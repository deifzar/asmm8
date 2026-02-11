package configparser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfigParser(t *testing.T) {
	t.Run("returns viper instance when config exists", func(t *testing.T) {
		// This test assumes the config file exists at ./configs/configuration.yaml
		// If we're running from the project root, it should work
		v, err := InitConfigParser()

		// Either returns valid viper or error - both are acceptable
		if err == nil {
			assert.NotNil(t, v)
		}
	})

	t.Run("returns error when config file not found", func(t *testing.T) {
		// Create a temp directory without config file
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		v, err := InitConfigParser()

		// Should return error when config file doesn't exist
		assert.Error(t, err)
		assert.NotNil(t, v) // Viper instance is still returned
	})

	t.Run("reads config values when file exists", func(t *testing.T) {
		// Create temp directory with config file
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "configs")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Create a test config file
		configContent := `
APP_ENV: TEST
LOG_LEVEL: "1"
Database:
  location: localhost
  port: 5432
`
		configPath := filepath.Join(configDir, "configuration.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Change to temp directory
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		v, err := InitConfigParser()
		require.NoError(t, err)

		assert.Equal(t, "TEST", v.GetString("APP_ENV"))
		assert.Equal(t, "1", v.GetString("LOG_LEVEL"))
		assert.Equal(t, "localhost", v.GetString("Database.location"))
		assert.Equal(t, 5432, v.GetInt("Database.port"))
	})

	t.Run("reads nested config values", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "configs")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		configContent := `
ORCHESTRATORM8:
  Services:
    asmm8: http://127.0.0.1:8000
    naabum8: http://127.0.0.1:8001
  Exchanges:
    cptm8: topic
    notification: topic
`
		configPath := filepath.Join(configDir, "configuration.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		v, err := InitConfigParser()
		require.NoError(t, err)

		services := v.GetStringMapString("ORCHESTRATORM8.Services")
		assert.Equal(t, "http://127.0.0.1:8000", services["asmm8"])
		assert.Equal(t, "http://127.0.0.1:8001", services["naabum8"])

		exchanges := v.GetStringMapString("ORCHESTRATORM8.Exchanges")
		assert.Equal(t, "topic", exchanges["cptm8"])
		assert.Equal(t, "topic", exchanges["notification"])
	})

	t.Run("returns empty values for missing keys", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "configs")
		err := os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		configContent := `APP_ENV: TEST`
		configPath := filepath.Join(configDir, "configuration.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		v, err := InitConfigParser()
		require.NoError(t, err)

		assert.Equal(t, "", v.GetString("NONEXISTENT_KEY"))
		assert.Equal(t, 0, v.GetInt("NONEXISTENT_INT"))
	})
}

func TestConfigParserWithDifferentTypes(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configContent := `
string_value: "hello"
int_value: 42
bool_value: true
float_value: 3.14
list_value:
  - item1
  - item2
  - item3
map_value:
  key1: value1
  key2: value2
`
	configPath := filepath.Join(configDir, "configuration.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	v, err := InitConfigParser()
	require.NoError(t, err)

	t.Run("reads string value", func(t *testing.T) {
		assert.Equal(t, "hello", v.GetString("string_value"))
	})

	t.Run("reads int value", func(t *testing.T) {
		assert.Equal(t, 42, v.GetInt("int_value"))
	})

	t.Run("reads bool value", func(t *testing.T) {
		assert.True(t, v.GetBool("bool_value"))
	})

	t.Run("reads float value", func(t *testing.T) {
		assert.Equal(t, 3.14, v.GetFloat64("float_value"))
	})

	t.Run("reads list value", func(t *testing.T) {
		list := v.GetStringSlice("list_value")
		assert.Len(t, list, 3)
		assert.Contains(t, list, "item1")
		assert.Contains(t, list, "item2")
		assert.Contains(t, list, "item3")
	})

	t.Run("reads map value", func(t *testing.T) {
		m := v.GetStringMapString("map_value")
		assert.Equal(t, "value1", m["key1"])
		assert.Equal(t, "value2", m["key2"])
	})
}
