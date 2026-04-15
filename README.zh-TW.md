# nosos

`nosos` 為 nos 生態系提供作業系統層級的工具函式。目前包含 `GracefulShutdown`——一個阻塞至程序收到終止信號後、在可設定期限內執行使用者提供的清理函式的輔助工具。

## 安裝

```bash
go get github.com/raaaaaaaay86/nosos
```

## GracefulShutdown

阻塞目前的 goroutine，直到收到設定的 OS 信號（預設為 `SIGINT` / `SIGTERM`），然後以受 `MaxWait` 限制的 timeout context 呼叫 `OnShutdown`。

### 基本用法

```go
package main

import (
    "context"
    "log/slog"
    "time"

    "github.com/raaaaaaaay86/nosos"
)

func main() {
    // ... 啟動 server、worker 等

    err := nosos.GracefulShutdown(context.Background(), nosos.GracefulShutdownSetup{
        ListenedSignals: nosos.DefaultShutdownSignals, // SIGINT + SIGTERM
        MaxWait:         30 * time.Second,
        OnShutdown: func(ctx context.Context) error {
            slog.Info("正在優雅關機...")
            // 停止 HTTP server、關閉 DB 連線、清空佇列等
            return nil
        },
    })
    if err != nil {
        slog.Error("graceful shutdown 發生錯誤", "error", err)
    }
}
```

### 搭配 Fiber HTTP Server

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

### 搭配多個元件

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

## API 說明

### `GracefulShutdownSetup`

| 欄位              | 型別                              | 說明                                                            |
|-------------------|-----------------------------------|-----------------------------------------------------------------|
| `OnShutdown`      | `func(ctx context.Context) error` | 收到信號後執行的清理 callback                                   |
| `MaxWait`         | `time.Duration`                   | 等待 `OnShutdown` 完成的最長時間（預設：60s）                   |
| `ListenedSignals` | `[]os.Signal`                     | 監聽的 OS 信號；使用 `DefaultShutdownSignals` 取得 SIGINT+SIGTERM |

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

阻塞至信號到來，然後在有限時間的 context 內執行 `OnShutdown`，並回傳其錯誤。
