# Email Microservice

Kafka consumer responsible for all asynchronous, transactional email in the system. It never receives HTTP traffic from clients — its only input is Kafka events published by `auth`, and its only output is SMTP.

---

## Features

- **Asynchronous, at-least-once-safe processing**: consumes from Kafka with manual offset commits, only marking a message as processed after the email is durably logged — a crash mid-send is retried, never silently dropped.
- **Exactly-once delivery semantics on top of at-least-once transport**: every email is deduplicated by event ID against a unique DB constraint (see [Idempotency](#idempotency)).
- **Auditable send log**: every email attempt is recorded with a `status` (`pending` / `sent` / `failed`) and `type` (`welcome` / `stress`) — this isn't fire-and-forget, it's a queryable record.
- **HTML templating** for consistent, professional email formatting.
- **Bounded concurrency**: a semaphore-backed worker pool processes messages from each poll batch in parallel without unbounded goroutine growth.
- **Rebalance-safe consumer**: uses Kafka's cooperative-sticky balancer and explicitly commits offsets before a partition is revoked, so a rebalance mid-processing never causes reprocessing storms or a stuck shutdown.
- **Built-in load testing consumer**: a second consumer (`stress_email`) exercises the same send path under synthetic load from `auth`'s `/api/stress` endpoint, and is what KEDA scales this service on (Kafka consumer lag — see the root README).

---

## Tech Stack

- **Language**: Go 1.25
- **Messaging**: Kafka via [`twmb/franz-go`](https://github.com/twmb/franz-go)
- **ORM**: [GORM](https://gorm.io/) (MySQL, PostgreSQL, or SQLite)
- **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Testing**: [Testify](https://github.com/stretchr/testify) + [Testcontainers](https://testcontainers.com/) (real MySQL + Kafka)

---

## Folder Structure

> For the general architecture patterns used here — the layered `handlers/actions/model` structure, the Repository Pattern, dependency injection, and typed config — see the **[Architecture section of the root README](../../README.md#architecture)**. This section covers only what's specific to `email`.

```text
internal/
├── bootstrap/
│   └── consumer.go       # Kafka client setup, signal handling, ordered graceful shutdown
├── domain/
│   ├── email/
│   │   ├── model/         # Email entity (status/type enums) + repository
│   │   ├── actions/         # SendWelcome, SendStress — the actual send + idempotency logic
│   │   └── handlers/         # Kafka message → DTO → action (thin adapter layer)
│   └── health/
├── infrastructure/
│   ├── providers/messaging/    # Kafka consumer: balancer, offset commits, worker pool
│   ├── config/
│   └── database/
└── internal/shared/                # go-app-shared submodule (Kafka DTOs, routing keys)
```

This is what lets the `stress_email` handler reuse the exact same `SendWelcome`-style action as the real welcome-email path — the load test genuinely exercises production code, not a stub.

---

## Idempotency

Kafka guarantees **at-least-once** delivery — a consumer restart, rebalance, or retry can redeliver the same message. This service turns that into effectively-once behavior at the database layer:

1. Each consumed message is identified by a stable event ID (`topic:partition:offset`).
2. `SendWelcome.Execute` inserts the `Email` row with `EventID` as a **unique-indexed column** *before* attempting SMTP.
3. If that insert collides (the event was already processed), the driver returns a typed MySQL duplicate-key error (`*mysql.MySQLError`, code `1062`), which the action checks explicitly via `errors.As` — not by string-matching the error message.
4. Retries are judged by **row status**, not row existence: only a row already marked `status = sent` is treated as "already handled." A `pending` or `failed` row from a prior SMTP hiccup is retried, so a transient SMTP outage doesn't permanently block a legitimate email.

---

## Consumers

| Consumer group | Topic | Payload | Action |
|---|---|---|---|
| `email.service` | `user.created` | `WelcomeEmail` (email, name, verification URL) | Renders `welcome_user.html`, sends via SMTP, logs as `type: welcome` |
| `email.service` | `stress.test` | `StressEmail` | Same send path as above, logged as `type: stress` — this is the traffic KEDA scales on |

---

## Messaging — Consuming a New Event

To add a new consumer without touching any messaging infrastructure:

**1. Add the DTO** to the shared module (`internal/shared/messaging/kafka/dtos/`):
```go
type PasswordReset struct {
    Email string `json:"email"`
    Token string `json:"token"`
}
```

**2. Create the action** in `internal/domain/email/actions/`, following the same idempotency pattern as `send_welcome.go` (insert-then-check-duplicate on a unique event ID column).

**3. Create the handler** in `internal/domain/email/handlers/`:
```go
func (h *PasswordReset) Handle(body []byte, eventID string) error {
    var dto dtos.PasswordReset
    if err := json.Unmarshal(body, &dto); err != nil {
        return fmt.Errorf("failed to unmarshal password reset dto: %w", err)
    }
    return h.action.Execute(dto.Email, dto.Token, eventID)
}
```

**4. Register it** in `internal/bootstrap/consumer.go`:
```go
provider.Register("email.service", "", "", "user.password_reset", handlers.NewPasswordReset(passwordResetAction))
```

---

## Shutdown Behavior

On `SIGTERM`/`SIGINT`, the consumer:
1. Cancels the polling context (stops fetching new batches).
2. Waits for in-flight workers to finish the batch they already picked up.
3. Commits final offsets and closes the Kafka client.
4. Closes the database connection.

This ordering matters: closing the DB or Kafka client while a worker is mid-send would either drop a message that was actually processed, or crash the handler.

---

## Environment Variables

| Variable | Description |
|---|---|
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USER` / `SMTP_PASS` | SMTP credentials (Mailpit in local/dev) |
| `KAFKA_BROKERS` | Kafka bootstrap servers |
| `DB_DRIVER` | Database driver for the email log |
| `LOG_LEVEL` | `debug` \| `info` \| `warn` \| `error` |

---

## Getting Started

```bash
go run cmd/consumer/main.go
```

Or from the repo root: `make up`, `make migrate`, `make test` (see the [root README](../../README.md)). Tests use Testcontainers to spin up real, ephemeral MySQL and Kafka instances — no mocked broker.
