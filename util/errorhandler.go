package util

import "github.com/gofiber/fiber/v2"

type ErrorHandler interface {
	HandleError(context *fiber.Ctx, err error) error
	HandleErrorStatus(context *fiber.Ctx, status int, err error) error
}
