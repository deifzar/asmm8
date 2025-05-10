package db8

import (
	"database/sql"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"

	"github.com/gofrs/uuid/v5"
	_ "github.com/lib/pq"
)

type Db8Generalscansettings8 struct {
	Db *sql.DB
}

func NewDb8Generalsettingsscan8(db *sql.DB) Db8GeneralscansettingsInterface {
	return &Db8Generalscansettings8{Db: db}
}

func (m *Db8Generalscansettings8) Get() (model8.Generalscansettings, error) {
	query, err := m.Db.Query("SELECT id, settings FROM cptm8generalscansettings")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Generalscansettings{}, err
	}
	var generalsettings model8.Generalscansettings
	if query != nil {
		if query.Next() {
			var (
				id       uuid.UUID
				settings model8.Settingsscan
			)
			err := query.Scan(&id, &settings)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return model8.Generalscansettings{}, err
			}
			generalsettings = model8.Generalscansettings{Id: id, Settings: settings}
		}
	}
	return generalsettings, nil
}
