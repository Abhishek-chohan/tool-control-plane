package observability

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
)

// NewLogger builds the process logger. Format is "json" or "text" (default
// text); JSON is the intended production setting so log aggregators get
// structured records instead of scraped free text.
func NewLogger(format string, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	var handler slog.Handler
	if strings.EqualFold(strings.TrimSpace(format), "json") {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}
	return slog.New(handler)
}

// RouteStdLog bridges the standard library logger through the slog handler:
// existing log.Printf calls across the services become structured records
// at Info level without touching every call site. Fatal-level std log
// helpers must not be used after this (they would bypass the handler and
// skip defers anyway); shutdown paths own their exits explicitly.
func RouteStdLog(logger *slog.Logger) {
	log.SetOutput(&stdLogWriter{logger: logger})
	log.SetFlags(0)
}

// stdLogWriter adapts an io.Writer interface onto a slog handler.
type stdLogWriter struct {
	logger *slog.Logger
}

func (w *stdLogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	if msg == "" {
		return len(p), nil
	}
	w.logger.Log(context.Background(), slog.LevelInfo, msg)
	return len(p), nil
}

// LogFormatFromEnv reads TOOLPLANE_LOG_FORMAT (json|text; default text).
func LogFormatFromEnv() string {
	return strings.TrimSpace(os.Getenv("TOOLPLANE_LOG_FORMAT"))
}
