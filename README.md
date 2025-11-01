# Notes App

A full-stack notes application with hierarchical organization (Users → Notebooks → Chapters → Notes) built with Go (Gin + GORM) backend and React (TypeScript + Vite) frontend.

## ✨ Features

### 🔐 Authentication & Security
- **Google OAuth Integration** - Secure login with Google accounts
- **Session Management** - Cookie-based authentication with secure storage
- **Protected Routes** - Client-side route protection for authenticated users

### 📚 Hierarchical Organization
- **4-Level Structure** - Users → Notebooks → Chapters → Notes
- **Drag & Drop** - Move chapters between notebooks and notes between chapters
- **Collapsible Navigation** - Expandable/collapsible sidebar with smooth animations
- **Active State Tracking** - Visual indication of currently selected note

### ✏️ Rich Text Editing
- **TipTap Editor** - Powerful WYSIWYG editor with markdown support
- **Real-time Preview** - Instant rendering of formatted content
- **Auto-save** - Automatic saving with visual feedback
- **Formatting Tools** - Bold, italic, headings, lists, code blocks, and more
- **Keyboard Shortcuts** - Efficient editing with standard shortcuts

### 🎨 Theming & Customization
- **7 Beautiful Themes**:
  - Claude (Orange)
  - Dark Matter (Tech)
  - Graphite (Minimal)
  - Gruvbox
  - Notebook
  - Supabase
  - T3 Chat (Pink)
- **Dark/Light Mode** - Per-theme light and dark variants with system preference support
- **Persistent Preferences** - Theme selection saved across sessions

### 🖱️ Context Menus
- **Right-click Actions** - Context-sensitive menus for all items
- **Quick Operations**:
  - View, Create, Rename, Delete for notebooks, chapters, and notes
  - Create new items in parent context
  - Consistent menu ordering across all levels

### 📱 Responsive Design
- **Collapsible Sidebars** - Left navigation and right metadata panels
- **Adaptive Layout** - Works on desktop and tablet devices
- **Tooltips** - Helpful hints on hover for truncated items
- **Smooth Animations** - Polished transitions and state changes

### 🔍 User Experience
- **Optimistic Updates** - Instant UI feedback with background sync
- **Error Handling** - Toast notifications for success/error states
- **Loading States** - Clear indicators during data operations
- **Keyboard Navigation** - Support for keyboard-first workflows

### 🤖 AI-Powered Features
- **Intelligent Reorganization** - AI analyzes and reorganizes your entire note structure
- **AI Chat Sidebar** - Interactive AI assistant for note management
- **Multi-Provider Support** - Works with OpenAI, Anthropic (Claude), and Google (Gemini)
- **Tool Calling** - AI can create, move, rename, and organize notebooks, chapters, and notes
- **Content Analysis** - AI understands note content to suggest optimal organization
- **Smart Naming** - Automatically generates clear, descriptive names
- **Meeting Transcription** - AI-powered meeting recording and transcription
- **Note Summarization** - Generate summaries and key points from meeting notes
- **Video Generation** - Create explanatory videos from note content

### 🗂️ Organization Tools
- **Nested Structure** - Unlimited notebooks, chapters, and notes
- **AI-Assisted Organization** - Let AI reorganize your notes intelligently
- **Bulk Operations** - Efficient management of multiple items
- **Smart Defaults** - Auto-expand active branches in navigation
- **Visual Hierarchy** - Icons and indentation for clear structure

### 🔄 Data Management
- **CRUD Operations** - Full create, read, update, delete for all entities
- **Cascade Delete** - Proper cleanup of child items
- **Move Operations** - Reorganize content via drag & drop
- **Query Optimization** - Efficient data fetching with React Query

The app features a modern, intuitive interface with:
- **Dual Sidebar Layout** - Navigation on the left, metadata on the right
- **Rich Text Editor** - Full-featured TipTap editor with formatting toolbar
- **Theme Selector** - Visual theme picker with color previews
- **Context Menus** - Right-click menus throughout the interface
- **Drag & Drop** - Intuitive content reorganization

## 🏗️ Architecture

```
Users
  └── Notebooks (Drives)
       └── Chapters (Folders)
            └── Notes (Files)
```

**Data Flow:**
1. User authenticates via Google OAuth
2. Frontend fetches user's notebooks with nested chapters and notes
3. TanStack Query manages caching and optimistic updates
4. User edits are auto-saved to the backend
5. Changes sync across all components via query invalidation

## 🛠️ Tech Stack

### Backend
- **Go 1.21+** - Modern, performant programming language
- **Gin** - Fast HTTP web framework
- **GORM** - Feature-rich ORM for database operations
- **Goth/Gothic** - Multi-provider OAuth authentication
- **PostgreSQL** - Robust relational database
- **Zerolog** - High-performance structured logging

### Frontend
- **React 18** - Modern UI library with hooks
- **TypeScript** - Type-safe JavaScript
- **Vite** - Lightning-fast build tool and dev server
- **TanStack Query** - Powerful data synchronization and caching
- **TipTap** - Headless WYSIWYG editor framework
- **Tailwind CSS v4** - Utility-first CSS framework
- **shadcn/ui** - Beautifully designed component library
- **Radix UI** - Unstyled, accessible component primitives
- **React Router** - Client-side routing
- **Axios** - Promise-based HTTP client
- **Sonner** - Toast notifications
- **dnd-kit** - Modern drag and drop toolkit
- **Lucide React** - Beautiful icon library
- **Zod** - TypeScript-first schema validation

## 📁 Project Structure

```
notes-app/
├── backend/
│   ├── cmd/
│   │   └── main.go                      # Application entry point
│   │
│   ├── db/
│   │   ├── db.go                        # Database connection setup
│   │   └── migrations/                  # SQL migration files
│   │
│   ├── internal/
│   │   ├── auth/
│   │   │   └── auth.go                  # OAuth configuration
│   │   │
│   │   ├── controllers/
│   │   │   ├── notebook.controller.go   # Notebook HTTP handlers
│   │   │   ├── chapter.controller.go    # Chapter HTTP handlers
│   │   │   └── note.controller.go       # Note HTTP handlers
│   │   │
│   │   ├── models/
│   │   │   ├── user.model.go            # User entity
│   │   │   ├── notebook.model.go        # Notebook entity
│   │   │   ├── chapter.model.go         # Chapter entity
│   │   │   └── note.model.go            # Note entity
│   │   │
│   │   └── middleware/
│   │       └── auth.middleware.go       # Authentication middleware
│   │
│   ├── .env.example                     # Environment template
│   ├── go.mod                           # Go dependencies
│   ├── go.sum                           # Dependency checksums
│   └── README.md                        # Backend documentation
│
└── frontend/
    ├── src/
    │   ├── components/
    │   │   ├── ui/                      # shadcn/ui components
    │   │   ├── Dashboard.tsx            # Main dashboard layout
    │   │   ├── Header.tsx               # Top navigation bar
    │   │   ├── left-sidebar-content.tsx # Navigation sidebar
    │   │   ├── right-sidebar-content.tsx# Metadata sidebar
    │   │   ├── ThemeSelector.tsx        # Theme picker dropdown
    │   │   ├── ModeToggle.tsx           # Light/dark mode toggle
    │   │   └── theme-provider.tsx       # Theme context provider
    │   │
    │   ├── hooks/
    │   │   └── auth.ts                  # Authentication hooks
    │   │
    │   ├── types/
    │   │   └── backend.ts               # TypeScript type definitions
    │   │
    │   ├── utils/
    │   │   ├── api.ts                   # Axios instance configuration
    │   │   ├── auth.ts                  # Auth utilities
    │   │   ├── notebook.ts              # Notebook API calls
    │   │   ├── chapter.ts               # Chapter API calls
    │   │   └── notes.ts                 # Notes API calls
    │   │
    │   ├── index.css                    # Global styles & theme tokens
    │   ├── App.tsx                      # Root component
    │   └── main.tsx                     # Application entry point
    │
    ├── public/                          # Static assets
    ├── .env.example                     # Environment variables template
    ├── package.json                     # Dependencies
    ├── tsconfig.json                    # TypeScript configuration
    ├── tailwind.config.js               # Tailwind CSS configuration
    ├── vite.config.ts                   # Vite configuration
    └── README.md                        # Frontend documentation
```

## Getting Started

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18 or higher
- **PostgreSQL** 14 or higher
- **Google OAuth Credentials** ([Get them here](https://console.cloud.google.com/))
- **AI API Key** (Optional, for AI features) - Get from [OpenAI](https://platform.openai.com/), [Anthropic](https://console.anthropic.com/), or [Google AI](https://ai.google.dev/)

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

### AI Features Setup (Optional)

To enable AI-powered features like intelligent reorganization:

1. **Login** to the app with Google OAuth
2. **Go to Profile** → Click your avatar in the top-right
3. **Navigate to Settings** → AI Credentials
4. **Add your API key** for one or more providers:
   - OpenAI (for GPT models)
   - Anthropic (for Claude models)
   - Google (for Gemini models)
5. **Open AI Chat** → Click the chat icon in the right sidebar
6. **Try reorganization** → Ask: "Please reorganize my notes"

**Note:** API keys are encrypted and stored securely in the database.

## 🔌 API Endpoints

### Authentication
- `GET /auth/google` - Initiate Google OAuth flow
- `GET /auth/google/callback` - OAuth callback handler
- `GET /auth/user` - Get current authenticated user details
- `GET /logout/google` - End user session and logout

### Notebooks
- `POST /notebook` - Create a new notebook
- `GET /notebook/:id` - Get notebook by ID with nested chapters and notes
- `GET /notebooks` - Get all notebooks for authenticated user
- `PUT /notebook/:id` - Update notebook name/properties
- `DELETE /notebook/:id` - Delete notebook and cascade delete children

### Chapters
- `POST /chapter` - Create a new chapter in a notebook
- `GET /chapter/:id` - Get chapter by ID with nested notes
- `PUT /chapter/:id` - Update chapter name/properties
- `PUT /chapter/:id/move` - Move chapter to different notebook
- `DELETE /chapter/:id` - Delete chapter and cascade delete notes

### Notes
- `POST /note` - Create a new note in a chapter
- `GET /note/:id` - Get note by ID with full content
- `PUT /note/:id` - Update note name/content
- `PUT /note/:id/move` - Move note to different chapter
- `DELETE /note/:id` - Delete note

**Authentication:** All endpoints (except auth routes) require valid session cookie.

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

## ⌨️ Keyboard Shortcuts

### Editor
- `Ctrl/Cmd + B` - Bold
- `Ctrl/Cmd + I` - Italic
- `Ctrl/Cmd + Shift + X` - Strikethrough
- `Ctrl/Cmd + Shift + H` - Highlight
- `Ctrl/Cmd + Alt + 1-6` - Headings 1-6
- `Ctrl/Cmd + Shift + 7` - Ordered list
- `Ctrl/Cmd + Shift + 8` - Bullet list
- `Ctrl/Cmd + Shift + 9` - Blockquote
- `Ctrl/Cmd + E` - Code
- `Ctrl/Cmd + Shift + E` - Code block
- `Ctrl/Cmd + Z` - Undo
- `Ctrl/Cmd + Shift + Z` - Redo

## 🎯 Key Features Explained

### AI-Powered Reorganization 🤖
**The standout feature that makes this notes app unique!**

- **Intelligent Analysis** - AI reads all your notes and understands their content
- **Automatic Organization** - Creates optimal notebook/chapter structure
- **Smart Naming** - Suggests clear, descriptive names for better discoverability
- **Content-Based Grouping** - Groups notes by topic, project, or theme
- **One Command** - Just ask: "Please reorganize my notes"

**Example prompts:**
```
"Reorganize my notes by topic"
"Create separate notebooks for work and personal"
"Organize my meeting notes by project"
"Clean up my messy structure"
```

**What the AI can do:**
- Create new notebooks and chapters
- Move notes and chapters to better locations
- Rename items for clarity
- Analyze content to determine themes
- Provide a summary of all changes

See [AI Reorganization Documentation](./docs/AI_REORGANIZATION_FEATURE.md) for detailed guide.

### Drag & Drop
- **Chapters** - Drag chapters between notebooks to reorganize
- **Notes** - Drag notes between chapters for better organization
- **Visual Feedback** - Highlighted drop zones during drag operations

### Auto-save
- Debounced saving (500ms after last edit)
- Visual indicator shows save status
- Optimistic updates for instant feedback

### Theme System
- CSS custom properties for consistent theming
- Separate light/dark variants per theme
- System preference detection
- Persistent storage in localStorage

### Context Menus
- Right-click on notebooks for notebook actions
- Right-click on chapters for chapter actions
- Right-click on notes for note actions
- Right-click on empty space to create notebooks

## 🚀 Future Enhancements

- [ ] Full-text search across all notes
- [ ] Tags and categories
- [ ] Markdown import/export
- [ ] Real-time collaboration
- [ ] Version history
- [ ] File attachments
- [ ] Mobile responsive design
- [ ] Offline support with PWA
- [ ] Keyboard shortcuts customization
- [ ] Templates for common note types

## 📝 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Goth](https://github.com/markbates/goth)
- [React](https://react.dev/)
- [Vite](https://vitejs.dev/)
- [TipTap](https://tiptap.dev/)
- [shadcn/ui](https://ui.shadcn.com/)
- [TanStack Query](https://tanstack.com/query)
- [dnd-kit](https://dndkit.com/)

