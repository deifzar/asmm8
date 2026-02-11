package orchestrator8

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig(t *testing.T) (*viper.Viper, func()) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configContent := `
APP_ENV: TEST
LOG_LEVEL: "1"
RabbitMQ:
  location: localhost
  port: 5672
  username: guest
  password: guest
ORCHESTRATORM8:
  Services:
    asmm8: http://127.0.0.1:8000
    naabum8: http://127.0.0.1:8001
  Exchanges:
    cptm8: topic
    notification: topic
  asmm8:
    Queue:
      - cptm8
      - qasmm8
      - "1"
    Routing-keys:
      - cptm8.asmm8.#
    Queue-arguments:
      x-max-length: 1
      x-overflow: reject-publish
    Consumer:
      - qasmm8
      - casmm8
      - "false"
    Publisher:
      - cptm8
      - cptm8.naabum8.get.scan
      - asmm8
`
	configPath := filepath.Join(configDir, "configuration.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)

	v := viper.New()
	v.AddConfigPath("./configs")
	v.SetConfigType("yaml")
	v.SetConfigName("configuration")
	err = v.ReadInConfig()
	require.NoError(t, err)

	cleanup := func() {
		os.Chdir(originalDir)
	}

	return v, cleanup
}

func TestOrchestrator8_Struct(t *testing.T) {
	t.Run("creates orchestrator with config", func(t *testing.T) {
		v, cleanup := setupTestConfig(t)
		defer cleanup()

		o := &Orchestrator8{Config: v}

		assert.NotNil(t, o.Config)
		assert.Equal(t, "TEST", o.Config.GetString("APP_ENV"))
	})
}

func TestOrchestrator8_ConfigReading(t *testing.T) {
	v, cleanup := setupTestConfig(t)
	defer cleanup()

	o := &Orchestrator8{Config: v}

	t.Run("reads services from config", func(t *testing.T) {
		services := o.Config.GetStringMapString("ORCHESTRATORM8.Services")
		assert.Equal(t, "http://127.0.0.1:8000", services["asmm8"])
		assert.Equal(t, "http://127.0.0.1:8001", services["naabum8"])
	})

	t.Run("reads exchanges from config", func(t *testing.T) {
		exchanges := o.Config.GetStringMapString("ORCHESTRATORM8.Exchanges")
		assert.Equal(t, "topic", exchanges["cptm8"])
		assert.Equal(t, "topic", exchanges["notification"])
	})

	t.Run("reads queue config for service", func(t *testing.T) {
		queue := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Queue")
		assert.Len(t, queue, 3)
		assert.Equal(t, "cptm8", queue[0])
		assert.Equal(t, "qasmm8", queue[1])
		assert.Equal(t, "1", queue[2])
	})

	t.Run("reads routing keys for service", func(t *testing.T) {
		routingKeys := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Routing-keys")
		assert.Len(t, routingKeys, 1)
		assert.Equal(t, "cptm8.asmm8.#", routingKeys[0])
	})

	t.Run("reads consumer config for service", func(t *testing.T) {
		consumer := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Consumer")
		assert.Len(t, consumer, 3)
		assert.Equal(t, "qasmm8", consumer[0])
		assert.Equal(t, "casmm8", consumer[1])
		assert.Equal(t, "false", consumer[2])
	})

	t.Run("reads publisher config for service", func(t *testing.T) {
		publisher := o.Config.GetStringSlice("ORCHESTRATORM8.asmm8.Publisher")
		assert.Len(t, publisher, 3)
		assert.Equal(t, "cptm8", publisher[0])
		assert.Equal(t, "cptm8.naabum8.get.scan", publisher[1])
		assert.Equal(t, "asmm8", publisher[2])
	})
}

func TestOrchestrator8_PublishToExchange_Validation(t *testing.T) {
	v, cleanup := setupTestConfig(t)
	defer cleanup()

	o := &Orchestrator8{Config: v}

	t.Run("returns error when exchange is empty", func(t *testing.T) {
		err := o.PublishToExchange("", "routing.key", "payload", "source")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Missing parameters")
	})

	t.Run("returns error when routing key is empty", func(t *testing.T) {
		err := o.PublishToExchange("exchange", "", "payload", "source")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Missing parameters")
	})

	t.Run("returns error when both are empty", func(t *testing.T) {
		err := o.PublishToExchange("", "", "payload", "source")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Missing parameters")
	})
}

func TestOrchestrator8Interface(t *testing.T) {
	// Test that Orchestrator8 implements Orchestrator8Interface
	t.Run("implements interface", func(t *testing.T) {
		v, cleanup := setupTestConfig(t)
		defer cleanup()

		var _ Orchestrator8Interface = &Orchestrator8{Config: v}
	})
}

func TestRoutingKeyParsing(t *testing.T) {
	// Test the routing key parsing logic used in CreateHandleAPICallByService
	testCases := []struct {
		name        string
		routingKey  string
		service     string
		httpMethod  string
		endpoint    string
	}{
		{
			name:       "get scan endpoint",
			routingKey: "cptm8.asmm8.get.scan",
			service:    "asmm8",
			httpMethod: "get",
			endpoint:   "scan",
		},
		{
			name:       "post scan endpoint",
			routingKey: "cptm8.asmm8.post.scan",
			service:    "asmm8",
			httpMethod: "post",
			endpoint:   "scan",
		},
		{
			name:       "naabum8 service",
			routingKey: "cptm8.naabum8.get.domain",
			service:    "naabum8",
			httpMethod: "get",
			endpoint:   "domain",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse routing key like the handler does
			parts := splitRoutingKey(tc.routingKey)
			require.Len(t, parts, 4)

			assert.Equal(t, tc.service, parts[1])
			assert.Equal(t, tc.httpMethod, parts[2])
			assert.Equal(t, tc.endpoint, parts[3])
		})
	}
}

// Helper function that mimics the routing key parsing in the handler
func splitRoutingKey(routingKey string) []string {
	result := []string{}
	current := ""
	for _, c := range routingKey {
		if c == '.' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func TestQueueArgumentsParsing(t *testing.T) {
	v, cleanup := setupTestConfig(t)
	defer cleanup()

	o := &Orchestrator8{Config: v}

	t.Run("parses queue arguments", func(t *testing.T) {
		qargs := o.Config.GetStringMap("ORCHESTRATORM8.asmm8.Queue-arguments")
		assert.NotNil(t, qargs)
		// x-max-length should be present
		_, exists := qargs["x-max-length"]
		assert.True(t, exists)
	})
}
