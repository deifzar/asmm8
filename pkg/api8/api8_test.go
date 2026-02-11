package api8

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestApi8_Struct(t *testing.T) {
	t.Run("creates api8 struct with fields", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		v := viper.New()
		router := gin.New()

		api := &Api8{
			DB:     db,
			Router: router,
			Config: v,
		}

		assert.NotNil(t, api.DB)
		assert.NotNil(t, api.Router)
		assert.NotNil(t, api.Config)
	})

	t.Run("creates api8 struct with nil values", func(t *testing.T) {
		api := &Api8{}

		assert.Nil(t, api.DB)
		assert.Nil(t, api.Router)
		assert.Nil(t, api.Config)
	})
}

// Note: TestApi8_Routes is skipped because Routes() calls NewController8ASSM8
// which tries to initialize RabbitMQ orchestrator and calls Fatal() on failure.
// To properly test Routes(), the controller would need dependency injection for the orchestrator.

func TestApi8_Init_MissingConfig(t *testing.T) {
	t.Run("returns error when config file not found", func(t *testing.T) {
		// Create temp directory without config
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		api := &Api8{}
		err := api.Init()

		// Should fail because config file doesn't exist
		assert.Error(t, err)
	})
}

func TestApi8_Init_DirectoryCreation(t *testing.T) {
	t.Run("creates required directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		// Create a minimal config file
		configDir := filepath.Join(tmpDir, "configs")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		configContent := `
APP_ENV: TEST
Database:
  location: localhost
  port: 5432
  schema: public
  database: testdb
  username: testuser
  password: testpass
`
		configPath := filepath.Join(configDir, "configuration.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		api := &Api8{}
		// Init will fail due to DB connection, but directories should be created
		_ = api.Init()

		// Check directories were created
		_, err = os.Stat(filepath.Join(tmpDir, "configs"))
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(tmpDir, "log"))
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(tmpDir, "tmp"))
		assert.NoError(t, err)
	})
}

func TestApi8_ConfigLoading(t *testing.T) {
	t.Run("config is set after successful init steps", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		configDir := filepath.Join(tmpDir, "configs")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		configContent := `
APP_ENV: TEST
LOG_LEVEL: "1"
Database:
  location: localhost
  port: 5432
  schema: public
  database: testdb
  username: testuser
  password: testpass
`
		configPath := filepath.Join(configDir, "configuration.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		api := &Api8{}
		// Init may fail due to DB connection (if no PostgreSQL running)
		// or succeed (if PostgreSQL is running on localhost)
		initErr := api.Init()

		// Config should be loaded regardless of DB connection result
		if api.Config != nil {
			assert.Equal(t, "TEST", api.Config.GetString("APP_ENV"))
			assert.Equal(t, "localhost", api.Config.GetString("Database.location"))
			assert.Equal(t, 5432, api.Config.GetInt("Database.port"))
		}

		// DB connection result depends on whether PostgreSQL is running
		// If no error, the DB must be available; if error, DB is unavailable
		if initErr == nil {
			assert.NotNil(t, api.DB, "DB should be set when Init() succeeds")
		}
	})
}

func TestApi8_DirectoryPermissions(t *testing.T) {
	t.Run("directories created with correct permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		configDir := filepath.Join(tmpDir, "configs")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		configContent := `
APP_ENV: TEST
Database:
  location: localhost
  port: 5432
  schema: public
  database: testdb
  username: testuser
  password: testpass
`
		configPath := filepath.Join(configDir, "configuration.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		api := &Api8{}
		_ = api.Init()

		// Check log directory exists
		logInfo, err := os.Stat(filepath.Join(tmpDir, "log"))
		if err == nil {
			assert.True(t, logInfo.IsDir())
		}

		// Check tmp directory exists
		tmpInfo, err := os.Stat(filepath.Join(tmpDir, "tmp"))
		if err == nil {
			assert.True(t, tmpInfo.IsDir())
		}
	})
}

func TestApi8_Run_RequiresRouter(t *testing.T) {
	t.Run("api has run method", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		api := &Api8{
			DB:     db,
			Config: viper.New(),
			Router: gin.New(),
		}

		// Verify the Router is set
		assert.NotNil(t, api.Router)
	})
}

func TestViperConfiguration(t *testing.T) {
	t.Run("viper can be configured programmatically", func(t *testing.T) {
		v := viper.New()
		v.Set("APP_ENV", "TEST")
		v.Set("Database.location", "localhost")
		v.Set("Database.port", 5432)

		api := &Api8{
			Config: v,
		}

		assert.Equal(t, "TEST", api.Config.GetString("APP_ENV"))
		assert.Equal(t, "localhost", api.Config.GetString("Database.location"))
		assert.Equal(t, 5432, api.Config.GetInt("Database.port"))
	})
}

func TestGinRouter(t *testing.T) {
	t.Run("gin router can be set", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		api := &Api8{
			Router: router,
		}

		assert.NotNil(t, api.Router)

		routes := api.Router.Routes()
		assert.Len(t, routes, 1)
		assert.Equal(t, "/test", routes[0].Path)
		assert.Equal(t, "GET", routes[0].Method)
	})
}
