package event

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"os"
	"time"
	"github.com/d-smith/statusapi-sls/awsctx"
)

type StatusEvent struct {
	CorrelationId string `json:"correlation_id"`
	EventId       string `json:"event_id"`
	State         string `json:"state"`
}

var (
	instanceTable = os.Getenv("INSTANCE_TABLE")
)

type EventSvc struct{}

func NewEventSvc() *EventSvc {
	return &EventSvc{}
}

func (es *EventSvc) StoreEvent(awsContext *awsctx.AWSContext, event *StatusEvent) error {
	now := time.Now()
	timestampMillis := now.UnixNano() / 1000000
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"correlationId": {
				S: aws.String(event.CorrelationId),
			},
			"eventTimestamp": {
				N: aws.String(fmt.Sprintf("%d", timestampMillis)),
			},
			"eventId": {
				S: aws.String(event.EventId),
			},
			"state": {
				S: aws.String(event.State),
			},
		},
		TableName: aws.String(instanceTable),
	}
	_, err := awsContext.DynamoDBSvc.PutItem(input)

	return err
}
