package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildWHEREQueryForDomains(t *testing.T) {
	tests := []struct {
		name     string
		domains  []string
		expected string
	}{
		{
			name:     "empty slice",
			domains:  []string{},
			expected: "",
		},
		{
			name:     "single domain",
			domains:  []string{"example.com"},
			expected: `name = "example.com"`,
		},
		{
			name:     "two domains",
			domains:  []string{"example.com", "test.com"},
			expected: `name = "example.com" OR name = "test.com"`,
		},
		{
			name:     "three domains",
			domains:  []string{"a.com", "b.com", "c.com"},
			expected: `name = "a.com" OR name = "b.com" OR name = "c.com"`,
		},
		{
			name:     "domain with subdomain",
			domains:  []string{"sub.example.com"},
			expected: `name = "sub.example.com"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildWHEREQueryForDomains(tt.domains)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "mixed duplicates",
			input:    []string{"a", "b", "a", "c", "b", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "single element",
			input:    []string{"x"},
			expected: []string{"x"},
		},
		{
			name:     "consecutive duplicates",
			input:    []string{"a", "a", "b", "b", "c", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "preserves order of first occurrence",
			input:    []string{"z", "a", "z", "b", "a"},
			expected: []string{"z", "a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDuplicates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDifference(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []string
		slice2   []string
		expected []string
	}{
		{
			name:     "both empty",
			slice1:   []string{},
			slice2:   []string{},
			expected: nil,
		},
		{
			name:     "first empty",
			slice1:   []string{},
			slice2:   []string{"a", "b"},
			expected: nil,
		},
		{
			name:     "second empty",
			slice1:   []string{"a", "b"},
			slice2:   []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "no overlap",
			slice1:   []string{"a", "b", "c"},
			slice2:   []string{"x", "y", "z"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "complete overlap",
			slice1:   []string{"a", "b", "c"},
			slice2:   []string{"a", "b", "c"},
			expected: nil,
		},
		{
			name:     "partial overlap",
			slice1:   []string{"a", "b", "c", "d"},
			slice2:   []string{"b", "d"},
			expected: []string{"a", "c"},
		},
		{
			name:     "subset in second",
			slice1:   []string{"x"},
			slice2:   []string{"x", "y", "z"},
			expected: nil,
		},
		{
			name:     "preserves order",
			slice1:   []string{"z", "a", "m", "b"},
			slice2:   []string{"a", "b"},
			expected: []string{"z", "m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Difference(tt.slice1, tt.slice2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidIPAddress(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "valid IPv4",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "valid IPv4 localhost",
			ip:       "127.0.0.1",
			expected: true,
		},
		{
			name:     "valid IPv4 zeros",
			ip:       "0.0.0.0",
			expected: true,
		},
		{
			name:     "valid IPv4 max",
			ip:       "255.255.255.255",
			expected: true,
		},
		{
			name:     "valid IPv6 localhost",
			ip:       "::1",
			expected: true,
		},
		{
			name:     "valid IPv6 full",
			ip:       "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected: true,
		},
		{
			name:     "valid IPv6 compressed",
			ip:       "2001:db8::1",
			expected: true,
		},
		{
			name:     "invalid - empty string",
			ip:       "",
			expected: false,
		},
		{
			name:     "invalid - hostname",
			ip:       "example.com",
			expected: false,
		},
		{
			name:     "invalid - too many octets",
			ip:       "192.168.1.1.1",
			expected: false,
		},
		{
			name:     "invalid - octet out of range",
			ip:       "256.1.1.1",
			expected: false,
		},
		{
			name:     "invalid - letters in IPv4",
			ip:       "192.168.a.1",
			expected: false,
		},
		{
			name:     "invalid - missing octet",
			ip:       "192.168.1",
			expected: false,
		},
		{
			name:     "invalid - leading zeros rejected in Go 1.17+",
			ip:       "192.168.01.1",
			expected: false, // Go's net.ParseIP rejects leading zeros since Go 1.17
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIPAddress(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteTempFile(t *testing.T) {
	t.Run("creates file with content", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test_output.txt")

		data := []string{"line1", "line2", "line3"}
		WriteTempFile(filePath, data)

		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		expected := "line1\nline2\nline3\n"
		assert.Equal(t, expected, string(content))
	})

	t.Run("creates file with empty list", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "empty_output.txt")

		WriteTempFile(filePath, []string{})

		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		assert.Equal(t, "", string(content))
	})

	t.Run("appends to existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "append_output.txt")

		WriteTempFile(filePath, []string{"first"})
		WriteTempFile(filePath, []string{"second"})

		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		expected := "first\nsecond\n"
		assert.Equal(t, expected, string(content))
	})

	t.Run("creates file with single element", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "single_output.txt")

		WriteTempFile(filePath, []string{"only_line"})

		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		expected := "only_line\n"
		assert.Equal(t, expected, string(content))
	})
}

func TestCheckTool(t *testing.T) {
	t.Run("finds existing tool", func(t *testing.T) {
		// 'go' should always be available in a Go testing environment
		result := checkTool("go")
		assert.True(t, result)
	})

	t.Run("does not find non-existent tool", func(t *testing.T) {
		result := checkTool("nonexistent_tool_that_should_never_exist_12345")
		assert.False(t, result)
	})
}
