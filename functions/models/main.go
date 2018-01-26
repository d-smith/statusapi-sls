package main

import (
    	"fmt"
    	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"
	"encoding/json"
)

var (
	modelAPI = NewModel()
)

func handleGet(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: "models get\n", StatusCode: 200}, nil
}

func handlePost(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var model Model
	err := json.Unmarshal([]byte(request.Body), &model)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400 }, nil
	}

	err = modelAPI.CreateModel(&model)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500 }, nil
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func handlePut(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: "models put\n", StatusCode: 200}, nil
}

func handleOtherRequests(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: 200}, nil
}


func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	var response events.APIGatewayProxyResponse
	var err error

	method := strings.ToLower(request.HTTPMethod)
	switch method {
	case "get":
		response, err = handleGet(request)
	case "post":
		response, err = handlePost(request)
	case "put":
		response, err = handlePut(request)
	default:
		response, err = handleOtherRequests(request)
	}

	return response, err
}

func main() {
	lambda.Start(Handler)
}