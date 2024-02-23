package circuitbreaker

import (
	"go.uber.org/zap"
	"sync"
	"time"
)

type status int32

const (
	statusClosed   status = 0
	statusOpen     status = 1
	statusHalfOpen status = 2
)

type circuitBreaker struct {
	log                        *zap.Logger
	maxErrorRate               float64
	timeToHalfOpen             time.Duration
	timeToClosed               time.Duration
	storage                    storage
	statusChangeWatcherRunning *sync.WaitGroup
	nextStatusChangeChannel    chan time.Time
	stopStatusChanger          chan struct{}

	// Basically only writing to currentStatus is protected by mu
	// Since the storage (tumblingTimeWindow) is entirely in atomic operations
	// and reading from currentStatus is okay to not be synchronized (and shouldn't be).
	// Using mu for storage may become a necessity once slidingTimeWindow is fully implemented.
	mu            sync.Mutex
	currentStatus status
}

func (cb *circuitBreaker) Init(timeWindow time.Duration) {
	cb.currentStatus = statusClosed
	cb.mu = sync.Mutex{}
	cb.statusChangeWatcherRunning = &sync.WaitGroup{}
	cb.nextStatusChangeChannel = make(chan time.Time, 2)
	cb.stopStatusChanger = make(chan struct{}, 1)
	cb.storage.init(timeWindow)

	cb.statusChangeWatcherRunning.Add(1)
	go func() {
		defer cb.statusChangeWatcherRunning.Done()

		for {
			select {
			case <-cb.stopStatusChanger:
				cb.log.Debug("Quitting status changer")
				return
			case nextStatusChange := <-cb.nextStatusChangeChannel:
				cb.log.Debug("Sleeping until " + nextStatusChange.Format(time.RFC3339))
				time.Sleep(nextStatusChange.Sub(time.Now()))
			default:
				cb.mu.Lock()
				switch cb.currentStatus {
				case statusOpen:
					cb.log.Debug("Changed status from Open to HalfOpen")
					cb.currentStatus = statusHalfOpen
					cb.mu.Unlock()
					cb.nextStatusChangeChannel <- time.Now().Add(cb.timeToClosed)
				case statusHalfOpen:
					cb.log.Debug("Changed status from HalfOpen to Closed")
					cb.currentStatus = statusClosed
					cb.mu.Unlock()
				default:
					cb.mu.Unlock()
				}
				// We don't need to immediately check this again
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (cb *circuitBreaker) AllowRequest() bool {
	return cb.currentStatus != statusOpen
}

func (cb *circuitBreaker) AddSuccess(time time.Time) {
	cb.storage.addSuccess(time)
}

func (cb *circuitBreaker) AddError(time time.Time) {
	cb.storage.addError(time)

	if cb.currentStatus == statusHalfOpen || cb.storage.getErrorRate() >= cb.maxErrorRate {
		cb.mu.Lock()
		defer cb.mu.Unlock()
		cb.log.Debug(
			"Changed status to Open",
			zap.String("OldStatus", cb.GetStatusAsName()),
			zap.Float64("ErrorRate", cb.storage.getErrorRate()),
		)
		cb.currentStatus = statusOpen
		cb.nextStatusChangeChannel <- time.Add(cb.timeToHalfOpen)
	}
}

func (cb *circuitBreaker) Stop() {
	cb.stopStatusChanger <- struct{}{}
	cb.statusChangeWatcherRunning.Wait()
}

func (cb *circuitBreaker) GetStatusAsName() string {
	switch cb.currentStatus {
	case statusClosed:
		return "Closed"
	case statusHalfOpen:
		return "HalfOpen"
	case statusOpen:
		return "Open"
	}

	return ""
}
