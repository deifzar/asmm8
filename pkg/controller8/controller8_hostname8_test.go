package controller8

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHostnameControllerTest(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	controller := NewController8Hostname8(db)
	router := gin.New()

	router.POST("/domain/:id/hostname", controller.InsertHostname)
	router.GET("/domain/:id/hostname", controller.GetAllHostname)
	router.GET("/domain/:id/hostname/:hostnameid", controller.GetOneHostname)
	router.PUT("/domain/:id/hostname/:hostnameid", controller.UpdateHostname)
	router.DELETE("/domain/:id/hostname/:hostnameid", controller.DeleteHostname)

	return router, mock, func() { db.Close() }
}

func TestController8Hostname8_InsertHostname(t *testing.T) {
	// Note: The InsertHostname function uses Hostname8Uri which expects both :id and :hostnameid
	// but the insert route only provides :id. This test documents the actual behavior.
	// A successful insert test would require fixing the controller to use a different URI struct.

	t.Run("missing required fields", func(t *testing.T) {
		router, _, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		domainId := uuid.Must(uuid.NewV4())

		body := map[string]interface{}{
			"enabled": true,
			// missing name
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/domain/"+domainId.String()+"/hostname", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid domain uuid", func(t *testing.T) {
		router, _, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		body := map[string]interface{}{
			"name": "sub.example.com",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/domain/invalid-uuid/hostname", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestController8Hostname8_GetAllHostname(t *testing.T) {
	t.Run("returns all hostnames", func(t *testing.T) {
		router, mock, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		domainId := uuid.Must(uuid.NewV4())
		id1 := uuid.Must(uuid.NewV4())
		id2 := uuid.Must(uuid.NewV4())
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(id1, "sub1.example.com", now, true, domainId, true).
			AddRow(id2, "sub2.example.com", now, false, domainId, true)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname ORDER BY name")).
			WillReturnRows(rows)

		req := httptest.NewRequest(http.MethodGet, "/domain/"+domainId.String()+"/hostname", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "sub1.example.com")
		assert.Contains(t, w.Body.String(), "sub2.example.com")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestController8Hostname8_GetOneHostname(t *testing.T) {
	t.Run("returns hostname by id", func(t *testing.T) {
		router, mock, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		domainId := uuid.Must(uuid.NewV4())
		hostnameId := uuid.Must(uuid.NewV4())
		now := time.Now()

		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(hostnameId, "sub.example.com", now, true, domainId, true)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE id = $1 AND domainid = $2")).
			WithArgs(hostnameId, domainId).
			WillReturnRows(rows)

		req := httptest.NewRequest(http.MethodGet, "/domain/"+domainId.String()+"/hostname/"+hostnameId.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "sub.example.com")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid domain uuid", func(t *testing.T) {
		router, _, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		hostnameId := uuid.Must(uuid.NewV4())

		req := httptest.NewRequest(http.MethodGet, "/domain/invalid-uuid/hostname/"+hostnameId.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid hostname uuid", func(t *testing.T) {
		router, _, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		domainId := uuid.Must(uuid.NewV4())

		req := httptest.NewRequest(http.MethodGet, "/domain/"+domainId.String()+"/hostname/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestController8Hostname8_UpdateHostname(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		router, mock, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		domainId := uuid.Must(uuid.NewV4())
		hostnameId := uuid.Must(uuid.NewV4())
		now := time.Now()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE cptm8hostname SET name = $1, enabled = $2, live = $4 WHERE id = $6 AND domainid = $6")).
			WithArgs("updated.example.com", true, true, hostnameId, domainId).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		rows := sqlmock.NewRows([]string{"id", "name", "foundfirsttime", "live", "domainid", "enabled"}).
			AddRow(hostnameId, "updated.example.com", now, true, domainId, true)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, foundfirsttime, live, domainid, enabled FROM ONLY cptm8hostname WHERE id = $1 AND domainid = $2")).
			WithArgs(hostnameId, domainId).
			WillReturnRows(rows)

		body := map[string]interface{}{
			"name":    "updated.example.com",
			"enabled": true,
			"live":    true,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/domain/"+domainId.String()+"/hostname/"+hostnameId.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "updated.example.com")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("missing required fields", func(t *testing.T) {
		router, _, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		domainId := uuid.Must(uuid.NewV4())
		hostnameId := uuid.Must(uuid.NewV4())

		body := map[string]interface{}{
			"enabled": true,
			// missing name
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/domain/"+domainId.String()+"/hostname/"+hostnameId.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestController8Hostname8_DeleteHostname(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		router, mock, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		domainId := uuid.Must(uuid.NewV4())
		hostnameId := uuid.Must(uuid.NewV4())

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ONLY cptm8hostname WHERE id = $1")).
			WithArgs(hostnameId).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		req := httptest.NewRequest(http.MethodDelete, "/domain/"+domainId.String()+"/hostname/"+hostnameId.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid hostname uuid", func(t *testing.T) {
		router, _, cleanup := setupHostnameControllerTest(t)
		defer cleanup()

		domainId := uuid.Must(uuid.NewV4())

		req := httptest.NewRequest(http.MethodDelete, "/domain/"+domainId.String()+"/hostname/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNewController8Hostname8(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	controller := NewController8Hostname8(db)
	assert.NotNil(t, controller)
}
