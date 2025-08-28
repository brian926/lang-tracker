package main

import (
	"lang-tracker/internal/db"
	"lang-tracker/internal/handler"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	db.Init()
	lambda.Start(handler.Handler)
}
