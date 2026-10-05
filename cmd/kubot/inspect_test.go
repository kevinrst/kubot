package main

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kevinrst/kubot/internal/k8s"
	"github.com/kevinrst/kubot/internal/model"
)

func TestSuggestWorkload_typo(t *testing.T) {
	snap := &k8s.Snapshot{
		Deployments: []appsv1.Deployment{
			{ObjectMeta: metav1.ObjectMeta{Name: "payments-api", Namespace: "default"}},
		},
	}
	got := suggestWorkload(snap, "payment-api")
	if len(got) != 1 || got[0] != "payments-api" {
		t.Fatalf("expected payments-api suggestion, got %v", got)
	}
	if got := suggestWorkload(snap, "totally-unrelated-xyz"); len(got) != 0 {
		t.Fatalf("expected no suggestions, got %v", got)
	}
}

func TestExitCode(t *testing.T) {
	f := func(sev string) []model.Finding {
		return []model.Finding{{Severity: sev, Resource: "pod/x", Reason: "test"}}
	}
	cases := []struct {
		findings []model.Finding
		failOn   string
		want     int
	}{
		{nil, "critical", exitClean},
		{f("warning"), "critical", exitClean},
		{f("critical"), "critical", exitCritical},
		{f("warning"), "warn", exitWarn},
		{f("critical"), "warn", exitCritical},
		{f("critical"), "none", exitClean},
		{f("note"), "info", exitWarn},
		{f("note"), "warn", exitClean},
	}
	for _, c := range cases {
		if got := exitCode(c.findings, c.failOn); got != c.want {
			t.Errorf("exitCode(%+v,%q)=%d want %d", c.findings, c.failOn, got, c.want)
		}
	}
}

func TestValidFailOnAndFormat(t *testing.T) {
	for _, ok := range []string{"critical", "warn", "info", "none"} {
		if !validFailOn(ok) {
			t.Errorf("validFailOn(%q)=false", ok)
		}
	}
	if validFailOn("bogus") {
		t.Error("validFailOn(bogus)=true")
	}
	for _, ok := range []string{"text", "json"} {
		if !validFormat(ok) {
			t.Errorf("validFormat(%q)=false", ok)
		}
	}
	if validFormat("sarif") {
		t.Error("validFormat(sarif) should be false until the format ships")
	}
}
