package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kevinrst/kubot/internal/model"
)

func TestPrintReport_JSONMatchesSchemaShape(t *testing.T) {
	rep := model.Report{
		SchemaVersion: model.SchemaVersion,
		Status:        "critical",
		Issues: []model.Finding{{
			Severity:  model.SeverityCritical,
			Resource:  "pod/payments-api-xxx",
			Namespace: "default",
			Reason:    "pod_oom_killed",
			Message:   "Container was OOMKilled",
		}},
		Checked: []string{"pods"},
	}
	var buf bytes.Buffer
	if err := PrintReport(&buf, rep, Options{JSON: true}); err != nil {
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

func TestPrintReport_TextFindingsFirst(t *testing.T) {
	rep := model.Report{
		SchemaVersion: model.SchemaVersion,
		Status:        "warning",
		Issues: []model.Finding{{
			Severity:  model.SeverityWarning,
			Resource:  "service/shop",
			Namespace: "default",
			Reason:    "service_no_endpoints",
			Message:   "Service has no ready endpoints",
		}},
		Checked: []string{"pods", "services"},
	}
	var buf bytes.Buffer
	if err := PrintReport(&buf, rep, Options{NoColor: true}); err != nil {
		t.Fatalf("PrintReport: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "WARNING") || !strings.Contains(out, "service/shop") {
		t.Fatalf("text report missing severity/resource:\n%s", out)
	}
	if !strings.Contains(out, "checked") {
		t.Fatalf("text report missing checked line:\n%s", out)
	}
}
