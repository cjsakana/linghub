# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project State

**Linghub** is a Go web backend for an AI chat application. MVP stage: users register/login, create conversations, send messages to a large language model (via OpenAI-compatible API), and receive SSE-streamed responses. Messages are stored in MySQL; agent orchestration is handled by CloudWeGo's Eino framework.

Stack: Gin (HTTP), GORM (ORM), Eino (LLM/agent), MySQL, JWT (auth), bcrypt (passwords).

## Directory Structure

```
linghub/
  main.go                          # Entry point: loads config, inits DB, wires dependencies, starts Gin
  .env.example                     # Environment variable template (rename to .env for local dev)
  go.mod / go.sum                  # Go module manifest and lock file
  internal/
    config/
      config.go                   # Reads environment variables, centralizes configuration
    model/
      user.go                     # GORM model: User (username, password hash)
      conversation.go             # GORM model: Conversation (user_id, title)
      message.go                  # GORM model: Message (conversation_id, role, content)
    database/
      database.go                 # gorm.Open + AutoMigrate
    jwtutil/
      jwt.go                      # Token generation/validation (HS256, 7-day expiry default)
    middleware/
      auth.go                     # JWT Bearer token middleware; extracts user_id into gin.Context
    service/
      auth_service.go            # Register/login logic (bcrypt + JWT generation)
      chat_service.go            # Eino ChatModel + ChatModelAgent + Runner; orchestrates AI chat, saves messages
    handler/
      auth_handler.go            # POST /api/auth/register, /api/auth/login
      conversation_handler.go    # POST /conversations, GET /conversations, GET /conversations/:id/messages
      chat_handler.go            # POST /conversations/:id/messages (SSE stream endpoint)
    router/
      router.go                  # Gin route setup; public auth routes + protected chat routes
```

## Layered Architecture

**Handler** (Gin HTTP layer) → **Service** (business logic: auth, chat) → **Model + Database** (GORM ORM, MySQL schema).
- No dependency injection framework; dependencies are wired manually in `main.go`.
- No middleware beyond JWT auth; Gin defaults handle CORS, panic recovery, etc.

## Commands

### Local Development Setup

```bash
# 1. Copy environment template and edit with real values
cp .env.example .env
# Edit .env: set DB_DSN, OPENAI_API_KEY, OPENAI_BASE_URL, JWT_SECRET

# 2. Ensure go.mod has all dependencies (already done, but can re-sync if needed)
go mod tidy

# 3. Build the project
go build -o linghub .

# 4. Run the server (requires .env and MySQL accessible at DB_DSN)
./linghub
# Server listens on http://localhost:8080 by default; set PORT env var to change
```

### Testing

```bash
# Run all tests (currently: JWT token gen/validation, bcrypt password hashing)
go test ./...

# Run tests for a specific package
go test ./internal/jwtutil
go test ./internal/service
```

### Code Quality

```bash
# Lint / static analysis
go vet ./...

# Format code
gofmt -w .

# (Optional: run golangci-lint if installed)
# golangci-lint run ./...
```

## Key Features

- **User Auth**: Registration (bcrypt), login, JWT token generation (7-day default expiry)
- **Conversations**: Create, list, persist multi-turn history
- **AI Chat**: Eino ChatModelAgent wraps OpenAI-compatible API; supports any provider (local, 3rd-party proxy, Azure)
- **Streaming**: SSE endpoint (`POST /api/conversations/:id/messages`) returns AI response increments in real-time
- **Database**: MySQL with GORM; three tables (users, conversations, messages); auto-migrated on startup

## API Endpoints

### Public (no auth)
- `POST /api/auth/register` — `{username, password}` → `{user_id, token}`
- `POST /api/auth/login` — `{username, password}` → `{user_id, token}`

### Protected (Bearer token required in `Authorization` header)
- `POST /api/conversations` — `{title}` → creates conversation
- `GET /api/conversations` — list user's conversations
- `GET /api/conversations/:id/messages` — list messages in a conversation
- `POST /api/conversations/:id/messages` — `{message}` → SSE stream of AI response

All protected endpoints validate `conversation.user_id == authenticated_user_id` to prevent cross-user access.

## Environment Variables

See `.env.example` for a full template. Key ones:
- `PORT` (default 8080)
- `DB_DSN` — MySQL connection string (e.g., `root:pass@tcp(localhost:3306)/linghub?charset=utf8mb4&parseTime=True`)
- `JWT_SECRET` — signing key for tokens; use a strong random string
- `JWT_EXPIRATION_HOURS` (default 168 = 7 days)
- `OPENAI_API_KEY` — your LLM provider API key
- `OPENAI_BASE_URL` — defaults to `https://api.openai.com/v1`; can be any OpenAI-compatible endpoint
- `OPENAI_MODEL` — e.g., `gpt-4o-mini`, `gpt-3.5-turbo`, etc.
- `SYSTEM_PROMPT` — system instruction for the AI (default Chinese-friendly)

## Verification & Next Steps

✅ **Completed**:
- `go build` — compiles without error
- `go vet` — passes static analysis
- Code structured in standard Go patterns (handler/service/model)
- All 14 Go files present and accounted for

⚠️ **Not Yet Tested** (requires external services):
- End-to-end auth flow (requires MySQL)
- Full chat with AI (requires OpenAI-compatible API + MySQL)

To test locally, you'll need:
1. A running MySQL server with a database named `linghub` (or adjust `DB_DSN`)
2. Valid `OPENAI_API_KEY` and `OPENAI_BASE_URL`
3. Then: `cp .env.example .env`, edit, and `./linghub` or `go run main.go`
4. Use `curl`, Postman, or a frontend to call the endpoints

**Note**: The project is production-ready in terms of structure and code quality, but no real data has flowed through it yet. First real test should be:
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password123"}'
```

