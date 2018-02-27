package event

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/d-smith/statusapi-sls/awsctx"
	"log"
	"os"
	"strconv"
	"time"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type StatusEvent struct {
	TransactionId string `json:"txn_id"`
	TimeStamp     int64  `json:"timestamp"`
	EventId       string `json:"event_id"`
	Step          string `json:"step"`
	StepState     string `json:"step_state"`
}

var (
	instanceTable = os.Getenv("INSTANCE_TABLE")
)

type EventSvc struct{}

func NewEventSvc() *EventSvc {
	return &EventSvc{}
}

func (es *EventSvc) GetStatusEventsForTxn(awsContext *awsctx.AWSContext, txnId string) (map[string]StatusEvent, map[string]StatusEvent, error) {

	log.Println("EventSvc GetStatusEventsForTxn")
	keyCond := expression.Key("transactionId").Equal(expression.Value(txnId))

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()
	if err != nil {
		log.Println("Error building expression", err.Error())
		return nil, nil, err
	}


	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression: expr.KeyCondition(),
		TableName:              aws.String(instanceTable),
	}

	qout, err := awsContext.DynamoDBSvc.Query(input)
	if err != nil {
		log.Println("Error executing DynamoDB query")
		return nil, nil, err
	}

	activeEvents := make(map[string]StatusEvent)
	completedEvents := make(map[string]StatusEvent)

	items := qout.Items
	for _, item := range items {

		step := *item["step"].S
		stepState := *item["step_state"].S

		event := StatusEvent{
			TransactionId: txnId,
			EventId:       *item["eventId"].S,
			Step:          step,
			StepState:     stepState,
		}

		switch stepState {
		case "active":
			activeEvents[step] = event
		case "completed":
			completedEvents[step] = event
		default:
			fmt.Println("Warning: unrecognized step state", stepState)
		}
	}

	return activeEvents, completedEvents, nil

}

func (es *EventSvc) GetEventsForTxn(awsContext *awsctx.AWSContext, txnId string) ([]StatusEvent, error) {
	keyCond := expression.Key("transactionId").Equal(expression.Value(txnId))

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()
	if err != nil {
		log.Println("Error building expression", err.Error())
		return nil, err
	}

	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression: expr.KeyCondition(),
		TableName:              aws.String(instanceTable),
	}

	qout, err := awsContext.DynamoDBSvc.Query(input)
	if err != nil {
		return nil, err
	}

	var events []StatusEvent
	items := qout.Items
	for _, item := range items {

		ts, convertErr := strconv.ParseInt(*item["eventTimestamp"].N, 10, 64)
		if convertErr != nil {
			log.Printf("WARNING: Error converting timestamp from ddb string to int: %s", convertErr)
		}

		event := StatusEvent{
			TransactionId: *item["transactionId"].S,
			TimeStamp:     ts,
			EventId:       *item["eventId"].S,
			Step:          *item["step"].S,
			StepState:     *item["step_state"].S,
		}

		events = append(events, event)
	}

	return events, err
}

func (es *EventSvc) StoreEvent(awsContext *awsctx.AWSContext, event *StatusEvent) error {
	now := time.Now()
	timestampMillis := now.UnixNano() / 1000000
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"transactionId": {
				S: aws.String(event.TransactionId),
			},
			"eventTimestamp": {
				N: aws.String(fmt.Sprintf("%d", timestampMillis)),
			},
			"eventId": {
				S: aws.String(event.EventId),
			},
			"step": {
				S: aws.String(event.Step),
			},
			"step_state": {
				S: aws.String(event.StepState),
			},
		},
		TableName: aws.String(instanceTable),
	}
	_, err := awsContext.DynamoDBSvc.PutItem(input)

	return err
}
