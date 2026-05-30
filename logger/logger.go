package logger

import (
	"log/slog"
	"os"

	"github.com/walnuts1018/esp32-thermohygrometer-exporter/config"
)

func CreateLogger(level slog.Level, logType config.LogType) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if logType == config.LogTypeJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
