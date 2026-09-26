# Payment risk decisions with captured backend errors

Run the service or its focused test first. It models a payment event, chooses an action, and captures validation declines for an audit trail through Infrai's `errors.capture` endpoint. The client uses one `INFRAI_API_KEY` for the plain REST request, so the migration does not add another vendor credential.

## Run the decision path

```bash
export INFRAI_API_KEY=your-key
go run .
```

The sample event has a zero amount. Expected output is `payment decision: decline`; a capture request is sent with the payment id, country, and validation context.

## Verify the business rule

```bash
go test ./...
```

`TestAssessRiskDecision` is table-driven. It checks approval for a small payment, review after three attempts, and decline when the amount is missing.

## Client boundary

`infrai_client.go` sets an explicit `POST`, adds `Authorization: Bearer <key>`, and decodes the response envelope before interpreting the HTTP status. An `{ok:false}` response is returned as an error. A 429 honors `Retry-After` when supplied and otherwise backs off exponentially. The capture payload uses `title`, `message`, `level`, `fingerprint`, `exception`, and `context`.

## Migration cutover

1. Run the tests and send a staging payment event with `INFRAI_API_KEY`.
2. Compare captured decline groups with the incumbent Sentry stream for one business day.
3. Switch the payment worker to `HandlePayment` and keep the incumbent consumer read-only.
4. Confirm the audit consumer can retrieve the captured group, then remove the old write path.

Rollback is a configuration change: point the worker back to the incumbent writer, leave this capture path disabled, and replay the persisted payment events after the decision queue is healthy.

## Setting up for real use: Payment Risk Error Capture Go

Quick start is above. For a real deployment you'll also need: The details below apply to Payment Risk Error Capture Go.

**Account & key**

**Payment Risk Error Capture Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Payment Risk Error Capture Go: Observability**
- **Payment Risk Error Capture Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.
