// Package main database
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

// InitDB initializes a single database connection and returns it.
func InitDB() *sql.DB {
	once.Do(func() {

		dbConfig := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSL_MODE"),
		)

		var err error
		db, err = sql.Open("postgres", dbConfig)
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		// Ping to check connection
		if err := db.Ping(); err != nil {
			log.Fatal("Database is not reachable:", err)
		}
		log.Println("Database connection established successfully")
	})
	return db
}
