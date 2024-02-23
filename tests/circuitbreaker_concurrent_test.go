package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
)

func TestHandlesConcurrentRequestsWell(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-init.yaml")

	requestWg := &sync.WaitGroup{}
	requestWg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer requestWg.Done()
			SendRequest(t, "/", 200)
		}()
	}

	requestWg.Wait()

	StopRoadrunner(wg, stopCh)
}

func TestHandlesConcurrentRequestsWithSomeErrorsWell(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-init.yaml")

	requestWg := &sync.WaitGroup{}
	requestWg.Add(1100)
	for i := 0; i < 1000; i++ {
		go func() {
			defer requestWg.Done()
			SendRequest(t, "/", 200)
		}()
	}

	for i := 0; i < 100; i++ {
		go func() {
			defer requestWg.Done()
			SendRequest(t, "/error", 500)
		}()
	}

	requestWg.Wait()

	StopRoadrunner(wg, stopCh)
}

func TestHandlesConcurrentRequestsWithSomeErrorsAndOpeningCBWell(t *testing.T) {
	wg, stopCh := StartRoadrunner(t, "configs/.rr-cb-init.yaml")

	openCbCounter := atomic.Int64{}

	requestWg := &sync.WaitGroup{}
	requestWg.Add(750)
	for i := 0; i < 500; i++ {
		go func() {
			defer requestWg.Done()
			// perform a request to trigger a middleware
			r, err := http.Get("http://127.0.0.1:17876/")
			assert.NoError(t, err)
			require.NotNil(t, r)

			_ = r.Body.Close()

			assert.Contains(t, [2]int{200, 503}, r.StatusCode)

			if r.StatusCode == 503 {
				openCbCounter.Add(1)
			}
		}()

		// Every second request gets an error as well, so at 500 total there should be 250 errors, also triggering the CB
		if i%2 == 0 {
			go func() {
				defer requestWg.Done()
				// perform a request to trigger a middleware
				r, err := http.Get("http://127.0.0.1:17876/error")
				assert.NoError(t, err)
				require.NotNil(t, r)

				_ = r.Body.Close()

				assert.Contains(t, [2]int{500, 503}, r.StatusCode)

				if r.StatusCode == 503 {
					openCbCounter.Add(1)
				}
			}()
		}
	}

	requestWg.Wait()

	StopRoadrunner(wg, stopCh)

	assert.Greater(t, openCbCounter.Load(), int64(0))
}
