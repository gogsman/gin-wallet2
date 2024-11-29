package main

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

func InitDB() *sql.DB {
	once.Do(func() {
		connStr := "user=postgres password=123456 dbname=wallet sslmode=disable"
		var err error
		db, err = sql.Open("postgres", connStr)
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
