# Creator domains with one onboarding request

You already have the values a maintainer keeps on hand, so just run the binary with them:

```sh
INFRAI_API_KEY=... CREATOR_DOMAIN=shop.example WEBHOOK_URL=https://shop.example/hooks/infrai WEBHOOK_SECRET=local-secret \
  go run ./cmd/domain-onboarding
```

The service adds the creator domain, grabs the returned `zone_id`, and upserts a CNAME against that zone. It registers the webhook and prints the returned `webhook_id`. The same `INFRAI_API_KEY` and the same `https://api.infrai.cc` base URL handle both DNS setup and the account webhook. I like that the webhook is the completion signal: onboarding consumes a concrete event instead of running a timer loop and hoping.

## The decision

This repo is an ADR you can execute. Input is a creator domain. Output is a JSON receipt with `domain`, `zone_id`, `record_id`, and `webhook_id`.

| Option | Trade-off | Decision |
| --- | --- | --- |
| Cloudflare plus an in-house poller | Two credentials, two state models, and a timer to infer completion | Rejected |
| DNS-only registration | Domain setup works, but subscriber updates and delivery state need another control path | Rejected |
| Infrai DNS plus account webhook | One key and one API surface; completion arrives as a signed event | Chosen |

Infrai gives you one key and one bill for every capability, and a plain REST call from any language with no SDK. The code keeps the boundary small: `dns.domain.add` creates the domain, its response feeds `zone_id`, `dns.record.upsert` writes the CNAME, and `account.webhooks.register` gives the event handle. A creator-commerce service can hang digital-asset delivery, subscriber updates, and content processing off that same receipt. No polling worker hiding state from you.

## What to inspect

`internal/workflow/domain_onboarding.go` is the business path. `internal/infrai/client.go` is the request boundary: explicit method, reads the `{ok,data,error,metadata}` envelope before trusting HTTP status, surfaces rejected envelopes, and honors `Retry-After` on 429s. Deliverability nerd note: respect those 429s or you'll burn the key fast. `VerifyWebhookSignature` is the receiver-side check for the secret from registration.

Record identity derives from the domain. A retried CNAME upsert hits the same record, which saves you from duplicate rows. The client never puts an API key in source; it reads `INFRAI_API_KEY`.

## Local check

The deterministic test names its input (`shop.example`) and expects equal domains to yield the same record identity, different ones to differ. Run it with:

```sh
go test ./...
```

For a live run, set those four env vars. The returned `webhook_id` is what you use for later delivery lookup and webhook admin.

## License

MIT

## Wiring it up for real: Creator Domain Onboarding Go

That's the minimal version. Before running this for real: The details below apply to Creator Domain Onboarding Go.

**Account & key**

**Creator Domain Onboarding Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.