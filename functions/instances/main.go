package main

import (
    "fmt"
    "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"
	"net/http"
)


func listInstances() (events.APIGatewayProxyResponse, error) {
	fmt.Println("list instances\n")
	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: http.StatusOK}, nil
}

func getInstance(id string)(events.APIGatewayProxyResponse, error) {
	fmt.Println("get instance\n")
	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: http.StatusOK}, nil
}

func handleGet(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//Is there an id from the path?
	var id string
	if len(request.PathParameters) > 0 {
		id = request.PathParameters["id"]
	}

	switch id {
	case "":
		return listInstances()
	default:
		return getInstance(id)
	}

}



func handleOtherRequests(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: http.StatusNotImplemented}, nil
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	var response events.APIGatewayProxyResponse
	var err error

	method := strings.ToLower(request.HTTPMethod)
	switch method {
	case "get":
		response, err = handleGet(request)
	default:
		response, err = handleOtherRequests(request)
	}

	return response, err
}

func main() {
	lambda.Start(Handler)
}