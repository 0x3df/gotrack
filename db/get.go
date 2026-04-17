package db

import (
	"encoding/json"

	"dailytrack/models"
)

func GetAllEntries() ([]models.Entry, error) {
	db, err := Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT date, data FROM daily_entries ORDER BY date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.Entry
	for rows.Next() {
		var e models.Entry
		var dataStr string
		if err := rows.Scan(&e.Date, &dataStr); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(dataStr), &e.Data)
		entries = append(entries, e)
	}
	return entries, nil
}

func GetEntryForDate(date string) (*models.Entry, error) {
	db, err := Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var e models.Entry
	var dataStr string
	row := db.QueryRow(`SELECT date, data FROM daily_entries WHERE date = ?`, date)
	if err := row.Scan(&e.Date, &dataStr); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(dataStr), &e.Data)
	return &e, nil
}
