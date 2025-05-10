package active

import (
	"deifzar/asmm8/pkg/active/alterx"
	"deifzar/asmm8/pkg/active/dnsx"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"
	"deifzar/asmm8/pkg/utils"
	"sync"
)

type ActiveRunner struct {
	SeedDomains []string
	Results     int
	Subdomains  map[string][]string
}

func (r *ActiveRunner) CheckLiveSubdomains(threads int) map[string][]string {
	return dnsx.RunDnsxConfirmLiveSubdomains(r.Subdomains, threads)
}

func (r *ActiveRunner) RunActiveEnum(wordlist string, threads int, prevResults map[string][]string) map[string][]string {
	var wg sync.WaitGroup
	var results model8.Result8
	results.Hostnames = make(map[string][]string)
	for _, domain := range r.SeedDomains {
		wg.Add(2)
		dnsx_results := make(chan string)
		go dnsx.RunDnsxIn(domain, wordlist, threads, dnsx_results, &wg)
		go dnsx.RunDnsxOut(domain, dnsx_results, &results, &wg)
	}

	wg.Wait()

	// Uniq results after DNS brute
	log8.BaseLogger.Info().Msg("Cleaning results after DNS bruteforce and creating temp files for DNS alterations")
	for _, domain := range r.SeedDomains {
		results.Hostnames[domain] = append(results.Hostnames[domain], prevResults[domain]...)
		clean := utils.RemoveDuplicates(results.Hostnames[domain])
		results.Hostnames[domain] = clean

		tempFile := "./tmp/tempDomain-" + domain + ".txt"
		// Create new tempDomain files for alterx
		utils.WriteTempFile(tempFile, results.Hostnames[domain])
	}

	// Run DNS alteration
	for _, domain := range r.SeedDomains {
		wg.Add(2)
		input := "./tmp/tempDomain-" + domain + ".txt"
		output := "./tmp/alterxDomain-" + domain + ".txt"
		alterx_results := make(chan string)
		go alterx.RunAlterxIn(domain, threads, input, output, alterx_results, &wg)
		go alterx.RunAlterxOut(domain, alterx_results, &results, &wg)
	}

	wg.Wait()

	log8.BaseLogger.Info().Msg("Cleaning results after DNS alterations")
	for _, domain := range r.SeedDomains {
		clean := utils.RemoveDuplicates(results.Hostnames[domain])
		results.Hostnames[domain] = clean
	}

	return results.Hostnames
}

// ***************************************************
// Deifzar: We do not check here if the subdomains are alive
//
// func (r *ActiveRunner) RunHttpx() {
// 	httpx.RunHttpx(r.SeedDomains, r.Subdomains)
// }
