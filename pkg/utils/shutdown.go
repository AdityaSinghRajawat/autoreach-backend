package utils

import "context"

var shutdownCtxI, beginShutdownI = context.WithCancel(context.Background())

// http.Server.Shutdown would otherwise wait for event streams until its timeout.
func BeginShutdown() {
	beginShutdownI()
}

// UntilShutdown is ctx, also cancelled when the process begins shutting down.
func UntilShutdown(ctx context.Context) (context.Context, context.CancelFunc) {
	merged, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(shutdownCtxI, cancel)

	return merged, func() {
		stop()
		cancel()
	}
}
