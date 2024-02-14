package circuitbreaker

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

type Status int32

const (
	pluginName = "circuitbreaker"

	StatusClosed   Status = 0
	StatusOpen     Status = 1
	StatusHalfOpen Status = 2
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

	maxErrorRate   float64
	timeToHalfOpen time.Duration
	timeToClosed   time.Duration
	errors         map[int]bool
	codeWhenOpen   int

	status  atomic.Int32
	storage storage

	nextStatusChange time.Time
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

	p.writersPool = sync.Pool{
		New: func() any {
			wr := new(writer)
			wr.code = -1
			return wr
		},
	}

	p.maxErrorRate = conf.MaxErrorRate
	p.timeToHalfOpen = conf.TimeToHalfOpen
	p.timeToClosed = conf.TimeToClosed
	p.codeWhenOpen = conf.CodeWhenOpen
	p.errors = make(map[int]bool)

	for configuredError := range conf.Errors {
		p.errors[configuredError] = true
	}

	p.storage = &tumblingTimeWindow{}
	p.storage.Init(conf.TimeWindow)

	return nil
}

func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := p.checkStatusChange(start)

		if status == StatusOpen {
			w.WriteHeader(p.codeWhenOpen)
			return
		}

		// overwrite original rw, because we need to get the return status code
		rrWriter := p.getWriter(w)
		defer p.putWriter(rrWriter)

		next.ServeHTTP(rrWriter, r)

		if p.errors[rrWriter.code] {
			p.storage.AddError(start)

			if status == StatusHalfOpen || p.storage.GetErrorRate() > p.maxErrorRate {
				p.changeStatus(status, StatusOpen, start)
			}
		} else {
			p.storage.AddSuccess(start)
		}
	})
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) checkStatusChange(start time.Time) Status {
	status := Status(p.status.Load())

	if status > StatusClosed && p.nextStatusChange.Before(start) {
		nextStatus := status
		switch status {
		case StatusOpen:
			nextStatus = StatusHalfOpen
		case StatusHalfOpen:
			nextStatus = StatusClosed
		}

		status = p.changeStatus(status, nextStatus, start)
	}

	return status
}

func (p *Plugin) changeStatus(oldStatus Status, status Status, start time.Time) Status {
	// We are protected here since status changes won't happen quickly
	if p.status.CompareAndSwap(int32(oldStatus), int32(status)) {
		switch status {
		case StatusOpen:
			p.nextStatusChange = start.Add(p.timeToHalfOpen)
		case StatusHalfOpen:
			p.nextStatusChange = start.Add(p.timeToClosed)
		}

		return status
	}

	return oldStatus
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
