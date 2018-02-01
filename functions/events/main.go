package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type AWSContext struct {
	ddbSvc dynamodbiface.DynamoDBAPI
}

var (
	eventAPI = NewEventSvc()
)

func checkInputs(event *StatusEvent) error {
	if event.State == "" || event.CorrelationId == "" || event.EventId == "" {
		return errors.New("Event payload missing mandatory fields")
	}

	return nil
}

func processRequest(awsContext *AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	var event StatusEvent

	err := json.Unmarshal([]byte(request.Body), &event)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	err = checkInputs(&event)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400}, nil
	}

	err = eventAPI.StoreEvent(awsContext, &event)

	fmt.Printf("event %s processed\n", event.EventId)

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func makeHandler(awsContext *AWSContext) func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return processRequest(awsContext, request)
	}
}

func main() {
	var awsContext AWSContext

	sess := session.New()
	svc := dynamodb.New(sess)

	awsContext.ddbSvc = svc

	handler := makeHandler(&awsContext)
	lambda.Start(handler)
}
