package main

import (
	"testing"

	"github.com/kevinrst/kubot/internal/model"
)

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
