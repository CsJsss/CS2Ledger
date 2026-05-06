package logfx

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/utils/configfx"
)

type Logger struct{ *slog.Logger }

func NewLogger(cfg configfx.Config) *Logger {
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      parseLevel(cfg.Log.Level),
		TimeFormat: time.TimeOnly,
	})
	return &Logger{slog.New(handler)}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *Logger) WithComponent(name string) *Logger {
	return &Logger{l.With("component", name)}
}

func WithComponent(name string) fx.Option {
	return fx.Decorate(func(log *Logger) *Logger {
		return log.WithComponent(name)
	})
}

var Module = fx.Module("logfx", fx.Provide(NewLogger))
