import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getNote } from '@/utils/notes'
import { getUserNotebooks } from '@/utils/notebook'
import { Header } from '@/components/Header'
import type { AuthenticatedUser } from '@/types/backend'
import { Loader2 } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeKatex from 'rehype-katex'
import 'katex/dist/katex.min.css'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { useTheme } from 'next-themes'

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
  const { theme } = useTheme()

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

  // Truncate note name if too long
  const truncateNoteName = (name: string, maxLength: number = 10) => {
    return name.length > maxLength ? name.substring(0, maxLength) + '...' : name
  }

  const breadcrumbs = [
    { label: 'Dashboard', href: '/' },
    ...(notebook ? [{ label: notebook.name, href: `/${notebookId}` }] : []),
    ...(chapter ? [{ label: chapter.name, href: `/${notebookId}/${chapterId}` }] : []),
    { label: truncateNoteName(note.name) },
  ]

  return (
    <div className="flex flex-col h-screen">
      <Header
        user={user}
        breadcrumbs={breadcrumbs}
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
                <ReactMarkdown
                  remarkPlugins={[remarkGfm, remarkMath]}
                  rehypePlugins={[rehypeKatex]}
                  components={{
                    code({ node, inline, className, children, ...props }: any) {
                      const match = /language-(\w+)/.exec(className || '')
                      const language = match ? match[1] : 'text'
                      
                      return !inline && match ? (
                        <SyntaxHighlighter
                          style={theme === 'dark' ? oneDark : oneLight}
                          language={language}
                          PreTag="div"
                          customStyle={{
                            margin: '1rem 0',
                            borderRadius: '0.5rem',
                          }}
                        >
                          {String(children).replace(/\n$/, '')}
                        </SyntaxHighlighter>
                      ) : (
                        <code className="px-1.5 py-0.5 bg-muted rounded text-sm font-mono" {...props}>
                          {children}
                        </code>
                      )
                    },
                    // Style links
                    a({ children, href, ...props }: any) {
                      return (
                        <a 
                          href={href} 
                          target="_blank" 
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                          {...props}
                        >
                          {children}
                        </a>
                      )
                    },
                    // Style tables
                    table({ children, ...props }: any) {
                      return (
                        <div className="overflow-x-auto my-4">
                          <table className="min-w-full divide-y divide-border" {...props}>
                            {children}
                          </table>
                        </div>
                      )
                    },
                    // Style blockquotes
                    blockquote({ children, ...props }: any) {
                      return (
                        <blockquote 
                          className="border-l-4 border-primary/50 pl-4 italic my-4 text-muted-foreground"
                          {...props}
                        >
                          {children}
                        </blockquote>
                      )
                    },
                  }}
                >
                  {note.content}
                </ReactMarkdown>
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

