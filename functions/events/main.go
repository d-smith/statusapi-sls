package main

import (
    "fmt"
    "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"encoding/json"
)

type StatusEvent struct {
	CorrelationId string `json:"correlation_id"`
	EventId string `json:"event_id"`
	ModelIds []string `json:"model_ids"`
	State string `json:"state"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	var event StatusEvent

	err := json.Unmarshal([]byte(request.Body), &event)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	fmt.Printf("event %s processed\n", event.EventId)

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}