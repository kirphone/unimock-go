package main

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"unimock/database"
	"unimock/errorhandlers"
	"unimock/templates"
	"unimock/triggers"
)

func main() {
	loggingConfig := zap.NewProductionConfig()
	loggingConfig.Sampling = nil
	loggingConfig.Encoding = "console"
	loggingConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := loggingConfig.Build()

	if err != nil {
		zap.L().Fatal("Не удалось инициализировать логгер\n" + err.Error())
		return
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	sqlDB, err := database.InitDatabaseConnection()

	if err != nil {
		zap.L().Fatal(err.Error())
		return
	}

	app := fiber.New()
	//app.Use(logger.New())
	triggerService := triggers.NewService(sqlDB)
	err = triggerService.UpdateFromDb()
	if err != nil {
		zap.L().Fatal(err.Error())
		return
	}

	templateService := templates.NewService(sqlDB)
	err = templateService.UpdateFromDb()
	if err != nil {
		zap.L().Fatal(err.Error())
		return
	}

	errorHandler := &errorhandlers.ErrorHandler{}
	triggerHandler := triggers.NewHandler(triggerService, errorHandler)
	templateHandler := templates.NewHandler(templateService, errorHandler)

	api := app.Group("/api")
	triggersController := api.Group("/triggers")
	triggersController.Get("", triggerHandler.GetTriggers)
	triggersController.Post("", triggerHandler.AddTrigger)
	triggersController.Get("/:id", triggerHandler.GetTriggerById)
	triggersController.Put("/:id", triggerHandler.UpdateTrigger)

	templateController := api.Group("/templates")
	templateController.Get("", templateHandler.GetTemplates)
	templateController.Post("", templateHandler.AddTemplate)
	templateController.Get("/:id", templateHandler.GetTemplateById)
	templateController.All("/:id/process", templateHandler.ProcessSpecificTemplate)

	app.Listen(":8080")
}
