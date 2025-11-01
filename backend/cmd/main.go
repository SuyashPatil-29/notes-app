package main

import (
	"backend/db"
	"backend/internal/auth"
	"backend/internal/controllers"
	"backend/internal/middleware"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Warn().Msg("Error loading .env file, using environment variables")
	}

	// Initialize database
	db.InitDB()

	// Initialize calendar OAuth
	auth.InitCalendarOAuth()

	// Initialize Clerk SDK
	clerkSecretKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecretKey == "" {
		log.Fatal().Msg("CLERK_SECRET_KEY environment variable is required")
	}
	clerk.SetKey(clerkSecretKey)

	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public routes (no authentication required)
	public := r.Group("/")
	{
		// Calendar OAuth callback routes - public for OAuth flow
		public.GET("/api/calendar/google/callback", auth.GoogleCalendarCallback)
		public.GET("/api/calendar/microsoft/callback", auth.MicrosoftCalendarCallback)

		// Public content routes
		public.GET("/public/:notebookId", controllers.GetPublicNotebook)
		public.GET("/public/:notebookId/:chapterId", controllers.GetPublicChapter)
		public.GET("/public/:notebookId/:chapterId/:noteId", controllers.GetPublicNote)
		public.GET("/public/user/:email", controllers.GetPublicUserProfile)
	}

	// Protected routes (authentication required via Clerk)
	protected := r.Group("/")
	protected.Use(middleware.ClerkMiddleware())
	protected.Use(middleware.RequireAuth())
	{
		// Auth routes
		protected.GET("/auth/user", auth.GetCurrentUser)

		// Onboarding routes
		protected.GET("/onboarding", auth.GetOnboardingStatus)
		protected.POST("/onboarding", auth.CompleteOnboarding)
		protected.DELETE("/onboarding", auth.ResetOnboarding) // Dev: Reset onboarding

		// AI credentials routes
		protected.GET("/settings/ai-credentials", auth.GetAICredentials)
		protected.POST("/settings/ai-credentials", auth.SetAICredential)
		protected.DELETE("/settings/ai-credentials", auth.DeleteAICredential)

		// Chat/AI routes
		protected.POST("/api/chat", controllers.ChatHandler)
		protected.POST("/api/generate", controllers.GenerateHandler)
		protected.GET("/api/dump", controllers.DumpHandler)

		// Notebook routes
		protected.POST("/notebook", controllers.CreateNotebook)
		protected.GET("/notebooks", controllers.GetUserNotebooks)
		protected.GET("/notebooks/:id/chapters", controllers.GetChaptersByNotebook) // More specific route first
		protected.GET("/notebook/:id", controllers.GetNotebookById)
		protected.PUT("/notebook/:id", controllers.UpdateNotebook)
		protected.DELETE("/notebook/:id", controllers.DeleteNotebook)

		// Chapter routes
		protected.POST("/chapter", controllers.CreateChapter)
		protected.GET("/chapters/:id/notes", controllers.GetNotesByChapter) // More specific route first
		protected.GET("/chapter/:id", controllers.GetChapterById)
		protected.PUT("/chapter/:id", controllers.UpdateChapter)
		protected.PATCH("/chapter/:id/move", controllers.MoveChapter)
		protected.DELETE("/chapter/:id", controllers.DeleteChapter)

		// Note routes
		protected.POST("/note", controllers.CreateNote)
		protected.GET("/note/:id", controllers.GetNoteById)
		protected.PUT("/note/:id", controllers.UpdateNote)
		protected.PATCH("/note/:id/move", controllers.MoveNote)
		protected.DELETE("/note/:id", controllers.DeleteNote)
		protected.POST("/note/:id/generate-video", controllers.GenerateNoteVideo)
		protected.DELETE("/note/:id/video", controllers.DeleteNoteVideo)

		// Publishing routes
		protected.POST("/notebook/:id/publish", controllers.PublishNotebook)
		protected.PUT("/notebook/:id/published-notes", controllers.UpdatePublishedNotes)
		protected.POST("/notebook/:id/unpublish", controllers.UnpublishNotebook)
		protected.PATCH("/note/:id/publish", controllers.PublishNote)

		// Meeting routes
		protected.POST("/meeting/start", controllers.StartMeetingRecording)
		protected.GET("/meetings", controllers.GetUserMeetings)
		protected.GET("/meeting/:id/transcript", controllers.GetMeetingTranscript)
		protected.POST("/meetings/backfill-videos", controllers.BackfillVideoURLs)

		// Calendar routes
		protected.POST("/api/calendar-auth/:provider", auth.BeginCalendarOAuth) // Initiate OAuth flow
		protected.GET("/api/calendars", controllers.GetUserCalendars)
		protected.POST("/api/calendars/sync-missing", controllers.SyncMissingCalendars)
		protected.DELETE("/api/calendars/:id", controllers.DisconnectCalendar)
		protected.GET("/api/calendars/:id/events", controllers.GetCalendarEvents)
		protected.POST("/api/calendars/:id/sync", controllers.SyncCalendarEvents)
		protected.POST("/api/calendar-events/:eventId/schedule-bot", controllers.ScheduleBotForEvent)
		protected.DELETE("/api/calendar-events/:eventId/cancel-bot", controllers.CancelBotForEvent)
	}

	// Webhook routes (no authentication required for external services)
	webhook := r.Group("/webhooks")
	{
		webhook.POST("/recall", controllers.HandleRecallWebhook)
		webhook.POST("/calendar/sync", controllers.HandleCalendarWebhook)
		webhook.POST("/clerk", auth.UserCreatedWebhook) // Clerk user sync webhook
	}

	r.Run(":8080")
}
