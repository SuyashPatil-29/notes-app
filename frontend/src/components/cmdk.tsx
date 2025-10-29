import { Command } from 'cmdk'
import { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Book, BookOpen, FileText, Search } from 'lucide-react'
import type { Notebook } from '@/types/backend'

interface CommandMenuProps {
  notebooks: Notebook[] | undefined
}

const CommandMenu = ({ notebooks }: CommandMenuProps) => {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)

  // Toggle the menu when ⌘K is pressed
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
      }
    }

    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [])

  // Handle Escape key to close the dialog
  useEffect(() => {
    if (!open) return

    const down = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        e.stopPropagation()
        setOpen(false)
        setSearch('')
      }
    }

    document.addEventListener('keydown', down, { capture: true })
    return () => document.removeEventListener('keydown', down, { capture: true })
  }, [open])

  // Auto-focus input when dialog opens
  useEffect(() => {
    if (open) {
      // Small delay to ensure the dialog is rendered
      setTimeout(() => {
        inputRef.current?.focus()
      }, 0)
    } else {
      setSearch('')
    }
  }, [open])

  const handleSelect = (path: string) => {
    setOpen(false)
    setSearch('')
    navigate(path)
  }

  // Build searchable items
  const notebookItems = notebooks?.map(notebook => ({
    type: 'notebook' as const,
    id: notebook.id,
    name: notebook.name,
    path: `/${notebook.id}`,
    notebook: notebook.name,
  })) || []

  const chapterItems = notebooks?.flatMap(notebook =>
    notebook.chapters?.map(chapter => ({
      type: 'chapter' as const,
      id: chapter.id,
      name: chapter.name,
      path: `/${notebook.id}/${chapter.id}`,
      notebook: notebook.name,
    })) || []
  ) || []

  const noteItems = notebooks?.flatMap(notebook =>
    notebook.chapters?.flatMap(chapter =>
      chapter.notes?.map(note => ({
        type: 'note' as const,
        id: note.id,
        name: note.name,
        path: `/${notebook.id}/${chapter.id}/${note.id}`,
        notebook: notebook.name,
        chapter: chapter.name,
        content: note.content,
      })) || []
    ) || []
  ) || []

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm">
      <div className="fixed left-1/2 top-[15vh] w-full max-w-2xl -translate-x-1/2">
        <Command 
          className="rounded-xl border-2 border-primary/30 bg-card shadow-2xl overflow-hidden"
          label="Global Command Menu"
          shouldFilter={true}
        >
          <div className="flex items-center border-b-2 border-border px-4 py-3 bg-background/50">
            <Search className="mr-3 h-5 w-5 shrink-0 text-primary" />
            <Command.Input 
              ref={inputRef}
              placeholder="Search notebooks, chapters, and notes..."
              value={search}
              onValueChange={setSearch}
              autoFocus
              className="flex-1 bg-transparent outline-none text-base font-medium text-foreground placeholder:text-muted-foreground"
            />
          </div>
          
          <Command.List className="max-h-[400px] overflow-y-auto p-2">
            <Command.Empty className="py-12 text-center text-sm font-medium text-foreground">
              No results found.
            </Command.Empty>

            {notebookItems.length > 0 && (
              <Command.Group className="px-2 pb-2">
                <div className="px-2 py-2 text-xs font-bold uppercase tracking-wider text-primary">
                  Notebooks
                </div>
                {notebookItems.map((item) => (
                  <Command.Item
                    key={item.id}
                    value={`notebook-${item.name}`}
                    onSelect={() => handleSelect(item.path)}
                    className="flex items-center gap-3 rounded-lg px-3 py-3 cursor-pointer data-[selected=true]:bg-primary data-[selected=true]:text-primary-foreground transition-colors mb-1"
                  >
                    <Book className="h-5 w-5 shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-semibold truncate">{item.name}</div>
                    </div>
                  </Command.Item>
                ))}
              </Command.Group>
            )}

            {chapterItems.length > 0 && (
              <Command.Group className="px-2 pb-2">
                <div className="px-2 py-2 text-xs font-bold uppercase tracking-wider text-primary">
                  Chapters
                </div>
                {chapterItems.map((item) => (
                  <Command.Item
                    key={item.id}
                    value={`chapter-${item.name}-${item.notebook}`}
                    onSelect={() => handleSelect(item.path)}
                    className="flex items-center gap-3 rounded-lg px-3 py-3 cursor-pointer data-[selected=true]:bg-primary data-[selected=true]:text-primary-foreground transition-colors mb-1"
                  >
                    <BookOpen className="h-5 w-5 shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-semibold truncate">{item.name}</div>
                      <div className="text-xs text-foreground/70 data-[selected=true]:text-primary-foreground truncate">
                        in {item.notebook}
                      </div>
                    </div>
                  </Command.Item>
                ))}
              </Command.Group>
            )}

            {noteItems.length > 0 && (
              <Command.Group className="px-2 pb-2">
                <div className="px-2 py-2 text-xs font-bold uppercase tracking-wider text-primary">
                  Notes
                </div>
                {noteItems.map((item) => (
                  <Command.Item
                    key={item.id}
                    value={`note-${item.name}-${item.chapter}-${item.notebook}-${item.content}`}
                    onSelect={() => handleSelect(item.path)}
                    className="flex items-center gap-3 rounded-lg px-3 py-3 cursor-pointer data-[selected=true]:bg-primary data-[selected=true]:text-primary-foreground transition-colors mb-1"
                  >
                    <FileText className="h-5 w-5 shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-semibold truncate">{item.name}</div>
                      <div className="text-xs text-foreground/70 data-[selected=true]:text-primary-foreground truncate">
                        {item.chapter} • {item.notebook}
                      </div>
                    </div>
                  </Command.Item>
                ))}
              </Command.Group>
            )}
          </Command.List>
          
          <div className="flex items-center gap-4 border-t-2 border-border px-4 py-3 text-xs font-medium text-foreground bg-background/50">
            <div className="flex items-center gap-2">
              <kbd className="inline-flex items-center justify-center min-w-[24px] h-6 px-2 text-[11px] font-bold rounded bg-muted border border-border shadow-sm">↑↓</kbd>
              <span>Navigate</span>
            </div>
            <div className="flex items-center gap-2">
              <kbd className="inline-flex items-center justify-center min-w-[24px] h-6 px-2 text-[11px] font-bold rounded bg-muted border border-border shadow-sm">↵</kbd>
              <span>Select</span>
            </div>
            <div className="flex items-center gap-2">
              <kbd className="inline-flex items-center justify-center min-w-[24px] h-6 px-2 text-[11px] font-bold rounded bg-muted border border-border shadow-sm">esc</kbd>
              <span>Close</span>
            </div>
          </div>
        </Command>
      </div>
      
      {/* Click outside to close */}
      <div 
        className="fixed inset-0 -z-10" 
        onClick={() => setOpen(false)}
      />
    </div>
  )
}

export default CommandMenu