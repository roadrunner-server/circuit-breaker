package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestPluginInit(t *testing.T) {
	wg, stopCh := StartRoadrunner(t)

	// perform a request to trigger a middleware
	r, err := http.Get("http://127.0.0.1:17876/")
	assert.NoError(t, err)
	require.NotNil(t, r)

	_ = r.Body.Close()

	assert.Equal(t, 200, r.StatusCode)

	StopRoadrunner(wg, stopCh)
}

func TestReturnsError(t *testing.T) {
	wg, stopCh := StartRoadrunner(t)

	// perform a request to trigger a middleware
	r, err := http.Get("http://127.0.0.1:17876/error")
	assert.NoError(t, err)
	require.NotNil(t, r)

	_ = r.Body.Close()

	assert.Equal(t, 500, r.StatusCode)

	StopRoadrunner(wg, stopCh)
}

func TestTriggerCb(t *testing.T) {
	wg, stopCh := StartRoadrunner(t)

	// perform a request to trigger a middleware
	r, err := http.Get("http://127.0.0.1:17876/error")
	assert.NoError(t, err)
	require.NotNil(t, r)

	_ = r.Body.Close()

	assert.Equal(t, 500, r.StatusCode)

	// perform a request to trigger a middleware
	r, err = http.Get("http://127.0.0.1:17876/")
	assert.NoError(t, err)
	require.NotNil(t, r)

	_ = r.Body.Close()

	assert.Equal(t, 503, r.StatusCode)

	StopRoadrunner(wg, stopCh)
}
