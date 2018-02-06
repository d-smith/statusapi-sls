package model

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/d-smith/statusapi-sls/awsctx"
	"os"
)

type Model struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

type ModelSvc struct{}

func NewModelSvc() *ModelSvc {
	return &ModelSvc{}
}

var (
	modelTable = os.Getenv("MODEL_TABLE")
)

func slice2SS(strings []string) []*string {
	var ss []*string
	for _, s := range strings {
		ss = append(ss, aws.String(s))
	}

	return ss
}

func ss2slice(ss []*string) []string {
	var outSlice []string
	for _, s := range ss {
		outSlice = append(outSlice, *s)
	}

	return outSlice
}

func (m *ModelSvc) GetStepsForModel(awsContext *awsctx.AWSContext, modelName string) ([]string, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(modelName),
			},
		},
		TableName: aws.String(modelTable),
	}

	result, err := awsContext.DynamoDBSvc.GetItem(input)
	if err != nil {
		return nil, err
	}

	steps := result.Item["steps"].SS

	return ss2slice(steps), nil
}

func (m *ModelSvc) CreateModel(awsContext *awsctx.AWSContext, model *Model) error {
	fmt.Printf("Creating model %s with steps %v", model.Name, model.Steps)

	uniqueName := expression.AttributeNotExists(expression.Name("name"))
	uniqueNameCond, _ := expression.NewBuilder().WithCondition(uniqueName).Build()

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(model.Name),
			},
			"steps": {
				SS: slice2SS(model.Steps),
			},
		},
		TableName:                aws.String(modelTable),
		ConditionExpression:      uniqueNameCond.Condition(),
		ExpressionAttributeNames: uniqueNameCond.Names(),
	}
	_, err := awsContext.DynamoDBSvc.PutItem(input)

	return err
}
