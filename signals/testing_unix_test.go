//go:build !windows

package signals

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSignalInjector_UnixMultipleHandlers(t *testing.T) {
	// Test SIGUSR1
	t.Run("SIGUSR1", func(t *testing.T) {
		manager := NewManager()
		injector := NewInjector(manager)

		usr1Received := make(chan bool, 1)

		_, _ = manager.Handle(syscall.SIGUSR1, func(ctx context.Context, sig os.Signal) error {
			usr1Received <- true
			return nil
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		go func() {
			_ = manager.Listen(ctx)
		}()

		if err := injector.WaitForListen(500 * time.Millisecond); err != nil {
			t.Fatalf("WaitForListen() failed: %v", err)
		}

		if err := injector.Inject(syscall.SIGUSR1); err != nil {
			t.Fatalf("Inject(SIGUSR1) failed: %v", err)
		}

		select {
		case <-usr1Received:
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("SIGUSR1 handler not called")
		}
	})

	// Test SIGUSR2
	t.Run("SIGUSR2", func(t *testing.T) {
		manager := NewManager()
		injector := NewInjector(manager)

		usr2Received := make(chan bool, 1)

		_, _ = manager.Handle(syscall.SIGUSR2, func(ctx context.Context, sig os.Signal) error {
			usr2Received <- true
			return nil
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		go func() {
			_ = manager.Listen(ctx)
		}()

		if err := injector.WaitForListen(500 * time.Millisecond); err != nil {
			t.Fatalf("WaitForListen() failed: %v", err)
		}

		if err := injector.Inject(syscall.SIGUSR2); err != nil {
			t.Fatalf("Inject(SIGUSR2) failed: %v", err)
		}

		select {
		case <-usr2Received:
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("SIGUSR2 handler not called")
		}
	})
}
