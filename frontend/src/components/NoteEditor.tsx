import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getNote, generateNoteVideo, deleteNoteVideo } from '@/utils/notes'
import { getUserNotebooks } from '@/utils/notebook'
import { Header } from '@/components/Header'
import { Skeleton } from '@/components/ui/skeleton'
import { NoteVideoPlayer } from '@/components/NoteVideoPlayer'
import type { AuthenticatedUser } from '@/types/backend'
import { Loader2, Calendar, Clock, Video, VideoOff } from 'lucide-react'
import { toast } from 'sonner'
import { useOrganizationContext } from '@/contexts/OrganizationContext'
import 'katex/dist/katex.min.css'
import '@/prosemirror.css'

import {
  EditorCommand,
  EditorCommandEmpty,
  EditorCommandItem,
  EditorCommandList,
  EditorContent,
  type EditorInstance,
  EditorRoot,
} from "novel";
import {
  handleCommandNavigation,
  ImageResizer
} from "novel/extensions";
import { useState, useEffect, useRef } from "react";
import { useDebouncedCallback } from "use-debounce";
import { defaultExtensions } from "@/lib/extensions";
import { ColorSelector } from "./selectors/color-selector";
import { LinkSelector } from "./selectors/link-selector";
import { MathSelector } from "./selectors/math-selector";
import { NodeSelector } from "./selectors/node-selector";
import { Separator } from "./ui/separator";
import type { Notes } from "@/types/backend";
import { RealtimeAvatarStack } from '@/components/realtime-avatar-stack'
import { RealtimeCursors } from '@/components/realtime-cursors'
import { isRealtimeConfigured } from '@/utils/check-realtime'
import { useRealtimePresenceRoom } from '@/hooks/use-realtime-presence-room'
import { Eye } from 'lucide-react'

import GenerativeMenuSwitch from "./generative/generative-menu-switch";

// Yjs collaboration imports
import * as Y from 'yjs'
import { Awareness } from 'y-protocols/awareness'
import Collaboration from '@tiptap/extension-collaboration'
import CollaborationCaret from '@tiptap/extension-collaboration-caret'
import { SupabaseYjsProvider } from '@/lib/supabase-yjs-provider'
import { useUser, useSession } from '@clerk/clerk-react'
// import { uploadFn } from "./image-upload";
import { TextButtons } from "./selectors/text-buttons";
import { slashCommand, suggestionItems } from "./slash-command";

import hljs from "highlight.js";
import { Button } from './ui/button'

// Extensions will be configured dynamically in the component to include collaboration
// const extensions = [...defaultExtensions, slashCommand];

interface NoteEditorProps {
  user: AuthenticatedUser | null
  userLoading?: boolean
}

export function NoteEditor({ user, userLoading = false }: NoteEditorProps) {
  const { notebookId, chapterId, noteId } = useParams<{
    notebookId: string
    chapterId: string
    noteId: string
  }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { activeOrg } = useOrganizationContext()
  const { user: clerkUser } = useUser()
  const { session } = useSession()
  const [charsCount, setCharsCount] = useState();
  const [isSaving, setIsSaving] = useState(false);
  const [isAutoSaving, setIsAutoSaving] = useState(false);
  const [isGeneratingVideo, setIsGeneratingVideo] = useState(false);
  const [isDeletingVideo, setIsDeletingVideo] = useState(false);
  const autoSaveTimerRef = useRef<number | null>(null);
  const hasUnsavedChanges = useRef(false);

  const [openAI, setOpenAI] = useState(false);
  const [openNode, setOpenNode] = useState(false);
  const [openColor, setOpenColor] = useState(false);
  const [openLink, setOpenLink] = useState(false);

  // Yjs collaboration state
  const ydoc = useRef<Y.Doc | null>(null)
  const provider = useRef<SupabaseYjsProvider | null>(null)
  const [collaborationSynced, setCollaborationSynced] = useState(false)
  const [collaborationStatus, setCollaborationStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('connecting')
  const [providerReady, setProviderReady] = useState(false)
  const [awareness, setAwareness] = useState<Awareness | null>(null)

  // Realtime presence for viewer count (keeps existing presence tracking)
  const { users: realtimeUsers } = useRealtimePresenceRoom(noteId ? `note-${noteId}` : '');
  const viewerCount = Object.keys(realtimeUsers).length;

  const { data: noteResponse, isLoading, error } = useQuery({
    queryKey: ['note', noteId],
    queryFn: async () => {
      return await getNote(noteId!);
    },
    enabled: !!noteId && !!user, // Also require user to be loaded
    refetchInterval: 5000, // Refetch every 5 seconds to catch AI-generated videos and sync updates
    refetchOnWindowFocus: false,
    retry: 1, // Retry once on failure
  })

  const { data: notebooks } = useQuery({
    queryKey: ['userNotebooks', activeOrg?.id],
    queryFn: () => getUserNotebooks(activeOrg?.id),
    enabled: !!user,
  })

  const note: Notes = noteResponse?.data
  const notebook = notebooks?.find((n) => n.id === notebookId)
  const chapter = notebook?.chapters?.find((c) => c.id === chapterId)

  // With Yjs collaboration, we don't use initialContent from React
  // Content is managed by the Yjs document, which is synced via the provider
  // Initial content is set when the document is first created via onCreate callback

  //Apply Codeblock Highlighting on the HTML from editor.getHTML()
  const highlightCodeblocks = (content: string) => {
    const doc = new DOMParser().parseFromString(content, "text/html");
    doc.querySelectorAll("pre code").forEach((el) => {
      // @ts-ignore
      // https://highlightjs.readthedocs.io/en/latest/api.html?highlight=highlightElement#highlightelement
      hljs.highlightElement(el);
    });
    return new XMLSerializer().serializeToString(doc);
  };

  const debouncedUpdates = useDebouncedCallback(async (editor: EditorInstance) => {
    const json = editor.getJSON();
    setCharsCount(editor.storage.characterCount.words());
    window.localStorage.setItem("html-content", highlightCodeblocks(editor.getHTML()));
    window.localStorage.setItem("novel-content", JSON.stringify(json));
    window.localStorage.setItem("markdown", editor.storage.markdown.getMarkdown());
  }, 500);

  // Cleanup auto-save timer on unmount - MUST be before any conditional returns
  useEffect(() => {
    return () => {
      if (autoSaveTimerRef.current) {
        window.clearTimeout(autoSaveTimerRef.current)
      }
    }
  }, [])

  // Initialize Yjs collaboration
  useEffect(() => {
    if (!noteId || !clerkUser || !session) return

    // Reset provider ready state
    setProviderReady(false)

    // Create Yjs document and provider
    ydoc.current = new Y.Doc()

    // Function to get auth token
    const getAuthToken = async () => {
      try {
        const token = await session.getToken()
        return token
      } catch (error) {
        console.error('[NoteEditor] Failed to get auth token:', error)
        return null
      }
    }

    provider.current = new SupabaseYjsProvider(noteId, ydoc.current, clerkUser, getAuthToken)

    // Store awareness in state
    setAwareness(provider.current.awareness)

    // Setup event handlers
    provider.current.on('synced', () => {
      console.log('[NoteEditor] Yjs document synced')
      setCollaborationSynced(true)
      setCollaborationStatus('connected')

      // Ensure awareness is fully initialized before marking as ready
      setTimeout(() => {
        if (provider.current?.awareness) {
          console.log('[NoteEditor] Provider ready with awareness')
          setProviderReady(true)
        }
      }, 100) // Small delay to ensure everything is initialized
    })

    provider.current.on('status', ({ status }) => {
      console.log('[NoteEditor] Collaboration status:', status)
      setCollaborationStatus(status as any)
    })

    provider.current.on('error', (error) => {
      console.error('[NoteEditor] Collaboration error:', error)
      setCollaborationStatus('error')
      toast.error('Collaboration error: ' + error.message)
    })

    // Cleanup on unmount
    return () => {
      provider.current?.destroy()
      ydoc.current?.destroy()
      provider.current = null
      ydoc.current = null
      setAwareness(null)
      setProviderReady(false)
    }
  }, [noteId, clerkUser, session])

  console.log({ providerReady }, ":", { ydoc }, ":", { awareness })

  if (userLoading || isLoading) {
    return (
      <div className="flex flex-col h-screen">
        <Header
          user={null}
          breadcrumbs={[
            { label: 'Dashboard', href: '/' },
            { label: 'Loading...' },
          ]}
        />
        <main className="flex-1 overflow-auto p-6">
          <div className="max-w-4xl mx-auto space-y-4">
            <Skeleton className="h-10 w-3/4" />
            <div className="space-y-2">
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-2/3" />
            </div>
            <Skeleton className="h-64 w-full" />
          </div>
        </main>
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

  // Removed old edit lock handlers - collaboration is always active with Yjs

  // Save function is now handled automatically by Yjs provider
  // Manual save will trigger a sync to backend
  const handleSave = async (isAutoSave = false) => {
    if (!provider.current || !noteId) {
      if (!isAutoSave) {
        toast.error("Collaboration not ready")
      }
      return
    }

    if (isAutoSave) {
      setIsAutoSaving(true)
    } else {
      setIsSaving(true)
    }

    try {
      // Trigger immediate sync to backend
      await provider.current.sync()

      if (!isAutoSave) {
        toast.success("Note synced successfully!")
      }
    } catch (error) {
      console.error("Failed to sync note:", error)
      if (!isAutoSave) {
        toast.error("Failed to sync note")
      }
    } finally {
      if (isAutoSave) {
        setIsAutoSaving(false)
      } else {
        setIsSaving(false)
      }
    }
  }

  const handleGenerateVideo = async () => {
    if (!noteId) {
      toast.error("Note not available")
      return
    }

    setIsGeneratingVideo(true)
    try {
      await generateNoteVideo(noteId)
      // Invalidate queries to refetch the updated note with video data
      queryClient.invalidateQueries({ queryKey: ['note', noteId] })
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
      toast.success("Video generated successfully!")
    } catch (error) {
      console.error("Failed to generate video:", error)
      toast.error("Failed to generate video")
    } finally {
      setIsGeneratingVideo(false)
    }
  }

  const handleDeleteVideo = async () => {
    if (!noteId) {
      toast.error("Note not available")
      return
    }

    setIsDeletingVideo(true)
    try {
      await deleteNoteVideo(noteId)
      // Invalidate queries to refetch the updated note without video data
      queryClient.invalidateQueries({ queryKey: ['note', noteId] })
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })
      toast.success("Video removed successfully!")
    } catch (error) {
      console.error("Failed to delete video:", error)
      toast.error("Failed to remove video")
    } finally {
      setIsDeletingVideo(false)
    }
  }

  // Format date for display
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  // Generate consistent color for user based on ID
  const generateUserColor = (userId: string): string => {
    const colors = [
      '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A',
      '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E2',
      '#F8B739', '#52B788', '#E07A5F', '#81B29A',
    ]

    let hash = 0
    for (let i = 0; i < userId.length; i++) {
      hash = userId.charCodeAt(i) + ((hash << 5) - hash)
    }

    return colors[Math.abs(hash) % colors.length]
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
      {/* Real-time cursors overlay - always visible when configured */}
      {isRealtimeConfigured() && <RealtimeCursors roomName={`note-${noteId}`} />}
      <div className="flex-1 overflow-auto">
        <div className="max-w-4xl mx-auto px-6 py-8">
          {/* Note Metadata Section */}
          <div className="mb-6 space-y-4">
            <div className='flex items-center justify-between gap-2'>
              <h1 className="text-4xl font-bold tracking-tight">{note.name}</h1>
              {isRealtimeConfigured() && (
                <div className="flex flex-col items-end gap-2">
                  {/* Collaboration status */}
                  {collaborationStatus === 'connected' && (
                    <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-primary/10 text-primary border border-primary/20">
                      <div className="h-2 w-2 rounded-full bg-primary animate-pulse" />
                      <span>Live collaboration active</span>
                    </div>
                  )}
                  {collaborationStatus === 'connecting' && (
                    <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-muted text-muted-foreground border border-border">
                      <div className="h-2 w-2 rounded-full bg-muted-foreground animate-pulse" />
                      <span>Connecting...</span>
                    </div>
                  )}

                  {/* Viewers section */}
                  <div className="flex flex-col items-center gap-3">
                    <RealtimeAvatarStack roomName={`note-${noteId}`} />
                    <div className="flex flex-col items-end gap-1">
                      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                        <Eye className="h-3.5 w-3.5" />
                        <span>
                          {viewerCount === 0
                            ? 'No collaborators'
                            : viewerCount === 1
                              ? '1 collaborator'
                              : `${viewerCount} collaborators`}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>

            <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                <span>Created: {formatDate(note.createdAt)}</span>
              </div>
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4" />
                <span>Updated: {formatDate(note.updatedAt)}</span>
              </div>
              {note.meetingRecordingId && (
                <div className="flex items-center gap-2 px-2 py-1 bg-blue-100 dark:bg-blue-900/30 rounded-md">
                  <Video className="h-4 w-4 text-secondary dark:text-secondary" />
                  <span className="text-blue-800 dark:text-blue-200 font-medium">Generated from Meeting</span>
                </div>
              )}
            </div>

            <Separator />

            {/* AI Summary for Meeting Notes */}
            {note.aiSummary && (
              <div className="bg-linear-to-r from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
                <h3 className="text-sm font-semibold text-blue-900 dark:text-blue-100 mb-2 flex items-center gap-2">
                  <Video className="h-4 w-4" />
                  Meeting Summary
                </h3>
                <p className="text-sm text-blue-800 dark:text-blue-200 leading-relaxed">
                  {note.aiSummary}
                </p>
              </div>
            )}
          </div>

          {/* Path and Save - Sticky */}
          <div className="sticky top-0 z-20 bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/80 border-b -mx-6 px-6 py-3 mb-6">
            <div className="flex items-center w-full justify-end">
              <div className='flex items-center gap-2'>
                {!note.hasVideo && (
                  <Button
                    variant="outline"
                    size="sm"
                    className="ml-2"
                    onClick={handleGenerateVideo}
                    disabled={isGeneratingVideo}
                  >
                    {isGeneratingVideo ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Generating...
                      </>
                    ) : (
                      <>
                        <Video className="mr-2 h-4 w-4" />
                        Generate Video
                      </>
                    )}
                  </Button>
                )}
                <Button
                  variant="ghost"
                  className="ml-2"
                  onClick={() => handleSave(false)}
                  disabled={isSaving}
                >
                  {isSaving ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    "Save"
                  )}
                </Button>
              </div>
            </div>
          </div>

          {/* Video Player - conditionally rendered */}
          {note.hasVideo && note.videoData && (
            <div className="mb-6">
              <div className="border rounded-lg p-4 bg-muted/30">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold flex items-center gap-2">
                    <Video className="h-5 w-5" />
                    Note Video
                  </h2>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleDeleteVideo}
                    disabled={isDeletingVideo}
                    className="text-destructive hover:text-destructive"
                  >
                    {isDeletingVideo ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Removing...
                      </>
                    ) : (
                      <>
                        <VideoOff className="mr-2 h-4 w-4" />
                        Remove Video
                      </>
                    )}
                  </Button>
                </div>
                <NoteVideoPlayer videoData={note.videoData} />
              </div>
            </div>
          )}

          <div className="relative w-full max-w-5xl">
            <div className="flex absolute right-5 top-5 z-10 mb-5 gap-2">
              {/* Collaboration status indicator */}
              {collaborationSynced && collaborationStatus === 'connected' && (
                <div className="rounded-lg bg-primary/10 border border-primary/20 px-2 py-1 text-xs text-primary flex items-center gap-1">
                  <div className="h-2 w-2 rounded-full bg-primary" />
                  Synced
                </div>
              )}
              {!providerReady && (
                <div className="rounded-lg bg-muted border border-border px-2 py-1 text-xs text-muted-foreground flex items-center gap-1">
                  <Loader2 className="h-3 w-3 animate-spin" />
                  Loading collaboration...
                </div>
              )}
              {isAutoSaving && (
                <div className="rounded-lg bg-accent/50 border border-accent px-2 py-1 text-xs text-accent-foreground flex items-center gap-1">
                  <Loader2 className="h-3 w-3 animate-spin" />
                  Syncing...
                </div>
              )}
              <div className={charsCount ? "rounded-lg bg-accent px-2 py-1 text-sm text-muted-foreground" : "hidden"}>
                {charsCount} Words
              </div>
            </div>
            {providerReady && ydoc.current && awareness ? (
              <EditorRoot key={`${noteId}-${note.updatedAt}-collab`}>
                <EditorContent
                  extensions={[
                    ...defaultExtensions,
                    slashCommand,
                    Collaboration.configure({
                      document: ydoc.current,
                    }),
                    CollaborationCaret.configure({
                      provider: provider.current,
                      user: {
                        name: clerkUser?.fullName || clerkUser?.username || 'Anonymous',
                        color: generateUserColor(clerkUser?.id || 'default'),
                      },
                    }),
                  ]}
                  className="relative min-h-[500px] w-full max-w-5xl border rounded-lg transition-colors border-border/40 hover:border-border/60 focus-within:border-border"
                  editorProps={{
                    handleDOMEvents: {
                      keydown: (_view, event) => handleCommandNavigation(event),
                    },
                    // handlePaste: (view, event) => handleImagePaste(view, event, uploadFn),
                    // handleDrop: (view, event, _slice, moved) => handleImageDrop(view, event, moved, uploadFn),
                    attributes: {
                      class:
                        "prose prose-lg dark:prose-invert prose-headings:font-title font-default focus:outline-none max-w-full",
                    },
                  }}
                  onCreate={async ({ editor }) => {
                    // Editor created callback
                    debouncedUpdates(editor);

                    // If this is a new document (from JSON conversion), set initial content
                    if (provider.current?.initialContent) {
                      console.log('[NoteEditor] Setting initial content from JSON');
                      try {
                        const parsed = JSON.parse(provider.current.initialContent);
                        if (parsed && parsed.type === 'doc') {
                          // Use setContent to initialize the Yjs document with content
                          editor.commands.setContent(parsed);

                          // Mark as initialized in backend
                          setTimeout(async () => {
                            try {
                              await provider.current?.markInitialized();
                              console.log('[NoteEditor] Document initialized successfully');
                            } catch (error) {
                              console.error('[NoteEditor] Failed to mark document as initialized:', error);
                            }
                          }, 1000); // Wait for Tiptap to fully sync content to Yjs
                        }
                      } catch (error) {
                        console.error('[NoteEditor] Failed to set initial content:', error);
                      }
                    }
                  }}
                  onUpdate={({ editor }) => {
                    hasUnsavedChanges.current = true;
                    debouncedUpdates(editor);

                    // Reset auto-save timer on every update
                    if (autoSaveTimerRef.current) {
                      window.clearTimeout(autoSaveTimerRef.current)
                    }
                    autoSaveTimerRef.current = window.setTimeout(() => {
                      handleSave(true)
                    }, 60000) // 1 minute
                  }}
                  slotAfter={<ImageResizer />}
                >
                  <EditorCommand className="z-50 h-auto max-h-[330px] overflow-y-auto rounded-md border border-muted bg-background px-1 py-2 shadow-md transition-all">
                    <EditorCommandEmpty className="px-2 text-muted-foreground">No results</EditorCommandEmpty>
                    <EditorCommandList>
                      {suggestionItems.map((item) => (
                        <EditorCommandItem
                          value={item.title}
                          onCommand={(val) => item.command?.(val)}
                          className="flex w-full items-center space-x-2 rounded-md px-2 py-1 text-left text-sm hover:bg-accent aria-selected:bg-accent"
                          key={item.title}
                        >
                          <div className="flex h-10 w-10 items-center justify-center rounded-md border border-muted bg-background">
                            {item.icon}
                          </div>
                          <div>
                            <p className="font-medium">{item.title}</p>
                            <p className="text-xs text-muted-foreground">{item.description}</p>
                          </div>
                        </EditorCommandItem>
                      ))}
                    </EditorCommandList>
                  </EditorCommand>

                  {/* Editor toolbar - always available with collaboration */}
                  <GenerativeMenuSwitch open={openAI} onOpenChange={setOpenAI}>
                    <Separator orientation="vertical" />
                    <NodeSelector open={openNode} onOpenChange={setOpenNode} />
                    <Separator orientation="vertical" />

                    <LinkSelector open={openLink} onOpenChange={setOpenLink} />
                    <Separator orientation="vertical" />
                    <MathSelector />
                    <Separator orientation="vertical" />
                    <TextButtons />
                    <Separator orientation="vertical" />
                    <ColorSelector open={openColor} onOpenChange={setOpenColor} />
                  </GenerativeMenuSwitch>
                </EditorContent>
              </EditorRoot>
            ) : (
              <div className="relative min-h-[500px] w-full max-w-5xl border rounded-lg transition-colors border-border/40 flex items-center justify-center">
                <div className="text-center space-y-2">
                  <Loader2 className="h-8 w-8 animate-spin mx-auto text-blue-500" />
                  <p className="text-sm text-muted-foreground">Initializing collaboration...</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

