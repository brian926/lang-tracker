package service

import (
	"context"
	"lang-tracker/internal/db"
	"lang-tracker/internal/models"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func GetStats(ctx context.Context, userID, language string) (*models.StatsResponse, error) {
	out, err := db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("LanguageLogs"),
		KeyConditionExpression: aws.String("userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, err
	}

	var logs []models.LogItem
	err = attributevalue.UnmarshalListOfMaps(out.Items, &logs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	year, week := now.ISOWeek()

	total := 0
	todayMins := 0
	weekMins := 0
	monthMins := 0
	activityTotals := make(map[string]int)

	for _, log := range logs {
		if log.Language != language {
			continue
		}
		total += log.Minutes
		activityTotals[log.ActivityType] += log.Minutes

		dt, _ := time.Parse("2006-01-02", log.Date)
		if log.Date == today {
			todayMins += log.Minutes
		}
		y, w := dt.ISOWeek()
		if y == year && w == week {
			weekMins += log.Minutes
		}
		if dt.Year() == now.Year() && dt.Month() == now.Month() {
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
