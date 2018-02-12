package main

import (
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/d-smith/statusapi-sls/awsctx"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/aws/aws-sdk-go/aws"
	"encoding/json"
	"log"
)

type dynamoDBMockery struct {
	dynamodbiface.DynamoDBAPI
	names      map[string]string
	scanresult *dynamodb.ScanOutput
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

func (m *dynamoDBMockery) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	scanResult := m.scanresult
	if scanResult == nil {
		scanResult = &dynamodb.ScanOutput{}
	}
	return scanResult, nil
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
			events.APIGatewayProxyRequest{Body: `{"steps": ["Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			200,
			nil,
		},
		{
			"handle existing model",
			events.APIGatewayProxyRequest{Body: `{"steps": ["Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			400,
			nil,
		},
		{
			"handle malformed request",
			events.APIGatewayProxyRequest{Body: `{"steps":Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "x"}`},
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

func makeOutputWithModelName(names... string)*dynamodb.ScanOutput {
	scanOutput := &dynamodb.ScanOutput{}
	for _, name := range names {
		itemdata := make(map[string]*dynamodb.AttributeValue)
		itemdata["name"] = &dynamodb.AttributeValue{S:aws.String(name)}
		scanOutput.Items = append(scanOutput.Items, itemdata)
	}

	return scanOutput
}

func TestModelList(t *testing.T) {
	tests := []struct {
		name    string
		request events.APIGatewayProxyRequest
		scanResult *dynamodb.ScanOutput
		expect  int
		payload []string
		err     error
	}{
		{
			"no results",
			events.APIGatewayProxyRequest{},
			&dynamodb.ScanOutput{},
			200,
			nil,
			nil,
		},
		{
			"one results",
			events.APIGatewayProxyRequest{},
			makeOutputWithModelName("model1"),
			200,
			[]string{"model1"},
			nil,
		},
		{
			"two results",
			events.APIGatewayProxyRequest{},
			makeOutputWithModelName("model1", "model2"),
			200,
			[]string{"model1","model2"},
			nil,
		},
	}

	var awsContext awsctx.AWSContext
	var myMock dynamoDBMockery
	awsContext.DynamoDBSvc = &myMock

	for _, test := range tests {
		myMock.scanresult = test.scanResult
		t.Run(test.name, func(t *testing.T) {
			response, err := handleGet(&awsContext, test.request)

			assert.IsType(t, test.err, err)
			assert.Equal(t, test.expect, response.StatusCode)

			var output []string
			err = json.Unmarshal([]byte(response.Body),&output)
			if assert.Nil(t,err) {
				assert.True(t, samePayload(test.payload, output))
			}
		})

	}
}

func samePayload(p1, p2 []string) bool {
	if len(p1) != len(p2) {
		log.Printf("---> Lengths of payloads differ: %v vs %v\n", p1, p2)
		return false
	}

	for i, v1 := range p1 {
		if v1 != p2[i] {
			log.Printf("---> slice content differs: %s vs %s\n", v1, p2[i])
			return false
		}
	}

	return true
}
