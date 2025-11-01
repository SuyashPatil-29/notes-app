import { useState, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Book, BookOpen, FileText, Video } from 'lucide-react'
import { Header } from '@/components/Header'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getUserNotebooks, createNotebook } from '@/utils/notebook'
import type { AuthenticatedUser } from '@/types/backend'
import { createId } from '@paralleldrive/cuid2'
import { useQueryClient } from '@tanstack/react-query'
import { createNote } from '@/utils/notes'
import { createChapter as createChapterApi } from '@/utils/chapter'
import { toast } from 'sonner'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '@/components/ui/select'
import { getPreviewText } from '@/utils/markdown'
import { MeetingsList } from '@/components/MeetingsList'

interface DashboardProps {
  user: AuthenticatedUser | null
  userLoading?: boolean
}

export function Dashboard({ user, userLoading = false }: DashboardProps) {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'notebooks' | 'chapters' | 'notes' | 'meetings'>('notebooks')
  const queryClient = useQueryClient()

  // Create Note modal state
  const [createNoteOpen, setCreateNoteOpen] = useState(false)
  const [newNoteName, setNewNoteName] = useState('')
  const [selectedNotebookId, setSelectedNotebookId] = useState<string>('')
  const [selectedChapterId, setSelectedChapterId] = useState<string>('')

  // Create Notebook modal state
  const [createNotebookOpen, setCreateNotebookOpen] = useState(false)
  const [newNotebookName, setNewNotebookName] = useState('')

  // Create Chapter modal state
  const [createChapterOpen, setCreateChapterOpen] = useState(false)
  const [newChapterName, setNewChapterName] = useState('')
  const [selectedNotebookIdForChapter, setSelectedNotebookIdForChapter] = useState<string>('')

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

  const openCreateNote = () => {
    // Default notebook/chapter selections
    const firstNotebook = notebooks && notebooks[0]
    const firstChapter = firstNotebook?.chapters?.[0]
    setSelectedNotebookId(firstNotebook?.id || '')
    setSelectedChapterId(firstChapter?.id || '')
    setNewNoteName('')
    setCreateNoteOpen(true)
  }

  const openCreateNotebook = () => {
    setNewNotebookName('')
    setCreateNotebookOpen(true)
  }

  const openCreateChapter = () => {
    const firstNotebook = notebooks && notebooks[0]
    setSelectedNotebookIdForChapter(firstNotebook?.id || '')
    setNewChapterName('')
    setCreateChapterOpen(true)
  }

  const handleNotebookChange = (nbId: string) => {
    setSelectedNotebookId(nbId)
    const nb = notebooks?.find(n => n.id === nbId)
    const firstChapter = nb?.chapters?.[0]
    setSelectedChapterId(firstChapter?.id || '')
  }

  const handleCreateNoteSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedChapterId || !newNoteName.trim()) return
    try {
      const id = createId()
      await createNote({
        id,
        name: newNoteName.trim(),
        content: '',
        isPublic: false,
        chapterId: selectedChapterId,
        chapter: {} as any,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
      toast.success('Note created successfully!')
      setCreateNoteOpen(false)
      await queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
      navigate(`/${selectedNotebookId}/${selectedChapterId}/${id}`)
    } catch (err: any) {
      toast.error('Failed to create note')
    }
  }

  const handleCreateNotebookSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!user || !newNotebookName.trim()) return
    try {
      const id = createId()
      await createNotebook({
        id,
        name: newNotebookName.trim(),
        userId: user.id,
        chapters: [],
        isPublic: false,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
      toast.success('Notebook created successfully!')
      setCreateNotebookOpen(false)
      await queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
      navigate(`/${id}`)
    } catch (err: any) {
      toast.error('Failed to create notebook')
    }
  }

  const handleCreateChapterSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedNotebookIdForChapter || !newChapterName.trim()) return
    try {
      const id = createId()
      await createChapterApi({
        id,
        name: newChapterName.trim(),
        isPublic: false,
        notebookId: selectedNotebookIdForChapter,
        notebook: {} as any,
        notes: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
      toast.success('Chapter created successfully!')
      setCreateChapterOpen(false)
      await queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
      navigate(`/${selectedNotebookIdForChapter}/${id}`)
    } catch (err: any) {
      toast.error('Failed to create chapter')
    }
  }

  console.log('userLoading:', userLoading)
  console.log('user:', user)

  return (
    <div className="flex flex-col h-screen">
      <Header user={user} breadcrumbs={[{ label: 'Dashboard' }]} />
      
      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        {userLoading ? (
          <div className="max-w-6xl mx-auto px-6 py-12 space-y-8">
            {/* Loading Skeleton */}
            <div className="space-y-2">
              <Skeleton className="h-9 w-64" />
              <Skeleton className="h-5 w-96" />
            </div>
            <div className="h-12" />
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-40 rounded-lg" />
              ))}
            </div>
          </div>
        ) : user ? (
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
              <button
                onClick={() => setActiveTab('meetings')}
                className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                  activeTab === 'meetings'
                    ? 'text-primary'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <Video className="h-4 w-4" />
                Meetings
                {activeTab === 'meetings' && (
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
                      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                      {/* Create Note tile */}
                      <button
                        onClick={openCreateNotebook}
                        className="border-2 border-dashed border-border/80 rounded-2xl p-6 h-40 flex items-center justify-center text-muted-foreground hover:text-foreground hover:border-foreground/30 transition-colors"
                        title="Create new note"
                      >
                        <span className="text-4xl leading-none">+</span>
                      </button>
                      {(notebooks ?? []).map((notebook) => (
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
                    {!notebooks?.length && (
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
                      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                      {/* Create Note tile */}
                      <button
                        onClick={openCreateChapter}
                        className="border-2 border-dashed border-border/80 rounded-2xl p-6 h-40 flex items-center justify-center text-muted-foreground hover:text-foreground hover:border-foreground/30 transition-colors"
                        title="Create new note"
                      >
                        <span className="text-4xl leading-none">+</span>
                      </button>
                      {allChapters.length > 0 && allChapters.map((chapter) => (
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
                    {allChapters.length === 0 && (
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
                      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                      {/* Create Note tile */}
                      <button
                        onClick={openCreateNote}
                        className="border-2 border-dashed border-border/80 rounded-2xl p-6 h-40 flex items-center justify-center text-muted-foreground hover:text-foreground hover:border-foreground/30 transition-colors"
                        title="Create new note"
                      >
                        <span className="text-4xl leading-none">+</span>
                      </button>
                      {allNotes.length > 0 && allNotes.map((note) => (
                          <button
                            key={note.id}
                            onClick={() => navigate(`/${note.notebookId}/${note.chapterId}/${note.id}`)}
                            className="bg-card border border-border rounded-lg p-6 hover:border-primary/50 transition-colors text-left h-40 flex flex-col justify-between"
                          >
                            <div className="space-y-2">
                              <div className="flex items-start justify-between">
                                <FileText className="h-8 w-8 text-primary" />
                              </div>
                              <div>
                                <h3 className="text-lg font-semibold text-card-foreground line-clamp-1">{note.name}</h3>
                                <p className="text-sm text-muted-foreground line-clamp-2 mt-1">
                                  {getPreviewText(note.content, 100)}
                                </p>
                              </div>
                            </div>
                            <p className="text-xs text-muted-foreground">
                              {note.chapterName} • {note.notebookName}
                            </p>
                          </button>
                        ))}
                      </div>
                    {allNotes.length === 0 && (
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

                {/* Meetings Tab */}
                {activeTab === 'meetings' && (
                  <div>
                    <MeetingsList />
                  </div>
                )}
              </>
            )}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center min-h-[calc(100vh-80px)] space-y-6">
            <div className="text-center space-y-3">
              <h2 className="text-4xl font-bold text-foreground">Welcome to Notes App</h2>
              <p className="text-lg text-muted-foreground max-w-md">
                Get started by creating your first notebook
              </p>
            </div>
          </div>
        )}
      </main>

      {/* Create Note Modal */}
      <Dialog open={createNoteOpen} onOpenChange={setCreateNoteOpen}>
        <DialogContent className="sm:max-w-[480px]">
          <form onSubmit={handleCreateNoteSubmit}>
            <DialogHeader>
              <DialogTitle>Create Note</DialogTitle>
              <DialogDescription>Select a destination and name your note.</DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label>Notebook</Label>
                <Select value={selectedNotebookId} onValueChange={handleNotebookChange}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a notebook" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectLabel>Notebooks</SelectLabel>
                      {notebooks?.map(nb => (
                        <SelectItem key={nb.id} value={nb.id}>{nb.name}</SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label>Chapter</Label>
                <Select value={selectedChapterId} onValueChange={setSelectedChapterId}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a chapter" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectLabel>Chapters</SelectLabel>
                      {notebooks?.find(n => n.id === selectedNotebookId)?.chapters?.map(ch => (
                        <SelectItem key={ch.id} value={ch.id}>{ch.name}</SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label>Note name</Label>
                <Input value={newNoteName} onChange={e => setNewNoteName(e.target.value)} placeholder="e.g. Daily notes" />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setCreateNoteOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={!selectedChapterId || !newNoteName.trim()}>Create</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Create Notebook Modal */}
      <Dialog open={createNotebookOpen} onOpenChange={setCreateNotebookOpen}>
        <DialogContent className="sm:max-w-[480px]">
          <form onSubmit={handleCreateNotebookSubmit}>
            <DialogHeader>
              <DialogTitle>Create Notebook</DialogTitle>
              <DialogDescription>Name your new notebook.</DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label>Notebook name</Label>
                <Input value={newNotebookName} onChange={e => setNewNotebookName(e.target.value)} placeholder="e.g. My Ideas" />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setCreateNotebookOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={!newNotebookName.trim()}>Create</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Create Chapter Modal */}
      <Dialog open={createChapterOpen} onOpenChange={setCreateChapterOpen}>
        <DialogContent className="sm:max-w-[480px]">
          <form onSubmit={handleCreateChapterSubmit}>
            <DialogHeader>
              <DialogTitle>Create Chapter</DialogTitle>
              <DialogDescription>Select a notebook and name your chapter.</DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label>Notebook</Label>
                <Select value={selectedNotebookIdForChapter} onValueChange={setSelectedNotebookIdForChapter}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a notebook" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectLabel>Notebooks</SelectLabel>
                      {notebooks?.map(nb => (
                        <SelectItem key={nb.id} value={nb.id}>{nb.name}</SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label>Chapter name</Label>
                <Input value={newChapterName} onChange={e => setNewChapterName(e.target.value)} placeholder="e.g. Basics" />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setCreateChapterOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={!selectedNotebookIdForChapter || !newChapterName.trim()}>Create</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}

