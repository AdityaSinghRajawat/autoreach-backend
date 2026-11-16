package unit

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestRunConcurrentlyRunsEveryTaskWithinTheLimit(t *testing.T) {
	setup.Config(t)

	running, peak, finished := atomic.Int32{}, atomic.Int32{}, atomic.Int32{}
	tasks := make([]utils.ConcurrentTask, config.GetConcurrentTaskLimit()*3)

	for index := range tasks {
		tasks[index] = utils.ConcurrentTask{Name: "task", Run: func(context.Context) error {
			current := running.Add(1)

			for {
				seen := peak.Load()
				if current <= seen || peak.CompareAndSwap(seen, current) {
					break
				}
			}

			time.Sleep(5 * time.Millisecond)
			running.Add(-1)
			finished.Add(1)

			return nil
		}}
	}

	require.NoError(t, utils.RunConcurrently(context.Background(), tasks...))
	assert.Equal(t, int32(len(tasks)), finished.Load())
	assert.LessOrEqual(t, int(peak.Load()), config.GetConcurrentTaskLimit())
}

func TestRunConcurrentlyReportsFailuresAndPanicsByTaskName(t *testing.T) {
	setup.Config(t)

	failure := errors.New("vendor down")

	err := utils.RunConcurrently(context.Background(),
		utils.ConcurrentTask{Name: "fetch", Run: func(context.Context) error { return failure }},
	)
	require.ErrorIs(t, err, failure)
	assert.Contains(t, err.Error(), "fetch")

	err = utils.RunConcurrently(
		context.Background(),
		utils.ConcurrentTask{
			Name: "parse",
			Run:  func(context.Context) error { panic("bad input") },
		},
	)
	require.ErrorIs(t, err, config.ErrTaskPanicked)
	assert.Contains(t, err.Error(), "parse")
}

func TestRunConcurrentlyCancelsTheOthersAfterAFailure(t *testing.T) {
	setup.Config(t)

	cancelled := atomic.Bool{}

	err := utils.RunConcurrently(context.Background(),
		utils.ConcurrentTask{Name: "fails", Run: func(context.Context) error {
			return errors.New("stop")
		}},
		utils.ConcurrentTask{Name: "waits", Run: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				cancelled.Store(true)
			case <-time.After(5 * time.Second):
			}

			return nil
		}},
	)
	require.Error(t, err)
	assert.True(t, cancelled.Load(), "the shared context is cancelled once one task fails")
}

func TestBackgroundWorkFinishesBeforeTheDrainReturnsAndPanicsAreContained(t *testing.T) {
	setup.Config(t)

	done := atomic.Bool{}

	utils.RunInBackground(context.Background(), "panics", func(context.Context) { panic("boom") })
	utils.RunInBackground(context.Background(), "slow", func(context.Context) {
		time.Sleep(20 * time.Millisecond)
		done.Store(true)
	})

	utils.DrainBackgroundGoroutines()
	assert.True(t, done.Load())
}

func TestBackgroundWorkOutlivesTheRequestThatStartedIt(t *testing.T) {
	setup.Config(t)

	request, cancel := context.WithCancel(context.Background())
	stillAlive := atomic.Bool{}

	utils.RunInBackground(request, "after response", func(ctx context.Context) {
		time.Sleep(10 * time.Millisecond)
		stillAlive.Store(ctx.Err() == nil)
	})
	cancel()

	utils.DrainBackgroundGoroutines()
	assert.True(t, stillAlive.Load())
}
