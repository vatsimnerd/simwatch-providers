package vatspydata

import (
	simwatchproviders "github.com/vatsimnerd/simwatch-providers"
)

type Config struct {
	DataURL       string                       `mapstructure:"data_url,omitempty"`
	BoundariesURL string                       `mapstructure:"boundaries_url,omitempty"`
	Poll          simwatchproviders.PollConfig `mapstructure:"poll"`
	Boot          simwatchproviders.BootConfig `mapstructure:"boot,omitempty"`
}
