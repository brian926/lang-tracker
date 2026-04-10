package handler

import (
	"context"
	"encoding/json"
	"errors"
	"lang-tracker/internal/models"
	"lang-tracker/internal/service"
	"log/slog"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Services groups the dependencies the handler needs.
type Services struct {
	Log   *service.LogService
	Stats *service.StatsService
}

var jsonHeaders = map[string]string{"Content-Type": "application/json"}

// Handler is the Lambda / HTTP entry point.
func (s *Services) Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var request models.Request
	if err := json.Unmarshal([]byte(req.Body), &request); err != nil {
		return errResponse(400, "invalid JSON body"), nil
	}

	switch request.Action {
	case "log":
		if err := validateLogRequest(request); err != nil {
			return errResponse(400, err.Error()), nil
		}
		if err := s.Log.LogActivity(ctx, request); err != nil {
			slog.ErrorContext(ctx, "LogActivity failed", "err", err)
			return errResponse(500, "internal server error"), nil
		}
		return jsonResponse(200, map[string]string{"message": "Log saved"}), nil

	case "stats":
		if err := validateStatsRequest(request); err != nil {
			return errResponse(400, err.Error()), nil
		}
		stats, err := s.Stats.GetStats(ctx, request.UserID, request.Language)
		if err != nil {
			slog.ErrorContext(ctx, "GetStats failed", "err", err)
			return errResponse(500, "internal server error"), nil
		}
		return jsonResponse(200, stats), nil

	default:
		return errResponse(400, "unknown action: must be \"log\" or \"stats\""), nil
	}
}

// ToAPIGatewayRequest converts a Request into an APIGatewayProxyRequest.
// Used by the local dev server to bridge plain HTTP into the Lambda handler.
func ToAPIGatewayRequest(r models.Request) events.APIGatewayProxyRequest {
	b, _ := json.Marshal(r)
	return events.APIGatewayProxyRequest{Body: string(b)}
}

// --- validation -------------------------------------------------------

func validateLogRequest(r models.Request) error {
	var errs []string
	if strings.TrimSpace(r.UserID) == "" {
		errs = append(errs, "userId is required")
	}
	if strings.TrimSpace(r.Language) == "" {
		errs = append(errs, "language is required")
	}
	if strings.TrimSpace(r.ActivityType) == "" {
		errs = append(errs, "activityType is required")
	}
	if r.Minutes <= 0 {
		errs = append(errs, "minutes must be a positive integer")
	}
	if r.Minutes > 1440 {
		errs = append(errs, "minutes cannot exceed 1440 (24 hours)")
	}
	if strings.TrimSpace(r.Date) == "" {
		errs = append(errs, "date is required (formats: YYYY-MM-DD or MM/DD/YYYY)")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func validateStatsRequest(r models.Request) error {
	var errs []string
	if strings.TrimSpace(r.UserID) == "" {
		errs = append(errs, "userId is required")
	}
	if strings.TrimSpace(r.Language) == "" {
		errs = append(errs, "language is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// --- response helpers -------------------------------------------------

func jsonResponse(status int, body any) events.APIGatewayProxyResponse {
	b, err := json.Marshal(body)
	if err != nil {
		slog.Error("failed to marshal response", "err", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    jsonHeaders,
			Body:       `{"error":"internal server error"}`,
		}
	}
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    jsonHeaders,
		Body:       string(b),
	}
}

func errResponse(status int, msg string) events.APIGatewayProxyResponse {
	return jsonResponse(status, map[string]string{"error": msg})
}
