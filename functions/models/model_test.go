package main

import (
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/d-smith/statusapi-sls/awsctx"
)

type dynamoDBMockery struct {
	dynamodbiface.DynamoDBAPI
	names map[string]string
}

func (m *dynamoDBMockery) PutItem(item *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	if m.names == nil {
		m.names = make(map[string]string)
	}

	name := *item.Item["name"].S
	_, ok := m.names[name]
	var err error
	if ok == true {
		origErr := errors.New("bam")
		err = awserr.New(dynamodb.ErrCodeConditionalCheckFailedException, "been there done that", origErr)
	} else {
		m.names[name] = name
	}
	var out dynamodb.PutItemOutput
	return &out, err
}

func TestModelCreate(t *testing.T) {
	tests := []struct {
		name    string
		request events.APIGatewayProxyRequest
		expect  int
		err     error
	}{
		{
			"handle full request",
			events.APIGatewayProxyRequest{Body: `{"states": ["Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			200,
			nil,
		},
		{
			"handle existing model",
			events.APIGatewayProxyRequest{Body: `{"states": ["Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			400,
			nil,
		},
		{
			"handle full request",
			events.APIGatewayProxyRequest{Body: `{"states":Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			400,
			nil,
		},
	}

	var awsContext awsctx.AWSContext
	var myMock dynamoDBMockery
	awsContext.DynamoDBSvc = &myMock

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := handlePost(&awsContext, test.request)
			assert.IsType(t, test.err, err)
			assert.Equal(t, test.expect, response.StatusCode)
		})

	}
}
