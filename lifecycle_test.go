package hollow

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestStartupHooks(t *testing.T) {
	app := &App{}
	stopErr := errors.New("stop startup")
	var order []string

	app.Startup(
		func() error {
			order = append(order, "start-1")
			return nil
		},
		func() error {
			order = append(order, "start-2")
			return stopErr
		},
		func() error {
			order = append(order, "start-3")
			return nil
		},
	)

	err := app.runStartupHooks()
	if !errors.Is(err, stopErr) {
		t.Fatalf("error=%v want=%v", err, stopErr)
	}
	want := []string{"start-1", "start-2"}
	if !slices.Equal(order, want) {
		t.Fatalf("order=%v want=%v", order, want)
	}
}

func TestShutdownHooks(t *testing.T) {
	app := &App{}
	firstErr := errors.New("first shutdown")
	lastErr := errors.New("last shutdown")
	var order []string

	app.Shutdown(
		func(context.Context) error {
			order = append(order, "stop-1")
			return firstErr
		},
		func(context.Context) error {
			order = append(order, "stop-2")
			return nil
		},
		func(context.Context) error {
			order = append(order, "stop-3")
			return lastErr
		},
	)

	err := app.runShutdownHooks(context.Background())
	if !errors.Is(err, firstErr) || !errors.Is(err, lastErr) {
		t.Fatalf("error=%v want joined errors %v and %v", err, firstErr, lastErr)
	}
	want := []string{"stop-3", "stop-2", "stop-1"}
	if !slices.Equal(order, want) {
		t.Fatalf("order=%v want=%v", order, want)
	}
}
