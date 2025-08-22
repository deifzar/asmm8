package controller8

import (
	"database/sql"
	"deifzar/asmm8/pkg/active"
	"deifzar/asmm8/pkg/notification8"
	"deifzar/asmm8/pkg/orchestrator8"
	"time"

	// "deifzar/asmm8/pkg/configparser"
	"deifzar/asmm8/pkg/cleanup8"
	"deifzar/asmm8/pkg/db8"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"
	"deifzar/asmm8/pkg/passive"
	"deifzar/asmm8/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type Controller8ASSM8 struct {
	Db     *sql.DB
	Config *viper.Viper
	Orch   *orchestrator8.Orchestrator8
}

func NewController8ASSM8(db *sql.DB, config *viper.Viper) Controller8ASMM8Interface {
	return &Controller8ASSM8{Db: db, Config: config}
}

func (m *Controller8ASSM8) LaunchScan(c *gin.Context) {
	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}
	// Check that RabbitMQ relevant Queue is available.
	// If relevant queue does not exist, inform the user that there is one ASMM8 running at this moment and advise the user to wait for the latest results.
	orchestrator8, err := orchestrator8.NewOrchestrator8()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Fatal().Msg("Error connecting to the RabbitMQ server.")
	}
	amqp8 := orchestrator8.GetAmqp()
	queue_consumer := m.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
	qargs_consumer := m.Config.GetStringMap("ORCHESTRATORM8.asmm8.Queue-arguments")
	exchange := m.Config.GetStringSlice("ORCHESTRATORM8.naabum8.Queue")[0]
	if amqp8.ExistQueue(queue_consumer[1], qargs_consumer) {
		DB := m.Db
		domain8 := db8.NewDb8Domain8(DB)
		get, err := domain8.GetAllEnabled()
		if err != nil {
			// move on and call naabum8 scan
			log8.BaseLogger.Error().Msg("HTTP 500 Response - ASM8 Full scans failed - Error fetching all domains from DB to launch scan.")
			orchestrator8.PublishToExchangeAndCloseChannelConnection(exchange, "cptm8.naabum8.get.scan", nil, "asmm8")
			notification8.Helper.PublishSysErrorNotification("LaunchScan - Error fetching all domains from DB to launch scan", "urgent", "asmm8")
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "ASM8 Scans failed. Error fetching all domains from DB to launch scan."})
			return
		}
		if len(get) < 1 {
			// no domains enabled - move on and call naabum8 scan
			orchestrator8.PublishToExchangeAndCloseChannelConnection(exchange, "cptm8.naabum8.get.scan", nil, "asmm8")
			log8.BaseLogger.Info().Msg("ASM8 full scans API call success. No targets in scope")
			c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "ASM8 full scans finished. No target in scope"})
			return
		}
		// install the required tools
		err = utils.InstallTools()
		if err != nil {
			// move on and call naabum8 scan
			log8.BaseLogger.Error().Msg("HTTP 500 Response - ASM8 Full scans failed - Error during tools installation!")
			orchestrator8.PublishToExchangeAndCloseChannelConnection(exchange, "cptm8.naabum8.get.scan", nil, "asmm8")
			notification8.Helper.PublishSysErrorNotification("LaunchScan - Error during tools installation", "urgent", "asmm8")
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "Launching Full scans is not possible at this moment due to interal errors ocurring during the tools installation. Please, check the notification."})
			return
		}
		log8.BaseLogger.Info().Msg("ASM8 full scans API call success")
		c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "Launching ASM8 full scans. Please, check the notification."})
		// run active.
		go m.Active(true, orchestrator8, get)
	} else {
		// move on and call naabum8 scan
		log8.BaseLogger.Info().Msg("Full scans API call cannot launch the scans at this moment - RabbitMQ queues do not exist.")
		orchestrator8.PublishToExchangeAndCloseChannelConnection(exchange, "cptm8.naabum8.get.scan", nil, "asmm8")
		notification8.Helper.PublishSysErrorNotification("LaunchScan - Full scans API call cannot launch the scans at this moment - RabbitMQ queues do not exist.", "urgent", "asmm8")
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "HTTP 500 Response - ASMM8 scans failed - Launching ASMM8 Full scans are not possible at this moment due to non-existent RabbitMQ queues."})
		return
	}
}

func (m *Controller8ASSM8) LaunchActive(c *gin.Context) {
	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}
	DB := m.Db
	domain8 := db8.NewDb8Domain8(DB)
	get, err := domain8.GetAllEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "success", "msg": "ASM8 Active scans failed."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Warn().Msg("HTTP 500 Response - ASM8 Active scans failed - Error dumping all domains from DB to launch active scan.")
		return
	}
	if len(get) < 1 {
		// no domains enabled
		log8.BaseLogger.Info().Msg("ASMM8 Active scans API call success. No targets in scope")
		c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "ASMM8 Active scans finished. No target in scope"})
		return
	}
	// install required tools
	err = utils.InstallTools()
	if err != nil {
		log8.BaseLogger.Error().Msg("HTTP 500 Response - ASM8 Active scans failed - Error during tools installation!")
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "Launching Active scans is not possible at this moment due to interal errors ocurring during tools installation. Please, check the notification."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": nil, "msg": "Launching Active scans. Please, check the notification."})
	log8.BaseLogger.Info().Msg("Active scans API call success")
	// run active.
	go m.Active(false, nil, get)
}

func (m *Controller8ASSM8) LaunchPassive(c *gin.Context) {
	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}
	DB := m.Db
	domain8 := db8.NewDb8Domain8(DB)
	get, err := domain8.GetAllEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "success", "msg": "ASM8 Passive scans failed."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Warn().Msg("HTTP 500 Response - ASM8 Passive scans - Error dumping all domains from DB to launch passive scan.")
		return
	}
	if len(get) < 1 {
		// no domains enabled
		log8.BaseLogger.Info().Msg("ASMM8 Pasive scans API call success. No targets in scope")
		c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "ASMM8 Pasive scans finished. No target in scope"})
		return
	}
	// install required tools
	err = utils.InstallTools()
	if err != nil {
		log8.BaseLogger.Error().Msg("HTTP 500 Response - ASM8 Passive scans failed - Error during tools installation!")
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "Launching Passive scans is not possible at this moment due to interal errors ocurring during tools installation. Please, check the notification."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "Launching Passive scans.. Please, check the notification."})
	log8.BaseLogger.Info().Msg("Passive scans API call success")
	go m.Passive(get)
}

func (m *Controller8ASSM8) LauchCheckLive(c *gin.Context) {
	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}
	DB := m.Db
	domain8 := db8.NewDb8Domain8(DB)
	get, err := domain8.GetAllEnabled()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "success", "msg": "ASM8 Check live scans failed"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Warn().Msg("HTTP 500 Response - ASM8 Check live scans failed - Error dumping all domains from DB to check live systems")
		return
	}
	if len(get) < 1 {
		// no domains enabled
		log8.BaseLogger.Info().Msg("ASMM8 Check Live scans API call success. No targets in scope")
		c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "ASMM8 Check Live scans finished. No target in scope"})
		return
	}
	// install required tools
	err = utils.InstallTools()
	if err != nil {
		log8.BaseLogger.Error().Msg("HTTP 500 Response - ASM8 Check live scans failed - Error during tools installation!")
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "Launching Check live scans is not possible at this moment due to interal errors ocurring during tools installation. Please, check the notification."})
		return
	}
	log8.BaseLogger.Info().Msg("Check live API call success")
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": nil, "msg": "Check alive scans are running"})
	go m.CheckLive(get)

	// Notifications that everything went OK via Queue messages have occurend earlier in Active function
	// return
}

func (m *Controller8ASSM8) Active(fullScan bool, orch8 orchestrator8.Orchestrator8Interface, target []model8.Domain8) {
	var PassiveRunner passive.PassiveRunner
	PassiveRunner.Subdomains = make(map[string][]string)
	var ActiveRunner active.ActiveRunner
	ActiveRunner.Subdomains = make(map[string][]string)
	prevResults := make(map[string][]string)
	var err error
	var changes_occurred bool = false

	for _, d8 := range target {
		PassiveRunner.SeedDomains = append(PassiveRunner.SeedDomains, d8.Name)
		ActiveRunner.SeedDomains = append(ActiveRunner.SeedDomains, d8.Name)
		prevResults[d8.Name], err = m.GetPrevSubdomains(d8.Id, d8.Name)
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Warn().Msgf("Active scans: error getting old subdomains for `%s`", d8.Name)
		}
	}

	log8.BaseLogger.Info().Msg("Active scans: Running Passive scans")
	// run passive enumeration and get the results
	passiveResults := PassiveRunner.RunPassiveEnum(prevResults)
	log8.BaseLogger.Info().Msg("Active scans: Passive scans have concluded")
	PassiveRunner.Subdomains = passiveResults

	wordlist := m.Config.GetString("ASMM8.activeWordList")
	threads := m.Config.GetInt("ASMM8.activeThreads")

	log8.BaseLogger.Info().Msg("Active scans: Running Active scans.")
	activeResults := ActiveRunner.RunActiveEnum(wordlist, threads, passiveResults)
	log8.BaseLogger.Info().Msg("Active scans: Active scans have concluded")
	ActiveRunner.Subdomains = activeResults
	log8.BaseLogger.Info().Msg("Active scans: Fetching scan settings for newly found hostnames from database.")
	generalscansettings8 := db8.NewDb8Generalsettingsscan8(m.Db)
	settings, err := generalscansettings8.Get()
	var scandefaultenabled bool = true
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Warn().Msgf("Active scans: error fetching scan settings for newly found hostnames. Set value to `false` by default")
	}
	if settings.Settings.Scannewlyfoundhostname {
		scandefaultenabled = settings.Settings.Scannewlyfoundhostname
	}
	log8.BaseLogger.Info().Msg("Active scans: Updating results in database.")
	hostname8 := db8.NewDb8Hostname8(m.Db)
	for _, d8 := range target {
		notify, err := hostname8.InsertBatch(d8.Id, scandefaultenabled, ActiveRunner.Subdomains[d8.Name])
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Warn().Msgf("Active scans: error inserting batch for `%s`", d8.Name)
		}
		if !changes_occurred && notify {
			changes_occurred = true
		}
	}
	if fullScan {
		// call naabum8 scan
		cptm8_exchange := m.Config.GetStringSlice("ORCHESTRATORM8.naabum8.Queue")[0]
		orch8.PublishToExchangeAndCloseChannelConnection(cptm8_exchange, "cptm8.naabum8.get.scan", nil, "asmm8")
		if changes_occurred {
			// send notification
			notification8.Helper.PublishSecurityNotificationAdmin("New hostnames have been found", "normal", "asmm8")
			notification8.Helper.PublishSecurityNotificationUser("New hostnames have been found", "normal", "asmm8")
		}

	}
	// Scans have finished.
	log8.BaseLogger.Info().Msg("Active scans: Active scan has concluded.")
}

func (m *Controller8ASSM8) Passive(target []model8.Domain8) {
	var PassiveRunner passive.PassiveRunner
	PassiveRunner.Subdomains = make(map[string][]string)
	prevResults := make(map[string][]string)
	var err error

	for _, d8 := range target {
		PassiveRunner.SeedDomains = append(PassiveRunner.SeedDomains, d8.Name)
		prevResults[d8.Name], err = m.GetPrevSubdomains(d8.Id, d8.Name)
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Warn().Msgf("Passive scans: error getting old subdomains for `%s`", d8.Name)
		}
	}
	log8.BaseLogger.Info().Msg("Passive scans: Running Passive scans.")
	// run passive enumeration and get the results
	passiveResults := PassiveRunner.RunPassiveEnum(prevResults)
	PassiveRunner.Subdomains = passiveResults
	log8.BaseLogger.Info().Msg("Active scans: Fetching scan settings for newly found hostnames from database.")
	generalscansettings8 := db8.NewDb8Generalsettingsscan8(m.Db)
	settings, err := generalscansettings8.Get()
	var scandefaultenabled bool = true
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Warn().Msgf("Active scans: error fetching scan settings for newly found hostnames. Set value to `false` by default")
	}
	if settings.Settings.Scannewlyfoundhostname {
		scandefaultenabled = settings.Settings.Scannewlyfoundhostname
	}
	// Update database
	log8.BaseLogger.Info().Msg("Passive scans: Updating results in database.")
	hostname8 := db8.NewDb8Hostname8(m.Db)
	for _, d8 := range target {
		notify, err := hostname8.InsertBatch(d8.Id, scandefaultenabled, PassiveRunner.Subdomains[d8.Name])
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Warn().Msgf("Passive scans: error inserting batch for `%s`", d8.Name)
		}
		if notify {
			// send notification to rabbitMQ
			log8.BaseLogger.Info().Msgf("`%t`", notify)
		}
	}
	//
	log8.BaseLogger.Info().Msg("Passive scans: Passive scan has concluded.")
}

func (m *Controller8ASSM8) CheckLive(target []model8.Domain8) error {
	var ActiveRunner active.ActiveRunner
	prevResults := make(map[string][]string)
	deadResults := make(map[string][]string)
	hostname8 := db8.NewDb8Hostname8(m.Db)
	var err error

	threads := m.Config.GetInt("ASMM8.activeThreads")

	for _, d8 := range target {
		prevResults[d8.Name], err = m.GetPrevSubdomains(d8.Id, d8.Name)
		if err != nil {
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Warn().Msgf("Check Live scan: error getting old subdomains for `%s`", d8.Name)
			return err
		}
		ActiveRunner.Subdomains[d8.Name] = prevResults[d8.Name]
	}
	log8.BaseLogger.Info().Msg("Check Live scan: Running check live scans.")
	activeResults := ActiveRunner.CheckLiveSubdomains(threads)
	// Here we have to clean the hostnames in the database with Rollback situations
	for _, d8 := range target {
		if activeResults[d8.Name] == nil {
			_, err := hostname8.UpdateLiveColumnByParentID(d8.Id, false)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				log8.BaseLogger.Warn().Msgf("Check Live scan: Error updating all dead hostnames under parent domain `%s`", d8.Name)
			} else {
				log8.BaseLogger.Info().Msgf("Check Live scan: Updated dead domains under parent domain `%s`", d8.Name)
			}
		} else {
			//iterate through prevResults and return results not found in activeResults
			deadResults[d8.Name] = utils.Difference(prevResults[d8.Name], activeResults[d8.Name])
			if deadResults[d8.Name] != nil {
				for _, name := range deadResults[d8.Name] {
					_, err := hostname8.UpdateLiveColumnByName(name, false)
					if err != nil {
						log8.BaseLogger.Debug().Msg(err.Error())
						log8.BaseLogger.Warn().Msgf("Check Live scan: Error updating the dead hostname '%s'\n", name)
					} else {
						log8.BaseLogger.Info().Msgf("Check Live scan: Updated the dead hostname `%s`", name)
					}
				}
			}
		}
	}
	return nil
}

func (m *Controller8ASSM8) GetPrevSubdomains(domainid uuid.UUID, domainName string) ([]string, error) {
	hostname8 := db8.NewDb8Hostname8(m.Db)
	get, err := hostname8.GetAllHostnameByDomainid(domainid)
	if err != nil {
		return nil, err
	} else {
		var subdomains []string
		for _, h8 := range get {
			subdomains = append(subdomains, h8.Name)
		}
		return subdomains, nil
	}
}

// func (m *Controller8ASSM8) RabbitMQBringConsumerBackAndPublishMessage() error {
// 	// RabbitMQ queue and consumer for asmm8 should be back to be available.
// 	orchestrator8, err := orchestrator8.NewOrchestrator8()
// 	amqp8 := orchestrator8.GetAmqp()
// 	defer amqp8.CloseChannel()
// 	defer amqp8.CloseConnection()
// 	if err != nil {
// 		log8.BaseLogger.Debug().Msg(err.Error())
// 		log8.BaseLogger.Fatal().Msg("Error connecting to the RabbitMQ server.")
// 		return err
// 	}
// 	orchestrator8.CreateHandleCPTM8()
// 	orchestrator8.ActivateConsumerByService("asmm8")

// 	// Publish message
// 	queue := m.Config.GetStringSlice("ORCHESTRATORM8.naabum8.Queue")
// 	log8.BaseLogger.Info().Msg("RabbitMQ publishing message to queue for NaabuM8 service.")
// 	err = amqp8.Publish(queue[0], "cptm8.naabum8.get.scan", "")
// 	if err != nil {
// 		log8.BaseLogger.Debug().Msg(err.Error())
// 		log8.BaseLogger.Error().Msgf("rabbitMQ publishing message to queue for NaabuM8 service failed")
// 		return err
// 	}
// 	log8.BaseLogger.Info().Msg("RabbitMQ publishing message to NaabuM8 queue service success.")
// 	return nil
// }

// func (m *Controller8ASSM8) RabbitMQBringConsumerBack() error {

// 	// RabbitMQ queue and consumer for asmm8 should be back to be available.
// 	orchestrator8, err := orchestrator8.NewOrchestrator8()
// 	amqp8 := orchestrator8.GetAmqp()
// 	defer amqp8.CloseChannel()
// 	defer amqp8.CloseConnection()
// 	if err != nil {
// 		log8.BaseLogger.Debug().Msg(err.Error())
// 		log8.BaseLogger.Fatal().Msg("Error connecting to the RabbitMQ server.")
// 		return err
// 	}
// 	orchestrator8.CreateHandleCPTM8()
// 	orchestrator8.ActivateConsumerByService("asmm8")
// 	return nil
// }

// func (m *Controller8ASSM8) RabbitMQPublishMessage() error {
// 	orchestrator8, err := orchestrator8.NewOrchestrator8()
// 	amqp8 := orchestrator8.GetAmqp()
// 	defer amqp8.CloseChannel()
// 	defer amqp8.CloseConnection()
// 	if err != nil {
// 		log8.BaseLogger.Debug().Msg(err.Error())
// 		log8.BaseLogger.Fatal().Msg("Error connecting to the RabbitMQ server.")
// 		return err
// 	}
// 	queue := m.Config.GetStringSlice("ORCHESTRATORM8.naabum8.Queue")
// 	log8.BaseLogger.Info().Msg("RabbitMQ publishing message to queue for NaabuM8 service.")
// 	err = amqp8.Publish(queue[0], "cptm8.naabum8.get.scan", "")
// 	if err != nil {
// 		log8.BaseLogger.Debug().Msg(err.Error())
// 		log8.BaseLogger.Error().Msgf("rabbitMQ publishing message to queue for NaabuM8 service failed")
// 		return err
// 	}
// 	log8.BaseLogger.Info().Msg("RabbitMQ publishing message to NaabuM8 queue service success.")
// 	return nil
// }
