package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/d-smith/statusapi-sls/awsctx"
	"github.com/d-smith/statusapi-sls/event"
)

var (
	eventAPI = event.NewEventSvc()
)

func checkInputs(event *event.StatusEvent) error {
	step_state := event.StepState

	if event.Step == "" || event.TransactionId == "" || event.EventId == "" || step_state == "" {
		return errors.New("Event payload missing mandatory fields")
	}

	if step_state != "active" && step_state != "completed" {
		return errors.New("Invalid step state")
	}

	return nil
}

func processRequest(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	var event event.StatusEvent

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

func makeHandler(awsContext *awsctx.AWSContext) func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return processRequest(awsContext, request)
	}
}

func main() {
	var awsContext awsctx.AWSContext

	sess := session.New()
	svc := dynamodb.New(sess)

	awsContext.DynamoDBSvc = svc

	handler := makeHandler(&awsContext)
	lambda.Start(handler)
}
