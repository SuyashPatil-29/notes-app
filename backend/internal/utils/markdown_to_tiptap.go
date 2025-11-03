package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

// TipTap node types
type TipTapDoc struct {
	Type    string        `json:"type"`
	Content []TipTapNode `json:"content"`
}

type TipTapNode struct {
	Type    string                 `json:"type"`
	Content []TipTapNode          `json:"content,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Marks   []TipTapMark          `json:"marks,omitempty"`
}

type TipTapMark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`
}

// MarkdownToTipTap converts markdown string to TipTap JSON
func MarkdownToTipTap(markdown string) (string, error) {
	if markdown == "" {
		// Return empty TipTap document
		doc := TipTapDoc{
			Type:    "doc",
			Content: []TipTapNode{},
		}
		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	}

	// Try to parse as JSON first - if it's already TipTap JSON, return as-is
	var testDoc TipTapDoc
	if err := json.Unmarshal([]byte(markdown), &testDoc); err == nil && testDoc.Type == "doc" {
		// Already valid TipTap JSON
		return markdown, nil
	}

	// Parse markdown into TipTap nodes
	nodes := parseMarkdownToNodes(markdown)

	doc := TipTapDoc{
		Type:    "doc",
		Content: nodes,
	}

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func parseMarkdownToNodes(markdown string) []TipTapNode {
	lines := strings.Split(markdown, "\n")
	nodes := []TipTapNode{}
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Skip empty lines (but add paragraph breaks for consecutive empty lines)
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// Headings
		if strings.HasPrefix(line, "#") {
			node := parseHeading(line)
			if node != nil {
				nodes = append(nodes, *node)
			}
			i++
			continue
		}

		// Code blocks
		if strings.HasPrefix(line, "```") {
			codeNode, linesConsumed := parseCodeBlock(lines, i)
			if codeNode != nil {
				nodes = append(nodes, *codeNode)
			}
			i += linesConsumed
			continue
		}

		// Bullet lists
		if strings.HasPrefix(strings.TrimSpace(line), "- ") || strings.HasPrefix(strings.TrimSpace(line), "* ") {
			listNode, linesConsumed := parseBulletList(lines, i)
			if listNode != nil {
				nodes = append(nodes, *listNode)
			}
			i += linesConsumed
			continue
		}

		// Ordered lists
		matched, _ := regexp.MatchString(`^\s*\d+\.\s`, line)
		if matched {
			listNode, linesConsumed := parseOrderedList(lines, i)
			if listNode != nil {
				nodes = append(nodes, *listNode)
			}
			i += linesConsumed
			continue
		}

		// Horizontal rule
		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "***" || strings.TrimSpace(line) == "___" {
			nodes = append(nodes, TipTapNode{Type: "horizontalRule"})
			i++
			continue
		}

		// Default: paragraph
		node := parseParagraph(line)
		if node != nil {
			nodes = append(nodes, *node)
		}
		i++
	}

	return nodes
}

func parseHeading(line string) *TipTapNode {
	re := regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 3 {
		level := len(matches[1])
		text := matches[2]
		return &TipTapNode{
			Type: "heading",
			Attrs: map[string]interface{}{
				"level": level,
			},
			Content: parseInlineContent(text),
		}
	}
	return nil
}

func parseCodeBlock(lines []string, startIdx int) (*TipTapNode, int) {
	firstLine := lines[startIdx]
	language := strings.TrimPrefix(firstLine, "```")
	language = strings.TrimSpace(language)

	codeLines := []string{}
	i := startIdx + 1

	// Find closing ```
	for i < len(lines) {
		if strings.HasPrefix(lines[i], "```") {
			break
		}
		codeLines = append(codeLines, lines[i])
		i++
	}

	codeContent := strings.Join(codeLines, "\n")

	node := &TipTapNode{
		Type: "codeBlock",
		Attrs: map[string]interface{}{
			"language": language,
		},
		Content: []TipTapNode{
			{
				Type: "text",
				Text: codeContent,
			},
		},
	}

	// Return node and number of lines consumed (including closing ```)
	linesConsumed := i - startIdx + 1
	return node, linesConsumed
}

func parseBulletList(lines []string, startIdx int) (*TipTapNode, int) {
	listItems := []TipTapNode{}
	i := startIdx

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "- ") && !strings.HasPrefix(line, "* ") {
			break
		}

		// Remove bullet prefix
		text := strings.TrimPrefix(line, "- ")
		text = strings.TrimPrefix(text, "* ")
		text = strings.TrimSpace(text)

		listItem := TipTapNode{
			Type: "listItem",
			Content: []TipTapNode{
				{
					Type:    "paragraph",
					Content: parseInlineContent(text),
				},
			},
		}
		listItems = append(listItems, listItem)
		i++
	}

	node := &TipTapNode{
		Type:    "bulletList",
		Content: listItems,
	}

	return node, i - startIdx
}

func parseOrderedList(lines []string, startIdx int) (*TipTapNode, int) {
	listItems := []TipTapNode{}
	i := startIdx
	re := regexp.MustCompile(`^\s*\d+\.\s+(.*)$`)

	for i < len(lines) {
		line := lines[i]
		matches := re.FindStringSubmatch(line)
		if len(matches) != 2 {
			break
		}

		text := strings.TrimSpace(matches[1])

		listItem := TipTapNode{
			Type: "listItem",
			Content: []TipTapNode{
				{
					Type:    "paragraph",
					Content: parseInlineContent(text),
				},
			},
		}
		listItems = append(listItems, listItem)
		i++
	}

	node := &TipTapNode{
		Type:    "orderedList",
		Content: listItems,
	}

	return node, i - startIdx
}

func parseParagraph(line string) *TipTapNode {
	text := strings.TrimSpace(line)
	if text == "" {
		return nil
	}

	return &TipTapNode{
		Type:    "paragraph",
		Content: parseInlineContent(text),
	}
}

func parseInlineContent(text string) []TipTapNode {
	// This is a simplified version - you can expand this to handle:
	// **bold**, *italic*, `code`, etc.

	nodes := []TipTapNode{}

	// Handle inline code first
	codeRe := regexp.MustCompile("`([^`]+)`")
	segments := []struct {
		text   string
		isCode bool
	}{}

	lastIdx := 0
	for _, match := range codeRe.FindAllStringIndex(text, -1) {
		if match[0] > lastIdx {
			segments = append(segments, struct {
				text   string
				isCode bool
			}{text: text[lastIdx:match[0]], isCode: false})
		}
		segments = append(segments, struct {
			text   string
			isCode bool
		}{text: text[match[0]+1 : match[1]-1], isCode: true})
		lastIdx = match[1]
	}
	if lastIdx < len(text) {
		segments = append(segments, struct {
			text   string
			isCode bool
		}{text: text[lastIdx:], isCode: false})
	}

	for _, segment := range segments {
		if segment.isCode {
			nodes = append(nodes, TipTapNode{
				Type: "text",
				Text: segment.text,
				Marks: []TipTapMark{
					{Type: "code"},
				},
			})
		} else {
			// Handle bold, italic, etc. in non-code segments
			nodes = append(nodes, parseFormattedText(segment.text)...)
		}
	}

	// If no special formatting found, return plain text
	if len(nodes) == 0 {
		nodes = append(nodes, TipTapNode{
			Type: "text",
			Text: text,
		})
	}

	return nodes
}

func parseFormattedText(text string) []TipTapNode {
	nodes := []TipTapNode{}

	// Bold: **text**
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	// Italic: *text*
	italicRe := regexp.MustCompile(`\*([^*]+)\*`)

	// For simplicity, just handle plain text for now
	// You can expand this to properly parse nested formatting
	if boldRe.MatchString(text) || italicRe.MatchString(text) {
		// Simplified: replace formatting markers
		cleanText := boldRe.ReplaceAllString(text, "$1")
		cleanText = italicRe.ReplaceAllString(cleanText, "$1")
		nodes = append(nodes, TipTapNode{
			Type: "text",
			Text: cleanText,
		})
	} else {
		nodes = append(nodes, TipTapNode{
			Type: "text",
			Text: text,
		})
	}

	return nodes
}

