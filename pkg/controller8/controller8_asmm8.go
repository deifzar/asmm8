package controller8

import (
	"database/sql"
	"deifzar/asmm8/pkg/active"
	"deifzar/asmm8/pkg/notification8"
	"deifzar/asmm8/pkg/orchestrator8"
	"strconv"
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

type contextKey string

const (
	deliveryTagKey contextKey = "rabbitmq_delivery_tag"
)

type Controller8ASSM8 struct {
	Db     *sql.DB
	Config *viper.Viper
	Orch   orchestrator8.Orchestrator8Interface
}

func NewController8ASSM8(db *sql.DB, config *viper.Viper) Controller8ASMM8Interface {
	orch, err := orchestrator8.NewOrchestrator8()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Fatal().Msg("Error initializing orchestrator8 in controller constructor")
	}
	return &Controller8ASSM8{Db: db, Config: config, Orch: orch}
}

func (m *Controller8ASSM8) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "asmm8",
	})
}

func (m *Controller8ASSM8) ReadinessCheck(c *gin.Context) {
	dbHealthy := true
	rbHealthy := true
	if err := m.Db.Ping(); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Database ping failed during readiness check")
		dbHealthy = false
	}
	dbStatus := "unhealthy"
	if dbHealthy {
		dbStatus = "healthy"
	}

	queue_consumer := m.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
	qargs_consumer := m.Config.GetStringMap("ORCHESTRATORM8.asmm8.Queue-arguments")

	if !m.Orch.ExistQueue(queue_consumer[1], qargs_consumer) || !m.Orch.ExistConsumersForQueue(queue_consumer[1], qargs_consumer) {
		rbHealthy = false
	}

	rbStatus := "unhealthy"
	if rbHealthy {
		rbStatus = "healthy"
	}

	if dbHealthy && rbHealthy {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "asmm8",
			"checks": gin.H{
				"database": dbStatus,
				"rabbitmq": rbStatus,
			},
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "asmm8",
			"checks": gin.H{
				"database": dbStatus,
				"rabbitmq": rbStatus,
			},
		})
	}
}

// handleNotificationErrorOnFullscan handles errors when fullscan is true by publishing to RabbitMQ and sending error notifications
func (m *Controller8ASSM8) handleNotificationErrorOnFullscan(fullscan bool, message string, urgency string) {
	if fullscan {
		publishingdetails := m.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Publisher")
		m.Orch.PublishToExchange(publishingdetails[0], publishingdetails[1], nil, publishingdetails[2])
		notification8.PoolHelper.PublishSysErrorNotification(message, urgency, "asmm8")
		log8.BaseLogger.Info().Msg("Published message to RabbitMQ for next service (naabu8)")
	}
}

func (m *Controller8ASSM8) LaunchScan(c *gin.Context) {
	// Clean up old files in tmp directory (older than 24 hours)
	cleanup := cleanup8.NewCleanup8()
	if err := cleanup.CleanupDirectory("tmp", 24*time.Hour); err != nil {
		log8.BaseLogger.Error().Err(err).Msg("Failed to cleanup tmp directory")
		// Don't return error here as cleanup failure shouldn't prevent startup
	}
	// Extract delivery tag from request header (set by RabbitMQ handler)
	deliveryTagStr := c.GetHeader("X-RabbitMQ-Delivery-Tag")
	var deliveryTag uint64
	if deliveryTagStr != "" {
		if tag, err := strconv.ParseUint(deliveryTagStr, 10, 64); err == nil {
			deliveryTag = tag
			log8.BaseLogger.Debug().Msgf("Scan triggered via RabbitMQ (deliveryTag: %d)", deliveryTag)
		}
	}
	// Check that RabbitMQ relevant Queue is available.
	// If relevant queue does not exist, inform the user that there is one ASMM8 running at this moment and advise the user to wait for the latest results.
	queue_consumer := m.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
	qargs_consumer := m.Config.GetStringMap("ORCHESTRATORM8.asmm8.Queue-arguments")
	publishingdetails := m.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Publisher")
	if m.Orch.ExistQueue(queue_consumer[1], qargs_consumer) {
		DB := m.Db
		domain8 := db8.NewDb8Domain8(DB)
		get, err := domain8.GetAllEnabled()
		if err != nil {
			// move on and call naabum8 scan
			log8.BaseLogger.Error().Msg("HTTP 500 Response - ASM8 Full scans failed - Error fetching all domains from DB to launch scan.")
			m.handleNotificationErrorOnFullscan(true, "LaunchScan - Error fetching all domains from DB to launch scan", "normal")
			// ACK the message since we're not going to retry
			if deliveryTag > 0 {
				m.Orch.AckScanCompletion(deliveryTag, true) // Mark as "completed" even though it failed early
				// OR use: m.Orch.NackScanMessage(deliveryTag, false) // Don't requeue - permanent failure
			}
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "ASM8 Scans failed. Error fetching all domains from DB to launch scan."})
			return
		}
		if len(get) < 1 {
			// no domains enabled - move on and call naabum8 scan
			m.Orch.PublishToExchange(publishingdetails[0], publishingdetails[1], nil, publishingdetails[2])
			log8.BaseLogger.Info().Msg("ASM8 full scans API call success. No targets in scope")
			// ACK the message since we're not going to retry
			if deliveryTag > 0 {
				m.Orch.AckScanCompletion(deliveryTag, true) // Mark as "completed" even though it failed early
				// OR use: m.Orch.NackScanMessage(deliveryTag, false) // Don't requeue - permanent failure
			}
			c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "ASM8 full scans finished. No target in scope"})
			return
		}
		// install the required tools
		err = utils.InstallTools()
		if err != nil {
			// move on and call naabum8 scan
			log8.BaseLogger.Error().Msg("HTTP 500 Response - ASM8 Full scans failed - Error during tools installation!")
			m.handleNotificationErrorOnFullscan(true, "LaunchScan - Error during tools installation", "normal")
			// ACK the message since we're not going to retry
			if deliveryTag > 0 {
				m.Orch.AckScanCompletion(deliveryTag, true) // Mark as "completed" even though it failed early
				// OR use: m.Orch.NackScanMessage(deliveryTag, false) // Don't requeue - permanent failure
			}
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": "Launching Full scans is not possible at this moment due to interal errors ocurring during the tools installation. Please, check the notification."})
			return
		}
		log8.BaseLogger.Info().Msg("ASM8 full scans API call success")
		c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "Launching ASM8 full scans. Please, check the notification."})
		// run active.
		go m.Active(true, get, deliveryTag)
	} else {
		// move on and call naabum8 scan
		log8.BaseLogger.Info().Msg("Full scans API call cannot launch the scans at this moment - RabbitMQ queues do not exist.")
		m.handleNotificationErrorOnFullscan(true, "LaunchScan - Full scans API call cannot launch the scans at this moment - RabbitMQ queues do not exist.", "normal")
		// If this was a RabbitMQ-triggered scan, NACK it since we can't process
		if deliveryTag > 0 {
			m.Orch.NackScanMessage(deliveryTag, false) // Don't requeue - permanent config issue
		}
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
	go m.Active(false, get, 0)
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

func (m *Controller8ASSM8) Active(fullScan bool, target []model8.Domain8, deliveryTag uint64) {
	var err error
	var scanCompleted bool = false
	var scanFailed bool = false
	var changes_occurred bool = false
	// Ensure we always publish to exchange AND ACK/NACK at the end
	if fullScan {
		defer func() {
			// Recover from panic if any
			if r := recover(); r != nil {
				log8.BaseLogger.Error().Msgf("PANIC recovered in ASMM8 scans: %v", r)
				scanCompleted = false
				scanFailed = true
			}
			// Always publish, but with different payload based on status
			var payload any = nil
			// call naabum8 scan
			if scanFailed {
				payload = map[string]interface{}{
					"status":  "warning",
					"message": "ASMM8 scan is showing warnings. Please, check!",
				}
			} else if !scanCompleted {
				payload = map[string]interface{}{
					"status":  "incomplete",
					"message": "ASMM8 scan did not complete. Unexpected errors.",
				}
			} else {
				payload = map[string]interface{}{
					"status":  "complete",
					"message": "ASMM8 scan run successfully!",
				}
			}
			publishingdetails := m.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Publisher")
			pubErr := m.Orch.PublishToExchange(publishingdetails[0], publishingdetails[1], payload, publishingdetails[2])
			if pubErr != nil {
				log8.BaseLogger.Error().Msgf("Failed to publish to exchange: %v", pubErr)
				// Retry once after brief delay
				time.Sleep(5 * time.Second)
				retryErr := m.Orch.PublishToExchange(publishingdetails[0], publishingdetails[1], payload, publishingdetails[2])
				if retryErr != nil {
					log8.BaseLogger.Error().Msgf("Retry failed: %v", retryErr)
					// Last resort: urgent notification
					notification8.PoolHelper.PublishSysErrorNotification(
						"CRITICAL: Failed to notify naabum8 after ASMM8 scan",
						"urgent",
						"asmm8",
					)
				}
			} else {
				log8.BaseLogger.Info().Msg("Published message to RabbitMQ for next service (naabum8)")
			}
			// ACK or NACK the RabbitMQ message if deliveryTag is set
			if deliveryTag > 0 {
				ackErr := m.Orch.AckScanCompletion(deliveryTag, scanCompleted)
				if ackErr != nil {
					log8.BaseLogger.Error().Msgf("Failed to ACK/NACK message (deliveryTag: %d): %v", deliveryTag, ackErr)
				}
			}
		}()
	}

	var PassiveRunner passive.PassiveRunner
	PassiveRunner.Subdomains = make(map[string][]string)
	var ActiveRunner active.ActiveRunner
	ActiveRunner.Subdomains = make(map[string][]string)
	prevResults := make(map[string][]string)

	for _, d8 := range target {
		PassiveRunner.SeedDomains = append(PassiveRunner.SeedDomains, d8.Name)
		ActiveRunner.SeedDomains = append(ActiveRunner.SeedDomains, d8.Name)
		prevResults[d8.Name], err = m.GetPrevSubdomains(d8.Id, d8.Name)
		if err != nil {
			scanFailed = true
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Warn().Msgf("Active scans: error getting old subdomains for `%s`", d8.Name)
		}
	}

	log8.BaseLogger.Info().Msg("Active scans: Running Passive scans")
	// run passive enumeration and get the results
	passiveResults, err := PassiveRunner.RunPassiveEnum(prevResults)
	if err != nil {
		scanFailed = true
		log8.BaseLogger.Error().Msgf("Passive scan failed: %v", err)
	}
	log8.BaseLogger.Info().Msg("Active scans: Passive scans have concluded")
	PassiveRunner.Subdomains = passiveResults

	wordlist := m.Config.GetString("ASMM8.activeWordList")
	threads := m.Config.GetInt("ASMM8.activeThreads")

	log8.BaseLogger.Info().Msg("Active scans: Running Active scans.")
	activeResults, err := ActiveRunner.RunActiveEnum(wordlist, threads, passiveResults)
	if err != nil {
		scanFailed = true
		log8.BaseLogger.Error().Msgf("Active scan failed: %v", err)
	}
	log8.BaseLogger.Info().Msg("Active scans: Active scans have concluded")
	ActiveRunner.Subdomains = activeResults
	log8.BaseLogger.Info().Msg("Active scans: Fetching scan settings for newly found hostnames from database.")
	generalscansettings8 := db8.NewDb8Generalsettingsscan8(m.Db)
	settings, err := generalscansettings8.Get()
	var scandefaultenabled bool = true
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Warn().Msgf("Active scans: error fetching scan settings for newly found hostnames. Set value to `false` by default")
	} else if settings.Settings.Scannewlyfoundhostname {
		scandefaultenabled = settings.Settings.Scannewlyfoundhostname
	}
	log8.BaseLogger.Info().Msg("Active scans: Updating results in database.")
	hostname8 := db8.NewDb8Hostname8(m.Db)
	for _, d8 := range target {
		notify, err := hostname8.InsertBatch(d8.Id, scandefaultenabled, ActiveRunner.Subdomains[d8.Name])
		if err != nil {
			scanFailed = true
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Warn().Msgf("Active scans: error inserting batch for `%s`", d8.Name)
		}
		if !changes_occurred && notify {
			changes_occurred = true
		}
	}
	// Scans have finished.
	scanCompleted = true
	if changes_occurred {
		// send notification
		notification8.PoolHelper.PublishSecurityNotificationAdmin("New hostnames have been found", "normal", "asmm8")
		notification8.PoolHelper.PublishSecurityNotificationUser("New hostnames have been found", "normal", "asmm8")
	}
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
	passiveResults, err := PassiveRunner.RunPassiveEnum(prevResults)
	if err != nil {
		log8.BaseLogger.Error().Msgf("Passive scan failed: %v", err)
	}
	PassiveRunner.Subdomains = passiveResults
	log8.BaseLogger.Info().Msg("Passive scans: Fetching scan settings for newly found hostnames from database.")
	generalscansettings8 := db8.NewDb8Generalsettingsscan8(m.Db)
	settings, err := generalscansettings8.Get()
	var scandefaultenabled bool = true
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Warn().Msgf("Passive scans: error fetching scan settings for newly found hostnames. Set value to `false` by default")
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
