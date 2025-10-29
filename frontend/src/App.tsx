import { useState } from 'react'
import { Routes, Route } from 'react-router-dom'
import { getUserNotebooks, createNotebook, updateNotebook, deleteNotebook } from '@/utils/notebook'
import { createChapter, updateChapter, deleteChapter } from '@/utils/chapter'
import { createNote, updateNote, deleteNote } from '@/utils/notes'
import { useUser } from '@/hooks/auth'
import { toast, Toaster } from 'sonner'
import { LeftSidebarContent } from '@/components/left-sidebar-content'
import { RightSidebarContent } from '@/components/right-sidebar-content'
import { Dashboard } from '@/components/Dashboard'
import { NotebookView } from '@/components/NotebookView'
import { ChapterView } from '@/components/ChapterView'
import { NoteEditor } from '@/components/NoteEditor'
import { CreateNotebookDialog } from '@/components/SidebarDialogs/CreateNotebookDialog'
import { CreateChapterDialog } from '@/components/SidebarDialogs/CreateChapterDialog'
import { CreateNoteDialog } from '@/components/SidebarDialogs/CreateNoteDialog'
import { RenameDialog } from '@/components/SidebarDialogs/RenameDialog'
import { DeleteConfirmDialog } from '@/components/SidebarDialogs/DeleteConfirmDialog'
import { LeftSidebarProvider, LeftSidebarInset } from '@/components/ui/left-sidebar'
import { RightSidebarProvider, RightSidebarInset } from '@/components/ui/right-sidebar'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import type { Chapter, Notes, Notebook } from '@/types/backend'
import { createId } from '@paralleldrive/cuid2'
import CommandMenu from './components/cmdk'

function App() {
  const { user, loading: userLoading } = useUser()
  const queryClient = useQueryClient()
  const [createNotebookDialog, setCreateNotebookDialog] = useState(false)

  const [createChapterDialog, setCreateChapterDialog] = useState<{
    open: boolean
    notebookId: string | null
    notebookName?: string
  }>({
    open: false,
    notebookId: null,
  })

  const [createNoteDialog, setCreateNoteDialog] = useState<{
    open: boolean
    chapterId: string | null
    chapterName?: string
  }>({
    open: false,
    chapterId: null,
  })

  const [renameNotebookDialog, setRenameNotebookDialog] = useState<{
    open: boolean
    notebookId: string | null
    currentName: string
  }>({
    open: false,
    notebookId: null,
    currentName: "",
  })

  const [deleteNotebookDialog, setDeleteNotebookDialog] = useState<{
    open: boolean
    notebookId: string | null
    notebookName: string
  }>({
    open: false,
    notebookId: null,
    notebookName: "",
  })

  const [renameChapterDialog, setRenameChapterDialog] = useState<{
    open: boolean
    chapterId: string | null
    currentName: string
  }>({
    open: false,
    chapterId: null,
    currentName: "",
  })

  const [deleteChapterDialog, setDeleteChapterDialog] = useState<{
    open: boolean
    chapterId: string | null
    chapterName: string
  }>({
    open: false,
    chapterId: null,
    chapterName: "",
  })

  const [renameNoteDialog, setRenameNoteDialog] = useState<{
    open: boolean
    noteId: string | null
    currentName: string
  }>({
    open: false,
    noteId: null,
    currentName: "",
  })

  const [deleteNoteDialog, setDeleteNoteDialog] = useState<{
    open: boolean
    noteId: string | null
    noteName: string
  }>({
    open: false,
    noteId: null,
    noteName: "",
  })

  const {data : userNotebooks, isLoading: userNotebooksLoading} = useQuery({
    queryKey: ['userNotebooks'],
    queryFn: getUserNotebooks,
    refetchOnWindowFocus: false,
    enabled: !!user,
  })

  // Context menu handlers
  const handleCreateNotebook = () => {
    setCreateNotebookDialog(true)
  }

  const handleCreateNotebookSubmit = async (notebookName: string) => {
    if (!user) {
      toast.error("Please login first")
      return
    }

    try {
      const newNotebook: Notebook = {
        id: createId(),
        name: notebookName,
        userId: user.id,
        chapters: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }

      await createNotebook(newNotebook)
      toast.success("Notebook created successfully!")
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to create notebook")
      }
      throw error
    }
  }

  const handleCreateChapter = (notebookId: string) => {
    const notebook = userNotebooks?.find((n) => n.id === notebookId)
    setCreateChapterDialog({
      open: true,
      notebookId,
      notebookName: notebook?.name,
    })
  }

  const handleCreateChapterSubmit = async (chapterName: string) => {
    if (!createChapterDialog.notebookId) return

    try {
      const newChapter: Chapter = {
        id: createId(),
        name: chapterName,
        notebookId: createChapterDialog.notebookId,
        notebook: {} as any, // Will be populated by backend
        notes: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }

      await createChapter(newChapter)
      toast.success("Chapter created successfully!")
      
      // Invalidate and refetch notebooks to get the updated data
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to create chapter")
      }
      throw error
    }
  }

  const handleCreateNote = (chapterId: string) => {
    // Find chapter name for better UX
    let chapterName = ""
    for (const notebook of userNotebooks || []) {
      const chapter = notebook.chapters?.find((c) => c.id === chapterId)
      if (chapter) {
        chapterName = chapter.name
        break
      }
    }
    setCreateNoteDialog({
      open: true,
      chapterId,
      chapterName,
    })
  }

  const handleCreateNoteSubmit = async (noteName: string) => {
    if (!createNoteDialog.chapterId) return

    try {
      const newNote: Notes = {
        id: createId(),
        name: noteName,
        content: "",
        chapterId: createNoteDialog.chapterId,
        chapter: {} as any,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }

      await createNote(newNote)
      toast.success("Note created successfully!")
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to create note")
      }
      throw error
    }
  }

  const handleRenameNotebook = (notebookId: string) => {
    const notebook = userNotebooks?.find((n) => n.id === notebookId)
    if (!notebook) return
    setRenameNotebookDialog({
      open: true,
      notebookId,
      currentName: notebook.name,
    })
  }

  const handleRenameNotebookSubmit = async (newName: string) => {
    if (!renameNotebookDialog.notebookId) return

    try {
      await updateNotebook(renameNotebookDialog.notebookId, { name: newName })
      toast.success("Notebook renamed successfully!")
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to rename notebook")
      }
      throw error
    }
  }

  const handleDeleteNotebook = (notebookId: string) => {
    const notebook = userNotebooks?.find((n) => n.id === notebookId)
    if (!notebook) return
    setDeleteNotebookDialog({
      open: true,
      notebookId,
      notebookName: notebook.name,
    })
  }

  const handleDeleteNotebookConfirm = async () => {
    if (!deleteNotebookDialog.notebookId) return

    try {
      await deleteNotebook(deleteNotebookDialog.notebookId)
      toast.success("Notebook deleted successfully!")
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to delete notebook")
      }
      throw error
    }
  }

  const handleRenameChapter = (chapterId: string) => {
    // Find chapter
    let chapter = null
    for (const notebook of userNotebooks || []) {
      const found = notebook.chapters?.find((c) => c.id === chapterId)
      if (found) {
        chapter = found
        break
      }
    }
    if (!chapter) return
    setRenameChapterDialog({
      open: true,
      chapterId,
      currentName: chapter.name,
    })
  }

  const handleRenameChapterSubmit = async (newName: string) => {
    if (!renameChapterDialog.chapterId) return

    try {
      await updateChapter(renameChapterDialog.chapterId, { name: newName })
      toast.success("Chapter renamed successfully!")
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to rename chapter")
      }
      throw error
    }
  }

  const handleDeleteChapter = (chapterId: string) => {
    // Find chapter name
    let chapterName = ""
    for (const notebook of userNotebooks || []) {
      const chapter = notebook.chapters?.find((c) => c.id === chapterId)
      if (chapter) {
        chapterName = chapter.name
        break
      }
    }
    setDeleteChapterDialog({
      open: true,
      chapterId,
      chapterName,
    })
  }

  const handleDeleteChapterConfirm = async () => {
    if (!deleteChapterDialog.chapterId) return

    try {
      await deleteChapter(deleteChapterDialog.chapterId)
      toast.success("Chapter deleted successfully!")
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to delete chapter")
      }
      throw error
    }
  }

  const handleRenameNote = (noteId: string) => {
    // Find note
    let note = null
    for (const notebook of userNotebooks || []) {
      for (const chapter of notebook.chapters || []) {
        const found = chapter.notes?.find((n) => n.id === noteId)
        if (found) {
          note = found
          break
        }
      }
      if (note) break
    }
    if (!note) return
    setRenameNoteDialog({
      open: true,
      noteId,
      currentName: note.name,
    })
  }

  const handleRenameNoteSubmit = async (newName: string) => {
    if (!renameNoteDialog.noteId) return

    try {
      await updateNote(renameNoteDialog.noteId, { name: newName })
      toast.success("Note renamed successfully!")
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to rename note")
      }
      throw error
    }
  }

  const handleDeleteNote = (noteId: string) => {
    // Find note name
    let noteName = ""
    for (const notebook of userNotebooks || []) {
      for (const chapter of notebook.chapters || []) {
        const note = chapter.notes?.find((n) => n.id === noteId)
        if (note) {
          noteName = note.name
          break
        }
      }
      if (noteName) break
    }
    setDeleteNoteDialog({
      open: true,
      noteId,
      noteName,
    })
  }

  const handleDeleteNoteConfirm = async () => {
    if (!deleteNoteDialog.noteId) return

    try {
      await deleteNote(deleteNoteDialog.noteId)
      toast.success("Note deleted successfully!")
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
    } catch (error: any) {
      if (error.response) {
        const status = error.response.status
        const message = error.response.data?.message || error.response.data?.error || error.message
        toast.error(`Error ${status}: ${message}`)
      } else {
        toast.error("Failed to delete note")
      }
      throw error
    }
  }

  if (userLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-muted-foreground text-lg">Loading...</div>
      </div>
    )
  }

  return (
    <>
      <Toaster />
      <LeftSidebarProvider defaultOpen>
        <LeftSidebarContent 
          notebooks={userNotebooks}
          loading={userNotebooksLoading}
          onCreateNotebook={handleCreateNotebook}
          onCreateChapter={handleCreateChapter}
          onRenameNotebook={handleRenameNotebook}
          onDeleteNotebook={handleDeleteNotebook}
          onCreateNote={handleCreateNote}
          onRenameChapter={handleRenameChapter}
          onDeleteChapter={handleDeleteChapter}
          onRenameNote={handleRenameNote}
          onDeleteNote={handleDeleteNote}
        />
        <LeftSidebarInset>
          <RightSidebarProvider defaultOpen={false}>
            <RightSidebarInset>
              <Routes>
                <Route path="/" element={<Dashboard user={user} />} />
                <Route path="/:notebookId" element={<NotebookView user={user} onCreateChapter={handleCreateChapter} />} />
                <Route path="/:notebookId/:chapterId" element={<ChapterView user={user} onCreateNote={handleCreateNote} />} />
                <Route path="/:notebookId/:chapterId/:noteId" element={<NoteEditor user={user} />} />
              </Routes>
            </RightSidebarInset>
            <RightSidebarContent />
          </RightSidebarProvider>
        </LeftSidebarInset>
      </LeftSidebarProvider>

      <CreateNotebookDialog
        open={createNotebookDialog}
        onOpenChange={setCreateNotebookDialog}
        onSubmit={handleCreateNotebookSubmit}
      />

      <CreateChapterDialog
        open={createChapterDialog.open}
        onOpenChange={(open) =>
          setCreateChapterDialog({ ...createChapterDialog, open })
        }
        onSubmit={handleCreateChapterSubmit}
        notebookName={createChapterDialog.notebookName}
      />

      <CreateNoteDialog
        open={createNoteDialog.open}
        onOpenChange={(open) =>
          setCreateNoteDialog({ ...createNoteDialog, open })
        }
        onSubmit={handleCreateNoteSubmit}
        chapterName={createNoteDialog.chapterName}
      />

      <RenameDialog
        open={renameNotebookDialog.open}
        onOpenChange={(open) =>
          setRenameNotebookDialog({ ...renameNotebookDialog, open })
        }
        onSubmit={handleRenameNotebookSubmit}
        title="Rename Notebook"
        currentName={renameNotebookDialog.currentName}
        itemType="Notebook"
      />

      <DeleteConfirmDialog
        open={deleteNotebookDialog.open}
        onOpenChange={(open) =>
          setDeleteNotebookDialog({ ...deleteNotebookDialog, open })
        }
        onConfirm={handleDeleteNotebookConfirm}
        title="Delete Notebook"
        description="Are you sure you want to delete this notebook? This action cannot be undone and will delete all chapters and notes within this notebook."
        itemName={deleteNotebookDialog.notebookName}
        itemType="Notebook"
      />

      <RenameDialog
        open={renameChapterDialog.open}
        onOpenChange={(open) =>
          setRenameChapterDialog({ ...renameChapterDialog, open })
        }
        onSubmit={handleRenameChapterSubmit}
        title="Rename Chapter"
        currentName={renameChapterDialog.currentName}
        itemType="Chapter"
      />

      <DeleteConfirmDialog
        open={deleteChapterDialog.open}
        onOpenChange={(open) =>
          setDeleteChapterDialog({ ...deleteChapterDialog, open })
        }
        onConfirm={handleDeleteChapterConfirm}
        title="Delete Chapter"
        description="Are you sure you want to delete this chapter? This action cannot be undone and will delete all notes within this chapter."
        itemName={deleteChapterDialog.chapterName}
        itemType="Chapter"
      />

      <RenameDialog
        open={renameNoteDialog.open}
        onOpenChange={(open) =>
          setRenameNoteDialog({ ...renameNoteDialog, open })
        }
        onSubmit={handleRenameNoteSubmit}
        title="Rename Note"
        currentName={renameNoteDialog.currentName}
        itemType="Note"
      />

      <DeleteConfirmDialog
        open={deleteNoteDialog.open}
        onOpenChange={(open) =>
          setDeleteNoteDialog({ ...deleteNoteDialog, open })
        }
        onConfirm={handleDeleteNoteConfirm}
        title="Delete Note"
        description="Are you sure you want to delete this note? This action cannot be undone."
        itemName={deleteNoteDialog.noteName}
        itemType="Note"
      />

      <CommandMenu notebooks={userNotebooks} />
    </>
  )
}

export default App
