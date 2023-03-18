package errorhandlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
	"reflect"
	"strconv"
	"unimock/templates"
	"unimock/triggers"
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

type ErrorHandler struct{}

func (errorHandler *ErrorHandler) HandleErrorStatus(context *fiber.Ctx, status int, err error) error {
	zap.L().Error(err.Error())
	resp := NewExceptionResponse(err.Error())
	return context.Status(status).JSON(resp)
}

func (errorHandler *ErrorHandler) HandleError(context *fiber.Ctx, err error) error {
	if _, ok := err.(*templates.TemplateValidationException); ok {
		return errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	} else if _, ok := err.(*templates.TemplateNotFoundException); ok {
		return errorHandler.HandleErrorStatus(context, fiber.StatusNotFound, err)
	} else if _, ok := err.(*triggers.TriggerValidationException); ok {
		return errorHandler.HandleErrorStatus(context, fiber.StatusBadRequest, err)
	} else if _, ok := err.(*triggers.TriggerNotFoundException); ok {
		return errorHandler.HandleErrorStatus(context, fiber.StatusNotFound, err)
	} else if err2, ok := err.(*sqlite.Error); ok {
		return errorHandler.HandleSqlError(context, err2)
	} else {
		errType := reflect.TypeOf(err).String()
		zap.L().Error(errType + ": " + err.Error())
		return err
	}
}

func (errorHandler *ErrorHandler) HandleSqlError(context *fiber.Ctx, err *sqlite.Error) error {
	if err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return errorHandler.HandleErrorStatus(context, fiber.StatusConflict, err)
	} else {
		zap.L().Error("Unknown error code: " + strconv.Itoa(err.Code()) + " " + sqlite.ErrorCodeString[err.Code()] + ": " + err.Error())
		return err
	}
}
