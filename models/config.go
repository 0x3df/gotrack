package models

import "github.com/google/uuid"

type TrackerType string

const (
	TrackerBinary   TrackerType = "binary"
	TrackerDuration TrackerType = "duration" // stored as float64 minutes
	TrackerCount    TrackerType = "count"    // stored as float64
	TrackerNumeric  TrackerType = "numeric"  // stored as float64
	TrackerRating   TrackerType = "rating"   // stored as float64, 1-5
	TrackerText     TrackerType = "text"     // stored as string
)

type Tracker struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Type   TrackerType `json:"type"`
	Target float64     `json:"target,omitempty"` // 0 = no target
	Order  int         `json:"order"`
}

func NewTracker(name string, t TrackerType) Tracker {
	return Tracker{
		ID:   uuid.New().String(),
		Name: name,
		Type: t,
	}
}

type Category struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Color    string    `json:"color"` // lipgloss color string
	Order    int       `json:"order"`
	Trackers []Tracker `json:"trackers"`
}

func NewCategory(name, color string) Category {
	return Category{
		ID:    uuid.New().String(),
		Name:  name,
		Color: color,
	}
}

type Config struct {
	SetupComplete bool       `json:"setup_complete"`
	Categories    []Category `json:"categories"`
}

type Entry struct {
	Date string                 `json:"date"`
	Data map[string]interface{} `json:"data"`
}
