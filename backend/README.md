# Atlas - Backend

Go backend API for Atlas, a hierarchical knowledge management application with Google OAuth authentication and RESTful endpoints.

## 🛠️ Tech Stack

- **Go 1.21+** - Modern, performant programming language
- **Gin** - Fast HTTP web framework with middleware support
- **GORM** - Feature-rich ORM with PostgreSQL support
- **Goth/Gothic** - Multi-provider OAuth authentication library
- **PostgreSQL** - Robust relational database
- **Zerolog** - High-performance structured logging
- **godotenv** - Environment variable management

## ✨ Features

### 🔐 Authentication & Security
- **Google OAuth 2.0**: Secure user authentication
- **Session Management**: Cookie-based sessions with Gothic
- **CORS Configuration**: Secure cross-origin requests
- **Protected Routes**: Middleware-based route protection
- **User Context**: Automatic user extraction from session

### 📊 Data Models
- **Hierarchical Structure**: Users → Notebooks → Chapters → Notes
- **Cascade Operations**: Automatic cleanup of child entities
- **Timestamps**: CreatedAt and UpdatedAt for all models
- **Associations**: Proper foreign key relationships
- **Preloading**: Efficient nested data loading

### 🔄 API Features
- **RESTful Design**: Standard HTTP methods and status codes
- **JSON Responses**: Consistent response format
- **Error Handling**: Comprehensive error messages
- **Query Optimization**: Eager loading to prevent N+1 queries
- **Transaction Support**: ACID-compliant operations

## 📁 Project Structure

```
backend/
├── cmd/
│   └── main.go                 # Application entry point
│
├── db/
│   ├── db.go                   # Database connection setup
│   └── migrations/             # SQL migration files
│
├── internal/
│   ├── auth/
│   │   └── auth.go            # OAuth configuration
│   │
│   ├── controllers/
│   │   ├── notebook.controller.go  # Notebook HTTP handlers
│   │   ├── chapter.controller.go   # Chapter HTTP handlers
│   │   └── note.controller.go      # Note HTTP handlers
│   │
│   ├── models/
│   │   ├── user.model.go       # User entity
│   │   ├── notebook.model.go   # Notebook entity
│   │   ├── chapter.model.go    # Chapter entity
│   │   └── note.model.go       # Note entity
│   │
│   └── middleware/
│       └── auth.middleware.go  # Authentication middleware
│
├── .env                        # Environment variables (not in git)
├── .env.example                # Environment template
├── go.mod                      # Go module dependencies
└── go.sum                      # Dependency checksums
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
- Google OAuth credentials ([Get them here](https://console.cloud.google.com/))

### Installation

1. **Navigate to backend directory:**
   ```bash
   cd backend
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Setup environment variables:**
   ```bash
   cp .env.example .env
   ```

4. **Configure `.env` file:**
   ```env
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=notes_db
   
   # Server
   PORT=8080
   
   # Google OAuth
   GOOGLE_CLIENT_ID=your_client_id
   GOOGLE_CLIENT_SECRET=your_client_secret
   GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback
   
   # Session
   SESSION_SECRET=your_session_secret_key
   
   # Frontend URL (for CORS)
   FRONTEND_URL=http://localhost:5173
   ```

5. **Create database:**
   ```sql
   CREATE DATABASE notes_db;
   ```

6. **Run the application:**
   ```bash
   go run cmd/main.go
   ```

   Server will start on `http://localhost:8080`

### Development Mode

Run with hot reload (using air or similar):
```bash
air
```

## 🔌 API Endpoints

### Authentication Routes
```
GET  /auth/google              - Initiate Google OAuth flow
GET  /auth/google/callback     - OAuth callback handler
GET  /auth/user                - Get current user details
GET  /logout/google            - Logout and clear session
```

### Notebook Routes (Protected)
```
POST   /notebook               - Create new notebook
GET    /notebook/:id           - Get notebook with nested data
GET    /notebooks              - Get all user's notebooks
PUT    /notebook/:id           - Update notebook
DELETE /notebook/:id           - Delete notebook (cascade)
```

### Chapter Routes (Protected)
```
POST   /chapter                - Create new chapter
GET    /chapter/:id            - Get chapter with notes
PUT    /chapter/:id            - Update chapter
PUT    /chapter/:id/move       - Move chapter to different notebook
DELETE /chapter/:id            - Delete chapter (cascade)
```

### Note Routes (Protected)
```
POST   /note                   - Create new note
GET    /note/:id               - Get note by ID
PUT    /note/:id               - Update note content
PUT    /note/:id/move          - Move note to different chapter
DELETE /note/:id               - Delete note
```

## 📦 Database Models

### User Model
```go
type User struct {
    ID        uint        `gorm:"primaryKey"`
    Name      string      `gorm:"not null"`
    Email     string      `gorm:"uniqueIndex;not null"`
    ImageUrl  *string
    Notebooks []Notebook  `gorm:"constraint:OnDelete:CASCADE"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Notebook Model
```go
type Notebook struct {
    ID        uint       `gorm:"primaryKey"`
    Name      string     `gorm:"not null"`
    UserID    uint       `gorm:"not null;index"`
    User      User
    Chapters  []Chapter  `gorm:"constraint:OnDelete:CASCADE"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Chapter Model
```go
type Chapter struct {
    ID         uint       `gorm:"primaryKey"`
    Name       string     `gorm:"not null"`
    NotebookID uint       `gorm:"not null;index"`
    Notebook   Notebook
    Notes      []Note     `gorm:"constraint:OnDelete:CASCADE"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### Note Model
```go
type Note struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"not null"`
    Content   string    `gorm:"type:text"`
    ChapterID uint      `gorm:"not null;index"`
    Chapter   Chapter
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## 🔒 Authentication Flow

1. **User clicks "Login with Google"**
   - Frontend redirects to `/auth/google`

2. **OAuth Flow**
   - User authenticates with Google
   - Google redirects to callback URL
   - Backend creates/updates user in database
   - Session is created and stored

3. **Authenticated Requests**
   - Session cookie sent with each request
   - Middleware validates session
   - User context injected into request

4. **Logout**
   - Session is destroyed
   - Cookie is cleared

## 🛡️ Middleware

### Authentication Middleware
```go
// Protects routes by checking session
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        c.Set("user", user)
        c.Next()
    }
}
```

### CORS Middleware
```go
// Allows cross-origin requests from frontend
config := cors.DefaultConfig()
config.AllowOrigins = []string{os.Getenv("FRONTEND_URL")}
config.AllowCredentials = true
router.Use(cors.New(config))
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `secret` |
| `DB_NAME` | Database name | `notes_db` |
| `PORT` | Server port | `8080` |
| `GOOGLE_CLIENT_ID` | OAuth client ID | `xxx.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | OAuth secret | `GOCSPX-xxx` |
| `GOOGLE_CALLBACK_URL` | OAuth callback | `http://localhost:8080/auth/google/callback` |
| `SESSION_SECRET` | Session encryption key | `random-secret-key` |
| `FRONTEND_URL` | Frontend URL for CORS | `http://localhost:5173` |

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add authorized redirect URIs:
   - Development: `http://localhost:8080/auth/google/callback`
   - Production: `https://yourdomain.com/auth/google/callback`
6. Copy credentials to `.env`

## 🗄️ Database Operations

### Auto-Migration
```go
// Automatically creates/updates tables
db.AutoMigrate(&models.User{}, &models.Notebook{}, 
               &models.Chapter{}, &models.Note{})
```

### Cascade Deletes
```go
// Deleting a notebook automatically deletes its chapters and notes
db.Delete(&notebook)
```

### Eager Loading
```go
// Load nested relationships in one query
db.Preload("Chapters.Notes").Find(&notebook)
```

## 📊 Query Optimization

### N+1 Prevention
```go
// BAD: N+1 queries
for _, notebook := range notebooks {
    db.Find(&notebook.Chapters)  // Extra query per notebook
}

// GOOD: Single query with preload
db.Preload("Chapters").Find(&notebooks)
```

### Indexing Strategy
- Foreign keys are indexed
- Email is unique indexed
- Composite indexes where needed

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...

# Run specific package tests
go test ./internal/controllers/...
```

## 🏗️ Building

### Development Build
```bash
go build -o bin/server cmd/main.go
```

### Production Build
```bash
go build -ldflags="-s -w" -o bin/server cmd/main.go
```

The `-ldflags="-s -w"` flag strips debug information for smaller binary.

## 🚢 Deployment

### Using Docker
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o server cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

### Environment-based Config
Use different `.env` files for different environments:
- `.env.development`
- `.env.staging`
- `.env.production`

## 🔐 Security Best Practices

⚠️ **Before Production Deployment:**

- [ ] Use HTTPS everywhere
- [ ] Generate cryptographically secure session secrets
- [ ] Implement Redis/PostgreSQL session storage (not cookie store)
- [ ] Add CSRF protection
- [ ] Implement rate limiting
- [ ] Add request validation and sanitization
- [ ] Enable security headers (HSTS, CSP, etc.)
- [ ] Implement audit logging
- [ ] Use prepared statements (GORM does this)
- [ ] Add SQL injection protection (GORM provides this)
- [ ] Implement proper error handling (don't expose internal errors)

## 📈 Performance Tips

1. **Connection Pooling**: Configure GORM connection pool
   ```go
   sqlDB, _ := db.DB()
   sqlDB.SetMaxIdleConns(10)
   sqlDB.SetMaxOpenConns(100)
   ```

2. **Query Optimization**: Use `Select` to fetch only needed fields
   ```go
   db.Select("id", "name").Find(&notebooks)
   ```

3. **Batch Operations**: Use batch inserts for multiple records
   ```go
   db.CreateInBatches(notes, 100)
   ```

## 🐛 Debugging

### Enable SQL Logging
```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```

### Structured Logging with Zerolog
```go
log.Info().
    Str("user_id", userID).
    Str("action", "create_notebook").
    Msg("Notebook created successfully")
```

## 📚 Resources

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Guide](https://gorm.io/docs/)
- [Goth OAuth](https://github.com/markbates/goth)
- [Go Best Practices](https://go.dev/doc/effective_go)

## 🤝 Contributing

When adding new features:
1. Follow Go idioms and conventions
2. Add proper error handling
3. Write tests for new functionality
4. Update API documentation
5. Use GORM best practices
6. Add structured logging
