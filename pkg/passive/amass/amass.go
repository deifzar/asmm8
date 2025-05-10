package amass

// @Deifzar
// ASMM8 will decomissioned this file in future versions due to the following reasons:

// Amass has been predesigned and it is not longer a single tool.
// Amass is a system of systems with a number of tools integrated in a Docker compose file.
// Amass aims to integrate tools to work together as an ASM Platform.
// Info: https://github.com/owasp-amass/amass/issues/1042
// Caffix: [...] Soon, there will no longer be an amass tool, since it's now a system of systems. It's easiest to setup this system in your environment using Docker Compose. All the components are still being built from open-source software.

import (
	"bytes"
	"deifzar/asmm8/pkg/log8"
	"deifzar/asmm8/pkg/model8"
	"os/exec"
	"strings"
	"sync"
)

func RunAmassIn(seedDomain string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	log8.BaseLogger.Info().Msgf("Running `Amass` on %s\n", seedDomain)
	var out, outerr bytes.Buffer
	cmd := exec.Command("amass", "enum", "-passive", "-config", "./amassconfig.yaml", "-log", "./amasserror.log", "-nocolor", "-d", seedDomain)
	cmd.Stdout = &out
	cmd.Stderr = &outerr

	err := cmd.Run()

	if err != nil {
		log8.BaseLogger.Debug().Msgf("`Amass` reported the following err %s", outerr.String())
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("An error has occured with `Amass`")
		close(results)
		return
	}
	log8.BaseLogger.Info().Msgf("Running `oam_subs` on %s\n", seedDomain)
	var out2, outerr2 bytes.Buffer
	cmd = exec.Command("oam_subs", "-names", "-config", "./amassconfig.yaml", "-d", seedDomain)
	cmd.Stdout = &out2
	cmd.Stderr = &outerr2

	err = cmd.Run()

	if err != nil {
		log8.BaseLogger.Debug().Msgf("`oam_subs` reported the following err %s", outerr2.String())
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("An error has occured with `oam_subs`")
		close(results)
		return
	}

	for _, domain := range strings.Split(out2.String(), "\n") {
		if strings.Contains(domain, seedDomain) && len(domain) != 0 {
			results <- domain
		}
	}
	close(results)
	log8.BaseLogger.Info().Msgf("`Amass` Run completed for %s\n", seedDomain)
}

func RunAmassOut(seedDomain string, results <-chan string, r *model8.Result8, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case a, ok := <-results:
			if ok {
				r.Hostnames[seedDomain] = append(r.Hostnames[seedDomain], a)
			} else {
				log8.BaseLogger.Info().Msgf("`Amass out` run completed for %s\n", seedDomain)
				return
			}
		}
	}
}
