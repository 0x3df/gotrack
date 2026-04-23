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

// AddDurationToEntry adds minutes to a duration tracker on the entry for date.
// It preserves all other tracker values and creates the entry when needed.
func AddDurationToEntry(cfg *models.Config, date string, trackerID string, minutes float64) error {
	if cfg == nil {
		return fmt.Errorf("no config loaded")
	}
	if minutes <= 0 {
		return fmt.Errorf("minutes must be positive")
	}

	var tracker models.Tracker
	found := false
	for _, cat := range cfg.Categories {
		for _, t := range cat.Trackers {
			if t.ID == trackerID {
				tracker = t
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return fmt.Errorf("tracker %q not found", trackerID)
	}
	if tracker.Type != models.TrackerDuration {
		return fmt.Errorf("%s is not a duration tracker", tracker.Name)
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

	current := 0.0
	if raw, ok := data[trackerID]; ok {
		switch v := raw.(type) {
		case float64:
			current = v
		case int:
			current = float64(v)
		default:
			return fmt.Errorf("%s has non-numeric existing value", tracker.Name)
		}
	}
	data[trackerID] = current + minutes

	return UpsertEntry(&models.Entry{Date: date, Data: data})
}
