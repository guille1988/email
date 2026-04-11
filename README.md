### Email Microservice in Go

This is a specialized email microservice built with **Go**, designed to handle asynchronous email dispatching using **Kafka**. It follows clean architecture principles and integrates with SMTP services to send transactional emails like welcome messages, notifications, and more.

---

### 🚀 Features

*   **Asynchronous Processing**: Consumes email tasks from Kafka topics to ensure high availability and non-blocking operations.
*   **Email Templates**: Uses HTML templates for consistent and professional email formatting.
*   **SMTP Integration**: Securely sends emails via standard SMTP protocols.
*   **Email Logging**: Keeps a record of sent emails in the database for auditing and retry logic.
*   **Clean Architecture**: Strict separation of concerns (domain, infrastructure, and application layers).
*   **Containerized**: Fully Dockerized for seamless integration with the microservices ecosystem.
*   **Database Migrations**: Built-in tools for managing database schema for email logs.
*   **Testing Suite**: Includes integration tests using Testcontainers for MySQL and Kafka.

---

### 🛠 Tech Stack

*   **Language**: Go 1.25+
*   **Messaging**: [Kafka (twmb/franz-go)](https://github.com/twmb/franz-go)
*   **Email Client**: [Go-Mail](https://github.com/go-mail/mail)
*   **ORM**: [GORM](https://gorm.io/) (MySQL, PostgreSQL, SQLite)
*   **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
*   **Testing**: [Testify](https://github.com/stretchr/testify) & [Testcontainers](https://testcontainers.com/)

---

### 📋 Prerequisites

*   [Docker](https://www.docker.com/) and Docker Compose.
*   [Go](https://golang.org/) (optional, for local development).
*   `make` (utility to run Makefile commands from the root).

---

### ⚙️ Getting Started

1.  **Clone the repository** (if not done yet):
    ```bash
    git clone <repository-url>
    cd email
    ```

2.  **Environment Setup**:
    Ensure the `.env` file is configured with your SMTP credentials and Kafka broker address.

3.  **Run the Consumer**:
    This service primarily operates as a Kafka consumer.
    ```bash
    go run cmd/consumer/main.go
    ```

---

### 🛠 Development Commands

From the root `Makefile`, you can manage this service:

| Command | Description |
| :--- | :--- |
| `make up` | Start all infrastructure including Kafka and MySQL. |
| `make migrate` | Run database migrations for the email service. |
| `make compile` | Compile the email consumer binary. |
| `make test` | Run tests for the email microservice. |

---

### 📩 Message Consumers

#### Welcome Email (`email.service`)
*   **Topic**: `user.created`
*   **Payload**: `WelcomeEmail` (contains user email, name, verification URL)
*   **Action**: Renders `welcome_user.html` template and sends it to the recipient.

---

### 📨 Messaging — Consuming a new message

To consume a new message from Kafka, follow these 4 steps without touching any messaging infrastructure files:

**1. Create the DTO** in `internal/shared/messaging/kafka/dtos/`:
```go
// internal/shared/messaging/kafka/dtos/password_reset.go
type PasswordReset struct {
    Email string `json:"email"`
    Token string `json:"token"`
}
```

**2. Create the action** in `internal/domain/email/actions/`:
```go
// internal/domain/email/actions/send_password_reset.go
func (a *SendPasswordReset) Execute(email, token string) error { ... }
```

**3. Create the handler** in `internal/domain/email/handlers/`:
```go
// internal/domain/email/handlers/password_reset.go
func (h *PasswordReset) Handle(body []byte) error {
    var dto dtos.PasswordReset
    if err := json.Unmarshal(body, &dto); err != nil {
        return fmt.Errorf("failed to unmarshal password reset dto: %w", err)
    }
    return h.action.Execute(dto.Email, dto.Token)
}
```

**4. Register the handler** in `internal/bootstrap/consumer.go`:
```go
provider.Register(
    "email.service",
    "",
    "",
    "user.password_reset",
    handlers.NewPasswordReset(passwordResetAction),
)
```

No infrastructure files need to be modified.

---

### 📂 Project Structure

```text
├── cmd/                # Entry points (Consumer, API, Migrations)
├── internal/
│   ├── bootstrap/      # App initialization logic (Kafka, DB)
│   ├── domain/         # Business logic (Email module)
│   │   └── email/      # Email actions, handlers, templates, entities
│   ├── infrastructure/ # Frameworks & Drivers (DB, Kafka, Logger)
│   ├── shared/         # Shared DTOs for messaging
├── tests/              # Integration and Unit tests
├── Dockerfile          # Production build configuration
└── go.mod              # Dependencies
```

---

### 🔐 Environment Variables

Key configurations:
*   `SMTP_HOST`: SMTP server address.
*   `SMTP_PORT`: SMTP server port.
*   `SMTP_USER`: Authentication user.
*   `SMTP_PASS`: Authentication password.
*   `KAFKA_BROKERS`: Kafka broker address (e.g. `kafka:9092`).
*   `DB_DRIVER`: Database driver for logging emails.

---

### 🧪 Testing

Run tests using the project root Makefile:
```bash
make test
```
The tests use Testcontainers to spin up ephemeral MySQL and Kafka instances.
