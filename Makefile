compile: models instances events authorizer

models:
	GOOS=linux go build -o bin/models functions/models/*.go

instances:
	GOOS=linux go build -o bin/instances functions/instances/main.go

events:
	GOOS=linux go build -o bin/events functions/events/*.go

authorizer:
	GOOS=linux go build -o bin/authorizer functions/authorizer/*.go
