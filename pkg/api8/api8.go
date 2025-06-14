package api8

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"

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
	conn, err2 := db.OpenConnection()
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
	orchestrator8.CreateHandleAPICallByService("asmm8")
	orchestrator8.ActivateConsumerByService("asmm8")

	a.DB = conn
	a.Config = v
	return nil
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

	a.Router = r
}

func (a *Api8) Run(addr string) {
	a.Router.Run(addr)
}
