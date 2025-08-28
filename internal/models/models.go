package models

// Incoming request
type Request struct {
	Action       string `json:"action"`
	UserID       string `json:"userId"`
	Language     string `json:"language"`
	ActivityType string `json:"activityType"`
	Minutes      int    `json:"minutes"`
	Date         string `json:"date"`
}

// DynamoDB log item
type LogItem struct {
	UserID       string `dynamodbav:"userId"`
	LogID        string `dynamodbav:"logId"`
	Language     string `dynamodbav:"language"`
	ActivityType string `dynamodbav:"activityType"`
	Minutes      int    `dynamodbav:"minutes"`
	Date         string `dynamodbav:"date"`
}

// Stats response
type StatsResponse struct {
	TotalHours  float64            `json:"totalHours"`
	Today       float64            `json:"today"`
	ThisWeek    float64            `json:"thisWeek"`
	ThisMonth   float64            `json:"thisMonth"`
	Percentages map[string]float64 `json:"percentages"`
}
