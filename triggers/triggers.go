package triggers

import (
	"database/sql"
	"fmt"
	"strings"
)

const InsertQuery = "INSERT INTO Triggers (type, expression, headers, comment, active) VALUES (?,?,?,?,?)"

type Trigger struct {
	Id          int64
	TriggerType string
	Expression  string
	Description string
	IsActive    bool
	Headers     map[string]string
}

func (trigger *Trigger) validate() bool {
	return trigger.TriggerType != ""
}

type TriggerValidationException struct {
	message string
}

func (e *TriggerValidationException) Error() string {
	return e.message
}

type TriggerService struct {
	triggers map[int64]*Trigger
	db       *sql.DB
}

func NewService(db *sql.DB) *TriggerService {
	return &TriggerService{
		triggers: make(map[int64]*Trigger),
		db:       db,
	}
}

func (service *TriggerService) getTriggers() []*Trigger {
	triggerValues := make([]*Trigger, 0, len(service.triggers))

	for _, value := range service.triggers {
		triggerValues = append(triggerValues, value)
	}

	return triggerValues
}

func (service *TriggerService) addTrigger(trigger *Trigger) error {
	if !trigger.validate() {
		return &TriggerValidationException{message: "Не указан тип триггера"}
	}
	insertStatement, err := service.db.Prepare(InsertQuery)
	if err != nil {
		return err
	}
	res, err := insertStatement.Exec(trigger.TriggerType, trigger.Expression, buildHeadersForDb(trigger.Headers), trigger.Description, trigger.IsActive)
	if err != nil {
		return err
	}

	trigger.Id, err = res.LastInsertId()
	if err != nil {
		return err
	}

	err = insertStatement.Close()
	if err != nil {
		return err
	}

	service.triggers[trigger.Id] = trigger
	return nil
}

func buildHeadersForDb(headers map[string]string) string {
	var res strings.Builder

	for key, value := range headers {
		res.WriteString(fmt.Sprintf("%s=%s", key, value))
	}
	return res.String()
}
