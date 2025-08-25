package controller8

import (
	"net/http"

	amqpM8 "deifzar/asmm8/pkg/amqpM8"
	"deifzar/asmm8/pkg/log8"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Controller8Health struct {
	Config *viper.Viper
}

func NewController8Health(config *viper.Viper) *Controller8Health {
	return &Controller8Health{
		Config: config,
	}
}

// GetRabbitMQHealth returns the health status of RabbitMQ connections and consumers
func (h *Controller8Health) GetRabbitMQHealth(c *gin.Context) {
	log8.BaseLogger.Info().Msg("Health check requested for RabbitMQ connections")

	healthStatus := make(map[string]interface{})

	// Get connection from pool to check health
	err := amqpM8.WithPooledConnection(func(conn amqpM8.PooledAmqpInterface) error {
		// Check connection status
		connectionStatus := conn.GetConnectionStatus()
		healthStatus["connection"] = connectionStatus
		healthStatus["is_connected"] = conn.IsConnected()

		// Get consumer health information
		consumerHealth := conn.GetConsumerHealth()
		healthStatus["consumers"] = consumerHealth

		// Get active consumers
		activeConsumers := conn.GetActiveConsumers()
		healthStatus["active_consumers"] = activeConsumers

		return nil
	})

	if err != nil {
		log8.BaseLogger.Error().Msgf("Failed to get RabbitMQ health status: %v", err)
		healthStatus["error"] = err.Error()
		healthStatus["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, healthStatus)
		return
	}

	// Determine overall health status
	overallHealthy := true
	if connected, ok := healthStatus["is_connected"].(bool); ok && !connected {
		overallHealthy = false
	}

	if overallHealthy {
		healthStatus["status"] = "healthy"
		c.JSON(http.StatusOK, healthStatus)
	} else {
		healthStatus["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, healthStatus)
	}
}

// GetConsumerHealth returns detailed health information for a specific consumer
func (h *Controller8Health) GetConsumerHealth(c *gin.Context) {
	consumerName := c.Param("consumer")
	if consumerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "consumer name is required"})
		return
	}

	log8.BaseLogger.Info().Msgf("Health check requested for consumer: %s", consumerName)

	var consumerHealth interface{}
	var found bool

	err := amqpM8.WithPooledConnection(func(conn amqpM8.PooledAmqpInterface) error {
		health, exists := conn.GetConsumerHealthByName(consumerName)
		consumerHealth = health
		found = exists
		return nil
	})

	if err != nil {
		log8.BaseLogger.Error().Msgf("Failed to get consumer health for %s: %v", consumerName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "consumer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"consumer_name": consumerName,
		"health":        consumerHealth,
	})
}