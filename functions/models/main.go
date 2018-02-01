package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"net/http"
	"strings"
	"github.com/d-smith/statusapi-sls/awsctx"
	"github.com/d-smith/statusapi-sls/model"
)



var (
	modelAPI = model.NewModel()
)

func listModels() (events.APIGatewayProxyResponse, error) {
	fmt.Println("listModels")
	return events.APIGatewayProxyResponse{Body: "models get\n", StatusCode: 200}, nil
}

func getModel(name string) (events.APIGatewayProxyResponse, error) {
	fmt.Println("getModel", name)
	return events.APIGatewayProxyResponse{Body: "models get\n", StatusCode: 200}, nil
}

func handleGet(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//Is there a name from the path?
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

func handlePost(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var model model.Model
	err := json.Unmarshal([]byte(request.Body), &model)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400}, nil
	}

	err = modelAPI.CreateModel(awsContext, &model)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return events.APIGatewayProxyResponse{Body: "Model with provide name already exists.", StatusCode: http.StatusBadRequest}, nil
			}
		}
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func handlePut(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

func makeHandler(awsContext *awsctx.AWSContext) func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		fmt.Println("Received body: ", request.Body)

		var response events.APIGatewayProxyResponse
		var err error

		method := strings.ToLower(request.HTTPMethod)
		switch method {
		case "get":
			response, err = handleGet(awsContext, request)
		case "post":
			response, err = handlePost(awsContext, request)
		case "put":
			response, err = handlePut(awsContext, request)
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
