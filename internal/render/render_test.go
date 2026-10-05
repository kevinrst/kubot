package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kevinrst/kubot/internal/model"
)

func testReport() model.Report {
	return model.Report{
		SchemaVersion: model.SchemaVersion,
		Cluster:       model.ClusterInfo{Context: "kind-test"},
		Status:        "critical",
		Issues: []model.Finding{
			{
				Severity:  model.SeverityCritical,
				Resource:  "pod/payments-api-xxx",
				Namespace: "default",
				Reason:    "pod_oom_killed",
				Message:   `Container "api" was OOMKilled (exit 137)`,
			},
			{
				Severity:  model.SeverityWarning,
				Resource:  "service/shop",
				Namespace: "default",
				Reason:    "service_no_endpoints",
				Message:   "Service selector matches no pods",
			},
		},
		Checked: []string{"pods", "services"},
	}
}

func TestPrintReport_JSONMatchesSchemaShape(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintReport(&buf, testReport(), Options{JSON: true}); err != nil {
		t.Fatalf("PrintReport: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if decoded["schema_version"] != model.SchemaVersion {
		t.Fatalf("schema_version missing: %+v", decoded)
	}
	if decoded["status"] != "critical" {
		t.Fatalf("status missing: %+v", decoded)
	}
}

func TestPrintText_compact(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintText(&buf, testReport(), false, Options{Width: 100}); err != nil {
		t.Fatalf("PrintText: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"connected", "kind-test", "Cluster health:", "/100",
		"CRITICAL", "WARNING", "●",
		"pod/payments-api-xxx", "service/shop",
		"checked", "--full", "--json",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("compact report missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Evidence:") {
		t.Errorf("compact report must not show evidence:\n%s", out)
	}
}

func TestPrintText_full(t *testing.T) {
	rep := testReport()
	rep.Issues[0].Evidence = map[string]any{"memory_limit": "64Mi"}
	rep.Issues[0].Recommendation = "Increase the limit."
	var buf bytes.Buffer
	if err := PrintText(&buf, rep, false, Options{Width: 100, Full: true}); err != nil {
		t.Fatalf("PrintText: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Evidence:", "memory_limit", "64Mi", "Recommendation:", "Increase the limit."} {
		if !strings.Contains(out, want) {
			t.Errorf("full report missing %q:\n%s", want, out)
		}
	}
}

func TestHealthScore(t *testing.T) {
	if got := healthScore(nil); got != 100 {
		t.Errorf("empty = %d, want 100", got)
	}
	fs := []model.Finding{{Severity: "critical"}, {Severity: "warning"}, {Severity: "note"}}
	if got := healthScore(fs); got != 86 {
		t.Errorf("got %d, want 86", got)
	}
}
