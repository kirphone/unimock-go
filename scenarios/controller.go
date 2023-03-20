package scenarios

import (
	"github.com/gofiber/fiber/v2"
	"strconv"
	"unimock/util"
)

type ScenarioHandler struct {
	scenarioService *ScenarioService
}

func NewHandler(service *ScenarioService) *ScenarioHandler {
	return &ScenarioHandler{
		scenarioService: service,
	}
}

func (handler *ScenarioHandler) GetOrderedStepsByTriggerId(context *fiber.Ctx) error {
	triggerId, err := strconv.ParseInt(context.Params("triggerId"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("triggerId", err)
	}

	steps := handler.scenarioService.GetOrderedStepsByTriggerId(triggerId)

	return context.JSON(steps)
}
