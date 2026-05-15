package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is a global variable that holds our database connection pool.
// Since it is capitalized, we can access it from our handlers later via config.DB
var DB *gorm.DB

// ConnectDatabase initializes the connection to Supabase
func ConnectDatabase() {
	// 1. Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Warning: No .env file found. Relying on system environment variables.")
	}

	// 2. Fetch the DATABASE_URL string
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ ERROR: DATABASE_URL is not set in the .env file")
	}

	// 3. Open the connection using GORM
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ ERROR: Failed to connect to the database: \n", err)
	}

	fmt.Println("✅ Successfully connected to Supabase PostgreSQL!")

	// 4. Assign the connection to our global variable
	DB = database
}
