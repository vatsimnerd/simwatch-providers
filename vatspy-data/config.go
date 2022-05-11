package vatspydata

import "time"

type Config struct {
	DataURL       string `mapstructure:"data_url,omitempty"`
	BoundariesURL string `mapstructure:"boundaries_url,omitempty"`
	Poll          struct {
		Period  time.Duration `mapstructure:"period,omitempty"`
		Timeout time.Duration `mapstructure:"timeout,omitempty"`
	} `mapstructure:"poll"`
	Boot struct {
		Retries       int           `mapstructure:"retries,omitempty"`
		RetryCooldown time.Duration `mapstructure:"retry_cooldown,omitempty"`
	} `mapstructure:"boot,omitempty"`
}
