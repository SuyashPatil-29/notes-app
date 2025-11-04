import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Loader2 } from "lucide-react"

interface RenameDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (newName: string) => Promise<void>
  title: string
  description?: string
  currentName?: string
  placeholder?: string
  itemType: "Notebook" | "Chapter" | "Note" | "Task Board"
}

export function RenameDialog({
  open,
  onOpenChange,
  onSubmit,
  title,
  description,
  currentName = "",
  placeholder,
  itemType,
}: RenameDialogProps) {
  const [name, setName] = useState(currentName)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setName(currentName)
    }
  }, [open, currentName])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || name === currentName) return

    setIsSubmitting(true)
    try {
      await onSubmit(name)
      onOpenChange(false)
    } catch (error) {
      console.error(`Error renaming ${itemType.toLowerCase()}:`, error)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            {description && <DialogDescription>{description}</DialogDescription>}
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">{itemType} Name</Label>
              <Input
                id="name"
                placeholder={placeholder || `Enter ${itemType.toLowerCase()} name`}
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isSubmitting}
                autoFocus
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting || !name.trim() || name === currentName}
            >
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Rename
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

