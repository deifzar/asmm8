package subfinder

import (
	"bytes"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"
	"os/exec"
	"strings"
	"sync"
)

func RunSubfinderIn(seedDomain string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	log8.BaseLogger.Info().Msgf("Running `Subfinder` on %s\n", seedDomain)
	var out, outerr bytes.Buffer
	cmd := exec.Command("subfinder", "-d", seedDomain, "-silent", "-all", "-config", "./configs/subfinderconfig.yaml", "-pc", "./configs/subfinderprovider-config.yaml")
	cmd.Stdout = &out
	cmd.Stderr = &outerr

	err := cmd.Run()

	if err != nil {
		log8.BaseLogger.Debug().Msgf("`Subfinder` reported the following err %s", outerr.String())
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("An error has ocurred with `Subfinder`")
		close(results)
		return
	}

	for _, domain := range strings.Split(out.String(), "\n") {
		if strings.Contains(domain, seedDomain) && len(domain) != 0 {
			results <- domain
		}
	}
	close(results)
	log8.BaseLogger.Info().Msgf("`Subfinder in` run completed for %s\n", seedDomain)
}

func RunSubfinderOut(seedDomain string, results <-chan string, r *model8.Result8, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case a, ok := <-results:
			if ok {
				r.Hostnames[seedDomain] = append(r.Hostnames[seedDomain], a)
			} else {
				log8.BaseLogger.Info().Msgf("`Subfinder out` run completed for %s\n", seedDomain)
				return
			}
		}
	}
}
