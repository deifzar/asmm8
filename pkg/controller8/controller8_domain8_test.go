package controller8

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupDomainControllerTest(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	controller := NewController8Domain8(db)
	router := gin.New()

	router.POST("/domain", controller.InsertDomain)
	router.GET("/domain", controller.GetAllDomain)
	router.GET("/domain/:id", controller.GetOneDomain)
	router.PUT("/domain/:id", controller.UpdateDomain)
	router.DELETE("/domain/:id", controller.DeleteDomain)

	return router, mock, func() { db.Close() }
}

func TestController8Domain8_InsertDomain(t *testing.T) {
	t.Run("successful insert", func(t *testing.T) {
		router, mock, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO cptm8domain(name, companyname, enabled) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING"))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO cptm8domain(name, companyname, enabled) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING")).
			WithArgs("example.com", "Example Inc", true).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		body := map[string]interface{}{
			"name":        "example.com",
			"companyname": "Example Inc",
			"enabled":     true,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/domain", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("missing required fields", func(t *testing.T) {
		router, _, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		body := map[string]interface{}{
			"name": "example.com",
			// missing companyname
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/domain", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "failed")
	})

	t.Run("invalid json body", func(t *testing.T) {
		router, _, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodPost, "/domain", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestController8Domain8_GetAllDomain(t *testing.T) {
	t.Run("returns all domains", func(t *testing.T) {
		router, mock, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		id1 := uuid.Must(uuid.NewV4())
		id2 := uuid.Must(uuid.NewV4())

		rows := sqlmock.NewRows([]string{"id", "name", "companyname", "enabled"}).
			AddRow(id1, "example.com", "Example Inc", true).
			AddRow(id2, "test.com", "Test Corp", false)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, companyname, enabled FROM cptm8domain")).
			WillReturnRows(rows)

		req := httptest.NewRequest(http.MethodGet, "/domain", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "example.com")
		assert.Contains(t, w.Body.String(), "test.com")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty list", func(t *testing.T) {
		router, mock, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"id", "name", "companyname", "enabled"})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, companyname, enabled FROM cptm8domain")).
			WillReturnRows(rows)

		req := httptest.NewRequest(http.MethodGet, "/domain", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestController8Domain8_GetOneDomain(t *testing.T) {
	t.Run("returns domain by id", func(t *testing.T) {
		router, mock, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		id := uuid.Must(uuid.NewV4())

		rows := sqlmock.NewRows([]string{"id", "name", "companyname", "enabled"}).
			AddRow(id, "example.com", "Example Inc", true)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, companyname, enabled FROM cptm8domain WHERE id = $1")).
			WithArgs(id).
			WillReturnRows(rows)

		req := httptest.NewRequest(http.MethodGet, "/domain/"+id.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "example.com")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid uuid", func(t *testing.T) {
		router, _, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodGet, "/domain/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "failed")
	})
}

func TestController8Domain8_UpdateDomain(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		router, mock, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		id := uuid.Must(uuid.NewV4())

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta("UPDATE cptm8domain SET name = $1, companyname = $2, enabled = $3 WHERE id = $4"))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE cptm8domain SET name = $1, companyname = $2, enabled = $3 WHERE id = $4")).
			WithArgs("updated.com", "Updated Inc", true, id).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		rows := sqlmock.NewRows([]string{"id", "name", "companyname", "enabled"}).
			AddRow(id, "updated.com", "Updated Inc", true)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, companyname, enabled FROM cptm8domain WHERE id = $1")).
			WithArgs(id).
			WillReturnRows(rows)

		body := map[string]interface{}{
			"name":        "updated.com",
			"companyname": "Updated Inc",
			"enabled":     true,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/domain/"+id.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "updated.com")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid uuid", func(t *testing.T) {
		router, _, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		body := map[string]interface{}{
			"name":        "updated.com",
			"companyname": "Updated Inc",
			"enabled":     true,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/domain/invalid-uuid", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required fields", func(t *testing.T) {
		router, _, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		id := uuid.Must(uuid.NewV4())
		body := map[string]interface{}{
			"name": "updated.com",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/domain/"+id.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestController8Domain8_DeleteDomain(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		router, mock, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		id := uuid.Must(uuid.NewV4())

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM cptm8domain WHERE id = $1")).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		req := httptest.NewRequest(http.MethodDelete, "/domain/"+id.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid uuid", func(t *testing.T) {
		router, _, cleanup := setupDomainControllerTest(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodDelete, "/domain/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNewController8Domain8(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	controller := NewController8Domain8(db)
	assert.NotNil(t, controller)
}
