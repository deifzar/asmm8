package controller8

import "github.com/gin-gonic/gin"

type Controller8Domain8Interface interface {
	InsertDomain(*gin.Context)
	GetAllDomain(*gin.Context)
	GetOneDomain(*gin.Context)
	UpdateDomain(*gin.Context)
	DeleteDomain(*gin.Context)
}
