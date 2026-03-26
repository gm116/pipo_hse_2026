package logging

import (
	"log/slog"
	"os"
	"strings"
)

func New(levelRaw string) *slog.Logger {
	var level slog.Level
	switch strings.ToUpper(levelRaw) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(h)
}
