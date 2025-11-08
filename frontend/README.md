# Atlas - Frontend

Modern React frontend for Atlas, a hierarchical knowledge management application with rich text editing, drag & drop, and beautiful theming.

## 🛠️ Tech Stack

- **React 18** - Modern UI library with hooks and concurrent features
- **TypeScript** - Type-safe development
- **Vite** - Lightning-fast build tool with HMR
- **TanStack Query (React Query)** - Powerful data fetching and state management
- **TipTap** - Headless WYSIWYG editor with ProseMirror
- **Tailwind CSS v4** - Utility-first CSS with custom design system
- **shadcn/ui** - High-quality component library
- **Radix UI** - Unstyled, accessible primitives
- **React Router** - Type-safe client-side routing
- **dnd-kit** - Modern drag and drop toolkit
- **Sonner** - Beautiful toast notifications
- **Lucide React** - Consistent icon system
- **Axios** - Promise-based HTTP client
- **Zod** - Runtime type validation

## ✨ Features

### 🎨 Theming System
- **7 Beautiful Themes**: Claude, Dark Matter, Graphite, Gruvbox, Notebook, Supabase, T3 Chat
- **Light/Dark Modes**: Each theme has both variants
- **System Preference**: Automatic detection of OS preference
- **Persistent Storage**: Theme selection saved to localStorage
- **CSS Custom Properties**: Easy theme customization

### ✏️ Rich Text Editor
- **TipTap Integration**: Full-featured WYSIWYG editor
- **Formatting Tools**: Bold, italic, strikethrough, headings, lists, quotes, code
- **Auto-save**: Debounced saving with visual feedback
- **Keyboard Shortcuts**: Standard shortcuts for all formatting
- **Markdown Support**: Compatible with markdown syntax

### 🗂️ Hierarchical Navigation
- **Collapsible Sidebar**: Smooth expand/collapse animations
- **Drag & Drop**: Move chapters and notes between containers
- **Active State**: Visual indication of current note
- **Context Menus**: Right-click actions throughout
- **Tooltips**: Helper text for truncated items

### 🚀 Performance Optimizations
- **Optimistic Updates**: Instant UI feedback
- **Query Caching**: Smart data caching with TanStack Query
- **Code Splitting**: Lazy loading for better initial load
- **React Compiler**: Automatic memoization (experimental)

## 📁 Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/                    # shadcn/ui components
│   │   ├── Dashboard.tsx          # Main dashboard layout
│   │   ├── Header.tsx             # Top navigation bar
│   │   ├── left-sidebar-content.tsx  # Navigation sidebar
│   │   ├── right-sidebar-content.tsx # Metadata sidebar
│   │   ├── ThemeSelector.tsx      # Theme picker dropdown
│   │   ├── ModeToggle.tsx         # Light/dark mode toggle
│   │   └── theme-provider.tsx     # Theme context provider
│   │
│   ├── hooks/
│   │   └── auth.ts                # Authentication hooks
│   │
│   ├── types/
│   │   └── backend.ts             # TypeScript type definitions
│   │
│   ├── utils/
│   │   ├── api.ts                 # Axios instance configuration
│   │   ├── auth.ts                # Auth utilities
│   │   ├── notebook.ts            # Notebook API calls
│   │   ├── chapter.ts             # Chapter API calls
│   │   └── notes.ts               # Notes API calls
│   │
│   ├── index.css                  # Global styles & theme tokens
│   ├── App.tsx                    # Root component
│   └── main.tsx                   # Application entry point
│
├── public/                        # Static assets
├── .env.example                   # Environment variables template
├── package.json                   # Dependencies
├── tsconfig.json                  # TypeScript configuration
├── tailwind.config.js             # Tailwind CSS configuration
└── vite.config.ts                 # Vite configuration
```

## 🚀 Getting Started

### Prerequisites

- Node.js 18+ or Bun 1.0+
- Backend server running on `http://localhost:8080`

### Installation

1. **Install dependencies:**
   ```bash
   bun install
   # or
   npm install
   ```

2. **Copy environment variables (if needed):**
   ```bash
   cp .env.example .env
   ```

3. **Start development server:**
   ```bash
   bun dev
   # or
   npm run dev
   ```

   The app will be available at `http://localhost:5173`

### Building for Production

```bash
bun run build
# or
npm run build
```

Preview production build:
```bash
bun run preview
# or
npm run preview
```

## 🎯 Key Features Explained

### Context Menus
Right-click on any item to access contextual actions:
- **Notebooks**: View, New Chapter, Rename, Delete
- **Chapters**: View, New Note, Rename, Delete
- **Notes**: View, New Note, Rename, Delete
- **Empty Space**: New Notebook

### Drag & Drop System
Powered by dnd-kit for accessible, smooth interactions:
- Drag chapters between notebooks
- Drag notes between chapters
- Visual feedback with drop zones
- Optimistic updates for instant feedback

### Auto-save Mechanism
- Debounced saving (500ms after last edit)
- Visual "Saving..." indicator
- Automatic retry on failure
- No manual save button needed

### Theme Architecture
```css
/* Each theme defines CSS custom properties */
html[data-theme='notebook'] {
  --background: ...;
  --foreground: ...;
  --primary: ...;
  /* ... more tokens */
}

/* Dark mode overrides */
html[data-theme='notebook'].dark {
  --background: ...;
  /* ... dark variants */
}
```

## ⌨️ Keyboard Shortcuts

### Editor
- `Cmd/Ctrl + B` - Bold
- `Cmd/Ctrl + I` - Italic
- `Cmd/Ctrl + Shift + X` - Strikethrough
- `Cmd/Ctrl + Shift + H` - Highlight
- `Cmd/Ctrl + Alt + 1-6` - Headings
- `Cmd/Ctrl + Shift + 7` - Ordered list
- `Cmd/Ctrl + Shift + 8` - Bullet list
- `Cmd/Ctrl + Shift + 9` - Blockquote
- `Cmd/Ctrl + E` - Inline code
- `Cmd/Ctrl + Shift + E` - Code block
- `Cmd/Ctrl + Z` - Undo
- `Cmd/Ctrl + Shift + Z` - Redo

## 📦 Component Library

### UI Components (shadcn/ui)
- Button, Input, Dialog, Dropdown Menu
- Sidebar (Left & Right)
- Context Menu
- Tooltip
- Collapsible

### Custom Components
- **Dashboard**: Main layout with dual sidebars
- **ThemeSelector**: Visual theme picker
- **ModeToggle**: Light/dark mode switcher
- **Editor**: TipTap-based rich text editor

## 🔧 Configuration

### Environment Variables
```bash
# API Base URL (optional, defaults to localhost:8080)
VITE_API_URL=http://localhost:8080
```

### Tailwind Configuration
- Custom design system with CSS variables
- Extended color palette from CSS tokens
- Custom shadow and radius utilities
- Typography system

## 🧪 Development Tips

### React Compiler
The React Compiler is enabled for automatic memoization:
- Improves performance
- Eliminates need for manual `useMemo`/`useCallback`
- May impact dev/build performance slightly

### Code Splitting
Routes are lazy-loaded for optimal bundle size:
```typescript
const Dashboard = lazy(() => import('./components/Dashboard'))
```

### Type Safety
All API responses are typed:
```typescript
type Notebook = {
  id: string
  name: string
  chapters?: Chapter[]
  // ... more fields
}
```

## 🔗 API Integration

The frontend communicates with the backend via:
- **Authentication**: Cookie-based sessions
- **Data Fetching**: TanStack Query with auto-caching
- **Optimistic Updates**: Instant UI feedback
- **Error Handling**: Toast notifications

## 🎨 Styling Approach

1. **Tailwind CSS**: Utility classes for layout
2. **CSS Variables**: Theme tokens for colors
3. **shadcn/ui**: Pre-built accessible components
4. **Custom CSS**: Complex animations and transitions

## 🐛 Debugging

```bash
# Check TypeScript errors
bun run type-check

# Run linter
bun run lint

# Format code
bun run format
```

## 📚 Resources

- [React Documentation](https://react.dev/)
- [TanStack Query](https://tanstack.com/query/latest)
- [TipTap Guide](https://tiptap.dev/docs/editor/getting-started/overview)
- [shadcn/ui](https://ui.shadcn.com/)
- [Tailwind CSS](https://tailwindcss.com/)
- [dnd-kit](https://docs.dndkit.com/)

## 🤝 Contributing

When adding new features:
1. Follow existing TypeScript patterns
2. Use TanStack Query for data fetching
3. Maintain type safety throughout
4. Add proper error handling
5. Update relevant documentation
