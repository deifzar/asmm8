package model8

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Hostname8 struct {
	Id             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Foundfirsttime time.Time `json:"foundfirsttime"`
	Live           bool      `json:"live"`
	Domainid       uuid.UUID `json:"dmainid"`
	Enabled        bool      `json:"enabled"`
}

type PostHostname8 struct {
	Name    string `json:"name" binding:"required"` //binding:"hostname_rfc1123"
	Enabled bool   `json:"enabled" binding:"boolean"`
	// IpAddress      string    `json:"ipAddress" binding:"omitempty,ip"`
	// FoundFirstTime time.Time `json:"foundFirstTime" binding:"required"`
	Live bool `json:"live" binding:"boolean"`
}

type Hostname8Uri struct {
	Domainid string `uri:"id" binding:"uuid"`
	ID       string `uri:"hostnameid" binding:"uuid"`
}
