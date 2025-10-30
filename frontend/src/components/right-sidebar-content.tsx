import { FileText, MessageSquare } from "lucide-react"
import { toast } from "sonner"
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
    PromptInputTextarea,
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
import { Tool, ToolHeader, ToolContent, ToolInput, ToolOutput } from "@/components/ai/tool"
import { Input } from "@/components/ui/input"
import { useUser } from "@/hooks/auth"
import api from "@/utils/api"
import { SelectGroup, SelectLabel, SelectSeparator } from "@/components/ui/select"
import { MentionNotesPopover, MentionedNotesBadges } from "@/components/ai/mention-notes"
import { getUserNotebooks } from "@/utils/notebook"
import type { Notebook } from "@/types/backend"

// Helper function to render tool output with nice formatting
function renderToolOutput(toolName: string, result: any) {
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
                                    <div key={idx} className="p-2 bg-muted/50 rounded border">
                                        <p className="font-medium text-foreground">{note.name}</p>
                                        <p className="text-xs text-muted-foreground">
                                            {note.notebookName} → {note.chapterName}
                                        </p>
                                        <p className="mt-1 text-xs text-muted-foreground">{note.preview}</p>
                                        <p className="mt-1 text-xs text-muted-foreground">Updated: {note.updatedAt}</p>
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
                    <div>
                        <p className="font-semibold text-foreground">{result.name}</p>
                        <p className="text-xs text-muted-foreground">
                            {result.notebookName} → {result.chapterName}
                        </p>
                        <p className="text-xs text-muted-foreground">
                            Created: {result.createdAt} • Updated: {result.updatedAt}
                        </p>
                    </div>
                    <div className="p-3 bg-muted/50 rounded border max-h-96 overflow-y-auto">
                        <pre className="text-xs whitespace-pre-wrap font-mono">{result.content}</pre>
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
                                    <p className="text-xs text-muted-foreground mt-1">{note.preview}</p>
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
                                <p className="font-semibold text-green-600 dark:text-green-400">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="font-medium text-foreground">{result.noteName}</p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    {result.notebookName} → {result.chapterName}
                                </p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    Size: {result.contentSize} characters • Created: {result.createdAt}
                                </p>
                                <p className="text-xs text-muted-foreground mt-2">
                                    Note ID: <code className="px-1 py-0.5 bg-muted rounded text-xs">{result.noteId}</code>
                                </p>
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
                            <div className="p-3 bg-muted/50 rounded border">
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
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="text-xs text-muted-foreground">
                                    <span className="line-through">{result.oldName}</span> → <span className="font-medium text-foreground">{result.newName}</span>
                                </p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    {result.notebookName} → {result.chapterName}
                                </p>
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
                                <p className="font-semibold text-green-600 dark:text-green-400">
                                    ✓ {result.message}
                                </p>
                            </div>
                            <div className="p-3 bg-muted/50 rounded border">
                                <p className="font-medium text-foreground">{result.noteName}</p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    {result.notebookName} → {result.chapterName}
                                </p>
                                <p className="text-xs text-muted-foreground mt-1">
                                    New size: {result.contentSize} characters • Updated: {result.updatedAt}
                                </p>
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
    const { user } = useUser()
    const [model, setModel] = useState<Model>("gpt-5-mini")
    const [files, setFiles] = useState<FileList | null>(null)
    const { open } = useRightSidebar()
    const inputRef = useRef<HTMLTextAreaElement>(null)
    const queryClient = useQueryClient()
    const processedNoteIds = useRef<Set<string>>(new Set())
    
    // Mention notes state
    const [mentionOpen, setMentionOpen] = useState(false)
    const [mentionedNotes, setMentionedNotes] = useState<MentionedNote[]>([])
    // Fetch notebooks to get all notes for mentions
    const { data: notebooks, isLoading: notebooksLoading } = useQuery<Notebook[]>({
        queryKey: ['userNotebooks'],
        queryFn: getUserNotebooks,
        enabled: open && !!user,
    })

    // Flatten all notes for mention selector
    const allNotes = useMemo(() => {
        if (!notebooks) return []
        return notebooks.flatMap(notebook =>
            notebook.chapters?.flatMap(chapter =>
                chapter.notes?.map(note => ({
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

    const { messages, input, handleInputChange, handleSubmit, status, setInput, append } = useChat({
        api: "http://localhost:8080/api/chat",
        body: {
            provider: modelToProvider[model],
            model,
        },
        credentials: "include",
        maxSteps: 5, // Enable multi-step tool calling
        onError: (error) => {
            console.error("Chat error:", error);
            // Show toast notification for errors
            const errorMessage = error.message || "An error occurred while processing your request";
            toast.error(errorMessage, {
                description: errorMessage.includes("API key")
                    ? "You can update your API key in Profile settings"
                    : undefined,
                duration: 5000,
            });
        },
    })

    // Handle @ mention detection
    const handleInputChangeWithMention = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        handleInputChange(e)
        
        const value = e.target.value
        const cursorPos = e.target.selectionStart
        
        // Check if @ was just typed
        if (value[cursorPos - 1] === '@') {
            setMentionOpen(true)
        } else if (mentionOpen && (value[cursorPos - 1] === ' ' || !value.includes('@'))) {
            // Close on space or if @ is removed
            setMentionOpen(false)
        }
    }

    // Handle keyboard events for mention popover
    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === 'Escape' && mentionOpen) {
            e.preventDefault()
            setMentionOpen(false)
            // Refocus textarea after closing
            setTimeout(() => {
                inputRef.current?.focus()
            }, 50)
        }
    }

    // Handle note selection from mention
    const handleSelectNote = (note: MentionedNote) => {
        // Add to mentioned notes if not already added
        if (!mentionedNotes.find(n => n.id === note.id)) {
            setMentionedNotes([...mentionedNotes, note])
        }
        
        // Remove the @ symbol from input
        const newInput = input.slice(0, -1) + ' '
        setInput(newInput)
        setMentionOpen(false)
        
        // Refocus textarea with a small delay
        setTimeout(() => {
            inputRef.current?.focus()
        }, 100)
    }

    // Remove mentioned note
    const handleRemoveMentionedNote = (noteId: string) => {
        setMentionedNotes(mentionedNotes.filter(n => n.id !== noteId))
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

    // Invalidate cache when notes are modified
    useEffect(() => {
        const lastMessage = messages[messages.length - 1]
        if (lastMessage?.role === 'assistant' && lastMessage.id) {
            lastMessage.parts?.forEach((part: any, partIndex: number) => {
                if (part.type === 'tool-invocation') {
                    const toolInvocation = part.toolInvocation as any
                    const toolName = toolInvocation.toolName
                    
                    // List of tools that modify notes/chapters and need cache invalidation
                    const modifyingTools = ['createNote', 'moveNote', 'moveChapter', 'renameNote', 'deleteNote', 'updateNoteContent']
                    
                    if (modifyingTools.includes(toolName) && toolInvocation.result?.success === true) {
                        // Use message ID + part index + tool name for truly unique operation key
                        const operationKey = `${lastMessage.id}-${partIndex}-${toolName}`
                        
                        // Only process if we haven't seen this exact tool invocation before
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
                            }

                            const messages = toastMessages[toolName] || {
                                loading: 'Processing...',
                                success: 'Changes saved!',
                                error: 'Failed to refresh'
                            }

                            // Use toast.promise for better UX with loading state
                            toast.promise(
                                Promise.all([
                                    queryClient.refetchQueries({ queryKey: ['userNotebooks'] }),
                                    queryClient.refetchQueries({ queryKey: ['notes'] }),
                                    // Also refetch individual note queries to update the note view
                                    queryClient.refetchQueries({ queryKey: ['note'] })
                                ]),
                                messages
                            )
                        }
                    }
                }
            })
        }
    }, [messages, queryClient])

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
            <RightSidebarContentWrapper className="flex flex-col h-full p-0">
                <Conversation className="flex-1">
                    <ConversationContent className="px-6">
                        {messages.length === 0 && (
                            <div className="flex flex-col items-center justify-center h-full text-center space-y-3">
                                <div className="rounded-full bg-muted p-4">
                                    <MessageSquare className="size-8 text-muted-foreground" />
                                </div>
                                <div>
                                    <h3 className="font-semibold text-lg">
                                        {hasSelectedApiKey ? "Start a conversation" : "Set up your API keys"}
                                    </h3>
                                    <p className="text-sm text-muted-foreground mt-1">
                                        {hasSelectedApiKey
                                            ? "Ask me about your notes, notebooks, or anything else!"
                                            : "Configure your API keys in settings to start chatting"
                                        }
                                    </p>
                                    {hasSelectedApiKey && (
                                        <p className="text-xs text-muted-foreground mt-2">
                                            Try: "What notes do I have?" or "Search for notes about..."
                                        </p>
                                    )}
                                </div>
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
                                                <Response>{mainMessage}</Response>
                                            </div>
                                        </MessageContent>
                                    </Message>
                                )
                            }

                            if (role === "assistant") {
                                return (
                                    <Message from={role} key={m.id}>
                                        <MessageAvatar role={role} />
                                        <MessageContent>
                                            {m.parts?.map((part, index) => {
                                                switch (part.type) {
                                                    case "text":
                                                        return (
                                                            <Response key={index}>
                                                                {part.text}
                                                            </Response>
                                                        )

                                                    case "tool-invocation":
                                                        const toolInvocation = part.toolInvocation as any
                                                        const toolState = toolInvocation.result
                                                            ? "output-available"
                                                            : "input-available"

                                                        return (
                                                            <Tool key={index} defaultOpen>
                                                                <ToolHeader
                                                                    type={toolInvocation.toolName}
                                                                    state={toolState}
                                                                />
                                                                <ToolContent>
                                                                    {toolInvocation.args && (
                                                                        <ToolInput input={toolInvocation.args} />
                                                                    )}
                                                                    {toolInvocation.result && (
                                                                        <ToolOutput
                                                                            output={renderToolOutput(toolInvocation.toolName, toolInvocation.result)}
                                                                            errorText={undefined}
                                                                        />
                                                                    )}
                                                                </ToolContent>
                                                            </Tool>
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
                                        </MessageContent>
                                    </Message>
                                )
                            }

                            return null
                        })}

                        {/* Loading indicator */}
                        {status === "submitted" && (
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
                        <MentionedNotesBadges
                            notes={mentionedNotes}
                            onRemove={handleRemoveMentionedNote}
                        />
                        <div className="relative overflow-visible">
                            <MentionNotesPopover
                                open={mentionOpen}
                                notes={allNotes}
                                onSelectNote={handleSelectNote}
                                isLoading={notebooksLoading}
                            />
                            <PromptInputTextarea
                                ref={inputRef}
                                value={input}
                                placeholder={hasSelectedApiKey ? "Ask me anything... (Type @ to mention notes)" : "Set up API keys in settings to start chatting..."}
                                onChange={handleInputChangeWithMention}
                                onKeyDown={handleKeyDown}
                                disabled={status === "streaming" || !hasSelectedApiKey}
                            />
                        </div>
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
                            <PromptInputSubmit
                                status={status}
                                disabled={!input.trim() || status === "streaming" || !hasSelectedApiKey}
                            />
                        </PromptInputToolbar>
                    </PromptInput>
                </div>
            </RightSidebarContentWrapper>
            <RightSidebarRail />
        </RightSidebar>
    )
}
