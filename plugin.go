package circuitbreaker

import (
	"net/http"
	"sync"
	"time"

	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

const (
	pluginName = "circuitbreaker"
)

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if the config section exists.
	Has(name string) bool
}

type Logger interface {
	NamedLogger(name string) *zap.Logger
}

type Plugin struct {
	log         *zap.Logger
	writersPool sync.Pool
	cb          circuitBreaker

	errors       map[int]bool
	codeWhenOpen int
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

	// create a named logger for this middleware
	p.log = logger.NamedLogger(pluginName)

	// at this point we're able to safely init defaults
	conf.InitDefault()

	// Validate the config
	err = conf.Valid()
	if err != nil {
		return err
	}

	p.writersPool = sync.Pool{
		New: func() any {
			wr := new(writer)
			wr.code = -1
			return wr
		},
	}

	p.cb = circuitBreaker{}
	p.cb.log = p.log
	p.cb.maxErrorRate = conf.MaxErrorRate
	p.cb.timeToHalfOpen = conf.TimeToHalfOpen
	p.cb.timeToClosed = conf.TimeToClosed
	p.cb.storage = &tumblingTimeWindow{}
	p.cb.Init(conf.TimeWindow)

	p.codeWhenOpen = conf.CodeWhenOpen
	p.errors = make(map[int]bool)

	for _, configuredError := range conf.Errors {
		p.errors[configuredError] = true
	}

	return nil
}

func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !p.cb.AllowRequest() {
			p.log.Debug("Did not allow request due to status being " + p.cb.GetStatusAsName())
			w.WriteHeader(p.codeWhenOpen)
			return
		}

		// overwrite original rw, because we need to get the return status code
		rrWriter := p.getWriter(w)
		defer p.putWriter(rrWriter)

		next.ServeHTTP(rrWriter, r)

		if p.errors[rrWriter.code] {
			p.cb.AddError(time.Now())
		} else {
			p.cb.AddSuccess(time.Now())
		}
	})
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) getWriter(w http.ResponseWriter) *writer {
	wr := p.writersPool.Get().(*writer)
	wr.w = w
	return wr
}

func (p *Plugin) putWriter(w *writer) {
	w.code = -1
	w.w = nil
	p.writersPool.Put(w)
}
