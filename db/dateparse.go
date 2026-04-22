package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NormalizeDate accepts a user-supplied date string and returns it in
// ISO YYYY-MM-DD form. Accepted inputs:
//
//	"t", "today"           → today
//	"y", "yesterday"       → yesterday
//	"-N"                   → N days ago
//	"YYYY-MM-DD"           → passthrough
//	"YYYY/MM/DD"           → slash variant
//	"MM-DD" / "M-D"        → current year (nearest past occurrence if future)
//	"MM/DD" / "M/D"        → current year (nearest past occurrence if future)
func NormalizeDate(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	now := time.Now()
	switch s {
	case "t", "today":
		return now.Format("2006-01-02"), nil
	case "y", "yesterday":
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	}
	if strings.HasPrefix(s, "-") {
		if n, err := strconv.Atoi(s[1:]); err == nil && n >= 0 {
			return now.AddDate(0, 0, -n).Format("2006-01-02"), nil
		}
	}
	normalized := strings.ReplaceAll(s, "/", "-")
	for _, layout := range []string{"2006-01-02", "2006-1-2"} {
		if t, err := time.Parse(layout, normalized); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	for _, layout := range []string{"01-02", "1-2"} {
		if t, err := time.Parse(layout, normalized); err == nil {
			candidate := time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
			if candidate.After(now) {
				candidate = candidate.AddDate(-1, 0, 0)
			}
			return candidate.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("use YYYY-MM-DD, MM-DD, t, y, or -N")
}
