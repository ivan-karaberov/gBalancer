package models

import (
	"fmt"
	"gBalancer/logger"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Loads environment variables from the .env file.
func InitENV(filename string) {
	if err := godotenv.Load(filename); err != nil {
		logrus.Fatalf("Error loading file %s > %v", filename, err)
	}
}

// Performs migration
func InitDB(db *gorm.DB, filename string) {
	err := db.AutoMigrate(&RateLimits{})
	if err != nil {
		logrus.Fatalf("Failed to migrate database > %v", err)
	}

	logrus.Info("Database initialized successfully")
}

// Create new connection to DB
func NewDBConnection() *gorm.DB {
	InitENV(".env")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"))

	var err error
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.CustomGormLogger(),
	})
	if err != nil {
		logrus.Fatal(err.Error())
	}

	logrus.Info("Successfully connected to the database")

	var tables []string
	result := DB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)
	if result.Error != nil {
		logrus.Fatalf("Failed retrieving table list: %s", result.Error.Error())
	}

	found := false
	for _, table := range tables {
		if table == "rate_limits" {
			found = true
			break
		}
	}

	if !found {
		InitDB(DB, "db.sql")
	}

	return DB
}
