package log8

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {
	// Note: GetLogger uses sync.Once, so we can only test the first initialization
	// Subsequent calls return the same logger instance

	t.Run("returns a zerolog logger", func(t *testing.T) {
		// Create a temp directory for log files
		tmpDir := t.TempDir()

		// Change to temp dir so log directory is created there
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		logger := GetLogger("test.log")

		// Verify logger is not zero value
		assert.NotEqual(t, zerolog.Logger{}, logger)
	})

	t.Run("BaseLogger is set after GetLogger call", func(t *testing.T) {
		// BaseLogger should be set after GetLogger is called
		// Note: This depends on the previous test having run first due to sync.Once
		assert.NotEqual(t, zerolog.Logger{}, BaseLogger)
	})

	t.Run("logger can write messages", func(t *testing.T) {
		// Test that the logger can be used without panicking
		assert.NotPanics(t, func() {
			BaseLogger.Info().Msg("test message")
			BaseLogger.Debug().Str("key", "value").Msg("debug message")
			BaseLogger.Warn().Int("count", 5).Msg("warning message")
		})
	})
}

func TestLoggerLevels(t *testing.T) {
	// Test that different log levels work
	t.Run("info level logging", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BaseLogger.Info().Msg("info test")
		})
	})

	t.Run("debug level logging", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BaseLogger.Debug().Msg("debug test")
		})
	})

	t.Run("warn level logging", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BaseLogger.Warn().Msg("warn test")
		})
	})

	t.Run("error level logging", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BaseLogger.Error().Msg("error test")
		})
	})
}

func TestLoggerWithFields(t *testing.T) {
	t.Run("logs with string field", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BaseLogger.Info().Str("domain", "example.com").Msg("processing domain")
		})
	})

	t.Run("logs with int field", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BaseLogger.Info().Int("count", 42).Msg("found subdomains")
		})
	})

	t.Run("logs with bool field", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BaseLogger.Info().Bool("enabled", true).Msg("feature status")
		})
	})

	t.Run("logs with multiple fields", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BaseLogger.Info().
				Str("domain", "example.com").
				Int("subdomains", 10).
				Bool("active", true).
				Msg("scan complete")
		})
	})

	t.Run("logs with error field", func(t *testing.T) {
		assert.NotPanics(t, func() {
			err := os.ErrNotExist
			BaseLogger.Error().Err(err).Msg("file not found")
		})
	})
}
