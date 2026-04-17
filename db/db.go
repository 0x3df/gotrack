package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func GetDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", "dailytrack")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "data.db"), nil
}

func Open() (*sql.DB, error) {
	dbPath, err := GetDBPath()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InitDB() error {
	db, err := Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Drop old table for clean break (removes day_rating column)
	db.Exec(`DROP TABLE IF EXISTS daily_entries`)

	schema := `
	CREATE TABLE IF NOT EXISTS daily_entries (
		date DATE PRIMARY KEY,
		data TEXT NOT NULL DEFAULT '{}'
	);
	`
	_, err = db.Exec(schema)
	return err
}
