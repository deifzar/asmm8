package controller8

import (
	"database/sql"
	"deifzar/asmm8/pkg/db8"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	_ "github.com/lib/pq"
)

type Controller8Domain8 struct {
	Db *sql.DB
}

func NewController8Domain8(db *sql.DB) Controller8Domain8Interface {
	return &Controller8Domain8{Db: db}
}

func (m *Controller8Domain8) DeleteDomain(c *gin.Context) {
	DB := m.Db
	var uri model8.Domain8Uri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "Delete domain failed - Check parameters in URL."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - DeleteDomain")
		return
	}
	domain8 := db8.NewDb8Domain8(DB)
	id, err := uuid.FromString(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "Delete domain failed - Check UUID parameters in URL."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - DeleteDomain - Check UUID parameters in URL.")
		return
	}
	delete, err := domain8.DeleteDomain(id)
	if delete {
		c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "delete domain successfully"})
		log8.BaseLogger.Info().Msg("200 HTTP Response - DeleteDomain")
		return
	} else {
		// http.StatusInternalServerError
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "delete domain failed"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - DeleteDomain")
		return
	}
}

func (m *Controller8Domain8) GetAllDomain(c *gin.Context) {
	DB := m.Db
	domain8 := db8.NewDb8Domain8(DB)
	get, err := domain8.GetAllDomain()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "get all domain failed"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - GetAllDomain")
		return
	} else {
		c.JSON(200, gin.H{"status": "success", "data": get, "msg": "get all domain success"})
		log8.BaseLogger.Info().Msg("200 HTTP Response - GetAllDomain")
		return
	}
}

func (m *Controller8Domain8) GetOneDomain(c *gin.Context) {
	DB := m.Db
	var uri model8.Domain8Uri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "get one domain failed - Check URL parameters"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - GetOneDomain")
		return
	}
	domain8 := db8.NewDb8Domain8(DB)
	id, err := uuid.FromString(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "get one domain failed - Check UUID URL parameters"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - GetOneDomain  - Check UUID parameters in URL.")
		return
	}
	get, err := domain8.GetOneDomain(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "get one domain failed."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - GetOneDomain")
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": get, "msg": "get one domain success"})
		log8.BaseLogger.Info().Msg("200 HTTP Response - GetOneDomain")
		return
	}
}

func (m *Controller8Domain8) UpdateDomain(c *gin.Context) {
	DB := m.Db
	var post model8.PostDomain8
	var uri model8.Domain8Uri
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "UpdateDomain failed - Check Body parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - UpdateDomain")
		return
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "UpdateDomain failed - Check URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - UpdateDomain")
		return
	}
	domain8 := db8.NewDb8Domain8(DB)
	id, err := uuid.FromString(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "UpdateDomain failed - Check UUID URL parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - UpdateDomain - Check UUID parameters in URL.")
		return
	}
	update, err := domain8.UpdateDomain(id, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "UpdateDomain failed."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - GetOneDomain")
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": update, "msg": "Update domain success"})
		log8.BaseLogger.Info().Msg("200 HTTP Response - Update Domain")
		return
	}
}

func (m *Controller8Domain8) InsertDomain(c *gin.Context) {
	DB := m.Db
	var post model8.PostDomain8
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "msg": "InsertDomain failed - Check Body parameters."})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("400 HTTP Response - InsertDomain")
		return
	}
	domain8 := db8.NewDb8Domain8(DB)
	insert, err := domain8.InsertDomain(post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "msg": "InsertDomain failed"})
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Info().Msg("500 HTTP Response - InsertDomain")
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": insert, "msg": "Insert domain success"})
		log8.BaseLogger.Info().Msg("200 HTTP Response - Insert Domain")
		return
	}
}
