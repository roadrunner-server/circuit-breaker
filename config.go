package circuitbreaker

import (
	"github.com/roadrunner-server/errors"
	"time"
)

type Config struct {
	MaxErrorRate   float64       `mapstructure:"max_error_rate"`
	TimeToHalfOpen time.Duration `mapstructure:"time_to_halfopen"`
	TimeToClosed   time.Duration `mapstructure:"time_to_closed"`
	Errors         []int         `mapstructure:"error_codes"`
	CodeWhenOpen   int           `mapstructure:"code_when_open"`
	TimeWindow     time.Duration `mapstructure:"time_window"`
}

func (c *Config) InitDefault() {
	if c.TimeToHalfOpen == 0 {
		c.TimeToHalfOpen = time.Minute
	}

	if c.TimeToClosed == 0 {
		c.TimeToClosed = time.Minute
	}

	if c.CodeWhenOpen == 0 {
		c.CodeWhenOpen = 503
	}

	if c.TimeWindow == 0 {
		c.TimeWindow = 5 * time.Minute
	}
}

func (c *Config) Valid() error {
	if c.MaxErrorRate == 0 || c.MaxErrorRate > 1.0 {
		return errors.E("The value for max_error_rate has to be in the range (0, 1.0]")
	}

	if len(c.Errors) == 0 {
		return errors.E("The array `error_codes` needs to be populated to enable circuitbreaker")
	}

	return nil
}
