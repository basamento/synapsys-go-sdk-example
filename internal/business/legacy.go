package business

import (
	"log/slog"
	"sync"
	"time"
)

// LegacyListener represents existing logic with Close rather than context.Context.
type LegacyListener struct {
	mu             sync.Mutex
	stop           chan struct{}
	closeRequested bool
}

func NewLegacyListener() *LegacyListener { return &LegacyListener{} }

func (l *LegacyListener) Run(logger *slog.Logger) error {
	l.mu.Lock()
	if l.closeRequested {
		l.closeRequested = false
		l.mu.Unlock()
		logger.Info("legacy listener observed an early close")
		return nil
	}
	stop := make(chan struct{})
	l.stop = stop
	l.mu.Unlock()

	logger.Info("legacy listener started")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	defer func() {
		l.mu.Lock()
		if l.stop == stop {
			l.stop = nil
		}
		l.mu.Unlock()
	}()

	for cycle := 1; ; cycle++ {
		select {
		case <-stop:
			logger.Info("legacy listener closed")
			return nil
		case <-ticker.C:
			logger.Info("legacy listener completed cycle", "cycle", cycle)
		}
	}
}

func (l *LegacyListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.stop != nil {
		close(l.stop)
		l.stop = nil
	} else {
		l.closeRequested = true
	}
	return nil
}
