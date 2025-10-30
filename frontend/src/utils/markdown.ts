// Markdown conversion utilities for ProseMirror content

export interface ProseMirrorNode {
  type: string;
  content?: ProseMirrorNode[];
  text?: string;
  attrs?: Record<string, any>;
  marks?: ProseMirrorMark[];
}

export interface ProseMirrorMark {
  type: string;
  attrs?: Record<string, any>;
}

export interface ProseMirrorDoc {
  type: string;
  content: ProseMirrorNode[];
}

// Detect if content is ProseMirror JSON or plain markdown
export const isJsonContent = (content: string): boolean => {
  if (!content || typeof content !== 'string') return false;
  try {
    const parsed = JSON.parse(content);
    return parsed && typeof parsed === 'object' && parsed.type === 'doc';
  } catch {
    return false;
  }
};

// Convert ProseMirror JSON to markdown
export const proseMirrorToMarkdown = (content: string): string => {
  if (!content) return '';

  // If it's plain markdown, return as-is
  if (!isJsonContent(content)) {
    return content;
  }

  try {
    const doc: ProseMirrorDoc = JSON.parse(content);
    return convertNodesToMarkdown(doc.content || []);
  } catch (error) {
    console.error('Error parsing ProseMirror content:', error);
    return content; // Fallback to original content
  }
};

// Convert ProseMirror nodes to markdown
const convertNodesToMarkdown = (nodes: ProseMirrorNode[], indent: string = ''): string => {
  return nodes.map(node => convertNodeToMarkdown(node, indent)).join('');
};

// Convert a single ProseMirror node to markdown
const convertNodeToMarkdown = (node: ProseMirrorNode, indent: string = ''): string => {
  const { type, content, text, attrs } = node;

  switch (type) {
    case 'heading':
      const level = attrs?.level || 1;
      const headingMarks = node.content ? convertInlineMarks(node.content[0]?.text || '', node.content[0]?.marks || []) : '';
      return `${'#'.repeat(level)} ${headingMarks}\n\n`;

    case 'paragraph':
      if (!content || content.length === 0) return '\n';
      const paragraphText = convertInlineContent(content);
      return `${indent}${paragraphText}\n\n`;

    case 'bulletList':
      return convertListItems(content || [], indent, '- ');

    case 'orderedList':
      return convertOrderedListItems(content || [], indent, 1);

    case 'listItem':
      // This should be handled by the list conversion above
      return convertNodesToMarkdown(content || [], indent);

    case 'codeBlock':
      const language = attrs?.language || '';
      const codeContent = text || '';
      return `\`\`\`${language}\n${codeContent}\n\`\`\`\n\n`;

    case 'blockquote':
      const quoteContent = convertNodesToMarkdown(content || [], '> ');
      return quoteContent + '\n';

    case 'horizontalRule':
      return '---\n\n';

    case 'hardBreak':
      return '\n';

    case 'text':
      return convertInlineMarks(text || '', node.marks || []);

    default:
      // Handle unknown node types by trying to extract text content
      if (content && content.length > 0) {
        return convertNodesToMarkdown(content, indent);
      }
      return '';
  }
};

// Convert inline content (text with marks)
const convertInlineContent = (content: ProseMirrorNode[]): string => {
  return content.map(node => {
    if (node.type === 'text') {
      return convertInlineMarks(node.text || '', node.marks || []);
    }
    // Handle inline nodes like links, etc.
    return convertInlineNode(node);
  }).join('');
};

// Convert inline marks (bold, italic, code, links, etc.)
const convertInlineMarks = (text: string, marks: ProseMirrorMark[]): string => {
  let result = text;

  // Apply marks in reverse order (innermost first)
  const sortedMarks = [...marks].reverse();

  for (const mark of sortedMarks) {
    switch (mark.type) {
      case 'bold':
        result = `**${result}**`;
        break;
      case 'italic':
        result = `*${result}*`;
        break;
      case 'code':
        result = `\`${result}\``;
        break;
      case 'strike':
        result = `~~${result}~~`;
        break;
      case 'link':
        const href = mark.attrs?.href || '';
        result = `[${result}](${href})`;
        break;
      // Add more mark types as needed
      default:
        break;
    }
  }

  return result;
};

// Convert inline nodes (like links that are nodes rather than marks)
const convertInlineNode = (node: ProseMirrorNode): string => {
  switch (node.type) {
    case 'link':
      const href = node.attrs?.href || '';
      const linkText = node.content ? convertInlineContent(node.content) : '';
      return `[${linkText}](${href})`;
    default:
      return node.content ? convertInlineContent(node.content) : '';
  }
};

// Convert bullet list items
const convertListItems = (items: ProseMirrorNode[], indent: string, bullet: string): string => {
  return items.map(item => {
    const itemContent = convertNodesToMarkdown(item.content || [], '');
    const lines = itemContent.trim().split('\n');
    const firstLine = lines[0] ? `${indent}${bullet}${lines[0]}` : `${indent}${bullet}`;
    const restLines = lines.slice(1).map(line => `${indent}  ${line}`).join('\n');
    return firstLine + (restLines ? '\n' + restLines : '') + '\n';
  }).join('') + '\n';
};

// Convert ordered list items
const convertOrderedListItems = (items: ProseMirrorNode[], indent: string, startNum: number): string => {
  let result = '';
  let num = startNum;

  for (const item of items) {
    const itemContent = convertNodesToMarkdown(item.content || [], '');
    const lines = itemContent.trim().split('\n');
    const firstLine = lines[0] ? `${indent}${num}. ${lines[0]}` : `${indent}${num}.`;
    const restLines = lines.slice(1).map(line => `${indent}   ${line}`).join('\n');
    result += firstLine + (restLines ? '\n' + restLines : '') + '\n';
    num++;
  }

  return result + '\n';
};

// Get markdown content from any note content (handles both JSON and plain markdown)
export const getMarkdownContent = (content: string): string => {
  return proseMirrorToMarkdown(content);
};

// Check if content contains markdown formatting
export const hasMarkdownContent = (content: string): boolean => {
  if (!content) return false;

  // Check for common markdown patterns
  const markdownPatterns = [
    /^#{1,6}\s/m,  // Headers
    /\*\*.*\*\*/,  // Bold
    /\*.*\*/,      // Italic
    /`.*`/,        // Inline code
    /```[\s\S]*?```/, // Code blocks
    /\[.*\]\(.*\)/, // Links
    /^[-*+]\s/m,   // Unordered lists
    /^\d+\.\s/m,   // Ordered lists
  ];

  return markdownPatterns.some(pattern => pattern.test(content));
};

// Extract plain text from ProseMirror nodes (for previews)
const extractTextFromNodes = (nodes: ProseMirrorNode[]): string => {
  return nodes.map(node => extractTextFromNode(node)).join(' ');
};

// Extract plain text from a single ProseMirror node
const extractTextFromNode = (node: ProseMirrorNode): string => {
  const { type, content, text } = node;

  switch (type) {
    case 'text':
      return text || '';
    case 'heading':
    case 'paragraph':
    case 'listItem':
    case 'blockquote':
      return content ? extractTextFromNodes(content) : '';
    case 'bulletList':
    case 'orderedList':
      return content ? extractTextFromNodes(content) : '';
    case 'codeBlock':
      return text || '';
    case 'hardBreak':
      return ' ';
    default:
      return content ? extractTextFromNodes(content) : '';
  }
};

// Get preview text (plain text without formatting) from content
export const getPreviewText = (content: string, maxLength: number = 150): string => {
  if (!content) return 'Empty note';

  // If it's JSON content, extract plain text
  if (isJsonContent(content)) {
    try {
      const doc: ProseMirrorDoc = JSON.parse(content);
      const plainText = extractTextFromNodes(doc.content || []);
      const trimmed = plainText.trim().replace(/\s+/g, ' ');
      if (!trimmed) return 'Empty note';
      return trimmed.length > maxLength 
        ? trimmed.substring(0, maxLength) + '...'
        : trimmed;
    } catch (error) {
      console.error('Error extracting text from JSON:', error);
      return 'Empty note';
    }
  }

  // For markdown, strip formatting for preview
  const plainText = content
    .replace(/```[\s\S]*?```/g, '') // Remove code blocks
    .replace(/`[^`]+`/g, '') // Remove inline code
    .replace(/#{1,6}\s/g, '') // Remove headers
    .replace(/\*\*([^*]+)\*\*/g, '$1') // Remove bold
    .replace(/\*([^*]+)\*/g, '$1') // Remove italic
    .replace(/~~([^~]+)~~/g, '$1') // Remove strikethrough
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1') // Remove links
    .replace(/^[-*+]\s/gm, '') // Remove list bullets
    .replace(/^\d+\.\s/gm, '') // Remove ordered list numbers
    .replace(/\s+/g, ' ') // Normalize whitespace
    .trim();

  if (!plainText) return 'Empty note';
  
  return plainText.length > maxLength 
    ? plainText.substring(0, maxLength) + '...'
    : plainText;
};
