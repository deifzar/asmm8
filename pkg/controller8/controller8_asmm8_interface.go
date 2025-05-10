package controller8

import "github.com/gin-gonic/gin"

type Controller8ASMM8Interface interface {
	LaunchScan(*gin.Context)
	LaunchPassive(*gin.Context)
	LaunchActive(*gin.Context)
	LauchCheckLive(*gin.Context)
	// RabbitMQBringConsumerBackAndPublishMessage() error
	// RabbitMQBringConsumerBack() error
	// RabbitMQPublishMessage() error
}
