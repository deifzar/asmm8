package model8

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomain8_JSONMarshaling(t *testing.T) {
	id := uuid.Must(uuid.NewV4())
	domain := Domain8{
		Id:          id,
		Name:        "example.com",
		Companyname: "Example Inc",
		Enabled:     true,
	}

	jsonBytes, err := json.Marshal(domain)
	require.NoError(t, err)

	var decoded Domain8
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, domain.Id, decoded.Id)
	assert.Equal(t, domain.Name, decoded.Name)
	assert.Equal(t, domain.Companyname, decoded.Companyname)
	assert.Equal(t, domain.Enabled, decoded.Enabled)
}

func TestPostDomain8_JSONMarshaling(t *testing.T) {
	post := PostDomain8{
		Name:        "example.com",
		Companyname: "Example Inc",
		Enabled:     true,
	}

	jsonBytes, err := json.Marshal(post)
	require.NoError(t, err)

	var decoded PostDomain8
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, post.Name, decoded.Name)
	assert.Equal(t, post.Companyname, decoded.Companyname)
	assert.Equal(t, post.Enabled, decoded.Enabled)
}

func TestHostname8_JSONMarshaling(t *testing.T) {
	id := uuid.Must(uuid.NewV4())
	domainId := uuid.Must(uuid.NewV4())
	now := time.Now().Truncate(time.Second) // Truncate for comparison

	hostname := Hostname8{
		Id:             id,
		Name:           "sub.example.com",
		Foundfirsttime: now,
		Live:           true,
		Domainid:       domainId,
		Enabled:        true,
	}

	jsonBytes, err := json.Marshal(hostname)
	require.NoError(t, err)

	var decoded Hostname8
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, hostname.Id, decoded.Id)
	assert.Equal(t, hostname.Name, decoded.Name)
	assert.Equal(t, hostname.Live, decoded.Live)
	assert.Equal(t, hostname.Domainid, decoded.Domainid)
	assert.Equal(t, hostname.Enabled, decoded.Enabled)
}

func TestPostHostname8_JSONMarshaling(t *testing.T) {
	post := PostHostname8{
		Name:    "sub.example.com",
		Enabled: true,
		Live:    true,
	}

	jsonBytes, err := json.Marshal(post)
	require.NoError(t, err)

	var decoded PostHostname8
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, post.Name, decoded.Name)
	assert.Equal(t, post.Enabled, decoded.Enabled)
	assert.Equal(t, post.Live, decoded.Live)
}

func TestResult8_Hostnames(t *testing.T) {
	result := Result8{
		Hostnames: map[string][]string{
			"example.com": {"sub1.example.com", "sub2.example.com"},
			"test.com":    {"api.test.com"},
		},
	}

	assert.Len(t, result.Hostnames, 2)
	assert.Len(t, result.Hostnames["example.com"], 2)
	assert.Contains(t, result.Hostnames["example.com"], "sub1.example.com")
}

func TestNotificationConstants(t *testing.T) {
	// Test Notificationevent constants
	assert.Equal(t, Notificationevent("message"), Message)
	assert.Equal(t, Notificationevent("security"), Security)
	assert.Equal(t, Notificationevent("system_error"), Error)
	assert.Equal(t, Notificationevent("system_warning"), Warning)

	// Test Notificationchannel constants
	assert.Equal(t, Notificationchannel("app"), App)
	assert.Equal(t, Notificationchannel("email"), Email)
	assert.Equal(t, Notificationchannel("sms"), Sms)

	// Test Roletype constants
	assert.Equal(t, Roletype("user"), RoleUser)
	assert.Equal(t, Roletype("admin"), RoleAdmin)
}

func TestNotification8_JSONMarshaling(t *testing.T) {
	id := uuid.Must(uuid.NewV4())
	userId := uuid.Must(uuid.NewV4())
	now := time.Now()

	notification := Notification8{
		Id:       &id,
		Userid:   &userId,
		Userrole: RoleAdmin,
		Type:     Security,
		Message:  "Security alert",
		Metadata: NotificationMetadata8{
			Severity:    "high",
			Channeltype: App,
			Eventtype:   Security,
		},
		Read:       false,
		Created_at: &now,
	}

	jsonBytes, err := json.Marshal(notification)
	require.NoError(t, err)

	var decoded Notification8
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, notification.Id, decoded.Id)
	assert.Equal(t, notification.Userrole, decoded.Userrole)
	assert.Equal(t, notification.Type, decoded.Type)
	assert.Equal(t, notification.Message, decoded.Message)
	assert.Equal(t, notification.Metadata.Severity, decoded.Metadata.Severity)
}

func TestNotificationMetadata8_JSONMarshaling(t *testing.T) {
	senderId := uuid.Must(uuid.NewV4())

	metadata := NotificationMetadata8{
		Room:         "general",
		Relativepath: "/notifications",
		Senderid:     &senderId,
		Sendername:   "System",
		Senderemail:  "system@example.com",
		Severity:     "critical",
		Channeltype:  Email,
		Eventtype:    Error,
	}

	jsonBytes, err := json.Marshal(metadata)
	require.NoError(t, err)

	var decoded NotificationMetadata8
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, metadata.Room, decoded.Room)
	assert.Equal(t, metadata.Severity, decoded.Severity)
	assert.Equal(t, metadata.Sendername, decoded.Sendername)
}

func TestDomain8Uri_Fields(t *testing.T) {
	uri := Domain8Uri{
		ID: "550e8400-e29b-41d4-a716-446655440000",
	}

	assert.NotEmpty(t, uri.ID)
}

func TestHostname8Uri_Fields(t *testing.T) {
	uri := Hostname8Uri{
		Domainid: "550e8400-e29b-41d4-a716-446655440000",
		ID:       "550e8400-e29b-41d4-a716-446655440001",
	}

	assert.NotEmpty(t, uri.Domainid)
	assert.NotEmpty(t, uri.ID)
}
