package tests

import (
	cb "github.com/roadrunner-server/circuit-breaker/v4"
	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	httpPlugin "github.com/roadrunner-server/http/v4"
	"github.com/roadrunner-server/logger/v4"
	"github.com/roadrunner-server/server/v4"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"
)

func StartRoadrunner(t *testing.T) (wg *sync.WaitGroup, stopCh chan struct{}) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2023.3.5",
		Path:    "configs/.rr-cb-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&cb.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg = &sync.WaitGroup{}
	wg.Add(1)

	stopCh = make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	// used for GitHub
	time.Sleep(time.Second)

	return wg, stopCh
}

func StopRoadrunner(wg *sync.WaitGroup, stopCh chan struct{}) {
	stopCh <- struct{}{}
	wg.Wait()
}
