import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getNote } from '@/utils/notes'
import { getUserNotebooks } from '@/utils/notebook'
import { Header } from '@/components/Header'
import type { AuthenticatedUser } from '@/types/backend'
import { Loader2, Calendar, Clock } from 'lucide-react'
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
import { useState } from "react";
import { useDebouncedCallback } from "use-debounce";
import { defaultExtensions } from "@/lib/extensions";
import { ColorSelector } from "./selectors/color-selector";
import { LinkSelector } from "./selectors/link-selector";
import { MathSelector } from "./selectors/math-selector";
import { NodeSelector } from "./selectors/node-selector";
import { Separator } from "./ui/separator";
import type { Notes } from "@/types/backend";

import GenerativeMenuSwitch from "./generative/generative-menu-switch";
// import { uploadFn } from "./image-upload";
import { TextButtons } from "./selectors/text-buttons";
import { slashCommand, suggestionItems } from "./slash-command";

import hljs from "highlight.js";

const extensions = [...defaultExtensions, slashCommand];

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
  const [charsCount, setCharsCount] = useState();

  const [openAI, setOpenAI] = useState(false);
  const [openNode, setOpenNode] = useState(false);
  const [openColor, setOpenColor] = useState(false);
  const [openLink, setOpenLink] = useState(false);

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

  const note : Notes = noteResponse?.data
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
    if (!note?.content) {
      return undefined;
    }
    
    // Try to parse as JSON first (new format)
    try {
      const parsed = JSON.parse(note.content);
      return parsed;
    } catch (error) {
      // If it's not JSON, treat it as markdown and convert
      return markdownToJSON(note.content);
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
      <div className="flex-1 overflow-auto">
        <div className="max-w-4xl mx-auto px-6 py-8">
          {/* Note Metadata Section */}
          <div className="mb-6 space-y-4">
            <div>
              <h1 className="text-4xl font-bold tracking-tight">{note.name}</h1>
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
              <div className="ml-auto flex items-center gap-2">
                <span className="text-xs font-mono bg-accent px-2 py-1 rounded">Path: {note.chapter.notebook.name} / {note.chapter.name} / {note.name}</span>
              </div>
            </div>
            
            <Separator />
          </div>

          <div className="relative w-full max-w-5xl">
            <div className="flex absolute right-5 top-5 z-10 mb-5 gap-2">
              <div className={charsCount ? "rounded-lg bg-accent px-2 py-1 text-sm text-muted-foreground" : "hidden"}>
                {charsCount} Words
              </div>
            </div>
            <EditorRoot key={`${noteId}-${note.updatedAt}`}>
              <EditorContent
                initialContent={getInitialContent()}
                extensions={extensions}
                className="relative min-h-[500px] w-full max-w-5xl border border-border/40 rounded-lg transition-colors hover:border-border/60 focus-within:border-border"
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
                  onUpdate={({ editor }) => {
                    debouncedUpdates(editor);
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
          </div>
        </div>
      </div>
    </div>
  )
}

