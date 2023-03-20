package triggers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"strconv"
	"unimock/util"
)

type TriggerHandler struct {
	triggerService *TriggerService
}

func NewHandler(service *TriggerService) *TriggerHandler {
	return &TriggerHandler{
		triggerService: service,
	}
}

func (handler *TriggerHandler) GetTriggers(context *fiber.Ctx) error {
	return context.JSON(handler.triggerService.GetTriggers())
}

func (handler *TriggerHandler) GetTriggerById(context *fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("id", err)
	}

	trigger, err := handler.triggerService.GetTriggerById(id)
	if err != nil {
		return err
	}

	return context.JSON(trigger)
}

func (handler *TriggerHandler) AddTrigger(context *fiber.Ctx) error {
	trigger := new(Trigger)
	if err := json.Unmarshal(context.Body(), trigger); err != nil {
		return &TriggerValidationException{message: err.Error()}
	}
	if err := handler.triggerService.AddTrigger(trigger); err != nil {
		return err
	}
	return nil
}

func (handler *TriggerHandler) UpdateTrigger(context *fiber.Ctx) error {
	trigger := new(Trigger)
	if err := json.Unmarshal(context.Body(), trigger); err != nil {
		return &TriggerValidationException{message: err.Error()}
	}
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("id", err)
	}
	trigger.Id = id
	if err := handler.triggerService.UpdateTrigger(trigger); err != nil {
		return err
	}
	return nil
}

func (handler *TriggerHandler) ProcessMessage(context *fiber.Ctx) error {
	inputMessage := util.Message{
		Body:    string(context.Body()),
		Headers: context.GetReqHeaders(),
	}

	log.Debug().Any("headers", inputMessage.Headers).Str("body", inputMessage.Body).Msg("Получено сообщение")

	outputMessage, err := handler.triggerService.ProcessMessage(&inputMessage)
	if err != nil {
		return err
	}

	for key, value := range outputMessage.Headers {
		context.Append(key, value)
	}

	return context.SendString(outputMessage.Body)
}
