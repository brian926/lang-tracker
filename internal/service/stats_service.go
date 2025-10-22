package service

import (
	"context"
	"fmt"
	"lang-tracker/internal/db"
	"lang-tracker/internal/models"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

func GetStats(ctx context.Context, userID, language string) (*models.StatsResponse, error) {
	out, err := db.QueryByUserId(ctx, userID, TableName)
	if err != nil {
		return nil, err
	}

	var logs []models.LogItem
	err = attributevalue.UnmarshalListOfMaps(out.Items, &logs)
	if err != nil {
		return nil, err
	}

	// Define possible date formats
	formats := []string{"2006-01-02", "01/02/2006"}

	var dt time.Time

	today := time.Now()
	now := time.Now()
	year, week := now.ISOWeek()
	y1, m1, d1 := today.Date()

	total := 0
	todayMins := 0
	weekMins := 0
	monthMins := 0
	activityTotals := make(map[string]int)

	fmt.Println(logs)

	for _, log := range logs {
		// Try parsing the date with each format
		for _, format := range formats {
			dt, err = time.Parse(format, log.Date)
			if err == nil {
				fmt.Println("Parsed date:", dt)
				break
			}
		}
		if err != nil {
			fmt.Printf("Error parsing date %s: %v\n", log.Date, err)
			continue
		}

		y2, m2, d2 := dt.Date()

		if log.Language != language {
			continue
		}

		total += log.Minutes
		activityTotals[log.ActivityType] += log.Minutes

		// Check for today
		if d1 == d2 {
			todayMins += log.Minutes
		}

		// Check for this week
		y, w := dt.ISOWeek()
		if y == year && w == week {
			weekMins += log.Minutes
		}

		// Check for this month
		if y2 == y1 && m2 == m1 {
			monthMins += log.Minutes
		}
	}

	percentages := make(map[string]float64)
	for act, mins := range activityTotals {
		percentages[act] = (float64(mins) / 60000.0) * 100 // out of 1000 hrs = 60000 mins
	}

	return &models.StatsResponse{
		TotalHours:  float64(total) / 60.0,
		Today:       float64(todayMins) / 60.0,
		ThisWeek:    float64(weekMins) / 60.0,
		ThisMonth:   float64(monthMins) / 60.0,
		Percentages: percentages,
	}, nil
}
