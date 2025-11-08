# Atlas

A full-stack knowledge management application with hierarchical organization (Users → Notebooks → Chapters → Notes) built with Go (Gin + GORM) backend and React (TypeScript + Vite) frontend. **Now featuring WhatsApp bot integration and AI-powered note creation!**

## ✨ Features

### 📱 WhatsApp Bot Integration (NEW!)
- **Natural Language Commands** - Create notes by simply typing in plain English
- **AI-Powered Organization** - AI automatically decides notebook and chapter placement
- **Auto Content Generation** - AI generates detailed content for your notes
- **Instant Creation** - No confirmations, no steps - just type and go!
- **Command Support**:
  - `add [note title]` - Create a note with AI organization and content
  - `/cancel` - Cancel any ongoing operation
  - `/retrieve [note name]` - View note content in markdown
  - `/list notes` - List all your notes
  - `/help` - Get command help
- **Smart Markdown Conversion** - Automatically converts between TipTap JSON and Markdown
- **Auto-refresh Frontend** - See changes from WhatsApp instantly in the app

**Example Usage:**
```
You: add steps to deploy app in backend go
Bot: 🤖 Creating your note with AI...
Bot: ✅ Note created successfully!
     📝 Title: steps to deploy app in backend go
     📓 Notebook: Development
     📑 Chapter: Backend
     🤖 AI-generated content added!
```

### 🔐 Authentication & Security
- **Clerk Authentication** - Modern, secure authentication with custom UI
- **Multiple Auth Methods** - Email/password and OAuth (Google, GitHub)
- **Email Verification** - Required verification for new accounts
- **Session Management** - Secure JWT-based sessions handled by Clerk
- **Protected Routes** - Client-side route protection for authenticated users
- **WhatsApp Account Linking** - Secure token-based WhatsApp authentication

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
- **Markdown Support** - Import/export and compatibility with markdown

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
- **Auto-refresh** - Frontend automatically updates every 30 seconds to catch WhatsApp changes

### 🤖 AI-Powered Features
- **WhatsApp Natural Language** - Just type what you want to create
- **Intelligent Organization** - AI analyzes note titles and chooses the best location
- **Smart Notebook Selection** - Reuses existing notebooks when appropriate
- **Auto Chapter Creation** - Creates chapters as needed for better organization
- **Content Generation** - AI writes detailed, relevant content for your notes
- **Multi-Provider Support** - Works with OpenAI, Anthropic (Claude), and Google (Gemini)
- **Meeting Transcription** - AI-powered meeting recording and transcription
- **Note Summarization** - Generate summaries and key points from meeting notes
- **Video Generation** - Create explanatory videos from note content
- **AI Chat Sidebar** - Interactive AI assistant for note management
- **Tool Calling** - AI can create, move, rename, and organize notebooks, chapters, and notes

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
- **Real-time Sync** - Changes from WhatsApp appear in the app automatically

## 🏗️ Architecture

```
Users
  └── Notebooks (Drives)
       └── Chapters (Folders)
            └── Notes (Files)
```

**Data Flow:**
1. User authenticates via Clerk (email/password or OAuth)
2. Frontend receives JWT session token
3. API calls include token in Authorization header
4. Backend validates token with Clerk
5. Frontend fetches user's notebooks with nested chapters and notes
6. TanStack Query manages caching and optimistic updates
7. User edits are auto-saved to the backend
8. Changes sync across all components via query invalidation
9. WhatsApp bot creates notes independently with AI assistance

**WhatsApp Integration Flow:**
1. User links WhatsApp account via secure token
2. User sends natural language command (e.g., "add meeting notes")
3. Bot parses command and uses AI to determine notebook/chapter
4. AI generates relevant content for the note
5. Backend creates notebook/chapter if needed, saves note
6. Frontend auto-refreshes to show new content
7. User can retrieve and manage notes via WhatsApp or web

## 🛠️ Tech Stack

### Backend
- **Go 1.21+** - Modern, performant programming language
- **Gin** - Fast HTTP web framework
- **GORM** - Feature-rich ORM for database operations
- **Clerk Go SDK** - Modern authentication and user management
- **SQLite** - Lightweight embedded database (development)
- **PostgreSQL** - Production database with full-text search
- **Zerolog** - High-performance structured logging
- **OpenAI Go SDK** - AI integration for content generation
- **WhatsApp Business API** - Official Meta WhatsApp API integration

### Frontend
- **React 19** - Modern UI library with hooks
- **TypeScript** - Type-safe JavaScript
- **Vite** - Lightning-fast build tool and dev server
- **Clerk React SDK** - Custom authentication UI components
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

### AI & WhatsApp
- **OpenAI GPT-4** - Content generation and organization
- **Meta WhatsApp Business API** - Message sending and webhooks
- **Natural Language Processing** - Command parsing and intent detection
- **TipTap ↔ Markdown Conversion** - Bidirectional content format conversion

## 📁 Project Structure

```
atlas/
├── backend/
│   ├── cmd/
│   │   └── main.go                      # Application entry point
│   │
│   ├── internal/
│   │   ├── controllers/
│   │   │   ├── whatsapp.controller.go   # WhatsApp webhook handlers
│   │   │   ├── notebook.controller.go   # Notebook HTTP handlers
│   │   │   ├── chapter.controller.go    # Chapter HTTP handlers
│   │   │   └── note.controller.go       # Note HTTP handlers
│   │   │
│   │   ├── services/
│   │   │   ├── ai_service.go            # AI content generation & organization
│   │   │   ├── whatsapp_message_processor.go  # Message routing
│   │   │   ├── whatsapp_auth_service.go # WhatsApp authentication
│   │   │   └── whatsapp_audit_service.go # Message logging
│   │   │
│   │   ├── whatsapp/
│   │   │   ├── commands/
│   │   │   │   ├── add_note_command.go  # /add command with AI
│   │   │   │   ├── cancel_command.go    # /cancel command
│   │   │   │   ├── retrieve_note_command.go  # /retrieve command
│   │   │   │   ├── list_command.go      # /list command
│   │   │   │   └── help_command.go      # /help command
│   │   │   ├── command.go               # Command registry
│   │   │   └── natural_language_parser.go # NL command parsing
│   │   │
│   │   ├── utils/
│   │   │   ├── markdown_to_tiptap.go    # Markdown → TipTap JSON
│   │   │   └── tiptap_to_markdown.go    # TipTap JSON → Markdown
│   │   │
│   │   ├── models/
│   │   │   ├── whatsapp_user.model.go   # WhatsApp users
│   │   │   ├── whatsapp_message.model.go # Message audit log
│   │   │   ├── notebook.model.go        # Notebook entity
│   │   │   ├── chapter.model.go         # Chapter entity
│   │   │   └── note.model.go            # Note entity
│   │   │
│   │   └── middleware/
│   │       └── auth.middleware.go       # Authentication middleware
│   │
│   └── pkg/whatsapp/
│       └── client.go                    # WhatsApp API client
│
└── frontend/
    ├── src/
    │   ├── pages/
    │   │   └── whatsapp-auth-page.tsx   # WhatsApp linking page
    │   │
    │   ├── components/
    │   │   ├── Dashboard.tsx            # Main dashboard layout
    │   │   ├── Header.tsx               # Top navigation bar
    │   │   └── ... (other components)
    │   │
    │   └── utils/
    │       └── api.ts                   # API client with auto-refresh
    │
    └── ...
```

## Getting Started

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18 or higher (or **Bun** for faster installs)
- **Clerk Account** - [Sign up for free](https://clerk.com/)
- **OpenAI API Key** - Get from [OpenAI](https://platform.openai.com/)
- **WhatsApp Business Account** (Optional, for WhatsApp features) - [Get started](https://business.whatsapp.com/)

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
   ```env
   # Clerk Authentication
   CLERK_SECRET_KEY=sk_test_...
   
   # OpenAI
   OPENAI_API_KEY=sk-proj-...
   
   # WhatsApp (Optional)
   WHATSAPP_ACCESS_TOKEN=your_meta_access_token
   WHATSAPP_PHONE_NUMBER_ID=your_phone_number_id
   WHATSAPP_VERIFY_TOKEN=your_verify_token
   WHATSAPP_WEBHOOK_SECRET=your_webhook_secret
   
   # Database
   DATABASE_URL=sqlite.db
   
   # Server
   PORT=8080
   FRONTEND_URL=http://localhost:5173
   ```

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

3. **Configure environment variables:**
   Create a `.env.local` file:
   ```bash
   VITE_CLERK_PUBLISHABLE_KEY=pk_test_...
   ```

4. **Start development server:**
   ```bash
   bun dev
   # or
   npm run dev
   ```

   Frontend will start on `http://localhost:5173`

### WhatsApp Bot Setup (Optional)

1. **Set up WhatsApp Business Account:**
   - Go to [Meta Business Suite](https://business.facebook.com/)
   - Create a WhatsApp Business app
   - Get your access token and phone number ID

2. **Configure Webhook:**
   - Set webhook URL to `https://yourdomain.com/api/whatsapp/webhook`
   - Use your `WHATSAPP_VERIFY_TOKEN` for verification
   - Subscribe to `messages` events

3. **Link Your WhatsApp Account:**
   - Login to the web app
   - Send any message to your WhatsApp Business number
   - Bot will send you an authentication link
   - Click the link to complete linking

4. **Start Creating Notes:**
   ```
   add meeting notes from today
   add recipe for chocolate cake
   add deployment checklist
   ```

## 🔌 API Endpoints

### WhatsApp Endpoints
- `GET /api/whatsapp/webhook` - Webhook verification
- `POST /api/whatsapp/webhook` - Receive messages
- `POST /api/whatsapp/link` - Link WhatsApp account

### Authentication
- `GET /auth/user` - Get current authenticated user details
- `POST /onboarding` - Mark user onboarding as complete
- `GET /onboarding` - Get onboarding completion status
- `POST /settings/ai-credentials` - Set encrypted AI API credentials
- `GET /settings/ai-credentials` - Get list of configured AI providers
- `DELETE /settings/ai-credentials` - Delete an AI API credential

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

**Authentication:** All endpoints (except public routes and webhooks) require valid Clerk session token in the Authorization header.

## 🤖 WhatsApp Bot Commands

### Natural Language (No Slash)
```
add [note title]
```
**Example:** `add steps to deploy app in backend go`
- AI automatically organizes into appropriate notebook and chapter
- AI generates detailed content
- Creates notebooks/chapters as needed
- Instant creation with one command

### Slash Commands
- `/help` - Show all available commands
- `/cancel` - Cancel any ongoing operation
- `/retrieve [note name]` - View note content in markdown
- `/list notes` - List all your notes
- `/list notebooks` - List all notebooks
- `/list chapters` - List chapters in a notebook
- `/delete [note name]` - Delete a note

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

## 🚀 Future Enhancements

- [ ] Voice note transcription via WhatsApp
- [ ] Image upload and OCR via WhatsApp
- [ ] Collaborative editing in real-time
- [ ] Mobile app (iOS/Android)
- [ ] WhatsApp group integration for team notes
- [ ] Scheduled reminders via WhatsApp
- [ ] Full-text search across all notes
- [ ] Tags and categories
- [ ] Markdown import/export
- [ ] Version history
- [ ] File attachments
- [ ] Offline support with PWA

## 📝 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Clerk](https://clerk.com/)
- [React](https://react.dev/)
- [Vite](https://vitejs.dev/)
- [TipTap](https://tiptap.dev/)
- [shadcn/ui](https://ui.shadcn.com/)
- [TanStack Query](https://tanstack.com/query)
- [dnd-kit](https://dndkit.com/)
- [OpenAI](https://openai.com/)
- [Meta WhatsApp Business API](https://developers.facebook.com/docs/whatsapp)

## License

This project is licensed under the MIT License.
