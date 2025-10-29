# backend

A Go backend API built with Gin and GORM.

## Project Structure

```
backend/
├── cmd/api/              # Application entry point
├── internal/             # Private application code
│   ├── config/           # Configuration management
│   ├── models/           # Data models
│   ├── repositories/     # Database operations
│   ├── services/         # Business logic
│   ├── controllers/      # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   └── routes/           # Route definitions
├── pkg/utils/            # Shared utilities
├── db/migrations/        # Database migrations
├── .env                  # Environment variables (not in git)
└── .env.example          # Example environment variables
```

## Setup

1. Copy `.env.example` to `.env` and configure your environment variables:
   ```bash
   cp .env.example .env
   ```

2. Update the database credentials in `.env`

3. Run the application:
   ```bash
   go run cmd/main.go
   ```

## Dependencies

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [GORM](https://gorm.io/) - ORM library
- [godotenv](https://github.com/joho/godotenv) - Environment variable loader
- [JWT](https://github.com/golang-jwt/jwt) - JSON Web Token implementation

## Development

Run the application:
```bash
go run cmd/main.go
```

Run tests:
```bash
go test ./...
```

Build the application:
```bash
go build -o bin/api cmd/main.go
```
