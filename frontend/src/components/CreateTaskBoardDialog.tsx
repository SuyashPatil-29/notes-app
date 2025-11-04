import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { createTaskBoard } from '@/utils/tasks'
import { validateTaskBoardName } from '@/utils/tasks'
import type { TaskBoard } from '@/types/backend'

interface CreateTaskBoardDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onTaskBoardCreated?: (taskBoard: TaskBoard) => void
}

export function CreateTaskBoardDialog({
  open,
  onOpenChange,
  onTaskBoardCreated,
}: CreateTaskBoardDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const [nameError, setNameError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    // Validate name
    const nameValidationError = validateTaskBoardName(name)
    if (nameValidationError) {
      setNameError(nameValidationError)
      return
    }

    setIsCreating(true)
    setNameError(null)

    try {
      const newTaskBoard = await createTaskBoard({
        name: name.trim(),
        description: description.trim() || undefined,
        isStandalone: true,
      })

      toast.success('Task board created successfully!')
      onTaskBoardCreated?.(newTaskBoard)
      handleClose()
    } catch (error) {
      console.error('Failed to create task board:', error)
      toast.error('Failed to create task board')
    } finally {
      setIsCreating(false)
    }
  }

  const handleClose = () => {
    setName('')
    setDescription('')
    setNameError(null)
    onOpenChange(false)
  }

  const handleNameChange = (value: string) => {
    setName(value)
    if (nameError) {
      setNameError(null)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create New Task Board</DialogTitle>
          <DialogDescription>
            Create a standalone Kanban board to organize your tasks.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="Enter board name"
              className={nameError ? 'border-destructive' : ''}
            />
            {nameError && (
              <p className="text-sm text-destructive">{nameError}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter board description (optional)"
              rows={3}
            />
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={isCreating}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isCreating || !name.trim()}
            >
              {isCreating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                'Create Board'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}