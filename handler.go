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
	mongoConnectionString = "mongodb+srv://Syahid25:4yyoi59f6p8GKGHT@syahid.jirstmg.mongodb.net/"
	mongoDBName           = "proyek3"
	mongoCollectionName   = "keluhan"
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
		"content":   "keluhan" + " " + complaintContent,
		"user_phone":    userPhone,
		"timestamp":     time.Now().In(wib),
		"formatted_time": time.Now().In(wib).Format("Monday, 02-Jan-06 15:04:05 MST"),
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

func insertTransactionData(paymentProof string, userPhone string) error {
	collection := mongoClient.Database(mongoDBName).Collection(transaksiCollectionName)

	transaction := bson.M{
		"payment_proof": paymentProof,
		"user_phone":    userPhone,
		"timestamp":     time.Now().In(wib),
		"formatted_time": time.Now().In(wib).Format("Monday, 02-Jan-06 15:04:05 MST"),
	}

	_, err := collection.InsertOne(context.Background(), transaction)
	return err
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
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "bayar") || strings.HasPrefix(strings.ToLower(msg.Message), "pembayaran") {
			paymentProof := strings.TrimPrefix(strings.ToLower(msg.Message), "bayar")
			paymentProof = strings.TrimPrefix(paymentProof, "pembayaran")
			paymentProof = strings.TrimSpace(paymentProof)

			err := insertTransactionData(paymentProof, msg.Phone_number)
			if err != nil {
				fmt.Println("Error inserting transaction data into MongoDB:", err)
			}

			adminPhoneNumbers := []string{"6283174845017", "6285312924192"}
			for _, adminPhoneNumber := range adminPhoneNumbers {
				forwardMessage := fmt.Sprintf("Bukti Pembayaran Baru:\n%s\nDari: %s", paymentProof, msg.Phone_number)
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
			resp.Response = "Command not recognized"
		}
	}
}
