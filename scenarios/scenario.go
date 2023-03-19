package scenarios

import (
	"database/sql"
	"sort"
)

type ScenarioStep struct {
	Id          int64
	OrderNumber int32
	Value       string
	TriggerId   int64
	StepType    ScenarioStepType
}

type ScenarioStepType string

const (
	TemplateProcessing ScenarioStepType = "template_processing"
	SendResponse                        = "send_response"
	Delay                               = "delay"
)

type Steps []*ScenarioStep

func (steps Steps) Len() int           { return len(steps) }
func (steps Steps) Swap(i, j int)      { steps[i], steps[j] = steps[j], steps[i] }
func (steps Steps) Less(i, j int) bool { return steps[i].OrderNumber < steps[j].OrderNumber }

type ScenarioService struct {
	steps map[int64]Steps
	db    *sql.DB
}

func NewService(db *sql.DB) *ScenarioService {
	return &ScenarioService{
		steps: make(map[int64]Steps),
		db:    db,
	}
}

func (service *ScenarioService) GetOrderedStepsByTriggerId(triggerId int64) Steps {
	steps := service.steps[triggerId]
	sort.Sort(steps)
	return steps
}
