package model

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

func slice2L(strings []string) []*dynamodb.AttributeValue {
	l, err := dynamodbattribute.MarshalList(strings)
	if err != nil {
		fmt.Printf("WARNING: error converting string slice to list: %s", err.Error())
	}

	return l
}

func l2slice(l []*dynamodb.AttributeValue) []string {
	var s []string
	err := dynamodbattribute.UnmarshalList(l, &s)
	if err != nil {
		fmt.Printf("WARNING: error converting list to string slice: %s", err.Error())
	}

	return s
}

func (m *ModelSvc) ListModels(awsContext *awsctx.AWSContext) ([]string, error) {

	proj := expression.NamesList(expression.Name("name"))
	expr, err := expression.NewBuilder().WithProjection(proj).Build()
	if err != nil {
		fmt.Println(nil, err)
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
		TableName:                aws.String(modelTable),
	}

	result, err := awsContext.DynamoDBSvc.Scan(input)
	if err != nil {
		return nil, err
	}

	var models = []string{}
	for _, item := range result.Items {
		models = append(models, *item["name"].S)
	}

	return models, nil
}

func (m *ModelSvc) getModel(awsContext *awsctx.AWSContext, modelName string) (*dynamodb.GetItemOutput, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(modelName),
			},
		},
		TableName: aws.String(modelTable),
	}

	return awsContext.DynamoDBSvc.GetItem(input)
}

func (m *ModelSvc) GetStepsForModel(awsContext *awsctx.AWSContext, modelName string) ([]string, error) {

	result, err := m.getModel(awsContext, modelName)
	if err != nil {
		return nil, err
	}

	steps := result.Item["steps"].L

	return l2slice(steps), nil
}

func (m *ModelSvc) GetModel(awsContext *awsctx.AWSContext, modelName string) (*Model, error) {
	result, err := m.getModel(awsContext, modelName)
	if err != nil {
		return nil, err
	}

	model := &Model{
		Name:  modelName,
		Steps: l2slice(result.Item["steps"].L),
	}

	return model, nil
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
				L: slice2L(model.Steps),
			},
		},
		TableName:                aws.String(modelTable),
		ConditionExpression:      uniqueNameCond.Condition(),
		ExpressionAttributeNames: uniqueNameCond.Names(),
	}
	_, err := awsContext.DynamoDBSvc.PutItem(input)

	return err
}

func (m *ModelSvc) UpdateModel(awsContext *awsctx.AWSContext, model *Model) error {
	fmt.Printf("Updating model %s with steps %v", model.Name, model.Steps)

	uniqueName := expression.AttributeExists(expression.Name("name"))
	uniqueNameCond, _ := expression.NewBuilder().WithCondition(uniqueName).Build()

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(model.Name),
			},
			"steps": {
				L: slice2L(model.Steps),
			},
		},
		TableName:                aws.String(modelTable),
		ConditionExpression:      uniqueNameCond.Condition(),
		ExpressionAttributeNames: uniqueNameCond.Names(),
	}
	_, err := awsContext.DynamoDBSvc.PutItem(input)

	return err
}
