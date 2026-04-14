package nosos

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var DefaultShutdownSignals = []os.Signal{
	syscall.SIGINT,  // Interrupt(2): Triggered by Ctrl+C
	syscall.SIGTERM, // Terminated(15): Triggered by "kill <pid>"
}

type GracefulShutdownSetup struct {
	OnShutdown      func(ctx context.Context) error
	MaxWait         time.Duration
	ListenedSignals []os.Signal
}

func (g GracefulShutdownSetup) GetMaxWait() time.Duration {
	if int64(g.MaxWait) == 0 {
		return 60 * time.Second
	}
	return g.MaxWait
}

func (g GracefulShutdownSetup) Handle(ctx context.Context) error {
	if g.OnShutdown == nil {
		return nil
	}

	tctx, cancel := context.WithTimeout(ctx, g.GetMaxWait())
	defer cancel()

	ch := make(chan error)

	go func(ctx context.Context, ch chan<- error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("graceful shutdown panic", "recover", r)
			}
		}()

		if err := g.OnShutdown(ctx); err != nil {
			ch <- err
		}

		ch <- nil
	}(tctx, ch)

	select {
	case <-tctx.Done():
		return nil
	case err := <-ch:
		if err != nil {
			return err
		}

		return nil
	}
}

func GracefulShutdown(ctx context.Context, setup GracefulShutdownSetup) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, setup.ListenedSignals...)

	signal := <-ch

	slog.Info("received shutdown signal", "signal", signal.String())

	if setup.OnShutdown != nil {
		return setup.Handle(ctx)
	}

	return nil
}
