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
)

// MongoDB configuration
const (
	mongoConnectionString   = "mongodb+srv://Syahid25:4yyoi59f6p8GKGHT@syahid.jirstmg.mongodb.net/"
	mongoDBName             = "proyek3"
	mongoCollectionName     = "keluhan"
	transaksiCollectionName = "transaksi"
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

func insertComplaintData(complaintContent string, userPhone string) error {
	collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

	complaint := bson.M{
		"content":         "keluhan" + " " + complaintContent,
		"user_phone":      userPhone,
		"timestamp":       time.Now().In(wib),
		"formatted_time":  time.Now().In(wib).Format("Monday, 02-Jan-06 15:04:05 MST"),
	}

	_, err := collection.InsertOne(context.Background(), complaint)
	return err
}

func getAllComplaints() ([]string, error) {
	collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

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
		complaintStr := fmt.Sprintf("Timestamp: %v\nUser Phone: %s\nComplaint Content: %s\n\n",
			complaint["timestamp"], complaint["user_phone"], complaint["content"])
		complaints = append(complaints, complaintStr)
	}

	return complaints, nil
}

func insertTransactionData(paymentProof string, userPhone string, buktitf string) error {
	collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

	transaction := bson.M{
		"payment_proof":   paymentProof,
		"user_phone":      userPhone,
		"buktitf":         buktitf,
		"timestamp":       time.Now().In(wib),
		"formatted_time":  time.Now().In(wib).Format("Monday, 02-Jan-06 15:04:05 MST"),
	}

	_, err := collection.InsertOne(context.Background(), transaction)
	return err
}

// ...

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
		transactionStr := fmt.Sprintf("Transaksi Number: %v\nTimestamp: %v\nUser Phone: %s\nPayment Proof: %s\n\n",
			transaction["transaksi_number"], transaction["timestamp"], transaction["user_phone"], transaction["payment_proof"])
		transactions = append(transactions, transactionStr)
	}

	return transactions, nil
}


// Function to delete a complaint by content
func deleteComplaintByContent(complaintContent string) error {
	collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

	// Delete the complaint with the specified content
	_, err := collection.DeleteOne(context.Background(), bson.M{"content": complaintContent})
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
				forwardMessage := fmt.Sprintf("Ada Masalah Baru:\n%s\nDari: %s", complaintContent, msg.Phone_number)
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
				complaintContentToDelete := strings.TrimPrefix(strings.ToLower(msg.Message), "deletekeluhan ")
				complaintContentToDelete = strings.TrimSpace(complaintContentToDelete)

				err := deleteComplaintByContent("keluhan" + " " + complaintContentToDelete)
				if err != nil {
					fmt.Println("Error deleting complaint from MongoDB:", err)
				}

				adminReply := fmt.Sprintf("Keluhan dengan konten '%s' berhasil dihapus.", complaintContentToDelete)
				adminDT := &wa.TextMessage{
					To:       msg.Phone_number,
					IsGroup:  false,
					Messages: adminReply,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), adminDT, "https://api.wa.my.id/api/send/message/text")
			} else {
				resp.Response = "You are not authorized to access this command."
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
					reply := "Invalid Transaksi Number format. Please provide a valid Transaksi Number."
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
	
				adminReply := fmt.Sprintf("Transaction with Transaksi Number '%d' successfully deleted.", transaksiNumberToDelete)
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
	
				adminReply := "All transactions successfully deleted."
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
						forwardMessage := fmt.Sprintf("Bukti Pembayaran Baru:\n%s\nDari: %s", paymentProof, msg.Phone_number)
						forwardDT := &wa.TextMessage{
							To:       adminPhoneNumber,
							IsGroup:  false,
							Messages: forwardMessage,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), forwardDT, "https://api.wa.my.id/api/send/message/text")
					}
	
					// Insert transaction data into MongoDB
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
			} else {
				resp.Response = "Command not recognized"
			}
		}
	}