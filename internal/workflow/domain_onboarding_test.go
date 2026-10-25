package workflow

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"custom-domain-delivery/internal/infrai"
)

func TestWebhookSignatureAndStableRecordDecision(t *testing.T) {
	secret, payload := "local-secret", []byte(`{"event":"account.updated"}`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	if !infrai.VerifyWebhookSignature(secret, payload, signature) {
		t.Fatal("valid webhook signature was rejected")
	}
	if stableID("shop.example") != stableID("shop.example") {
		t.Fatal("same domain must keep the same record identity")
	}
	if stableID("shop.example") == stableID("other.example") {
		t.Fatal("different domains must not share a record identity")
	}
}
