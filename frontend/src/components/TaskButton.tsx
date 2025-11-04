import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Loader2, CheckSquare, Plus, Eye, Sparkles, ListTodo } from 'lucide-react'
import { toast } from 'sonner'
import { getTasksForNote, generateTasksFromNote, createTaskBoard } from '@/utils/tasks'
import type { TaskBoard } from '@/types/backend'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface TaskButtonProps {
  noteId: string
  className?: string
}

export function TaskButton({ noteId, className }: TaskButtonProps) {
  const navigate = useNavigate()
  const [taskBoard, setTaskBoard] = useState<TaskBoard | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isGenerating, setIsGenerating] = useState(false)
  const [isCreatingManual, setIsCreatingManual] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showCreateDialog, setShowCreateDialog] = useState(false)

  // Check if tasks exist for this note
  useEffect(() => {
    const checkExistingTasks = async () => {
      if (!noteId) return

      setIsLoading(true)
      setError(null)

      try {
        const response = await getTasksForNote(noteId)
        setTaskBoard(response.taskBoard)
      } catch (err) {
        console.error('Failed to check existing tasks:', err)
        setError('Failed to load task information')
      } finally {
        setIsLoading(false)
      }
    }

    checkExistingTasks()
  }, [noteId])

  const handleGenerateTasks = async () => {
    if (!noteId) {
      toast.error('Note not available')
      return
    }

    setIsGenerating(true)
    setError(null)

    try {
      const response = await generateTasksFromNote(noteId)
      setTaskBoard(response.taskBoard)
      toast.success('Tasks generated successfully!')
      
      // Navigate to the newly created task board
      if (response.taskBoard?.id) {
        navigate(`/kanban/${response.taskBoard.id}`)
      }
    } catch (err: any) {
      console.error('Failed to generate tasks:', err)

      // Handle specific error cases
      if (err.response?.status === 409) {
        toast.error('Task board already exists for this note')
        // Refresh to get the existing task board
        try {
          const response = await getTasksForNote(noteId)
          setTaskBoard(response.taskBoard)
          if (!response.taskBoard) return
          navigate(`/kanban/${response.taskBoard.id}`)
        } catch (refreshErr) {
          console.error('Failed to refresh task board:', refreshErr)
        }
      } else {
        toast.error('Failed to generate tasks from note content')
        setError('Failed to generate tasks')
      }
    } finally {
      setIsGenerating(false)
    }
  }

  const handleViewTasks = () => {
    if (taskBoard) {
      navigate(`/kanban/${taskBoard.id}`)
    }
  }

  const handleCreateManual = async () => {
    if (!noteId) {
      toast.error('Note not available')
      return
    }

    setIsCreatingManual(true)
    setShowCreateDialog(false)

    try {
      const newTaskBoard = await createTaskBoard({
        name: 'Task Board',
        description: 'Manually created task board',
        noteId: noteId,
        isStandalone: false,
      })

      setTaskBoard(newTaskBoard)
      toast.success('Task board created successfully!')

      // Navigate to the newly created task board
      if (newTaskBoard?.id) {
        navigate(`/kanban/${newTaskBoard.id}`)
      }
    } catch (err: any) {
      console.error('Failed to create task board:', err)

      // Handle specific error cases
      if (err.response?.status === 409) {
        toast.error('Task board already exists for this note')
        // Refresh to get the existing task board
        try {
          const response = await getTasksForNote(noteId)
          setTaskBoard(response.taskBoard)
          if (response.taskBoard) {
            navigate(`/kanban/${response.taskBoard.id}`)
          }
        } catch (refreshErr) {
          console.error('Failed to refresh task board:', refreshErr)
        }
      } else {
        toast.error('Failed to create task board')
      }
    } finally {
      setIsCreatingManual(false)
    }
  }

  const handleAIGenerate = () => {
    setShowCreateDialog(false)
    handleGenerateTasks()
  }

  // Loading state
  if (isLoading) {
    return (
      <Button
        variant="outline"
        size="sm"
        disabled
        className={className}
      >
        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
        Loading...
      </Button>
    )
  }

  // Error state - show retry option
  if (error && !taskBoard) {
    return (
      <Button
        variant="outline"
        size="sm"
        onClick={() => window.location.reload()}
        className={className}
      >
        <CheckSquare className="mr-2 h-4 w-4" />
        Retry
      </Button>
    )
  }

  // Task board exists - show "View Tasks" button
  if (taskBoard) {
    const taskCount = taskBoard.tasks?.length || 0

    return (
      <Button
        variant="outline"
        size="sm"
        onClick={handleViewTasks}
        className={className}
      >
        <Eye className="mr-2 h-4 w-4" />
        View Tasks ({taskCount})
      </Button>
    )
  }

  // No task board exists - show "Create Tasks" button
  return (
    <>
      <Button
        variant="outline"
        size="sm"
        onClick={() => setShowCreateDialog(true)}
        disabled={isGenerating || isCreatingManual}
        className={className}
      >
        {isGenerating || isCreatingManual ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            {isGenerating ? 'Generating...' : 'Creating...'}
          </>
        ) : (
          <>
            <Plus className="mr-2 h-4 w-4" />
            Create Tasks
          </>
        )}
      </Button>

      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Create Task Board</DialogTitle>
            <DialogDescription>
              Choose how you'd like to create your task board
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-3 py-4">
            {/* AI Generated Option */}
            <button
              onClick={handleAIGenerate}
              className="group relative flex items-start gap-4 rounded-lg border border-border p-4 text-left transition-all hover:border-primary hover:bg-accent/50"
            >
              <div className="flex-1 space-y-1">
                <h3 className="font-semibold text-foreground">AI Generated Tasks</h3>
                <p className="text-sm text-muted-foreground">
                  Let AI analyze your note content and automatically generate relevant tasks
                </p>
              </div>
            </button>

            {/* Manual Option */}
            <button
              onClick={handleCreateManual}
              className="group relative flex items-start gap-4 rounded-lg border border-border p-4 text-left transition-all hover:border-primary hover:bg-accent/50"
            >
              <div className="flex-1 space-y-1">
                <h3 className="font-semibold text-foreground">Create Manually</h3>
                <p className="text-sm text-muted-foreground">
                  Start with an empty task board and add tasks yourself
                </p>
              </div>
            </button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}