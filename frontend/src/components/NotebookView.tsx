import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getUserNotebooks } from '@/utils/notebook'
import { Header } from '@/components/Header'
import { Button } from '@/components/ui/button'
import { PublishNotebookDialog } from '@/components/SidebarDialogs/PublishNotebookDialog'
import { BookOpen, Plus, Globe, Lock } from 'lucide-react'
import { useState } from 'react'
import { isNotebookPublished, getPublishedNoteCount } from '@/utils/publish'
import type { AuthenticatedUser } from '@/types/backend'

interface NotebookViewProps {
  user: AuthenticatedUser | null
  onCreateChapter?: (notebookId: string) => void
}

export function NotebookView({ user, onCreateChapter }: NotebookViewProps) {
  const { notebookId } = useParams<{ notebookId: string }>()
  const navigate = useNavigate()
  const [publishDialogOpen, setPublishDialogOpen] = useState(false)

  const { data: notebooks } = useQuery({
    queryKey: ['userNotebooks'],
    queryFn: getUserNotebooks,
    enabled: !!user,
  })

  const notebook = notebooks?.find((n) => n.id === notebookId)

  if (!notebook) {
    return (
      <div className="flex flex-col h-screen">
        <Header user={user} breadcrumbs={[{ label: 'Notebook Not Found' }]} />
        <main className="flex-1 overflow-auto flex items-center justify-center">
          <div className="text-center space-y-4">
            <h2 className="text-2xl font-bold text-foreground">Notebook Not Found</h2>
            <p className="text-muted-foreground">The notebook you're looking for doesn't exist.</p>
            <Button onClick={() => navigate('/')}>Go to Dashboard</Button>
          </div>
        </main>
      </div>
    )
  }

  const breadcrumbs = [
    { label: 'Dashboard', href: '/' },
    { label: notebook.name }
  ]

  const isPublished = isNotebookPublished(notebook)
  const publishedNoteCount = getPublishedNoteCount(notebook)

  const headerActions = (
    <Button
      variant={isPublished ? "default" : "outline"}
      onClick={() => setPublishDialogOpen(true)}
    >
      {isPublished ? (
        <>
          <Globe className="mr-2 h-4 w-4" />
          Published ({publishedNoteCount})
        </>
      ) : (
        <>
          <Lock className="mr-2 h-4 w-4" />
          Publish
        </>
      )}
    </Button>
  )

  return (
    <div className="flex flex-col h-screen">
      <Header user={user} breadcrumbs={breadcrumbs} actions={headerActions} />
      
      <main className="flex-1 overflow-auto">
        <div className="max-w-6xl mx-auto px-6 py-12 space-y-8">
          {/* Header Section */}
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <h2 className="text-3xl font-bold text-foreground">{notebook.name}</h2>
              <p className="text-muted-foreground">
                {notebook.chapters?.length || 0} chapter{notebook.chapters?.length !== 1 ? 's' : ''}
              </p>
            </div>
            <Button onClick={() => onCreateChapter?.(notebookId!)}>
              <Plus className="mr-2 h-4 w-4" />
              New Chapter
            </Button>
          </div>

          {/* Chapters Grid */}
          {notebook.chapters && notebook.chapters.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {notebook.chapters.map((chapter) => (
                <button
                  key={chapter.id}
                  onClick={() => navigate(`/${notebookId}/${chapter.id}`)}
                  className="bg-card border border-border rounded-lg p-6 space-y-3 hover:border-primary/50 transition-colors text-left"
                >
                  <div className="flex items-start justify-between">
                    <BookOpen className="h-8 w-8 text-primary" />
                    <span className="text-sm text-muted-foreground">
                      {chapter.notes?.length || 0} note{chapter.notes?.length !== 1 ? 's' : ''}
                    </span>
                  </div>
                  <div className="space-y-1">
                    <h3 className="text-lg font-semibold text-card-foreground">{chapter.name}</h3>
                    <p className="text-sm text-muted-foreground">
                      Click to view all notes in this chapter
                    </p>
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-16 space-y-4">
              <BookOpen className="h-16 w-16 text-muted-foreground" />
              <div className="text-center space-y-2">
                <h3 className="text-xl font-semibold text-foreground">No Chapters Yet</h3>
                <p className="text-muted-foreground">Create your first chapter to start organizing your notes.</p>
              </div>
              <Button onClick={() => onCreateChapter?.(notebookId!)}>
                <Plus className="mr-2 h-4 w-4" />
                Create Chapter
              </Button>
            </div>
          )}
        </div>
      </main>

      <PublishNotebookDialog
        open={publishDialogOpen}
        onOpenChange={setPublishDialogOpen}
        notebookId={notebookId ?? ""}
      />
    </div>
  )
}

