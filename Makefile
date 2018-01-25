compile:
	GOOS=linux go build -o bin/models functions/models/main.go
	GOOS=linux go build -o bin/instances functions/instances/main.go
	GOOS=linux go build -o bin/events functions/events/main.go
