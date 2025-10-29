package main

import (
	"backend/db"
	"backend/internal/auth"
	"backend/internal/controllers"
	"backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	db.InitDB()

	// Initialize authentication
	auth.InitAuth()

	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
	}))

	// Public routes (no authentication required)
	public := r.Group("/")
	{
		// OAuth routes - public for authentication flow
		public.GET("/auth/:provider", auth.BeginAuth)
		public.GET("/auth/:provider/callback", auth.AuthCallback)
	}

	// Protected routes (authentication required)
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		// Auth routes
		protected.GET("/auth/user", auth.GetCurrentUser)
		protected.GET("/logout/:provider", auth.Logout)

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
	}

	r.Run(":8080")
}
