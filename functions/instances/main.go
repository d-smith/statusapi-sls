package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/d-smith/statusapi-sls/awsctx"
	"github.com/d-smith/statusapi-sls/event"
	"github.com/d-smith/statusapi-sls/instance"
	"log"
	"net/http"
	"strings"
)

var (
	instanceSvc = instance.NewInstanceSvc()
	eventsSvc   = event.NewEventSvc()
)

func listInstances() (events.APIGatewayProxyResponse, error) {
	fmt.Println("list instances\n")
	return events.APIGatewayProxyResponse{Body: "other\n", StatusCode: http.StatusOK}, nil
}

func getModelStates(awsContext *awsctx.AWSContext, tenant, id, model string) (events.APIGatewayProxyResponse, error) {
	log.Println("get model states")

	states, err := instanceSvc.StatusForInstance(awsContext, tenant, id, model)
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

func retrieveInstance(awsContext *awsctx.AWSContext, tenant, id string) (events.APIGatewayProxyResponse, error) {
	txnEvents, err := eventsSvc.GetEventsForTxn(awsContext, tenant, id)
	if err != nil {
		fmt.Println("Error retrieving events", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	bodyOut, err := json.Marshal(&txnEvents)
	if err != nil {
		fmt.Println("Error marshalling response", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(bodyOut), StatusCode: http.StatusOK}, nil
}

func getInstance(awsContext *awsctx.AWSContext, tenant, id string, queryStringParams map[string]string) (events.APIGatewayProxyResponse, error) {
	fmt.Println("get instance\n")
	model := queryStringParams["model"]
	if model != "" {
		return getModelStates(awsContext, tenant, id, model)
	}

	return retrieveInstance(awsContext, tenant, id)
}

func handleGet(awsContext *awsctx.AWSContext, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tenant := request.RequestContext.Authorizer["tenant"]

	//Is there an id from the path?
	var id string
	if len(request.PathParameters) > 0 {
		id = request.PathParameters["id"]
	}

	switch id {
	case "":
		return listInstances()
	default:
		return getInstance(awsContext, tenant.(string), id, request.QueryStringParameters)
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
