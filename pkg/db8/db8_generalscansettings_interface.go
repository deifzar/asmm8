package db8

import (
	"deifzar/asmm8/pkg/model8"

	_ "github.com/lib/pq"
)

type Db8GeneralscansettingsInterface interface {
	Get() (model8.Generalscansettings, error)
}
