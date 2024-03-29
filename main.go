package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"
	"unimock/database"
	"unimock/errorhandlers"
	"unimock/scenarios"
	"unimock/templates"
	"unimock/triggers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Не удалось инициализировать конфигурацию\n %w", err))
		return
	}

	embeddedMonitor := viper.GetBool("monitoring.embedded")
	prometheusMonitor := viper.GetBool("monitoring.prometheus")

	setupLogger()

	sqlDB, err := initDatabase()

	if err != nil {
		log.Error().Err(err).Msg("При соединении с базой данных произошла ошибка")
		return
	}

	app := fiber.New(fiber.Config{
		BodyLimit:    50 * 1024 * 1024,
		ErrorHandler: errorhandlers.FinalErrorHandler,
	})

	app.Static("/", "./public")
	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

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
	api.Use(Middleware())
	triggersController := api.Group("/triggers")
	triggersController.Get("", triggerHandler.GetTriggers)
	triggersController.Post("", triggerHandler.AddTrigger)
	triggersController.Get("/:id", triggerHandler.GetTriggerById)
	triggersController.Put("/:id", triggerHandler.UpdateTrigger)
	triggersController.Delete("/:id", triggerHandler.DeleteTrigger)

	templateController := api.Group("/templates")
	templateController.Get("", templateHandler.GetTemplates)
	templateController.Post("", templateHandler.AddTemplate)
	templateController.Get("/:id", templateHandler.GetTemplateById)
	templateController.All("/:id/process*", templateHandler.ProcessSpecificTemplate)
	templateController.Put("/:id", templateHandler.UpdateTemplate)
	templateController.Delete("/:id", templateHandler.DeleteTemplate)

	scenarioController := api.Group("/steps")
	scenarioController.Get("/field/triggerId/:triggerId", scenarioHandler.GetOrderedStepsByTriggerId)
	scenarioController.Post("", scenarioHandler.AddStep)
	scenarioController.Put("/:id", scenarioHandler.UpdateStep)
	scenarioController.Put("/field/triggerId/:triggerId", scenarioHandler.UpdateStepsForTrigger)

	api.All("/http/process*", triggerHandler.ProcessMessage)

	if prometheusMonitor {
		app.Get("/metrics", func(c *fiber.Ctx) error {
			handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
			handler(c.Context())
			return nil
		})
	}

	if embeddedMonitor {
		app.Get("/monitor", monitor.New(monitor.Config{Title: "Unimock Metrics Page"}))
	}

	startServer(app)
}

func setupLogger() {
	loggingLevel := viper.GetString("logging.level")
	logFile := viper.GetString("logging.file")
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
}

func initDatabase() (*sql.DB, error) {
	dbFile := viper.GetString("db.file")
	sqlHistoryDirectory := viper.GetString("db.sql_history.directory")
	return database.InitDatabaseConnection(dbFile, sqlHistoryDirectory)
}

func startServer(app *fiber.App) {
	viper.SetDefault("server.tls", false)
	useTLS := viper.GetBool("server.tls")
	serverPort := viper.GetInt("server.port")

	if useTLS {
		cert, err := tls.LoadX509KeyPair(viper.GetString("server.cert_name"), viper.GetString("server.key_name"))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load SSL certificate")
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		ln, err := tls.Listen("tcp", ":443", tlsConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set up server")
		}

		// Start the server with SSL
		if err := app.Listener(ln); err != nil {
			log.Error().Err(err).Msg("Failed to start server")
		}
	} else {
		err := app.Listen(":" + strconv.Itoa(serverPort))
		if err != nil {
			log.Error().Err(err).Msg("")
		}
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
