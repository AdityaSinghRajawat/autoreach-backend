package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerServer "github.com/AdityaSinghRajawat/autoreach-backend/worker/server"
)

const adminPasswordEnv = "ADMIN_PASSWORD"

func main() {
	config.Init()

	mode, arguments := resolveMode(os.Args[1:])
	utils.InitLogger(mode)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, mode, arguments)

	cancel()

	if err != nil {
		utils.LogErrorWithoutCtx(err, "mode", mode)
		os.Exit(1)
	}
}

func resolveMode(arguments []string) (string, []string) {
	if len(arguments) == 0 {
		return config.GetModeServer(), nil
	}

	return arguments[0], arguments[1:]
}

func run(ctx context.Context, mode string, arguments []string) error {
	switch mode {
	case config.GetModeServer():
		return startApp(ctx)
	case config.GetModeWorker():
		return startWorker(ctx)
	case config.GetModeAdminCreate():
		return createAdmin(ctx, arguments)
	default:
		return fmt.Errorf("%w: %q", config.ErrUnknownMode, mode)
	}
}

func startApp(ctx context.Context) error {
	app, err := NewApp(ctx)
	if err != nil {
		return err
	}
	defer app.Close()

	return app.Start(ctx)
}

func startWorker(ctx context.Context) error {
	logger := utils.InitLogger(config.GetModeWorker())

	deps, err := dependencies.NewDependencies(ctx, logger, config.GetWorkerDBMaxConns())
	if err != nil {
		return fmt.Errorf("build dependencies: %w", err)
	}
	defer deps.Close()
	defer utils.DrainBackgroundGoroutines()

	worker, err := workerServer.NewWorkerServer(ctx, deps)
	if err != nil {
		return err
	}

	return worker.Start(ctx)
}

func createAdmin(ctx context.Context, arguments []string) error {
	if len(arguments) != 1 {
		return config.ErrCreateAdminUsage
	}

	logger := utils.InitLogger(config.GetModeAdminCreate())

	deps, err := dependencies.NewDependencies(ctx, logger, config.GetDBMaxConns())
	if err != nil {
		return fmt.Errorf("build dependencies: %w", err)
	}
	defer deps.Close()

	user, customErr := deps.Services.Auth.CreateAdmin(
		ctx,
		arguments[0],
		os.Getenv(adminPasswordEnv),
	)
	if customErr != nil {
		return customErr
	}

	utils.LogInfo(ctx, "admin created", "user_id", user.ID)

	return nil
}
