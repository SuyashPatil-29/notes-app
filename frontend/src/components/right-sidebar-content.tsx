import { MessageSquare } from "lucide-react"
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

const modelToProvider = {
    "gpt-4o-mini": "openai",
    "gpt-4o": "openai",
    "claude-3-7-sonnet-latest": "anthropic",
    "o1": "openai",
    "gemini-2.5-pro-preview-03-25": "google",
} as const

type Model = keyof typeof modelToProvider

export function RightSidebarContent() {
    const [model, setModel] = useState<Model>("gpt-4o-mini")
    const [files, setFiles] = useState<FileList | null>(null)
    const { user } = useUser()
    const { open } = useRightSidebar()
    const inputRef = useRef<HTMLTextAreaElement>(null)

    const { messages, input, handleInputChange, handleSubmit, status, error } = useChat({
        api: "http://localhost:8080/api/chat",
        body: {
            provider: modelToProvider[model],
            model,
        },
        credentials: "include",
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
                {error && (
                    <div className="mx-4 mt-4 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                        Error: {error.message}
                    </div>
                )}

                <Conversation className="flex-1">
                    <ConversationContent className="px-6">
                        {messages.length === 0 && (
                            <div className="flex flex-col items-center justify-center h-full text-center space-y-3">
                                <div className="rounded-full bg-muted p-4">
                                    <MessageSquare className="size-8 text-muted-foreground" />
                                </div>
                                <div>
                                    <h3 className="font-semibold text-lg">Start a conversation</h3>
                                    <p className="text-sm text-muted-foreground mt-1">
                                        Ask me anything or try using tools
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
                            placeholder="Ask me anything..."
                            onChange={handleInputChange}
                            disabled={status === "streaming"}
                        />
                        <PromptInputToolbar>
                            <PromptInputTools>
                                <PromptInputModelSelect value={model} onValueChange={(value) => setModel(value as Model)}>
                                    <PromptInputModelSelectTrigger className="w-auto">
                                        <PromptInputModelSelectValue />
                                    </PromptInputModelSelectTrigger>
                                    <PromptInputModelSelectContent>
                                        {Object.entries(modelToProvider).map(([modelName, provider]) => (
                                            <PromptInputModelSelectItem key={modelName} value={modelName}>
                                                {modelName} ({provider})
                                            </PromptInputModelSelectItem>
                                        ))}
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
                            <PromptInputSubmit status={status} disabled={!input.trim() || status === "streaming"} />
                        </PromptInputToolbar>
                    </PromptInput>
                </div>
            </RightSidebarContentWrapper>
            <RightSidebarRail />
        </RightSidebar>
    )
}
