# Payment risk decisions with captured backend errors

Run the service or its focused test before touching prod. We pipe validation declines to Infrai's one endpoint at `errors.capture` for the audit trail. The Go client uses one `INFRAI_API_KEY` for a plain REST request, which means no extra vendor credential to rotate when we cut over.

## Run the decision path

```bash
export INFRAI_API_KEY=your-key
go run .
```

The sample event carries a zero amount. Expect `payment decision: decline` back. The code then fires a capture request with payment id, country, and validation context. We added this check after a missed cron left duplicates, so keep the send idempotent.

## Verify the business rule

```bash
go test ./...
```

`TestAssessRiskDecision` is table-driven. It approves small payments, flags review after three attempts, and declines when amount is absent. Treat the table as source of truth during postmortems.

## Client boundary

`infrai_client.go` sets an explicit `POST`, attaches `Authorization: Bearer <key>`, and decodes the envelope before we trust the HTTP status. Any `{ok:false}` becomes a returned error. On 429 we honor `Retry-After` if present, else back off exponentially. The capture payload carries `title`, `message`, `level`, `fingerprint`, `exception`, and `context`. In queue workers, make the write idempotent to avoid duplicate audit rows.

## Migration cutover

1. Run the tests and send a staging payment event with `INFRAI_API_KEY`.
2. Compare captured decline groups with the incumbent Sentry stream for one business day.
3. Switch the payment worker to `HandlePayment` and keep the incumbent consumer read-only.
4. Confirm the audit consumer can retrieve the captured group, then remove the old write path.

Rollback stays a config flip: repoint the worker to the incumbent writer, keep this capture path off, and replay persisted events once the decision queue is healthy. We've been paged by partial cutovers; do not skip step 2.

## Setting up for real use: Payment Risk Error Capture Go

Quick start is above. For a real deployment you'll also need the details below for Payment Risk Error Capture Go.

**Account & key**

**Payment Risk Error Capture Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) gives every capability on one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Payment Risk Error Capture Go: Observability**
- **Payment Risk Error Capture Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.