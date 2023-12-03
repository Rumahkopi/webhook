package gcf

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/whatsauth/webhook"
)

func init() {
	functions.HTTP("WebHook", webhook.Post)
}
