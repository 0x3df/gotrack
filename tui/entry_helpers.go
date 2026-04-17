package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"dailytrack/models"
)

func validateEntryDate(s string) error {
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return fmt.Errorf("use YYYY-MM-DD")
	}
	return nil
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
