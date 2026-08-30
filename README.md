# Synapsys Go SDK example

A vanilla Go application using
[`github.com/basamento/synapsys-go-sdk`](https://github.com/basamento/synapsys-go-sdk).
It registers 24 remotely controlled processes—twelve Endless and twelve
Progressive—and emits execution-scoped `slog` records from every process so
Synapsys Core's log ingestion can be exercised.

The business logic in `internal/business` imports only the Go standard library.
`internal/processes` is the thin Synapsys wiring layer. Nothing runs
automatically: every process starts idle and executes only after Core sends a
start command.

## Requirements

- Go 1.22 or later
- A reachable Synapsys Core instance
- A worker token created in Core

## Run

Set the deployment values and start the application.

Windows PowerShell:

```powershell
$env:SYNAPSYS_CORE_URL = "http://localhost:8080"
$env:SYNAPSYS_CORE_TOKEN = "syn_your_worker_token"
go run .
```

Linux or macOS:

```bash
export SYNAPSYS_CORE_URL=http://localhost:8080
export SYNAPSYS_CORE_TOKEN=syn_your_worker_token
go run .
```

`.env.example` lists all useful settings as a deployment reference; Go does not
load `.env` files automatically. Open <http://127.0.0.1:3005> for the local status
page and <http://127.0.0.1:3005/json> for its JSON form. In Core, find the
`go-sdk-example` worker and start or stop processes there.

## Process catalogue

### Endless

| Process | Demonstrates |
| --- | --- |
| `invoice-poller` | Canonical cooperative polling loop. |
| `payment-reconciler` | Shared ordinary business service. |
| `feed-subscriber` | A long wait interrupted through `context.Context`. |
| `cache-warmer` | An Endless `on-stop` cleanup hook. |
| `session-reaper` | A 30-second wait that still stops promptly. |
| `metrics-publisher` | A captured log record every second. |
| `dead-letter-retrier` | Handled errors that do not fail the process. |
| `legacy-nightly-sync` | Context-free legacy logic unblocked by `Close`. |
| `queue-consumer` | A conventional consumer loop. |
| `temperature-monitor` | A small cooperative monitor. |
| `slow-stopper` | A non-interruptible safe point that remains `stopping`. |
| `crashing-watchdog` | A deliberate panic recovered as `failed`. |

### Progressive

| Process | Demonstrates |
| --- | --- |
| `monthly-report` | Normal finite work. |
| `bulk-import` | Cooperative cancellation between rows. |
| `flaky-importer` | Deliberate errors; every third invocation succeeds. |
| `checksum-cruncher` | Cancellable CPU-like rounds. |
| `data-migration` | A long multi-step batch. |
| `pdf-renderer` | Mixed finite steps. |
| `search-index-rebuild` | Concurrent finite work isolated in its own goroutine. |
| `email-blast` | Sixty small cancellable operations. |
| `archive-compactor` | Progressive cleanup only after an in-flight stop. |
| `tax-calculator` | Context-free finite business logic. |
| `payroll-exporter` | An error captured and reported as `failed`. |
| `uncancellable-batch` | No context; remains `stopping` until it returns. |

The deliberate failures are part of the example. They demonstrate that an error
or panic fails one process without taking down the worker.

## Logging

Each declaration obtains `worker.Logger(processName)` and passes that standard
`*slog.Logger` into ordinary business logic. The SDK tees Info-or-higher records
to the application's logger and Core while the corresponding run is active.
There is no global stdout/stderr replacement, so concurrent goroutines cannot be
misattributed.

Do not log tokens, passwords, personal information, or other secrets from a
managed process: captured process logs are transmitted to Core.

## Run without Core

```powershell
$env:SYNAPSYS_ENABLED = "false"
go run .
```

Disabled mode performs no validation, network access, heartbeat, or capture. The
local page still displays all 24 registered processes.

## Verify

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
```

## Layout

```text
main.go                         application lifecycle and signal handling
internal/business/             reusable logic with no Synapsys import
internal/processes/register.go thin wiring for all 24 declarations
internal/status/handler.go     standard-library HTML and JSON status endpoint
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
