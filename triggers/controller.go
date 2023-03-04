package triggers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
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
	zap.L().Info("Получен запрос на вывод всех триггеров")
	return context.JSON(handler.triggerService.getTriggers())
}

func (handler *TriggerHandler) AddTrigger(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на добавление триггера")
	trigger := new(Trigger)
	if err := json.Unmarshal(context.Body(), trigger); err != nil {
		zap.L().Error(err.Error())
		return context.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := handler.triggerService.addTrigger(trigger); err != nil {
		zap.L().Error(err.Error())
		if _, ok := err.(*TriggerValidationException); ok {
			resp := util.NewExceptionResponse(err.Error())
			return context.Status(fiber.StatusBadRequest).JSON(resp)
		}
		return err
	}
	return nil
}
