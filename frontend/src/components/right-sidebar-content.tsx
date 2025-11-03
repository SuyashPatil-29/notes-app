import { FileText, MessageSquare, Copy, RotateCcw, SquareIcon } from "lucide-react"
import { toast } from "sonner"
import { useNavigate, useLocation } from "react-router-dom"
import {
    RightSidebar,
    RightSidebarContent as RightSidebarContentWrapper,
    RightSidebarHeader,
    RightSidebarMenu,
    RightSidebarMenuItem,
    RightSidebarMenuButton,
    RightSidebarRail,
    useRightSidebar,
} from "@/components/ui/right-sidebar"
import { Skeleton } from "@/components/ui/skeleton"
import { useChat } from "@ai-sdk/react"
import React, { useState, useRef, useEffect, useMemo } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import {
    Conversation,
    ConversationContent,
    ConversationScrollButton,
} from "@/components/ai/conversation"
import { Message, MessageContent, MessageAvatar } from "@/components/ai/message"
import {
    PromptInput,
    PromptInputSubmit,
    PromptInputToolbar,
    PromptInputTools,
    PromptInputModelSelect,
    PromptInputModelSelectTrigger,
    PromptInputModelSelectContent,
    PromptInputModelSelectItem,
    PromptInputModelSelectValue,
} from "@/components/ai/prompt-input"
import { Response } from "@/components/ai/response"
import { Reasoning, ReasoningTrigger, ReasoningContent } from "@/components/ai/reasoning"
import { Task, TaskTrigger, TaskContent, TaskItem } from "@/components/ai/task"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { useUser } from "@/hooks/auth"
import { useAuth } from "@clerk/clerk-react"
import api from "@/utils/api"
import { SelectGroup, SelectLabel, SelectSeparator } from "@/components/ui/select"
import { MentionTagsInput } from "@/components/ai/mention-tags-input"
import { Suggestions, Suggestion } from "@/components/ai/suggestion"
import { getUserNotebooks } from "@/utils/notebook"
import type { Notebook } from "@/types/backend"
import { useOrganizationContext } from "@/contexts/OrganizationContext"
import { getPreviewText } from "@/utils/markdown"

// Special marker format for mentions in text: @[noteId|noteName]
const MENTION_REGEX = /@\[([^\|]+)\|([^\]]+)\]/g;

// Helper function to render message content with mention badges
function renderMessageWithMentions(text: string): React.ReactNode {
    const parts: React.ReactNode[] = []
    let lastIndex = 0
    let match: RegExpExecArray | null

    // Create a new regex instance to avoid lastIndex issues
    const regex = new RegExp(MENTION_REGEX.source, 'g')
    
    while ((match = regex.exec(text)) !== null) {
        const [, noteId, noteName] = match
        const matchStart = match.index

        // Add text before the mention
        if (matchStart > lastIndex) {
            parts.push(text.slice(lastIndex, matchStart))
        }

        // Add the mention badge
        parts.push(
            <span
                key={`mention-${noteId}-${matchStart}`}
                className="inline-flex items-center gap-1 px-2 py-0.5 mx-0.5 rounded-md bg-primary-foreground/20 text-primary-foreground text-sm font-medium whitespace-nowrap"
            >
                <FileText className="h-3 w-3 shrink-0" />
                {noteName}
            </span>
        )

        lastIndex = regex.lastIndex
    }

    // Add remaining text after the last mention
    if (lastIndex < text.length) {
        parts.push(text.slice(lastIndex))
    }

    // If no mentions were found, return the original text
    return parts.length === 0 ? text : parts
}

// Helper function to render tool output with nice formatting
function renderToolOutput(toolName: string, result: any, onNavigateToNote?: (notebookId: string, chapterId: string, noteId: string) => void) {
    // Handle error case
    if (result.error) {
        return (
            <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                <p className="font-semibold">Error</p>
                <p>{result.error}</p>
            </div>
        )
    }

    // Render based on tool name
    switch (toolName) {
        case "searchNotes":
            return (
                <div className="space-y-2 text-sm">
                    {result.count > 0 ? (
                        <>
                            <p className="font-semibold text-foreground">
                                Found {result.count} note{result.count !== 1 ? 's' : ''} matching "{result.query}"
                            </p>
                            <div className="space-y-2">
                                {result.results.map((note: any, idx: number) => (
                                    <div key={idx} className="p-2 bg-muted/50 rounded border space-y-2">
                                        <div>
                                            <p className="font-medium text-foreground">{note.name}</p>
                                            <p className="text-xs text-muted-foreground">
                                                {note.notebookName} → {note.chapterName}
                                            </p>
                                            <p className="mt-1 text-xs text-muted-foreground">
                                                {note.content ? getPreviewText(note.content, 100) : (note.preview || 'Empty note')}
                                            </p>
                                            <p className="mt-1 text-xs text-muted-foreground">Updated: {note.updatedAt}</p>
                                        </div>
                                        {onNavigateToNote && note.notebookId && note.chapterId && note.id && (
                                            <button
                                                onClick={() => onNavigateToNote(note.notebookId, note.chapterId, note.id)}
                                                className="w-full px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-md transition-colors"
                                            >
                                                Open Note →
                                            </button>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </>
                    ) : (
                        <p className="text-muted-foreground">{result.message}</p>
                    )}
                </div>
            )

        case "listNotebooks":
            return (
                <div className="space-y-2 text-sm">
                    {result.count > 0 ? (
                        <>
                            <p className="font-semibold text-foreground">Found {result.count} notebook{result.count !== 1 ? 's' : ''}</p>
                            <div className="space-y-1">
                                {result.notebooks.map((notebook: any, idx: number) => (
                                    <div key={idx} className="p-2 bg-muted/50 rounded border">
                                        <p className="font-medium text-foreground">{notebook.name}</p>
                                        <p className="text-xs text-muted-foreground">
                                            {notebook.chapterCount} chapter{notebook.chapterCount !== 1 ? 's' : ''} •
                                            Updated: {notebook.updatedAt}
                                        </p>
                                    </div>
                                ))}
                            </div>
                        </>
                    ) : (
                        <p className="text-muted-foreground">{result.message}</p>
                    )}
                </div>
            )

        case "listChapters":
            return (
                <div className="space-y-2 text-sm">
                    <p className="font-semibold text-foreground">
                        Chapters in "{result.notebookName}"
                    </p>
                    {result.count > 0 ? (
                        <div className="space-y-1">
                            {result.chapters.map((chapter: any, idx: number) => (
                                <div key={idx} className="p-2 bg-muted/50 rounded border">
                                    <p className="font-medium text-foreground">{chapter.name}</p>
                                    <p className="text-xs text-muted-foreground">
                                        {chapter.noteCount} note{chapter.noteCount !== 1 ? 's' : ''} •
                                        Created: {chapter.createdAt}
                                    </p>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <p className="text-muted-foreground">{result.message}</p>
                    )}
                </div>
            )

        case "getNoteContent":
            return (
                <div className="space-y-2 text-sm">
                    <div className="p-3 bg-muted/50 rounded border space-y-2">
                        <div>
                            <p className="font-semibold text-foreground">{result.name}</p>
                            <p className="text-xs text-muted-foreground">
                                {result.notebookName} → {result.chapterName}
                            </p>
                            <p className="text-xs text-muted-foreground">
                                Created: {result.createdAt} • Updated: {result.updatedAt}
                            </p>
                        </div>
                        {onNavigateToNote && result.notebookId && result.chapterId && result.id && (
                            <button
                                onClick={() => onNavigateToNote(result.notebookId, result.chapterId, result.id)}
                                className="w-full px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-md transition-colors"
                            >
                                Open Note →
                            </button>
                        )}
                    </div>
                    <div className="p-3 bg-muted/50 rounded border max-h-96 overflow-y-auto">
                        <p className="text-xs whitespace-pre-wrap">{getPreviewText(result.content || '', 500)}</p>
                    </div>
                </div>
            )

        case "listNotesInChapter":
            return (
                <div className="space-y-2 text-sm">
                    <p className="font-semibold text-foreground">
                        Notes in "{result.chapterName}" ({result.notebookName})
                    </p>
                    {result.count > 0 ? (
                        <div className="space-y-1">
                            {result.notes.map((note: any, idx: number) => (
                                <div key={idx} className="p-2 bg-muted/50 rounded border">
                                    <p className="font-medium text-foreground">{note.name}</p>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        {note.content ? getPreviewText(note.content, 100) : (note.preview || 'Empty note')}
                                    </p>
                                    <p className="text-xs text-muted-foreground mt-1">Updated: {note.updatedAt}</p>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <p className="text-muted-foreground">{result.message}</p>
                    )}
                </div>
            )

        case "createNote":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-green-500/10 border border-green-500/20 rounded-md">
                                <p className="font-semibold text-primary dark:text-primary">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border space-y-2">
                                <div>
                                    <p className="font-medium text-foreground">{result.noteName}</p>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        {result.notebookName} → {result.chapterName}
                                    </p>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        Size: {result.contentSize} characters • Created: {result.createdAt}
                                    </p>
                                </div>
                                {onNavigateToNote && result.notebookId && result.chapterId && result.noteId && (
                                    <button
                                        onClick={() => onNavigateToNote(result.notebookId, result.chapterId, result.noteId)}
                                        className="w-full px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-md transition-colors"
                                    >
                                        Open Note →
                                    </button>
                                )}
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to create note</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "moveNote":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-blue-500/10 border border-blue-500/20 rounded-md">
                                <p className="font-semibold text-blue-600 dark:text-blue-400">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border space-y-2">
                                <div>
                                    <p className="font-medium text-foreground">{result.noteName}</p>
                                    <div className="mt-2 space-y-1">
                                        <p className="text-xs text-muted-foreground">
                                            <span className="font-semibold">From:</span> {result.fromNotebook} → {result.fromChapter}
                                        </p>
                                        <p className="text-xs text-muted-foreground">
                                            <span className="font-semibold">To:</span> {result.toNotebook} → {result.toChapter}
                                        </p>
                                    </div>
                                </div>
                                {onNavigateToNote && result.notebookId && result.chapterId && result.noteId && (
                                    <button
                                        onClick={() => onNavigateToNote(result.notebookId, result.chapterId, result.noteId)}
                                        className="w-full px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-md transition-colors"
                                    >
                                        Open Note →
                                    </button>
                                )}
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to move note</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "moveChapter":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-blue-500/10 border border-blue-500/20 rounded-md">
                                <p className="font-semibold text-blue-600 dark:text-blue-400">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="font-medium text-foreground">{result.chapterName}</p>
                                <div className="mt-2 space-y-1">
                                    <p className="text-xs text-muted-foreground">
                                        <span className="font-semibold">From:</span> {result.fromNotebook}
                                    </p>
                                    <p className="text-xs text-muted-foreground">
                                        <span className="font-semibold">To:</span> {result.toNotebook}
                                    </p>
                                    <p className="text-xs text-primary mt-2">
                                        📦 Moved with {result.notesCount} note{result.notesCount !== 1 ? 's' : ''}
                                    </p>
                                </div>
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to move chapter</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "renameNote":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-purple-500/10 border border-purple-500/20 rounded-md">
                                <p className="font-semibold text-purple-600 dark:text-purple-400">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border space-y-2">
                                <div>
                                    <p className="text-xs text-muted-foreground">
                                        <span className="line-through">{result.oldName}</span> → <span className="font-medium text-foreground">{result.newName}</span>
                                    </p>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        {result.notebookName} → {result.chapterName}
                                    </p>
                                </div>
                                {onNavigateToNote && result.notebookId && result.chapterId && result.noteId && (
                                    <button
                                        onClick={() => onNavigateToNote(result.notebookId, result.chapterId, result.noteId)}
                                        className="w-full px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-md transition-colors"
                                    >
                                        Open Note →
                                    </button>
                                )}
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to rename note</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "deleteNote":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-md">
                                <p className="font-semibold text-red-600 dark:text-red-400">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="font-medium text-foreground">{result.noteName}</p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    {result.notebookName} → {result.chapterName}
                                </p>
                                <p className="text-xs text-orange-600 dark:text-orange-400 mt-2">
                                    This note has been permanently deleted.
                                </p>
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to delete note</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "updateNoteContent":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-green-500/10 border border-green-500/20 rounded-md">
                                <p className="font-semibold text-primary dark:text-primary">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border space-y-2">
                                <div>
                                    <p className="font-medium text-foreground">{result.noteName}</p>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        {result.notebookName} → {result.chapterName}
                                    </p>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        New size: {result.contentSize} characters • Updated: {result.updatedAt}
                                    </p>
                                </div>
                                {onNavigateToNote && result.notebookId && result.chapterId && result.noteId && (
                                    <button
                                        onClick={() => onNavigateToNote(result.notebookId, result.chapterId, result.noteId)}
                                        className="w-full px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-md transition-colors"
                                    >
                                        Open Note →
                                    </button>
                                )}
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to update note</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "createNotebook":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-green-500/10 border border-green-500/20 rounded-md">
                                <p className="font-semibold text-primary dark:text-primary">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="font-medium text-foreground">📚 {result.name}</p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    Created: {result.createdAt}
                                </p>
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to create notebook</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "createChapter":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-green-500/10 border border-green-500/20 rounded-md">
                                <p className="font-semibold text-primary dark:text-primary">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="font-medium text-foreground">📖 {result.chapterName}</p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    in {result.notebookName}
                                </p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    Created: {result.createdAt}
                                </p>
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to create chapter</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "renameNotebook":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-purple-500/10 border border-purple-500/20 rounded-md">
                                <p className="font-semibold text-purple-600 dark:text-purple-400">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="text-xs text-muted-foreground">
                                    <span className="line-through">{result.oldName}</span> → <span className="font-medium text-foreground">📚 {result.newName}</span>
                                </p>
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to rename notebook</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        case "renameChapter":
            return (
                <div className="space-y-2 text-sm">
                    {result.success ? (
                        <>
                            <div className="p-3 bg-purple-500/10 border border-purple-500/20 rounded-md">
                                <p className="font-semibold text-purple-600 dark:text-purple-400">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="text-xs text-muted-foreground">
                                    <span className="line-through">{result.oldName}</span> → <span className="font-medium text-foreground">📖 {result.newName}</span>
                                </p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    in {result.notebookName}
                                </p>
                            </div>
                        </>
                    ) : (
                        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                            <p className="font-semibold">Failed to rename chapter</p>
                            <p>{result.error || "Unknown error"}</p>
                        </div>
                    )}
                </div>
            )

        default:
            // Fallback to JSON for unknown tools
            return (
                <pre className="p-2 text-xs bg-muted/50 rounded overflow-x-auto">
                    {JSON.stringify(result, null, 2)}
                </pre>
            )
    }
}

const modelToProvider = {
    // OpenAI models (cheap to expensive)
    "gpt-5-mini": "openai",
    "gpt-5": "openai",
    "gpt-4o-mini": "openai",
    "gpt-4o": "openai",
    "gpt-3.5-turbo": "openai",

    // Anthropic models (cheap to expensive)
    "claude-haiku-4.5": "anthropic",
    "claude-haiku-3.5": "anthropic",
    "claude-sonnet-4.5": "anthropic",
    "claude-sonnet-3.5": "anthropic",

    // Google models (cheap to expensive)
    "gemini-2.5-flash-lite": "google",
    "gemini-2.5-flash": "google",
    "gemini-2.0-flash-lite": "google",
    "gemini-2.0-flash": "google",
    "gemini-2.5-pro": "google",
} as const;


type Model = keyof typeof modelToProvider

interface ApiKeyStatus {
    openai: boolean;
    anthropic: boolean;
    google: boolean;
}

interface MentionedNote {
    id: string
    name: string
    content: string
    chapterName: string
    notebookName: string
}

export function RightSidebarContent() {
    const { user, loading: userLoading } = useUser()
    const { getToken } = useAuth()
    const { activeOrg } = useOrganizationContext()
    const navigate = useNavigate()
    const location = useLocation()

    // Extract noteId from URL path: /:notebookId/:chapterId/:noteId
    const noteId = useMemo(() => {
        const pathParts = location.pathname.split('/').filter(Boolean)
        // If we have 3 parts, the last one is the noteId
        if (pathParts.length === 3) {
            return pathParts[2]
        }
        return null
    }, [location.pathname])

    const [model, setModel] = useState<Model>("gpt-5-mini")
    const [files, setFiles] = useState<FileList | null>(null)
    const [authToken, setAuthToken] = useState<string | null>(null)
    const { open } = useRightSidebar()
    const inputRef = useRef<HTMLTextAreaElement>(null)

    // Fetch auth token when component mounts or user changes
    useEffect(() => {
        const fetchToken = async () => {
            const token = await getToken()
            setAuthToken(token || '')
        }
        fetchToken()
    }, [getToken, user]) // Add user dependency to refetch when user changes
    const queryClient = useQueryClient()
    const processedNoteIds = useRef<Set<string>>(new Set())

    // Mention notes state
    const [mentionedNotes, setMentionedNotes] = useState<MentionedNote[]>([])
    // Fetch notebooks to get all notes for mentions
    const { data: notebooks, isLoading: notebooksLoading } = useQuery<Notebook[]>({
        queryKey: ['userNotebooks', activeOrg?.id],
        queryFn: () => getUserNotebooks(activeOrg?.id),
        enabled: open && !!user,
    })

    // Flatten all notes for mention selector
    const allNotes = useMemo(() => {
        if (!notebooks) return []
        return notebooks.flatMap((notebook: Notebook) =>
            notebook.chapters?.flatMap((chapter: any) =>
                chapter.notes?.map((note: any) => ({
                    id: note.id,
                    name: note.name,
                    content: note.content,
                    chapterName: chapter.name,
                    notebookName: notebook.name,
                })) || []
            ) || []
        )
    }, [notebooks])

    // Use useQuery to manage API key status so it can react to invalidations
    const { data: fetchedApiKeyStatus } = useQuery<ApiKeyStatus>({
        queryKey: ['api-key-status'],
        queryFn: async () => {
            const response = await api.get("/settings/ai-credentials");
            const providers = response.data.providers || {};
            return {
                openai: providers.openai || false,
                anthropic: providers.anthropic || false,
                google: providers.google || false,
            };
        },
        enabled: open, // Only fetch when sidebar is open
        staleTime: 0, // Always refetch when invalidated
    });

    // Default to false if data is not yet loaded
    const apiKeyStatus = fetchedApiKeyStatus || { openai: false, anthropic: false, google: false };

    // Check if the selected model's API key is configured
    const selectedProvider = modelToProvider[model];
    const hasSelectedApiKey = apiKeyStatus[selectedProvider as keyof ApiKeyStatus] || false;

    const chatFetch = React.useCallback(async (input: RequestInfo | URL, init?: RequestInit) => {
        try {
            const freshToken = await getToken();
            const headers = new Headers(init?.headers);
            if (freshToken) {
                headers.set('Authorization', `Bearer ${freshToken}`);
            }
            return fetch(input, {
                ...init,
                headers,
            });
        } catch (error) {
            return fetch(input, init);
        }
    }, []);

    // Memoize body object to prevent re-initialization
    const chatBody = React.useMemo(() => ({
        provider: modelToProvider[model],
        model,
        organizationId: activeOrg?.id || null,
    }), [model, activeOrg?.id]);

    const onError = React.useCallback((error: Error) => {
        const errorMessage = error.message || "An error occurred while processing your request";
        toast.error(errorMessage, {
            description: errorMessage.includes("API key")
                ? "You can update your API key in Profile settings"
                : undefined,
            duration: 5000,
        });
    }, []);

    const { messages, input, handleSubmit, status, setInput, append, setMessages, stop, reload, error } = useChat({
        key: authToken || 'no-token', // Force re-initialization when token changes
        api: "http://localhost:8080/api/chat",
        body: chatBody,
        credentials: "include",
        fetch: chatFetch,
        maxSteps: 10, // Enable multi-step tool calling
        onError,
    })


    // Handle mentioned notes change from MentionTagsInput
    const handleMentionedNotesChange = React.useCallback((notes: MentionedNote[]) => {
        setMentionedNotes(notes)
    }, [])

    // Navigate to a note
    const handleNavigateToNote = React.useCallback((notebookId: string, chapterId: string, noteId: string) => {
        if (!notebookId || !chapterId || !noteId) {
            toast.error("Cannot navigate: Missing notebook or chapter information", { duration: 3000 })
            return
        }

        const path = `/${notebookId}/${chapterId}/${noteId}`
        navigate(path)
        toast.success("Navigating to note...", { duration: 2000 })
    }, [navigate])

    // Clear chat history
    const handleClearChat = () => {
        setMessages([])
        toast.success("Chat cleared", { duration: 2000 })
    }

    // Copy response text to clipboard
    const handleCopyResponse = (text: string) => {
        navigator.clipboard.writeText(text).then(() => {
            toast.success("Copied to clipboard", { duration: 2000 })
        }).catch(() => {
            toast.error("Failed to copy", { duration: 2000 })
        })
    }

    // Handle suggestion click
    const handleSuggestionClick = (suggestion: string) => {
        setInput(suggestion)
        // Auto-focus the input and position cursor at the end
        setTimeout(() => {
            if (inputRef.current) {
                inputRef.current.focus()
                // Position cursor at the end of the text
                const length = suggestion.length
                inputRef.current.setSelectionRange(length, length)
            }
        }, 50)
    }

    const handleSubmitWithFiles = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()

        if (!input.trim() && !files) return

        // Build the message content with mentioned notes
        let messageContent = input

        if (mentionedNotes.length > 0) {
            const notesContext = mentionedNotes.map(note =>
                `\n\n<<<REFERENCED_NOTE>>>\n**Referenced Note: ${note.name}**\n*Location: ${note.notebookName} → ${note.chapterName}*\n\n${note.content}`
            ).join('\n')

            messageContent = `${input}${notesContext}`
        }

        if (files) {
            handleSubmit(e, {
                experimental_attachments: files,
            })
            setFiles(null)
        } else {
            // Use append to send the enriched message
            await append({
                role: 'user',
                content: messageContent,
            })
            setInput('')
        }

        // Clear mentioned notes after sending
        setMentionedNotes([])
    }

    // Focus input when sidebar opens
    useEffect(() => {
        if (open && inputRef.current) {
            // Small delay to ensure the sidebar animation is complete
            const timer = setTimeout(() => {
                inputRef.current?.focus()
            }, 150)
            return () => clearTimeout(timer)
        }
    }, [open])

    // Refocus after message is sent (when status changes back to ready)
    useEffect(() => {
        if (status === "ready" && messages.length > 0 && inputRef.current) {
            const timer = setTimeout(() => {
                inputRef.current?.focus()
            }, 50)
            return () => clearTimeout(timer)
        }
    }, [status, messages.length])

    // Auto-navigate to notes being created or updated
    useEffect(() => {
        // Only process when there are messages
        if (messages.length === 0) return

        // Get the last assistant message
        const lastMessage = messages[messages.length - 1]
        if (lastMessage.role !== "assistant") return

        // Check if this message has tool invocations
        lastMessage.parts?.forEach((part) => {
            if (part.type === "tool-invocation") {
                const toolInvocation = part.toolInvocation as any

                // For createNote: navigate when result is available (note just created)
                if (
                    toolInvocation.toolName === "createNote" &&
                    toolInvocation.result?.success &&
                    toolInvocation.result?.noteId &&
                    toolInvocation.result?.notebookId &&
                    toolInvocation.result?.chapterId
                ) {
                    const noteId = toolInvocation.result.noteId
                    const operationKey = `createNote-${noteId}`

                    if (!processedNoteIds.current.has(operationKey)) {
                        processedNoteIds.current.add(operationKey)
                        setTimeout(() => {
                            handleNavigateToNote(
                                toolInvocation.result.notebookId,
                                toolInvocation.result.chapterId,
                                noteId
                            )
                        }, 500)
                    }
                }

                // For updateNoteContent: navigate IMMEDIATELY when we see the args (before streaming content)
                // This happens as soon as the AI decides to update a note
                if (
                    toolInvocation.toolName === "updateNoteContent" &&
                    toolInvocation.args?.noteId
                ) {
                    const noteId = toolInvocation.args.noteId
                    const operationKey = `updateNoteContent-${noteId}-nav`

                    // Navigate as soon as we see the noteId in args
                    if (!processedNoteIds.current.has(operationKey)) {
                        processedNoteIds.current.add(operationKey)

                        let navigationAttempted = false

                        // Try to get navigation info from cached note data
                        const noteData = queryClient.getQueryData(['note', noteId]) as any

                        if (noteData?.data?.chapter?.notebook?.id && noteData?.data?.chapter?.id) {
                            // Navigate immediately using cached data
                            handleNavigateToNote(
                                noteData.data.chapter.notebook.id,
                                noteData.data.chapter.id,
                                noteId
                            )
                            navigationAttempted = true
                        } else if (toolInvocation.result?.notebookId && toolInvocation.result?.chapterId) {
                            // If result is available with full path info, use it
                            handleNavigateToNote(
                                toolInvocation.result.notebookId,
                                toolInvocation.result.chapterId,
                                noteId
                            )
                            navigationAttempted = true
                        } else {
                            // If note isn't cached, try to find it in notebooks cache
                            const notebooks = queryClient.getQueryData(['userNotebooks']) as any[]
                            if (notebooks) {
                                for (const notebook of notebooks) {
                                    for (const chapter of notebook.chapters || []) {
                                        const note = chapter.notes?.find((n: any) => n.id === noteId)
                                        if (note) {
                                            handleNavigateToNote(notebook.id, chapter.id, noteId)
                                            navigationAttempted = true
                                            break
                                        }
                                    }
                                    if (navigationAttempted) break
                                }
                            }
                        }

                    }
                }

                // For getNoteContent: Don't auto-navigate - let user click the button
                // The result has an "Open Note" button for manual navigation

                // For moveNote: navigate to the note's new location when result is available
                if (
                    toolInvocation.toolName === "moveNote" &&
                    toolInvocation.result?.success &&
                    toolInvocation.result?.noteId &&
                    toolInvocation.result?.notebookId &&
                    toolInvocation.result?.chapterId
                ) {
                    const noteId = toolInvocation.result.noteId
                    const operationKey = `moveNote-${noteId}`

                    if (!processedNoteIds.current.has(operationKey)) {
                        processedNoteIds.current.add(operationKey)
                        setTimeout(() => {
                            handleNavigateToNote(
                                toolInvocation.result.notebookId,
                                toolInvocation.result.chapterId,
                                noteId
                            )
                        }, 500)
                    }
                }

                // For renameNote: navigate to the note when result is available
                if (
                    toolInvocation.toolName === "renameNote" &&
                    toolInvocation.result?.success &&
                    toolInvocation.result?.noteId &&
                    toolInvocation.result?.notebookId &&
                    toolInvocation.result?.chapterId
                ) {
                    const noteId = toolInvocation.result.noteId
                    const operationKey = `renameNote-${noteId}`

                    if (!processedNoteIds.current.has(operationKey)) {
                        processedNoteIds.current.add(operationKey)
                        setTimeout(() => {
                            handleNavigateToNote(
                                toolInvocation.result.notebookId,
                                toolInvocation.result.chapterId,
                                noteId
                            )
                        }, 500)
                    }
                }
            }
        })
    }, [messages, handleNavigateToNote, queryClient])

    // Note: Real-time streaming preview removed to avoid flickering
    // The note will update automatically once the updateNoteContent tool completes

    // Invalidate cache when notes are modified
    useEffect(() => {
        const lastMessage = messages[messages.length - 1]
        if (lastMessage?.role === 'assistant' && lastMessage.id) {
            lastMessage.parts?.forEach((part: any, partIndex: number) => {
                if (part.type === 'tool-invocation') {
                    const toolInvocation = part.toolInvocation as any
                    const toolName = toolInvocation.toolName

                    const modifyingTools = ['createNote', 'moveNote', 'moveChapter', 'renameNote', 'deleteNote', 'updateNoteContent', 'createNotebook', 'createChapter', 'renameNotebook', 'renameChapter']
                    const isModifyingTool = modifyingTools.includes(toolName)
                    const hasSuccessResult = toolInvocation.result?.success === true

                    if (isModifyingTool && hasSuccessResult) {
                        const operationKey = `${lastMessage.id}-${partIndex}-${toolName}`

                        if (!processedNoteIds.current.has(operationKey)) {
                            processedNoteIds.current.add(operationKey)

                            // Toast messages for different operations
                            const toastMessages: Record<string, { loading: string, success: string, error: string }> = {
                                createNote: {
                                    loading: 'Creating note...',
                                    success: 'Note created!',
                                    error: 'Failed to refresh'
                                },
                                moveNote: {
                                    loading: 'Moving note...',
                                    success: 'Note moved!',
                                    error: 'Failed to refresh'
                                },
                                moveChapter: {
                                    loading: 'Moving chapter...',
                                    success: 'Chapter moved!',
                                    error: 'Failed to refresh'
                                },
                                renameNote: {
                                    loading: 'Renaming note...',
                                    success: 'Note renamed!',
                                    error: 'Failed to refresh'
                                },
                                deleteNote: {
                                    loading: 'Deleting note...',
                                    success: 'Note deleted!',
                                    error: 'Failed to refresh'
                                },
                                updateNoteContent: {
                                    loading: 'Updating note...',
                                    success: 'Note updated!',
                                    error: 'Failed to refresh'
                                },
                                createNotebook: {
                                    loading: 'Creating notebook...',
                                    success: 'Notebook created!',
                                    error: 'Failed to refresh'
                                },
                                createChapter: {
                                    loading: 'Creating chapter...',
                                    success: 'Chapter created!',
                                    error: 'Failed to refresh'
                                },
                                renameNotebook: {
                                    loading: 'Renaming notebook...',
                                    success: 'Notebook renamed!',
                                    error: 'Failed to refresh'
                                },
                                renameChapter: {
                                    loading: 'Renaming chapter...',
                                    success: 'Chapter renamed!',
                                    error: 'Failed to refresh'
                                },
                            }

                            const messages = toastMessages[toolName] || {
                                loading: 'Processing...',
                                success: 'Changes saved!',
                                error: 'Failed to refresh'
                            }

                            // Build the list of queries to refetch
                            const queriesToRefetch: Promise<any>[] = [
                                queryClient.refetchQueries({ queryKey: ['userNotebooks'] }),
                                queryClient.refetchQueries({ queryKey: ['notes'] }),
                            ]

                            const noteId = toolInvocation.result?.noteId
                            if (noteId && ['createNote', 'updateNoteContent', 'moveNote', 'renameNote'].includes(toolName)) {
                                const invalidatePromise = queryClient.invalidateQueries({ 
                                    queryKey: ['note', noteId],
                                    refetchType: 'active'
                                })
                                
                                queriesToRefetch.push(invalidatePromise)
                            }

                            // Use toast.promise for better UX with loading state
                            toast.promise(
                                Promise.all(queriesToRefetch),
                                messages
                            )
                        }
                    }
                }
            })
        }
    }, [messages, queryClient])

    // Show skeleton while loading user
    if (userLoading) {
        return (
            <RightSidebar collapsible="offcanvas">
                <RightSidebarHeader className="h-16 border-b">
                    <RightSidebarMenu>
                        <RightSidebarMenuItem>
                            <RightSidebarMenuButton size="lg">
                                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                                    <MessageSquare className="size-4" />
                                </div>
                                <div className="grid flex-1 text-left text-sm leading-tight">
                                    <span className="truncate font-semibold">AI Chat</span>
                                </div>
                            </RightSidebarMenuButton>
                        </RightSidebarMenuItem>
                    </RightSidebarMenu>
                </RightSidebarHeader>
                <RightSidebarContentWrapper className="flex flex-col h-full p-6">
                    <div className="space-y-4">
                        {/* Message skeleton items */}
                        <div className="flex items-start gap-3">
                            <Skeleton className="h-8 w-8 rounded-full" />
                            <div className="flex-1 space-y-2">
                                <Skeleton className="h-4 w-3/4" />
                                <Skeleton className="h-4 w-full" />
                                <Skeleton className="h-4 w-2/3" />
                            </div>
                        </div>
                        <div className="flex items-start gap-3">
                            <Skeleton className="h-8 w-8 rounded-full" />
                            <div className="flex-1 space-y-2">
                                <Skeleton className="h-4 w-2/3" />
                                <Skeleton className="h-4 w-full" />
                            </div>
                        </div>
                    </div>
                </RightSidebarContentWrapper>
                <RightSidebarRail />
            </RightSidebar>
        )
    }

    return (
        <RightSidebar collapsible="offcanvas">
            <RightSidebarHeader className="h-16 border-b">
                <RightSidebarMenu>
                    <RightSidebarMenuItem>
                        <RightSidebarMenuButton size="lg">
                            <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                                <MessageSquare className="size-4" />
                            </div>
                            <div className="grid flex-1 text-left text-sm leading-tight">
                                <span className="truncate font-semibold">AI Chat</span>
                            </div>
                            {messages.length > 0 && (
                                <Button
                                    onClick={handleClearChat}
                                    title="Clear chat"
                                >
                                    Clear Chat
                                </Button>
                            )}
                        </RightSidebarMenuButton>
                    </RightSidebarMenuItem>
                </RightSidebarMenu>
            </RightSidebarHeader>
            <RightSidebarContentWrapper className="flex flex-col h-full p-0">
                <Conversation className="flex-1">
                    <ConversationContent className="px-6">
                        {messages.length === 0 && (
                            <div className="flex flex-col items-center justify-center h-full text-center space-y-4 px-4">
                                <div className="rounded-full bg-muted p-4">
                                    <MessageSquare className="size-8 text-muted-foreground" />
                                </div>
                                <div className="space-y-2">
                                    <h3 className="font-semibold text-lg">
                                        {hasSelectedApiKey ? "Start a conversation" : "Set up your API keys"}
                                    </h3>
                                    <p className="text-sm text-muted-foreground">
                                        {hasSelectedApiKey
                                            ? "Ask me about your notes, notebooks, or anything else!"
                                            : "Configure your API keys in settings to start chatting"
                                        }
                                    </p>
                                </div>
                                {hasSelectedApiKey && (
                                    <div className="w-full space-y-3 mt-2">
                                        <p className="text-xs text-muted-foreground">Try these suggestions:</p>
                                        <Suggestions className="justify-center w-full">
                                            <Suggestion
                                                suggestion="What notes do I have?"
                                                onClick={handleSuggestionClick}
                                            />
                                        </Suggestions>
                                        <Suggestions className="justify-center w-full">
                                            <Suggestion
                                                suggestion="Search for notes about..."
                                                onClick={handleSuggestionClick}
                                            />
                                        </Suggestions>
                                        <Suggestions className="justify-center w-full">
                                            <Suggestion
                                                suggestion="List all my notebooks"
                                                onClick={handleSuggestionClick}
                                            />
                                        </Suggestions>
                                        <Suggestions className="justify-center w-full">
                                            <Suggestion
                                                suggestion="Create a new note"
                                                onClick={handleSuggestionClick}
                                            />
                                        </Suggestions>
                                        <Suggestions className="justify-center w-full">
                                            <Suggestion
                                                suggestion="Help me organize my notes"
                                                onClick={handleSuggestionClick}
                                            />
                                        </Suggestions>
                                        <Suggestions className="justify-center w-full">
                                            <Suggestion
                                                suggestion="Show me my recent notes"
                                                onClick={handleSuggestionClick}
                                            />
                                        </Suggestions>
                                    </div>
                                )}
                            </div>
                        )}
                        {messages.map((m) => {
                            const role = m.role

                            // Skip data messages
                            if (role === "data") return null

                            if (role === "user") {
                                // Parse content to separate message from referenced notes
                                const content = m.content || ''
                                const parts = content.split('\n\n<<<REFERENCED_NOTE>>>\n')
                                const mainMessage = parts[0]
                                const referencedNotes = parts.slice(1)

                                return (
                                    <Message from={role} key={m.id}>
                                        <MessageAvatar role={role} image={user?.imageUrl} />
                                        <MessageContent>
                                            <div className="space-y-2">
                                                {referencedNotes.length > 0 && (
                                                    <div className="flex flex-wrap items-center gap-1.5">
                                                        {referencedNotes.map((noteBlock, idx) => {
                                                            const lines = noteBlock.split('\n')
                                                            const noteName = lines[0]?.replace('**Referenced Note: ', '').replace('**', '') || ''

                                                            return (
                                                                <React.Fragment key={idx}>
                                                                    <div className="inline-flex items-center gap-1.5 px-2 py-1 bg-muted/50 rounded-md border border-border/50 text-xs">
                                                                        <FileText className="h-3 w-3 text-muted-foreground shrink-0" />
                                                                        <span className="font-medium text-foreground">{noteName}</span>
                                                                    </div>
                                                                </React.Fragment>
                                                            )
                                                        })}
                                                    </div>
                                                )}
                                                <div className="text-sm">{renderMessageWithMentions(mainMessage)}</div>
                                            </div>
                                        </MessageContent>
                                    </Message>
                                )
                            }

                            if (role === "assistant") {
                                // Group parts by type - separate tool invocations from text/reasoning
                                const toolParts = m.parts?.filter(p => p.type === "tool-invocation") || []
                                const nonToolParts = m.parts?.filter(p => p.type !== "tool-invocation") || []

                                return (
                                    <React.Fragment key={m.id}>
                                        {/* Render tool invocations in their own message */}
                                        {toolParts.length > 0 && (
                                            <Message from={role}>
                                                <MessageAvatar role={role} />
                                                <MessageContent>
                                                    {toolParts.map((part, index) => {
                                                        const toolInvocation = part.toolInvocation as any
                                                        const isComplete = !!toolInvocation.result
                                                        const taskTitle = isComplete
                                                            ? `✓ ${toolInvocation.toolName}`
                                                            : `⏳ ${toolInvocation.toolName}...`

                                                        return (
                                                            <Task key={index} defaultOpen={false} className="mt-3 first:mt-0">
                                                                <TaskTrigger title={taskTitle} />
                                                                <TaskContent>
                                                                    {toolInvocation.result && (
                                                                        <TaskItem>
                                                                            {renderToolOutput(toolInvocation.toolName, toolInvocation.result, handleNavigateToNote)}
                                                                        </TaskItem>
                                                                    )}
                                                                    {!toolInvocation.result && toolInvocation.args && (
                                                                        <TaskItem>
                                                                            <div className="text-muted-foreground text-xs">
                                                                                Running with: {Object.keys(toolInvocation.args).join(', ')}
                                                                            </div>
                                                                        </TaskItem>
                                                                    )}
                                                                </TaskContent>
                                                            </Task>
                                                        )
                                                    })}
                                                </MessageContent>
                                            </Message>
                                        )}

                                        {/* Render text and reasoning in a separate message */}
                                        {nonToolParts.length > 0 && (
                                            <Message from={role}>
                                                <MessageAvatar role={role} />
                                                <MessageContent>
                                                    {nonToolParts.map((part, index) => {
                                                        switch (part.type) {
                                                            case "text":
                                                                return (
                                                                    <div key={index}>
                                                                        <Response>
                                                                            {part.text}
                                                                        </Response>
                                                                    </div>
                                                                )

                                                            case "reasoning":
                                                                return (
                                                                    <Reasoning
                                                                        key={index}
                                                                        isStreaming={status === "streaming"}
                                                                        defaultOpen={false}
                                                                    >
                                                                        <ReasoningTrigger />
                                                                        <ReasoningContent>
                                                                            {part.reasoning || ""}
                                                                        </ReasoningContent>
                                                                    </Reasoning>
                                                                )

                                                            default:
                                                                return null
                                                        }
                                                    })}
                                                    {/* Action buttons at the bottom right */}
                                                    <div className="flex items-center justify-end gap-2 mt-3 pt-2 border-t border-border/30">
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => reload()}
                                                            className="h-8 gap-1.5"
                                                            title="Regenerate response"
                                                        >
                                                            <RotateCcw className="h-3.5 w-3.5" />
                                                            <span className="text-xs">Regenerate</span>
                                                        </Button>
                                                        {nonToolParts.some(p => p.type === "text" && p.text) && (
                                                            <Button
                                                                variant="ghost"
                                                                size="sm"
                                                                onClick={() => {
                                                                    const textPart = nonToolParts.find(p => p.type === "text" && p.text)
                                                                    if (textPart?.type === "text") {
                                                                        handleCopyResponse(textPart.text)
                                                                    }
                                                                }}
                                                                className="h-8 gap-1.5"
                                                                title="Copy response"
                                                            >
                                                                <Copy className="h-3.5 w-3.5" />
                                                                <span className="text-xs">Copy</span>
                                                            </Button>
                                                        )}
                                                    </div>
                                                </MessageContent>
                                            </Message>
                                        )}
                                    </React.Fragment>
                                )
                            }

                            return null
                        })}

                        {/* Error message */}
                        {status === "error" && (
                            <Message from="assistant">
                                <MessageAvatar role="assistant" />
                                <MessageContent>
                                    <div className="text-sm text-destructive">
                                        <p className="font-semibold">Error while processing your request</p>
                                        <p>Error : {error?.message}</p>
                                        <p>Cause : {JSON.stringify(error)}</p>
                                    </div>
                                </MessageContent>
                            </Message>
                        )}

                        {/* Loading indicator */}
                        {status === "submitted" || status === "streaming" || status !== "ready" && (
                            <Message from="assistant">
                                <MessageAvatar role="assistant" />
                                <MessageContent>
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <div className="flex space-x-1">
                                            <div className="w-2 h-2 bg-current rounded-full animate-bounce" />
                                            <div className="w-2 h-2 bg-current rounded-full animate-bounce [animation-delay:0.2s]" />
                                            <div className="w-2 h-2 bg-current rounded-full animate-bounce [animation-delay:0.4s]" />
                                        </div>
                                        <span className="text-sm">Thinking...</span>
                                    </div>
                                </MessageContent>
                            </Message>
                        )}
                    </ConversationContent>
                    <ConversationScrollButton />
                </Conversation>

                <div className="p-6 pt-4 border-t bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
                    <PromptInput onSubmit={handleSubmitWithFiles} className="shadow-lg">
                        <MentionTagsInput
                            value={input}
                            onChange={setInput}
                            placeholder={hasSelectedApiKey ? "Ask me anything... (Type @ to mention notes)" : "Set up API keys in settings to start chatting..."}
                            disabled={status === "streaming" || status === "submitted" || !hasSelectedApiKey}
                            allNotes={allNotes}
                            notesLoading={notebooksLoading}
                            currentNoteId={noteId}
                            onMentionedNotesChange={handleMentionedNotesChange}
                            onKeyDown={(e) => {
                                // Submit on Enter (without Shift)
                                if (e.key === 'Enter' && !e.shiftKey) {
                                    e.preventDefault();
                                    if (input.trim() && hasSelectedApiKey && status !== "streaming" && status !== "submitted") {
                                        handleSubmitWithFiles(e as any);
                                    }
                                }
                            }}
                        />
                        <PromptInputToolbar>
                            <PromptInputTools>
                                <PromptInputModelSelect value={model} onValueChange={(value) => setModel(value as Model)}>
                                    <PromptInputModelSelectTrigger className="w-auto">
                                        <PromptInputModelSelectValue />
                                    </PromptInputModelSelectTrigger>
                                    <PromptInputModelSelectContent>
                                        <SelectGroup>
                                            <SelectLabel>OpenAI</SelectLabel>
                                            {Object.entries(modelToProvider)
                                                .filter(([, provider]) => provider === "openai")
                                                .map(([modelName]) => (
                                                    <PromptInputModelSelectItem key={modelName} value={modelName}>
                                                        {modelName} (openai)
                                                    </PromptInputModelSelectItem>
                                                ))}
                                        </SelectGroup>
                                        <SelectSeparator />
                                        <SelectGroup>
                                            <SelectLabel>Anthropic</SelectLabel>
                                            {Object.entries(modelToProvider)
                                                .filter(([, provider]) => provider === "anthropic")
                                                .map(([modelName]) => (
                                                    <PromptInputModelSelectItem key={modelName} value={modelName}>
                                                        {modelName} (anthropic)
                                                    </PromptInputModelSelectItem>
                                                ))}
                                        </SelectGroup>
                                        <SelectSeparator />
                                        <SelectGroup>
                                            <SelectLabel>Google</SelectLabel>
                                            {Object.entries(modelToProvider)
                                                .filter(([, provider]) => provider === "google")
                                                .map(([modelName]) => (
                                                    <PromptInputModelSelectItem key={modelName} value={modelName}>
                                                        {modelName} (google)
                                                    </PromptInputModelSelectItem>
                                                ))}
                                        </SelectGroup>
                                    </PromptInputModelSelectContent>
                                </PromptInputModelSelect>

                                <Input
                                    id="file-upload"
                                    type="file"
                                    multiple
                                    onChange={(event) => setFiles(event.target.files || null)}
                                    className="hidden"
                                />
                            </PromptInputTools>
                            {/* Submit/Stop button */}
                            {status === "streaming" || status === "submitted" ? (
                                <Button
                                    type="button"
                                    variant="default"
                                    size="icon"
                                    onClick={() => stop()}
                                    className="gap-1.5 rounded-lg"
                                    title="Stop generating"
                                >
                                    <SquareIcon className="size-4" />
                                </Button>
                            ) : (
                                <PromptInputSubmit
                                    status={status}
                                    disabled={!input.trim() || !hasSelectedApiKey}
                                />
                            )}
                        </PromptInputToolbar>
                    </PromptInput>
                </div>
            </RightSidebarContentWrapper>
            <RightSidebarRail />
        </RightSidebar>
    )
}
