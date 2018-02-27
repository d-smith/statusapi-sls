package model

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/d-smith/statusapi-sls/awsctx"
	"os"
	"log"
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

func (m *ModelSvc) ListModels(awsContext *awsctx.AWSContext, tenant string) ([]string, error) {
	log.Println("ModelSvc ListModels")
	keyCond := expression.Key("tenant").Equal(expression.Value(tenant))

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()
	if err != nil {
		log.Println("Error building expression", err.Error())
		return nil,  err
	}


	input := &dynamodb.QueryInput{
		ExpressionAttributeNames:expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression: expr.KeyCondition(),
		TableName:              aws.String(modelTable),
	}

	qout, err := awsContext.DynamoDBSvc.Query(input)
	if err != nil {
		log.Println("Error executing DynamoDB query")
		return nil, err
	}


	var models = []string{}
	for _, item := range qout.Items {
		models = append(models, *item["name"].S)
	}

	return models, nil
}

func (m *ModelSvc) getModel(awsContext *awsctx.AWSContext, tenant, modelName string) (*dynamodb.GetItemOutput, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(modelName),
			},
			"tenant": {
				S: aws.String(tenant),
			},
		},
		TableName: aws.String(modelTable),
	}

	return awsContext.DynamoDBSvc.GetItem(input)


}

func (m *ModelSvc) GetStepsForModel(awsContext *awsctx.AWSContext, tenant,modelName string) ([]string, error) {

	result, err := m.getModel(awsContext, tenant, modelName)
	if err != nil {
		return nil, err
	}

	steps := result.Item["steps"].L

	return l2slice(steps), nil
}

func (m *ModelSvc) GetModel(awsContext *awsctx.AWSContext, tenant,modelName string) (*Model, error) {
	result, err := m.getModel(awsContext, tenant,modelName)
	if err != nil {
		return nil, err
	}

	model := &Model{
		Name:  modelName,
		Steps: l2slice(result.Item["steps"].L),
	}

	return model, nil
}

func (m *ModelSvc) CreateModel(awsContext *awsctx.AWSContext, tenant string, model *Model) error {
	fmt.Printf("Creating model %s with steps %v", model.Name, model.Steps)

	uniqueName := expression.AttributeNotExists(expression.Name("name"))
	uniqueNameForTenant := uniqueName.And(expression.AttributeNotExists(expression.Name("tenant")))
	uniqueNameCond, _ := expression.NewBuilder().WithCondition(uniqueNameForTenant).Build()

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(model.Name),
			},
			"tenant": {
				S: aws.String(tenant),
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

func (m *ModelSvc) UpdateModel(awsContext *awsctx.AWSContext, tenant string,model *Model) error {
	fmt.Printf("Updating model %s with steps %v", model.Name, model.Steps)

	uniqueName := expression.AttributeExists(expression.Name("name"))
	uniqueNameCond, _ := expression.NewBuilder().WithCondition(uniqueName).Build()

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(model.Name),
			},
			"tenant": {
				S: aws.String(tenant),
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
