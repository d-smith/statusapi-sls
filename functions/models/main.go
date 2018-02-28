package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/d-smith/statusapi-sls/awsctx"
	"github.com/d-smith/statusapi-sls/model"
	"net/http"
	"strings"
)

var (
	modelAPI = model.NewModelSvc()
)

func listModels(awsContext *awsctx.AWSContext, tenant string) (events.APIGatewayProxyResponse, error) {
	fmt.Println("listModels")
	models, err := modelAPI.ListModels(awsContext, tenant)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	modelsOut, _ := json.Marshal(&models)

	return events.APIGatewayProxyResponse{Body: string(modelsOut), StatusCode: http.StatusOK}, nil

}

func getModel(awsContext *awsctx.AWSContext, tenant, name string) (events.APIGatewayProxyResponse, error) {
	fmt.Println("getModel", name)
	model, err := modelAPI.GetModel(awsContext, tenant, name)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusBadRequest}, nil
			}
		}

		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	body, err := json.Marshal(&model)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: http.StatusOK}, nil
}

func handleGet(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("request content: %+v", request.RequestContext)

	authZContext := request.RequestContext.Authorizer
	tenant :=  authZContext["tenant"].(string)

	//Is there a name from the path?
	var name string
	if len(request.PathParameters) > 0 {
		name = request.PathParameters["name"]
	}

	switch name {
	case "":
		return listModels(awsContext, tenant)
	default:
		return getModel(awsContext, tenant,name)
	}

}

func handlePost(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var model model.Model
	err := json.Unmarshal([]byte(request.Body), &model)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusBadRequest}, nil
	}

	tenant := request.RequestContext.Authorizer["tenant"].(string)

	err = modelAPI.CreateModel(awsContext, tenant, &model)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return events.APIGatewayProxyResponse{Body: "Model with provide name already exists.", StatusCode: http.StatusBadRequest}, nil
			}
		}
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: http.StatusOK}, nil
}

func handlePut(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var model model.Model
	err := json.Unmarshal([]byte(request.Body), &model)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusBadRequest}, nil
	}

	tenant := request.RequestContext.Authorizer["tenant"].(string)

	err = modelAPI.UpdateModel(awsContext, tenant, &model)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return events.APIGatewayProxyResponse{Body: "No such model to update exists.", StatusCode: http.StatusBadRequest}, nil
			}
		}
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: http.StatusOK}, nil
}

func handleOtherRequests(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: http.StatusBadRequest}, nil
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
