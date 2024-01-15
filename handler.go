package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/aiteung/atapi"
	"github.com/aiteung/atmessage"
	"github.com/aiteung/module/model"
	"github.com/whatsauth/wa"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"github.com/bwmarrin/discordgo"
)

func Post(w http.ResponseWriter, r *http.Request) {
	var msg model.IteungMessage
	var resp atmessage.Response
	json.NewDecoder(r.Body).Decode(&msg)
	link := "https://medium.com/@gilarwahibul/whatsauth-free-2fa-otp-notif-whatsapp-gateway-api-gratis-ab4d04f80601"
	if r.Header.Get("Secret") == os.Getenv("SECRET") {
		if msg.Message == "loc" || msg.Message == "Loc" || msg.Message == "lokasi" || msg.LiveLoc {
			location, err := ReverseGeocode(msg.Latitude, msg.Longitude)
			if err != nil {
				location = "Unknown Location"
			}

			reply := fmt.Sprintf("Hai hai haiii kamu pasti lagi di %s \n Koordinatenya : %s - %s\n Cara Penggunaan WhatsAuth Ada di link dibawah ini"+
				"yaa kak %s\n", location,
				strconv.Itoa(int(msg.Longitude)), strconv.Itoa(int(msg.Latitude)), link)
			dt := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
		} else if msg.Message == "hallo" || msg.Message == "Hallo" {
			// Respond to "hallo" command
			reply := "Hallo juga! Apa yang bisa saya bantu?"
			dt := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://api.wa.my.id/api/send/message/text")
		} else if strings.HasPrefix(msg.Message, "report") || strings.HasPrefix(msg.Message, "Report") {
			// Handle the reporting process here
			reportContent := strings.TrimPrefix(msg.Message, "report")
			reportContent = strings.TrimPrefix(reportContent, "Report")
			reportContent = strings.TrimSpace(reportContent)
	
			// Send the report to the Discord webhook
			discordWebhookURL := "https://discord.com/api/webhooks/1170045069018026025/df2hxPLEoQOblNvZK3b4RxwzNp7JVtZ66nqSiRLrGBbFpxeBtbhLyH3isvEtpLOoTbKj" // Replace with your Discord webhook URL
			reportNotification := fmt.Sprintf("🚨 New Report Received!\n📝 Report Content: %s\n📱 Reporter's Number: %s", reportContent, msg.Phone_number)
	
			// Send report notification to Discord
			err := sendToDiscordWebhook(discordWebhookURL, reportNotification)
			if err != nil {
				// Handle the error (e.g., log it)
			}
	
			// Send acknowledgment to the reporter
			reply := "Terima kasih! Laporan Anda telah diterima dan akan segera ditindaklanjuti."
	
			// Send acknowledgment to reporter via WhatsApp
			ackDT := &wa.TextMessage{
				To:       msg.Phone_number,
				IsGroup:  false,
				Messages: reply,
			}
			resp, _ = atapi.PostStructWithToken[atmessage.Response](os.Getenv("TOKEN"), os.Getenv("TOKEN"), ackDT, "https://api.wa.my.id/api/send/message/text")
		} else {
			resp.Response = "Secret Salah"
		}
}
}
func sendToDiscordWebhook(webhookURL, message string) error {
	// Create a new HTTP POST request to the Discord webhook
	req, err := http.NewRequest("POST", webhookURL, strings.NewReader(fmt.Sprintf(`{"content": "%s"}`, message)))
	if err != nil {
		return err
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Perform the HTTP request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK status code: %d", resp.StatusCode)
	}

	return nil
}

func ReverseGeocode(latitude, longitude float64) (string, error) {
	// OSM Nominatim API endpoint
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%f&lon=%f", latitude, longitude)

	// Make a GET request to the API
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Decode the response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	// Extract the place name from the response
	displayName, ok := result["display_name"].(string)
	if !ok {
		return "", fmt.Errorf("unable to extract display_name from the API response")
	}

	return displayName, nil
}

func Liveloc(w http.ResponseWriter, r *http.Request) {
	var msg model.IteungMessage
	var resp atmessage.Response
	json.NewDecoder(r.Body).Decode(&msg)

	// Reverse geocode to get the place name
	location, err := ReverseGeocode(msg.Latitude, msg.Longitude)
	if err != nil {
		// Handle the error (e.g., log it) and set a default location name
		location = "Unknown Location"
	}

	reply := fmt.Sprintf("Hai hai haiii kamu pasti lagi di %s \n Koordinatenya : %s - %s\n", location,
		strconv.Itoa(int(msg.Longitude)), strconv.Itoa(int(msg.Latitude)))

	if r.Header.Get("Secret") == os.Getenv("SECRET") {
		dt := &wa.TextMessage{
			To:       msg.Phone_number,
			IsGroup:  false,
			Messages: reply,
		}
		resp, _ = atapi.PostStructWithToken[atmessage.Response]("Token", os.Getenv("TOKEN"), dt, "https://cloud.wa.my.id/api/send/message/text")
	} else {
		resp.Response = "Secret Salah"
	}
	fmt.Fprintf(w, resp.Response)
}

func GetRandomString(strings []string) string {
	randomIndex := rand.Intn(len(strings))
	return strings[randomIndex]
}
