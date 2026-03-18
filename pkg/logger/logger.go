package logger

import (
	"log/slog"
	"os"
	"strings"
)

func New(level string, appEnv string) *slog.Logger {
	options := &slog.HandlerOptions{Level: parseLevel(level)}

	if strings.EqualFold(appEnv, "production") {
		return slog.New(slog.NewJSONHandler(os.Stdout, options))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, options))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
