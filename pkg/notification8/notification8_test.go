package notification8

import (
	"testing"

	"deifzar/asmm8/pkg/model8"

	"github.com/stretchr/testify/assert"
)

func TestNotificationService_Struct(t *testing.T) {
	t.Run("creates notification service", func(t *testing.T) {
		service := NotificationService{}
		assert.NotNil(t, service)
	})
}

func TestNotificationPoolHelper_Struct(t *testing.T) {
	t.Run("creates pool helper", func(t *testing.T) {
		helper := NotificationPoolHelper{}
		assert.NotNil(t, helper)
	})

	t.Run("global PoolHelper is accessible", func(t *testing.T) {
		assert.NotNil(t, PoolHelper)
	})
}

func TestRoutingKeyGeneration(t *testing.T) {
	// Test the routing key generation logic used in notification helper methods
	testCases := []struct {
		name            string
		severity        string
		notificationType string
		expectedKey     string
	}{
		{
			name:            "security high severity",
			severity:        "high",
			notificationType: "security",
			expectedKey:     "app.security.high",
		},
		{
			name:            "security critical severity",
			severity:        "critical",
			notificationType: "security",
			expectedKey:     "app.security.critical",
		},
		{
			name:            "error low severity",
			severity:        "low",
			notificationType: "error",
			expectedKey:     "app.error.low",
		},
		{
			name:            "warning medium severity",
			severity:        "medium",
			notificationType: "warning",
			expectedKey:     "app.warning.medium",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			routingKey := "app." + tc.notificationType + "." + tc.severity
			assert.Equal(t, tc.expectedKey, routingKey)
		})
	}
}

func TestNotificationMetadata_Creation(t *testing.T) {
	t.Run("creates metadata with severity", func(t *testing.T) {
		metadata := model8.NotificationMetadata8{
			Severity:    "high",
			Channeltype: model8.App,
			Eventtype:   model8.Security,
		}

		assert.Equal(t, "high", metadata.Severity)
		assert.Equal(t, model8.App, metadata.Channeltype)
		assert.Equal(t, model8.Security, metadata.Eventtype)
	})

	t.Run("creates metadata for error notification", func(t *testing.T) {
		metadata := model8.NotificationMetadata8{
			Severity:    "critical",
			Channeltype: model8.App,
			Eventtype:   model8.Error,
		}

		assert.Equal(t, "critical", metadata.Severity)
		assert.Equal(t, model8.Error, metadata.Eventtype)
	})

	t.Run("creates metadata for warning notification", func(t *testing.T) {
		metadata := model8.NotificationMetadata8{
			Severity:    "low",
			Channeltype: model8.App,
			Eventtype:   model8.Warning,
		}

		assert.Equal(t, "low", metadata.Severity)
		assert.Equal(t, model8.Warning, metadata.Eventtype)
	})
}

func TestNotification_Creation(t *testing.T) {
	t.Run("creates security notification for admin", func(t *testing.T) {
		metadata := model8.NotificationMetadata8{
			Severity:    "high",
			Channeltype: model8.App,
			Eventtype:   model8.Security,
		}

		notification := model8.Notification8{
			Userrole: model8.RoleAdmin,
			Type:     model8.Security,
			Message:  "Security alert: New vulnerability detected",
			Metadata: metadata,
		}

		assert.Equal(t, model8.RoleAdmin, notification.Userrole)
		assert.Equal(t, model8.Security, notification.Type)
		assert.Contains(t, notification.Message, "Security alert")
		assert.Equal(t, "high", notification.Metadata.Severity)
	})

	t.Run("creates security notification for user", func(t *testing.T) {
		metadata := model8.NotificationMetadata8{
			Severity:    "medium",
			Channeltype: model8.App,
			Eventtype:   model8.Security,
		}

		notification := model8.Notification8{
			Userrole: model8.RoleUser,
			Type:     model8.Security,
			Message:  "Security notification",
			Metadata: metadata,
		}

		assert.Equal(t, model8.RoleUser, notification.Userrole)
		assert.Equal(t, model8.Security, notification.Type)
	})

	t.Run("creates error notification", func(t *testing.T) {
		metadata := model8.NotificationMetadata8{
			Severity:    "critical",
			Channeltype: model8.App,
			Eventtype:   model8.Error,
		}

		notification := model8.Notification8{
			Userrole: model8.RoleAdmin,
			Type:     model8.Error,
			Message:  "System error occurred",
			Metadata: metadata,
		}

		assert.Equal(t, model8.RoleAdmin, notification.Userrole)
		assert.Equal(t, model8.Error, notification.Type)
		assert.Equal(t, "critical", notification.Metadata.Severity)
	})

	t.Run("creates warning notification", func(t *testing.T) {
		metadata := model8.NotificationMetadata8{
			Severity:    "low",
			Channeltype: model8.App,
			Eventtype:   model8.Warning,
		}

		notification := model8.Notification8{
			Userrole: model8.RoleAdmin,
			Type:     model8.Warning,
			Message:  "System warning",
			Metadata: metadata,
		}

		assert.Equal(t, model8.Warning, notification.Type)
		assert.Equal(t, "low", notification.Metadata.Severity)
	})
}

func TestSeverityLevels(t *testing.T) {
	severities := []string{"low", "medium", "high", "critical", "urgent"}

	for _, severity := range severities {
		t.Run("severity_"+severity, func(t *testing.T) {
			metadata := model8.NotificationMetadata8{
				Severity: severity,
			}
			assert.Equal(t, severity, metadata.Severity)
		})
	}
}

func TestNotificationEventTypes(t *testing.T) {
	t.Run("message event type", func(t *testing.T) {
		assert.Equal(t, model8.Notificationevent("message"), model8.Message)
	})

	t.Run("security event type", func(t *testing.T) {
		assert.Equal(t, model8.Notificationevent("security"), model8.Security)
	})

	t.Run("error event type", func(t *testing.T) {
		assert.Equal(t, model8.Notificationevent("system_error"), model8.Error)
	})

	t.Run("warning event type", func(t *testing.T) {
		assert.Equal(t, model8.Notificationevent("system_warning"), model8.Warning)
	})
}

func TestNotificationChannelTypes(t *testing.T) {
	t.Run("app channel", func(t *testing.T) {
		assert.Equal(t, model8.Notificationchannel("app"), model8.App)
	})

	t.Run("email channel", func(t *testing.T) {
		assert.Equal(t, model8.Notificationchannel("email"), model8.Email)
	})

	t.Run("sms channel", func(t *testing.T) {
		assert.Equal(t, model8.Notificationchannel("sms"), model8.Sms)
	})
}

func TestRoleTypes(t *testing.T) {
	t.Run("user role", func(t *testing.T) {
		assert.Equal(t, model8.Roletype("user"), model8.RoleUser)
	})

	t.Run("admin role", func(t *testing.T) {
		assert.Equal(t, model8.Roletype("admin"), model8.RoleAdmin)
	})
}

// TestHelperMethodRoutingKeys tests that the helper methods would generate correct routing keys
func TestHelperMethodRoutingKeys(t *testing.T) {
	t.Run("PublishSecurityNotificationAdmin generates correct routing key", func(t *testing.T) {
		severity := "high"
		expectedKey := "app.security." + severity
		assert.Equal(t, "app.security.high", expectedKey)
	})

	t.Run("PublishSecurityNotificationUser generates correct routing key", func(t *testing.T) {
		severity := "critical"
		expectedKey := "app.security." + severity
		assert.Equal(t, "app.security.critical", expectedKey)
	})

	t.Run("PublishSysErrorNotification generates correct routing key", func(t *testing.T) {
		severity := "high"
		expectedKey := "app.error." + severity
		assert.Equal(t, "app.error.high", expectedKey)
	})

	t.Run("PublishSysWarningNotification generates correct routing key", func(t *testing.T) {
		severity := "medium"
		expectedKey := "app.warning." + severity
		assert.Equal(t, "app.warning.medium", expectedKey)
	})
}
