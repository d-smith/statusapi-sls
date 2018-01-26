package main

import (
	"testing"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestModelCreate(t *testing.T) {
	tests := []struct {
		name string
		request events.APIGatewayProxyRequest
		expect int
		err error
	}{
		{
			"handle full request",
			events.APIGatewayProxyRequest{Body: `{"states": ["Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			200,
			nil,
		},
		{
			"handle full request",
			events.APIGatewayProxyRequest{Body: `{"states":Order Received", "Assembling Pizza", "Cooking Pizza", "Pizza Ready"], "name": "model1"}`},
			400,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response,err := handlePost(test.request)
			assert.IsType(t, test.err, err)
			assert.Equal(t, test.expect, response.StatusCode)
		})

	}
}