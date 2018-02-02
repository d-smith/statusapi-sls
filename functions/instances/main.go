package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"strings"
	"github.com/d-smith/statusapi-sls/instance"
	"encoding/json"
	"github.com/d-smith/statusapi-sls/awsctx"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
)

var (
	instanceSvc = instance.NewInstanceSvc()
)

func listInstances() (events.APIGatewayProxyResponse, error) {
	fmt.Println("list instances\n")
	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: http.StatusOK}, nil
}

func getModelStates(awsContext *awsctx.AWSContext, id, model string) (events.APIGatewayProxyResponse, error) {
	log.Println("get model states")
	states, err :=  instanceSvc.StatusForInstance(awsContext, id, model)
	if err != nil {
		fmt.Println("Error building model states", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	bodyOut, err := json.Marshal(&states)
	if err != nil {
		fmt.Println("Error marshalling response", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(bodyOut), StatusCode: http.StatusOK}, nil
}

func getInstance(awsContext *awsctx.AWSContext, id string, queryStringParams map[string]string) (events.APIGatewayProxyResponse, error) {
	fmt.Println("get instance\n")
	model := queryStringParams["model"]
	if model != "" {
		return getModelStates(awsContext, id, model)
	}

	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: http.StatusOK}, nil
}

func handleGet(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//Is there an id from the path?
	var id string
	if len(request.PathParameters) > 0 {
		id = request.PathParameters["id"]
	}

	switch id {
	case "":
		return listInstances()
	default:
		return getInstance(awsContext, id, request.QueryStringParameters)
	}

}

func handleOtherRequests(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: http.StatusNotImplemented}, nil
}
func makeHandler(awsContext *awsctx.AWSContext) func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		fmt.Println("Received body: ", request.Body)

		var response events.APIGatewayProxyResponse
		var err error

		method := strings.ToLower(request.HTTPMethod)
		switch method {
		case "get":
			response, err = handleGet(awsContext, request)
		default:
			response, err = handleOtherRequests(request)
		}

		return response, err
	}
}

func main() {
	var awsContext awsctx.AWSContext

	sess := session.New()
	svc := dynamodb.New(sess)

	awsContext.DynamoDBSvc = svc

	handler := makeHandler(&awsContext)

	lambda.Start(handler)
}
