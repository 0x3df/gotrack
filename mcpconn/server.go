package mcpconn

import (
	"context"
	"fmt"
	"time"

	"dailytrack/db"
	"dailytrack/models"

	mcplib "github.com/modelcontextprotocol/go-sdk/mcp"
)

const mcpInstructions = `This server exposes your local GoTrack workspace (SQLite + config).
Workflow: call gotrack_schema to list categories and trackers (names, UUIDs, types, units).
Use gotrack_entries for raw day rows in a date range. Use gotrack_insights for per-tracker
stats (series tail, momentum, binary streaks, text latest). Use gotrack_log to upsert fields
for one date (same rules as the gotrack log CLI). Avoid running heavy exports while the TUI
is open; SQLite is single-writer.`

// Run starts the MCP server on stdio. Callers should ensure logging does not write to stdout.
func Run(ctx context.Context) error {
	if err := db.InitDB(); err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	s := mcplib.NewServer(&mcplib.Implementation{Name: "gotrack", Version: "1.0"}, &mcplib.ServerOptions{
		Instructions: mcpInstructions,
	})

	mcplib.AddTool(s, &mcplib.Tool{
		Name:        "gotrack_schema",
		Description: "Return the full GoTrack config: categories, trackers (id, name, type, unit, targets), and setup flag.",
	}, handleSchema)

	mcplib.AddTool(s, &mcplib.Tool{
		Name:        "gotrack_entries",
		Description: "List daily entries between two ISO dates (inclusive), newest first. Defaults: last 60 days through today. Caps result count.",
	}, handleEntries)

	mcplib.AddTool(s, &mcplib.Tool{
		Name:        "gotrack_insights",
		Description: "Structured stats for one tracker: numeric/rating/duration/count series tail, momentum vs prior window, target hit rate; binary streak/consistency; text latest. Optional date filter on entry rows.",
	}, handleInsights)

	mcplib.AddTool(s, &mcplib.Tool{
		Name:        "gotrack_log",
		Description: "Upsert tracker values for one date (merges with existing row). Keys are tracker names or UUIDs; values are JSON scalars/strings matching tracker types.",
	}, handleLog)

	return s.Run(ctx, &mcplib.StdioTransport{})
}

func requireSetup(cfg *models.Config) error {
	if cfg == nil || !cfg.SetupComplete {
		return fmt.Errorf("gotrack is not set up yet — run `gotrack` once to create a workspace")
	}
	return nil
}

type schemaArgs struct{}

type schemaOut struct {
	SetupComplete bool           `json:"setup_complete"`
	Config        *models.Config `json:"config"`
}

func handleSchema(_ context.Context, _ *mcplib.CallToolRequest, _ schemaArgs) (*mcplib.CallToolResult, schemaOut, error) {
	cfg, err := db.LoadConfig()
	if err != nil {
		return nil, schemaOut{}, fmt.Errorf("load config: %w", err)
	}
	if cfg == nil {
		return nil, schemaOut{SetupComplete: false}, nil
	}
	return nil, schemaOut{SetupComplete: cfg.SetupComplete, Config: cfg}, nil
}

type entriesArgs struct {
	From       string `json:"from,omitempty" jsonschema:"inclusive YYYY-MM-DD start; default ~60 days before 'to'"`
	To         string `json:"to,omitempty" jsonschema:"inclusive YYYY-MM-DD end; default today"`
	MaxEntries int    `json:"max_entries,omitempty" jsonschema:"max rows (default 500, hard cap 2000)"`
}

type entriesOut struct {
	From      string         `json:"from"`
	To        string         `json:"to"`
	Truncated bool           `json:"truncated"`
	Count     int            `json:"count"`
	Entries   []models.Entry `json:"entries"`
}

func handleEntries(_ context.Context, _ *mcplib.CallToolRequest, in entriesArgs) (*mcplib.CallToolResult, entriesOut, error) {
	cfg, err := db.LoadConfig()
	if err != nil {
		return nil, entriesOut{}, fmt.Errorf("load config: %w", err)
	}
	if err := requireSetup(cfg); err != nil {
		return nil, entriesOut{}, err
	}

	from, to, err := resolveEntryDateRange(in.From, in.To)
	if err != nil {
		return nil, entriesOut{}, err
	}

	max := in.MaxEntries
	if max <= 0 {
		max = 500
	}
	if max > 2000 {
		max = 2000
	}

	entries, err := db.GetEntriesBetween(from, to)
	if err != nil {
		return nil, entriesOut{}, err
	}
	truncated := false
	if len(entries) > max {
		entries = entries[:max]
		truncated = true
	}

	return nil, entriesOut{
		From:      from,
		To:        to,
		Truncated: truncated,
		Count:     len(entries),
		Entries:   entries,
	}, nil
}

func resolveEntryDateRange(from, to string) (string, string, error) {
	today := time.Now().Format("2006-01-02")
	var err error
	if to == "" {
		to = today
	} else if to, err = db.NormalizeDate(to); err != nil {
		return "", "", fmt.Errorf("to: %w", err)
	}
	if from == "" {
		end, perr := time.Parse("2006-01-02", to)
		if perr != nil {
			return "", "", fmt.Errorf("parse to: %w", perr)
		}
		from = end.AddDate(0, 0, -60).Format("2006-01-02")
	} else if from, err = db.NormalizeDate(from); err != nil {
		return "", "", fmt.Errorf("from: %w", err)
	}
	if from > to {
		return "", "", fmt.Errorf("from must be on or before to")
	}
	return from, to, nil
}

type insightsArgs struct {
	Tracker string `json:"tracker" jsonschema:"tracker name or UUID"`
	From    string `json:"from,omitempty" jsonschema:"optional inclusive YYYY-MM-DD; omit with to for all-time"`
	To      string `json:"to,omitempty" jsonschema:"optional inclusive YYYY-MM-DD"`
	Window  int    `json:"window,omitempty" jsonschema:"momentum window (default 7)"`
	Tail    int    `json:"tail,omitempty" jsonschema:"max series points (default 60)"`
}

type insightsOut struct {
	Insights *InsightsPayload `json:"insights"`
}

func handleInsights(_ context.Context, _ *mcplib.CallToolRequest, in insightsArgs) (*mcplib.CallToolResult, insightsOut, error) {
	cfg, err := db.LoadConfig()
	if err != nil {
		return nil, insightsOut{}, fmt.Errorf("load config: %w", err)
	}
	if err := requireSetup(cfg); err != nil {
		return nil, insightsOut{}, err
	}
	if in.Tracker == "" {
		return nil, insightsOut{}, fmt.Errorf("tracker is required")
	}

	var entries []models.Entry
	var fromOut, toOut string
	switch {
	case in.From == "" && in.To == "":
		entries, err = db.GetAllEntries()
		if err != nil {
			return nil, insightsOut{}, err
		}
	default:
		from, to, rerr := resolveEntryDateRange(in.From, in.To)
		if rerr != nil {
			return nil, insightsOut{}, rerr
		}
		fromOut, toOut = from, to
		entries, err = db.GetEntriesBetween(from, to)
		if err != nil {
			return nil, insightsOut{}, err
		}
	}

	pl, err := buildInsights(cfg, entries, in.Tracker, in.Window, in.Tail)
	if err != nil {
		return nil, insightsOut{}, err
	}
	pl.From = fromOut
	pl.To = toOut
	return nil, insightsOut{Insights: pl}, nil
}

type logArgs struct {
	Date   string                 `json:"date,omitempty" jsonschema:"YYYY-MM-DD or today/yesterday/-N (default today)"`
	Values map[string]interface{} `json:"values" jsonschema:"tracker name or id -> JSON value"`
}

type logOut struct {
	Date       string `json:"date"`
	FieldCount int    `json:"field_count"`
	Message    string `json:"message"`
}

func handleLog(_ context.Context, _ *mcplib.CallToolRequest, in logArgs) (*mcplib.CallToolResult, logOut, error) {
	cfg, err := db.LoadConfig()
	if err != nil {
		return nil, logOut{}, fmt.Errorf("load config: %w", err)
	}
	if err := requireSetup(cfg); err != nil {
		return nil, logOut{}, err
	}
	if len(in.Values) == 0 {
		return nil, logOut{}, fmt.Errorf("values must be a non-empty object")
	}
	date := in.Date
	if date == "" {
		date = "today"
	}
	date, err = db.NormalizeDate(date)
	if err != nil {
		return nil, logOut{}, fmt.Errorf("date: %w", err)
	}
	if err := db.UpsertEntryLog(cfg, date, in.Values); err != nil {
		return nil, logOut{}, err
	}
	after, err := db.GetEntryForDate(date)
	if err != nil {
		return nil, logOut{}, err
	}
	n := 0
	if after != nil && after.Data != nil {
		n = len(after.Data)
	}
	return nil, logOut{
		Date:       date,
		FieldCount: n,
		Message:    "saved",
	}, nil
}
