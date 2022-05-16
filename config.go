package simwatchproviders

import "time"

type BootConfig struct {
	Retries       int           `mapstructure:"retries,omitempty"`
	RetryCooldown time.Duration `mapstructure:"retry_cooldown,omitempty"`
}

type PollConfig struct {
	Period  time.Duration `mapstructure:"period,omitempty"`
	Timeout time.Duration `mapstructure:"timeout,omitempty"`
}
