package tests

import (
	"testing"
)

func TestPluginInit(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-init.yaml")

	SendRequest(t, "/", 200)

	StopRoadrunner(wg, stopCh)
}

func TestReturnsError(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-init.yaml")

	SendRequest(t, "/error", 500)

	StopRoadrunner(wg, stopCh)
}

func TestTriggerCb(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-init.yaml")

	SendRequest(t, "/error", 500)
	SendRequest(t, "/", 503)

	StopRoadrunner(wg, stopCh)
}
