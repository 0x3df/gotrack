package db

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"

	"dailytrack/models"
)

// ExportBundle is the JSON export/import envelope.
type ExportBundle struct {
	Version int            `json:"version"`
	Config  *models.Config `json:"config"`
	Entries []models.Entry `json:"entries"`
}

// ExportJSON writes the full config + all entries as pretty JSON.
func ExportJSON(w io.Writer) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	entries, err := GetAllEntries()
	if err != nil {
		return err
	}
	bundle := ExportBundle{Version: 1, Config: cfg, Entries: entries}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(bundle)
}

// ExportCSV writes a wide CSV: first column is date, remaining columns are
// one per tracker (in category order). Binary values are "true"/"false",
// numeric values are the float, text is the string.
func ExportCSV(w io.Writer) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("no config on disk")
	}
	entries, err := GetAllEntries()
	if err != nil {
		return err
	}

	var trackers []models.Tracker
	header := []string{"date"}
	for _, cat := range cfg.Categories {
		for _, t := range cat.Trackers {
			trackers = append(trackers, t)
			header = append(header, fmt.Sprintf("%s:%s", cat.Name, t.Name))
		}
	}

	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Date < entries[j].Date })
	for _, e := range entries {
		row := []string{e.Date}
		for _, t := range trackers {
			row = append(row, csvValue(e.Data[t.ID]))
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func csvValue(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return ""
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case string:
		return x
	}
	return fmt.Sprintf("%v", v)
}

// ImportJSON reads an ExportBundle and upserts its entries. If overwriteConfig
// is true, the on-disk config is replaced; otherwise the existing config is
// kept. Returns the number of entries written.
func ImportJSON(r io.Reader, overwriteConfig bool, dryRun bool) (int, error) {
	var bundle ExportBundle
	if err := json.NewDecoder(r).Decode(&bundle); err != nil {
		return 0, fmt.Errorf("decode: %w", err)
	}
	if dryRun {
		return len(bundle.Entries), nil
	}
	if overwriteConfig && bundle.Config != nil {
		if err := SaveConfig(bundle.Config); err != nil {
			return 0, fmt.Errorf("save config: %w", err)
		}
	}
	written := 0
	for i := range bundle.Entries {
		e := bundle.Entries[i]
		if err := UpsertEntry(&e); err != nil {
			return written, fmt.Errorf("upsert %s: %w", e.Date, err)
		}
		written++
	}
	return written, nil
}
