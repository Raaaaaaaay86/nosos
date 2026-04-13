package nosos

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var DefaultShutdownSignals = []os.Signal{
	syscall.SIGINT,  // Interrupt(2): Triggered by Ctrl+C
	syscall.SIGTERM, // Terminated(15): Triggered by "kill <pid>"
}

func WaitForShutdown(ctx context.Context, callback func(ctx context.Context) error, listenedSignals ...os.Signal) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, listenedSignals...)

	signal := <-ch

	slog.Info("received shutdown signal", "signal", signal.String())

	if callback != nil {
		return callback(ctx)
	}

	return nil
}
