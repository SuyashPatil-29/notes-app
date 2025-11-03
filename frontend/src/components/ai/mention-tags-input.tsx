import { cn } from "@/lib/utils";
import React, { useState, useRef, useEffect } from "react";
import { MentionNotesPopover } from "./mention-notes";

interface MentionedNote {
    id: string;
    name: string;
    content: string;
    chapterName: string;
    notebookName: string;
}

// Special marker format for mentions in text: @[noteId|noteName]
const MENTION_REGEX = /@\[([^\|]+)\|([^\]]+)\]/g;

interface MentionTagsInputProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
    disabled?: boolean;
    allNotes: MentionedNote[];
    notesLoading?: boolean;
    currentNoteId?: string | null;
    onMentionedNotesChange?: (notes: MentionedNote[]) => void;
    onKeyDown?: (e: React.KeyboardEvent<HTMLDivElement>) => void;
    className?: string;
}

export function MentionTagsInput({
    value,
    onChange,
    placeholder = "Ask me anything... (Type @ to mention notes)",
    disabled = false,
    allNotes,
    notesLoading = false,
    currentNoteId = null,
    onMentionedNotesChange,
    onKeyDown,
    className,
}: MentionTagsInputProps) {
    const [showMentionPopover, setShowMentionPopover] = useState(false);
    const editableRef = useRef<HTMLDivElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const isComposingRef = useRef(false);

    // Extract mentioned notes from value
    useEffect(() => {
        const matches = Array.from(value.matchAll(MENTION_REGEX));
        const noteIds = matches.map(m => m[1]);
        const notes = noteIds
            .map(id => allNotes.find(n => n.id === id))
            .filter((n): n is MentionedNote => n !== undefined);

        onMentionedNotesChange?.(notes);
    }, [value, allNotes, onMentionedNotesChange]);

    // Save cursor position by counting characters consistently
    const saveCursorPosition = () => {
        const selection = window.getSelection();
        if (!selection || selection.rangeCount === 0 || !editableRef.current) return null;

        const range = selection.getRangeAt(0);
        let offset = 0;
        let found = false;

        const traverseToRange = (node: Node): boolean => {
            if (found) return true;

            if (node === range.startContainer) {
                offset += range.startOffset;
                found = true;
                return true;
            }

            if (node.nodeType === Node.TEXT_NODE) {
                const textContent = node.textContent || '';
                // Filter out zero-width spaces for counting
                const visibleLength = textContent.replace(/\u200B/g, '').length;
                offset += visibleLength;
            } else if (node.nodeType === Node.ELEMENT_NODE) {
                const el = node as HTMLElement;
                if (el.contentEditable === 'false' && el.hasAttribute('data-mention-id')) {
                    // Count badge as 1 character
                    offset += 1;
                } else {
                    for (const child of Array.from(node.childNodes)) {
                        if (traverseToRange(child)) return true;
                    }
                }
            }
            return false;
        };

        traverseToRange(editableRef.current);
        return offset;
    };

    // Restore cursor position
    const restoreCursorPosition = (targetOffset: number) => {
        if (!editableRef.current) return;

        const selection = window.getSelection();
        if (!selection) return;

        const range = document.createRange();
        let currentOffset = 0;
        let found = false;

        const traverseNodes = (node: Node): boolean => {
            if (found) return true;

            if (node.nodeType === Node.TEXT_NODE) {
                const textContent = node.textContent || '';
                const visibleText = textContent.replace(/\u200B/g, '');
                const visibleLength = visibleText.length;

                if (currentOffset + visibleLength >= targetOffset) {
                    const nodeOffset = targetOffset - currentOffset;
                    // Account for zero-width spaces when setting actual position
                    let actualOffset = 0;
                    let visibleCount = 0;
                    for (let i = 0; i < textContent.length && visibleCount < nodeOffset; i++) {
                        if (textContent[i] !== '\u200B') {
                            visibleCount++;
                        }
                        actualOffset = i + 1;
                    }

                    range.setStart(node, Math.min(actualOffset, textContent.length));
                    range.setEnd(node, Math.min(actualOffset, textContent.length));
                    found = true;
                    return true;
                }
                currentOffset += visibleLength;
            } else if (node.nodeType === Node.ELEMENT_NODE) {
                const el = node as HTMLElement;
                if (el.contentEditable === 'false' && el.hasAttribute('data-mention-id')) {
                    if (currentOffset === targetOffset) {
                        // Cursor is right before the badge
                        range.setStartBefore(node);
                        range.setEndBefore(node);
                        found = true;
                        return true;
                    }
                    currentOffset += 1; // Count badge as 1 character
                } else {
                    for (const child of Array.from(node.childNodes)) {
                        if (traverseNodes(child)) return true;
                    }
                }
            }
            return false;
        };

        traverseNodes(editableRef.current);

        if (found) {
            selection.removeAllRanges();
            selection.addRange(range);
        } else {
            // If not found, place cursor at the end
            range.selectNodeContents(editableRef.current);
            range.collapse(false);
            selection.removeAllRanges();
            selection.addRange(range);
        }
    };

    // Render content with inline badges
    useEffect(() => {
        if (!editableRef.current || isComposingRef.current) return;

        const editable = editableRef.current;
        const wasActive = document.activeElement === editable;
        const cursorOffset = wasActive ? saveCursorPosition() : null;

        // Clear current content
        editable.innerHTML = '';

        if (!value) {
            return;
        }

        let lastIndex = 0;
        const matches = Array.from(value.matchAll(MENTION_REGEX));

        matches.forEach((match) => {
            const [fullMatch, noteId, noteName] = match;
            const matchIndex = match.index!;

            // Add text before mention
            if (matchIndex > lastIndex) {
                const textNode = document.createTextNode(value.slice(lastIndex, matchIndex));
                editable.appendChild(textNode);
            }

            // Add mention badge
            const badge = document.createElement('span');
            badge.className = 'inline-flex items-center gap-1 px-2 py-0.5 mx-0.5 rounded-md bg-primary/10 text-primary text-sm font-medium whitespace-nowrap';
            badge.contentEditable = 'false';
            badge.setAttribute('data-mention-id', noteId);
            badge.style.userSelect = 'none';

            const icon = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
            icon.setAttribute('class', 'w-3 h-3');
            icon.setAttribute('fill', 'none');
            icon.setAttribute('stroke', 'currentColor');
            icon.setAttribute('viewBox', '0 0 24 24');
            const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
            path.setAttribute('stroke-linecap', 'round');
            path.setAttribute('stroke-linejoin', 'round');
            path.setAttribute('stroke-width', '2');
            path.setAttribute('d', 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z');
            icon.appendChild(path);
            badge.appendChild(icon);

            const nameSpan = document.createElement('span');
            nameSpan.textContent = noteName;
            badge.appendChild(nameSpan);

            const removeBtn = document.createElement('button');
            removeBtn.type = 'button';
            removeBtn.className = 'ml-1 opacity-70 hover:opacity-100 focus:outline-none';
            const xIcon = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
            xIcon.setAttribute('class', 'w-3 h-3');
            xIcon.setAttribute('fill', 'none');
            xIcon.setAttribute('stroke', 'currentColor');
            xIcon.setAttribute('viewBox', '0 0 24 24');
            const xPath = document.createElementNS('http://www.w3.org/2000/svg', 'path');
            xPath.setAttribute('stroke-linecap', 'round');
            xPath.setAttribute('stroke-linejoin', 'round');
            xPath.setAttribute('stroke-width', '2');
            xPath.setAttribute('d', 'M6 18L18 6M6 6l12 12');
            xIcon.appendChild(xPath);
            removeBtn.appendChild(xIcon);
            removeBtn.onclick = (e) => {
                e.preventDefault();
                e.stopPropagation();
                handleRemoveMention(noteId);
            };
            badge.appendChild(removeBtn);

            editable.appendChild(badge);

            // Add zero-width space after badge for cursor positioning
            editable.appendChild(document.createTextNode('\u200B'));

            lastIndex = matchIndex + fullMatch.length;
        });

        // Add remaining text
        if (lastIndex < value.length) {
            const textNode = document.createTextNode(value.slice(lastIndex));
            editable.appendChild(textNode);
        }

        // Restore cursor position only if the element was active
        if (wasActive && cursorOffset !== null) {
            requestAnimationFrame(() => {
                restoreCursorPosition(cursorOffset);
            });
        }
    }, [value]);

    const getPlainText = (element: HTMLElement): string => {
        let text = '';
        for (const node of Array.from(element.childNodes)) {
            if (node.nodeType === Node.TEXT_NODE) {
                text += node.textContent;
            } else if (node.nodeType === Node.ELEMENT_NODE) {
                const el = node as HTMLElement;
                if (el.contentEditable === 'false' && el.hasAttribute('data-mention-id')) {
                    // This is a mention badge - find its original marker in value
                    const noteId = el.getAttribute('data-mention-id');
                    const match = value.match(new RegExp(`@\\[${noteId}\\|([^\\]]+)\\]`));
                    if (match) {
                        text += match[0];
                    }
                } else {
                    text += getPlainText(el);
                }
            }
        }
        return text.replace(/\u200B/g, ''); // Remove zero-width spaces
    };

    const handleInput = () => {
        if (isComposingRef.current || !editableRef.current) return;

        const text = getPlainText(editableRef.current);

        // Check if @ was just typed at the end
        if (text.endsWith('@')) {
            setShowMentionPopover(true);
        } else if (showMentionPopover && !text.includes('@')) {
            setShowMentionPopover(false);
        }

        onChange(text);
    };

    const handleSelectNote = (note: MentionedNote) => {
        // Replace the trailing @ with mention marker
        const newValue = value.slice(0, -1) + `@[${note.id}|${note.name}] `;
        onChange(newValue);
        setShowMentionPopover(false);

        // Refocus editable div and move cursor to end
        setTimeout(() => {
            if (editableRef.current) {
                editableRef.current.focus();

                // Move cursor to the very end
                const selection = window.getSelection();
                const range = document.createRange();

                // Select all content
                range.selectNodeContents(editableRef.current);
                // Collapse to the end
                range.collapse(false);

                selection?.removeAllRanges();
                selection?.addRange(range);
            }
        }, 100);
    };

    const handleRemoveMention = (noteId: string) => {
        const regex = new RegExp(`@\\[${noteId}\\|[^\\]]+\\]\\s?`, 'g');
        const newValue = value.replace(regex, '');
        onChange(newValue);
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
        // Close mention popover on Escape
        if (e.key === 'Escape' && showMentionPopover) {
            e.preventDefault();
            setShowMentionPopover(false);
            editableRef.current?.focus();
            return;
        }

        onKeyDown?.(e);
    };

    const handleCompositionStart = () => {
        isComposingRef.current = true;
    };

    const handleCompositionEnd = () => {
        isComposingRef.current = false;
        handleInput();
    };

    return (
        <div className={cn("relative", className)}>
            {/* Mention popover */}
            <div className="relative" ref={containerRef}>
                <MentionNotesPopover
                    open={showMentionPopover}
                    notes={allNotes}
                    onSelectNote={handleSelectNote}
                    isLoading={notesLoading}
                    currentNoteId={currentNoteId}
                />

                {/* Contenteditable div with inline badges */}
                <div
                    ref={editableRef}
                    contentEditable={!disabled}
                    onInput={handleInput}
                    onKeyDown={handleKeyDown}
                    onCompositionStart={handleCompositionStart}
                    onCompositionEnd={handleCompositionEnd}
                    data-placeholder={!value ? placeholder : undefined}
                    className={cn(
                        'w-full rounded-none border-none p-3 shadow-none outline-none ring-0',
                        'bg-transparent dark:bg-transparent',
                        'focus-visible:ring-0',
                        !value && 'empty:before:content-[attr(data-placeholder)] empty:before:text-muted-foreground',
                        disabled && 'opacity-50 cursor-not-allowed',
                        'overflow-y-auto whitespace-pre-wrap wrap-break-word'
                    )}
                    style={{
                        minHeight: '120px',
                        maxHeight: '200px',
                    }}
                />
            </div>
        </div>
    );
}

