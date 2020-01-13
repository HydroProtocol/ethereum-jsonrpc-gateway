package core

import (
	"context"
	"fmt"
	"strings"
)

type Config struct {
	Upstreams               []string `json:"upstreams"`
	Strategy                string   `json:"strategy"`
	MethodLimitationEnabled bool     `json:"methodLimitationEnabled"`
	AllowedMethods          []string `json:"allowedMethods"`
	ContractWhitelist       []string `json:"contractWhitelist"`
}

type RunningConfig struct {
	ctx                     context.Context
	stop                    context.CancelFunc
	Upstreams               []Upstream
	Strategy                IStrategy
	MethodLimitationEnabled bool
	allowedMethods          map[string]bool
	allowedCallContracts    map[string]bool
}

var currentRunningConfig *RunningConfig

func BuildRunningConfigFromConfig(parentContext context.Context, cfg *Config) (*RunningConfig, error) {
	ctx, stop := context.WithCancel(parentContext)

	rcfg := &RunningConfig{
		ctx:  ctx,
		stop: stop,
	}

	currentRunningConfig = rcfg

	for _, url := range cfg.Upstreams {

		// hack, refactor this sometime

		var primaryUrl string
		var oldTrieUrl string

		if strings.Contains(url, ",") {
			urls := strings.Split(url, ",")
			primaryUrl = urls[0]
			oldTrieUrl = urls[1]
		} else {
			primaryUrl = url
			oldTrieUrl = url
		}

		rcfg.Upstreams = append(rcfg.Upstreams, newUpstream(ctx, primaryUrl, oldTrieUrl))
	}

	if len(rcfg.Upstreams) == 0 {
		return nil, fmt.Errorf("need upstreams")
	}

	switch cfg.Strategy {
	case "NAIVE":
		if len(rcfg.Upstreams) > 1 {
			panic(fmt.Errorf("naive proxy strategy require exact 1 upstream"))
		}
		rcfg.Strategy = newNaiveProxy()
	case "RACE":
		if len(rcfg.Upstreams) < 2 {
			panic(fmt.Errorf("race proxy strategy require more than 1 upstream"))
		}
		rcfg.Strategy = newRaceProxy()
	case "FALLBACK":
		if len(rcfg.Upstreams) < 2 {
			panic(fmt.Errorf("fallback proxy strategy require more than 1 upstream"))
		}
		rcfg.Strategy = newFallbackProxy()
	default:
		return nil, fmt.Errorf("blank of unsupported strategy: %s", cfg.Strategy)
	}

	rcfg.MethodLimitationEnabled = cfg.MethodLimitationEnabled

	rcfg.allowedMethods = make(map[string]bool)
	for i := 0; i < len(cfg.AllowedMethods); i++ {
		rcfg.allowedMethods[cfg.AllowedMethods[i]] = true
	}

	rcfg.allowedCallContracts = make(map[string]bool)
	for i := 0; i < len(cfg.ContractWhitelist); i++ {
		rcfg.allowedCallContracts[strings.ToLower(cfg.ContractWhitelist[i])] = true
	}

	return rcfg, nil
}
