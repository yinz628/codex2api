package database

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func newSQLiteTimeRangeTestDB(t *testing.T) *DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "codex2api.db")
	db, err := New("sqlite", dbPath)
	if err != nil {
		t.Fatalf("New(sqlite) returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestSQLiteChartAggregationAcceptsOffsetRange(t *testing.T) {
	db := newSQLiteTimeRangeTestDB(t)
	ctx := context.Background()

	_, err := db.conn.ExecContext(ctx, `
		INSERT INTO usage_logs (
			account_id, endpoint, model, total_tokens, status_code, duration_ms,
			input_tokens, output_tokens, reasoning_tokens, cached_tokens, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, 1, "/v1/responses", "gpt-5.4", 42, 200, 1200, 20, 22, 0, 0, "2026-03-30 13:30:00")
	if err != nil {
		t.Fatalf("insert usage log: %v", err)
	}

	start, err := time.Parse(time.RFC3339, "2026-03-30T21:00:00+08:00")
	if err != nil {
		t.Fatalf("parse start: %v", err)
	}
	end, err := time.Parse(time.RFC3339, "2026-03-30T22:00:00+08:00")
	if err != nil {
		t.Fatalf("parse end: %v", err)
	}

	agg, err := db.GetChartAggregation(ctx, start, end, 5)
	if err != nil {
		t.Fatalf("GetChartAggregation: %v", err)
	}
	if len(agg.Timeline) != 1 {
		t.Fatalf("timeline size = %d, want 1", len(agg.Timeline))
	}
	if agg.Timeline[0].Requests != 1 {
		t.Fatalf("requests = %d, want 1", agg.Timeline[0].Requests)
	}
}

func TestSQLitePagedUsageLogsAcceptOffsetRange(t *testing.T) {
	db := newSQLiteTimeRangeTestDB(t)
	ctx := context.Background()

	_, err := db.conn.ExecContext(ctx, `
		INSERT INTO usage_logs (
			account_id, endpoint, model, total_tokens, status_code, duration_ms,
			input_tokens, output_tokens, reasoning_tokens, cached_tokens, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, 1, "/v1/responses", "gpt-5.4", 42, 200, 1200, 20, 22, 0, 0, "2026-03-30 13:30:00")
	if err != nil {
		t.Fatalf("insert usage log: %v", err)
	}

	start, err := time.Parse(time.RFC3339, "2026-03-30T21:00:00+08:00")
	if err != nil {
		t.Fatalf("parse start: %v", err)
	}
	end, err := time.Parse(time.RFC3339, "2026-03-30T22:00:00+08:00")
	if err != nil {
		t.Fatalf("parse end: %v", err)
	}

	page, err := db.ListUsageLogsByTimeRangePaged(ctx, UsageLogFilter{
		Start:    start,
		End:      end,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListUsageLogsByTimeRangePaged: %v", err)
	}
	if len(page.Logs) != 1 {
		t.Fatalf("page log size = %d, want 1", len(page.Logs))
	}
	if page.Total != 1 {
		t.Fatalf("page total = %d, want 1", page.Total)
	}
}

func TestSQLiteAccountEventTrendAcceptsOffsetRange(t *testing.T) {
	db := newSQLiteTimeRangeTestDB(t)
	ctx := context.Background()

	_, err := db.conn.ExecContext(ctx, `
		INSERT INTO account_events (account_id, event_type, source, created_at)
		VALUES ($1, $2, $3, $4)
	`, 1, "added", "test", "2026-03-30 13:30:00")
	if err != nil {
		t.Fatalf("insert account event: %v", err)
	}

	start, err := time.Parse(time.RFC3339, "2026-03-30T21:00:00+08:00")
	if err != nil {
		t.Fatalf("parse start: %v", err)
	}
	end, err := time.Parse(time.RFC3339, "2026-03-30T22:00:00+08:00")
	if err != nil {
		t.Fatalf("parse end: %v", err)
	}

	trend, err := db.GetAccountEventTrend(ctx, start, end, 60)
	if err != nil {
		t.Fatalf("GetAccountEventTrend: %v", err)
	}
	if len(trend) != 1 {
		t.Fatalf("trend size = %d, want 1", len(trend))
	}
	if trend[0].Added != 1 {
		t.Fatalf("added = %d, want 1", trend[0].Added)
	}
}
