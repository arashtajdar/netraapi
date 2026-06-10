package config

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/go-sql-driver/mysql"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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

	runMigrations(DB, name)
}

// runMigrations automatically applies database migrations using golang-migrate
func runMigrations(db *sql.DB, dbName string) {
	log.Println("⏳ Checking database migrations...")
	
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("Could not start sql migration driver: %v", err)
	}

	// Use the embedded migrations filesystem
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("Could not load embedded migrations: %v", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatalf("Migration failed to initialize: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("An error occurred while running migrations: %v", err)
	}

	log.Println("✅ Database migrations applied successfully!")
}
