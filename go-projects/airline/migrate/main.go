package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

// LoadEnv loads the .env file
func LoadEnv() {
	// Get the absolute path of the .env file relative to the current working directory
	envPath, err := filepath.Abs("../go-projects/airline/.env") // Adjust based on your structure
	fmt.Println(envPath)
	if err != nil {
		log.Println("Could not resolve .env path:", err)
		return
	}

	// Load the .env file
	err = godotenv.Load(envPath)
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

func main() {
	LoadEnv()

	// get DB_PASSWORD from .env
	dbPassword := os.Getenv("DB_PASSWORD")

	//create a dsn string from this password to connect to local mysql
	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/airline?parseTime=true", dbPassword)

	// Open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	fmt.Println("Successfully connected to MySQL database!")
	//RunMigrations()

}

func RunMigrations() {
	// Database connection string
	const dbURL = "mysql://root:password@tcp(localhost:3306)/airline"
	m, err := migrate.New(
		"file://db/migrations", // Path to migration files
		dbURL,
	)
	if err != nil {
		log.Fatalf("Could not initialize migration: %v", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations applied successfully!")
}
