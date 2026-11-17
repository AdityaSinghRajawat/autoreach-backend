package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type ConcurrentTask struct {
	Name string
	Run  func(ctx context.Context) error
}

type GoroutineManager struct {
	waitGroup sync.WaitGroup
	slots     chan struct{}
}

var (
	backgroundGoroutinesI    *GoroutineManager
	backgroundGoroutinesOnce sync.Once
)

func getBackgroundGoroutines() *GoroutineManager {
	backgroundGoroutinesOnce.Do(func() {
		backgroundGoroutinesI = &GoroutineManager{
			slots: make(chan struct{}, config.GetBackgroundMaxGoroutines()),
		}
	})

	return backgroundGoroutinesI
}

func RunInBackground(ctx context.Context, name string, work func(ctx context.Context)) {
	manager := getBackgroundGoroutines()
	detached := context.WithoutCancel(ctx)

	select {
	case manager.slots <- struct{}{}:
	default:
		runRecovered(detached, name, work)

		return
	}

	manager.waitGroup.Add(1)

	go func() {
		defer manager.waitGroup.Done()
		defer func() { <-manager.slots }()

		runRecovered(detached, name, work)
	}()
}

func DrainBackgroundGoroutines() {
	manager := getBackgroundGoroutines()
	drained := make(chan struct{})

	go func() {
		manager.waitGroup.Wait()
		close(drained)
	}()

	select {
	case <-drained:
	case <-time.After(config.GetBackgroundDrainTimeout()):
		LogInfoWithoutCtx("background goroutines still running after drain timeout")
	}
}

func RunConcurrently(ctx context.Context, tasks ...ConcurrentTask) error {
	return RunConcurrentlyWithLimit(ctx, config.GetConcurrentTaskLimit(), tasks...)
}

func RunConcurrentlyWithLimit(ctx context.Context, limit int, tasks ...ConcurrentTask) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(max(limit, 1))

	for _, task := range tasks {
		group.Go(func() (err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					err = fmt.Errorf("%s: %w: %v", task.Name, config.ErrTaskPanicked, recovered)
				}
			}()

			if err := task.Run(groupCtx); err != nil {
				return fmt.Errorf("%s: %w", task.Name, err)
			}

			return nil
		})
	}

	return group.Wait()
}

func runRecovered(ctx context.Context, name string, work func(ctx context.Context)) {
	defer func() {
		if recovered := recover(); recovered != nil {
			LogError(ctx, fmt.Errorf("%s: %w: %v", name, config.ErrTaskPanicked, recovered))
		}
	}()

	work(ctx)
}
