package main

import (
    "fmt"
    "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"
)

func handleGet(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: "models get\n", StatusCode: 200}, nil
}

func handlePost(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: "models post\n", StatusCode: 200}, nil
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