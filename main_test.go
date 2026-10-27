package main

import (
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// serveTestSetup starts serve on a random port and returns its address, the
// signal channel, and a channel carrying serve's return value.
func serveTestSetup(t *testing.T, handler http.Handler) (string, chan os.Signal, <-chan error) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: time.Second}
	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(slog.New(slog.DiscardHandler), srv, ln, sigCh)
	}()
	return "http://" + ln.Addr().String(), sigCh, errCh
}

// A termination signal drains in-flight requests: serve returns only after
// they finish, and the client still gets a complete response.
func TestServeDrainsInFlightRequests(t *testing.T) {
	var handlerFinished atomic.Bool
	started := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		time.Sleep(300 * time.Millisecond)
		handlerFinished.Store(true)
		_, _ = w.Write([]byte("drained"))
	})

	addr, sigCh, errCh := serveTestSetup(t, handler)

	type result struct {
		body string
		err  error
	}
	resCh := make(chan result, 1)
	go func() {
		resp, err := http.Get(addr + "/slow")
		if err != nil {
			resCh <- result{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		resCh <- result{body: string(body), err: err}
	}()

	<-started // request is in flight
	sigCh <- syscall.SIGTERM

	if err := <-errCh; err != nil {
		t.Fatalf("serve returned %v, want nil", err)
	}
	if !handlerFinished.Load() {
		t.Error("serve returned before the in-flight request finished")
	}

	res := <-resCh
	if res.err != nil {
		t.Fatalf("in-flight request failed: %v", res.err)
	}
	if res.body != "drained" {
		t.Errorf("body = %q, want %q", res.body, "drained")
	}
}

// serve owns the signal handler, so it has to reap it on every exit path, not
// only the one a termination signal drives.
func TestServeStopsSignalHandlerOnListenerError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: http.NewServeMux(), ReadHeaderTimeout: time.Second}
	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)

	go func() {
		errCh <- serve(slog.New(slog.DiscardHandler), srv, ln, sigCh)
	}()

	// Break Accept so Serve fails with something other than ErrServerClosed.
	if err := ln.Close(); err != nil {
		t.Fatalf("closing listener: %v", err)
	}

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("serve returned nil, want the listener error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after the listener closed")
	}

	// A stranded handler would still be reading sigCh and consume this.
	sigCh <- syscall.SIGTERM
	time.Sleep(100 * time.Millisecond)
	if len(sigCh) != 1 {
		t.Error("signal handler still running after serve returned")
	}
}
