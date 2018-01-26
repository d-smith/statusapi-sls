compile: models instances events

models:
	GOOS=linux go build -o bin/models functions/models/*.go

instances:
	GOOS=linux go build -o bin/instances functions/instances/main.go

events:
	GOOS=linux go build -o bin/events functions/events/main.go
