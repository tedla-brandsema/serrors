package serrors

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestStructuredError(t *testing.T) {
	ctx := context.Background()
	baseErr := errors.New("failed to connect")

	err := New(ctx, "CreateUser", "could not create user", baseErr, map[string]any{
		"user_id": 1234,
		"retry":   true,
	})

	fmt.Println(err)

	var se *StructuredError
	if !errors.As(err, &se) {
		t.Fatal("expected StructuredError")
	}

	if se.Fields["user_id"] != 1234 {
		t.Errorf("expected user_id=1234, got %v", se.Fields["user_id"])
	}
}
