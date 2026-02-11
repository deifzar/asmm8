package db8

import (
	"database/sql"
	"deifzar/asmm8/pkg/model8"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHostnameMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Db8Hostname8) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &Db8Hostname8{Db: db}
	return db, mock, repo
}

func TestNewDb8Hostname8(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDb8Hostname8(db)
	assert.NotNil(t, repo)
}

func TestDb8Hostname8_GetAllHostname(t *testing.T) {
	t.Run("returns all hostnames", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		id1 := uuid.Must(uuid.NewV4())
		id2 := uuid.Must(uuid.NewV4())
		domainId := uuid.Must(uuid.NewV4())
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(id1, "sub1.example.com", now, true, domainId, true).
			AddRow(id2, "sub2.example.com", now, false, domainId, true)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname ORDER BY name")).
			WillReturnRows(rows)

		hostnames, err := repo.GetAllHostname()

		assert.NoError(t, err)
		assert.Len(t, hostnames, 2)
		assert.Equal(t, "sub1.example.com", hostnames[0].Name)
		assert.Equal(t, "sub2.example.com", hostnames[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no hostnames", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname ORDER BY name")).
			WillReturnRows(rows)

		hostnames, err := repo.GetAllHostname()

		assert.NoError(t, err)
		assert.Empty(t, hostnames)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on query failure", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname ORDER BY name")).
			WillReturnError(errors.New("query error"))

		hostnames, err := repo.GetAllHostname()

		assert.Error(t, err)
		assert.Empty(t, hostnames)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_GetAllHostnameByDomainid(t *testing.T) {
	t.Run("returns hostnames for domain", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		id := uuid.Must(uuid.NewV4())
		domainId := uuid.Must(uuid.NewV4())
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(id, "sub.example.com", now, true, domainId, true)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE domainid = $1")).
			WithArgs(domainId).
			WillReturnRows(rows)

		hostnames, err := repo.GetAllHostnameByDomainid(domainId)

		assert.NoError(t, err)
		assert.Len(t, hostnames, 1)
		assert.Equal(t, domainId, hostnames[0].Domainid)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_GetAllHostnameIDsByDomainid(t *testing.T) {
	t.Run("returns hostname IDs for domain", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		id1 := uuid.Must(uuid.NewV4())
		id2 := uuid.Must(uuid.NewV4())
		domainId := uuid.Must(uuid.NewV4())

		rows := sqlmock.NewRows([]string{"id"}).
			AddRow(id1).
			AddRow(id2)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM ONLY cptm8hostname WHERE domainid = $1")).
			WithArgs(domainId).
			WillReturnRows(rows)

		ids, err := repo.GetAllHostnameIDsByDomainid(domainId)

		assert.NoError(t, err)
		assert.Len(t, ids, 2)
		assert.Contains(t, ids, id1)
		assert.Contains(t, ids, id2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_GetOneHostnameByIdAndDomainid(t *testing.T) {
	t.Run("returns hostname by id and domain", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		id := uuid.Must(uuid.NewV4())
		domainId := uuid.Must(uuid.NewV4())
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(id, "sub.example.com", now, true, domainId, true)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE id = $1 AND domainid = $2")).
			WithArgs(id, domainId).
			WillReturnRows(rows)

		hostname, err := repo.GetOneHostnameByIdAndDomainid(id, domainId)

		assert.NoError(t, err)
		assert.Equal(t, id, hostname.Id)
		assert.Equal(t, "sub.example.com", hostname.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_GetOneHostnameByName(t *testing.T) {
	t.Run("returns hostname by name", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		id := uuid.Must(uuid.NewV4())
		domainId := uuid.Must(uuid.NewV4())
		now := time.Now()
		name := "sub.example.com"

		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(id, name, now, true, domainId, true)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE name = $1")).
			WithArgs(name).
			WillReturnRows(rows)

		hostname, err := repo.GetOneHostnameByName(name)

		assert.NoError(t, err)
		assert.Equal(t, name, hostname.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_InsertHostname(t *testing.T) {
	t.Run("successful insert", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		domainId := uuid.Must(uuid.NewV4())
		hostnameId := uuid.Must(uuid.NewV4())
		now := time.Now()
		post := model8.PostHostname8{
			Name:    "new.example.com",
			Enabled: true,
			Live:    true,
		}

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO cptm8hostname(name, foundfirsttime, live, domainid) VALUES ($1, NOW(), true, $2) ON CONFLICT DO NOTHING"))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO cptm8hostname(name, foundfirsttime, live, domainid) VALUES ($1, NOW(), true, $2) ON CONFLICT DO NOTHING")).
			WithArgs(post.Name, domainId).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// GetOneHostnameByName call
		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(hostnameId, post.Name, now, true, domainId, true)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE name = $1")).
			WithArgs(post.Name).
			WillReturnRows(rows)

		id, err := repo.InsertHostname(domainId, post)

		assert.NoError(t, err)
		assert.Equal(t, hostnameId, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("begin transaction error", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		domainId := uuid.Must(uuid.NewV4())
		post := model8.PostHostname8{
			Name: "new.example.com",
		}

		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		id, err := repo.InsertHostname(domainId, post)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_InsertBatch(t *testing.T) {
	t.Run("successful batch insert with changes", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		domainId := uuid.Must(uuid.NewV4())
		names := []string{"sub1.example.com", "sub2.example.com"}

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO cptm8hostname(name, foundfirsttime, live, enabled, domainid)
		VALUES ($1, NOW(), true, $2, $3)
		ON CONFLICT (name) DO UPDATE SET
			live = EXCLUDED.live
		WHERE cptm8hostname.live != EXCLUDED.live
	`))

		// First insert
		mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO cptm8hostname(name, foundfirsttime, live, enabled, domainid)
		VALUES ($1, NOW(), true, $2, $3)
		ON CONFLICT (name) DO UPDATE SET
			live = EXCLUDED.live
		WHERE cptm8hostname.live != EXCLUDED.live
	`)).
			WithArgs(names[0], true, domainId).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Second insert
		mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO cptm8hostname(name, foundfirsttime, live, enabled, domainid)
		VALUES ($1, NOW(), true, $2, $3)
		ON CONFLICT (name) DO UPDATE SET
			live = EXCLUDED.live
		WHERE cptm8hostname.live != EXCLUDED.live
	`)).
			WithArgs(names[1], true, domainId).
			WillReturnResult(sqlmock.NewResult(2, 1))

		mock.ExpectCommit()

		changed, err := repo.InsertBatch(domainId, true, names)

		assert.NoError(t, err)
		assert.True(t, changed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("batch insert with no changes", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		domainId := uuid.Must(uuid.NewV4())
		names := []string{"existing.example.com"}

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO cptm8hostname(name, foundfirsttime, live, enabled, domainid)
		VALUES ($1, NOW(), true, $2, $3)
		ON CONFLICT (name) DO UPDATE SET
			live = EXCLUDED.live
		WHERE cptm8hostname.live != EXCLUDED.live
	`))

		mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO cptm8hostname(name, foundfirsttime, live, enabled, domainid)
		VALUES ($1, NOW(), true, $2, $3)
		ON CONFLICT (name) DO UPDATE SET
			live = EXCLUDED.live
		WHERE cptm8hostname.live != EXCLUDED.live
	`)).
			WithArgs(names[0], true, domainId).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

		mock.ExpectCommit()

		changed, err := repo.InsertBatch(domainId, true, names)

		assert.NoError(t, err)
		assert.False(t, changed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("batch insert with empty list", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		domainId := uuid.Must(uuid.NewV4())
		names := []string{}

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO cptm8hostname(name, foundfirsttime, live, enabled, domainid)
		VALUES ($1, NOW(), true, $2, $3)
		ON CONFLICT (name) DO UPDATE SET
			live = EXCLUDED.live
		WHERE cptm8hostname.live != EXCLUDED.live
	`))
		mock.ExpectCommit()

		changed, err := repo.InsertBatch(domainId, true, names)

		assert.NoError(t, err)
		assert.False(t, changed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_UpdateHostname(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		id := uuid.Must(uuid.NewV4())
		domainId := uuid.Must(uuid.NewV4())
		now := time.Now()
		post := model8.PostHostname8{
			Name:    "updated.example.com",
			Enabled: true,
			Live:    true,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE cptm8hostname SET name = $1, enabled = $2, live = $4 WHERE id = $6 AND domainid = $6")).
			WithArgs(post.Name, post.Enabled, post.Live, id, domainId).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// GetOneHostnameByIdAndDomainid call
		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(id, post.Name, now, post.Live, domainId, post.Enabled)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE id = $1 AND domainid = $2")).
			WithArgs(id, domainId).
			WillReturnRows(rows)

		hostname, err := repo.UpdateHostname(domainId, id, post)

		assert.NoError(t, err)
		assert.Equal(t, post.Name, hostname.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_UpdateLiveColumnByParentID(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		domainId := uuid.Must(uuid.NewV4())

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE cptm8hostname SET live = $1 WHERE domainid = $2")).
			WithArgs(false, domainId).
			WillReturnResult(sqlmock.NewResult(0, 5))
		mock.ExpectCommit()

		success, err := repo.UpdateLiveColumnByParentID(domainId, false)

		assert.NoError(t, err)
		assert.True(t, success)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_UpdateLiveColumnByName(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		name := "sub.example.com"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE cptm8hostname SET live = $1 WHERE name = $2")).
			WithArgs(true, name).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		success, err := repo.UpdateLiveColumnByName(name, true)

		assert.NoError(t, err)
		assert.True(t, success)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_DeleteHostnameByID(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		id := uuid.Must(uuid.NewV4())

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ONLY cptm8hostname WHERE id = $1")).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		success, err := repo.DeleteHostnameByID(id)

		assert.NoError(t, err)
		assert.True(t, success)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error with rollback", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		id := uuid.Must(uuid.NewV4())

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ONLY cptm8hostname WHERE id = $1")).
			WithArgs(id).
			WillReturnError(errors.New("delete error"))
		mock.ExpectRollback()

		success, err := repo.DeleteHostnameByID(id)

		assert.Error(t, err)
		assert.False(t, success)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_DeleteAllByParentID(t *testing.T) {
	t.Run("successful delete all by domain", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		domainId := uuid.Must(uuid.NewV4())

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ONLY cptm8hostname WHERE domainid = $1")).
			WithArgs(domainId).
			WillReturnResult(sqlmock.NewResult(0, 10))
		mock.ExpectCommit()

		success, err := repo.DeleteAllByParentID(domainId)

		assert.NoError(t, err)
		assert.True(t, success)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDb8Hostname8_DeleteHostnameByName(t *testing.T) {
	t.Run("successful delete by name", func(t *testing.T) {
		db, mock, repo := setupHostnameMock(t)
		defer db.Close()

		name := "sub.example.com"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ONLY cptm8hostname WHERE name = $1")).
			WithArgs(name).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		success, err := repo.DeleteHostnameByName(name)

		assert.NoError(t, err)
		assert.True(t, success)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
