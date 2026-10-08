package bootstrap

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitDrainsServerErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	httpErrCh := make(chan error, 1)
	grpcErrCh := make(chan error, 1)
	httpErrCh <- errors.New("http down")

	go func() {
		<-ctx.Done()
		close(httpErrCh)
		close(grpcErrCh)
	}()

	done := make(chan struct{})
	go func() {
		wait(cancel, httpErrCh, grpcErrCh, time.Second)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("wait did not return after the server channels closed")
	}
	if ctx.Err() == nil {
		t.Fatal("wait did not cancel the process context")
	}
}
