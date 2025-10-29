import { MessageSquare } from "lucide-react"
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
import React, { useState, useRef, useEffect } from "react"
import { useQuery } from "@tanstack/react-query"
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

export function RightSidebarContent() {
    const { user } = useUser()
    const [model, setModel] = useState<Model>("gpt-5-mini")
    const [files, setFiles] = useState<FileList | null>(null)
    const { open } = useRightSidebar()
    const inputRef = useRef<HTMLTextAreaElement>(null)
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

    const { messages, input, handleInputChange, handleSubmit, status } = useChat({
        api: "http://localhost:8080/api/chat",
        body: {
            provider: modelToProvider[model],
            model,
        },
        credentials: "include",
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

    const handleSubmitWithFiles = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        if (files) {
            handleSubmit(e, {
                experimental_attachments: files,
            })
            setFiles(null)
        } else {
            handleSubmit(e)
        }
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
                                            ? "Ask me anything or try using tools"
                                            : "Configure your API keys in settings to start chatting"
                                        }
                                    </p>
                                </div>
                            </div>
                        )}
                        {messages.map((m) => {
                            const role = m.role

                            // Skip data messages
                            if (role === "data") return null

                            if (role === "user") {
                                return (
                                    <Message from={role} key={m.id}>
                                        <MessageAvatar role={role} image={user?.imageUrl} />
                                        <MessageContent>
                                            <Response>{m.content}</Response>
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
                                                                            output={
                                                                                <pre className="p-2 text-xs">
                                                                                    {JSON.stringify(toolInvocation.result, null, 2)}
                                                                                </pre>
                                                                            }
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
                        <PromptInputTextarea
                            ref={inputRef}
                            value={input}
                            placeholder={hasSelectedApiKey ? "Ask me anything..." : "Set up API keys in settings to start chatting..."}
                            onChange={handleInputChange}
                            disabled={status === "streaming" || !hasSelectedApiKey}
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
