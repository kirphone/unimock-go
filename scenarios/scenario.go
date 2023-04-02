package scenarios

import (
	"database/sql"
	"sort"
	"sync"
	"time"
	"unimock/templates"
	"unimock/util"
)

const SelectAllQuery = "SELECT * FROM scenario_steps"
const SelectByTriggerIdQuery = "SELECT * FROM scenario_steps where trigger_id = ?"
const InsertQuery = "INSERT INTO scenario_steps (order_number, value, trigger_id, step_type) VALUES (?,?,?,?)"
const UpdateQuery = "UPDATE scenario_steps SET order_number = ?, value = ?, trigger_id = ?, step_type = ? where id = ?"

type ScenarioStep struct {
	Id          int64            `json:"id"`
	OrderNumber int              `json:"order_number"`
	Value       int64            `json:"value"`
	TriggerId   int64            `json:"trigger_id"`
	StepType    ScenarioStepType `json:"step_type"`
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
	mut             sync.RWMutex
}

func NewService(db *sql.DB, templateService *templates.TemplateService) *ScenarioService {
	return &ScenarioService{
		steps:           make(map[int64]Steps),
		db:              db,
		templateService: templateService,
	}
}

func (service *ScenarioService) AddStep(step *ScenarioStep) error {
	insertStatement, err := service.db.Prepare(InsertQuery)
	if err != nil {
		return err
	}

	defer insertStatement.Close()
	res, err := insertStatement.Exec(step.OrderNumber, step.Value, step.TriggerId, step.StepType)
	if err != nil {
		return err
	}

	step.Id, err = res.LastInsertId()
	if err != nil {
		return err
	}

	service.mut.Lock()
	service.steps[step.TriggerId] = append(service.steps[step.TriggerId], step)
	service.mut.Unlock()
	return nil
}

func (service *ScenarioService) UpdateStep(step *ScenarioStep) error {
	updateStatement, err := service.db.Prepare(UpdateQuery)
	if err != nil {
		return err
	}
	defer updateStatement.Close()
	_, err = updateStatement.Exec(step.OrderNumber, step.Value, step.TriggerId, step.StepType, step.Id)
	if err != nil {
		return err
	}

	service.mut.Lock()
	stepIndex, err := findStepIndexByID(service.steps[step.TriggerId], step.Id)
	if err != nil {
		return service.UpdateFromDb()
	}
	service.steps[step.TriggerId][stepIndex] = step
	service.mut.Unlock()
	return nil
}

func (service *ScenarioService) UpdateStepsForTrigger(steps Steps, triggerId int64) (Steps, error) {
	tx, err := service.db.Begin()
	if err != nil {
		return nil, err
	}
	updateStatement, err := service.db.Prepare(UpdateQuery)
	if err != nil {
		return nil, err
	}

	defer updateStatement.Close()

	insertStatement, err := service.db.Prepare(InsertQuery)
	if err != nil {
		return nil, err
	}

	defer insertStatement.Close()

	for i, _ := range steps {
		if steps[i].Id == -1 {
			_, err := insertStatement.Exec(steps[i].OrderNumber, steps[i].Value, steps[i].TriggerId, steps[i].StepType)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		} else {
			_, err = updateStatement.Exec(steps[i].OrderNumber, steps[i].Value, steps[i].TriggerId, steps[i].StepType, steps[i].Id)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if err := service.updateStepsForTriggerFromDb(triggerId); err != nil {
		return nil, err
	}

	return service.GetOrderedStepsByTriggerId(triggerId), nil
}

func (service *ScenarioService) UpdateFromDb() error {
	rows, err := service.db.Query(SelectAllQuery)
	if err != nil {
		return err
	}

	service.mut.Lock()
	service.steps = make(map[int64]Steps)

	for rows.Next() {
		var step ScenarioStep
		err = rows.Scan(&step.Id, &step.OrderNumber, &step.Value, &step.TriggerId, &step.StepType)
		if err != nil {
			return err
		}

		service.steps[step.TriggerId] = append(service.steps[step.TriggerId], &step)
	}
	service.mut.Unlock()
	return nil
}

func (service *ScenarioService) updateStepsForTriggerFromDb(triggerId int64) error {
	rows, err := service.db.Query(SelectByTriggerIdQuery, triggerId)
	if err != nil {
		return err
	}

	service.mut.Lock()
	service.steps[triggerId] = make(Steps, 0)

	for rows.Next() {
		var step ScenarioStep
		err = rows.Scan(&step.Id, &step.OrderNumber, &step.Value, &step.TriggerId, &step.StepType)
		if err != nil {
			return err
		}

		service.steps[step.TriggerId] = append(service.steps[step.TriggerId], &step)
	}
	service.mut.Unlock()
	return nil
}

func (service *ScenarioService) GetOrderedStepsByTriggerId(triggerId int64) Steps {
	service.mut.RLock()
	steps, ok := service.steps[triggerId]
	if !ok {
		service.mut.RUnlock()
		return Steps{}
	}
	stepsCopy := make(Steps, len(steps))
	copy(stepsCopy, steps)
	service.mut.RUnlock()

	sort.Sort(stepsCopy)
	return stepsCopy
}

func (service *ScenarioService) ProcessMessage(inputMessage *util.Message, triggerId int64) (*util.Message, error) {
	steps := service.GetOrderedStepsByTriggerId(triggerId)
	if len(steps) == 0 {
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

func findStepIndexByID(steps Steps, id int64) (int, error) {
	for i := range steps {
		if steps[i].Id == id {
			return i, nil
		}
	}
	return 0, &StepNotFoundException{}
}
