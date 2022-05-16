package ourairports

import (
	simwatchproviders "github.com/vatsimnerd/simwatch-providers"
)

type Config struct {
	URL  string                       `mapstructure:"url,omitempty"`
	Poll simwatchproviders.PollConfig `mapstructure:"poll"`
	Boot simwatchproviders.BootConfig `mapstructure:"boot,omitempty"`
}
