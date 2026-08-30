package status

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/basamento/synapsys-go-sdk"
)

func TestHandlerReportsRegisteredProcesses(t *testing.T) {
	worker, err := synapsys.New(synapsys.WithEnabled(false), synapsys.WithWorkerName("test-worker"))
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Register(synapsys.Progressive("report", func() error { return nil })); err != nil {
		t.Fatal(err)
	}

	jsonResponse := httptest.NewRecorder()
	Handler(worker).ServeHTTP(jsonResponse, httptest.NewRequest(http.MethodGet, "/json", nil))
	if jsonResponse.Code != http.StatusOK {
		t.Fatalf("JSON status = %d", jsonResponse.Code)
	}
	var body view
	if err := json.Unmarshal(jsonResponse.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Worker != "test-worker" || len(body.Processes) != 1 || body.Processes[0].Name != "report" {
		t.Fatalf("unexpected JSON body: %#v", body)
	}

	htmlResponse := httptest.NewRecorder()
	Handler(worker).ServeHTTP(htmlResponse, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(htmlResponse.Body.String(), "test-worker") || !strings.Contains(htmlResponse.Body.String(), "report") {
		t.Fatalf("HTML did not contain worker and process: %s", htmlResponse.Body.String())
	}
}
