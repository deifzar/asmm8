package active

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActiveRunner_Initialization(t *testing.T) {
	t.Run("creates runner with seed domains", func(t *testing.T) {
		runner := ActiveRunner{
			SeedDomains: []string{"example.com", "test.com"},
			Results:     0,
			Subdomains:  make(map[string][]string),
		}

		assert.Len(t, runner.SeedDomains, 2)
		assert.Contains(t, runner.SeedDomains, "example.com")
		assert.Contains(t, runner.SeedDomains, "test.com")
		assert.Equal(t, 0, runner.Results)
		assert.NotNil(t, runner.Subdomains)
	})

	t.Run("creates runner with empty domains", func(t *testing.T) {
		runner := ActiveRunner{
			SeedDomains: []string{},
			Results:     0,
			Subdomains:  make(map[string][]string),
		}

		assert.Empty(t, runner.SeedDomains)
	})

	t.Run("creates runner with existing subdomains", func(t *testing.T) {
		subdomains := map[string][]string{
			"example.com": {"sub1.example.com", "sub2.example.com", "api.example.com"},
		}

		runner := ActiveRunner{
			SeedDomains: []string{"example.com"},
			Subdomains:  subdomains,
		}

		assert.Len(t, runner.Subdomains["example.com"], 3)
		assert.Contains(t, runner.Subdomains["example.com"], "api.example.com")
	})
}

func TestActiveRunner_RunActiveEnum_NoDomains(t *testing.T) {
	// Test with empty domain list - should return empty results without calling external tools
	runner := ActiveRunner{
		SeedDomains: []string{},
		Results:     0,
		Subdomains:  make(map[string][]string),
	}

	prevResults := make(map[string][]string)
	results, err := runner.RunActiveEnum("wordlist.txt", 10, prevResults)

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestActiveRunner_CheckLiveSubdomains_NoSubdomains(t *testing.T) {
	// Test with empty subdomains - should return empty results
	runner := ActiveRunner{
		SeedDomains: []string{"example.com"},
		Subdomains:  make(map[string][]string),
	}

	// Note: This test may require external tool 'dnsx' to be installed
	// If dnsx is not installed, this will return empty results
	results := runner.CheckLiveSubdomains(10)

	// With no subdomains to check, should return empty or nil
	assert.NotNil(t, results)
}
