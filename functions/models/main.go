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


func listModels() (events.APIGatewayProxyResponse, error) {
	fmt.Println("listModels")
	return events.APIGatewayProxyResponse{Body: "models get\n", StatusCode: 200}, nil
}

func getModel(name string)(events.APIGatewayProxyResponse, error) {
	fmt.Println("getModel", name)
	return events.APIGatewayProxyResponse{Body: "models get\n", StatusCode: 200}, nil
}

func handleGet(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//Is there an id from the path?
	var name string
	if len(request.PathParameters) > 0 {
		name = request.PathParameters["name"]
	}

	switch name {
	case "":
		return listModels()
	default:
		return getModel(name)
	}

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
	var name string
	if len(request.PathParameters) > 0 {
		name = request.PathParameters["name"]
	}

	fmt.Println("update model", name)
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