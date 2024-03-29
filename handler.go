package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aiteung/atapi"
	"github.com/aiteung/atmessage"
	"github.com/aiteung/module/model"
	"github.com/whatsauth/wa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv"
)

// MongoDB configuration
const (
	mongoConnectionString   = "mongodb+srv://Syahid25:4yyoi59f6p8GKGHT@syahid.jirstmg.mongodb.net/"
	mongoDBName             = "proyek3"
	mongoCollectionName     = "keluhan"
)

// MongoDB client
var mongoClient *mongo.Client

// Timezone for Indonesia (WIB - Western Indonesian Time)
var wib *time.Location

func init() {
	// Set the timezone for WIB
	wib, _ = time.LoadLocation("Asia/Jakarta")

	// Initialize MongoDB client
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoConnectionString))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}

	mongoClient = client
}

func closeMongoClient() {
	if mongoClient != nil {
		err := mongoClient.Disconnect(context.Background())
		if err != nil {
			fmt.Println("Error disconnecting MongoDB:", err)
		}
	}
}
const (
	transaksiCollectionName = "transaksi"
	transaksiStatusPending  = "Pending"
	transaksiStatusProcessing = "Proses"
	transaksiStatusCompleted  = "Sukses"
	transaksiStatusFailed    = "Gagal"
)
func insertTransactionData(paymentProof string, userPhone string, buktitf string) error {
	collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

	// Get the current count of transactions to determine the transaction number
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	transaksiNumber := count + 1

	transaction := bson.M{
		"transaksi_number": transaksiNumber,
		"payment_proof":    paymentProof,
		"user_phone":       userPhone,
		"buktitf":          buktitf,
		"status":           transaksiStatusPending, // Add status field
		"timestamp":        time.Now().In(wib),
		"formatted_time":   time.Now().In(wib).Format("Monday, 02-Jan-06 15:04:05 MST"),
	}

	_, err = collection.InsertOne(context.Background(), transaction)
	return err
}
// Function to get a transaction by its number
func getTransactionByNumber(transaksiNumber int) (bson.M, error) {
	collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)
	var transaction bson.M
	err := collection.FindOne(context.Background(), bson.M{"transaksi_number": transaksiNumber}).Decode(&transaction)
	return transaction, err
}
// Function to get a transaction status message
func getTransactionStatusMessage(transaction bson.M) string {
	var statusMessage string
	switch transaction["status"].(string) {
	case transaksiStatusPending:
		statusMessage = "Transaksi kamu sedang Pending."
	case transaksiStatusProcessing:
		statusMessage = "Transaksi kamu sedang Proses."
	case transaksiStatusCompleted:
		statusMessage = "Transaksi kamu sudah Sukses."
	case transaksiStatusFailed:
		statusMessage = "Transaksi kamu Gagal."
	default:
		statusMessage = "Unknown transaction status."
	}
	return statusMessage
}
func insertComplaintData(complaintContent string, userPhone string) error {
    collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

    // Get the current count of complaints to determine the complaint number
    count, err := collection.CountDocuments(context.Background(), bson.M{})
    if err != nil {
        return err
    }

    complaintNumber := count + 1

    // Prepare complaint document
    complaint := bson.M{
        "complaint_number": complaintNumber,
        "content":          complaintContent,
        "user_phone":       userPhone,
        "timestamp":        time.Now().In(wib),
        "formattedTime":    time.Now().In(wib).Format("Monday, 02-Jan-06 15:04:05 MST"),
    }

    // Insert document into MongoDB
    _, err = collection.InsertOne(context.Background(), complaint)
    return err
}
func getTransactionsByUserPhone(userPhone string) ([]string, error) {
	collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

	cursor, err := collection.Find(context.Background(), bson.M{"user_phone": userPhone})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var transactions []string
	for cursor.Next(context.Background()) {
		var transaction bson.M
		err := cursor.Decode(&transaction)
		if err != nil {
			return nil, err
		}
		transactionStr := fmt.Sprintf("Transaksi Number: %v\nTimestamp: %v\nPayment Proof: %s\nStatus: %s\n\n",
			transaction["transaksi_number"], transaction["formatted_time"], transaction["payment_proof"], transaction["status"])
		transactions = append(transactions, transactionStr)
	}

	return transactions, nil
}
//
//
func getAllComplaints() ([]string, error) {
    collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

    // Find all documents in the collection
    cursor, err := collection.Find(context.Background(), bson.M{})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())

    var complaints []string
    for cursor.Next(context.Background()) {
        var complaint bson.M
        err := cursor.Decode(&complaint)
        if err != nil {
            return nil, err
        }

        complaintStr := fmt.Sprintf("Complaint Number: %v\nTimestamp: %s\nUser Phone: %s\nComplaint Content: %s\n\n",
            complaint["complaint_number"], complaint["formattedTime"], complaint["user_phone"], complaint["content"])
        complaints = append(complaints, complaintStr)
    }

    return complaints, nil
}

func getAllTransactions() ([]string, error) {
	collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var transactions []string
	for cursor.Next(context.Background()) {
		var transaction bson.M
		err := cursor.Decode(&transaction)
		if err != nil {
			return nil, err
		}
		transactionStr := fmt.Sprintf("Transaksi Number: %v\nTimestamp: %v\nUser Phone: %s\nPayment Proof: %s\n status: %s\n\n",
			transaction["transaksi_number"], transaction["formatted_time"], transaction["user_phone"], transaction["payment_proof"], transaction["status"])
		transactions = append(transactions, transactionStr)
	}

	return transactions, nil
}

func deleteComplaintByContent(complaintContent string) error {
    collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

    // Delete the complaint with the specified content
    _, err := collection.DeleteOne(context.Background(), bson.M{"content": complaintContent})
    return err
}

// Function to delete a complaint by number
func deleteComplaintByNumber(complaintNumber int) error {
    collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

    // Delete the complaint with the specified number
    _, err := collection.DeleteOne(context.Background(), bson.M{"complaint_number": complaintNumber})
    return err
}

// Function to delete all complaints
func deleteAllComplaints() error {
	collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

	// Delete all documents in the collection
	_, err := collection.DeleteMany(context.Background(), bson.M{})
	return err
}
func deleteTransactionByNumber(transaksiNumber int) error {
	collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

	// Delete the transaction with the specified Transaksi Number
	_, err := collection.DeleteOne(context.Background(), bson.M{"transaksi_number": transaksiNumber})
	return err
}
// Function to delete all transactions
func deleteAllTransactions() error {
	collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

	// Delete all documents in the collection
	_, err := collection.DeleteMany(context.Background(), bson.M{})
	return err
}

func Post(w http.ResponseWriter, r *http.Request) {
	var msg model.IteungMessage
	var resp atmessage.Response
	json.NewDecoder(r.Body).Decode(&msg)

	if r.Header.Get("Secret") == os.Getenv("SECRET") {
		if strings.ToLower(msg.Message) == "ada masalah" {
			reply := "Silahkan tuliskan keluhan dan masalah Anda."
			dt := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
			} else if strings.HasPrefix(strings.ToLower(msg.Message), "order") {
				// Echo back the user's message
				reply := "Kamu sudah yakin dengan :\n" + msg.Message + "\nKamu bisa bayar melalui:\nBank BCA: 321321312 \nNo Dana: 088883211232\nNo Gopay: 088883211232 \nJika Kamu sudah membayarkan kamu bisa lakukan :\n bayar [link bukti screenshot transfer]\n\nJika Kamu Kesulitan bisa chat kontak support berikut ini:\nhttps://wa.me/6285312924192\nhttps://wa.me/6283174845017"
			
				dt := &wa.TextMessage{
					To:       msg.Phone_number,
					IsGroup:  false,
					Messages: reply,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "keluhan") || strings.HasPrefix(strings.ToLower(msg.Message), "masalah") {
			complaintContent := strings.TrimPrefix(strings.ToLower(msg.Message), "keluhan")
			complaintContent = strings.TrimPrefix(complaintContent, "masalah")
			complaintContent = strings.TrimSpace(complaintContent)

			err := insertComplaintData(complaintContent, msg.Phone_number)
			if err != nil {
				fmt.Println("Error inserting complaint data into MongoDB:", err)
			}

			adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
			for _, adminPhoneNumber := range adminPhoneNumbers {
				forwardMessage := fmt.Sprintf("Ada Masalah Baru:\n%s\nDari: https://wa.me/%s", complaintContent, msg.Phone_number)
				forwardDT := &wa.TextMessage{
					To:       adminPhoneNumber,
					IsGroup:  false,
					Messages: forwardMessage,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), forwardDT, "https://api.wa.my.id/api/send/message/text")
			}

			reply := "Terimakasih!!. Keluhan Anda telah kami terima. Silahkan tunggu respon dari admin Rumah Kopi dalam waktu 1 x 24 jam."
			ackDT := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "listkeluhan") {
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
			isAdmin := false
			for _, adminPhoneNumber := range adminPhoneNumbers {
				if msg.Phone_number == adminPhoneNumber {
					isAdmin = true
					break
				}
			}

			if isAdmin {
				complaints, err := getAllComplaints()
				if err != nil {
					fmt.Println("Error retrieving complaints from MongoDB:", err)
				}

				adminReply := "Daftar Keluhan:\n\n" + strings.Join(complaints, "\n")
				adminDT := &wa.TextMessage{
					To:       msg.Phone_number,
					IsGroup:  false,
					Messages: adminReply,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), adminDT, "https://api.wa.my.id/api/send/message/text")
			} else {
				resp.Response = "You are not authorized to access this command."
			}
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "deletekeluhan") {
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
			isAdmin := false
			for _, adminPhoneNumber := range adminPhoneNumbers {
				if msg.Phone_number == adminPhoneNumber {
					isAdmin = true
					break
				}
			}
		
			if isAdmin {
				// Check if the user provided a complaint number to delete
				parts := strings.Fields(msg.Message)
				if len(parts) == 2 {
					complaintNumberToDelete, err := strconv.Atoi(parts[1])
					if err != nil {
						resp.Response = "Invalid complaint number. Please provide a valid complaint number to delete."
						return
					}
		
					err = deleteComplaintByNumber(complaintNumberToDelete)
					if err != nil {
						fmt.Println("Error deleting complaint from MongoDB:", err)
					}
		
					adminReply := fmt.Sprintf("Keluhan dengan nomor '%d' berhasil dihapus.", complaintNumberToDelete)
					adminDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: adminReply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), adminDT, "https://api.wa.my.id/api/send/message/text")
				} else {
					resp.Response = "Please provide a valid complaint number to delete."
				}
			}
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "deleteallkeluhan") {
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
			isAdmin := false
			for _, adminPhoneNumber := range adminPhoneNumbers {
				if msg.Phone_number == adminPhoneNumber {
					isAdmin = true
					break
				}
			}

			if isAdmin {
				err := deleteAllComplaints()
				if err != nil {
					fmt.Println("Error deleting all complaints from MongoDB:", err)
				}

				adminReply := "Semua keluhan berhasil dihapus."
				adminDT := &wa.TextMessage{
					To:       msg.Phone_number,
					IsGroup:  false,
					Messages: adminReply,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), adminDT, "https://api.wa.my.id/api/send/message/text")
			} else {
				resp.Response = "You are not authorized to access this command."
			}
		} else 	if strings.HasPrefix(strings.ToLower(msg.Message), "listbayar") {
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
			isAdmin := false
			for _, adminPhoneNumber := range adminPhoneNumbers {
				if msg.Phone_number == adminPhoneNumber {
					isAdmin = true
					break
				}
			}
	
			if isAdmin {
				transactions, err := getAllTransactions()
				if err != nil {
					fmt.Println("Error retrieving transactions from MongoDB:", err)
				}
	
				adminReply := "Daftar Transaksi:\n\n" + strings.Join(transactions, "\n")
				adminDT := &wa.TextMessage{
					To:       msg.Phone_number,
					IsGroup:  false,
					Messages: adminReply,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), adminDT, "https://api.wa.my.id/api/send/message/text")
			} else {
				resp.Response = "You are not authorized to access this command."
			}
		}else if strings.HasPrefix(strings.ToLower(msg.Message), "deletebayar") {
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
			isAdmin := false
			for _, adminPhoneNumber := range adminPhoneNumbers {
				if msg.Phone_number == adminPhoneNumber {
					isAdmin = true
					break
				}
			}
	
			if isAdmin {
				transaksiNumberToDeleteStr := strings.TrimPrefix(strings.ToLower(msg.Message), "deletebayar")
				transaksiNumberToDeleteStr = strings.TrimSpace(transaksiNumberToDeleteStr)
	
				// Convert the Transaksi Number to integer
				transaksiNumberToDelete, err := strconv.Atoi(transaksiNumberToDeleteStr)
				if err != nil {
					// Handle invalid Transaksi Number format
					reply := "salah caranya weh , berikut caranya\ndeletebayar [idnya]"
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
					return
				}
	
				err = deleteTransactionByNumber(transaksiNumberToDelete)
				if err != nil {
					fmt.Println("Error deleting transaction from MongoDB:", err)
					reply := "Error deleting transaction. Please try again."
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
					return
				}
	
				adminReply := fmt.Sprintf("transaksi dengan id '%d' berhasil di hapus.", transaksiNumberToDelete)
				adminDT := &wa.TextMessage{
					To:       msg.Phone_number,
					IsGroup:  false,
					Messages: adminReply,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), adminDT, "https://api.wa.my.id/api/send/message/text")
			}
		}else if strings.HasPrefix(strings.ToLower(msg.Message), "deleteallbayar") {
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
			isAdmin := false
			for _, adminPhoneNumber := range adminPhoneNumbers {
				if msg.Phone_number == adminPhoneNumber {
					isAdmin = true
					break
				}
			}
	
			if isAdmin {
				err := deleteAllTransactions()
				if err != nil {
					fmt.Println("Error deleting all transactions from MongoDB:", err)
					reply := "Error deleting all transactions. Please try again."
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
					return
				}
	
				adminReply := "Semua transaksi berhasil terhapus."
				adminDT := &wa.TextMessage{
					To:       msg.Phone_number,
					IsGroup:  false,
					Messages: adminReply,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), adminDT, "https://api.wa.my.id/api/send/message/text")
			}
			} else if strings.HasPrefix(strings.ToLower(msg.Message), "bayar") || strings.HasPrefix(strings.ToLower(msg.Message), "pembayaran") {
				paymentProof := strings.TrimPrefix(strings.ToLower(msg.Message), "bayar")
				paymentProof = strings.TrimPrefix(paymentProof, "pembayaran")
				paymentProof = strings.TrimSpace(paymentProof)
			
				// Ensure that the client provides a valid image link
				if strings.HasPrefix(paymentProof, "http") {
					// Extract the image link from the message
					buktitf := paymentProof
			
					adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
					for _, adminPhoneNumber := range adminPhoneNumbers {
						// Forward image link to admin
						forwardMessage := fmt.Sprintf("Bukti Pembayaran Baru:\n%s\nDari: https://wa.me/%s\nKamu Wajib Approve pembayaran dengan cara:\napprovebayar [id transaksi]\nrejectbayar [id transaksi]\n Mau Liat id transaksi? pake command listbayar lahh brow...", paymentProof, msg.Phone_number)
						forwardDT := &wa.TextMessage{
							To:       adminPhoneNumber,
							IsGroup:  false,
							Messages: forwardMessage,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), forwardDT, "https://api.wa.my.id/api/send/message/text")
					}
			
					// Insert transaction data into MongoDB with status pending
					err := insertTransactionData(paymentProof, msg.Phone_number, buktitf)
					if err != nil {
						fmt.Println("Error inserting transaction data into MongoDB:", err)
					}
			
					reply := "Terimakasih!!. Bukti pembayaran Anda telah kami terima. Silahkan tunggu proses verifikasi."
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
				} else {
					// Respond if the provided input is not a valid URL
					reply := "salah cuyyy caranya itu: \n bayar [link bukti fotonya]"
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
				}
				} else if strings.HasPrefix(strings.ToLower(msg.Message), "approvebayar") {
					// Extract the transaction number from the message
					transaksiNumberStr := strings.TrimPrefix(strings.ToLower(msg.Message), "approvebayar")
					transaksiNumberStr = strings.TrimSpace(transaksiNumberStr)
				
					// Convert the transaction number to integer
					transaksiNumber, err := strconv.Atoi(transaksiNumberStr)
					if err != nil {
						// Handle invalid transaction number format
						reply := "Format kamu salah harusnya: \napprovebayar [idnya]"
						ackDT := &wa.TextMessage{
							To:       msg.Phone_number,
							IsGroup:  false,
							Messages: reply,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
						return
					}
				
					// Update the transaction status to processing
					collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)
					_, err = collection.UpdateOne(context.Background(), bson.M{"transaksi_number": transaksiNumber}, bson.D{{"$set", bson.D{{"status", transaksiStatusProcessing}}}})
					if err != nil {
						fmt.Println("Error updating transaction status in MongoDB:", err)
						reply := "Error updating transaction status. Please try again."
						ackDT := &wa.TextMessage{
							To:       msg.Phone_number,
							IsGroup:  false,
							Messages: reply,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
						return
					}
				
					// Find the transaction by transaction number
					var transaction bson.M
					err = collection.FindOne(context.Background(), bson.M{"transaksi_number": transaksiNumber}).Decode(&transaction)
					if err != nil {
						fmt.Println("Error retrieving transaction from MongoDB:", err)
						reply := "Error retrieving transaction. Please try again."
						ackDT := &wa.TextMessage{
							To:       msg.Phone_number,
							IsGroup:  false,
							Messages: reply,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
						return
					}
						
							// Send a confirmation message to the user
							userPhone := transaction["user_phone"].(string)
							reply := fmt.Sprintf("Selamat!! Pembayaran Anda telah kami terima dan akan segera diproses dan akan kita kirimkan. kamu bisa cek statusmu melalui perintah :\ncekstatus")
							ackDT := &wa.TextMessage{
								To:       userPhone,
								IsGroup:  false,
								Messages: reply,
							}
							resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")

							// Send a confirmation message to the admin
							adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
							for _, adminPhoneNumber := range adminPhoneNumbers {
								forwardMessage := fmt.Sprintf("Pembayaran dengan Transaksi Number [%d] sudah terkirimkan ke user [%s]", transaksiNumber, userPhone)
								forwardDT := &wa.TextMessage{
									To:       adminPhoneNumber,
									IsGroup:  false,
									Messages: forwardMessage,
								}
								_, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), forwardDT, "https://api.wa.my.id/api/send/message/text")
					}
					} else if strings.HasPrefix(strings.ToLower(msg.Message), "rejectbayar") {
					// Extract the transaction number from the message
					transaksiNumberStr := strings.TrimPrefix(strings.ToLower(msg.Message), "rejectbayar")
					transaksiNumberStr = strings.TrimSpace(transaksiNumberStr)
				
					// Convert the transaction number to integer
					transaksiNumber, err := strconv.Atoi(transaksiNumberStr)
					if err != nil {
						// Handle invalid transaction number format
						reply := "Format kamu salah harusnya: \nrejectbayar [idnya]"
						ackDT := &wa.TextMessage{
							To:       msg.Phone_number,
							IsGroup:  false,
							Messages: reply,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
						return
					}
				
					// Update the transaction status to completed with a value indicating that the payment was rejected
					collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)
					_, err = collection.UpdateOne(context.Background(), bson.M{"transaksi_number": transaksiNumber}, bson.D{{"$set", bson.D{{"status", transaksiStatusFailed}, {"payment_status", "rejected"}}}})
					if err != nil {
						fmt.Println("Error updating transaction status in MongoDB:", err)
						reply := "Error updating transaction status. Please try again."
						ackDT := &wa.TextMessage{
							To:       msg.Phone_number,
							IsGroup:  false,
							Messages: reply,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
						return
					}
				
					// Find the transaction by transaction number
					var transaction bson.M
					err = collection.FindOne(context.Background(), bson.M{"transaksi_number": transaksiNumber}).Decode(&transaction)
					if err != nil {
						fmt.Println("Error retrieving transaction from MongoDB:", err)
						reply := "Error retrieving transaction. Please try again."
						ackDT := &wa.TextMessage{
							To:       msg.Phone_number,
							IsGroup:  false,
							Messages: reply,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
						return
					}
				
					// Send a message to the user indicating that the payment has been rejected
					userPhone := transaction["user_phone"].(string)
					reply := fmt.Sprintf("Maaf!! Pembayaran Anda lalukan itu tidak sesuai maka kami tolak. Silahkan lakukan transaksi ulang ya...")
					ackDT := &wa.TextMessage{
						To:       userPhone,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
					// Send a confirmation message to the admin
					adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
					for _, adminPhoneNumber := range adminPhoneNumbers {
						forwardMessage := fmt.Sprintf("Pembayaran dengan Transaksi Number [%d] sudah terkirimkan ke user [%s]", transaksiNumber, userPhone)
						forwardDT := &wa.TextMessage{
							To:       adminPhoneNumber,
							IsGroup:  false,
							Messages: forwardMessage,
						}
						_, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), forwardDT, "https://api.wa.my.id/api/send/message/text")
					}
				}else if strings.HasPrefix(strings.ToLower(msg.Message), "cekstatus") {
					// Extract the transaction number from the message
					transaksiNumberStr := strings.TrimPrefix(strings.ToLower(msg.Message), "cekstatus")
					transaksiNumberStr = strings.TrimSpace(transaksiNumberStr)
			
					var transactions []string
					var err error
					userPhone := msg.Phone_number

					// If the user provides a transaction number, show the status for that transaction
					if transaksiNumberStr != "" {
						// Convert the transaction number to integer
						transaksiNumber, err := strconv.Atoi(transaksiNumberStr)
						if err != nil {
							// Handle invalid transaction number format
							reply := "format kamu salah weh!"
							ackDT := &wa.TextMessage{
								To:       msg.Phone_number,
								IsGroup:  false,
								Messages: reply,
							}
							resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
							return
						}
			
						// Get the transaction by its number
						transaction, err := getTransactionByNumber(transaksiNumber)
						if err != nil {
							fmt.Println("Error retrieving transaction from MongoDB:", err)
							reply := "Error retrieving transaction. Please try again."
							ackDT := &wa.TextMessage{
								To:       msg.Phone_number,
								IsGroup:  false,
								Messages: reply,
							}
							resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
							return
						}
			
						// Send the transaction status message to the user
						statusMessage := getTransactionStatusMessage(transaction)
						reply := fmt.Sprintf("Berikut adalah status transaksi kamu :\n%s", statusMessage)
						ackDT := &wa.TextMessage{
							To:       msg.Phone_number,
							IsGroup:  false,
							Messages: reply,
						}
						transactions = append(transactions, ackDT.Messages)
			
					} else {
						// Show all transactions for the user
						transactions, err = getTransactionsByUserPhone(userPhone)
						if err != nil {
							fmt.Println("Error retrieving transactions from MongoDB:", err)
							reply := "Error retrieving transactions. Please try again."
							ackDT := &wa.TextMessage{
								To:       msg.Phone_number,
								IsGroup:  false,
								Messages: reply,
							}
							resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
							return
						}
					}
			
					// Send the transaction history to the user
					reply := "Jika sudah pesanan sampai kamu bisa gunakan perintah berikut ini untuk menyelesaikan status pengiriman:\npesanansampai [id nomer riwayat transaksi]\nBerikut adalah riwayat transaksi kamu:\n" + strings.Join(transactions, "\n")
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
					} else if strings.HasPrefix(strings.ToLower(msg.Message), "pesanansampai") {
						// Extract the transaction number from the message
						transaksiNumberStr := strings.TrimPrefix(strings.ToLower(msg.Message), "pesanansampai")
						transaksiNumberStr = strings.TrimSpace(transaksiNumberStr)
					
						// Convert the transaction number to integer
						transaksiNumber, err := strconv.Atoi(transaksiNumberStr)
						if err != nil {
							// Handle invalid transaction number format
							reply := "salah caranya kak , caranya adalah :\npesanansampai [number riwayat transaksi]"
							ackDT := &wa.TextMessage{
								To:       msg.Phone_number,
								IsGroup:  false,
								Messages: reply,
							}
							resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
							return
						}
					
						// Update the transaction status to completed
						collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)
						_, err = collection.UpdateOne(context.Background(), bson.M{"transaksi_number": transaksiNumber}, bson.D{{"$set", bson.D{{"status", transaksiStatusCompleted}}}})
						if err != nil {
							fmt.Println("Error updating transaction status in MongoDB:", err)
							reply := "Error updating transaction status. Please try again."
							ackDT := &wa.TextMessage{
								To:       msg.Phone_number,
								IsGroup:  false,
								Messages: reply,
							}
							resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
							return
						}
					
						// Find the transaction by transaction number
						var transaction bson.M
						err = collection.FindOne(context.Background(), bson.M{"transaksi_number": transaksiNumber}).Decode(&transaction)
						if err != nil {
							fmt.Println("Error retrieving transaction from MongoDB:", err)
							reply := "Error retrieving transaction. Please try again."
							ackDT := &wa.TextMessage{
								To:       msg.Phone_number,
								IsGroup:  false,
								Messages: reply,
							}
							resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
							return
						}
					
						// Send a message to the user indicating that the transaction is completed
						userPhone := transaction["user_phone"].(string)
						reply := fmt.Sprintf("transaksi kamu dengan number %d telah di selesaikan.", transaksiNumber)
						ackDT := &wa.TextMessage{
							To:       userPhone,
							IsGroup:  false,
							Messages: reply,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
					}else {
				resp.Response = "Command not recognized"
			}
		}
	}