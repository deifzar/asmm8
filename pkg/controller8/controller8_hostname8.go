package controller8

import (
	"database/sql"
	"deifzar/asmm8/pkg/db8"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"

	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/gofrs/uuid/v5"
)

type Controller8Hostname8 struct {
	Db *sql.DB
}

func NewController8Hostname8(db *sql.DB) Controller8Hostname8Interface {
	return &Controller8Hostname8{Db: db}
}

func (m *Controller8Hostname8) DeleteHostname(c *gin.Context) {
	DB := m.Db
	var uri model8.Hostname8Uri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "DeleteHostname failed - Check URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - DeleteHostname")
		return
	}
	hostname8 := db8.NewDb8Hostname8(DB)
	id, err := uuid.FromString(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "DeleteHostname failed - Check UUID URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - DeleteHostname")
		return
	}
	_, err = hostname8.DeleteHostnameByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "delete hostname failed"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - DeleteDomain")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "delete hostname success"})
	log8.BaseLogger.Info().Msg("200 HTTP Response - Delete Hostname sucess")
}

func (m *Controller8Hostname8) GetAllHostname(c *gin.Context) {
	DB := m.Db
	hostname8 := db8.NewDb8Hostname8(DB)
	get, err := hostname8.GetAllHostname()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "success", "msg": "GetAllHostname failed"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - GetAllHostname")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": get, "msg": "get all hostname successfully"})
	log8.BaseLogger.Info().Msg("200 HTTP Response - GetAllHostname")
}

func (m *Controller8Hostname8) GetOneHostname(c *gin.Context) {
	DB := m.Db
	var uri model8.Hostname8Uri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "GetOneHostname failed - Check URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - DeleteHostname")
		return
	}
	hostname8 := db8.NewDb8Hostname8(DB)
	domainid, err := uuid.FromString(uri.Domainid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "GetOneHostname failed - Check UUID for domain URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - DeleteHostname")
		return
	}
	id, err := uuid.FromString(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "GetOneHostname failed - Check UUID for hostname URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - DeleteHostname")
		return
	}
	get, err := hostname8.GetOneHostnameByIdAndDomainid(id, domainid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "GetOneHostname failed."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - GetOneHostname")
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": get, "msg": "get hostname successfully"})
	log8.BaseLogger.Info().Msg("200 HTTP Response - GetOneHostname")
}

func (m *Controller8Hostname8) UpdateHostname(c *gin.Context) {
	DB := m.Db
	var post model8.PostHostname8
	var uri model8.Hostname8Uri
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "UpdateHostname failed - Check Body parameters"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - UpdateHostname")
		return
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "UpdateHostname failed - Check URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - UpdateHostname")
		return
	}
	domainid, err := uuid.FromString(uri.Domainid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "UpdateHostname failed - Check UUID for domain URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - UpdateHostname")
		return
	}
	id, err := uuid.FromString(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "UpdateHostname failed - Check UUID for hostname URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - UpdateHostname")
		return
	}
	hostname8 := db8.NewDb8Hostname8(DB)
	update, err := hostname8.UpdateHostname(domainid, id, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "UpdateHostname failed."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - UpdateHostname")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": update, "msg": "update hostname success"})
	log8.BaseLogger.Info().Msg("200 HTTP Response - UpdateHostname")
}

func (m *Controller8Hostname8) InsertHostname(c *gin.Context) {
	DB := m.Db
	var post model8.PostHostname8
	var uri model8.Hostname8Uri
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "Insert Hostname failed - Check Body parameters"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - Insert Hostname")
		return
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "Insert hostname failed - Check URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - Insert Hostname")
		return
	}
	domainid, err := uuid.FromString(uri.Domainid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "Insert hostname failed - Check UUID for domain URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - Insert Hostname - Check UUID parameters in URL.")
		return
	}
	hostname8 := db8.NewDb8Hostname8(DB)
	insert, err := hostname8.InsertHostname(domainid, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "Insert hostname failed"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - Insert Hostname")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": insert, "msg": "insert hostname success"})
	log8.BaseLogger.Debug().Msg("200 HTTP Response - Insert Hostname")
}
