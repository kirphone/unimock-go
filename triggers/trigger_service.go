package triggers

import (
	"database/sql"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
	"unimock/scenarios"
	"unimock/util"
)

const InsertQuery = "INSERT INTO triggers (type, expression, description, active, headers, subsystem) VALUES (?,?,?,?,?,?)"
const UpdateQuery = "UPDATE triggers SET type = ?, expression = ?, description = ?, active = ?, headers = ?, subsystem = ? where id = ?"
const SelectAllQuery = "SELECT * FROM triggers"
const DeleteQuery = "DELETE FROM triggers WHERE id = ?"

var successTriggerProcessingMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{Name: "trigger_success_requests_duration_histogram", Help: "Успешные обработки запросов триггером"},
	[]string{"trigger_id"})

var failedTriggerProcessingMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{Name: "trigger_failed_requests_duration_histogram", Help: "Неспешные обработки запросов триггером"},
	[]string{"trigger_id"})

type TriggerService struct {
	triggers        map[int64]TriggerInterface
	db              *sql.DB
	scenarioService *scenarios.ScenarioService
}

func NewService(db *sql.DB, scenarioService *scenarios.ScenarioService) *TriggerService {
	return &TriggerService{
		triggers:        make(map[int64]TriggerInterface),
		db:              db,
		scenarioService: scenarioService,
	}
}

func (service *TriggerService) GetTriggers() []TriggerInterface {
	triggerValues := make([]TriggerInterface, 0, len(service.triggers))

	for _, value := range service.triggers {
		triggerValues = append(triggerValues, value)
	}
	return triggerValues
}

func (service *TriggerService) GetTriggerById(id int64) (TriggerInterface, error) {
	trigger, ok := service.triggers[id]

	if !ok {
		return nil, &TriggerNotFoundException{
			message: fmt.Sprintf("Триггер с id = %d не найден", id),
		}
	} else {
		return trigger, nil
	}
}

func (service *TriggerService) AddTrigger(trigger TriggerInterface) error {
	if !trigger.validate() {
		return &TriggerValidationException{message: "Не указан тип триггера"}
	}
	insertStatement, err := service.db.Prepare(InsertQuery)
	if err != nil {
		return err
	}
	res, err := insertStatement.Exec(trigger.getType(), trigger.getExpression(), trigger.getDescription(),
		trigger.getIsActive(), buildHeadersForDb(trigger.getHeaders()), trigger.getSubsystem())
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	trigger.setId(id)

	err = insertStatement.Close()
	if err != nil {
		return err
	}

	err = trigger.prepare()
	if err != nil {
		return err
	}

	service.triggers[trigger.getId()] = trigger
	return nil
}

func (service *TriggerService) UpdateTrigger(trigger TriggerInterface) error {
	if !trigger.validate() {
		return &TriggerValidationException{message: "Не указан тип триггера"}
	}
	updateStatement, err := service.db.Prepare(UpdateQuery)
	if err != nil {
		return err
	}
	_, err = updateStatement.Exec(trigger.getType(), trigger.getExpression(), trigger.getDescription(),
		trigger.getIsActive(), buildHeadersForDb(trigger.getHeaders()), trigger.getSubsystem(), trigger.getId())
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

	service.triggers[trigger.getId()] = trigger
	return nil
}

func (service *TriggerService) DeleteTrigger(id int64) error {
	deleteStatement, err := service.db.Prepare(DeleteQuery)
	if err != nil {
		return err
	}
	_, err = deleteStatement.Exec(id)
	if err != nil {
		return err
	}

	err = deleteStatement.Close()
	if err != nil {
		return err
	}

	delete(service.triggers, id)
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

	service.triggers = make(map[int64]TriggerInterface)

	for rows.Next() {
		var baseTrigger Trigger
		var headersRow string
		err = rows.Scan(&baseTrigger.Id, &baseTrigger.TriggerType, &baseTrigger.Expression,
			&baseTrigger.Description, &baseTrigger.IsActive, &headersRow, &baseTrigger.Subsystem)
		if err != nil {
			return err
		}

		baseTrigger.Headers = getHeadersFromString(headersRow)
		trigger := CreateTriggerFromBaseTrigger(&baseTrigger)
		err = trigger.prepare()
		if err != nil {
			return err
		}

		service.triggers[trigger.getId()] = trigger
	}
	return nil
}

func (service *TriggerService) ProcessMessage(message *util.Message) (*util.Message, error) {
	for _, trigger := range service.triggers {
		if trigger.TriggerOnMessage(message) {
			log.Debug().Int64("triggerId", trigger.getId()).Msg("Выбран триггер")
			startTime := time.Now()
			msg, err := service.scenarioService.ProcessMessage(message, trigger.getId())
			duration := time.Since(startTime).Seconds()
			if err == nil {
				successTriggerProcessingMetric.WithLabelValues(strconv.FormatInt(trigger.getId(), 10)).
					Observe(duration)
			} else {
				failedTriggerProcessingMetric.WithLabelValues(strconv.FormatInt(trigger.getId(), 10)).
					Observe(duration)
			}
			return msg, err
		}
	}
	return nil, &TriggerNotFoundException{
		message: "Триггер для сообщения не найден",
	}
}
