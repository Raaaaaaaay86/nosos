# nosos

`nosos` provides OS-level utilities for the nos ecosystem. Currently it ships `GracefulShutdown`, a helper that blocks until the process receives a termination signal and then calls a user-supplied cleanup function within a configurable deadline.

## Installation

```bash
go get github.com/raaaaaaaay86/nosos
```

## GracefulShutdown

Blocks the current goroutine until one of the configured OS signals (`SIGINT` / `SIGTERM` by default) is received, then invokes `OnShutdown` with a timeout context bounded by `MaxWait`.

### Basic Usage

```go
package main

import (
    "context"
    "log/slog"
    "time"

    "github.com/raaaaaaaay86/nosos"
)

func main() {
    // ... start your server, workers, etc.

    err := nosos.GracefulShutdown(context.Background(), nosos.GracefulShutdownSetup{
        ListenedSignals: nosos.DefaultShutdownSignals, // SIGINT + SIGTERM
        MaxWait:         30 * time.Second,
        OnShutdown: func(ctx context.Context) error {
            slog.Info("shutting down gracefully...")
            // stop HTTP server, close DB connections, flush queues, etc.
            return nil
        },
    })
    if err != nil {
        slog.Error("graceful shutdown error", "error", err)
    }
}
```

### With a Fiber HTTP Server

```go
func Run(ctx context.Context) error {
    app := nosfiber.NewFiberApp(fiber.Config{}, fiber.ListenConfig{})

    if err := app.StartAsync(ctx, 8080); err != nil {
        return err
    }

    return nosos.GracefulShutdown(ctx, nosos.GracefulShutdownSetup{
        ListenedSignals: nosos.DefaultShutdownSignals,
        MaxWait:         15 * time.Second,
        OnShutdown: func(ctx context.Context) error {
            return app.Shutdown()
        },
    })
}
```

### With Multiple Components

```go
return nosos.GracefulShutdown(ctx, nosos.GracefulShutdownSetup{
    ListenedSignals: nosos.DefaultShutdownSignals,
    MaxWait:         30 * time.Second,
    OnShutdown: func(ctx context.Context) error {
        rabbitManager.Stop()
        kafkaManager.Stop()
        httpServer.Shutdown()
        return nil
    },
})
```

## API Reference

### `GracefulShutdownSetup`

| Field             | Type                              | Description                                                              |
|-------------------|-----------------------------------|--------------------------------------------------------------------------|
| `OnShutdown`      | `func(ctx context.Context) error` | Cleanup callback, called after signal is received                        |
| `MaxWait`         | `time.Duration`                   | Maximum time to wait for `OnShutdown` to complete (default: 60s)         |
| `ListenedSignals` | `[]os.Signal`                     | OS signals to listen for; use `DefaultShutdownSignals` for SIGINT+SIGTERM|

### `DefaultShutdownSignals`

```go
var DefaultShutdownSignals = []os.Signal{
    syscall.SIGINT,  // Ctrl+C
    syscall.SIGTERM, // kill <pid>
}
```

### `GracefulShutdown`

```go
func GracefulShutdown(ctx context.Context, setup GracefulShutdownSetup) error
```

Blocks until a signal is received, then runs `OnShutdown` inside a bounded context. Returns any error from `OnShutdown`.
