package service

import (
	"context"
	"fmt"
	"lang-tracker/internal/db"
	"lang-tracker/internal/models"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

// StatsService handles computing usage statistics.
type StatsService struct {
	DB        db.DynamoDBClient
	TableName string
}

// GetStats returns aggregated stats for a user+language combination.
func (s *StatsService) GetStats(ctx context.Context, userID, language string) (*models.StatsResponse, error) {
	items, err := db.QueryByUserId(ctx, s.DB, userID, s.TableName)
	if err != nil {
		return nil, err
	}

	var logs []models.LogItem
	if err := attributevalue.UnmarshalListOfMaps(items, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal log items: %w", err)
	}

	slog.DebugContext(ctx, "fetched log items", "count", len(logs), "userId", userID, "language", language)

	// Accepted date formats
	formats := []string{"2006-01-02", "01/02/2006"}

	now := time.Now()
	nowYear, nowMonth, nowDay := now.Date()
	nowYear2, nowWeek := now.ISOWeek()

	var (
		total     int
		todayMins int
		weekMins  int
		monthMins int
		actTotals = make(map[string]int)
	)

	for _, entry := range logs {
		// Filter by language first before any computation
		if entry.Language != language {
			continue
		}

		// Parse date — try each format, fresh error per iteration
		var (
			dt       time.Time
			parseErr error
		)
		for _, format := range formats {
			dt, parseErr = time.Parse(format, entry.Date)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			slog.WarnContext(ctx, "skipping entry with unparseable date",
				"logId", entry.LogID, "date", entry.Date)
			continue
		}

		entryYear, entryMonth, entryDay := dt.Date()
		entryYear2, entryWeek := dt.ISOWeek()

		total += entry.Minutes
		actTotals[entry.ActivityType] += entry.Minutes

		// Today: must match year, month, and day
		if entryYear == nowYear && entryMonth == nowMonth && entryDay == nowDay {
			todayMins += entry.Minutes
		}

		// This ISO week
		if entryYear2 == nowYear2 && entryWeek == nowWeek {
			weekMins += entry.Minutes
		}

		// This calendar month
		if entryYear == nowYear && entryMonth == nowMonth {
			monthMins += entry.Minutes
		}
	}

	// Percentages: each activity's share of total logged time for this language.
	// Returns 0 for each activity if no time has been logged yet.
	percentages := make(map[string]float64)
	if total > 0 {
		for act, mins := range actTotals {
			percentages[act] = (float64(mins) / float64(total)) * 100
		}
	}

	return &models.StatsResponse{
		TotalHours:  float64(total) / 60.0,
		Today:       float64(todayMins) / 60.0,
		ThisWeek:    float64(weekMins) / 60.0,
		ThisMonth:   float64(monthMins) / 60.0,
		Percentages: percentages,
	}, nil
}
