package serrors

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"testing"
)

type LoggerFunc func(w io.Writer) *slog.Logger

var ErrConnect = errors.New("failed to connect")

func TestStructuredError(t *testing.T) {
	tests := []struct {
		name       string
		newLogger  LoggerFunc
		err        error
		wantFields map[string]any
		wantMsg    string
		wantUnwrap error
	}{
		{
			name: "WithFieldsAndCause",
			newLogger: func(w io.Writer) *slog.Logger {
				return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{AddSource: true}))
			},
			err: New("CreateUser",
				"could not create user",
				ErrConnect,
				[]slog.Attr{
					slog.Int("user_id", 1234),
					slog.Bool("retry", true),
				}...,
			),
			wantFields: map[string]any{
				"op":      "CreateUser",
				"cause":   "failed to connect",
				"user_id": int64(1234),
				"retry":   true,
			},
			wantMsg:    "CreateUser: failed to connect",
			wantUnwrap: ErrConnect,
		},
		{
			name: "WithoutCause",
			newLogger: func(w io.Writer) *slog.Logger {
				return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: false}))
			},
			err: New("LoadConfig",
				"missing file",
				nil,
				[]slog.Attr{
					slog.String("file", "config.yaml"),
				}...,
			),
			wantFields: map[string]any{
				"op":   "LoadConfig",
				"file": "config.yaml",
			},
			wantMsg:    "missing file",
			wantUnwrap: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var se *StructuredError
			if !errors.As(test.err, &se) {
				t.Fatal("expected error to be StructuredError")
			}

			if got := se.Error(); got != test.wantMsg {
				t.Errorf("Error(): got %q, want %q", got, test.wantMsg)
			}

			if test.wantUnwrap != nil && !errors.Is(test.err, test.wantUnwrap) {
				t.Errorf("Unwrap(): expected to match %q", test.wantUnwrap)
			}

			val := se.LogValue()
			if val.Kind() != slog.KindGroup {
				t.Fatalf("LogValue() expected GroupValue, got %v", val.Kind())
			}

			attrs := val.Group()
			attrMap := make(map[string]any, len(attrs))
			for _, attr := range attrs {
				attrMap[attr.Key] = attr.Value.Any()
			}

			for key, wantVal := range test.wantFields {
				gotVal, ok := attrMap[key]
				if !ok {
					t.Errorf("missing attribute %q in structured log", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("attribute  %q: got %v, want %v", key, gotVal, wantVal)
				}
			}
			loggingIntegration(test.newLogger, se, t)
		})
	}
}

func loggingIntegration(lf LoggerFunc, se *StructuredError, t *testing.T) {
	var buf bytes.Buffer
	logger := lf(&buf)
	logger.Error(se.Error(), slog.Any("error", se))
	t.Log("log output:", buf.String())
}
