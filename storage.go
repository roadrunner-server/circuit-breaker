package circuitbreaker

import (
	"sync/atomic"
	"time"
)

type storage interface {
	init(time.Duration)
	addError(time.Time)
	addSuccess(time.Time)
	getErrorRate() float64
}

type tumblingTimeWindow struct {
	errors               atomic.Int64
	successes            atomic.Int64
	endTimeWindow        time.Time
	endTimeWindowCounter atomic.Int32
	timeWindowDuration   time.Duration
}

func (storage *tumblingTimeWindow) init(timeWindowDuration time.Duration) {
	storage.timeWindowDuration = timeWindowDuration
	storage.errors.Store(0)
	storage.successes.Store(0)
	storage.endTimeWindow = time.Now().Add(timeWindowDuration)
}

func (storage *tumblingTimeWindow) addError(start time.Time) {
	storage.removeExpiredCounts(start)
	storage.errors.Add(1)
}

func (storage *tumblingTimeWindow) addSuccess(start time.Time) {
	storage.removeExpiredCounts(start)
	storage.successes.Add(1)
}

func (storage *tumblingTimeWindow) getErrorRate() float64 {
	successes := storage.successes.Load()
	errors := storage.errors.Load()
	// Avoid divide by zero errors :)
	if errors == 0 {
		return 0.0
	}

	return float64(errors) / float64(successes+errors)
}

func (storage *tumblingTimeWindow) removeExpiredCounts(start time.Time) {
	if storage.endTimeWindow.Before(start) {
		windowCounter := storage.endTimeWindowCounter.Load()
		if storage.endTimeWindowCounter.CompareAndSwap(windowCounter, windowCounter+1) {
			storage.endTimeWindow = start.Add(storage.timeWindowDuration)
			storage.successes.Store(0)
			storage.errors.Store(0)
		}
	}
}

// TODO: This will still need to be synchronized somehow since we can't use atomics here
// TODO: Maybe a way to do it would be to change the map values to atomics
// TODO: And have a "currentSecondCounter" atomic or whatever that would be incremented by whoever first adds a key to the storage
type slidingTimeWindow struct {
	successes          map[int64]int64
	errors             map[int64]int64
	timeWindowDuration time.Duration
}

func (storage *slidingTimeWindow) init(timeWindow time.Duration) {
	storage.timeWindowDuration = timeWindow
}

func (storage *slidingTimeWindow) addError(start time.Time) {
	storage.removeExpiredCounts(start)
	key := start.Unix()

	if _, ok := storage.errors[key]; ok {
		storage.errors[key] += 1
	} else {
		storage.errors[key] = 1
	}
}

func (storage *slidingTimeWindow) addSuccess(start time.Time) {
	storage.removeExpiredCounts(start)
	key := start.Unix()

	if _, ok := storage.successes[key]; ok {
		storage.successes[key] += 1
	} else {
		storage.successes[key] = 1
	}
}

func (storage *slidingTimeWindow) getErrorRate() float64 {
	errors := int64(0)
	for _, value := range storage.errors {
		errors += value
	}
	successes := int64(0)
	for _, value := range storage.successes {
		successes += value
	}

	if errors == 0 {
		return 0.0
	}

	return float64(errors) / float64(successes+errors)
}

// Don't acquire storage.mu here since it has already been acquired
func (storage *slidingTimeWindow) removeExpiredCounts(start time.Time) {
	startTimeWindow := start.Add(-storage.timeWindowDuration)
	cutoff := startTimeWindow.Unix()
	for key := range storage.errors {
		if key < cutoff {
			delete(storage.errors, key)
		} else {
			break
		}
	}

	for key := range storage.successes {
		if key < cutoff {
			delete(storage.successes, key)
		} else {
			break
		}
	}
}
