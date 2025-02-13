package main

import (
	"context"
	"sync"
	"time"

	"github.com/openshift/assisted-installer-agent/src/commands"
	"github.com/openshift/assisted-installer-agent/src/config"
	"github.com/openshift/assisted-installer-agent/src/util"
	log "github.com/sirupsen/logrus"
)

func main() {
	agentConfig := config.ProcessArgs()
	config.ProcessDryRunArgs(&agentConfig.DryRunConfig)
	util.SetLogging("agent_next_step_runner", agentConfig.TextLogging, agentConfig.JournalLogging, agentConfig.StdoutLogging, agentConfig.ForcedHostID)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	toolRunnerFactory := commands.NewToolRunnerFactory()
	go commands.ProcessSteps(ctx, cancel, agentConfig, toolRunnerFactory, &wg, log.StandardLogger())

	if agentConfig.DryRunEnabled {
		log.Info(`Dry run enabled, will cancel goroutine on fake "reboot"`)
		for {
			if util.DryRebootHappened(&agentConfig.DryRunConfig) {
				log.Info("Dry reboot happened, exiting")
				cancel()
				break
			}

			time.Sleep(time.Second)
		}
	} else {
		// Nothing interesting to do, wait for the goroutine to finish naturally
		wg.Wait()
	}

	log.Info("next step runner exiting")
}
