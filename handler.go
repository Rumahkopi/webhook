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
)

// MongoDB client
var mongoClient *mongo.Client

// Initialize MongoDB client
func init() {
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

// Close MongoDB client
func closeMongoClient() {
	if mongoClient != nil {
		err := mongoClient.Disconnect(context.Background())
		if err != nil {
			fmt.Println("Error disconnecting MongoDB:", err)
		}
	}
}

// Function to insert complaint data into MongoDB
func insertComplaintData(complaintContent string, userPhone string) error {
	collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

	// Prepare complaint document
	complaint := bson.M{
		"content":   "keluhan" + complaintContent, // Prefix "keluhan" to the content
		"user_phone": userPhone,
		"timestamp":  time.Now(),
	}

	// Insert document into MongoDB
	_, err := collection.InsertOne(context.Background(), complaint)
	return err
}

// Function to retrieve all complaint data from MongoDB
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
		complaintStr := fmt.Sprintf("Timestamp: %v\nUser Phone: %s\nComplaint Content: %s\n\n",
			complaint["timestamp"], complaint["user_phone"], complaint["content"])
		complaints = append(complaints, complaintStr)
	}

	return complaints, nil
}

// Function to delete a complaint by content
func deleteComplaintByContent(complaintContent string) error {
	collection := mongoClient.Database(mongoDBName).Collection(mongoCollectionName)

	// Delete the complaint with the specified content
	_, err := collection.DeleteOne(context.Background(), bson.M{"content": complaintContent})
	return err
}

func Post(w http.ResponseWriter, r *http.Request) {
	var msg model.IteungMessage
	var resp atmessage.Response
	json.NewDecoder(r.Body).Decode(&msg)

	if r.Header.Get("Secret") == os.Getenv("SECRET") {
		if strings.ToLower(msg.Message) == "ada masalah" {
			// Respond to "ada masalah" command
			reply := "Silahkan tuliskan keluhan dan masalah Anda."
			dt := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "keluhan") || strings.HasPrefix(strings.ToLower(msg.Message), "masalah") {
			// Handle the complaint process here
			complaintContent := strings.TrimPrefix(strings.ToLower(msg.Message), "keluhan")
			complaintContent = strings.TrimPrefix(complaintContent, "masalah")
			complaintContent = strings.TrimSpace(complaintContent)

			// Insert complaint data into MongoDB
			err := insertComplaintData(complaintContent, msg.Phone_number)
			if err != nil {
				fmt.Println("Error inserting complaint data into MongoDB:", err)
				// Handle the error (e.g., log it)
			}

			// Forward the complaint to all admin phone numbers
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"} // Add more admin numbers as needed
			for _, adminPhoneNumber := range adminPhoneNumbers {
				forwardMessage := fmt.Sprintf("Ada Masalah Baru:\n%s\nDari: %s", complaintContent, msg.Phone_number)
				forwardDT := &wa.TextMessage{
					To:       adminPhoneNumber,
					IsGroup:  false,
					Messages: forwardMessage,
				}
				resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), forwardDT, "https://api.wa.my.id/api/send/message/text")
			}

			// Send acknowledgment to the user
			reply := "Terimakasih!!. Keluhan Anda telah kami terima. Silahkan tunggu respon dari admin Rumah Kopi dalam waktu 1 x 24 jam."
			ackDT := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "listkeluhan") {
			// Handle the "listkeluhan" command for admin
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"} // Add more admin numbers as needed

			// Check if the sender is an admin
			isAdmin := false
			for _, adminPhoneNumber := range adminPhoneNumbers {
				if msg.Phone_number == adminPhoneNumber {
					isAdmin = true
					break
				}
			}

			if isAdmin {
				// Retrieve all complaints from MongoDB
				complaints, err := getAllComplaints()
				if err != nil {
					fmt.Println("Error retrieving complaints from MongoDB:", err)
					// Handle the error (e.g., log it)
				}

				// Prepare and send the list of complaints to the admin
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
			// Handle the "deletekeluhan [complaintContent]" command for admin
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"} // Add more admin numbers as needed

			// Check if the sender is an admin
			isAdmin := false
			for _, adminPhoneNumber := range adminPhoneNumbers {
				if msg.Phone_number == adminPhoneNumber {
					isAdmin = true
					break
				}
			}

			if isAdmin {
				// Extract the complaint content from the command
				complaintContentToDelete := strings.TrimPrefix(strings.ToLower(msg.Message), "deletekeluhan ")
				complaintContentToDelete = strings.TrimSpace(complaintContentToDelete)

				// Delete the specified complaint
				err := deleteComplaintByContent("keluhan" + complaintContentToDelete)
				if err != nil {
					fmt.Println("Error deleting complaint from MongoDB:", err)
					// Handle the error (e.g., log it)
				}

				// Send acknowledgment to the admin
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
		} else {
			resp.Response = "Command not recognized"
		}
	}
}
