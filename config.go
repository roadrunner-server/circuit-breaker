package circuitbreaker

import "time"

type Config struct {
	MaxErrorRate   float64       `mapstructure:"max_error_rate"`
	TimeToHalfOpen time.Duration `mapstructure:"time_to_halfopen"`
	TimeToClosed   time.Duration `mapstructure:"time_to_closed"`
	Errors         []int         `mapstructure:"error_codes"`
	CodeWhenOpen   int           `mapstructure:"code_when_open"`
	TimeWindow     time.Duration `mapstructure:"time_window"`
}

func (c *Config) InitDefault() {
	if c.MaxErrorRate == 0.0 {
		// 200% will never happen so effectively disabled. Log a warning?
		c.MaxErrorRate = 2
	}

	if c.TimeToHalfOpen == 0 {
		c.TimeToHalfOpen = time.Minute
	}

	if c.TimeToClosed == 0 {
		c.TimeToClosed = time.Minute
	}

	// Errors array can be empty, effectively disabling this. Log a warning?

	if c.CodeWhenOpen == 0 {
		c.CodeWhenOpen = 503
	}

	if c.TimeWindow == 0 {
		c.TimeWindow = 5 * time.Minute
	}
}
