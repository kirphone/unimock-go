package scenarios

import (
	"database/sql"
	"sort"
	"time"
	"unimock/templates"
	"unimock/util"
)

const SelectAllQuery = "SELECT * FROM scenario_steps"

type ScenarioStep struct {
	Id          int64
	OrderNumber int32
	Value       int64
	TriggerId   int64
	StepType    ScenarioStepType
}

type ScenarioStepType string

const (
	TemplateProcessing ScenarioStepType = "template_processing"
	Delay                               = "delay"
)

type Steps []*ScenarioStep

func (steps Steps) Len() int           { return len(steps) }
func (steps Steps) Swap(i, j int)      { steps[i], steps[j] = steps[j], steps[i] }
func (steps Steps) Less(i, j int) bool { return steps[i].OrderNumber < steps[j].OrderNumber }

type ScenarioService struct {
	steps           map[int64]Steps
	db              *sql.DB
	templateService *templates.TemplateService
}

func NewService(db *sql.DB, templateService *templates.TemplateService) *ScenarioService {
	return &ScenarioService{
		steps:           make(map[int64]Steps),
		db:              db,
		templateService: templateService,
	}
}

func (service *ScenarioService) UpdateFromDb() error {
	rows, err := service.db.Query(SelectAllQuery)
	if err != nil {
		return err
	}

	service.steps = make(map[int64]Steps)

	for rows.Next() {
		var step ScenarioStep
		err = rows.Scan(&step.Id, &step.OrderNumber, &step.Value, &step.TriggerId, &step.StepType)
		if err != nil {
			return err
		}

		service.steps[step.TriggerId] = append(service.steps[step.TriggerId], &step)
	}
	return nil
}

func (service *ScenarioService) GetOrderedStepsByTriggerId(triggerId int64) Steps {
	steps := service.steps[triggerId]
	sort.Sort(steps)
	return steps
}

func (service *ScenarioService) ProcessMessage(inputMessage *util.Message, triggerId int64) (*util.Message, error) {
	steps := service.GetOrderedStepsByTriggerId(triggerId)
	if steps == nil || len(steps) == 0 {
		return &util.Message{}, nil
	}
	message := inputMessage
	for _, step := range steps {
		switch step.StepType {
		case TemplateProcessing:
			var err error
			message, err = service.templateService.ProcessMessage(step.Value, message)
			if err != nil {
				return nil, err
			}
		case Delay:
			time.Sleep(time.Duration(step.Value) * time.Millisecond)
		default:

		}
	}
	return message, nil
}
