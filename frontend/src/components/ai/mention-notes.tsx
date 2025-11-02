import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Badge } from "@/components/ui/badge"
import { X, FileText } from "lucide-react"
import { cn } from "@/lib/utils"
import { useEffect, useRef, useMemo } from "react"
import { getPreviewText } from "@/utils/markdown"

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
  currentNoteId?: string | null
}

export function MentionNotesPopover({
  open,
  notes,
  onSelectNote,
  isLoading = false,
  currentNoteId = null,
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

  // Split notes into current note and other notes
  const { currentNote, otherNotes } = useMemo(() => {
    if (!currentNoteId) {
      return { currentNote: null, otherNotes: notes }
    }
    
    const current = notes.find(note => note.id === currentNoteId)
    const others = notes.filter(note => note.id !== currentNoteId)
    
    return { currentNote: current || null, otherNotes: others }
  }, [notes, currentNoteId])

  if (!open) return null

  return (
    <div className="absolute bottom-full left-0 right-0 mb-2 z-50">
      <div className="w-full max-w-[450px] rounded-lg border bg-popover text-popover-foreground shadow-xl">
        <Command className="rounded-lg" shouldFilter={true}>
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
              <>
                {/* Current Note Group */}
                {currentNote && (
                  <CommandGroup heading="Current Note">
                    <CommandItem
                      key={currentNote.id}
                      value={`${currentNote.name} ${currentNote.notebookName} ${currentNote.chapterName}`}
                      onSelect={() => onSelectNote(currentNote)}
                      className="flex items-start gap-3 py-3 cursor-pointer"
                    >
                      <FileText className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
                      <div className="flex-1 space-y-1 overflow-hidden">
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-medium leading-none truncate">{currentNote.name}</p>
                          <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
                            Current
                          </Badge>
                        </div>
                        <p className="text-xs text-muted-foreground truncate">
                          {currentNote.notebookName} → {currentNote.chapterName}
                        </p>
                        <p className="text-xs text-muted-foreground line-clamp-2">
                          {getPreviewText(currentNote.content, 100)}
                        </p>
                      </div>
                    </CommandItem>
                  </CommandGroup>
                )}
                
                {/* All Notes Group */}
                <CommandGroup heading={`${otherNotes.length} other note${otherNotes.length !== 1 ? 's' : ''}`}>
                  {otherNotes.map((note) => (
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
                          {getPreviewText(note.content, 100)}
                        </p>
                      </div>
                    </CommandItem>
                  ))}
                </CommandGroup>
              </>
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

