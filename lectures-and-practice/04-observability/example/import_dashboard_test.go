package example_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestDashboardImport(t *testing.T) {
	script, err := filepath.Abs("import-dashboard.sh")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("dashboard.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]any
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := calls.Add(1)
		var payload struct {
			Dashboard map[string]any
			Overwrite bool
		}
		if r.Method != "POST" || r.URL.Path != "/api/dashboards/db" || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || !reflect.DeepEqual(payload.Dashboard, want) {
			t.Errorf("dashboard payload differs from dashboard.json: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if call > 1 && !payload.Overwrite {
			http.Error(w, "dashboard already exists", http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("GRAFANA_URL", server.URL+"/")
	for _, check := range []struct {
		overwrite string
		ok        bool
	}{{"", true}, {"false", false}, {"true", true}, {"invalid", false}} {
		t.Setenv("OVERWRITE", check.overwrite)
		command := exec.Command("sh", script)
		command.Dir = t.TempDir() // The dashboard path must not depend on the caller's cwd.
		output, err := command.CombinedOutput()
		if (err == nil) != check.ok {
			t.Fatalf("OVERWRITE=%q: err=%v, output=%s", check.overwrite, err, output)
		}
	}
	if calls.Load() != 3 {
		t.Fatalf("invalid OVERWRITE must not reach Grafana; got %d requests", calls.Load())
	}
}
