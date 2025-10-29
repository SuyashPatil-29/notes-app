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

interface CreateNotebookDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (name: string) => Promise<void>
}

export function CreateNotebookDialog({
  open,
  onOpenChange,
  onSubmit,
}: CreateNotebookDialogProps) {
  const [notebookName, setNotebookName] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!notebookName.trim()) return

    setIsSubmitting(true)
    try {
      await onSubmit(notebookName)
      setNotebookName("")
      onOpenChange(false)
    } catch (error) {
      console.error("Error creating notebook:", error)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create Notebook</DialogTitle>
            <DialogDescription>
              Create a new notebook to organize your notes
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Notebook Name</Label>
              <Input
                id="name"
                placeholder="e.g., Work, Personal, Study Notes, etc."
                value={notebookName}
                onChange={(e) => setNotebookName(e.target.value)}
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
            <Button type="submit" disabled={isSubmitting || !notebookName.trim()}>
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Notebook
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

