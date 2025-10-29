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

interface CreateChapterDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (name: string) => Promise<void>
  notebookName?: string
}

export function CreateChapterDialog({
  open,
  onOpenChange,
  onSubmit,
  notebookName,
}: CreateChapterDialogProps) {
  const [chapterName, setChapterName] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!chapterName.trim()) return

    setIsSubmitting(true)
    try {
      await onSubmit(chapterName)
      setChapterName("")
      onOpenChange(false)
    } catch (error) {
      console.error("Error creating chapter:", error)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create Chapter</DialogTitle>
            <DialogDescription>
              {notebookName
                ? `Add a new chapter to "${notebookName}"`
                : "Add a new chapter to your notebook"}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Chapter Name</Label>
              <Input
                id="name"
                placeholder="e.g., Introduction, Chapter 1, etc."
                value={chapterName}
                onChange={(e) => setChapterName(e.target.value)}
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
            <Button type="submit" disabled={isSubmitting || !chapterName.trim()}>
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Chapter
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

