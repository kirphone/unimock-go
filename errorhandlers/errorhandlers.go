package errorhandlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
	"reflect"
	"unimock/templates"
	"unimock/triggers"
	"unimock/util"
)

type ExceptionResponse struct {
	Message string
}

func NewExceptionResponse(message string) *ExceptionResponse {
	return &ExceptionResponse{message}
}

func (response *ExceptionResponse) setMessage(message string) {
	response.Message = message
}

func HandleErrorStatus(context *fiber.Ctx, status int, err error) error {
	log.Error().Err(err).Msg("")
	resp := NewExceptionResponse(err.Error())
	return context.Status(status).JSON(resp)
}

func HandleError(context *fiber.Ctx, err error) error {
	switch v := err.(type) {
	case *templates.TemplateValidationException:
		return HandleErrorStatus(context, fiber.StatusBadRequest, err)
	case *templates.TemplateNotFoundException:
		return HandleErrorStatus(context, fiber.StatusNotFound, err)
	case *triggers.TriggerValidationException:
		return HandleErrorStatus(context, fiber.StatusBadRequest, err)
	case *triggers.TriggerNotFoundException:
		return HandleErrorStatus(context, fiber.StatusNotFound, err)
	case *util.ParamValidationException:
		return HandleErrorStatus(context, fiber.StatusBadRequest, err)
	case *sqlite.Error:
		return HandleSqlError(context, v)
	default:
		context.Status(fiber.StatusInternalServerError)
		errType := reflect.TypeOf(err).String()
		log.Error().Err(err).Msgf("Unknown error: %s", errType)
		return err
	}
}

func HandleSqlError(context *fiber.Ctx, err *sqlite.Error) error {
	if err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE || err.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
		return HandleErrorStatus(context, fiber.StatusConflict, err)
	} else {
		log.Error().Err(err).Msgf("Unknown error code: %d %s", err.Code(), sqlite.ErrorCodeString[err.Code()])
		context.Status(fiber.StatusInternalServerError)
		return err
	}
}

func FinalErrorHandler(context *fiber.Ctx, err error) error {
	resp := NewExceptionResponse(err.Error())
	return context.JSON(resp)
}
