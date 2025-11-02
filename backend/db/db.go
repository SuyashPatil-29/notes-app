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
		PreferSimpleProtocol: false, // Enable prepared statements for local DB (faster)
	}), &gorm.Config{
		PrepareStmt:            true,                                // Enable prepared statements for local DB performance
		SkipDefaultTransaction: true,                                // Improves performance
		Logger:                 logger.Default.LogMode(logger.Info), // Enable SQL query logging to debug slow queries
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

	// Drop existing tables to start fresh with CUID string IDs
	// WARNING: This will delete all existing data! Only run once!
	// Uncomment below lines if you need to recreate the database schema
	// log.Info().Msg("Dropping existing tables to recreate with CUID string IDs...")
	// err = DB.Migrator().DropTable(&models.Notes{}, &models.Chapter{}, &models.Notebook{})
	// if err != nil {
	// 	log.Warn().Err(err).Msg("Failed to drop tables (they might not exist yet)")
	// }

	// migrate the schema
	log.Info().Msg("Migrating database schema...")
	if err := DB.AutoMigrate(
		&models.Notebook{},
		&models.Chapter{},
		&models.Notes{},
		&models.AICredential{},
		&models.MeetingRecording{},
		&models.Calendar{},
		&models.CalendarEvent{},
		&models.CalendarOAuthState{},
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate schema")
	}

	log.Info().Msg("Database schema migrated successfully")
}
