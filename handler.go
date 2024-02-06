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
type Transaction struct {
    TransaksiNumber int    `bson:"transaksi_number"`
    PaymentProof    string `bson:"payment_proof"`
    UserPhone       string `bson:"user_phone"`
    BuktiTf         string `bson:"buktitf"`
    Timestamp       time.Time `bson:"timestamp"`
    FormattedTime   string `bson:"formatted_time"`
    Status          string `bson:"status"`
}
func insertTransactionData(paymentProof string, userPhone string, buktitf string) error {
    collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

    count, err := collection.CountDocuments(context.Background(), bson.M{})
    if err != nil {
        return err
    }

    transaksiNumber := count + 1

    transaction := Transaction{
        TransaksiNumber: transaksiNumber,
        PaymentProof:    paymentProof,
        UserPhone:       userPhone,
        BuktiTf:         buktitf,
        Timestamp:       time.Now().In(wib),
        FormattedTime:   time.Now().In(wib).Format("Monday, 02-Jan-06 15:04:05 MST"),
        Status:          "Pending", // Set status to Pending initially
    }

    _, err = collection.InsertOne(context.Background(), transaction)
    return err
}
func approvePayment(transaksiNumber int, isApproved bool) error {
    collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

    filter := bson.M{"transaksi_number": transaksiNumber}
    update := bson.M{}

    if isApproved {
        update["$set"] = bson.M{"status": "Processing"}
    } else {
        update["$set"] = bson.M{"status": "Cancelled"}
    }

    _, err := collection.UpdateOne(context.Background(), filter, update)
    return err
}

func informUserAboutPaymentStatus(transaksiNumber int, userPhone string, status string) error {
    var message string

    switch status {
    case "Processing":
        message = "Pembayaran Anda telah diterima dan sedang diproses."
    case "Cancelled":
        message = "Maaf, pembayaran Anda telah ditolak."
    }

    dt := &wa.TextMessage{
        To:       userPhone,
        IsGroup:  false,
        Messages: message,
    }

    _, err := atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
    return err
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
		transactionStr := fmt.Sprintf("Transaksi Number: %v\nTimestamp: %v\nUser Phone: %s\nPayment Proof: %s\n\n",
			transaction["transaksi_number"], transaction["formatted_time"], transaction["user_phone"], transaction["payment_proof"])
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
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "beli") {
			// Echo back the user's message
			reply := "kamu sudah yakin dengan :\n" + msg.Message + "\nkamu bisa bayar melalui:\nBank BCA: 321321312 \nNo Dana: 088883211232\nNo Gopay: 088883211232 \nJika Kamu sudah membayarkan kamu bisa lakukan :\n bayar [link bukti screenshot transfer]\n\nJika Kamu Kesulitan bisa chat kontak support berikut ini:\nhttps://wa.me/6285312924192\nhttps://wa.me/6283174845017"
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
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "listbayar") {
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
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "deletebayar") {
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

				err = deleteTransactionByNumber(int64(transaksiNumberToDelete))
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
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "deleteallbayar") {
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
			if strings.HasPrefix(strings.ToLower(msg.Message), "cekstatus") {
				// Handle check status command
				parts := strings.Fields(msg.Message)
				if len(parts) == 1 {
					// Get user's phone number
					userPhone := msg.Phone_number

					// Retrieve user's transactions from MongoDB
					transactions, err := getTransactionsByUser(userPhone)
					if err != nil {
						fmt.Println("Error retrieving user's transactions from MongoDB:", err)
						return
					}

					// Construct message with transaction details
					var message string
					if len(transactions) == 0 {
						message = "Anda belum memiliki transaksi."
					} else {
						message = "Status Transaksi Anda:\n\n"
						for _, transaction := range transactions {
							message += fmt.Sprintf("Transaksi Number: %v\nTimestamp: %v\nStatus: %s\n\n",
								transaction.TransaksiNumber, transaction.FormattedTime, transaction.Status)
						}
					}

					// Send message to user
					dt := &wa.TextMessage{
						To:       userPhone,
						IsGroup:  false,
						Messages: message,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
				} else {
					reply := "Format pesan tidak valid. Gunakan format: cekstatus"
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
					return
				}
			} else if strings.HasPrefix(strings.ToLower(msg.Message), "bayar") || strings.HasPrefix(strings.ToLower(msg.Message), "pembayaran") {
				paymentProof := strings.TrimPrefix(strings.ToLower(msg.Message), "bayar")
				paymentProof = strings.TrimPrefix(paymentProof, "pembayaran")
				paymentProof = strings.TrimSpace(paymentProof)

				// Ensure that the client provides a valid image link
				if strings.HasPrefix(paymentProof, "http") {
					buktitf := paymentProof

					err := insertTransactionData(paymentProof, msg.Phone_number, buktitf)
					if err != nil {
						fmt.Println("Error inserting transaction data into MongoDB:", err)
					}

					adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
					for _, adminPhoneNumber := range adminPhoneNumbers {
						forwardMessage := fmt.Sprintf("Bukti Pembayaran Baru:\n%s\nDari: https://wa.me/%s", paymentProof, msg.Phone_number)
						forwardDT := &wa.TextMessage{
							To:       adminPhoneNumber,
							IsGroup:  false,
							Messages: forwardMessage,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), forwardDT, "https://api.wa.my.id/api/send/message/text")
					}

					reply := "Terimakasih!!. Bukti pembayaran Anda telah kami terima. Silahkan tunggu proses verifikasi."
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
				} else {
					reply := "Bukti pembayaran tidak valid. Silahkan masukkan link gambar bukti transfer."
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
				}
			} else if strings.HasPrefix(strings.ToLower(msg.Message), "setuju") || strings.HasPrefix(strings.ToLower(msg.Message), "tolak") {
				// Handle admin approval or rejection of payment
				parts := strings.Fields(msg.Message)
				if len(parts) >= 3 {
					transaksiNumber, err := strconv.Atoi(parts[1])
					if err != nil {
						reply := "Format pesan tidak valid. Gunakan format: setuju/tolak [nomor_transaksi]"
						ackDT := &wa.TextMessage{
							To:       msg.Phone_number,
							IsGroup:  false,
							Messages: reply,
						}
						resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
						return
					}

					isApproved := strings.ToLower(parts[0]) == "setuju"
					err = approvePayment(int64(transaksiNumber), isApproved)
					if err != nil {
						fmt.Println("Error updating payment status in MongoDB:", err)
						return
					}

					// Inform user about the payment status
					transaction, err := getTransactionByNumber(int64(transaksiNumber))
					if err != nil {
						fmt.Println("Error retrieving transaction data from MongoDB:", err)
						return
					}

					err = informUserAboutPaymentStatus(transaction.TransaksiNumber, transaction.UserPhone, transaction.Status)
					if err != nil {
						fmt.Println("Error sending payment status message to user:", err)
					}
				} else {
					reply := "Format pesan tidak valid. Gunakan format: setuju/tolak [nomor_transaksi]"
					ackDT := &wa.TextMessage{
						To:       msg.Phone_number,
						IsGroup:  false,
						Messages: reply,
					}
					resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
					return
				}
			}
		}

		json.NewEncoder(w).Encode(resp)
	}
}
	
func getTransactionByNumber(transaksiNumber int) (Transaction, error) {
    var transaction Transaction

    collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

    filter := bson.M{"transaksi_number": transaksiNumber}
    err := collection.FindOne(context.Background(), filter).Decode(&transaction)
    if err != nil {
        return Transaction{}, err
    }

    return transaction, nil
}
func getTransactionsByUser(userPhone string) ([]Transaction, error) {
    var transactions []Transaction

    collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

    filter := bson.M{"user_phone": userPhone}
    cursor, err := collection.Find(context.Background(), filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())

    for cursor.Next(context.Background()) {
        var transaction Transaction
        err := cursor.Decode(&transaction)
        if err != nil {
            return nil, err
        }
        transactions = append(transactions, transaction)
    }

    return transactions, nil
}