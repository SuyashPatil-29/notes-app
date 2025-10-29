import { useState } from "react"
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

interface CreateNoteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (name: string) => Promise<void>
  chapterName?: string
}

export function CreateNoteDialog({
  open,
  onOpenChange,
  onSubmit,
  chapterName,
}: CreateNoteDialogProps) {
  const [noteName, setnoteName] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!noteName.trim()) return

    setIsSubmitting(true)
    try {
      await onSubmit(noteName)
      setnoteName("")
      onOpenChange(false)
    } catch (error) {
      console.error("Error creating note:", error)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create Note</DialogTitle>
            <DialogDescription>
              {chapterName
                ? `Add a new note to "${chapterName}"`
                : "Add a new note to your chapter"}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Note Name</Label>
              <Input
                id="name"
                placeholder="e.g., My First Note, TODO List, etc."
                value={noteName}
                onChange={(e) => setnoteName(e.target.value)}
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
            <Button type="submit" disabled={isSubmitting || !noteName.trim()}>
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Note
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

