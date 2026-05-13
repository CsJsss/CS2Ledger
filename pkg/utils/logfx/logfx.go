package logfx

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/utils/configfx"
)

type Logger struct{ *slog.Logger }

func NewLogger(cfg configfx.Config) *Logger {
	level := parseLevel(cfg.Log.Level)
	opts := &tint.Options{Level: level, TimeFormat: time.DateTime}

	var w io.Writer = os.Stderr
	if cfg.Log.Path != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Log.Path), 0755); err != nil {
			panic("logfx: failed to create log directory: " + err.Error())
		}
		f, err := os.OpenFile(cfg.Log.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic("logfx: failed to open log file: " + err.Error())
		}
		opts.NoColor = true
		w = io.MultiWriter(os.Stderr, f)
	}

	handler := tint.NewHandler(w, opts)
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

func NewNop() *Logger {
	return &Logger{slog.New(slog.NewTextHandler(io.Discard, nil))}
}

var Module = fx.Module("logfx", fx.Provide(NewLogger))
