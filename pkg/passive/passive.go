package passive

import (
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"

	// "deifzar/asmm8/pkg/passive/amass"
	"deifzar/asmm8/pkg/passive/subfinder"
	"deifzar/asmm8/pkg/utils"
	"sync"
)

// Modify the structure to contemplate subdomains per domain
type PassiveRunner struct {
	SeedDomains []string
	Results     int
	Subdomains  map[string][]string
}

func (r *PassiveRunner) RunPassiveEnum(prevResults map[string][]string) (map[string][]string, error) {
	var wg sync.WaitGroup
	var results model8.Result8
	var scanError error
	var mu sync.Mutex

	results.Hostnames = make(map[string][]string)
	for _, domain := range r.SeedDomains {
		wg.Add(2)
		sf_results := make(chan string)
		// amass_results := make(chan string)
		log8.BaseLogger.Info().Msgf("Finding domains for %s\n", domain)
		go subfinder.RunSubfinderIn(domain, sf_results, &wg, &scanError, &mu)
		go subfinder.RunSubfinderOut(domain, sf_results, &results, &wg)
		// go amass.RunAmassIn(domain, amass_results, &wg)
		// go amass.RunAmassOut(domain, amass_results, &results, &wg)
	}

	wg.Wait()

	if scanError != nil {
		return results.Hostnames, scanError // Propagate error
	}

	log8.BaseLogger.Info().Msg("Cleaning results from passive scan.")
	for _, domain := range r.SeedDomains {
		results.Hostnames[domain] = append(results.Hostnames[domain], prevResults[domain]...)
		clean := utils.RemoveDuplicates(results.Hostnames[domain])
		results.Hostnames[domain] = clean
	}
	return results.Hostnames, nil
}
