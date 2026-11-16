package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
)

type App struct {
	router       http.Handler
	dependencies *dependencies.Dependencies
}

func NewApp(ctx context.Context) (*App, error) {
	logger := utils.InitLogger(config.GetModeServer())

	deps, err := dependencies.NewDependencies(ctx, logger, config.GetDBMaxConns())
	if err != nil {
		return nil, fmt.Errorf("build dependencies: %w", err)
	}

	return &App{
		router:       routes.NewRoutes(deps),
		dependencies: deps,
	}, nil
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:              ":" + config.GetPort(),
		Handler:           app.router,
		ReadHeaderTimeout: config.GetReadHeaderTimeout(),
		ReadTimeout:       config.GetReadTimeout(),
		WriteTimeout:      config.GetWriteTimeout(),
		IdleTimeout:       config.GetIdleTimeout(),
	}
	server.RegisterOnShutdown(utils.BeginShutdown)

	serverErrors := make(chan error, 1)

	go func() {
		utils.LogInfoWithoutCtx("http server listening", "port", config.GetPort())

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		config.GetServerShutdownTimeout(),
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	utils.DrainBackgroundGoroutines()
	utils.LogInfoWithoutCtx("http server stopped")

	return nil
}

func (app *App) Close() {
	app.dependencies.Close()
}
