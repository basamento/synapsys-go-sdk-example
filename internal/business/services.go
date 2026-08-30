// Package business contains ordinary application logic with no Synapsys dependency.
package business

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Services holds small pieces of shared state used by the example processes.
type Services struct {
	mu   sync.Mutex
	runs map[string]int
}

func NewServices() *Services {
	return &Services{runs: make(map[string]int)}
}

// RunContinuous models a listener, consumer, or polling loop.
func (s *Services) RunContinuous(ctx context.Context, logger *slog.Logger, activity string, interval time.Duration) error {
	run := s.nextRun(activity)
	logger.Info("business loop started", "activity", activity, "invocation", run)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	cycle := 0
	for {
		select {
		case <-ctx.Done():
			logger.Info("business loop observed cancellation", "activity", activity, "cycles", cycle)
			return ctx.Err()
		case <-ticker.C:
			cycle++
			logger.Info("business loop completed cycle", "activity", activity, "cycle", cycle)
		}
	}
}

// RunDeadLetterRetrier demonstrates handled errors that do not fail a process.
func (s *Services) RunDeadLetterRetrier(ctx context.Context, logger *slog.Logger) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if attempt%3 == 0 {
				logger.Error("dead-letter delivery failed; retry remains active", "attempt", attempt)
				continue
			}
			logger.Info("dead-letter delivered", "attempt", attempt)
		}
	}
}

// RunSlowStopper has a safe point that takes time after cancellation is requested.
func (s *Services) RunSlowStopper(ctx context.Context, logger *slog.Logger) error {
	logger.Info("slow operation entered non-interruptible section")
	select {
	case <-ctx.Done():
		time.Sleep(2 * time.Second)
		logger.Info("slow operation reached safe stop point")
		return ctx.Err()
	case <-time.After(20 * time.Second):
		return nil
	}
}

// RunCrashingWatchdog deliberately panics on two out of every three runs.
func (s *Services) RunCrashingWatchdog(ctx context.Context, logger *slog.Logger) error {
	run := s.nextRun("crashing-watchdog")
	logger.Info("watchdog started", "invocation", run)
	if err := wait(ctx, 750*time.Millisecond); err != nil {
		return err
	}
	if run%3 != 0 {
		panic(fmt.Sprintf("deliberate watchdog panic on invocation %d", run))
	}
	return s.RunContinuous(ctx, logger, "healthy-watchdog", time.Second)
}

// RunFinite models one finite unit of progressive work.
func (s *Services) RunFinite(ctx context.Context, logger *slog.Logger, job string, steps int, delay time.Duration) error {
	run := s.nextRun(job)
	logger.Info("finite job started", "job", job, "invocation", run, "steps", steps)
	for step := 1; step <= steps; step++ {
		if err := wait(ctx, delay); err != nil {
			logger.Info("finite job cancelled", "job", job, "completedSteps", step-1)
			return err
		}
		logger.Info("finite job completed step", "job", job, "step", step, "total", steps)
	}
	logger.Info("finite job completed", "job", job)
	return nil
}

// RunFlakyImporter deliberately fails on two out of every three invocations.
func (s *Services) RunFlakyImporter(ctx context.Context, logger *slog.Logger) error {
	run := s.nextRun("flaky-importer")
	logger.Info("flaky import started", "invocation", run)
	if err := wait(ctx, time.Second); err != nil {
		return err
	}
	if run%3 != 0 {
		return fmt.Errorf("deliberate import failure on invocation %d", run)
	}
	logger.Info("flaky import succeeded", "invocation", run)
	return nil
}

// RunPayrollExporter deliberately returns an unexpected business error.
func (s *Services) RunPayrollExporter(ctx context.Context, logger *slog.Logger) error {
	logger.Info("payroll export opened")
	if err := wait(ctx, time.Second); err != nil {
		return err
	}
	return errors.New("deliberate payroll export failure: downstream rejected the file")
}

// CleanupCache is an ordinary cleanup hook.
func (s *Services) CleanupCache(logger *slog.Logger) error {
	logger.Info("cache resources released")
	return nil
}

func (s *Services) nextRun(name string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs[name]++
	return s.runs[name]
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
