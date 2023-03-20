package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strconv"
	"time"
	"unimock/database"
	"unimock/errorhandlers"
	"unimock/scenarios"
	"unimock/templates"
	"unimock/triggers"
)

func main() {
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Не удалось инициализировать конфигурацию\n %w", err))
		return
	}

	serverPort := viper.GetInt("server.port")
	loggingLevel := viper.GetString("logging.level")
	logFile := viper.GetString("logging.file")
	dbFile := viper.GetString("db.file")
	level, err := zerolog.ParseLevel(loggingLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	fileLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    5,
		MaxBackups: 10,
		MaxAge:     14,
		Compress:   true,
	}

	zerolog.SetGlobalLevel(level)

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	log.Logger = log.Output(zerolog.MultiLevelWriter(os.Stdout, fileLogger))

	log.Info().Msgf("Log level is %s", level.String())
	sqlDB, err := database.InitDatabaseConnection(dbFile)

	if err != nil {
		log.Error().Err(err).Msg("При соединении с базой данных произошла ошибка")
		return
	}

	log.Info().Msg("Соединение с базой данных успешно установлено")

	app := fiber.New(fiber.Config{
		BodyLimit:    50 * 1024 * 1024,
		ErrorHandler: errorhandlers.FinalErrorHandler,
	})

	app.Use(Middleware())

	templateService := templates.NewService(sqlDB)
	err = templateService.UpdateFromDb()
	if err != nil {
		log.Fatal().Err(err).Msg("")
		return
	}

	scenarioService := scenarios.NewService(sqlDB, templateService)
	err = scenarioService.UpdateFromDb()
	if err != nil {
		log.Fatal().Err(err).Msg("")
		return
	}

	triggerService := triggers.NewService(sqlDB, scenarioService)
	err = triggerService.UpdateFromDb()
	if err != nil {
		log.Fatal().Err(err).Msg("")
		return
	}

	triggerHandler := triggers.NewHandler(triggerService)
	templateHandler := templates.NewHandler(templateService)
	scenarioHandler := scenarios.NewHandler(scenarioService)

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

	scenarioController := api.Group("/steps")
	scenarioController.Get("/field/templateId/:templateId", scenarioHandler.GetOrderedStepsByTriggerId)

	api.All("/http/process", triggerHandler.ProcessMessage)

	err = app.Listen(":" + strconv.Itoa(serverPort))
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}

func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		if err != nil {
			err = errorhandlers.HandleError(c, err)
		}

		duration := time.Since(start)

		entry := log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", duration)

		entry.Msg("")

		return err
	}
}
