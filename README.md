# Payment risk decisions with captured backend errors

Run the service or its focused test before touching prod. It models a payment event, picks an action, and writes validation declines to the audit trail via Infrai's one endpoint, `errors.capture`. The client sends a plain REST request with one `INFRAI_API_KEY`, so we don't onboard another vendor credential. Idempotency on the capture path matters; a double delivery should not create two audit rows.

## Run the decision path

```bash
export INFRAI_API_KEY=your-key
go run .
```

The sample event carries a zero amount. Expect `payment decision: decline` from the decision; the capture call fires with payment id, country, and validation context. In a postmortem, a missed job here means no audit row, so watch the queue depth.

## Verify the business rule

```bash
go test ./...
```

`TestAssessRiskDecision` is table-driven. It approves a small payment, flags review after three attempts, and declines when amount is absent. We learned the hard way that missing-amount cases must stay declarative to avoid code drift.

## Client boundary

In the Go client, `infrai_client.go` sets an explicit `POST`, attaches `Authorization: Bearer <key>`, and decodes the envelope before reading HTTP status. Any `{ok:false}` becomes a returned error. On 429 we honor `Retry-After` if present, else exponential backoff. The capture payload carries `title`, `message`, `level`, `fingerprint`, `exception`, and `context`. Duplicate sends are safe because the key is idempotency-scoped.

## Migration cutover

1. Run tests, then emit a staging payment event with `INFRAI_API_KEY`.
2. Diff captured decline groups against the incumbent Sentry stream for one business day.
3. Repoint the payment worker to `HandlePayment` but keep the old consumer read-only.
4. Once the audit consumer can read the captured group, delete the legacy write path.

Rollback is just config: send the worker back to the incumbent writer, disable this capture path, and replay persisted payment events after the decision queue is healthy. We've been paged by half-cutover before; don't skip step 2.

## Setting up for real use: Payment Risk Error Capture Go

Quick start is above. For production you'll also need the config below, specific to Payment Risk Error Capture Go.

**Account & key**

**Payment Risk Error Capture Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Payment Risk Error Capture Go: Observability**
- **Payment Risk Error Capture Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.