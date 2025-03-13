package databases

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Connect initializes a database connection
func Connect() (*sql.DB, error) {
	dbURL := os.Getenv("SUPABASE_URL")
	if dbURL == "" {
		log.Fatal("SUPABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected successfully!")
	return db, nil
}
