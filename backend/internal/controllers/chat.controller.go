package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"backend/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/pkg/utils"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/coder/aisdk-go"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	openaioption "github.com/openai/openai-go/option"
	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
)

// ChatRequest wraps the aisdk.Chat with optional extra fields
type ChatRequest struct {
	aisdk.Chat
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Thinking bool   `json:"thinking"`
}

var (
	lastMessages []aisdk.Message
)

func getOpenAIClient(apiKey string) *openai.Client {
	client := openai.NewClient(openaioption.WithAPIKey(apiKey))
	return &client
}

func getAnthropicClient(apiKey string) *anthropic.Client {
	client := anthropic.NewClient(anthropicoption.WithAPIKey(apiKey))
	return &client
}

func getGoogleClient(ctx context.Context, apiKey string) (*genai.Client, error) {
	return genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
}

// getUserAPIKey retrieves the user's API key for a specific provider
func getUserAPIKey(userID uint, provider string) (string, error) {
	var credential models.AICredential
	if err := db.DB.Where("user_id = ? AND provider = ?", userID, provider).First(&credential).Error; err != nil {
		return "", err
	}

	// Decrypt the API key
	apiKey, err := utils.Decrypt(credential.KeyCipher)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decrypt API key")
		return "", err
	}

	// Trim whitespace (leading/trailing spaces/newlines that might have been accidentally saved)
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return "", fmt.Errorf("decrypted API key is empty")
	}

	return apiKey, nil
}

// DumpHandler dumps the last messages to a JSON file
func DumpHandler(c *gin.Context) {
	data, _ := json.MarshalIndent(lastMessages, "", "  ")
	err := os.WriteFile("dump.json", data, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write dump file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dump messages"})
		return
	}
	log.Info().Msg("Dumped messages to dump.json")
	c.JSON(http.StatusOK, gin.H{"message": "Dumped to dump.json"})
}

// ChatHandler handles the chat API endpoint with multi-provider support
func ChatHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse chat request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user ID from session
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user's API key for the requested provider
	apiKey, err := getUserAPIKey(userID, req.Provider)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Uint("userID", userID).Msg("Failed to get user API key")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key not configured for this provider. Please set up your API key in settings."})
		return
	}

	// Define tools
	tools := []aisdk.Tool{{
		Name:        "test",
		Description: "A test tool. Only use if the user explicitly requests it.",
		Schema: aisdk.Schema{
			Required: []string{"message"},
			Properties: map[string]any{
				"message": map[string]any{
					"type": "string",
				},
			},
		},
	}}

	// Tool handler
	handleToolCall := func(toolCall aisdk.ToolCall) any {
		return map[string]string{
			"message": "It worked!",
		}
	}

	// Set data stream headers
	aisdk.WriteDataStreamHeaders(c.Writer)

	// Add system message if not present
	if len(req.Messages) == 0 || req.Messages[0].Role != "system" {
		req.Messages = append([]aisdk.Message{{
			Role:    "system",
			Content: "You are a helpful assistant. When using tools, always provide a text response after receiving the tool result to describe what happened. Do not make additional tool calls unless explicitly requested by the user.",
		}}, req.Messages...)
	}

	// Main streaming loop (handles tool calls)
	for {
		var stream aisdk.DataStream

		switch req.Provider {
		case "openai":
			messages, err := aisdk.MessagesToOpenAI(req.Messages)
			if err != nil {
				log.Error().Err(err).Msg("Failed to convert messages for OpenAI")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare messages"})
				return
			}

			reasoningEffort := openai.ReasoningEffort("")
			if req.Thinking {
				reasoningEffort = openai.ReasoningEffortMedium
			}

			client := getOpenAIClient(apiKey)
			stream = aisdk.OpenAIToDataStream(client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
				Model:               req.Model,
				Messages:            messages,
				ReasoningEffort:     reasoningEffort,
				Tools:               aisdk.ToolsToOpenAI(tools),
				MaxCompletionTokens: openai.Int(2048),
			}))

		case "anthropic":
			messages, system, err := aisdk.MessagesToAnthropic(req.Messages)
			if err != nil {
				log.Error().Err(err).Msg("Failed to convert messages for Anthropic")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare messages"})
				return
			}

			thinking := anthropic.ThinkingConfigParamUnion{}
			if req.Thinking {
				thinking = anthropic.ThinkingConfigParamOfEnabled(2048)
			}

			client := getAnthropicClient(apiKey)
			stream = aisdk.AnthropicToDataStream(client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
				Model:     anthropic.Model(req.Model),
				Messages:  messages,
				System:    system,
				MaxTokens: 4096,
				Thinking:  thinking,
				Tools:     aisdk.ToolsToAnthropic(tools),
			}))

		case "google":
			googleClient, err := getGoogleClient(ctx, apiKey)
			if err != nil {
				log.Error().Err(err).Msg("Google client not initialized")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Google client not configured"})
				return
			}

			messages, err := aisdk.MessagesToGoogle(req.Messages)
			if err != nil {
				log.Error().Err(err).Msg("Failed to convert messages for Google")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare messages"})
				return
			}

			var thinkingConfig *genai.ThinkingConfig
			if req.Thinking {
				thinkingConfig = &genai.ThinkingConfig{
					IncludeThoughts: true,
				}
			}

			googleTools, err := aisdk.ToolsToGoogle(tools)
			if err != nil {
				log.Error().Err(err).Msg("Failed to convert tools for Google")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare tools"})
				return
			}

			stream = aisdk.GoogleToDataStream(googleClient.Models.GenerateContentStream(ctx, req.Model, messages, &genai.GenerateContentConfig{
				Tools:          googleTools,
				ThinkingConfig: thinkingConfig,
			}))

		default:
			log.Error().Str("provider", req.Provider).Msg("Invalid provider")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
			return
		}

		if stream == nil {
			log.Error().Msg("Failed to create stream")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stream"})
			return
		}

		// Setup accumulator and tool calling
		var acc aisdk.DataStreamAccumulator
		stream = stream.WithToolCalling(handleToolCall)
		stream = stream.WithAccumulator(&acc)

		// Pipe the stream to the response writer
		err := stream.Pipe(c.Writer)
		if err != nil {
			log.Error().Err(err).Msg("Error piping AI response stream")

			// Try to extract error message from the error
			errorMsg := err.Error()
			userFriendlyMsg := errorMsg

			// Check if it's an API key error
			if strings.Contains(errorMsg, "401") || strings.Contains(errorMsg, "Unauthorized") || strings.Contains(errorMsg, "invalid_api_key") || strings.Contains(errorMsg, "Incorrect API key") {
				userFriendlyMsg = "Invalid API key. Please check your API key in Profile settings and ensure it's correct."
			} else if strings.Contains(errorMsg, "insufficient_quota") {
				userFriendlyMsg = "API quota exceeded. Please check your API usage limits."
			} else if strings.Contains(errorMsg, "rate_limit") {
				userFriendlyMsg = "Rate limit exceeded. Please try again in a moment."
			} else if strings.Contains(errorMsg, "model_not_found") {
				userFriendlyMsg = "The selected model is not available. Please try a different model."
			} else {
				// Generic error message
				userFriendlyMsg = "An error occurred while processing your request. Please try again."
			}

			// Return JSON error response instead of trying to write to stream
			c.JSON(http.StatusInternalServerError, gin.H{"error": userFriendlyMsg})
			return
		}

		// Update messages with accumulated content
		req.Messages = append(req.Messages, acc.Messages()...)
		lastMessages = req.Messages[:]

		// Continue if tool calls need to be processed
		if acc.FinishReason() == aisdk.FinishReasonToolCalls {
			continue
		}

		break
	}
}
