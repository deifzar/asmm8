package passive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassiveRunner_Initialization(t *testing.T) {
	t.Run("creates runner with seed domains", func(t *testing.T) {
		runner := PassiveRunner{
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
		runner := PassiveRunner{
			SeedDomains: []string{},
			Results:     0,
			Subdomains:  make(map[string][]string),
		}

		assert.Empty(t, runner.SeedDomains)
	})

	t.Run("creates runner with subdomains", func(t *testing.T) {
		subdomains := map[string][]string{
			"example.com": {"sub1.example.com", "sub2.example.com"},
		}

		runner := PassiveRunner{
			SeedDomains: []string{"example.com"},
			Subdomains:  subdomains,
		}

		assert.Len(t, runner.Subdomains["example.com"], 2)
	})
}

func TestPassiveRunner_RunPassiveEnum_NoDomains(t *testing.T) {
	// Test with empty domain list - should return empty results without calling external tools
	runner := PassiveRunner{
		SeedDomains: []string{},
		Results:     0,
		Subdomains:  make(map[string][]string),
	}

	prevResults := make(map[string][]string)
	results, err := runner.RunPassiveEnum(prevResults)

	assert.NoError(t, err)
	assert.Empty(t, results)
}
