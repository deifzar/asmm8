package model8

import (
	"github.com/gofrs/uuid/v5"
)

type Settingsscan struct {
	Scannewlyfoundhostname bool `json:"scannewlyfoundhostname"`
}

type Generalscansettings struct {
	Id       uuid.UUID    `json:"id"`
	Settings Settingsscan `json:"settings"`
}
