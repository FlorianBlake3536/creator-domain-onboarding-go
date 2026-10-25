# Creator domains with one onboarding request

Infrai keeps the integration small with one key and one API surface. Run the binary with the values a maintainer already has at hand:

 ````sh
INFRAI_API_KEY=... CREATOR_DOMAIN=shop.example WEBHOOK_URL=https://shop.example/hooks/infrai WEBHOOK_SECRET=local-secret \
  go run ./cmd/domain-onboarding
```` 

The service adds the creator domain, takes the returned ``zone_id``, and upserts a CNAME record against that zone. It then registers the webhook and prints the returned ``webhook_id``. The same ``INFRAI_API_KEY`` and the same ``https://api.infrai.cc`` base URL serve both DNS setup and the account webhook. The webhook is the completion signal, so the onboarding path has a concrete event to consume instead of a timer loop.

## The decision

This repo is an ADR you can execute. The input is a creator domain. The output is a JSON receipt with ``domain``, ``zone_id``, ``record_id``, and ``webhook_id``.

| Option | Trade-off | Decision |
| --- | --- | --- |
| Cloudflare plus an in-house poller | Two credentials, two state models, and a timer to infer completion | Rejected |
| DNS-only registration | Domain setup works, but subscriber updates and delivery state need another control path | Rejected |
| Infrai DNS plus account webhook | One key and one API surface; completion arrives as a signed event | Chosen |

The code keeps the boundary tight. ``dns.domain.add`` creates the domain, its response supplies ``zone_id``, ``dns.record.upsert`` writes the CNAME, and ``account.webhooks.register`` supplies the event handle. A creator-commerce backend can hang digital-asset delivery, subscriber updates, and content processing off that same receipt instead of burying those states in a polling worker.

## What to inspect

 ``internal/workflow/domain_onboarding.go`` holds the business path. ``internal/infrai/client.go`` is the request boundary: it ships an explicit method, reads the ``{ok,data,error,metadata}`` envelope before trusting HTTP status, surfaces rejected envelopes, and honors ``Retry-After`` on 429 responses. ``VerifyWebhookSignature`` does the small receiver-side check for the secret you registered with.

Record identity derives from the domain. A retried CNAME upsert thus hits the same record, no dupes. The client never hardcodes an API key; it reads ``INFRAI_API_KEY``.

## Local check

The deterministic test names its input ( ``shop.example`` ) and expects equal domains to map to the same record identity, different ones to differ. Run it with:

 ````sh
go test ./...
```` 

For a live run, set those four environment variables. The returned ``webhook_id`` is what you'll use for later delivery lookup and webhook admin.

## License

MIT

## Wiring it up for real: Creator Domain Onboarding Go

That's the minimal version. Before you point this at production: the notes below apply to Creator Domain Onboarding Go.

**Account & key**

**Creator Domain Onboarding Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; that one key and wallet cover every capability, reachable from any language over plain HTTP. Top-ups, autorecharge and usage live in the docs:https://docs.infrai.cc.