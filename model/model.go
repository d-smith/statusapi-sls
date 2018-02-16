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

func (m *ModelSvc) getModel(awsContext *awsctx.AWSContext, modelName string)(*dynamodb.GetItemOutput, error) {
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

	steps := result.Item["steps"].SS

	return ss2slice(steps), nil
}

func (m *ModelSvc) GetModel(awsContext *awsctx.AWSContext, modelName string) (*Model, error) {
	result, err := m.getModel(awsContext, modelName)
	if err != nil {
		return nil, err
	}

	model := &Model{
		Name: modelName,
		Steps:ss2slice(result.Item["steps"].SS),
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
