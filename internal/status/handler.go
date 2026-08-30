// Package status provides a tiny standard-library HTTP status page.
package status

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/basamento/synapsys-go-sdk"
)

type view struct {
	Worker    string                   `json:"worker"`
	Enabled   bool                     `json:"enabled"`
	Processes []synapsys.ProcessStatus `json:"processes"`
}

// Handler exposes HTML at / and the same live process snapshots as JSON at /json.
func Handler(worker *synapsys.Worker) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /json", func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(response).Encode(snapshot(worker)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("GET /", func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(response, snapshot(worker)); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	})
	return mux
}

func snapshot(worker *synapsys.Worker) view {
	settings := worker.Settings()
	return view{Worker: settings.WorkerName, Enabled: settings.Enabled, Processes: worker.Processes()}
}

var page = template.Must(template.New("status").Funcs(template.FuncMap{
	"runID": func(value *int64) string {
		if value == nil {
			return "—"
		}
		return fmt.Sprintf("%d", *value)
	},
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Synapsys Go SDK example</title>
  <style>
    body { max-width: 960px; margin: 2rem auto; padding: 0 1rem; font: 15px system-ui, sans-serif; color: #1f2937; }
    h1 { margin-bottom: .25rem; } p { color: #4b5563; } table { width: 100%; border-collapse: collapse; }
    th, td { padding: .6rem; border-bottom: 1px solid #e5e7eb; text-align: left; }
    th { color: #6b7280; font-size: 12px; text-transform: uppercase; }
    code { font-family: ui-monospace, monospace; } .state { font-weight: 600; }
  </style>
</head>
<body>
  <h1>Synapsys Go SDK example</h1>
  <p>Worker <code>{{.Worker}}</code> · SDK enabled: {{.Enabled}} · {{len .Processes}} registered processes</p>
  <table><thead><tr><th>Name</th><th>Type</th><th>State</th><th>Ack token</th><th>Run ID</th></tr></thead>
  <tbody>{{range .Processes}}<tr><td><code>{{.Name}}</code></td><td>{{.Type}}</td><td class="state">{{.State}}</td><td>{{.AckToken}}</td><td><code>{{runID .RunID}}</code></td></tr>{{end}}</tbody></table>
</body>
</html>`))
