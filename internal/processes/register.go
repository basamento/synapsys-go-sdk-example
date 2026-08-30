// Package processes is the thin wiring layer between business logic and Synapsys.
package processes

import (
	"context"
	"time"

	"github.com/basamento/synapsys-go-sdk"
	"github.com/basamento/synapsys-go-sdk-example/internal/business"
)

// Register adds all 24 example processes. Nothing runs until Core sends a start.
func Register(worker *synapsys.Worker, services *business.Services) error {
	legacy := business.NewLegacyListener()
	legacyLog := worker.Logger("legacy-nightly-sync")
	cacheLog := worker.Logger("cache-warmer")

	return worker.Register(
		continuous(worker, services, "invoice-poller", "poll invoices", 2*time.Second),
		continuous(worker, services, "payment-reconciler", "reconcile payments", 3*time.Second),
		continuous(worker, services, "feed-subscriber", "receive feed events", 8*time.Second),
		synapsys.EndlessContext("cache-warmer", func(ctx context.Context) error {
			return services.RunContinuous(ctx, cacheLog, "warm cache", 2*time.Second)
		}, synapsys.WithOnStop(func() error { return services.CleanupCache(cacheLog) })),
		continuous(worker, services, "session-reaper", "reap sessions", 30*time.Second),
		continuous(worker, services, "metrics-publisher", "publish metrics", time.Second),
		synapsys.EndlessContext("dead-letter-retrier", func(ctx context.Context) error {
			return services.RunDeadLetterRetrier(ctx, worker.Logger("dead-letter-retrier"))
		}),
		synapsys.Endless("legacy-nightly-sync", func() error {
			return legacy.Run(legacyLog)
		}, synapsys.WithOnStop(legacy.Close)),
		continuous(worker, services, "queue-consumer", "consume queue messages", 2*time.Second),
		continuous(worker, services, "temperature-monitor", "sample temperature", 1500*time.Millisecond),
		synapsys.EndlessContext("slow-stopper", func(ctx context.Context) error {
			return services.RunSlowStopper(ctx, worker.Logger("slow-stopper"))
		}),
		synapsys.EndlessContext("crashing-watchdog", func(ctx context.Context) error {
			return services.RunCrashingWatchdog(ctx, worker.Logger("crashing-watchdog"))
		}),

		finite(worker, services, "monthly-report", 5, 350*time.Millisecond),
		finite(worker, services, "bulk-import", 20, 150*time.Millisecond),
		synapsys.ProgressiveContext("flaky-importer", func(ctx context.Context) error {
			return services.RunFlakyImporter(ctx, worker.Logger("flaky-importer"))
		}),
		finite(worker, services, "checksum-cruncher", 10, 200*time.Millisecond),
		finite(worker, services, "data-migration", 15, 300*time.Millisecond),
		finite(worker, services, "pdf-renderer", 8, 250*time.Millisecond),
		finite(worker, services, "search-index-rebuild", 12, 225*time.Millisecond),
		finite(worker, services, "email-blast", 60, 75*time.Millisecond),
		synapsys.ProgressiveContext("archive-compactor", func(ctx context.Context) error {
			return services.RunFinite(ctx, worker.Logger("archive-compactor"), "compact archives", 20, 200*time.Millisecond)
		}, synapsys.WithOnStop(func() error {
			worker.Logger("archive-compactor").Info("partial archive files removed")
			return nil
		})),
		synapsys.Progressive("tax-calculator", func() error {
			return services.RunFinite(context.Background(), worker.Logger("tax-calculator"), "calculate taxes", 6, 250*time.Millisecond)
		}),
		synapsys.ProgressiveContext("payroll-exporter", func(ctx context.Context) error {
			return services.RunPayrollExporter(ctx, worker.Logger("payroll-exporter"))
		}),
		synapsys.Progressive("uncancellable-batch", func() error {
			return services.RunFinite(context.Background(), worker.Logger("uncancellable-batch"), "uncancellable batch", 8, 500*time.Millisecond)
		}),
	)
}

func continuous(worker *synapsys.Worker, services *business.Services, name, activity string, interval time.Duration) synapsys.Process {
	logger := worker.Logger(name)
	return synapsys.EndlessContext(name, func(ctx context.Context) error {
		return services.RunContinuous(ctx, logger, activity, interval)
	})
}

func finite(worker *synapsys.Worker, services *business.Services, name string, steps int, delay time.Duration) synapsys.Process {
	logger := worker.Logger(name)
	return synapsys.ProgressiveContext(name, func(ctx context.Context) error {
		return services.RunFinite(ctx, logger, name, steps, delay)
	})
}
