package vatsimapi

import "time"

type Config struct {
	URL  string `mapstructure:"url,omitempty"`
	Poll struct {
		Period  time.Duration `mapstructure:"period,omitempty"`
		Timeout time.Duration `mapstructure:"timeout,omitempty"`
	} `mapstructure:"poll"`
	Boot struct {
		Retries       int           `mapstructure:"retries,omitempty"`
		RetryCooldown time.Duration `mapstructure:"retry_cooldown,omitempty"`
	} `mapstructure:"boot,omitempty"`
}
