package db

import (
	"encoding/json"

	"dailytrack/models"
)

func UpsertEntry(entry *models.Entry) error {
	db, err := Open()
	if err != nil {
		return err
	}
	defer db.Close()

	dataJSON, err := json.Marshal(entry.Data)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO daily_entries (date, data)
		VALUES (?, ?)
		ON CONFLICT(date) DO UPDATE SET data=excluded.data
	`, entry.Date, string(dataJSON))
	return err
}
