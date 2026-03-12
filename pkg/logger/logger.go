package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	once sync.Once
)

func Init() {
	once.Do(func() {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}

		handler := slog.NewJSONHandler(os.Stdout, opts)
		logger := slog.New(handler)
		slog.SetDefault(logger)
	})
}
