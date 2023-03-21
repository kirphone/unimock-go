package triggers

import (
	"database/sql"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"regexp"
	"strings"
	"unimock/scenarios"
	"unimock/util"
)

const InsertQuery = "INSERT INTO triggers (type, expression, description, active, headers) VALUES (?,?,?,?,?)"
const UpdateQuery = "UPDATE triggers SET type = ?, expression = ?, description = ?, active = ?, headers = ? where id = ?"
const SelectAllQuery = "SELECT * FROM triggers"

type TriggerType string

const (
	Json  TriggerType = "json"
	Regex             = "regex"
)

const contentType = "Content-Type"
const contentTypeJSONValue = "application/json"
const dotAllRegexMod = "(?s)"

type Trigger struct {
	Id               int64
	TriggerType      TriggerType
	Expression       string
	Description      string
	IsActive         bool
	Headers          map[string]string
	expressionRegexp *regexp.Regexp
}

func (trigger *Trigger) validate() bool {
	return trigger.TriggerType != ""
}

type TriggerService struct {
	triggers        map[int64]*Trigger
	db              *sql.DB
	scenarioService *scenarios.ScenarioService
}

func NewService(db *sql.DB, scenarioService *scenarios.ScenarioService) *TriggerService {
	return &TriggerService{
		triggers:        make(map[int64]*Trigger),
		db:              db,
		scenarioService: scenarioService,
	}
}

func (service *TriggerService) GetTriggers() []*Trigger {
	triggerValues := make([]*Trigger, 0, len(service.triggers))

	for _, value := range service.triggers {
		triggerValues = append(triggerValues, value)
	}

	return triggerValues
}

func (service *TriggerService) GetTriggerById(id int64) (*Trigger, error) {
	trigger, ok := service.triggers[id]

	if !ok {
		return nil, &TriggerNotFoundException{
			message: fmt.Sprintf("Триггер с id = %d не найден", id),
		}
	} else {
		return trigger, nil
	}
}

func (service *TriggerService) AddTrigger(trigger *Trigger) error {
	if !trigger.validate() {
		return &TriggerValidationException{message: "Не указан тип триггера"}
	}
	insertStatement, err := service.db.Prepare(InsertQuery)
	if err != nil {
		return err
	}
	res, err := insertStatement.Exec(trigger.TriggerType, trigger.Expression, trigger.Description, trigger.IsActive, buildHeadersForDb(trigger.Headers))
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

	err = trigger.prepare()
	if err != nil {
		return err
	}

	service.triggers[trigger.Id] = trigger
	return nil
}

func (service *TriggerService) UpdateTrigger(trigger *Trigger) error {
	if !trigger.validate() {
		return &TriggerValidationException{message: "Не указан тип триггера"}
	}
	updateStatement, err := service.db.Prepare(UpdateQuery)
	if err != nil {
		return err
	}
	_, err = updateStatement.Exec(trigger.TriggerType, trigger.Expression, trigger.Description, trigger.IsActive, buildHeadersForDb(trigger.Headers), trigger.Id)
	if err != nil {
		return err
	}

	err = updateStatement.Close()
	if err != nil {
		return err
	}

	err = trigger.prepare()
	if err != nil {
		return err
	}

	service.triggers[trigger.Id] = trigger
	return nil
}

func buildHeadersForDb(headers map[string]string) string {
	var res strings.Builder

	for key, value := range headers {
		res.WriteString(fmt.Sprintf("%s=%s, ", key, value))
	}
	return res.String()
}

func getHeadersFromString(headersRow string) map[string]string {
	pairs := strings.Split(headersRow, ",")

	resMap := make(map[string]string, len(pairs))

	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			resMap[kv[0]] = kv[1]
		}
	}

	return resMap
}

func (service *TriggerService) UpdateFromDb() error {
	rows, err := service.db.Query(SelectAllQuery)
	if err != nil {
		return err
	}

	service.triggers = make(map[int64]*Trigger)

	for rows.Next() {
		var t Trigger
		var headersRow string
		err = rows.Scan(&t.Id, &t.TriggerType, &t.Expression, &t.Description, &t.IsActive, &headersRow)
		if err != nil {
			return err
		}

		t.Headers = getHeadersFromString(headersRow)
		err = t.prepare()
		if err != nil {
			return err
		}

		service.triggers[t.Id] = &t
	}
	return nil
}

func (service *TriggerService) ProcessMessage(message *util.Message) (*util.Message, error) {
	for _, trigger := range service.triggers {
		if trigger.TriggerOnMessage(message) {
			log.Debug().Int64("triggerId", trigger.Id).Msg("Выбран триггер")
			return service.scenarioService.ProcessMessage(message, trigger.Id)
		}
	}
	return nil, &TriggerNotFoundException{
		message: "Триггер для сообщения не найден",
	}
}

func containHeaders(messageHeaders map[string]string, triggerHeaders map[string]string) bool {
	for key, valueTrigger := range triggerHeaders {
		valueMessage, ok := messageHeaders[key]
		if !ok || valueMessage != valueTrigger {
			return false
		}
	}
	return true
}

func (trigger *Trigger) prepare() error {
	var err error
	switch trigger.TriggerType {
	case Regex:
		trigger.expressionRegexp, err = regexp.Compile(dotAllRegexMod + trigger.Expression)
		if err != nil {
			return &TriggerValidationException{message: err.Error()}
		}
	case Json:
		if _, ok := trigger.Headers[contentType]; !ok {
			trigger.Headers[contentType] = contentTypeJSONValue
		}
	}
	return nil
}

func (trigger *Trigger) TriggerOnMessage(message *util.Message) bool {
	if !trigger.IsActive || !containHeaders(message.Headers, trigger.Headers) {
		return false
	}

	result := false
	switch trigger.TriggerType {
	case Regex:
		result = trigger.expressionRegexp.MatchString(message.Body)
	case Json:
		result = gjson.Get(message.Body, trigger.Expression).Exists()
	}
	return result
}
