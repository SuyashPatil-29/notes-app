package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TipTapToMarkdown converts TipTap JSON to markdown string
func TipTapToMarkdown(tiptapJSON string) (string, error) {
	if tiptapJSON == "" {
		return "", nil
	}

	// Try to parse as TipTap JSON
	var doc TipTapDoc
	if err := json.Unmarshal([]byte(tiptapJSON), &doc); err != nil {
		// If not valid JSON, assume it's already plain text/markdown
		return tiptapJSON, nil
	}

	// If not a valid TipTap document, return as-is
	if doc.Type != "doc" {
		return tiptapJSON, nil
	}

	// Convert TipTap nodes to markdown
	return nodesToMarkdown(doc.Content, 0), nil
}

func nodesToMarkdown(nodes []TipTapNode, depth int) string {
	var result strings.Builder

	for i, node := range nodes {
		markdown := nodeToMarkdown(node, depth)
		result.WriteString(markdown)

		// Add spacing between certain block elements
		if i < len(nodes)-1 {
			nextNode := nodes[i+1]
			if shouldAddSpacing(node.Type, nextNode.Type) {
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}

func nodeToMarkdown(node TipTapNode, depth int) string {
	switch node.Type {
	case "paragraph":
		return paragraphToMarkdown(node) + "\n"

	case "heading":
		return headingToMarkdown(node) + "\n"

	case "codeBlock":
		return codeBlockToMarkdown(node) + "\n"

	case "bulletList":
		return bulletListToMarkdown(node, depth)

	case "orderedList":
		return orderedListToMarkdown(node, depth)

	case "listItem":
		return listItemToMarkdown(node, depth)

	case "horizontalRule":
		return "---\n"

	case "blockquote":
		return blockquoteToMarkdown(node, depth)

	case "text":
		return textToMarkdown(node)

	case "hardBreak":
		return "  \n"

	default:
		// For unknown types, try to extract text content
		if len(node.Content) > 0 {
			return nodesToMarkdown(node.Content, depth)
		}
		return ""
	}
}

func paragraphToMarkdown(node TipTapNode) string {
	if len(node.Content) == 0 {
		return ""
	}
	return nodesToMarkdown(node.Content, 0)
}

func headingToMarkdown(node TipTapNode) string {
	level := 1
	if node.Attrs != nil {
		if l, ok := node.Attrs["level"].(float64); ok {
			level = int(l)
		}
	}

	hashes := strings.Repeat("#", level)
	content := nodesToMarkdown(node.Content, 0)
	return fmt.Sprintf("%s %s", hashes, strings.TrimSpace(content))
}

func codeBlockToMarkdown(node TipTapNode) string {
	language := ""
	if node.Attrs != nil {
		if lang, ok := node.Attrs["language"].(string); ok && lang != "" {
			language = lang
		}
	}

	content := ""
	if len(node.Content) > 0 {
		content = nodesToMarkdown(node.Content, 0)
	}

	return fmt.Sprintf("```%s\n%s\n```", language, strings.TrimSpace(content))
}

func bulletListToMarkdown(node TipTapNode, depth int) string {
	var result strings.Builder
	for _, item := range node.Content {
		if item.Type == "listItem" {
			indent := strings.Repeat("  ", depth)
			itemContent := listItemContentToMarkdown(item, depth)
			result.WriteString(fmt.Sprintf("%s- %s", indent, itemContent))
		}
	}
	return result.String()
}

func orderedListToMarkdown(node TipTapNode, depth int) string {
	var result strings.Builder
	for i, item := range node.Content {
		if item.Type == "listItem" {
			indent := strings.Repeat("  ", depth)
			itemContent := listItemContentToMarkdown(item, depth)
			result.WriteString(fmt.Sprintf("%s%d. %s", indent, i+1, itemContent))
		}
	}
	return result.String()
}

func listItemToMarkdown(node TipTapNode, depth int) string {
	return listItemContentToMarkdown(node, depth)
}

func listItemContentToMarkdown(node TipTapNode, depth int) string {
	if len(node.Content) == 0 {
		return "\n"
	}

	var result strings.Builder
	for i, child := range node.Content {
		if child.Type == "paragraph" {
			// For first paragraph in list item, inline it
			if i == 0 {
				result.WriteString(nodesToMarkdown(child.Content, depth))
				result.WriteString("\n")
			} else {
				// Subsequent paragraphs are indented
				indent := strings.Repeat("  ", depth+1)
				result.WriteString(indent)
				result.WriteString(nodesToMarkdown(child.Content, depth))
				result.WriteString("\n")
			}
		} else if child.Type == "bulletList" || child.Type == "orderedList" {
			// Nested lists
			result.WriteString(nodeToMarkdown(child, depth+1))
		} else {
			result.WriteString(nodeToMarkdown(child, depth))
		}
	}

	return result.String()
}

func blockquoteToMarkdown(node TipTapNode, depth int) string {
	var result strings.Builder
	content := nodesToMarkdown(node.Content, depth)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	for _, line := range lines {
		result.WriteString("> ")
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

func textToMarkdown(node TipTapNode) string {
	text := node.Text

	if len(node.Marks) > 0 {
		for _, mark := range node.Marks {
			switch mark.Type {
			case "bold":
				text = fmt.Sprintf("**%s**", text)
			case "italic":
				text = fmt.Sprintf("*%s*", text)
			case "code":
				text = fmt.Sprintf("`%s`", text)
			case "strike":
				text = fmt.Sprintf("~~%s~~", text)
			case "link":
				if mark.Attrs != nil {
					if href, ok := mark.Attrs["href"].(string); ok {
						text = fmt.Sprintf("[%s](%s)", text, href)
					}
				}
			}
		}
	}

	return text
}

func shouldAddSpacing(currentType, nextType string) bool {
	// Add spacing between block-level elements
	blockElements := map[string]bool{
		"paragraph":      true,
		"heading":        true,
		"codeBlock":      true,
		"bulletList":     true,
		"orderedList":    true,
		"horizontalRule": true,
		"blockquote":     true,
	}

	return blockElements[currentType] && blockElements[nextType]
}
