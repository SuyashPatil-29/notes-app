import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getNote, updateNote, generateNoteVideo, deleteNoteVideo } from '@/utils/notes'
import { getUserNotebooks } from '@/utils/notebook'
import { Header } from '@/components/Header'
import { Skeleton } from '@/components/ui/skeleton'
import { NoteVideoPlayer } from '@/components/NoteVideoPlayer'
import type { AuthenticatedUser } from '@/types/backend'
import { Loader2, Calendar, Clock, Video, VideoOff, Pause } from 'lucide-react'
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
import { useRealtimeEditLock } from '@/hooks/use-realtime-edit-lock'
import { useRealtimePresenceRoom } from '@/hooks/use-realtime-presence-room'
import { Edit3, Eye } from 'lucide-react'

import GenerativeMenuSwitch from "./generative/generative-menu-switch";
// import { uploadFn } from "./image-upload";
import { TextButtons } from "./selectors/text-buttons";
import { slashCommand, suggestionItems } from "./slash-command";

import hljs from "highlight.js";
import { Button } from './ui/button'

const extensions = [...defaultExtensions, slashCommand];

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
  const [charsCount, setCharsCount] = useState();
  const [editorInstance, setEditorInstance] = useState<EditorInstance | null>(null);
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
  const [isEditMode, setIsEditMode] = useState(false);

  // Edit lock system
  const { currentEditor, requestEdit, releaseEdit, isEditing, canEdit } = useRealtimeEditLock(
    noteId ? `note-${noteId}` : ''
  );

  // Realtime presence for viewer count
  const { users: realtimeUsers } = useRealtimePresenceRoom(noteId ? `note-${noteId}` : '');
  const viewerCount = Object.keys(realtimeUsers).length;

  const { data: noteResponse, isLoading, error, refetch: refetchNote } = useQuery({
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

  // Convert markdown to ProseMirror JSON structure
  const markdownToJSON = (markdown: string) => {
    const lines = markdown.split('\n');
    const content: any[] = [];
    let i = 0;

    while (i < lines.length) {
      const line = lines[i];

      // Empty line
      if (line.trim() === '') {
        content.push({ type: 'paragraph' });
        i++;
        continue;
      }

      // Heading
      const headingMatch = line.match(/^(#{1,6})\s+(.+)$/);
      if (headingMatch) {
        const level = headingMatch[1].length;
        content.push({
          type: 'heading',
          attrs: { level },
          content: [{ type: 'text', text: headingMatch[2] }]
        });
        i++;
        continue;
      }

      // Code block
      if (line.trim().startsWith('```')) {
        const language = line.trim().slice(3).trim() || 'plaintext';
        const codeLines: string[] = [];
        i++; // Skip opening ```

        while (i < lines.length && !lines[i].trim().startsWith('```')) {
          codeLines.push(lines[i]);
          i++;
        }
        i++; // Skip closing ```

        content.push({
          type: 'codeBlock',
          attrs: { language },
          content: codeLines.length > 0
            ? [{ type: 'text', text: codeLines.join('\n') }]
            : undefined
        });
        continue;
      }

      // List items (bullet)
      if (line.match(/^[-*]\s+/)) {
        const listItems: any[] = [];
        while (i < lines.length && lines[i].match(/^[-*]\s+/)) {
          const text = lines[i].replace(/^[-*]\s+/, '');
          listItems.push({
            type: 'listItem',
            content: [{
              type: 'paragraph',
              content: [{ type: 'text', text }]
            }]
          });
          i++;
        }
        content.push({
          type: 'bulletList',
          content: listItems
        });
        continue;
      }

      // Regular paragraph
      content.push({
        type: 'paragraph',
        content: [{ type: 'text', text: line }]
      });
      i++;
    }

    return {
      type: 'doc',
      content: content.length > 0 ? content : [{ type: 'paragraph' }]
    };
  };

  // Parse note content safely - handle both JSON and plain text/markdown
  const getInitialContent = () => {
    if (!note?.content || note.content.trim() === '') {
      // Return empty document with a single paragraph
      return {
        type: 'doc',
        content: [{ type: 'paragraph' }]
      };
    }

    // Try to parse as JSON first (new format)
    try {
      const parsed = JSON.parse(note.content);
      // Validate it has proper structure
      if (parsed && parsed.type === 'doc') {
        return parsed;
      }
      // If not proper structure, return empty doc
      return {
        type: 'doc',
        content: [{ type: 'paragraph' }]
      };
    } catch (error) {
      // If it's not JSON, treat it as markdown and convert
      try {
        return markdownToJSON(note.content);
      } catch (mdError) {
        // Fallback to empty document if markdown conversion fails
        console.error('Failed to convert markdown:', mdError);
        return {
          type: 'doc',
          content: [{ type: 'paragraph' }]
        };
      }
    }
  }

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

  // Auto-release edit lock when unmounting or navigating away
  useEffect(() => {
    return () => {
      if (isEditing) {
        releaseEdit()
      }
    }
  }, [isEditing, releaseEdit])

  // Update editor editable state when isEditMode changes
  useEffect(() => {
    if (editorInstance) {
      editorInstance.setEditable(isEditMode)
    }
  }, [isEditMode, editorInstance])

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

  const handleStartEditing = async () => {
    try {
      // Request edit lock first
      const success = await requestEdit()
      if (!success) {
        toast.error('Unable to start editing. Someone else may be editing.')
        return
      }

      // Pull latest data from DB before enabling edit mode
      await refetchNote()
      
      // Enable edit mode
      setIsEditMode(true)
      toast.success('You are now editing with latest version')
    } catch (error) {
      console.error('Failed to start editing:', error)
      toast.error('Failed to start editing')
      await releaseEdit() // Release lock if something went wrong
    }
  }

  const handleStopEditing = async () => {
    try {
      // Save to DB first if there are unsaved changes
      if (editorInstance && hasUnsavedChanges.current) {
        setIsSaving(true)
        const content = editorInstance.getJSON()
        await updateNote(noteId!, {
          content: JSON.stringify(content)
        })
        hasUnsavedChanges.current = false
        setIsSaving(false)
      }
      
      // Release edit lock
      await releaseEdit()
      
      // Disable edit mode
      setIsEditMode(false)
      
      // Invalidate queries to sync all viewers
      await queryClient.invalidateQueries({ queryKey: ['note', noteId] })
      
      toast.success('Stopped editing and saved')
    } catch (error) {
      console.error('Failed to stop editing:', error)
      toast.error('Failed to save changes')
      setIsSaving(false)
    }
  }

  const handleSave = async (isAutoSave = false) => {
    if (!editorInstance || !noteId) {
      if (!isAutoSave) {
        toast.error("Editor not ready")
      }
      return
    }

    // Don't auto-save if not in edit mode
    if (isAutoSave && !isEditMode) {
      return
    }

    if (isAutoSave) {
      setIsAutoSaving(true)
    } else {
      setIsSaving(true)
    }

    try {
      const content = editorInstance.getJSON()
      await updateNote(noteId, {
        content: JSON.stringify(content)
      })

      // Invalidate cache to refetch the updated note
      queryClient.invalidateQueries({ queryKey: ['note', noteId] })
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] })

      hasUnsavedChanges.current = false

      if (isAutoSave) {
      } else {
        toast.success("Note saved successfully!")
      }
    } catch (error) {
      console.error("Failed to save note:", error)
      if (!isAutoSave) {
        toast.error("Failed to save note")
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
                  {/* Editing status badge - prominent if someone is editing */}
                  {(isEditing || currentEditor) && (
                    <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium ${isEditing
                        ? 'bg-green-500/15 text-green-700 dark:text-green-400 border border-green-500/30'
                        : 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border border-blue-500/30'
                      }`}>
                      <Edit3 className="h-3.5 w-3.5" />
                      <span>{isEditing ? "You're editing" : `${currentEditor?.name} is editing`}</span>
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
                            ? 'No viewers'
                            : viewerCount === 1
                              ? '1 viewer currently'
                              : `${viewerCount} viewers currently`}
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
                  <Video className="h-4 w-4 text-blue-600 dark:text-blue-400" />
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
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono bg-accent px-3 py-1.5 rounded">
                Path: {note.chapter.notebook.name} / {note.chapter.name} / {note.name}
              </span>
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
              {/* Edit control button */}
              {isRealtimeConfigured() && (
                <>
                  {!isEditing && canEdit && (
                    <Button
                      onClick={handleStartEditing}
                      size="sm"
                      variant="default"
                      className="gap-2"
                    >
                      <Edit3 className="h-3 w-3" />
                      Start Editing
                    </Button>
                  )}
                  {!isEditing && !canEdit && (
                    <Button
                      disabled
                      size="sm"
                      variant="outline"
                      className="gap-2 cursor-not-allowed"
                      title="Someone else is currently editing"
                    >
                      Edit (Locked)
                    </Button>
                  )}
                  {isEditing && (
                    <Button
                      onClick={handleStopEditing}
                      size="sm"
                      variant="destructive"
                      className="gap-2"
                    >
                      <Pause className="h-3 w-3" />
                      Stop Editing
                    </Button>
                  )}
                </>
              )}
              {isAutoSaving && (
                <div className="rounded-lg bg-blue-500/10 border border-blue-500/20 px-2 py-1 text-xs text-blue-600 dark:text-blue-400 flex items-center gap-1">
                  <Loader2 className="h-3 w-3 animate-spin" />
                  Auto-saving...
                </div>
              )}
              {!isEditMode && (
                <div className="rounded-lg bg-orange-500/10 border border-orange-500/20 px-2 py-1 text-xs text-orange-600 dark:text-orange-400 flex items-center gap-1">
                  Read-only
                </div>
              )}
              <div className={charsCount ? "rounded-lg bg-accent px-2 py-1 text-sm text-muted-foreground" : "hidden"}>
                {charsCount} Words
              </div>
            </div>
            <EditorRoot key={`${noteId}-${note.updatedAt}-${isEditMode ? 'edit' : 'read'}`}>
              <EditorContent
                initialContent={getInitialContent()}
                extensions={extensions}
                className={`relative min-h-[500px] w-full max-w-5xl border rounded-lg transition-colors ${isEditMode
                    ? 'border-border/40 hover:border-border/60 focus-within:border-border'
                    : 'border-border/20 cursor-default bg-muted/20'
                  }`}
                editorProps={{
                  editable: () => isEditMode,
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
                onCreate={({ editor }) => {
                  // Set editor instance when editor is first created
                  setEditorInstance(editor);
                  debouncedUpdates(editor);
                }}
                onUpdate={({ editor }) => {
                  setEditorInstance(editor);
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

                {isEditMode && (
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
                )}
              </EditorContent>
            </EditorRoot>
          </div>
        </div>
      </div>
    </div>
  )
}

