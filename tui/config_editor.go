package tui

import (
	"fmt"

	"dailytrack/models"
)

func trackerHasData(entries []models.Entry, trackerID string) bool {
	for _, entry := range entries {
		if _, ok := entry.Data[trackerID]; ok {
			return true
		}
	}
	return false
}

func categoryHasData(entries []models.Entry, cat models.Category) bool {
	for _, tracker := range cat.Trackers {
		if trackerHasData(entries, tracker.ID) {
			return true
		}
	}
	return false
}

func moveCategory(cfg *models.Config, categoryID string, delta int) bool {
	if cfg == nil {
		return false
	}
	for idx := range cfg.Categories {
		if cfg.Categories[idx].ID != categoryID {
			continue
		}
		next := idx + delta
		if next < 0 || next >= len(cfg.Categories) {
			return false
		}
		cfg.Categories[idx], cfg.Categories[next] = cfg.Categories[next], cfg.Categories[idx]
		refreshCategoryOrders(cfg)
		return true
	}
	return false
}

func moveTracker(cfg *models.Config, categoryID, trackerID string, delta int) bool {
	cat := findCategory(cfg, categoryID)
	if cat == nil {
		return false
	}
	for idx := range cat.Trackers {
		if cat.Trackers[idx].ID != trackerID {
			continue
		}
		next := idx + delta
		if next < 0 || next >= len(cat.Trackers) {
			return false
		}
		cat.Trackers[idx], cat.Trackers[next] = cat.Trackers[next], cat.Trackers[idx]
		refreshTrackerOrders(cat)
		return true
	}
	return false
}

func deleteCategory(cfg *models.Config, entries []models.Entry, categoryID string) error {
	if cfg == nil {
		return fmt.Errorf("missing config")
	}
	for idx := range cfg.Categories {
		if cfg.Categories[idx].ID != categoryID {
			continue
		}
		if categoryHasData(entries, cfg.Categories[idx]) {
			return fmt.Errorf("category has historical data")
		}
		cfg.Categories = append(cfg.Categories[:idx], cfg.Categories[idx+1:]...)
		refreshCategoryOrders(cfg)
		return nil
	}
	return fmt.Errorf("category not found")
}

func deleteTracker(cfg *models.Config, entries []models.Entry, categoryID, trackerID string) error {
	cat := findCategory(cfg, categoryID)
	if cat == nil {
		return fmt.Errorf("category not found")
	}
	for idx := range cat.Trackers {
		if cat.Trackers[idx].ID != trackerID {
			continue
		}
		if trackerHasData(entries, trackerID) {
			return fmt.Errorf("tracker has historical data")
		}
		cat.Trackers = append(cat.Trackers[:idx], cat.Trackers[idx+1:]...)
		refreshTrackerOrders(cat)
		return nil
	}
	return fmt.Errorf("tracker not found")
}

func findCategory(cfg *models.Config, categoryID string) *models.Category {
	if cfg == nil {
		return nil
	}
	for idx := range cfg.Categories {
		if cfg.Categories[idx].ID == categoryID {
			return &cfg.Categories[idx]
		}
	}
	return nil
}

func findTracker(cat *models.Category, trackerID string) *models.Tracker {
	if cat == nil {
		return nil
	}
	for idx := range cat.Trackers {
		if cat.Trackers[idx].ID == trackerID {
			return &cat.Trackers[idx]
		}
	}
	return nil
}

func refreshCategoryOrders(cfg *models.Config) {
	for idx := range cfg.Categories {
		cfg.Categories[idx].Order = idx
		refreshTrackerOrders(&cfg.Categories[idx])
	}
}

func refreshTrackerOrders(cat *models.Category) {
	for idx := range cat.Trackers {
		cat.Trackers[idx].Order = idx
	}
}
