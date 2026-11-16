package workerServer

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (ws *WorkerServer) Start(ctx context.Context) error {
	server, err := ws.dependencies.WorkerStrategy.InitializeServer(ctx, ws.aggregatorRegistry)
	if err != nil {
		return err
	}

	registerQueueMetrics(ws.dependencies.WorkerInspector)

	opsServer := ws.newOpsServer()
	opsErrors := make(chan error, 1)

	go func() {
		if err := opsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			opsErrors <- fmt.Errorf("worker ops server: %w", err)
		}
	}()

	ws.cronManager.StartJobs(ctx)

	// The consumer outlives the signal so tasks the cron jobs publish while stopping still run.
	consumerCtx, stopConsumer := context.WithCancel(context.WithoutCancel(ctx))
	defer stopConsumer()

	consumerErrors := make(chan error, 1)
	go func() {
		consumerErrors <- server.Start(consumerCtx, ws.handlerRegistry)
	}()

	var runErr error

	consumerStopped := false

	select {
	case runErr = <-opsErrors:
	case runErr = <-consumerErrors:
		consumerStopped = true
	case <-ctx.Done():
	}

	ws.cronManager.StopJobs()
	stopConsumer()

	if !consumerStopped {
		runErr = errors.Join(runErr, <-consumerErrors)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		config.GetServerShutdownTimeout(),
	)
	defer cancel()

	if err := opsServer.Shutdown(shutdownCtx); err != nil {
		utils.LogErrorWithoutCtx(fmt.Errorf("shutdown worker ops server: %w", err))
	}

	utils.LogInfoWithoutCtx("worker stopped")

	return runErr
}

func (ws *WorkerServer) newOpsServer() *http.Server {
	if !config.IsLocal() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middlewares.Recovery())
	router.GET("/health", handlers.Health)

	metrics := router.Group("")
	if config.IsProduction() {
		metrics.Use(
			middlewares.ErrorHandler(),
			middlewares.InternalSharedSecret(config.GetInternalSharedSecret()),
		)
	}

	metrics.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return &http.Server{
		Addr:              ":" + config.GetPort(),
		Handler:           router,
		ReadHeaderTimeout: config.GetReadHeaderTimeout(),
		ReadTimeout:       config.GetReadTimeout(),
		WriteTimeout:      config.GetWriteTimeout(),
		IdleTimeout:       config.GetIdleTimeout(),
	}
}
