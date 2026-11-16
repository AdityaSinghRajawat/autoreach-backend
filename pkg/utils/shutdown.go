package utils

import "context"

var shutdownCtxI, beginShutdownI = context.WithCancel(context.Background())

// BeginShutdown ends long-lived work such as event streams, which http.Server.Shutdown would
// otherwise wait for until its timeout.
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
