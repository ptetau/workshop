package storage

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Init(dbPath string) *sql.DB {
	// _journal_mode=WAL:     Enables concurrency
	// _busy_timeout=5000:    Wait 5s for a lock
	// _foreign_keys=on:      Enforce relations
	// _synchronous=NORMAL:   Trade small safety for high speed
	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on&_synchronous=NORMAL"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("Fatal: Could not open DB: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(1 * time.Hour)

	if err := db.Ping(); err != nil {
		log.Fatalf("Fatal: Database unreachable: %v", err)
	}

	log.Println("Database initialized successfully in WAL mode.")
	return db
}
