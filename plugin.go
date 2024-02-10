package circuitbreaker

import (
	"net/http"

	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

const pluginName = "circuitbreaker"

type Configurer interface {
	// Experimental checks if RR runs in experimental mode.
	Experimental() bool
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if the config section exists.
	Has(name string) bool
}

type Logger interface {
	NamedLogger(name string) *zap.Logger
}

type Plugin struct {
	log *zap.Logger
}

// Configurer and Logger would be provided by RR container
func (p *Plugin) Init(cfg Configurer, logger Logger) error {
	const op = errors.Op("circuitbreaker_middleware_init")
	if !cfg.Has(pluginName) {
		// we don't have plugin's configuration, tell the container that we have to remove it from the graph
		return errors.E(errors.Disabled)
	}

	conf := &Config{}
	// populate configuration
	err := cfg.UnmarshalKey(pluginName, conf)
	if err != nil {
		return errors.E(op, err)
	}

	// at this point we're able to safely init defaults
	conf.InitDefault()

	return nil
}

func (p *Plugin) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.log.Debug("hello world")
		// here we can use logger - p.log
	})
}

func (p *Plugin) Name() string {
	return pluginName
}
