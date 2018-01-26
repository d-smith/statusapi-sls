package main

import "fmt"

type Model struct {
	Name string `json:"name"`
	States []string `json:"states"`
}

type ModelSvc struct {}

func NewModel() *ModelSvc {
	return &ModelSvc{}
}

func (m *ModelSvc) CreateModel(model *Model) error {
	fmt.Printf("Creating model %s with states %v", model.Name, model.States)
	return nil
}
