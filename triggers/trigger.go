package triggers

import (
	"github.com/tidwall/gjson"
	"regexp"
	"strings"
	"unimock/util"
)

type TriggerType string

const (
	Json  TriggerType = "json"
	Regex TriggerType = "regex"
)

const contentType = "Content-Type"
const contentTypeJSONValue = "application/json"
const dotAllRegexMod = "(?s)"

type TriggerInterface interface {
	validate() bool
	getId() int64
	setId(id int64)
	getType() TriggerType
	setType(trigger TriggerType)
	getExpression() string
	setExpression(e string)
	getDescription() string
	setDescription(d string)
	getIsActive() bool
	setIsActive(a bool)
	getHeaders() map[string]string
	setHeaders(h map[string]string)
	getSubsystem() string
	setSubsystem(subsystem string)
	prepare() error
	TriggerOnMessage(message *util.Message) bool
}

type Trigger struct {
	Id          int64             `json:"id"`
	TriggerType TriggerType       `json:"type"`
	Expression  string            `json:"expression"`
	Description string            `json:"description"`
	IsActive    bool              `json:"is_active"`
	Headers     map[string]string `json:"headers"`
	Subsystem   string            `json:"subsystem"`
}

func (trigger *Trigger) validate() bool {
	return trigger.TriggerType != ""
}

func (trigger *Trigger) getId() int64 {
	return trigger.Id
}

func (trigger *Trigger) setId(id int64) {
	trigger.Id = id
}

func (trigger *Trigger) getType() TriggerType {
	return trigger.TriggerType
}

func (trigger *Trigger) setType(triggerType TriggerType) {
	trigger.TriggerType = triggerType
}

func (trigger *Trigger) getExpression() string {
	return trigger.Expression
}

func (trigger *Trigger) setExpression(expression string) {
	trigger.Expression = expression
}

func (trigger *Trigger) getDescription() string {
	return trigger.Description
}

func (trigger *Trigger) setDescription(description string) {
	trigger.Description = description
}

func (trigger *Trigger) getIsActive() bool {
	return trigger.IsActive
}

func (trigger *Trigger) setIsActive(isActive bool) {
	trigger.IsActive = isActive
}

func (trigger *Trigger) getHeaders() map[string]string {
	return trigger.Headers
}

func (trigger *Trigger) setHeaders(headers map[string]string) {
	trigger.Headers = headers
}

func (trigger *Trigger) getSubsystem() string {
	return trigger.Subsystem
}

func (trigger *Trigger) setSubsystem(subsystem string) {
	trigger.Subsystem = subsystem
}

type RegexTrigger struct {
	*Trigger
	expressionRegexp *regexp.Regexp
}

type JsonGsonTrigger struct {
	*Trigger
}

func (trigger *RegexTrigger) prepare() error {
	var err error
	trigger.expressionRegexp, err = regexp.Compile(dotAllRegexMod + trigger.Expression)
	if err != nil {
		return &TriggerValidationException{message: err.Error()}
	}
	return nil
}

func (trigger *JsonGsonTrigger) prepare() error {
	if _, ok := trigger.Headers[contentType]; !ok {
		trigger.Headers[contentType] = contentTypeJSONValue
	}
	return nil
}

func containHeaders(messageHeaders map[string]string, triggerHeaders map[string]string) bool {
	for key, valueTrigger := range triggerHeaders {
		valueMessage, ok := messageHeaders[key]
		if !ok {
			return false
		}

		splitValueMessage := strings.Split(valueMessage, ";")

		// Trim any whitespace
		for i, val := range splitValueMessage {
			splitValueMessage[i] = strings.TrimSpace(val)
		}

		contains := false
		for _, val := range splitValueMessage {
			if valueTrigger == val {
				contains = true
				break
			}
		}

		if !contains {
			return false
		}
	}
	return true
}

func (trigger *Trigger) TriggerOnMessage(message *util.Message) bool {
	return trigger.IsActive && containHeaders(message.Headers, trigger.Headers)
}

func (trigger *RegexTrigger) TriggerOnMessage(message *util.Message) bool {
	if !trigger.Trigger.TriggerOnMessage(message) {
		return false
	}

	return trigger.expressionRegexp.MatchString(message.Body)
}

func (trigger *JsonGsonTrigger) TriggerOnMessage(message *util.Message) bool {
	if !trigger.Trigger.TriggerOnMessage(message) {
		return false
	}

	return gjson.Get(message.Body, trigger.Expression).Exists()
}

func CreateTriggerFromBaseTrigger(baseTrigger *Trigger) (trigger TriggerInterface) {
	switch baseTrigger.TriggerType {
	case Regex:
		trigger = &RegexTrigger{Trigger: baseTrigger}
	case Json:
		trigger = &JsonGsonTrigger{Trigger: baseTrigger}
	}
	return trigger
}
