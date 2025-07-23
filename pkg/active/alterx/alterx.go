package alterx

import (
	"bytes"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func RunAlterxIn(seedDomain string, threads int, input string, output string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	log8.BaseLogger.Info().Msgf("Running DNS Permutations for %s\n", seedDomain)
	var out, outerr bytes.Buffer
	cmd := exec.Command("alterx", "-l", input, "-silent", "-o", output)
	cmd.Stdout = &out
	cmd.Stderr = &outerr

	err := cmd.Run()

	if err != nil {
		log8.BaseLogger.Debug().Msgf("`alterx` reported the following err %s", outerr.String())
		log8.BaseLogger.Debug().Msgf(err.Error())
		log8.BaseLogger.Error().Msg("An error has ocurred with `alterx`")
		close(results)
		return
	}

	out.Reset()
	outerr.Reset()

	cmd = exec.Command("dnsx", "-l", output, "-silent", "-a", "-cname", "-aaaa", "-t", strconv.Itoa(threads))
	cmd.Stdout = &out
	cmd.Stderr = &outerr

	err = cmd.Run()

	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Debug().Msgf("After alterations, `dnsx` reported the following err %s", outerr.String())
		log8.BaseLogger.Error().Msg("An error has ocurred with `dnsx` after alterations")
		close(results)
		return
	}

	for _, domain := range strings.Split(out.String(), "\n") {
		if strings.Contains(domain, seedDomain) && len(domain) != 0 {
			results <- domain
		}
	}
	close(results)
	err = os.Remove(input)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msgf("Error when deleting the file `%s`", input)
	}
	err = os.Remove(output)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msgf("Error when deleting the file `%s`", output)
	}
	log8.BaseLogger.Info().Msgf("DNS Permutations scan has concluded for %s.\n", seedDomain)
}

func RunAlterxOut(seedDomain string, results <-chan string, r *model8.Result8, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case a, ok := <-results:
			if ok {
				r.Hostnames[seedDomain] = append(r.Hostnames[seedDomain], a)
			} else {
				log8.BaseLogger.Info().Msgf("`DNS alterations` run completed for %s\n", seedDomain)
				return
			}
		}
	}
}

// func createDomainFile(seedDomains []string, sdomains map[string][]string) {
// 	for _, d := range seedDomains {
// 		file, err := os.OpenFile("./tmp/tempDomains-"+d+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

// 		if err != nil {
// 			panic(err)
// 		}

// 		datawriter := bufio.NewWriter(file)

// 		for _, data := range sdomains[d] {
// 			_, _ = datawriter.WriteString(data + "\n")
// 		}

// 		datawriter.Flush()
// 		file.Close()
// 	}
// }
