package circuitbreaker

import (
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

type Status int32

const (
	StatusClosed   Status = 0
	StatusOpen     Status = 1
	StatusHalfOpen Status = 2
)

type circuitBreaker struct {
	log                        *zap.Logger
	maxErrorRate               float64
	timeToHalfOpen             time.Duration
	timeToClosed               time.Duration
	storage                    storage
	statusChangeWatcherRunning atomic.Bool
	nextStatusChangeChannel    chan time.Time

	// Basically only writing to currentStatus is protected by mu
	// Since the storage (tumblingTimeWindow) is entirely in atomic operations
	// and reading from currentStatus is okay to not be synchronized (and shouldn't be).
	// Using mu for storage may become a necessity once slidingTimeWindow is fully implemented.
	mu            sync.Mutex
	currentStatus Status
}

func (cb *circuitBreaker) Init(timeWindow time.Duration) {
	cb.currentStatus = StatusClosed
	cb.mu = sync.Mutex{}
	cb.nextStatusChangeChannel = make(chan time.Time, 2)
	cb.storage.Init(timeWindow)
}

func (cb *circuitBreaker) AllowRequest() bool {
	return cb.currentStatus != StatusOpen
}

func (cb *circuitBreaker) AddSuccess(time time.Time) {
	cb.storage.AddSuccess(time)
}

func (cb *circuitBreaker) AddError(time time.Time) {
	cb.storage.AddError(time)

	if cb.currentStatus == StatusHalfOpen || cb.storage.GetErrorRate() >= cb.maxErrorRate {
		cb.mu.Lock()
		defer cb.mu.Unlock()
		cb.log.Debug(
			"Changed Status to Open",
			zap.String("OldStatus", cb.GetStatusAsName()),
			zap.Float64("ErrorRate", cb.storage.GetErrorRate()),
		)
		cb.currentStatus = StatusOpen
		cb.StatusChangeWatcher(time.Add(cb.timeToHalfOpen))
	}
}

func (cb *circuitBreaker) StatusChangeWatcher(nextStatusChangeTime time.Time) {
	// Send changed status change time
	cb.nextStatusChangeChannel <- nextStatusChangeTime

	// Watcher is running
	if cb.statusChangeWatcherRunning.CompareAndSwap(false, true) == false {
		return
	}

	go func() {
		defer func() { cb.statusChangeWatcherRunning.Store(false) }()

		for {
			if len(cb.nextStatusChangeChannel) > 0 {
				nextStatusChange := <-cb.nextStatusChangeChannel
				time.Sleep(nextStatusChange.Sub(time.Now()))
			} else {
				cb.mu.Lock()
				switch cb.currentStatus {
				case StatusOpen:
					cb.log.Debug("Changed Status from Open to HalfOpen")
					cb.currentStatus = StatusHalfOpen
					cb.mu.Unlock()
					time.Sleep(cb.timeToClosed)
				case StatusHalfOpen:
					cb.log.Debug("Changed Status from HalfOpen to Closed")
					cb.currentStatus = StatusClosed
					cb.mu.Unlock()
					return
				default:
					cb.log.Debug("Changed nothing")
					cb.mu.Unlock()
					return
				}
			}
		}
	}()
}

func (cb *circuitBreaker) GetStatusAsName() string {
	switch cb.currentStatus {
	case StatusClosed:
		return "Closed"
	case StatusHalfOpen:
		return "HalfOpen"
	case StatusOpen:
		return "Open"
	}

	return ""
}
