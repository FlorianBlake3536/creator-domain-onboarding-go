package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"custom-domain-delivery/internal/infrai"
)

type DomainRequest struct {
	Domain        string
	WebhookURL    string
	WebhookSecret string
}

type OnboardingResult struct {
	Domain    string
	ZoneID    string
	WebhookID string
	RecordID  string
}

type service struct{ api *infrai.Client }

func New(api *infrai.Client) *service { return &service{api: api} }

func (s *service) Onboard(input DomainRequest) (OnboardingResult, error) {
	if input.Domain == "" || input.WebhookURL == "" || input.WebhookSecret == "" {
		return OnboardingResult{}, fmt.Errorf("domain, webhook URL, and webhook secret are required")
	}
	var domain struct {
		ZoneID string `json:"zone_id"`
	}
	metadata := map[string]string{"request_id": stableID(input.Domain)}
	if err := s.api.Do("POST", "/v1/dns/domain/add", map[string]any{"domain": input.Domain, "metadata": metadata}, &domain); err != nil {
		return OnboardingResult{}, err
	}
	if domain.ZoneID == "" {
		return OnboardingResult{}, fmt.Errorf("domain response did not include zone_id")
	}
	var record struct {
		RecordID string `json:"record_id"`
	}
	if err := s.api.Do("PUT", "/v1/dns/record/upsert", map[string]any{
		"zone_id": domain.ZoneID, "record_type": "CNAME", "name": "www", "content": input.Domain, "metadata": metadata,
	}, &record); err != nil {
		return OnboardingResult{}, err
	}
	var hook struct {
		WebhookID string `json:"webhook_id"`
	}
	if err := s.api.Do("POST", "/v1/account/webhooks/register", map[string]any{
		"url": input.WebhookURL, "events": []string{"account.updated"}, "secret": input.WebhookSecret,
	}, &hook); err != nil {
		return OnboardingResult{}, err
	}
	if hook.WebhookID == "" {
		return OnboardingResult{}, fmt.Errorf("webhook response did not include webhook_id")
	}
	return OnboardingResult{Domain: input.Domain, ZoneID: domain.ZoneID, WebhookID: hook.WebhookID, RecordID: record.RecordID}, nil
}

func stableID(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}
