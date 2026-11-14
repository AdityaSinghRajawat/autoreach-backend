package mocks

import (
	"context"
	"strconv"
	"sync"

	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type WorkerClient struct {
	mutex     sync.Mutex
	Err       error
	Published []workerConfigs.PublishParams
}

var _ workerConfigs.IWorkerClient = (*WorkerClient)(nil)

func NewWorkerClient() *WorkerClient {
	return &WorkerClient{}
}

func (c *WorkerClient) Publish(
	_ context.Context,
	params workerConfigs.PublishParams,
) (*workerConfigs.PublishResult, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Err != nil {
		return nil, c.Err
	}

	c.Published = append(c.Published, params)

	return &workerConfigs.PublishResult{MessageID: "mock-" + strconv.Itoa(len(c.Published))}, nil
}

func (c *WorkerClient) Close() error {
	return nil
}

func (c *WorkerClient) SetErr(err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Err = err
}

func (c *WorkerClient) Snapshot() []workerConfigs.PublishParams {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return append([]workerConfigs.PublishParams(nil), c.Published...)
}
