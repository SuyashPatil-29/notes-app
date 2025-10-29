import { useState, useEffect } from "react"
import { useNavigate, useLocation } from "react-router-dom"
import { ChevronRight, Book, BookOpen, FileText, Plus, Pencil, Trash2, Loader, Eye } from "lucide-react"
import { moveChapter } from "@/utils/chapter"
import { moveNote } from "@/utils/notes"
import { toast } from "sonner"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { Notebook } from "@/types/backend"
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
  type DragOverEvent,
} from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import { useDraggable, useDroppable } from '@dnd-kit/core'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarMenuSubButton,
  SidebarRail,
} from "@/components/ui/sidebar"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from "@/components/ui/context-menu"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"

interface NotebookSidebarProps {
  notebooks: Notebook[] | undefined
  loading: boolean
  onCreateNotebook?: () => void
  onCreateChapter?: (notebookId: string) => void
  onRenameNotebook?: (notebookId: string) => void
  onDeleteNotebook?: (notebookId: string) => void
  onCreateNote?: (chapterId: string) => void
  onRenameChapter?: (chapterId: string) => void
  onDeleteChapter?: (chapterId: string) => void
  onRenameNote?: (noteId: string) => void
  onDeleteNote?: (noteId: string) => void
}

export function NotebookSidebar({
  notebooks,
  loading,
  onCreateNotebook,
  onCreateChapter,
  onRenameNotebook,
  onDeleteNotebook,
  onCreateNote,
  onRenameChapter,
  onDeleteChapter,
  onRenameNote,
  onDeleteNote
}: NotebookSidebarProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  
  // Extract noteId and chapterId from URL path
  // URL format: /:notebookId/:chapterId/:noteId
  const pathParts = location.pathname.split('/').filter(Boolean)
  const currentNoteId = pathParts.length === 3 ? pathParts[2] : undefined
  const currentChapterId = pathParts.length === 3 ? pathParts[1] : undefined

  // Controlled state for expanded notebooks and chapters
  const [expandedNotebooks, setExpandedNotebooks] = useState<Set<string>>(new Set())
  const [expandedChapters, setExpandedChapters] = useState<Set<string>>(new Set())
  
  // Drag and drop state with dnd-kit
  const [overId, setOverId] = useState<string | null>(null)
  
  // Configure sensors for dnd-kit
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8, // 8px of movement required before drag starts
      },
    })
  )
  
  // Optimistic chapter move mutation
  const moveChapterMutation = useMutation({
    mutationFn: ({ chapterId, targetNotebookId }: { chapterId: string, targetNotebookId: string }) => 
      moveChapter(chapterId, targetNotebookId),
    onMutate: async ({ chapterId, targetNotebookId }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['userNotebooks'] })
      
      // Snapshot the previous value
      const previousNotebooks = queryClient.getQueryData<Notebook[]>(['userNotebooks'])
      
      // Optimistically update
      queryClient.setQueryData<Notebook[]>(['userNotebooks'], (old) => {
        if (!old) return old
        
        const newNotebooks = JSON.parse(JSON.stringify(old)) as Notebook[]
        let movedChapter: any = null
        
        // Find and remove chapter from source notebook
        for (const notebook of newNotebooks) {
          const chapterIndex = notebook.chapters?.findIndex(ch => ch.id === chapterId)
          if (chapterIndex !== undefined && chapterIndex !== -1 && notebook.chapters) {
            movedChapter = notebook.chapters.splice(chapterIndex, 1)[0]
            break
          }
        }
        
        // Add chapter to target notebook
        if (movedChapter) {
          const targetNotebook = newNotebooks.find(nb => nb.id === targetNotebookId)
          if (targetNotebook) {
            if (!targetNotebook.chapters) {
              targetNotebook.chapters = []
            }
            targetNotebook.chapters.push(movedChapter)
          }
        }
        
        return newNotebooks
      })
      
      // If we're currently viewing a note in this chapter, navigate to the new location
      if (currentChapterId === chapterId && currentNoteId) {
        navigate(`/${targetNotebookId}/${chapterId}/${currentNoteId}`)
      }
      
      return { previousNotebooks }
    },
    onError: (err, _variables, context) => {
      // Rollback on error
      if (context?.previousNotebooks) {
        queryClient.setQueryData(['userNotebooks'], context.previousNotebooks)
      }
      toast.error('Failed to move chapter')
      console.error('Error moving chapter:', err)
    },
    onSuccess: () => {
      toast.success('Chapter moved successfully')
    },
    onSettled: () => {
      // Refetch to ensure consistency
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    }
  })
  
  // Optimistic note move mutation
  const moveNoteMutation = useMutation({
    mutationFn: ({ noteId, targetChapterId }: { noteId: string, targetChapterId: string }) => 
      moveNote(noteId, targetChapterId),
    onMutate: async ({ noteId, targetChapterId }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['userNotebooks'] })
      
      // Snapshot the previous value
      const previousNotebooks = queryClient.getQueryData<Notebook[]>(['userNotebooks'])
      
      // Find the target notebook ID for navigation
      let targetNotebookId: string | null = null
      if (previousNotebooks) {
        for (const notebook of previousNotebooks) {
          if (!notebook.chapters) continue
          const targetChapter = notebook.chapters.find(ch => ch.id === targetChapterId)
          if (targetChapter) {
            targetNotebookId = notebook.id
            break
          }
        }
      }
      
      // Optimistically update
      queryClient.setQueryData<Notebook[]>(['userNotebooks'], (old) => {
        if (!old) return old
        
        const newNotebooks = JSON.parse(JSON.stringify(old)) as Notebook[]
        let movedNote: any = null
        
        // Find and remove note from source chapter
        for (const notebook of newNotebooks) {
          if (!notebook.chapters) continue
          for (const chapter of notebook.chapters) {
            const noteIndex = chapter.notes?.findIndex(n => n.id === noteId)
            if (noteIndex !== undefined && noteIndex !== -1 && chapter.notes) {
              movedNote = chapter.notes.splice(noteIndex, 1)[0]
              break
            }
          }
          if (movedNote) break
        }
        
        // Add note to target chapter
        if (movedNote) {
          for (const notebook of newNotebooks) {
            if (!notebook.chapters) continue
            const targetChapter = notebook.chapters.find(ch => ch.id === targetChapterId)
            if (targetChapter) {
              if (!targetChapter.notes) {
                targetChapter.notes = []
              }
              targetChapter.notes.push(movedNote)
              break
            }
          }
        }
        
        return newNotebooks
      })
      
      // If we're currently viewing this note, navigate to the new location
      if (currentNoteId === noteId && targetNotebookId) {
        navigate(`/${targetNotebookId}/${targetChapterId}/${noteId}`)
      }
      
      return { previousNotebooks }
    },
    onError: (err, _variables, context) => {
      // Rollback on error
      if (context?.previousNotebooks) {
        queryClient.setQueryData(['userNotebooks'], context.previousNotebooks)
      }
      toast.error('Failed to move note')
      console.error('Error moving note:', err)
    },
    onSuccess: () => {
      toast.success('Note moved successfully')
    },
    onSettled: () => {
      // Refetch to ensure consistency
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    }
  })
  
  // Initialize and update expanded states when notebooks change
  useEffect(() => {
    if (notebooks && notebooks.length > 0) {
      // Keep all notebooks expanded
      setExpandedNotebooks(prev => {
        const newSet = new Set(prev)
        notebooks.forEach(notebook => newSet.add(notebook.id))
        return newSet
      })
      
      // Expand chapter if we're viewing a note
      if (currentChapterId) {
        setExpandedChapters(prev => {
          const newSet = new Set(prev)
          newSet.add(currentChapterId)
          return newSet
        })
      }
    }
  }, [notebooks, currentChapterId])
  
  // Toggle functions
  const toggleNotebook = (notebookId: string) => {
    setExpandedNotebooks(prev => {
      const newSet = new Set(prev)
      if (newSet.has(notebookId)) {
        newSet.delete(notebookId)
      } else {
        newSet.add(notebookId)
      }
      return newSet
    })
  }
  
  const toggleChapter = (chapterId: string) => {
    setExpandedChapters(prev => {
      const newSet = new Set(prev)
      if (newSet.has(chapterId)) {
        newSet.delete(chapterId)
      } else {
        newSet.add(chapterId)
      }
      return newSet
    })
  }
  
  // Helper to parse drag/drop IDs
  const parseId = (id: string) => {
    const [type, itemId] = id.split(':')
    return { type, itemId }
  }
  
  // dnd-kit handlers
  const handleDragStart = (_event: DragStartEvent) => {
    // Drag started
  }
  
  const handleDragOver = (event: DragOverEvent) => {
    const newOverId = event.over?.id as string || null
    setOverId(newOverId)
  }
  
  const handleDragEnd = (event: DragEndEvent) => {
    const activeId = event.active.id as string
    const overIdValue = event.over?.id as string
    
    if (!activeId || !overIdValue || activeId === overIdValue) {
      setOverId(null)
      return
    }
    
    const active = parseId(activeId)
    const over = parseId(overIdValue)
    
    // Move chapter to notebook
    if (active.type === 'chapter' && over.type === 'notebook') {
      moveChapterMutation.mutate({
        chapterId: active.itemId,
        targetNotebookId: over.itemId
      })
    }
    // Move note to chapter
    else if (active.type === 'note' && over.type === 'chapter') {
      moveNoteMutation.mutate({
        noteId: active.itemId,
        targetChapterId: over.itemId
      })
    }
    
    setOverId(null)
  }
  
  const handleDragCancel = () => {
    setOverId(null)
  }

  if (loading) {
    return (
      <Sidebar collapsible="icon">
        <SidebarHeader className="h-16 border-b">
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  <Loader className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">Notebooks</span>
                </div>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <div className="px-2 py-4 text-sm text-muted-foreground">
                Loading...
              </div>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarRail />
      </Sidebar>
    )
  }

  if (!notebooks || notebooks.length === 0) {
    return (
      <Sidebar collapsible="icon">
        <SidebarHeader className="h-16 border-b">
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  <Book className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">Notebooks</span>
                </div>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <ContextMenu>
          <ContextMenuTrigger asChild>
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupContent>
                  <div className="px-2 py-4 text-sm text-muted-foreground">
                    No notebooks yet. Create one to get started!
                  </div>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
          </ContextMenuTrigger>
          <ContextMenuContent>
            <ContextMenuItem onClick={() => onCreateNotebook?.()}>
              <Plus className="mr-2 h-4 w-4" />
              New Notebook
            </ContextMenuItem>
          </ContextMenuContent>
        </ContextMenu>
        <SidebarRail />
      </Sidebar>
    )
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <Sidebar collapsible="icon">
        <SidebarHeader className="h-16 border-b">
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  <Book className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">Notebooks</span>
          </div>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <ContextMenu>
          <ContextMenuTrigger asChild>
        <SidebarContent>
            <SidebarGroup>
              <SidebarGroupContent>
                <SidebarMenu>
          {notebooks.map((notebook) => (
                        <Collapsible
                          key={notebook.id}
                          asChild
                          open={expandedNotebooks.has(notebook.id)}
                          onOpenChange={() => toggleNotebook(notebook.id)}
                          className="group/collapsible"
                        >
                          <SidebarMenuItem>
                            <ContextMenu>
                              <ContextMenuTrigger asChild>
                                <DroppableNotebook 
                                  id={notebook.id}
                                  isOver={overId === `notebook:${notebook.id}`}
                                >
                                  <CollapsibleTrigger asChild>
                                    <SidebarMenuButton tooltip={notebook.name}>
                                      <Book />
                                      <span>{notebook.name}</span>
                                      <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                    </SidebarMenuButton>
                                  </CollapsibleTrigger>
                                </DroppableNotebook>
                              </ContextMenuTrigger>
                          <ContextMenuContent>
                            <ContextMenuItem onClick={() => navigate(`/${notebook.id}`)}>
                              <Eye className="mr-2 h-4 w-4" />
                              View Notebook
                            </ContextMenuItem>
                            <ContextMenuItem onClick={() => onCreateChapter?.(notebook.id)}>
                              <Plus className="mr-2 h-4 w-4" />
                              Create Chapter
                            </ContextMenuItem>
                            <ContextMenuSeparator />
                            <ContextMenuItem onClick={() => onRenameNotebook?.(notebook.id)}>
                              <Pencil className="mr-2 h-4 w-4" />
                              Rename Notebook
                            </ContextMenuItem>
                            <ContextMenuItem
                              onClick={() => onDeleteNotebook?.(notebook.id)}
                              className="text-destructive focus:text-destructive"
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete Notebook
                            </ContextMenuItem>
                          </ContextMenuContent>
                        </ContextMenu>
                        <CollapsibleContent>
                          <SidebarMenuSub>
                            {notebook.chapters?.map((chapter) => (
                              <Collapsible
                                key={chapter.id}
                                asChild
                                open={expandedChapters.has(chapter.id) || currentChapterId === chapter.id}
                                onOpenChange={() => toggleChapter(chapter.id)}
                                className="group/chapter-collapsible"
                              >
                                <SidebarMenuSubItem>
                                  <DroppableChapter 
                                    id={chapter.id}
                                    isOver={overId === `chapter:${chapter.id}`}
                                  >
                                    <DraggableChapter id={chapter.id}>
                                      <ContextMenu>
                                        <ContextMenuTrigger asChild>
                                          <CollapsibleTrigger asChild>
                                            <SidebarMenuSubButton>
                                              <BookOpen />
                                              <span>{chapter.name}</span>
                                              <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/chapter-collapsible:rotate-90" />
                                            </SidebarMenuSubButton>
                                          </CollapsibleTrigger>
                                        </ContextMenuTrigger>
                                      <ContextMenuContent>
                                        <ContextMenuItem onClick={() => navigate(`/${notebook.id}/${chapter.id}`)}>
                                          <Eye className="mr-2 h-4 w-4" />
                                          View Chapter
                                        </ContextMenuItem>
                                        <ContextMenuItem onClick={() => onCreateNote?.(chapter.id)}>
                                          <Plus className="mr-2 h-4 w-4" />
                                          Create Note
                                        </ContextMenuItem>
                                        <ContextMenuSeparator />
                                        <ContextMenuItem onClick={() => onRenameChapter?.(chapter.id)}>
                                          <Pencil className="mr-2 h-4 w-4" />
                                          Rename Chapter
                                        </ContextMenuItem>
                                        <ContextMenuItem
                                          onClick={() => onDeleteChapter?.(chapter.id)}
                                          className="text-destructive focus:text-destructive"
                                        >
                                          <Trash2 className="mr-2 h-4 w-4" />
                                          Delete Chapter
                                        </ContextMenuItem>
                                        </ContextMenuContent>
                                      </ContextMenu>
                                    </DraggableChapter>
                                  </DroppableChapter>
                                  <CollapsibleContent>
                                    <SidebarMenuSub>
                                      {chapter.notes?.map((note) => {
                                        const isActive = currentNoteId === note.id
                                        
                                        return (
                                        <SidebarMenuSubItem key={note.id}>
                                          <DraggableNote id={note.id}>
                                            <TooltipProvider delayDuration={300}>
                                              <Tooltip>
                                                <ContextMenu>
                                                  <ContextMenuTrigger asChild>
                                                    <TooltipTrigger asChild>
                                                      <SidebarMenuSubButton
                                                        asChild
                                                        isActive={isActive}
                                                      >
                                                        <button 
                                                          onClick={() => navigate(`/${notebook.id}/${chapter.id}/${note.id}`)}
                                                          className={`w-full min-w-0 ${isActive ? 'font-bold bg-accent' : ''}`}
                                                        >
                                                          <FileText className={`shrink-0 ${isActive ? 'text-primary' : ''}`} />
                                                          <span className={`truncate block ${isActive ? 'text-foreground' : ''}`}>{note.name}</span>
                                                        </button>
                                                      </SidebarMenuSubButton>
                                                    </TooltipTrigger>
                                                  </ContextMenuTrigger>
                                                  <ContextMenuContent>
                                                    <ContextMenuItem onClick={() => onRenameNote?.(note.id)}>
                                                      <Pencil className="mr-2 h-4 w-4" />
                                                      Rename Note
                                                    </ContextMenuItem>
                                                    <ContextMenuItem
                                                      onClick={() => onDeleteNote?.(note.id)}
                                                      className="text-destructive focus:text-destructive"
                                                    >
                                                      <Trash2 className="mr-2 h-4 w-4" />
                                                      Delete Note
                                                    </ContextMenuItem>
                                                  </ContextMenuContent>
                                                </ContextMenu>
                                                <TooltipContent side="right" align="start" sideOffset={5}>
                                                  <p className="text-sm">{note.name}</p>
                                                </TooltipContent>
                                              </Tooltip>
                                            </TooltipProvider>
                                          </DraggableNote>
                                        </SidebarMenuSubItem>
                                        )
                                      })}
                                    </SidebarMenuSub>
                                  </CollapsibleContent>
                                </SidebarMenuSubItem>
                              </Collapsible>
                            ))}
                          </SidebarMenuSub>
                        </CollapsibleContent>
                      </SidebarMenuItem>
                    </Collapsible>
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
      </SidebarContent>
        </ContextMenuTrigger>
        <ContextMenuContent>
          <ContextMenuItem onClick={() => onCreateNotebook?.()}>
            <Plus className="mr-2 h-4 w-4" />
            New Notebook
          </ContextMenuItem>
        </ContextMenuContent>
      </ContextMenu>
      <SidebarRail />
    </Sidebar>
    </DndContext>
  )
}

// Draggable Chapter Component - drag by entire element
function DraggableChapter({ id, children }: { id: string, children: React.ReactNode }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: `chapter:${id}`,
  })
  
  const style = {
    transform: CSS.Translate.toString(transform),
    opacity: isDragging ? 0.5 : 1,
    cursor: 'grab',
  }
  
  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      className="active:cursor-grabbing"
    >
      {children}
    </div>
  )
}

// Droppable Notebook Component
function DroppableNotebook({ id, isOver, children }: { id: string, isOver: boolean, children: React.ReactNode }) {
  const { setNodeRef } = useDroppable({
    id: `notebook:${id}`,
  })
  
  return (
    <div
      ref={setNodeRef}
      className={`transition-colors ${isOver ? 'bg-accent rounded-md ring-2 ring-primary' : ''}`}
    >
      {children}
    </div>
  )
}

// Draggable Note Component - drag by entire element
function DraggableNote({ id, children }: { id: string, children: React.ReactNode }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: `note:${id}`,
  })
  
  const style = {
    transform: CSS.Translate.toString(transform),
    opacity: isDragging ? 0.5 : 1,
    cursor: 'grab',
  }
  
  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      className="active:cursor-grabbing"
    >
      {children}
    </div>
  )
}

// Droppable Chapter Component
function DroppableChapter({ id, isOver, children }: { id: string, isOver: boolean, children: React.ReactNode }) {
  const { setNodeRef } = useDroppable({
    id: `chapter:${id}`,
  })
  
  return (
    <div
      ref={setNodeRef}
      className={`transition-colors ${isOver ? 'bg-accent rounded-md ring-2 ring-primary' : ''}`}
    >
      {children}
    </div>
  )
}
