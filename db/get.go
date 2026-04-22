package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"dailytrack/models"
)

// normalizeDate strips the time component the modernc sqlite driver adds when
// reading DATE columns ("2026-04-19T00:00:00Z" → "2026-04-19"). All internal
// code expects the plain ISO date.
func normalizeDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

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
		e.Date = normalizeDate(e.Date)
		if err := json.Unmarshal([]byte(dataStr), &e.Data); err != nil {
			return nil, fmt.Errorf("unmarshal entry %s: %w", e.Date, err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// GetEntriesBetween returns entries with date in [from, to] inclusive (YYYY-MM-DD),
// ordered newest-first, same convention as [GetAllEntries].
func GetEntriesBetween(from, to string) ([]models.Entry, error) {
	db, err := Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT date, data FROM daily_entries WHERE date >= ? AND date <= ? ORDER BY date DESC`,
		from, to,
	)
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
		e.Date = normalizeDate(e.Date)
		if err := json.Unmarshal([]byte(dataStr), &e.Data); err != nil {
			return nil, fmt.Errorf("unmarshal entry %s: %w", e.Date, err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	e.Date = normalizeDate(e.Date)
	if err := json.Unmarshal([]byte(dataStr), &e.Data); err != nil {
		return nil, fmt.Errorf("unmarshal entry %s: %w", e.Date, err)
	}
	return &e, nil
}
