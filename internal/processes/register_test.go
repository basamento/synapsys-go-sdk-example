package processes

import (
	"testing"

	"github.com/basamento/synapsys-go-sdk"
	"github.com/basamento/synapsys-go-sdk-example/internal/business"
)

func TestRegisterAddsTwelveProcessesOfEachType(t *testing.T) {
	worker, err := synapsys.New(synapsys.WithEnabled(false))
	if err != nil {
		t.Fatal(err)
	}
	if err := Register(worker, business.NewServices()); err != nil {
		t.Fatal(err)
	}

	counts := map[synapsys.ProcessType]int{}
	names := map[string]struct{}{}
	for _, process := range worker.Processes() {
		counts[process.Type]++
		if _, duplicate := names[process.Name]; duplicate {
			t.Fatalf("duplicate process name %q", process.Name)
		}
		names[process.Name] = struct{}{}
		if process.State != synapsys.StateIdle {
			t.Fatalf("%s started in state %s", process.Name, process.State)
		}
	}
	if counts[synapsys.TypeEndless] != 12 || counts[synapsys.TypeProgressive] != 12 {
		t.Fatalf("process counts = %#v, want 12 endless and 12 progressive", counts)
	}
}
