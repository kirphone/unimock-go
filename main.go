package main

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"unimock/database"
	"unimock/triggers"
)

func main() {
	loggingConfig := zap.NewProductionConfig()
	loggingConfig.Sampling = nil
	loggingConfig.Encoding = "console"
	loggingConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := loggingConfig.Build()

	if err != nil {
		println("Не удалось инициализировать логгер\n%s", err.Error())
		return
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	sqlDB, err := database.InitDatabaseConnection()

	if err != nil {
		return
	}

	app := fiber.New()
	//app.Use(logger.New())
	triggerService := triggers.NewService(sqlDB)
	triggerHandler := triggers.NewHandler(triggerService)

	control := app.Group("/TStub/control")
	triggersController := control.Group("/triggers")
	triggersController.Get("/get", triggerHandler.GetTriggers)
	triggersController.Post("/add", triggerHandler.AddTrigger)

	app.Listen(":8080")
}
