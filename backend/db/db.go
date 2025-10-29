package db

import (
	"backend/internal/models"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Failed to load .env file:", err)
	}

	dsn := os.Getenv("DB_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	log.Info().Msg("Database connected successfully")

	// Drop existing tables to start fresh with CUID string IDs
	// WARNING: This will delete all existing data! Only run once!
	// Uncomment below lines if you need to recreate the database schema
	// log.Info().Msg("Dropping existing tables to recreate with CUID string IDs...")
	// err = DB.Migrator().DropTable(&models.Notes{}, &models.Chapter{}, &models.Notebook{}, &models.User{})
	// if err != nil {
	// 	log.Warn().Err(err).Msg("Failed to drop tables (they might not exist yet)")
	// }

	// migrate the schema
	log.Info().Msg("Migrating database schema...")
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Notebook{},
		&models.Chapter{},
		&models.Notes{},
	); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate schema")
	}

	log.Info().Msg("Database schema migrated successfully")
}
