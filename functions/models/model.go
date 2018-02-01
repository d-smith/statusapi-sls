package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"os"
)

type Model struct {
	Name   string   `json:"name"`
	States []string `json:"states"`
}

type ModelSvc struct{}

func NewModel() *ModelSvc {
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

func (m *ModelSvc) CreateModel(awsContext *AWSContext, model *Model) error {
	fmt.Printf("Creating model %s with states %v", model.Name, model.States)

	uniqueName := expression.AttributeNotExists(expression.Name("name"))
	uniqueNameCond, _ := expression.NewBuilder().WithCondition(uniqueName).Build()

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(model.Name),
			},
			"states": {
				SS: slice2SS(model.States),
			},
		},
		TableName:                aws.String(modelTable),
		ConditionExpression:      uniqueNameCond.Condition(),
		ExpressionAttributeNames: uniqueNameCond.Names(),
	}
	_, err := awsContext.ddbSvc.PutItem(input)

	return err
}
