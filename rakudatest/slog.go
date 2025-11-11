package rakudatest

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
)

// THandler is a slog.Handler that writes log records to a *testing.T.
type THandler struct {
	t      *testing.T
	level  slog.Level
	attrs  []slog.Attr
	prefix string
}

// NewTHandler creates a new THandler that writes to the given testing object
// at the specified log level.
func NewTHandler(t *testing.T, level slog.Level) *THandler {
	return &THandler{
		t:     t,
		level: level,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *THandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle formats the log record and writes it to the testing object using t.Logf.
func (h *THandler) Handle(_ context.Context, r slog.Record) error {
	var attrs string
	r.Attrs(func(attr slog.Attr) bool {
		attrs += fmt.Sprintf(" %s=%v", attr.Key, attr.Value.Any())
		return true
	})
	h.t.Logf("%s: %s%s", r.Level, r.Message, attrs)
	return nil
}

// WithAttrs returns a new THandler with the given attributes.
func (h *THandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.attrs = append(newHandler.attrs, attrs...)
	return &newHandler
}

// WithGroup returns a new THandler with the given group name.
func (h *THandler) WithGroup(name string) slog.Handler {
	newHandler := *h
	newHandler.prefix += name + "."
	return &newHandler
}
