package serrors

import (
	"fmt"
	"log/slog"
)

type ErrAttr int

const (
	Cause ErrAttr = iota
	Operation
)

var errAttrStrings = [...]string{
	"cause",
	"op",
}

func (e ErrAttr) String() string {
	if e < 0 || int(e) > e.Size() {
		return fmt.Sprintf("unknown ErrAttr(%d)", e)
	}
	return errAttrStrings[e]
}

func (e ErrAttr) Size() int {
	return len(errAttrStrings)
}

// StructuredError represents a structured, optionally correlated error.
type StructuredError struct {
	Msg    string
	Op     string
	Err    error
	Fields []slog.Attr
}

func (e *StructuredError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Msg
}

func (e *StructuredError) Unwrap() error {
	return e.Err
}

func (e *StructuredError) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0)
	attrs = append(attrs, e.Fields...)
	attrs = append(attrs, slog.String(Operation.String(), e.Op))

	if e.Err != nil {
		attrs = append(attrs, slog.String(Cause.String(), e.Err.Error()))
	}

	return slog.GroupValue(attrs...)
}

// New creates a StructuredError with optional span correlation.
func New(op, msg string, err error, fields ...slog.Attr) error {
	return &StructuredError{
		Msg:    msg,
		Op:     op,
		Err:    err,
		Fields: fields,
	}
}
