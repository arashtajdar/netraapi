package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// ConnectDB initializes the MySQL connection pool
func ConnectDB() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, name)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Optimize connection pool for production
	DB.SetMaxOpenConns(25)                 // Limit max open connections to prevent DB saturation
	DB.SetMaxIdleConns(25)                 // Keep connections open for fast reuse
	DB.SetConnMaxLifetime(5 * time.Minute) // Retire old connections

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("✅ Successfully connected to MySQL database!")
}
