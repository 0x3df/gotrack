package mcpconn

import (
	"errors"

	"dailytrack/db"
	"dailytrack/models"
)

var errUnsupportedTrackerType = errors.New("insights not available for this tracker type; use gotrack_entries")

// InsightsPayload is returned by gotrack_insights (JSON fields use snake_case for LLMs).
type InsightsPayload struct {
	TrackerID   string `json:"tracker_id"`
	TrackerName string `json:"tracker_name"`
	TrackerType string `json:"tracker_type"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	SampleSize  int    `json:"sample_size"`

	Text *struct {
		Latest string `json:"latest,omitempty"`
	} `json:"text,omitempty"`

	Binary *struct {
		Done           int     `json:"done"`
		Total          int     `json:"total"`
		ConsistencyPct float64 `json:"consistency_pct"`
		CurrentStreak  int     `json:"current_streak"`
	} `json:"binary,omitempty"`

	Numeric *struct {
		SeriesOldestFirst []float64 `json:"series_oldest_first"`
		RecentAvg         *float64  `json:"recent_avg,omitempty"`
		PrevAvg           *float64  `json:"prev_avg,omitempty"`
		MomentumDelta     *float64  `json:"momentum_delta,omitempty"`
		MomentumOK        bool      `json:"momentum_ok"`
		TargetHits        *int      `json:"target_hits,omitempty"`
		TargetTotal       *int      `json:"target_total,omitempty"`
		TargetPct         *float64  `json:"target_hit_pct,omitempty"`
	} `json:"numeric,omitempty"`
}

func buildInsights(cfg *models.Config, entries []models.Entry, trackerKey string, window, tail int) (*InsightsPayload, error) {
	t, err := db.LookupTracker(cfg, trackerKey)
	if err != nil {
		return nil, err
	}
	if window <= 0 {
		window = 7
	}
	if tail <= 0 {
		tail = 60
	}

	out := &InsightsPayload{
		TrackerID:   t.ID,
		TrackerName: t.Name,
		TrackerType: string(t.Type),
		SampleSize:  len(entries),
	}

	switch t.Type {
	case models.TrackerBinary:
		done, total := db.BinaryStats(entries, t.ID)
		out.Binary = &struct {
			Done           int     `json:"done"`
			Total          int     `json:"total"`
			ConsistencyPct float64 `json:"consistency_pct"`
			CurrentStreak  int     `json:"current_streak"`
		}{
			Done:           done,
			Total:          total,
			ConsistencyPct: db.ConsistencyPct(entries, t.ID),
			CurrentStreak:  db.CurrentStreak(entries, t.ID),
		}
		return out, nil

	case models.TrackerDuration, models.TrackerCount, models.TrackerNumeric, models.TrackerRating:
		series := db.NumericSeries(entries, t.ID)
		if len(series) == 0 {
			out.Numeric = &struct {
				SeriesOldestFirst []float64 `json:"series_oldest_first"`
				RecentAvg         *float64  `json:"recent_avg,omitempty"`
				PrevAvg           *float64  `json:"prev_avg,omitempty"`
				MomentumDelta     *float64  `json:"momentum_delta,omitempty"`
				MomentumOK        bool      `json:"momentum_ok"`
				TargetHits        *int      `json:"target_hits,omitempty"`
				TargetTotal       *int      `json:"target_total,omitempty"`
				TargetPct         *float64  `json:"target_hit_pct,omitempty"`
			}{
				SeriesOldestFirst: []float64{},
				MomentumOK:        false,
			}
			return out, nil
		}
		start := 0
		if len(series) > tail {
			start = len(series) - tail
		}
		tailSeries := append([]float64(nil), series[start:]...)

		nBlock := struct {
			SeriesOldestFirst []float64 `json:"series_oldest_first"`
			RecentAvg         *float64  `json:"recent_avg,omitempty"`
			PrevAvg           *float64  `json:"prev_avg,omitempty"`
			MomentumDelta     *float64  `json:"momentum_delta,omitempty"`
			MomentumOK        bool      `json:"momentum_ok"`
			TargetHits        *int      `json:"target_hits,omitempty"`
			TargetTotal       *int      `json:"target_total,omitempty"`
			TargetPct         *float64  `json:"target_hit_pct,omitempty"`
		}{
			SeriesOldestFirst: tailSeries,
		}

		if ra, pa, d, ok := db.TrackerMomentum(entries, t.ID, window); ok {
			nBlock.MomentumOK = true
			nBlock.RecentAvg = &ra
			nBlock.PrevAvg = &pa
			nBlock.MomentumDelta = &d
		}

		if t.Target != nil {
			h, tot, pct := db.TargetHitRate(entries, t.ID, *t.Target, 0)
			nBlock.TargetHits = &h
			nBlock.TargetTotal = &tot
			nBlock.TargetPct = &pct
		}

		out.Numeric = &nBlock
		return out, nil

	case models.TrackerText:
		var latest string
		for _, e := range entries {
			if v, ok := e.Data[t.ID].(string); ok && v != "" {
				latest = v
				break
			}
		}
		out.Text = &struct {
			Latest string `json:"latest,omitempty"`
		}{Latest: latest}
		return out, nil

	default:
		return nil, errUnsupportedTrackerType
	}
}
