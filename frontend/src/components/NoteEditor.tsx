import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getNote } from '@/utils/notes'
import { getUserNotebooks } from '@/utils/notebook'
import { Header } from '@/components/Header'
import type { AuthenticatedUser } from '@/types/backend'
import { Loader2 } from 'lucide-react'

interface NoteEditorProps {
  user: AuthenticatedUser | null
}

export function NoteEditor({ user }: NoteEditorProps) {
  const { notebookId, chapterId, noteId } = useParams<{
    notebookId: string
    chapterId: string
    noteId: string
  }>()
  const navigate = useNavigate()

  const { data: noteResponse, isLoading, error } = useQuery({
    queryKey: ['note', noteId],
    queryFn: () => getNote(noteId!),
    enabled: !!noteId,
  })

  const { data: notebooks } = useQuery({
    queryKey: ['userNotebooks'],
    queryFn: getUserNotebooks,
    enabled: !!user,
  })

  const note = noteResponse?.data
  const notebook = notebooks?.find((n) => n.id === notebookId)
  const chapter = notebook?.chapters?.find((c) => c.id === chapterId)

  if (isLoading) {
    return (
      <div className="flex flex-col h-screen">
        <Header
          user={user}
          breadcrumbs={[
            { label: 'Dashboard', href: '/' },
            { label: 'Loading...' },
          ]}
        />
        <div className="flex-1 flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </div>
    )
  }

  if (error || !note) {
    return (
      <div className="flex flex-col h-screen">
        <Header
          user={user}
          breadcrumbs={[
            { label: 'Dashboard', href: '/' },
            { label: 'Error' },
          ]}
        />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center space-y-4">
            <p className="text-lg text-destructive">Failed to load note</p>
            <button
              onClick={() => navigate('/')}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              ← Back to Dashboard
            </button>
          </div>
        </div>
      </div>
    )
  }

  const breadcrumbs = [
    { label: 'Dashboard', href: '/' },
    ...(notebook ? [{ label: notebook.name, href: `/${notebookId}` }] : []),
    ...(chapter ? [{ label: chapter.name, href: `/${notebookId}/${chapterId}` }] : []),
    { label: note.name },
  ]

  return (
    <div className="flex flex-col h-screen">
      <Header
        user={user}
        breadcrumbs={breadcrumbs}
        showCloseButton={true}
        onClose={() => navigate('/')}
      />
      <div className="flex-1 overflow-auto">
        <div className="max-w-4xl mx-auto px-6 py-8 space-y-6">
          <div className="space-y-2">
            <h1 className="text-4xl font-bold text-foreground">{note.name}</h1>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span>Notebook ID: {notebookId}</span>
              <span>•</span>
              <span>Chapter ID: {chapterId}</span>
              <span>•</span>
              <span>Note ID: {noteId}</span>
            </div>
          </div>

          <div className="bg-card border border-border rounded-lg p-6">
            <div className="prose prose-neutral dark:prose-invert max-w-none">
              {note.content ? (
                <pre className="whitespace-pre-wrap font-sans">
                  {note.content}
                </pre>
              ) : (
                <p className="text-muted-foreground italic">
                  This note is empty. Start writing to add content.
                </p>
              )}
            </div>
          </div>

          <div className="text-xs text-muted-foreground space-y-1">
            <p>Created: {new Date(note.createdAt).toLocaleString()}</p>
            <p>Updated: {new Date(note.updatedAt).toLocaleString()}</p>
          </div>
        </div>
      </div>
    </div>
  )
}

