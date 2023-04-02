package scenarios

import (
	"encoding/json"
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

func (handler *ScenarioHandler) AddStep(context *fiber.Ctx) error {
	step := new(ScenarioStep)
	if err := json.Unmarshal(context.Body(), step); err != nil {
		return &StepValidationException{message: err.Error()}
	}
	if err := handler.scenarioService.AddStep(step); err != nil {
		return err
	}
	return context.JSON(step)
}

func (handler *ScenarioHandler) UpdateStep(context *fiber.Ctx) error {
	step := new(ScenarioStep)
	if err := json.Unmarshal(context.Body(), step); err != nil {
		return &StepValidationException{message: err.Error()}
	}
	id, err := strconv.ParseInt(context.Params("id"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("id", err)
	}
	step.Id = id
	if err := handler.scenarioService.UpdateStep(step); err != nil {
		return err
	}
	return nil
}

func (handler *ScenarioHandler) UpdateStepsForTrigger(context *fiber.Ctx) error {
	steps := make(Steps, 0)
	if err := json.Unmarshal(context.Body(), &steps); err != nil {
		return &StepValidationException{message: err.Error()}
	}
	triggerId, err := strconv.ParseInt(context.Params("triggerId"), 10, 64)
	if err != nil {
		return util.CreateParamValidationException("triggerId", err)
	}

	steps, err = handler.scenarioService.UpdateStepsForTrigger(steps, triggerId)
	if err != nil {
		return err
	}
	return context.JSON(steps)
}
