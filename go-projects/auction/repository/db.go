package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // Importing MySQL driver
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	instance *sql.DB
	once     sync.Once
)

// LoadEnv loads the .env file
func loadEnv() {
	// Get the absolute path of the .env file relative to the current working directory
	// get $WORKSPACE from os environment variable
	// get the absolute path of the .env file
	workspacePath := os.Getenv("WORKSPACE")
	if workspacePath == "" {
		log.Fatal("WORKSPACE environment variable is not set")
	}
	path := filepath.Join(workspacePath, "kodex/go-projects/airline/.env")
	log.Debug("workspace path:", path)

	envPath, err := filepath.Abs(path)
	log.Debug("loading env from path:", envPath)
	if err != nil {
		log.Println("Could not resolve .env path:", err)
		return
	}

	// Load the .env file
	err = godotenv.Load(envPath)
	if err != nil {
		log.Fatal("No .env file found, using system environment variables")
	}
}

// GetDB returns a singleton DB instance
func GetDB() *sql.DB {

	once.Do(func() {
		loadEnv()
		// get DB_PASSWORD from .env
		dbPassword := os.Getenv("DB_PASSWORD")
		//create a dsn string from this password to connect to local mysql
		dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/auction?parseTime=true", dbPassword)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}

		db.SetMaxOpenConns(25)                 // Limit max open connections
		db.SetMaxIdleConns(10)                 // Limit idle connections
		db.SetConnMaxLifetime(5 * time.Minute) // Reuse connections for up to 5 minutes
		db.SetConnMaxIdleTime(2 * time.Minute) // Close idle connections after 2 minutes
		err = db.Ping()
		if err != nil {
			log.Fatal("Database is not reachable:", err)
		}
		instance = db
	})

	return instance
}
