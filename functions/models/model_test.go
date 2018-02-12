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
	"github.com/d-smith/statusapi-sls/model"
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

func (m *dynamoDBMockery) GetItem(input *dynamodb.GetItemInput)(*dynamodb.GetItemOutput, error) {
	if  *input.Key["name"].S != "model1" {
		return nil, awserr.New(dynamodb.ErrCodeResourceNotFoundException, "nope", errors.New("whoops"))
	}

	out := &dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"name": {
			S: aws.String("model1"),
			},
			"steps": {
				SS: []*string{aws.String("s1"), aws.String("s2")},
			},
		},
	}

	return out, nil
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
			events.APIGatewayProxyRequest{HTTPMethod: "POST", Body: `{"steps": ["Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			200,
			nil,
		},
		{
			"handle existing model",
			events.APIGatewayProxyRequest{HTTPMethod: "POST", Body: `{"steps": ["Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			400,
			nil,
		},
		{
			"handle malformed request",
			events.APIGatewayProxyRequest{HTTPMethod: "POST", Body: `{"steps":Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "x"}`},
			400,
			nil,
		},
	}

	var awsContext awsctx.AWSContext
	var myMock dynamoDBMockery
	awsContext.DynamoDBSvc = &myMock

	handler := makeHandler(&awsContext)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := handler(test.request)
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
			events.APIGatewayProxyRequest{HTTPMethod: "GET"},
			&dynamodb.ScanOutput{},
			200,
			nil,
			nil,
		},
		{
			"one results",
			events.APIGatewayProxyRequest{HTTPMethod: "GET"},
			makeOutputWithModelName("model1"),
			200,
			[]string{"model1"},
			nil,
		},
		{
			"two results",
			events.APIGatewayProxyRequest{HTTPMethod: "GET"},
			makeOutputWithModelName("model1", "model2"),
			200,
			[]string{"model1","model2"},
			nil,
		},
	}

	var awsContext awsctx.AWSContext
	var myMock dynamoDBMockery
	awsContext.DynamoDBSvc = &myMock

	handler := makeHandler(&awsContext)

	for _, test := range tests {
		myMock.scanresult = test.scanResult
		t.Run(test.name, func(t *testing.T) {
			response, err := handler(test.request)

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


func TestModelGet(t *testing.T) {
	tests := []struct {
		name    string
		request events.APIGatewayProxyRequest
		scanResult *dynamodb.ScanOutput
		expect  int
		err     error
	}{
		{
			"model1",
			events.APIGatewayProxyRequest{HTTPMethod: "GET", PathParameters:map[string]string{"name":"model1"}},
			&dynamodb.ScanOutput{},
			200,
			nil,
		},
	}

	var awsContext awsctx.AWSContext
	var myMock dynamoDBMockery
	awsContext.DynamoDBSvc = &myMock

	handler := makeHandler(&awsContext)

	for _, test := range tests {
		response, err := handler(test.request)

		assert.IsType(t, test.err, err)
		assert.Equal(t, test.expect, response.StatusCode)

		var output model.Model
		err = json.Unmarshal([]byte(response.Body),&output)
		if assert.Nil(t,err) {
			assert.Equal(t, "model1", output.Name)
		}
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
