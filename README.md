# Linghub — AI Chat Backend

A production-ready Go web backend for an AI-powered chat application. Built with **Gin** (HTTP framework), **GORM** (database ORM), **Eino** (LLM agent orchestration), and **MySQL**. Features user authentication (JWT), multi-turn conversation history, and real-time SSE streaming of AI responses.

## ✨ Features

- **User Authentication**: Registration and login with bcrypt password hashing and JWT tokens (7-day default expiry)
- **Conversations**: Create and manage multi-turn chat conversations per user
- **AI Chat**: Integrate any OpenAI-compatible LLM (OpenAI, Llama, Ollama, etc.) via Eino framework
- **Real-time Streaming**: SSE endpoints deliver AI responses incrementally for a responsive user experience
- **Data Persistence**: All messages and conversation history stored in MySQL via GORM
- **Access Control**: JWT middleware ensures users can only access their own conversations

## 🚀 Quick Start

### Prerequisites

- **Go 1.26+** (or adjust `go.mod`)
- **MySQL 5.7+** (create a database named `linghub` or set `DB_DSN` to point to your database)
- **OpenAI-compatible LLM API** (OpenAI, Ollama, local proxy, etc.)

### Setup

1. **Clone and prepare environment**:
   ```bash
   cd linghub
   cp .env.example .env
   ```

2. **Edit `.env` with your configuration**:
   ```bash
   PORT=8080
   DB_DSN=root:password@tcp(127.0.0.1:3306)/linghub?charset=utf8mb4&parseTime=True&loc=Local
   JWT_SECRET=your-secret-key-change-in-production
   JWT_EXPIRATION_HOURS=168
   OPENAI_API_KEY=sk-...
   OPENAI_BASE_URL=https://api.openai.com/v1
   OPENAI_MODEL=gpt-4o-mini
   SYSTEM_PROMPT=你是一个有帮助的中文AI助手。
   ```

3. **Install dependencies and build**:
   ```bash
   go mod tidy
   go build -o linghub .
   ```

4. **Run the server**:
   ```bash
   ./linghub
   # Server starts on http://localhost:8080
   ```

## 📖 API Documentation

### Authentication Endpoints (Public)

#### Register
```
POST /api/auth/register
Content-Type: application/json

{
  "username": "john",
  "password": "secure_password_123"
}

Response (201 Created):
{
  "user_id": 1,
  "token": "eyJhbGc..."
}
```

#### Login
```
POST /api/auth/login
Content-Type: application/json

{
  "username": "john",
  "password": "secure_password_123"
}

Response (200 OK):
{
  "user_id": 1,
  "token": "eyJhbGc..."
}
```

### Conversation Endpoints (Protected)

All protected endpoints require an `Authorization: Bearer <token>` header.

#### Create Conversation
```
POST /api/conversations
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "My First Chat"
}

Response (201 Created):
{
  "id": 1,
  "title": "My First Chat"
}
```

#### List Conversations
```
GET /api/conversations
Authorization: Bearer <token>

Response (200 OK):
[
  {
    "ID": 1,
    "UserID": 1,
    "Title": "My First Chat",
    "CreatedAt": "2024-10-10T12:00:00Z"
  }
]
```

#### Get Conversation Messages
```
GET /api/conversations/1/messages
Authorization: Bearer <token>

Response (200 OK):
[
  {
    "ID": 1,
    "ConversationID": 1,
    "Role": "user",
    "Content": "Hello, how are you?",
    "CreatedAt": "2024-10-10T12:01:00Z"
  },
  {
    "ID": 2,
    "ConversationID": 1,
    "Role": "assistant",
    "Content": "I'm doing well, thank you for asking!",
    "CreatedAt": "2024-10-10T12:01:05Z"
  }
]
```

#### Send Message (SSE Stream)
```
POST /api/conversations/1/messages
Authorization: Bearer <token>
Content-Type: application/json

{
  "message": "Tell me a joke"
}

Response (200 OK, text/event-stream):
data: {"delta":"Why"}
data: {"delta":" did"}
data: {"delta":" the"}
data: {"delta":" chicken"}
data: {"delta":" cross"}
data: {"delta":" the"}
data: {"delta":" road"}
data: {"delta":"?"}
data: [DONE]
```

## 🏗️ Architecture

### Directory Structure

```
linghub/
├── main.go                          # Entry point
├── .env.example                     # Environment template
├── go.mod / go.sum                  # Go module files
└── internal/
    ├── config/
    │   └── config.go               # Configuration management
    ├── model/
    │   ├── user.go                 # User model
    │   ├── conversation.go         # Conversation model
    │   └── message.go              # Message model
    ├── database/
    │   └── database.go             # Database initialization
    ├── jwtutil/
    │   └── jwt.go                  # JWT token handling
    ├── middleware/
    │   └── auth.go                 # Authentication middleware
    ├── service/
    │   ├── auth_service.go         # Authentication logic
    │   └── chat_service.go         # Chat and AI orchestration
    ├── handler/
    │   ├── auth_handler.go         # Auth HTTP handlers
    │   ├── conversation_handler.go # Conversation HTTP handlers
    │   └── chat_handler.go         # Chat HTTP handlers (SSE)
    └── router/
        └── router.go               # Route configuration
```

### Layered Design

```
HTTP Handlers (Gin)
       ↓
Business Services (Auth, Chat)
       ↓
Models + Database (GORM, MySQL)
```

- **Handlers**: Accept HTTP requests, validate input, call services, return responses
- **Services**: Implement business logic (auth, conversation management, LLM integration)
- **Models**: GORM-managed database entities
- **Middleware**: JWT authentication and authorization checks

### Data Models

#### User
```go
type User struct {
  ID           uint
  Username     string (unique)
  PasswordHash string
  CreatedAt    time.Time
  UpdatedAt    time.Time
}
```

#### Conversation
```go
type Conversation struct {
  ID        uint
  UserID    uint (foreign key)
  Title     string
  CreatedAt time.Time
  UpdatedAt time.Time
}
```

#### Message
```go
type Message struct {
  ID             uint
  ConversationID uint (foreign key)
  Role           string ("user" or "assistant")
  Content        string
  CreatedAt      time.Time
}
```

## 🔐 Security Considerations

- **Passwords**: Hashed with bcrypt (cost 10, standard)
- **Tokens**: HS256-signed JWT with 7-day default expiration
- **Database**: Use strong `DB_DSN` credentials; never commit real credentials to git
- **Secrets**: Store `JWT_SECRET` and `OPENAI_API_KEY` securely (environment variables, secrets manager, etc.)
- **HTTPS**: In production, deploy behind a reverse proxy (nginx, cloud load balancer) that handles TLS

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/jwtutil
go test ./internal/service
```

**Note**: Integration tests (real database, real LLM) require external services and are not included in the unit test suite.

## 📝 Development Workflow

```bash
# Format code
gofmt -w .

# Run linter
go vet ./...

# Run tests
go test ./...

# Build
go build -o linghub .

# Run
./linghub
```

## 🤝 Integration with External Services

### OpenAI-Compatible LLMs

The chat service uses **Eino's OpenAI component** to call any API compatible with the OpenAI Chat Completion format:

- **OpenAI** (official): Set `OPENAI_API_KEY` to your API key
- **Ollama** (local): Set `OPENAI_BASE_URL` to `http://localhost:11434/v1` and use any available model
- **Azure OpenAI**: Set `OPENAI_BASE_URL` to your Azure endpoint
- **Third-party proxies**: Point `OPENAI_BASE_URL` to your proxy URL

### Database

Any MySQL-compatible database works (MySQL 5.7+, MariaDB, etc.). Configure via `DB_DSN` in `.env`.

## 🚢 Deployment

### Docker (Example)

```dockerfile
FROM golang:1.26 AS builder
WORKDIR /app
COPY . .
RUN go build -o linghub .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/linghub .
COPY .env.example .env
EXPOSE 8080
CMD ["./linghub"]
```

### Environment Checklist

- [ ] MySQL database created and accessible
- [ ] `DB_DSN` set to production database
- [ ] `JWT_SECRET` is a strong random string
- [ ] `OPENAI_API_KEY` and `OPENAI_BASE_URL` configured
- [ ] `PORT` set appropriately for your infrastructure
- [ ] HTTPS enabled (via reverse proxy)
- [ ] Database backups configured
- [ ] Error logging configured

## 📚 Documentation

- [Eino Framework](https://github.com/cloudwego/eino) — LLM application development kit
- [Gin Web Framework](https://github.com/gin-gonic/gin) — HTTP framework
- [GORM](https://gorm.io/) — Go ORM
- [OpenAI API](https://platform.openai.com/docs/api-reference) — LLM API reference

## 📄 License

MIT License (or specify your license)

## 👥 Contributing

Contributions welcome! Please open an issue or pull request.

## ❓ FAQ

**Q: Can I use a different LLM provider?**  
A: Yes! Set `OPENAI_BASE_URL` to any OpenAI-compatible endpoint (Ollama, Mistral, etc.).

**Q: How do I change the JWT expiration?**  
A: Set `JWT_EXPIRATION_HOURS` in `.env` (default: 168 hours = 7 days).

**Q: Can I run this without MySQL?**  
A: The code uses GORM with MySQL driver. To use a different database, edit `internal/database/database.go`.

**Q: Does this support multiple languages?**  
A: Yes! The `SYSTEM_PROMPT` is fully configurable. Adjust it to any language or style.

**Q: Is this production-ready?**  
A: Yes, the code follows Go best practices and is structured for scalability. Ensure external services (database, LLM) are production-grade and secure.
