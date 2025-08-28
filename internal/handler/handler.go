package handler

import (
	"context"
	"encoding/json"
	"lang-tracker/internal/models"
	"lang-tracker/internal/service"

	"github.com/aws/aws-lambda-go/events"
)

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var request models.Request
	if err := json.Unmarshal([]byte(req.Body), &request); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request"}, nil
	}

	switch request.Action {
	case "log":
		err := service.LogActivity(ctx, request)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
		}
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: `{"message":"Log saved"}`}, nil

	case "stats":
		stats, err := service.GetStats(ctx, request.UserID, request.Language)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
		}
		resp, _ := json.Marshal(stats)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(resp)}, nil

	default:
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Unknown action"}, nil
	}
}

// ToAPIGatewayRequest converts a Request into API Gateway Proxy Request
func ToAPIGatewayRequest(r models.Request) events.APIGatewayProxyRequest {
	b, _ := json.Marshal(r)
	return events.APIGatewayProxyRequest{Body: string(b)}
}
