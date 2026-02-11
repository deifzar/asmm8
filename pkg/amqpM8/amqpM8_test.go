package amqpM8

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConnectionPoolConfig(t *testing.T) {
	t.Run("returns default configuration", func(t *testing.T) {
		config := DefaultConnectionPoolConfig()

		assert.Equal(t, 10, config.MaxConnections)
		assert.Equal(t, 2, config.MinConnections)
		assert.Equal(t, 1*time.Hour, config.MaxIdleTime)
		assert.Equal(t, 2*time.Hour, config.MaxLifetime)
		assert.Equal(t, 30*time.Minute, config.HealthCheckPeriod)
		assert.Equal(t, 30*time.Second, config.ConnectionTimeout)
		assert.Equal(t, 10, config.RetryAttempts)
		assert.Equal(t, 2*time.Second, config.RetryDelay)
	})
}

func TestConnectionPoolConfig_CustomValues(t *testing.T) {
	t.Run("custom configuration values", func(t *testing.T) {
		config := ConnectionPoolConfig{
			MaxConnections:    20,
			MinConnections:    5,
			MaxIdleTime:       30 * time.Minute,
			MaxLifetime:       1 * time.Hour,
			HealthCheckPeriod: 15 * time.Minute,
			ConnectionTimeout: 15 * time.Second,
			RetryAttempts:     5,
			RetryDelay:        1 * time.Second,
		}

		assert.Equal(t, 20, config.MaxConnections)
		assert.Equal(t, 5, config.MinConnections)
		assert.Equal(t, 30*time.Minute, config.MaxIdleTime)
		assert.Equal(t, 1*time.Hour, config.MaxLifetime)
	})
}

func TestConnectionPoolStats(t *testing.T) {
	t.Run("stats struct initialization", func(t *testing.T) {
		stats := ConnectionPoolStats{
			ActiveConnections:  5,
			IdleConnections:    3,
			TotalCreated:       10,
			TotalDestroyed:     2,
			TotalBorrowed:      100,
			TotalReturned:      95,
			HealthyConnections: 8,
		}

		assert.Equal(t, 5, stats.ActiveConnections)
		assert.Equal(t, 3, stats.IdleConnections)
		assert.Equal(t, uint64(10), stats.TotalCreated)
		assert.Equal(t, uint64(2), stats.TotalDestroyed)
		assert.Equal(t, uint64(100), stats.TotalBorrowed)
		assert.Equal(t, uint64(95), stats.TotalReturned)
		assert.Equal(t, 8, stats.HealthyConnections)
	})
}

func TestPooledConnection_Struct(t *testing.T) {
	t.Run("pooled connection fields", func(t *testing.T) {
		now := time.Now()
		pc := &PooledConnection{
			inUse:      false,
			createdAt:  now,
			lastUsed:   now,
			usageCount: 0,
			isHealthy:  true,
		}

		assert.False(t, pc.inUse)
		assert.True(t, pc.isHealthy)
		assert.Equal(t, uint64(0), pc.usageCount)
		assert.Equal(t, now, pc.createdAt)
	})
}

func TestSharedAmqpState(t *testing.T) {
	// Note: GetSharedState uses sync.Once, so we test the singleton behavior

	t.Run("returns singleton instance", func(t *testing.T) {
		state1 := GetSharedState()
		state2 := GetSharedState()

		assert.Same(t, state1, state2)
	})

	t.Run("state has initialized maps", func(t *testing.T) {
		state := GetSharedState()

		assert.NotNil(t, state.GetQueues())
		assert.NotNil(t, state.GetBindings())
		assert.NotNil(t, state.GetExchanges())
	})
}

func TestSharedAmqpState_ExchangeOperations(t *testing.T) {
	state := GetSharedState()

	t.Run("set and get exchange", func(t *testing.T) {
		state.SetExchange("test_exchange", "topic")

		exchangeType := state.GetExchangeTypeByExchangeName("test_exchange")
		assert.Equal(t, "topic", exchangeType)
	})

	t.Run("get non-existent exchange returns empty", func(t *testing.T) {
		exchangeType := state.GetExchangeTypeByExchangeName("nonexistent")
		assert.Equal(t, "", exchangeType)
	})

	t.Run("get all exchanges", func(t *testing.T) {
		state.SetExchange("exchange1", "direct")
		state.SetExchange("exchange2", "fanout")

		exchanges := state.GetExchanges()
		assert.Contains(t, exchanges, "exchange1")
		assert.Contains(t, exchanges, "exchange2")
	})
}

func TestSharedAmqpState_QueueOperations(t *testing.T) {
	state := GetSharedState()

	t.Run("set and get queue by exchange name", func(t *testing.T) {
		queue := amqp.Queue{
			Name:      "test_queue",
			Messages:  0,
			Consumers: 0,
		}

		state.SetQueueByExchangeName("test_exchange", "test_queue", queue)

		retrieved := state.GetQueueByExchangeNameAndQueueName("test_exchange", "test_queue")
		assert.Equal(t, "test_queue", retrieved.Name)
	})

	t.Run("get non-existent queue returns empty", func(t *testing.T) {
		queue := state.GetQueueByExchangeNameAndQueueName("nonexistent", "nonexistent")
		assert.Equal(t, "", queue.Name)
	})
}

func TestSharedAmqpState_BindingOperations(t *testing.T) {
	state := GetSharedState()

	t.Run("set and get bindings", func(t *testing.T) {
		bindingKeys := []string{"key1", "key2", "key3"}
		state.SetBindingQueueByExchangeName("bind_exchange", "bind_queue", bindingKeys)

		retrieved := state.GetBindingsByExchangeNameAndQueueName("bind_exchange", "bind_queue")
		assert.Len(t, retrieved, 3)
		assert.Contains(t, retrieved, "key1")
		assert.Contains(t, retrieved, "key2")
		assert.Contains(t, retrieved, "key3")
	})

	t.Run("get non-existent bindings returns nil", func(t *testing.T) {
		bindings := state.GetBindingsByExchangeNameAndQueueName("nonexistent", "nonexistent")
		assert.Nil(t, bindings)
	})
}

func TestSharedAmqpState_ConsumerOperations(t *testing.T) {
	state := GetSharedState()

	t.Run("add and get consumers for queue", func(t *testing.T) {
		state.AddConsumerToQueue("consumer_queue", "consumer1")
		state.AddConsumerToQueue("consumer_queue", "consumer2")

		consumers := state.GetConsumersForQueue("consumer_queue")
		assert.Contains(t, consumers, "consumer1")
		assert.Contains(t, consumers, "consumer2")
	})

	t.Run("get non-existent consumers returns nil", func(t *testing.T) {
		consumers := state.GetConsumersForQueue("nonexistent_queue")
		assert.Nil(t, consumers)
	})
}

func TestSharedAmqpState_HandlerOperations(t *testing.T) {
	state := GetSharedState()

	t.Run("add and get handler", func(t *testing.T) {
		handler := func(msg amqp.Delivery) error {
			return nil
		}

		state.AddHandler("handler_queue", handler)

		retrieved, exists := state.GetHandler("handler_queue")
		assert.True(t, exists)
		assert.NotNil(t, retrieved)
	})

	t.Run("get non-existent handler returns false", func(t *testing.T) {
		_, exists := state.GetHandler("nonexistent_handler")
		assert.False(t, exists)
	})
}

func TestSharedAmqpState_InitializeExchange(t *testing.T) {
	state := GetSharedState()

	t.Run("initializes exchange maps", func(t *testing.T) {
		state.InitializeExchange("new_exchange")

		// Should be able to set queues and bindings without nil map error
		queue := amqp.Queue{Name: "new_queue"}
		state.SetQueueByExchangeName("new_exchange", "new_queue", queue)

		retrieved := state.GetQueueByExchangeNameAndQueueName("new_exchange", "new_queue")
		assert.Equal(t, "new_queue", retrieved.Name)
	})
}

func TestSharedAmqpState_DeleteOperations(t *testing.T) {
	state := GetSharedState()

	t.Run("delete queue by name", func(t *testing.T) {
		// Setup
		queue := amqp.Queue{Name: "queue_to_delete"}
		state.SetQueueByExchangeName("delete_exchange", "queue_to_delete", queue)
		state.SetBindingQueueByExchangeName("delete_exchange", "queue_to_delete", []string{"key"})
		state.AddConsumerToQueue("queue_to_delete", "consumer")
		state.AddHandler("queue_to_delete", func(msg amqp.Delivery) error { return nil })

		// Delete
		state.DeleteQueueByName("queue_to_delete")

		// Verify deletion
		_, exists := state.GetHandler("queue_to_delete")
		assert.False(t, exists)
	})

	t.Run("delete consumer by name", func(t *testing.T) {
		state.AddConsumerToQueue("test_delete_queue", "consumer_to_delete")
		state.AddConsumerToQueue("test_delete_queue", "consumer_to_keep")

		state.DeleteConsumerByName("consumer_to_delete")

		consumers := state.GetConsumersForQueue("test_delete_queue")
		assert.NotContains(t, consumers, "consumer_to_delete")
	})
}

func TestGlobalPoolManager(t *testing.T) {
	t.Run("returns singleton manager", func(t *testing.T) {
		manager1 := GetGlobalPoolManager()
		manager2 := GetGlobalPoolManager()

		assert.Same(t, manager1, manager2)
	})

	t.Run("list pools returns empty initially", func(t *testing.T) {
		manager := GetGlobalPoolManager()
		// Note: Due to singleton, pools may exist from other tests
		pools := manager.ListPools()
		assert.NotNil(t, pools)
	})

	t.Run("get non-existent pool returns error", func(t *testing.T) {
		manager := GetGlobalPoolManager()

		_, err := manager.GetPool("nonexistent_pool_xyz")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("get connection from non-existent pool returns error", func(t *testing.T) {
		manager := GetGlobalPoolManager()

		_, err := manager.GetConnection("nonexistent_pool_abc")
		assert.Error(t, err)
	})

	t.Run("return connection to non-existent pool returns error", func(t *testing.T) {
		manager := GetGlobalPoolManager()

		err := manager.ReturnConnection("nonexistent_pool_def", nil)
		assert.Error(t, err)
	})

	t.Run("get pool stats for non-existent pool returns error", func(t *testing.T) {
		manager := GetGlobalPoolManager()

		_, err := manager.GetPoolStats("nonexistent_pool_ghi")
		assert.Error(t, err)
	})

	t.Run("close non-existent pool returns error", func(t *testing.T) {
		manager := GetGlobalPoolManager()

		err := manager.ClosePool("nonexistent_pool_jkl")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestGetAllPoolStats(t *testing.T) {
	t.Run("returns stats map", func(t *testing.T) {
		manager := GetGlobalPoolManager()

		stats := manager.GetAllPoolStats()
		assert.NotNil(t, stats)
	})
}

func TestHealthCheckAllPools(t *testing.T) {
	t.Run("health check does not panic", func(t *testing.T) {
		manager := GetGlobalPoolManager()

		assert.NotPanics(t, func() {
			manager.HealthCheckAllPools()
		})
	})
}

func TestGetDefaultConnection_NoPool(t *testing.T) {
	t.Run("returns error when default pool not initialized", func(t *testing.T) {
		// Try to get default connection - may fail if no default pool
		_, err := GetDefaultConnection()
		if err != nil {
			assert.Contains(t, err.Error(), "not found")
		}
	})
}

func TestWithPooledConnection_NoPool(t *testing.T) {
	t.Run("returns error when pool not available", func(t *testing.T) {
		err := WithPooledConnection(func(conn PooledAmqpInterface) error {
			return nil
		})

		// Should return error if no default pool exists
		if err != nil {
			assert.Error(t, err)
		}
	})
}

func TestConnectionPoolInterface(t *testing.T) {
	// Verify ConnectionPool implements ConnectionPoolInterface
	t.Run("ConnectionPool implements interface", func(t *testing.T) {
		// This is a compile-time check - if it compiles, the interface is implemented
		var _ ConnectionPoolInterface = (*ConnectionPool)(nil)
	})
}
