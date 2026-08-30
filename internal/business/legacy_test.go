package business

import (
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestLegacyListenerHandlesCloseBeforeRun(t *testing.T) {
	listener := NewLegacyListener()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- listener.Run(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("listener missed an early close")
	}
}
