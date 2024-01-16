package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/aiteung/atapi"
	"github.com/aiteung/atmessage"
	"github.com/aiteung/module/model"
	"github.com/whatsauth/wa"
	"net/http"
	"os"
	"strings"
)

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

			// List of admin phone numbers
			adminPhoneNumbers := []string{"6283174845017", "6285312924192"} // Add more admin numbers as needed

			// Forward the complaint to all admin phone numbers
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
		} else if strings.HasPrefix(strings.ToLower(msg.Message), "beli ") {
			// Handle the "beli [text client]" command
			clientText := strings.TrimPrefix(strings.ToLower(msg.Message), "beli ")
			reply := fmt.Sprintf("Anda akan membayar %s melalui metode pembayaran berikut:\n1. BCA\n2. Dana\n3. Gopay", clientText)

			// Send payment information to the user
			dt := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
		} else if strings.ToLower(msg.Message) == "fixbayar" {
			// Respond to the "fixbayar" command
			reply := "Silahkan kirim hasil screenshotan pembayaran beserta teks 'Sudah Bayar'."

			// Send instructions to the user
			dt := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
		} else {
			resp.Response = "Command not recognized"
		}
	}
}
