package tests

import (
	"testing"
	"time"
)

func TestIgnoresRequestsOlderThanTimeWindowWithTumblingTimeWindow(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-ignore-old-errors.yaml")

	// Send successful response
	SendRequest(t, "/", 200)
	// Wait time_window defined in config file (1s)
	time.Sleep(time.Second)
	// Give it a little leeway
	time.Sleep(time.Millisecond * 100)
	// Send error and check that it triggered CB (it should)
	SendRequest(t, "/error", 500)
	SendRequest(t, "/", 503)

	StopRoadrunner(wg, stopCh)
}

func TestGoesHalfOpenAfterTimeIntervalAndImmediatelyClosesOnError(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-half-open.yaml")

	// Send error response
	SendRequest(t, "/error", 500)
	// CB open
	SendRequest(t, "/", 503)
	// Wait time_to_halfopen defined in config file (1s)
	time.Sleep(time.Second)
	// Give it a little leeway
	time.Sleep(time.Millisecond * 100)
	// Send error and check that it triggered CB (it should)
	SendRequest(t, "/error", 500)
	SendRequest(t, "/", 503)

	StopRoadrunner(wg, stopCh)
}

func TestGoesHalfOpenAfterTimeIntervalAndAllowSuccessfulRequests(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-half-open.yaml")

	// Send error response
	SendRequest(t, "/error", 500)
	// CB open
	SendRequest(t, "/", 503)
	// Wait time_to_halfopen defined in config file (1s)
	time.Sleep(time.Second)
	// Give it a little leeway
	time.Sleep(time.Millisecond * 100)
	// Send error and check that it triggered CB (it should)
	SendRequest(t, "/", 200)
	SendRequest(t, "/error", 500)
	SendRequest(t, "/", 503)

	StopRoadrunner(wg, stopCh)
}

func TestGoesHalfOpenAfterTimeIntervalAndAllowClosedIfOnlySuccessfulRequests(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-half-open.yaml")

	// Send error response
	SendRequest(t, "/error", 500)
	// CB open
	SendRequest(t, "/", 503)
	// Wait time_to_halfopen defined in config file (1s)
	time.Sleep(time.Second)
	// Send success
	SendRequest(t, "/", 200)
	// Wait time_to_closed defined in config file (2s)
	time.Sleep(time.Second * 2)
	// Send success, error, success, CB should not be triggered
	SendRequest(t, "/", 200)
	SendRequest(t, "/error", 500)
	SendRequest(t, "/", 200)

	StopRoadrunner(wg, stopCh)
}
