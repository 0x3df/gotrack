package db

import (
	"encoding/json"
	"fmt"
	"strings"

	"dailytrack/models"
)

func DeleteEntry(date string) error {
	db, err := Open()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`DELETE FROM daily_entries WHERE date = ?`, date)
	return err
}

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

// UpsertEntryLog merges tracker values into the entry for date (same semantics
// as the `gotrack log` CLI): unknown keys are resolved via [LookupTracker],
// values are coerced with [CoerceValueFromInterface], then the row is upserted.
func UpsertEntryLog(cfg *models.Config, date string, values map[string]interface{}) error {
	if cfg == nil {
		return fmt.Errorf("no config loaded")
	}
	if len(values) == 0 {
		return fmt.Errorf("no values to log")
	}

	existing, err := GetEntryForDate(date)
	if err != nil {
		return err
	}
	data := map[string]interface{}{}
	if existing != nil && existing.Data != nil {
		for k, v := range existing.Data {
			data[k] = v
		}
	}

	for key, raw := range values {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		t, err := LookupTracker(cfg, k)
		if err != nil {
			return err
		}
		v, err := CoerceValueFromInterface(t, raw)
		if err != nil {
			return fmt.Errorf("%s: %w", t.Name, err)
		}
		data[t.ID] = v
	}

	entry := &models.Entry{Date: date, Data: data}
	return UpsertEntry(entry)
}
