package controller8

import "github.com/gin-gonic/gin"

type Controller8ASMM8Interface interface {
	LaunchScan(*gin.Context)
	LaunchPassive(*gin.Context)
	LaunchActive(*gin.Context)
	LauchCheckLive(*gin.Context)
	HealthCheck(c *gin.Context)
	ReadinessCheck(c *gin.Context)
	handleNotificationErrorOnFullscan(fullscan bool, message string, urgency string)
	// RabbitMQBringConsumerBackAndPublishMessage() error
	// RabbitMQBringConsumerBack() error
	// RabbitMQPublishMessage() error
}
