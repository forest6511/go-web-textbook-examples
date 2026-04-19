package observability

import (
	"context"
	"log/slog"
)

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(hs ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: hs}
}

func (m *MultiHandler) Enabled(
	ctx context.Context, l slog.Level,
) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(
	ctx context.Context, r slog.Record,
) error {
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(as []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(as)
	}
	return &MultiHandler{handlers: next}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: next}
}
