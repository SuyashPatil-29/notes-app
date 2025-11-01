package services

import (
	"backend/pkg/recallai"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	openaioption "github.com/openai/openai-go/option"
	"github.com/rs/zerolog/log"
)

type AIService struct {
	client *openai.Client
}

// TranscriptAnalysis represents the AI analysis result of a meeting transcript
type TranscriptAnalysis struct {
	NotebookName string   `json:"notebook_name"`
	ChapterName  string   `json:"chapter_name"`
	NoteName     string   `json:"note_name"`
	Summary      string   `json:"summary"`
	KeyPoints    []string `json:"key_points"`
	ActionItems  []string `json:"action_items"`
}

// NewAIService creates a new AI service instance
func NewAIService() *AIService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Warn().Msg("OPENAI_API_KEY not set, AI analysis will be unavailable")
		return &AIService{client: nil}
	}

	client := openai.NewClient(openaioption.WithAPIKey(apiKey))

	return &AIService{
		client: &client,
	}
}

// AnalyzeTranscript analyzes a meeting transcript and determines organizational structure
func (s *AIService) AnalyzeTranscript(ctx context.Context, transcript string, existingNotebooks []string) (*TranscriptAnalysis, error) {
	if s.client == nil {
		return nil, fmt.Errorf("OpenAI client not initialized - check OPENAI_API_KEY")
	}

	if strings.TrimSpace(transcript) == "" {
		return nil, fmt.Errorf("transcript cannot be empty")
	}

	notebooksStr := "None"
	if len(existingNotebooks) > 0 {
		notebooksJSON, _ := json.Marshal(existingNotebooks)
		notebooksStr = string(notebooksJSON)
	}

	systemPrompt := `You are an AI assistant that analyzes meeting transcripts and organizes them into a hierarchical note-taking structure.

Your task:
1. Analyze the meeting transcript
2. Determine appropriate Notebook, Chapter, and Note names
3. Create a concise summary
4. Extract key points and action items

Notebook: High-level category (e.g., "Work", "Personal", "Project X")
Chapter: Sub-category within notebook (e.g., "Q1 Planning", "Team Meetings")
Note: Specific meeting/topic (e.g., "Sprint Planning - Jan 15")

If the user has existing notebooks, try to fit the meeting into an existing notebook when appropriate.

Action Items: Extract specific tasks, follow-ups, or next steps mentioned in the meeting.

Respond ONLY with valid JSON in this exact format:
{
  "notebook_name": "string",
  "chapter_name": "string", 
  "note_name": "string",
  "summary": "string (2-3 sentences)",
  "key_points": ["point 1", "point 2", "point 3"],
  "action_items": ["action 1", "action 2", "action 3"]
}`

	userPrompt := fmt.Sprintf(`Existing notebooks: %s

Meeting Transcript:
%s

Analyze this transcript and provide the organizational structure and summary.`, notebooksStr, transcript)

	log.Info().
		Int("transcript_length", len(transcript)).
		Int("existing_notebooks_count", len(existingNotebooks)).
		Msg("Analyzing transcript with OpenAI")

	resp, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model: openai.ChatModelGPT4o,

		MaxTokens:   openai.Int(1000),
		Temperature: openai.Float(0.3), // Lower temperature for more consistent results
	})

	if err != nil {
		log.Error().Err(err).Msg("OpenAI API error during transcript analysis")
		return nil, fmt.Errorf("openai API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Clean the response content - OpenAI sometimes wraps JSON in markdown code blocks
	content := resp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// Remove markdown code block markers if present
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var analysis TranscriptAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		log.Error().
			Err(err).
			Str("response_content", resp.Choices[0].Message.Content).
			Str("cleaned_content", content).
			Msg("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Validate the analysis result
	if analysis.NotebookName == "" || analysis.ChapterName == "" || analysis.NoteName == "" {
		return nil, fmt.Errorf("AI analysis incomplete - missing required fields")
	}

	log.Info().
		Str("notebook_name", analysis.NotebookName).
		Str("chapter_name", analysis.ChapterName).
		Str("note_name", analysis.NoteName).
		Int("key_points_count", len(analysis.KeyPoints)).
		Msg("Successfully analyzed transcript")

	return &analysis, nil
}

// FormatTranscriptAsMarkdown converts a transcript and analysis into formatted markdown
func (s *AIService) FormatTranscriptAsMarkdown(transcript []recallai.TranscriptEntry, analysis *TranscriptAnalysis) string {
	if analysis == nil {
		return "# Meeting Transcript\n\n*Analysis unavailable*\n\n"
	}

	var md strings.Builder

	// Add title
	md.WriteString(fmt.Sprintf("# %s\n\n", analysis.NoteName))

	// Add summary section
	md.WriteString("## Summary\n\n")
	if analysis.Summary != "" {
		md.WriteString(analysis.Summary)
		md.WriteString("\n\n")
	} else {
		md.WriteString("*No summary available*\n\n")
	}

	// Add key points section
	if len(analysis.KeyPoints) > 0 {
		md.WriteString("## Key Points\n\n")
		for _, point := range analysis.KeyPoints {
			md.WriteString(fmt.Sprintf("- %s\n", point))
		}
		md.WriteString("\n")
	}

	// Add action items section
	if len(analysis.ActionItems) > 0 {
		md.WriteString("## Action Items\n\n")
		for _, item := range analysis.ActionItems {
			md.WriteString(fmt.Sprintf("- [ ] %s\n", item))
		}
		md.WriteString("\n")
	}

	// Add participants section if we have transcript data
	if len(transcript) > 0 {
		participants := s.extractParticipants(transcript)
		if len(participants) > 0 {
			md.WriteString("## Participants\n\n")
			for _, participant := range participants {
				hostIndicator := ""
				if participant.IsHost {
					hostIndicator = " (Host)"
				}
				md.WriteString(fmt.Sprintf("- %s%s\n", participant.Name, hostIndicator))
			}
			md.WriteString("\n")
		}
	}

	// Add full transcript section
	md.WriteString("## Full Transcript\n\n")

	if len(transcript) == 0 {
		md.WriteString("*No transcript data available*\n")
		return md.String()
	}

	// Add horizontal rule for visual separation
	md.WriteString("---\n\n")

	currentParticipant := ""
	var currentSentence strings.Builder

	for i, entry := range transcript {
		// Collect all words for this entry
		var words []string
		for _, word := range entry.Words {
			words = append(words, word.Text)
		}

		if len(words) == 0 {
			continue
		}

		entryText := strings.Join(words, " ")

		// Check if this is a new speaker
		if entry.Participant.Name != currentParticipant {
			// Write previous sentence if exists
			if currentSentence.Len() > 0 {
				md.WriteString(currentSentence.String())
				md.WriteString("\n\n")
				currentSentence.Reset()
			}

			// Add spacing between speakers (except for the first one)
			if currentParticipant != "" {
				md.WriteString("\n")
			}

			currentParticipant = entry.Participant.Name

			// Write speaker name with better formatting
			hostIndicator := ""
			if entry.Participant.IsHost {
				hostIndicator = " 👤"
			}
			md.WriteString(fmt.Sprintf("### %s%s\n\n", currentParticipant, hostIndicator))
		}

		// Add the text to current sentence
		currentSentence.WriteString(entryText)

		// Add space if not the last entry
		if i < len(transcript)-1 {
			currentSentence.WriteString(" ")
		}
	}

	// Write final sentence if exists
	if currentSentence.Len() > 0 {
		md.WriteString(currentSentence.String())
		md.WriteString("\n")
	}

	return md.String()
}

// extractParticipants extracts unique participants from the transcript
func (s *AIService) extractParticipants(transcript []recallai.TranscriptEntry) []recallai.ParticipantInfo {
	participantMap := make(map[int]recallai.ParticipantInfo)

	for _, entry := range transcript {
		if _, exists := participantMap[entry.Participant.ID]; !exists {
			participantMap[entry.Participant.ID] = entry.Participant
		}
	}

	participants := make([]recallai.ParticipantInfo, 0, len(participantMap))
	for _, participant := range participantMap {
		participants = append(participants, participant)
	}

	return participants
}

// TranscriptToPlainText converts transcript entries to plain text for AI analysis
func (s *AIService) TranscriptToPlainText(transcript []recallai.TranscriptEntry) string {
	var text strings.Builder

	for _, entry := range transcript {
		text.WriteString(fmt.Sprintf("%s: ", entry.Participant.Name))
		for _, word := range entry.Words {
			text.WriteString(word.Text)
			text.WriteString(" ")
		}
		text.WriteString("\n")
	}

	return text.String()
}

// CreateFallbackAnalysis creates a basic analysis when AI is unavailable
func (s *AIService) CreateFallbackAnalysis(meetingURL string) *TranscriptAnalysis {
	// Extract some basic info from the meeting URL if possible
	notebookName := "Meetings"
	chapterName := "Recorded Meetings"

	// Try to determine platform from URL
	platform := "Unknown Platform"
	if strings.Contains(meetingURL, "meet.google.com") {
		platform = "Google Meet"
	} else if strings.Contains(meetingURL, "zoom.us") {
		platform = "Zoom"
	} else if strings.Contains(meetingURL, "teams.microsoft.com") {
		platform = "Microsoft Teams"
	}

	// Create a basic note name with timestamp
	noteName := fmt.Sprintf("%s Meeting - %s", platform,
		fmt.Sprintf("%d", time.Now().Unix()))

	return &TranscriptAnalysis{
		NotebookName: notebookName,
		ChapterName:  chapterName,
		NoteName:     noteName,
		Summary:      "Meeting transcript processed without AI analysis.",
		KeyPoints:    []string{"Meeting recorded and transcribed", "AI analysis unavailable"},
	}
}
