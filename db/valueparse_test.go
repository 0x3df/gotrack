package db

import (
	"testing"

	"dailytrack/models"
)

func TestCoerceValueFromInterface(t *testing.T) {
	tr := models.NewTracker("Mood", models.TrackerRating)
	v, err := CoerceValueFromInterface(tr, float64(4))
	if err != nil {
		t.Fatal(err)
	}
	if v != float64(4) {
		t.Fatalf("got %v", v)
	}
	_, err = CoerceValueFromInterface(tr, float64(6))
	if err == nil {
		t.Fatal("expected error for rating > 5")
	}
}
