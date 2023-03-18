package templates

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"strconv"
	"unimock/util"
)

type TemplateHandler struct {
	templateService *TemplateService
	errorHandler    util.ErrorHandler
}

func NewHandler(service *TemplateService, errorHandler util.ErrorHandler) *TemplateHandler {
	return &TemplateHandler{
		templateService: service,
		errorHandler:    errorHandler,
	}
}

func (handler *TemplateHandler) GetTemplates(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на вывод всех шаблонов")
	return context.JSON(handler.templateService.GetTemplates())
}

func (handler *TemplateHandler) GetTemplateById(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на вывод шаблона с id = " + context.Params("id"))
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return handler.errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	}

	template, err := handler.templateService.GetTemplateById(id)
	if err != nil {
		return handler.errorHandler.HandleError(context, err)
	}

	return context.JSON(template)
}

func (handler *TemplateHandler) AddTemplate(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на добавление триггера")
	template := new(Template)
	if err := json.Unmarshal(context.Body(), template); err != nil {
		return handler.errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	}
	if err := handler.templateService.AddTemplate(template); err != nil {
		return handler.errorHandler.HandleError(context, err)
	}
	return nil
}

func (handler *TemplateHandler) ProcessSpecificTemplate(context *fiber.Ctx) error {
	zap.L().Info("Получен запрос на обработку сообщения по шаблону")

	templateId, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return handler.errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	}

	inputMessage := util.Message{
		Body:    string(context.Body()),
		Headers: context.GetReqHeaders(),
	}

	outputMessage, err := handler.templateService.ProcessMessage(templateId, &inputMessage)
	if err != nil {
		return handler.errorHandler.HandleError(context, err)
	}

	for key, value := range outputMessage.Headers {
		context.Append(key, value)
	}

	return context.SendString(outputMessage.Body)
}
