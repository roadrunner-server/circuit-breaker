package circuitbreaker

import "time"

type Config struct {
	maxErrorRate   float32       `mapstructure:"max_error_rate"`
	timeToHalfOpen time.Duration `mapstructure:"time_to_halfopen"`
	timeToClosed   time.Duration `mapstructure:"time_to_closed"`
	errors         []int         `mapstructure:"error_codes"`
	codeWhenOpen   int           `mapstructure:"code_when_open"`
	timeWindow     time.Duration `mapstructure:"time_window"`
}

func (c *Config) InitDefault() {
	if c.maxErrorRate == 0.0 {
		// 200% will never happen so effectively disabled. Log a warning?
		c.maxErrorRate = 2
	}

	if c.timeToHalfOpen == 0 {
		c.timeToHalfOpen = time.Minute
	}

	if c.timeToClosed == 0 {
		c.timeToClosed = time.Minute
	}

	// Errors array can be empty, effectively disabling this. Log a warning?

	if c.codeWhenOpen == 0 {
		c.codeWhenOpen = 503
	}

	if c.timeWindow == 0 {
		c.timeWindow = 5 * time.Minute
	}
}
