curl -X POST `
    -H "Content-Type: application/json" `
    -d '{"userId": "1", "language": "French", "activityType": "Watching", "action": "log", "minutes": 60, "date": "10/15/2025"}' `
    http://localhost:8080/api

curl -X POST `
    -H "Content-Type: application/json" `
    -d '{"userId": "1", "language": "French", "action": "stats"}' `
    http://localhost:8080/api