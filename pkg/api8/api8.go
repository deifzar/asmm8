package api8

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	"deifzar/asmm8/pkg/cleanup8"
	"deifzar/asmm8/pkg/configparser"
	"deifzar/asmm8/pkg/controller8"
	"deifzar/asmm8/pkg/db8"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/orchestrator8"

	"github.com/gin-gonic/gin"
)

type Api8 struct {
	DB     *sql.DB
	Router *gin.Engine
	Config *viper.Viper
	// Orchestrator8  orchestrator8.Orchestrator8Interface
}

func (a *Api8) Init() error {
	// Create log and tmp directories if they don't exist
	if err := os.MkdirAll("configs", 0750); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create configs directory")
		return err
	}
	if err := os.MkdirAll("log", 0750); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create log directory")
		return err
	}
	if err := os.MkdirAll("tmp", 0750); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to create tmp directory")
		return err
	}

	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}

	v, err := configparser.InitConfigParser()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("Error initialising the config parser.")
		return err
	}

	location := v.GetString("Database.location")
	port := v.GetInt("Database.port")
	schema := v.GetString("Database.schema")
	database := v.GetString("Database.database")
	username := v.GetString("Database.username")
	password := v.GetString("Database.password")

	var db db8.Db8
	db.InitDatabase8(location, port, schema, database, username, password)
	connDB, err2 := db.OpenConnection()
	if err2 != nil {
		log8.BaseLogger.Error().Msg("Error connecting into DB.")
		return err2
	}

	orchestrator8, err := orchestrator8.NewOrchestrator8()
	if err != nil {
		log8.BaseLogger.Error().Msg("Error connecting to the RabbitMQ server.")
		return err
	}
	err = orchestrator8.InitOrchestrator()
	if err != nil {
		log8.BaseLogger.Error().Msg("Error bringing up the RabbitMQ exchanges.")
		return err
	}
	err = orchestrator8.ActivateQueueByService("asmm8")
	if err != nil {
		log8.BaseLogger.Error().Msg("Error bringing up the RabbitMQ queues for the `asmm8` service.")
		return err
	}
	err = orchestrator8.ActivateConsumerByService("asmm8")
	if err != nil {
		log8.BaseLogger.Error().Msg("Error activating consumer with dedicated connection for the `asmm8` service.")
		return err
	}

	a.DB = connDB
	a.Config = v
	return nil
}

// InitializeConsumerAfterReady starts a goroutine that waits for the API service
// to become ready (via /ready endpoint) before initializing RabbitMQ queues and consumers.
// This prevents consumers from receiving messages before the API can handle them.
func (a *Api8) InitializeConsumerAfterReady() {
	go func() {
		locationService := a.Config.GetString("ORCHESTRATORM8.Services.asmm8")
		requestURL := locationService + "/ready"

		log8.BaseLogger.Info().Msg("Waiting for API service to become ready before activating RabbitMQ consumer...")

		// Poll the /ready endpoint until the service is healthy
		maxRetries := 60 // 5 minutes total (60 * 5 seconds)
		retryCount := 0
		for {
			resp, err := http.Get(requestURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				log8.BaseLogger.Info().Msg("API service is ready. Initializing RabbitMQ consumer...")
				break
			}
			if resp != nil {
				resp.Body.Close()
			}

			retryCount++
			if retryCount >= maxRetries {
				log8.BaseLogger.Error().Msg("Timeout waiting for API service to become ready. Consumer will not be activated.")
				return
			}

			time.Sleep(5 * time.Second)
		}

		// Initialize RabbitMQ orchestrator
		orchestrator8, err := orchestrator8.NewOrchestrator8()
		if err != nil {
			log8.BaseLogger.Error().Err(err).Msg("Error connecting to the RabbitMQ server.")
			return
		}

		err = orchestrator8.InitOrchestrator()
		if err != nil {
			log8.BaseLogger.Error().Err(err).Msg("Error bringing up the RabbitMQ exchanges.")
			return
		}

		err = orchestrator8.ActivateQueueByService("asmm8")
		if err != nil {
			log8.BaseLogger.Error().Err(err).Msg("Error bringing up the RabbitMQ queues for the `asmm8` service.")
			return
		}

		err = orchestrator8.ActivateConsumerByService("asmm8")
		if err != nil {
			log8.BaseLogger.Error().Err(err).Msg("Error activating consumer with dedicated connection for the `asmm8` service.")
			return
		}

		log8.BaseLogger.Info().Msg("RabbitMQ consumer successfully activated for asmm8 service.")
	}()
}

func (a *Api8) Routes() {
	r := gin.Default()
	// domain CRUD
	contrDomain8 := controller8.NewController8Domain8(a.DB)
	r.GET("/domain", contrDomain8.GetAllDomain)
	r.POST("/domain", contrDomain8.InsertDomain)
	r.GET("/domain/:id", contrDomain8.GetOneDomain)
	r.PUT("/domain/:id", contrDomain8.UpdateDomain)
	r.DELETE("/domain/:id", contrDomain8.DeleteDomain)

	// hostname CRUD
	contrHostname8 := controller8.NewController8Hostname8(a.DB)
	r.GET("/domain/:id/hostname", contrHostname8.GetAllHostname)
	r.POST("/domain/:id/hostname", contrHostname8.InsertHostname)
	r.GET("/domain/:id/hostname/:hostnameid", contrHostname8.GetOneHostname)
	r.PUT("/domain/:id/hostname/:hostnameid", contrHostname8.UpdateHostname)
	r.DELETE("/domain/:id/hostname/:hostnameid", contrHostname8.DeleteHostname)

	// ASM Scan
	contrASSM8 := controller8.NewController8ASSM8(a.DB, a.Config)
	r.GET("/scan", contrASSM8.LaunchScan)
	r.GET("/scan/passive", contrASSM8.LaunchPassive)
	r.GET("/scan/active", contrASSM8.LaunchActive)
	r.GET("/scan/check", contrASSM8.LauchCheckLive)

	// Health live probes
	r.GET("/health", contrASSM8.HealthCheck)
	r.GET("/ready", contrASSM8.ReadinessCheck)

	a.Router = r
}

func (a *Api8) Run(addr string) {
	a.Router.Run(addr)
}
