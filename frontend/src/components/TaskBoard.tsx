import { useState, useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { Loader2, ArrowLeft, Settings, Trash2, Edit, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { toast } from 'sonner'
import { KanbanBoard } from './KanbanBoard'
import { getTaskBoard, updateTaskBoard, deleteTaskBoard } from '@/utils/tasks'
import type { Task, TaskBoard as TaskBoardType } from '@/types/backend'

interface TaskBoardProps {
  boardId: string
  onNavigateBack?: () => void
  onBoardDeleted?: () => void
  className?: string
}

export function TaskBoard({ 
  boardId, 
  onNavigateBack, 
  onBoardDeleted,
  className = "" 
}: TaskBoardProps) {
  const queryClient = useQueryClient()
  const [isEditingBoard, setIsEditingBoard] = useState(false)
  const [editName, setEditName] = useState("")
  const [editDescription, setEditDescription] = useState("")
  const [isUpdating, setIsUpdating] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const {
    data: taskBoard,
    isLoading,
    error,
    refetch
  } = useQuery({
    queryKey: ['taskBoard', boardId],
    queryFn: () => getTaskBoard(boardId),
    enabled: !!boardId,
    refetchOnWindowFocus: false,
    retry: 1,
  })

  // Initialize edit form when taskBoard loads
  useEffect(() => {
    if (taskBoard) {
      setEditName(taskBoard.name)
      setEditDescription(taskBoard.description || "")
    }
  }, [taskBoard])

  const handleTaskUpdate = (updatedTask: Task) => {
    // Optimistically update the task board in cache
    queryClient.setQueryData(['taskBoard', boardId], (oldData: TaskBoardType | undefined) => {
      if (!oldData) return oldData
      
      return {
        ...oldData,
        tasks: oldData.tasks.map(task => 
          task.id === updatedTask.id ? updatedTask : task
        )
      }
    })

    // Don't invalidate immediately - let the backend sync complete first
    // The child component will handle the backend update
  }

  const handleTaskDelete = (taskId: string) => {
    // Optimistically remove the task from cache
    queryClient.setQueryData(['taskBoard', boardId], (oldData: TaskBoardType | undefined) => {
      if (!oldData) return oldData
      
      return {
        ...oldData,
        tasks: oldData.tasks.filter(task => task.id !== taskId)
      }
    })

    // Invalidate userTaskBoards to update the task count in sidebar
    queryClient.invalidateQueries({ queryKey: ['userTaskBoards'] })
  }

  const handleTaskCreate = (newTask: Task) => {
    // Optimistically add the task to cache
    queryClient.setQueryData(['taskBoard', boardId], (oldData: TaskBoardType | undefined) => {
      if (!oldData) return oldData
      
      // If this is a temporary task, replace any existing temp tasks
      const isTemp = newTask.id.startsWith('temp-')
      let tasks = oldData.tasks
      
      if (!isTemp) {
        // Real task from backend - remove any temp tasks and add this one
        tasks = tasks.filter(t => !t.id.startsWith('temp-'))
      }
      
      // Check if task already exists (avoid duplicates)
      const existingIndex = tasks.findIndex(t => t.id === newTask.id)
      if (existingIndex >= 0) {
        // Update existing task
        tasks = tasks.map(t => t.id === newTask.id ? newTask : t)
      } else {
        // Add new task
        tasks = [...tasks, newTask]
      }
      
      return {
        ...oldData,
        tasks
      }
    })

    // Invalidate userTaskBoards to update the task count in sidebar
    queryClient.invalidateQueries({ queryKey: ['userTaskBoards'] })
  }

  const handleUpdateBoard = async () => {
    if (!taskBoard || !editName.trim()) {
      toast.error("Board name is required")
      return
    }

    setIsUpdating(true)

    try {
      const updatedBoard = await updateTaskBoard(boardId, {
        name: editName.trim(),
        description: editDescription.trim() || undefined,
      })

      // Update cache
      queryClient.setQueryData(['taskBoard', boardId], updatedBoard)
      queryClient.invalidateQueries({ queryKey: ['userTaskBoards'] })

      setIsEditingBoard(false)
      toast.success("Board updated successfully")
    } catch (error) {
      console.error('Failed to update board:', error)
      toast.error("Failed to update board")
    } finally {
      setIsUpdating(false)
    }
  }

  const handleDeleteBoard = async () => {
    if (!taskBoard) return

    const confirmMessage = taskBoard.tasks.length > 0
      ? `Are you sure you want to delete "${taskBoard.name}"? This will also delete all ${taskBoard.tasks.length} tasks in this board.`
      : `Are you sure you want to delete "${taskBoard.name}"?`

    if (!confirm(confirmMessage)) return

    setIsDeleting(true)

    try {
      await deleteTaskBoard(boardId)

      // Clear cache
      queryClient.removeQueries({ queryKey: ['taskBoard', boardId] })
      queryClient.invalidateQueries({ queryKey: ['userTaskBoards'] })

      toast.success("Board deleted successfully")
      onBoardDeleted?.()
    } catch (error) {
      console.error('Failed to delete board:', error)
      toast.error("Failed to delete board")
    } finally {
      setIsDeleting(false)
    }
  }

  const handleRefresh = () => {
    refetch()
    toast.success("Board refreshed")
  }

  if (isLoading) {
    return (
      <div className={`flex flex-col h-full ${className}`}>
        <div className="flex items-center justify-center h-full">
          <div className="text-center space-y-4">
            <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
            <p className="text-sm text-muted-foreground">Loading task board...</p>
          </div>
        </div>
      </div>
    )
  }

  if (error || !taskBoard) {
    return (
      <div className={`flex flex-col h-full ${className}`}>
        <div className="flex items-center justify-center h-full">
          <div className="text-center space-y-4">
            <p className="text-lg text-destructive">Failed to load task board</p>
            <div className="space-x-2">
              <Button variant="outline" onClick={handleRefresh}>
                Try Again
              </Button>
              {onNavigateBack && (
                <Button variant="ghost" onClick={onNavigateBack}>
                  <ArrowLeft className="mr-2 h-4 w-4" />
                  Go Back
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={`flex flex-col h-full ${className}`}>
      {/* In-Page Header */}
      <div className="border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
        <div className="flex items-center justify-between p-4">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <div className="min-w-0 flex-1">
              <h2 className="text-xl font-semibold truncate">{taskBoard.name}</h2>
              {taskBoard.description && (
                <p className="text-sm text-muted-foreground mt-0.5 line-clamp-1">
                  {taskBoard.description}
                </p>
              )}
              <div className="flex items-center gap-3 mt-1.5 text-xs text-muted-foreground">
                <span>{taskBoard.tasks.length} tasks</span>
                {taskBoard.noteId && taskBoard.note && (
                  <Link to={`/${taskBoard.note.chapter.notebookId}/${taskBoard.note.chapterId}/${taskBoard.noteId}`}>
                    <Badge variant="secondary" className="hover:bg-secondary/80 transition-colors cursor-pointer">
                      View Note
                    </Badge>
                  </Link>
                )}
                {taskBoard.isStandalone && (
                  <Badge variant="outline">
                    Standalone Board
                  </Badge>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <Button variant="outline" size="sm" onClick={handleRefresh}>
              <RotateCcw className="h-4 w-4" />
            </Button>
            
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <Settings className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => setIsEditingBoard(true)}>
                  <Edit className="mr-2 h-4 w-4" />
                  Edit Board
                </DropdownMenuItem>
                <DropdownMenuItem 
                  onClick={handleDeleteBoard}
                  className="text-destructive"
                  disabled={isDeleting}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete Board
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>

      {/* Kanban Board */}
      <div className="flex-1 overflow-hidden">
        <KanbanBoard
          taskBoard={taskBoard}
          onTaskUpdate={handleTaskUpdate}
          onTaskDelete={handleTaskDelete}
          onTaskCreate={handleTaskCreate}
          className="h-full"
        />
      </div>

      {/* Edit Board Dialog */}
      <Dialog open={isEditingBoard} onOpenChange={setIsEditingBoard}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Task Board</DialogTitle>
            <DialogDescription>
              Update the name and description of your task board.
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">Name</label>
              <Input
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                placeholder="Board name"
                className="mt-1"
              />
            </div>
            
            <div>
              <label className="text-sm font-medium">Description</label>
              <Textarea
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
                placeholder="Board description (optional)"
                rows={3}
                className="mt-1"
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsEditingBoard(false)}
              disabled={isUpdating}
            >
              Cancel
            </Button>
            <Button
              onClick={handleUpdateBoard}
              disabled={isUpdating || !editName.trim()}
            >
              {isUpdating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Updating...
                </>
              ) : (
                "Update Board"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}