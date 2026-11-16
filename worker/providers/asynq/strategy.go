package asynqProvider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type envelope struct {
	Payload  json.RawMessage   `json:"payload"`
	Metadata map[string]string `json:"metadata"`
}

type Strategy struct{}

func NewStrategy() *Strategy {
	return &Strategy{}
}

func redisOptions() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:        config.GetRedisAddr(),
		Password:    config.GetRedisPassword(),
		DB:          config.GetRedisDB(),
		DialTimeout: config.GetRedisConnectTimeout(),
	}
}

func (s *Strategy) InitializeClient(_ context.Context) (workerConfigs.IWorkerClient, error) {
	client := asynq.NewClient(redisOptions())
	if err := client.Ping(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("initialize asynq client: %w", err)
	}

	return &Client{client: client, inspector: asynq.NewInspector(redisOptions())}, nil
}

func (s *Strategy) InitializeServer(
	_ context.Context,
	aggregatorRegistry workerConfigs.WorkerAggregatorRegistry,
) (workerConfigs.IWorkerServer, error) {
	return newServer(aggregatorRegistry), nil
}

func encodeEnvelope(payload []byte, metadata map[string]string) ([]byte, error) {
	encoded, err := json.Marshal(envelope{Payload: payload, Metadata: metadata})
	if err != nil {
		return nil, fmt.Errorf("encode task envelope: %w", err)
	}

	return encoded, nil
}

func decodeEnvelope(raw []byte) (envelope, error) {
	decoded := envelope{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return envelope{}, fmt.Errorf("decode task envelope: %w", err)
	}

	if decoded.Metadata == nil {
		decoded.Metadata = map[string]string{}
	}

	return decoded, nil
}
