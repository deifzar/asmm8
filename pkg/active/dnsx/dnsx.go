package dnsx

import (
	"bytes"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"
	"deifzar/asmm8/pkg/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func RunDnsxIn(seedDomain string, wordlist string, threads int, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	log8.BaseLogger.Info().Msgf("Runing DNS Bruteforce for %s", seedDomain)
	var out, outerr bytes.Buffer
	cmd := exec.Command("dnsx", "-d", seedDomain, "-silent", "-w", wordlist, "-a", "-cname", "-aaaa", "-t", strconv.Itoa(threads))
	cmd.Stdout = &out
	cmd.Stderr = &outerr

	err := cmd.Run()

	if err != nil {
		log8.BaseLogger.Debug().Msgf("`dnsx` reported the following err %s", outerr.String())
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("An error has ocurred with `dnsx`")
		close(results)
		return
	}

	for _, domain := range strings.Split(out.String(), "\n") {
		if strings.Contains(domain, seedDomain) && len(domain) != 0 {
			log8.BaseLogger.Info().Msgf("DNS brute force - found hostname `%s`", domain)
			results <- domain
		}
	}
	close(results)
	log8.BaseLogger.Info().Msgf("DNS brute force has concluded for %s\n.", seedDomain)
}

func RunDnsxOut(seedDomain string, results <-chan string, r *model8.Result8, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case a, ok := <-results:
			if ok {
				r.Hostnames[seedDomain] = append(r.Hostnames[seedDomain], a)
			} else {
				log8.BaseLogger.Info().Msgf("`DNSx out` run completed for %s\n", seedDomain)
				return
			}
		}
	}
}

func RunDnsxConfirmLiveSubdomains(subdomains map[string][]string, threads int) map[string][]string {
	log8.BaseLogger.Info().Msgf("Runing DNS check!")
	var results = make(map[string][]string)
	for domain, list := range subdomains {
		utils.WriteTempFile("./tmp/subdomains.txt", list)
		cmd := exec.Command("dnsx", "-l", "./tmp/subdomains.txt", "-silent", "-a", "-cname", "-aaaa", "-t", strconv.Itoa(threads))

		var out, outerr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &outerr
		err := cmd.Run()

		if err != nil {
			log8.BaseLogger.Debug().Msgf("`dnsx` reported the following err %s", outerr.String())
			log8.BaseLogger.Debug().Msg(err.Error())
			log8.BaseLogger.Error().Msgf("An error has ocurred with `dnsx` while checking for live subdomains")
			return results
		}

		for _, sdomain := range strings.Split(out.String(), "\n") {
			if len(sdomain) != 0 {
				results[domain] = append(results[domain], sdomain)
				log8.BaseLogger.Info().Msgf("DNS live hostname `%s`", sdomain)
			}
		}
		os.Remove("./tmp/subdomains.txt")
	}
	return results
}
