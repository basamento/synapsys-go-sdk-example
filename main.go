package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/basamento/synapsys-go-sdk"
	"github.com/basamento/synapsys-go-sdk-example/internal/business"
	processwiring "github.com/basamento/synapsys-go-sdk-example/internal/processes"
	"github.com/basamento/synapsys-go-sdk-example/internal/status"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx); err != nil {
		slog.Error("application stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	worker, err := synapsys.New(synapsys.WithWorkerName("go-sdk-example"))
	if err != nil {
		return err
	}
	if err := processwiring.Register(worker, business.NewServices()); err != nil {
		return err
	}
	if err := worker.Start(); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              ":3005",
		Handler:           status.Handler(worker),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverError := make(chan error, 1)
	go func() {
		slog.Info("local status page listening", "url", "http://127.0.0.1:3005")
		serverError <- server.ListenAndServe()
	}()

	var runErr error
	select {
	case <-ctx.Done():
	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			runErr = err
		}
	}

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverErr := server.Shutdown(shutdown)
	workerErr := worker.Stop(shutdown)
	return errors.Join(runErr, serverErr, workerErr)
}
