package db8

import (
	"deifzar/asmm8/pkg/model8"

	"github.com/gofrs/uuid/v5"
)

type Db8Hostname8Interface interface {
	InsertHostname(uuid.UUID, model8.PostHostname8) (uuid.UUID, error)
	// return true or false if changes have occured
	InsertBatch(uuid.UUID, bool, []string) (bool, error)
	GetAllHostname() ([]model8.Hostname8, error)
	GetAllHostnameByDomainid(uuid.UUID) ([]model8.Hostname8, error)
	GetAllHostnameIDsByDomainid(uuid.UUID) ([]uuid.UUID, error)
	GetOneHostnameByIdAndDomainid(uuid.UUID, uuid.UUID) (model8.Hostname8, error)
	GetOneHostnameByName(string) (model8.Hostname8, error)
	UpdateHostname(uuid.UUID, uuid.UUID, model8.PostHostname8) (model8.Hostname8, error)
	UpdateLiveColumnByParentID(uuid.UUID, bool) (bool, error)
	UpdateLiveColumnByName(string, bool) (bool, error)
	DeleteHostnameByID(uuid.UUID) (bool, error)
	DeleteAllByParentID(uuid.UUID) (bool, error)
	DeleteHostnameByName(string) (bool, error)
}
