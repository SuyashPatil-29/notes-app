import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getUserNotebooks } from '@/utils/notebook'
import { Header } from '@/components/Header'
import { Button } from '@/components/ui/button'
import { FileText, Plus } from 'lucide-react'
import type { AuthenticatedUser } from '@/types/backend'

interface ChapterViewProps {
  user: AuthenticatedUser | null
  onCreateNote?: (chapterId: string) => void
}

export function ChapterView({ user, onCreateNote }: ChapterViewProps) {
  const { notebookId, chapterId } = useParams<{ notebookId: string; chapterId: string }>()
  const navigate = useNavigate()

  const { data: notebooks } = useQuery({
    queryKey: ['userNotebooks'],
    queryFn: getUserNotebooks,
    enabled: !!user,
  })

  const notebook = notebooks?.find((n) => n.id === notebookId)
  const chapter = notebook?.chapters?.find((c) => c.id === chapterId)

  if (!notebook || !chapter) {
    return (
      <div className="flex flex-col h-screen">
        <Header user={user} breadcrumbs={[{ label: 'Chapter Not Found' }]} />
        <main className="flex-1 overflow-auto flex items-center justify-center">
          <div className="text-center space-y-4">
            <h2 className="text-2xl font-bold text-foreground">Chapter Not Found</h2>
            <p className="text-muted-foreground">The chapter you're looking for doesn't exist.</p>
            <Button onClick={() => navigate('/')}>Go to Dashboard</Button>
          </div>
        </main>
      </div>
    )
  }

  const breadcrumbs = [
    { label: 'Dashboard', href: '/' },
    { label: notebook.name, href: `/${notebookId}` },
    { label: chapter.name }
  ]

  return (
    <div className="flex flex-col h-screen">
      <Header user={user} breadcrumbs={breadcrumbs} />
      
      <main className="flex-1 overflow-auto">
        <div className="max-w-6xl mx-auto px-6 py-12 space-y-8">
          {/* Header Section */}
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <h2 className="text-3xl font-bold text-foreground">{chapter.name}</h2>
              <p className="text-muted-foreground">
                {chapter.notes?.length || 0} note{chapter.notes?.length !== 1 ? 's' : ''}
              </p>
            </div>
            <Button onClick={() => onCreateNote?.(chapterId!)}>
              <Plus className="mr-2 h-4 w-4" />
              New Note
            </Button>
          </div>

          {/* Notes Grid */}
          {chapter.notes && chapter.notes.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {chapter.notes.map((note) => (
                <button
                  key={note.id}
                  onClick={() => navigate(`/${notebookId}/${chapterId}/${note.id}`)}
                  className="bg-card border border-border rounded-lg p-6 space-y-3 hover:border-primary/50 transition-colors text-left"
                >
                  <div className="flex items-start justify-between">
                    <FileText className="h-8 w-8 text-primary" />
                  </div>
                  <div className="space-y-1">
                    <h3 className="text-lg font-semibold text-card-foreground">{note.name}</h3>
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {note.content ? note.content.substring(0, 100) + (note.content.length > 100 ? '...' : '') : 'Empty note'}
                    </p>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Updated {new Date(note.updatedAt).toLocaleDateString()}
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-16 space-y-4">
              <FileText className="h-16 w-16 text-muted-foreground" />
              <div className="text-center space-y-2">
                <h3 className="text-xl font-semibold text-foreground">No Notes Yet</h3>
                <p className="text-muted-foreground">Create your first note to start writing.</p>
              </div>
              <Button onClick={() => onCreateNote?.(chapterId!)}>
                <Plus className="mr-2 h-4 w-4" />
                Create Note
              </Button>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}

