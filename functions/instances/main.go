package main

import (
    "fmt"
    "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)


func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	return events.APIGatewayProxyResponse{Body: "hi ya from instances\n", StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}