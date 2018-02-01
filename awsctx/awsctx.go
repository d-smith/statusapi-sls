package awsctx

import "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

type AWSContext struct {
	DynamoDBSvc dynamodbiface.DynamoDBAPI
}
