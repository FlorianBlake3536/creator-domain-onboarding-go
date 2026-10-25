package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"custom-domain-delivery/internal/infrai"
	"custom-domain-delivery/internal/workflow"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	domain := os.Getenv("CREATOR_DOMAIN")
	webhookURL := os.Getenv("WEBHOOK_URL")
	secret := os.Getenv("WEBHOOK_SECRET")
	result, err := workflow.New(infrai.NewClient(key)).Onboard(workflow.DomainRequest{Domain: domain, WebhookURL: webhookURL, WebhookSecret: secret})
	if err != nil {
		log.Fatal(err)
	}
	encoded, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(encoded))
}
