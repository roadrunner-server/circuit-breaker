package circuitbreaker

type Config struct {
	option1 string `mapstructure:"option_1"` // option1 - how it would be in the plugin. option_1 - how it would be in the .rr.yaml
	// here is your configuration
}

func (c *Config) InitDefault() {
	// init default values
	if c.option1 == "" {
		c.option1 = "option"
	}
}
