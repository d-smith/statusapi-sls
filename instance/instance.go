package instance

import (
	"fmt"
	"github.com/d-smith/statusapi-sls/awsctx"
	"github.com/d-smith/statusapi-sls/event"
	"github.com/d-smith/statusapi-sls/model"
	"log"
)

type InstanceSvc struct{}

func NewInstanceSvc() *InstanceSvc {
	return &InstanceSvc{}
}

type StepState struct {
	Step      string `json:"step"`
	StepState string `json:"step_state`
}

var (
	modelSvc = model.NewModelSvc()
	eventSvc = event.NewEventSvc()
)

func (is *InstanceSvc) StatusForInstance(awsContext *awsctx.AWSContext, transactionId, modelName string) ([]StepState, error) {
	log.Println("get status events for txn")
	active, completed, err := eventSvc.GetStatusEventsForTxn(awsContext, transactionId)
	if err != nil {
		return nil, err
	}

	log.Printf("active: %v\n", active)
	log.Printf("completed: %v\n", completed)

	log.Println("get steps events for model", modelName)
	steps, err := modelSvc.GetStepsForModel(awsContext, modelName)
	if err != nil {
		return nil, err
	}

	log.Println("join events and model states")
	var modelStates []StepState

	for _, step := range steps {
		fmt.Println("look for step ", step)
		var event event.StatusEvent
		var ok bool
		event, ok = active[step]
		if !ok {
			event, ok = completed[step]
		}

		if ok {
			log.Println("found step with state ", event.StepState)
			modelStates = append(modelStates, StepState{
				Step:      event.Step,
				StepState: event.StepState,
			})
		} else {
			log.Println("no step state found")
			modelStates = append(modelStates, StepState{
				Step:      step,
				StepState: "",
			})
		}
	}

	return modelStates, nil

}
