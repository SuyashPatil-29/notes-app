# Notes App

A full-stack notes application with hierarchical organization (Users → Notebooks → Chapters → Notes) built with Go (Gin + GORM) backend and React (TypeScript + Vite) frontend.

## Features

- 🔐 **Google OAuth Authentication** - Secure login with Google
- 📚 **Hierarchical Organization** - Users → Notebooks (Drives) → Chapters (Folders) → Notes (Files)
- 🎯 **RESTful API** - Clean and well-structured endpoints
- 🔄 **Session Management** - Cookie-based authentication
- 🎨 **Modern Frontend** - React with TypeScript and Vite
- 🗄️ **PostgreSQL Database** - Reliable data storage with GORM

## Architecture

```
Users
  └── Notebooks (Drives)
       └── Chapters (Folders)
            └── Notes (Files)
```

## Tech Stack

### Backend
- **Go** - Programming language
- **Gin** - Web framework
- **GORM** - ORM for database operations
- **Goth/Gothic** - OAuth authentication
- **PostgreSQL** - Database
- **Zerolog** - Structured logging

### Frontend
- **React** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Axios** - HTTP client

## Project Structure

```
notes-app/
├── backend/
│   ├── cmd/
│   │   └── main.go                 # Application entry point
│   ├── db/
│   │   ├── db.go                   # Database connection
│   │   └── migrations/             # Database migrations
│   ├── internal/
│   │   ├── auth/
│   │   │   └── auth.go            # Authentication logic
│   │   ├── controllers/
│   │   │   ├── notebook.controller.go
│   │   │   └── chapter.controller.go
│   │   ├── models/
│   │   │   ├── user.model.go
│   │   │   ├── notebook.model.go
│   │   │   ├── chapter.model.go
│   │   │   └── notes.model.go
│   │   └── middleware/             # Custom middleware
│   ├── go.mod
│   └── .env.example
│
└── frontend/
    ├── src/
    │   ├── components/             # React components
    │   ├── hooks/
    │   │   └── auth.ts            # Authentication hooks
    │   ├── types/
    │   │   └── backend.ts         # TypeScript types
    │   ├── utils/
    │   │   ├── api.ts             # Axios configuration
    │   │   ├── auth.ts            # Auth utilities
    │   │   ├── notebook.ts        # Notebook API calls
    │   │   └── chapter.ts         # Chapter API calls
    │   ├── App.tsx
    │   └── main.tsx
    ├── package.json
    └── .env.example
```

## Getting Started

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18 or higher
- **PostgreSQL** 14 or higher
- **Google OAuth Credentials** ([Get them here](https://console.cloud.google.com/))

### Backend Setup

1. **Navigate to backend directory:**
   ```bash
   cd backend
   ```

2. **Copy environment variables:**
   ```bash
   cp .env.example .env
   ```

3. **Configure `.env` file:**
   - Add your PostgreSQL credentials
   - Add your Google OAuth credentials
   - Set a secure session secret

4. **Install dependencies:**
   ```bash
   go mod download
   ```

5. **Run the server:**
   ```bash
   go run cmd/main.go
   ```

   Server will start on `http://localhost:8080`

### Frontend Setup

1. **Navigate to frontend directory:**
   ```bash
   cd frontend
   ```

2. **Install dependencies:**
   ```bash
   bun install
   # or
   npm install
   ```

3. **Copy environment variables (if needed):**
   ```bash
   cp .env.example .env
   ```

4. **Start development server:**
   ```bash
   bun dev
   # or
   npm run dev
   ```

   Frontend will start on `http://localhost:5173`

## API Endpoints

### Authentication
- `GET /auth/google` - Start Google OAuth flow
- `GET /auth/google/callback` - OAuth callback
- `GET /auth/user` - Get current authenticated user
- `GET /logout/google` - Logout

### Notebooks
- `POST /notebook` - Create a notebook
- `GET /notebook/:id` - Get notebook by ID
- `PUT /notebook/:id` - Update notebook
- `DELETE /notebook/:id` - Delete notebook

### Chapters
- `POST /chapter` - Create a chapter
- `GET /chapter/:id` - Get chapter by ID
- `PUT /chapter/:id` - Update chapter
- `DELETE /chapter/:id` - Delete chapter

## Database Models

### User
```go
type User struct {
    ID        uint
    Name      string
    Email     string
    ImageUrl  *string
    Notebooks []Notebook
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Notebook
```go
type Notebook struct {
    ID        uint
    Name      string
    UserID    uint
    Chapters  []Chapter
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Chapter
```go
type Chapter struct {
    ID         uint
    Name       string
    NotebookID uint
    Notes      []Notes
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### Notes
```go
type Notes struct {
    ID        uint
    Name      string
    Content   string
    ChapterID uint
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## Environment Variables

See `.env.example` files in `backend/` and `frontend/` directories for required environment variables.

## Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or select existing)
3. Enable Google+ API
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client ID"
5. Configure OAuth consent screen
6. Add authorized redirect URIs:
   - `http://localhost:8080/auth/google/callback` (development)
   - `https://yourdomain.com/auth/google/callback` (production)
7. Copy Client ID and Client Secret to `.env` file

## Development

### Running Tests
```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
bun test
# or
npm test
```

### Building for Production

**Backend:**
```bash
cd backend
go build -o bin/server cmd/main.go
./bin/server
```

**Frontend:**
```bash
cd frontend
bun build
# or
npm run build
```

## Security Considerations

⚠️ **This application is currently in development mode**

Before deploying to production:
- [ ] Use HTTPS everywhere
- [ ] Generate cryptographically secure session secrets
- [ ] Implement Redis/PostgreSQL for session storage (not cookie store)
- [ ] Add CSRF protection
- [ ] Add rate limiting
- [ ] Implement session rotation
- [ ] Use environment-based secret management
- [ ] Enable security headers (HSTS, CSP, etc.)
- [ ] Add audit logging

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Goth](https://github.com/markbates/goth)
- [React](https://react.dev/)
- [Vite](https://vitejs.dev/)

