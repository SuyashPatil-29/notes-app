package db

import (
	"backend/internal/models"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Failed to load .env file:", err)
	}

	dsn := os.Getenv("DB_URL")

	// Open connection with local PostgreSQL optimized settings
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Use simple protocol to avoid prepared statement conflicts with hot-reload
	}), &gorm.Config{
		PrepareStmt:            false,                                // Disable prepared statements for hot-reload compatibility
		SkipDefaultTransaction: true,                                 // Improves performance
		Logger:                 logger.Default.LogMode(logger.Error), // Only log errors, not slow queries during startup
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Configure connection pool for local PostgreSQL
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database instance")
	}
	sqlDB.SetMaxOpenConns(25)                  // Higher for local DB
	sqlDB.SetMaxIdleConns(10)                  // More idle connections for local
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Longer lifetime for local
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // Longer idle time for local

	log.Info().Msg("Database connected successfully")

	// Run migrations in background to not block server startup
	go func() {
		log.Info().Msg("Starting database schema migration in background...")
		if err := DB.AutoMigrate(
			&models.Notebook{},
			&models.Chapter{},
			&models.Notes{},
			&models.NoteLink{},
			&models.TaskBoard{},
			&models.Task{},
			&models.TaskAssignment{},
			&models.AICredential{},
			&models.OrganizationAPICredential{},
			&models.MeetingRecording{},
			&models.Calendar{},
			&models.CalendarEvent{},
			&models.CalendarOAuthState{},
			&models.YjsDocument{},
			&models.YjsUpdate{},
			&models.WhatsAppUser{},
			&models.WhatsAppConversationContext{},
			&models.WhatsAppGroupLink{},
			&models.WhatsAppMessage{},
		); err != nil {
			log.Error().Err(err).Msg("Failed to migrate schema")
		} else {
			log.Info().Msg("Database schema migrated successfully")
		}
	}()
}
