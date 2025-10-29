import { useState, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Book, BookOpen, FileText } from 'lucide-react'
import { Header } from '@/components/Header'
import { handleGoogleLogin } from '@/utils/auth'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getUserNotebooks } from '@/utils/notebook'
import type { AuthenticatedUser } from '@/types/backend'

interface DashboardProps {
  user: AuthenticatedUser | null
}

export function Dashboard({ user }: DashboardProps) {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'notebooks' | 'chapters' | 'notes'>('notebooks')

  const { data: notebooks, isLoading } = useQuery({
    queryKey: ['userNotebooks'],
    queryFn: getUserNotebooks,
    enabled: !!user,
  })

  // Flatten all chapters and notes for the tabs
  const allChapters = useMemo(() => {
    return notebooks?.flatMap(notebook => 
      notebook.chapters?.map(chapter => ({
        ...chapter,
        notebookName: notebook.name,
        notebookId: notebook.id,
      })) || []
    ) || []
  }, [notebooks])

  const allNotes = useMemo(() => {
    return notebooks?.flatMap(notebook =>
      notebook.chapters?.flatMap(chapter =>
        chapter.notes?.map(note => ({
          ...note,
          chapterName: chapter.name,
          chapterId: chapter.id,
          notebookName: notebook.name,
          notebookId: notebook.id,
        })) || []
      ) || []
    ) || []
  }, [notebooks])

  return (
    <div className="flex flex-col h-screen">
      <Header user={user} breadcrumbs={[{ label: 'Dashboard' }]} />
      
      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        {user ? (
          <div className="max-w-6xl mx-auto px-6 py-12 space-y-8">
            {/* Welcome Section */}
            <div className="space-y-2">
              <h2 className="text-3xl font-bold text-foreground">
                Welcome back, {user.name}!
              </h2>
              <p className="text-muted-foreground">
                {notebooks?.length || 0} notebook{notebooks?.length !== 1 ? 's' : ''} • {allChapters.length} chapter{allChapters.length !== 1 ? 's' : ''} • {allNotes.length} note{allNotes.length !== 1 ? 's' : ''}
              </p>
            </div>

            {/* Tabs */}
            <div className="flex items-center gap-1 border-b border-border">
              <button
                onClick={() => setActiveTab('notebooks')}
                className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                  activeTab === 'notebooks'
                    ? 'text-primary'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <Book className="h-4 w-4" />
                Notebooks
                {activeTab === 'notebooks' && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
                )}
              </button>
              <button
                onClick={() => setActiveTab('chapters')}
                className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                  activeTab === 'chapters'
                    ? 'text-primary'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <BookOpen className="h-4 w-4" />
                Chapters
                {activeTab === 'chapters' && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
                )}
              </button>
              <button
                onClick={() => setActiveTab('notes')}
                className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                  activeTab === 'notes'
                    ? 'text-primary'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <FileText className="h-4 w-4" />
                Notes
                {activeTab === 'notes' && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
                )}
              </button>
            </div>

            {/* Content Area */}
            {isLoading ? (
              <div className="flex items-center justify-center py-16">
                <div className="text-muted-foreground">Loading...</div>
              </div>
            ) : (
              <>
                {/* Notebooks Tab */}
                {activeTab === 'notebooks' && (
                  <>
                    {notebooks && notebooks.length > 0 ? (
                      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {notebooks.map((notebook) => (
                          <button
                            key={notebook.id}
                            onClick={() => navigate(`/${notebook.id}`)}
                            className="bg-card border border-border rounded-lg p-6 space-y-3 hover:border-primary/50 transition-colors text-left"
                          >
                            <div className="flex items-start justify-between">
                              <Book className="h-8 w-8 text-primary" />
                              <span className="text-sm text-muted-foreground">
                                {notebook.chapters?.length || 0} chapter{notebook.chapters?.length !== 1 ? 's' : ''}
                              </span>
                            </div>
                            <div className="space-y-1">
                              <h3 className="text-lg font-semibold text-card-foreground">{notebook.name}</h3>
                              <p className="text-sm text-muted-foreground">
                                Click to view all chapters
                              </p>
                            </div>
                          </button>
                        ))}
                      </div>
                    ) : (
                      <div className="flex flex-col items-center justify-center py-16 space-y-4">
                        <Book className="h-16 w-16 text-muted-foreground" />
                        <div className="text-center space-y-2">
                          <h3 className="text-xl font-semibold text-foreground">No Notebooks Yet</h3>
                          <p className="text-muted-foreground">Create your first notebook to start organizing your notes.</p>
                        </div>
                      </div>
                    )}
                  </>
                )}

                {/* Chapters Tab */}
                {activeTab === 'chapters' && (
                  <>
                    {allChapters.length > 0 ? (
                      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {allChapters.map((chapter) => (
                          <button
                            key={chapter.id}
                            onClick={() => navigate(`/${chapter.notebookId}/${chapter.id}`)}
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
                                in {chapter.notebookName}
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
                          <p className="text-muted-foreground">Create a chapter in a notebook to get started.</p>
                        </div>
                      </div>
                    )}
                  </>
                )}

                {/* Notes Tab */}
                {activeTab === 'notes' && (
                  <>
                    {allNotes.length > 0 ? (
                      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {allNotes.map((note) => (
                          <button
                            key={note.id}
                            onClick={() => navigate(`/${note.notebookId}/${note.chapterId}/${note.id}`)}
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
                              <p className="text-xs text-muted-foreground pt-1">
                                {note.chapterName} • {note.notebookName}
                              </p>
                            </div>
                          </button>
                        ))}
                      </div>
                    ) : (
                      <div className="flex flex-col items-center justify-center py-16 space-y-4">
                        <FileText className="h-16 w-16 text-muted-foreground" />
                        <div className="text-center space-y-2">
                          <h3 className="text-xl font-semibold text-foreground">No Notes Yet</h3>
                          <p className="text-muted-foreground">Create a note in a chapter to start writing.</p>
                        </div>
                      </div>
                    )}
                  </>
                )}
              </>
            )}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center min-h-[calc(100vh-80px)] space-y-6">
            <div className="text-center space-y-3">
              <h2 className="text-4xl font-bold text-foreground">Welcome to Notes App</h2>
              <p className="text-lg text-muted-foreground max-w-md">
                Sign in with your Google account to start organizing your thoughts and ideas
              </p>
            </div>
            <Button size="lg" onClick={handleGoogleLogin}>
              Login with Google
            </Button>
          </div>
        )}
      </main>
    </div>
  )
}

