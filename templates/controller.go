package templates

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"strconv"
	"unimock/util"
)

type TemplateHandler struct {
	templateService *TemplateService
}

func NewHandler(service *TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: service,
	}
}

func (handler *TemplateHandler) GetTemplates(context *fiber.Ctx) error {
	includeBody, err := strconv.ParseBool(context.Query("includeBody", "true"))
	if err != nil {
		includeBody = true
	}

	if includeBody {
		return context.JSON(handler.templateService.GetTemplates())
	} else {
		return context.JSON(handler.templateService.GetTemplatesWithoutBody())
	}
}

func (handler *TemplateHandler) GetTemplateById(context *fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("id", err)
	}

	template, err := handler.templateService.GetTemplateById(id)
	if err != nil {
		return err
	}

	return context.JSON(template)
}

func (handler *TemplateHandler) AddTemplate(context *fiber.Ctx) error {
	template := new(Template)
	if err := json.Unmarshal(context.Body(), template); err != nil {
		return &TemplateValidationException{message: err.Error()}
	}
	if err := handler.templateService.AddTemplate(template); err != nil {
		return err
	}
	return context.JSON(template)
}

func (handler *TemplateHandler) UpdateTemplate(context *fiber.Ctx) error {
	template := new(Template)
	if err := json.Unmarshal(context.Body(), template); err != nil {
		return &TemplateValidationException{message: err.Error()}
	}
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("id", err)
	}
	template.Id = id
	if err := handler.templateService.UpdateTemplate(template); err != nil {
		return err
	}
	return nil
}

func (handler *TemplateHandler) DeleteTemplate(context *fiber.Ctx) error {
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("id", err)
	}

	if err := handler.templateService.DeleteTemplate(id); err != nil {
		return err
	}
	return nil
}

func (handler *TemplateHandler) ProcessSpecificTemplate(context *fiber.Ctx) error {
	templateId, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("id", err)
	}

	inputMessage := util.Message{
		Body:    string(context.Body()),
		Headers: context.GetReqHeaders(),
	}

	log.Debug().Any("headers", inputMessage.Headers).Str("body", inputMessage.Body).Msg("Получено сообщение")

	outputMessage, err := handler.templateService.ProcessMessage(templateId, &inputMessage)
	if err != nil {
		return err
	}

	for key, value := range outputMessage.Headers {
		context.Append(key, value)
	}

	return context.SendString(outputMessage.Body)
}
