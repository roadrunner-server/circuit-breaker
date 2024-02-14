package circuitbreaker

import (
	"sync"
	"time"
)

type storage interface {
	Init(time.Duration)
	AddError(time.Time)
	AddSuccess(time.Time)
	GetErrorRate() float64
}

type tumblingTimeWindow struct {
	mu sync.Mutex

	// Everything here is protected by mu
	errors             int64
	successes          int64
	endTimeWindow      time.Time
	timeWindowDuration time.Duration
}

func (storage *tumblingTimeWindow) Init(timeWindowDuration time.Duration) {
	storage.mu = sync.Mutex{}
	storage.timeWindowDuration = timeWindowDuration
	storage.errors = 0
	storage.successes = 0
	storage.endTimeWindow = time.Now().Add(timeWindowDuration)
}

func (storage *tumblingTimeWindow) AddError(start time.Time) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	storage.removeExpiredCounts(start)

	storage.errors += 1
}

func (storage *tumblingTimeWindow) AddSuccess(start time.Time) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	storage.removeExpiredCounts(start)

	storage.successes += 1
}

func (storage *tumblingTimeWindow) GetErrorRate() float64 {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	// Avoid divide by zero errors :)
	if storage.successes == 0 {
		return float64(storage.errors)
	}

	return float64(storage.errors) / float64(storage.successes)
}

// Don't acquire storage.mu here since it has already been acquired
func (storage *tumblingTimeWindow) removeExpiredCounts(start time.Time) {
	if storage.endTimeWindow.Before(start) {
		storage.successes = 0
		storage.errors = 0
		storage.endTimeWindow = start.Add(storage.timeWindowDuration)
	}
}

// TODO Doesn't feel very nice, maybe try rewriting it. Probably just changing now.Second() to now.Unix() should be good enough
//type SlidingTimeWindow struct {
//	mu sync.Mutex
//
//	// Everything here is protected by mu
//	successes          map[int]int64
//	errors             map[int]int64
//	timeWindowDuration time.Duration
//}
//
//func (storage *SlidingTimeWindow) AddError() {
//	storage.mu.Lock()
//	defer storage.mu.Unlock()
//
//	now := time.Now()
//	storage.removeExpiredCounts(now)
//
//	if _, ok := storage.errors[now.Second()]; ok {
//		storage.errors[now.Second()] += 1
//	} else {
//		storage.errors[now.Second()] = 1
//	}
//}
//
//func (storage *SlidingTimeWindow) AddSuccess() {
//	storage.mu.Lock()
//	defer storage.mu.Unlock()
//
//	now := time.Now()
//	storage.removeExpiredCounts(now)
//
//	if _, ok := storage.successes[now.Second()]; ok {
//		storage.successes[now.Second()] += 1
//	} else {
//		storage.successes[now.Second()] = 1
//	}
//}
//
//func (storage *SlidingTimeWindow) GetErrorRate() float64 {
//	storage.mu.Lock()
//	defer storage.mu.Unlock()
//
//	errors := int64(0)
//	for _, value := range storage.errors {
//		errors += value
//	}
//	successes := int64(0)
//	for _, value := range storage.successes {
//		successes += value
//	}
//
//	if successes == 0 {
//		return float64(errors)
//	}
//
//	return float64(errors) / float64(successes)
//}
//
//// Don't acquire storage.mu here since it has already been acquired
//func (storage *SlidingTimeWindow) removeExpiredCounts(now time.Time) {
//	startTimeWindow := now.Add(-storage.timeWindowDuration)
//}
