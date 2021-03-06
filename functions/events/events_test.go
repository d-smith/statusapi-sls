package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/d-smith/statusapi-sls/awsctx"
	"github.com/stretchr/testify/assert"
	"testing"
)

type dynamoDBMockery struct {
	dynamodbiface.DynamoDBAPI
}

func (m *dynamoDBMockery) PutItem(item *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	var out dynamodb.PutItemOutput
	return &out, nil
}

func TestEventPost(t *testing.T) {

	requestCtx := events.APIGatewayProxyRequestContext{
		Authorizer:map[string]interface{}{
			"tenant":"purple",
		},
	}

	tests := []struct {
		name    string
		request events.APIGatewayProxyRequest
		expect  int
		err     error
	}{
		{
			"handle full request",
			events.APIGatewayProxyRequest{Body: `{"txn_id":"1a","event_id":"1","step":"Order Received","step_state":"active"}`, RequestContext:requestCtx},
			200,
			nil,
		},
		{
			"valid step state - completed",
			events.APIGatewayProxyRequest{Body: `{"txn_id":"1a","event_id":"1","step":"Order Received","step_state":"completed"}`, RequestContext:requestCtx},
			200,
			nil,
		},
		{
			"invalid step state",
			events.APIGatewayProxyRequest{Body: `{"txn_id":"1a","event_id":"1","step":"Order Received","step_state":"dunno"}`, RequestContext:requestCtx},
			400,
			nil,
		},
		{
			"handle wrong payload",
			events.APIGatewayProxyRequest{Body: `{"states":Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`, RequestContext:requestCtx},
			400,
			nil,
		},
	}

	var awsContext awsctx.AWSContext
	var myMock dynamoDBMockery
	awsContext.DynamoDBSvc = &myMock

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := processRequest(&awsContext, test.request)
			assert.IsType(t, test.err, err)
			assert.Equal(t, test.expect, response.StatusCode)
		})

	}
}
