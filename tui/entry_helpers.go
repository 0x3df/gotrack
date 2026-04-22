package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"dailytrack/db"
	"dailytrack/models"
)

func normalizeEntryDate(s string) (string, error) {
	return db.NormalizeDate(s)
}

func validateEntryDate(s string) error {
	_, err := normalizeEntryDate(s)
	return err
}

func prefillBoolValue(data map[string]interface{}, trackerID string) bool {
	if data == nil {
		return false
	}
	b, _ := data[trackerID].(bool)
	return b
}

func entryData(entry *models.Entry) map[string]interface{} {
	if entry == nil {
		return nil
	}
	return entry.Data
}

func prefillStringValue(data map[string]interface{}, trackerID string) string {
	if data == nil {
		return ""
	}
	switch v := data[trackerID].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

func prefillIntValue(data map[string]interface{}, trackerID string, fallback int) int {
	if data == nil {
		return fallback
	}
	switch v := data[trackerID].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return fallback
	}
}

func trackerUnit(t models.Tracker) string {
	if strings.TrimSpace(t.Unit) != "" {
		return strings.TrimSpace(t.Unit)
	}
	return models.DefaultUnit(t.Name, t.Type)
}

func formatFloatValue(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// roundAverageUp2 rounds v up to two decimal places (toward +∞ at cent resolution).
func roundAverageUp2(v float64) float64 {
	return math.Ceil(v*100) / 100
}

// formatAverageFixed2 formats an average rounded up to two decimal places.
func formatAverageFixed2(v float64) string {
	return strconv.FormatFloat(roundAverageUp2(v), 'f', 2, 64)
}

// formatAverageWithUnit formats a rounded-up average with the tracker's unit when set.
func formatAverageWithUnit(v float64, t models.Tracker) string {
	value := formatAverageFixed2(v)
	unit := trackerUnit(t)
	if unit == "" {
		return value
	}
	return value + " " + unit
}

func trackerInputLabel(t models.Tracker) string {
	unit := trackerUnit(t)
	if unit == "" {
		if t.Target == nil {
			return t.Name
		}
		return fmt.Sprintf("%s (target: %s)", t.Name, formatFloatValue(*t.Target))
	}
	if t.Target == nil {
		return fmt.Sprintf("%s (%s)", t.Name, unit)
	}
	return fmt.Sprintf("%s (%s, target: %s %s)", t.Name, unit, formatFloatValue(*t.Target), unit)
}

func formatValueWithUnit(v float64, t models.Tracker) string {
	value := formatFloatValue(v)
	unit := trackerUnit(t)
	if unit == "" {
		return value
	}
	return value + " " + unit
}
