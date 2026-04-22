package db

import (
	"fmt"
	"strconv"
	"strings"

	"dailytrack/models"
)

// CoerceValue converts a raw user-supplied string into the correct Go type
// for a tracker. Returned values use the same types stored in Entry.Data:
// bool for TrackerBinary, float64 for numeric-like trackers, string for text.
func CoerceValue(t models.Tracker, raw string) (interface{}, error) {
	raw = strings.TrimSpace(raw)
	switch t.Type {
	case models.TrackerBinary:
		switch strings.ToLower(raw) {
		case "true", "t", "yes", "y", "1", "on", "done":
			return true, nil
		case "false", "f", "no", "n", "0", "off":
			return false, nil
		}
		return nil, fmt.Errorf("binary value must be true/false (got %q)", raw)
	case models.TrackerDuration, models.TrackerCount, models.TrackerNumeric:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("numeric value expected (got %q)", raw)
		}
		return v, nil
	case models.TrackerRating:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("rating must be 1-5 (got %q)", raw)
		}
		if v < 1 || v > 5 {
			return nil, fmt.Errorf("rating must be 1-5 (got %v)", v)
		}
		return v, nil
	case models.TrackerText:
		return raw, nil
	}
	return nil, fmt.Errorf("unknown tracker type %q", t.Type)
}

// valueToCoerceString normalizes JSON-decoded scalars for [CoerceValue].
func valueToCoerceString(v interface{}) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	case bool:
		return fmt.Sprintf("%v", x), nil
	case int:
		return strconv.Itoa(x), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	default:
		return "", fmt.Errorf("unsupported JSON value type %T", v)
	}
}

// CoerceValueFromInterface converts a JSON-decoded value into the storage type
// for tracker t (same contract as [CoerceValue]).
func CoerceValueFromInterface(t models.Tracker, v interface{}) (interface{}, error) {
	if v == nil {
		return nil, fmt.Errorf("null value not allowed")
	}
	switch t.Type {
	case models.TrackerBinary:
		if b, ok := v.(bool); ok {
			return b, nil
		}
		s, err := valueToCoerceString(v)
		if err != nil {
			return nil, err
		}
		return CoerceValue(t, s)
	case models.TrackerText:
		if s, ok := v.(string); ok {
			return CoerceValue(t, s)
		}
		s, err := valueToCoerceString(v)
		if err != nil {
			return nil, err
		}
		return CoerceValue(t, s)
	default:
		s, err := valueToCoerceString(v)
		if err != nil {
			return nil, err
		}
		return CoerceValue(t, s)
	}
}

// LookupTracker resolves a case-insensitive name or ID against the config
// and returns the matching tracker or an error with a "did you mean" hint.
func LookupTracker(cfg *models.Config, key string) (models.Tracker, error) {
	if cfg == nil {
		return models.Tracker{}, fmt.Errorf("no config loaded")
	}
	needle := strings.ToLower(strings.TrimSpace(key))
	var names []string
	for _, cat := range cfg.Categories {
		for _, t := range cat.Trackers {
			if strings.EqualFold(t.ID, needle) || strings.EqualFold(t.Name, needle) {
				return t, nil
			}
			// allow slug-like matches (spaces collapsed)
			if strings.ReplaceAll(strings.ToLower(t.Name), " ", "") == strings.ReplaceAll(needle, " ", "") {
				return t, nil
			}
			names = append(names, t.Name)
		}
	}
	hint := ""
	if best := closestName(needle, names); best != "" {
		hint = fmt.Sprintf(" (did you mean %q?)", best)
	}
	return models.Tracker{}, fmt.Errorf("tracker %q not found%s", key, hint)
}

func closestName(needle string, names []string) string {
	best := ""
	bestScore := 0
	for _, n := range names {
		s := commonSubstringScore(strings.ToLower(n), needle)
		if s > bestScore {
			bestScore = s
			best = n
		}
	}
	if bestScore < 3 {
		return ""
	}
	return best
}

func commonSubstringScore(a, b string) int {
	score := 0
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] == b[i] {
			score++
		} else {
			break
		}
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		score += 3
	}
	return score
}
