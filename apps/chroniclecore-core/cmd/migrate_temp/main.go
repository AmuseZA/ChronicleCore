package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
    // 1. Resolve DB Path
    appData := os.Getenv("LOCALAPPDATA")
    if appData == "" {
        log.Fatal("LOCALAPPDATA not set")
    }
    dbPath := filepath.Join(appData, "ChronicleCore", "chronicle.db")
    
    fmt.Printf("Opening database: %s\n", dbPath)

    // 2. Open DB
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatalf("Failed to open DB: %v", err)
    }
    defer db.Close()

    // 3. Read Migration SQL
    // Assuming run from apps/chroniclecore-core
    migrationPath := filepath.Join("..", "..", "spec", "migrations", "001_currency_code.sql")
    fmt.Printf("Reading migration: %s\n", migrationPath)
    
    sqlBytes, err := ioutil.ReadFile(migrationPath)
    if err != nil {
        log.Fatalf("Failed to read migration file: %v", err)
    }

    // 4. Execute
    fmt.Println("Applying migration...")
    _, err = db.Exec(string(sqlBytes))
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    fmt.Println("✅ Migration applied successfully!")
}
