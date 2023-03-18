package triggers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"strconv"
	"unimock/util"
)

type TriggerHandler struct {
	triggerService *TriggerService
	errorHandler   util.ErrorHandler
}

func NewHandler(service *TriggerService, errorHandler util.ErrorHandler) *TriggerHandler {
	return &TriggerHandler{
		triggerService: service,
		errorHandler:   errorHandler,
	}
}

func (handler *TriggerHandler) GetTriggers(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на вывод всех триггеров")
	return context.JSON(handler.triggerService.GetTriggers())
}

func (handler *TriggerHandler) GetTriggerById(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на вывод триггера с id = " + context.Params("id"))
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return handler.errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	}

	trigger, err := handler.triggerService.GetTriggerById(id)
	if err != nil {
		return handler.errorHandler.HandleError(context, err)
	}

	return context.JSON(trigger)
}

func (handler *TriggerHandler) AddTrigger(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на добавление триггера")
	trigger := new(Trigger)
	if err := json.Unmarshal(context.Body(), trigger); err != nil {
		return handler.errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	}
	if err := handler.triggerService.AddTrigger(trigger); err != nil {
		return handler.errorHandler.HandleError(context, err)
	}
	return nil
}

func (handler *TriggerHandler) UpdateTrigger(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на обновление триггера с id = " + context.Params("id"))
	trigger := new(Trigger)
	if err := json.Unmarshal(context.Body(), trigger); err != nil {
		return handler.errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	}
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return handler.errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	}
	trigger.Id = id
	if err := handler.triggerService.UpdateTrigger(trigger); err != nil {
		return handler.errorHandler.HandleError(context, err)
	}
	return nil
}
