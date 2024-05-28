package logger

import (
	"bytes"
	"log/slog"
)

func NewTestLogger(opts *slog.HandlerOptions) (*slog.Logger, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	if opts == nil {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	}
	return slog.New(slog.NewJSONHandler(buf, opts)), buf
}
