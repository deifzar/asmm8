package controller8

import "github.com/gin-gonic/gin"

type Controller8Hostname8Interface interface {
	InsertHostname(*gin.Context)
	GetAllHostname(*gin.Context)
	GetOneHostname(*gin.Context)
	UpdateHostname(*gin.Context)
	DeleteHostname(*gin.Context)
}
