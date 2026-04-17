package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func GetDBPath() (string, error) {
	workspace, err := GetWorkspacePath()
	if err != nil {
		return "", err
	}
	if workspace == "" {
		return "", os.ErrNotExist
	}
	return filepath.Join(workspace, "data.db"), nil
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

	schema := `
	CREATE TABLE IF NOT EXISTS daily_entries (
		date DATE PRIMARY KEY,
		data TEXT NOT NULL DEFAULT '{}'
	);
	`
	_, err = db.Exec(schema)
	return err
}
