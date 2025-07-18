package db8

import (
	"database/sql"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"
	"time"

	"github.com/gofrs/uuid/v5"
	_ "github.com/lib/pq"
)

type Db8Hostname8 struct {
	Db *sql.DB
}

func NewDb8Hostname8(db *sql.DB) Db8Hostname8Interface {
	return &Db8Hostname8{Db: db}
}

func (m *Db8Hostname8) DeleteHostnameByID(id uuid.UUID) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	_, err = tx.Exec("DELETE FROM ONLY cptm8hostname WHERE id = $1", id)
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	return true, nil
}

func (m *Db8Hostname8) DeleteAllByParentID(domainid uuid.UUID) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	_, err = tx.Exec("DELETE FROM ONLY cptm8hostname WHERE domainid = $1", domainid)
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	return true, nil
}

func (m *Db8Hostname8) DeleteHostnameByName(name string) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	_, err = tx.Exec("DELETE FROM ONLY cptm8hostname WHERE name = $1", name)
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	return true, nil
}

func (m *Db8Hostname8) GetAllHostname() ([]model8.Hostname8, error) {
	query, err := m.Db.Query("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname ORDER BY name")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return []model8.Hostname8{}, err
	}
	var hostnames []model8.Hostname8
	if query != nil {
		for query.Next() {
			var (
				id             uuid.UUID
				name           string
				foundfirsttime time.Time
				live           bool
				domainid       uuid.UUID
				enabled        bool
			)
			err := query.Scan(&id, &name, &foundfirsttime, &live, &domainid, &enabled)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return nil, err
			}
			h := model8.Hostname8{Id: id, Name: name, Foundfirsttime: foundfirsttime, Live: live, Domainid: domainid, Enabled: enabled}
			hostnames = append(hostnames, h)
		}
	}
	return hostnames, nil
}

func (m *Db8Hostname8) GetAllHostnameByDomainid(domainid uuid.UUID) ([]model8.Hostname8, error) {
	query, err := m.Db.Query("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE domainid = $1", domainid)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return []model8.Hostname8{}, err
	}
	var hostnames []model8.Hostname8
	if query != nil {
		for query.Next() {
			var (
				id             uuid.UUID
				name           string
				foundfirsttime time.Time
				live           bool
				domainid       uuid.UUID
				enabled        bool
			)
			err := query.Scan(&id, &name, &foundfirsttime, &live, &domainid, &enabled)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return nil, err
			}
			h := model8.Hostname8{Id: id, Name: name, Foundfirsttime: foundfirsttime, Live: live, Domainid: domainid, Enabled: enabled}
			hostnames = append(hostnames, h)
		}
	}
	return hostnames, nil
}

func (m *Db8Hostname8) GetAllHostnameIDsByDomainid(domainid uuid.UUID) ([]uuid.UUID, error) {
	query, err := m.Db.Query("SELECT id FROM ONLY cptm8hostname WHERE domainid = $1", domainid)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return nil, err
	}
	var ids []uuid.UUID
	if query != nil {
		for query.Next() {
			var (
				id uuid.UUID
			)
			err := query.Scan(&id)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return nil, err
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (m *Db8Hostname8) GetOneHostnameByIdAndDomainid(id uuid.UUID, domainid uuid.UUID) (model8.Hostname8, error) {
	query, err := m.Db.Query("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE id = $1 AND domainid = $2", id, domainid)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Hostname8{}, err
	}
	var hostname model8.Hostname8
	if query != nil {
		for query.Next() {
			var (
				id             uuid.UUID
				name           string
				foundfirsttime time.Time
				live           bool
				domainid       uuid.UUID
				enabled        bool
			)
			err := query.Scan(&id, &name, &foundfirsttime, &live, &domainid, &enabled)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return model8.Hostname8{}, err
			}
			hostname = model8.Hostname8{Id: id, Name: name, Foundfirsttime: foundfirsttime, Live: live, Domainid: domainid, Enabled: enabled}
		}
	}
	return hostname, nil
}

func (m *Db8Hostname8) GetOneHostnameByName(name string) (model8.Hostname8, error) {
	query, err := m.Db.Query("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE name = $1", name)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Hostname8{}, err
	}
	var hostname model8.Hostname8
	if query != nil {
		for query.Next() {
			var (
				id             uuid.UUID
				name           string
				foundfirsttime time.Time
				live           bool
				domainid       uuid.UUID
				enabled        bool
			)
			err := query.Scan(&id, &name, &foundfirsttime, &live, &domainid, &enabled)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return model8.Hostname8{}, err
			}
			hostname = model8.Hostname8{Id: id, Name: name, Foundfirsttime: foundfirsttime, Live: live, Domainid: domainid, Enabled: enabled}
		}
	}
	return hostname, nil
}

func (m *Db8Hostname8) InsertHostname(domainid uuid.UUID, post model8.PostHostname8) (uuid.UUID, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return uuid.Nil, err
	}
	// enabled column value is `True` by default
	stmt, err := tx.Prepare("INSERT INTO cptm8hostname(name, foundfirsttime, live, domainid) VALUES ($1, NOW(), true, $2) ON CONFLICT DO NOTHING")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return uuid.Nil, err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(post.Name, domainid)
	if err2 != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err2.Error())
		return uuid.Nil, err2
	}
	err2 = tx.Commit()
	if err2 != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err2.Error())
		return uuid.Nil, err2
	}
	var h model8.Hostname8
	h, err = m.GetOneHostnameByName(post.Name)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return uuid.Nil, err
	}
	return h.Id, nil
}

// InsertBatch (domainid, enabled, names)
func (m *Db8Hostname8) InsertBatch(domainid uuid.UUID, enabled bool, names []string) (bool, error) {
	var changes_occurred bool = false
	tx, err := m.Db.Begin()
	if err != nil {
		return false, err
	}
	
	// Modified SQL to return information about what happened
	stmt, err := tx.Prepare(`
		INSERT INTO cptm8hostname(name, foundfirsttime, live, enabled, domainid) 
		VALUES ($1, NOW(), true, $2, $3) 
		ON CONFLICT (name) DO UPDATE SET 
			live = EXCLUDED.live
		WHERE cptm8hostname.live != EXCLUDED.live
	`)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	defer stmt.Close()
	
	var err2 error
	for _, n := range names {
		result, err2 := stmt.Exec(n, enabled, domainid)
		if err2 != nil {
			_ = tx.Rollback()
			log8.BaseLogger.Debug().Msg(err2.Error())
			return false, err2
		}
		
		// Check if any rows were affected (new insert or live status changed)
		if !changes_occurred {
			rows, _ := result.RowsAffected()
			if rows > 0 {
				changes_occurred = true
			}
		}
	}
	
	err2 = tx.Commit()
	if err2 != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err2.Error())
		return false, err2
	}
	return changes_occurred, nil
}

func (m *Db8Hostname8) UpdateHostname(domainid, id uuid.UUID, post model8.PostHostname8) (model8.Hostname8, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Hostname8{}, err
	}
	_, err = tx.Exec("UPDATE cptm8hostname SET name = $1, enabled = $2, live = $4 WHERE id = $6 AND domainid = $6", post.Name, post.Enabled, post.Live, id, domainid)
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Hostname8{}, err
	}
	err = tx.Commit()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Hostname8{}, err
	}
	var h model8.Hostname8
	h, err = m.GetOneHostnameByIdAndDomainid(id, domainid)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Hostname8{}, err
	}
	return h, nil
}

func (m *Db8Hostname8) UpdateLiveColumnByParentID(domainid uuid.UUID, live bool) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	_, err = tx.Exec("UPDATE cptm8hostname SET live = $1 WHERE domainid = $2", live, domainid)
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	return true, nil
}

func (m *Db8Hostname8) UpdateLiveColumnByName(name string, live bool) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	_, err = tx.Exec("UPDATE cptm8hostname SET live = $1 WHERE name = $2", live, name)
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	return true, nil
}
