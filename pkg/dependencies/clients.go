package dependencies

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/storage"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerProviders "github.com/AdityaSinghRajawat/autoreach-backend/worker/providers"
)

func (d *Dependencies) initClients(ctx context.Context, postgresMaxConns int) error {
	workerStrategy, err := workerProviders.NewWorkerStrategy()
	if err != nil {
		return err
	}

	d.WorkerStrategy = workerStrategy

	err = utils.RunConcurrently(
		ctx,
		utils.ConcurrentTask{
			Name: "connect postgres",
			Run: func(ctx context.Context) error {
				postgres, err := db.ConnectPostgres(ctx, postgresMaxConns)
				d.Postgres = postgres

				return err
			},
		},
		utils.ConcurrentTask{
			Name: "connect redis",
			Run: func(ctx context.Context) error {
				redis, err := db.ConnectRedis(ctx)
				d.Redis = redis

				return err
			},
		},
		utils.ConcurrentTask{
			Name: "connect worker client",
			Run: func(ctx context.Context) error {
				workerClient, err := workerStrategy.InitializeClient(ctx)
				d.WorkerClient = workerClient

				return err
			},
		},
	)
	if err != nil {
		return err
	}

	if d.FileStore == nil {
		fileStore, err := storage.NewFileStore(ctx)
		if err != nil {
			return err
		}

		d.FileStore = fileStore
	}

	return d.initClientHelpers()
}

func (d *Dependencies) initClientHelpers() error {
	transactionManager, err := db.NewPostgresTransactionManager(d.Postgres)
	if err != nil {
		return err
	}

	d.TransactionManager = transactionManager

	rateLimiter, err := utils.NewRateLimiter(d.Redis.GetClient())
	if err != nil {
		return err
	}

	d.RateLimiter = rateLimiter

	redisUtil, err := utils.NewRedisUtil(d.Redis.GetClient())
	if err != nil {
		return err
	}

	d.RedisUtil = redisUtil

	return nil
}
