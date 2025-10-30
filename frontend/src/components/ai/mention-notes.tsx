import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Badge } from "@/components/ui/badge"
import { X, FileText } from "lucide-react"
import { cn } from "@/lib/utils"
import { useEffect, useRef } from "react"

interface MentionedNote {
  id: string
  name: string
  content: string
  chapterName: string
  notebookName: string
}

interface MentionNotesPopoverProps {
  open: boolean
  notes: MentionedNote[]
  onSelectNote: (note: MentionedNote) => void
  isLoading?: boolean
}

export function MentionNotesPopover({
  open,
  notes,
  onSelectNote,
  isLoading = false,
}: MentionNotesPopoverProps) {
  const searchInputRef = useRef<HTMLInputElement>(null)

  // Focus the search input when popover opens
  useEffect(() => {
    if (open && searchInputRef.current) {
      // Small delay to ensure the popover is rendered
      const timer = setTimeout(() => {
        searchInputRef.current?.focus()
      }, 50)
      return () => clearTimeout(timer)
    }
  }, [open])

  if (!open) return null

  return (
    <div className="absolute bottom-full left-0 right-0 mb-2 z-50">
      <div className="w-full max-w-[450px] rounded-lg border bg-popover text-popover-foreground shadow-xl">
        <Command className="rounded-lg">
          <CommandInput 
            ref={searchInputRef}
            placeholder="Search notes..." 
            className="h-9" 
          />
          <CommandList className="max-h-[300px]">
            {isLoading ? (
              <div className="p-4 text-center text-sm text-muted-foreground">
                Loading notes...
              </div>
            ) : notes.length === 0 ? (
              <CommandEmpty className="py-6 text-center text-sm">
                No notes found. Create some notes first!
              </CommandEmpty>
            ) : (
              <CommandGroup heading={`${notes.length} note${notes.length !== 1 ? 's' : ''} available`}>
                {notes.map((note) => (
                  <CommandItem
                    key={note.id}
                    value={`${note.name} ${note.notebookName} ${note.chapterName}`}
                    onSelect={() => onSelectNote(note)}
                    className="flex items-start gap-3 py-3 cursor-pointer"
                  >
                    <FileText className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
                    <div className="flex-1 space-y-1 overflow-hidden">
                      <p className="text-sm font-medium leading-none truncate">{note.name}</p>
                      <p className="text-xs text-muted-foreground truncate">
                        {note.notebookName} → {note.chapterName}
                      </p>
                      <p className="text-xs text-muted-foreground line-clamp-2">
                        {note.content ? note.content.substring(0, 100) + (note.content.length > 100 ? '...' : '') : 'Empty note'}
                      </p>
                    </div>
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </div>
    </div>
  )
}

interface MentionedNotesBadgesProps {
  notes: MentionedNote[]
  onRemove: (noteId: string) => void
  className?: string
}

export function MentionedNotesBadges({
  notes,
  onRemove,
  className,
}: MentionedNotesBadgesProps) {
  if (notes.length === 0) return null

  return (
    <div className={cn("flex flex-wrap gap-2 p-2 border-b", className)}>
      {notes.map((note) => (
        <Badge
          key={note.id}
          variant="secondary"
          className="group gap-1 pl-2 pr-1.5 py-1 text-xs font-normal"
        >
          <FileText className="h-3 w-3" />
          <span className="max-w-[150px] truncate">{note.name}</span>
          <button
            type="button"
            onClick={(e) => {
              e.preventDefault()
              onRemove(note.id)
            }}
            className="ml-1 rounded-sm opacity-70 hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1"
          >
            <X className="h-3 w-3" />
          </button>
        </Badge>
      ))}
    </div>
  )
}

