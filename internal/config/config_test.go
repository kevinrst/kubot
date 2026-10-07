package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTOML(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), ".kubot.toml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadIgnore(t *testing.T) {
	p := writeTOML(t, `
[[ignore]]
finding = "pod_missing_resources"
object = "kube-system/*"
reason = "platform manages its own resources"
`)
	c, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(c.Ignore) != 1 || c.Ignore[0].Finding != "pod_missing_resources" {
		t.Fatalf("got %+v", c.Ignore)
	}
}

func TestLoadRejectsMissingReason(t *testing.T) {
	p := writeTOML(t, "[[ignore]]\nfinding = \"x\"\n")
	if _, err := Load(p); err == nil {
		t.Fatal("expected error for reasonless rule")
	}
}

func TestLoadRejectsCredentials(t *testing.T) {
	p := writeTOML(t, "password = \"hunter2\"\n")
	if _, err := Load(p); err == nil {
		t.Fatal("expected error for credential key")
	}
}

func TestMatch(t *testing.T) {
	c := &Config{Ignore: []Ignore{
		{Finding: "pod_missing_resources", Object: "kube-system/*", Reason: "platform"},
		{Finding: "pod_crashloop_backoff", Reason: "known flake"},
	}}
	if _, ok := c.Match("pod_missing_resources", "kube-system", "pod/coredns-x"); !ok {
		t.Error("namespace/resource glob should match")
	}
	if _, ok := c.Match("pod_missing_resources", "default", "pod/coredns-x"); ok {
		t.Error("wrong namespace must not match")
	}
	if _, ok := c.Match("pod_crashloop_backoff", "default", "pod/x"); !ok {
		t.Error("empty object should match everything")
	}
	if _, ok := c.Match("other", "default", "pod/x"); ok {
		t.Error("wrong finding must not match")
	}
}

func TestMatchBareResource(t *testing.T) {

	c := &Config{Ignore: []Ignore{
		{Finding: "f", Object: "pod/coredns-*", Reason: "r"},
	}}
	if _, ok := c.Match("f", "kube-system", "pod/coredns-abc"); !ok {
		t.Error("bare resource glob should match")
	}
}

func TestMatchNamespaceNameWithoutKind(t *testing.T) {
	c := &Config{Ignore: []Ignore{
		{Finding: "f", Object: "default/limitless-*", Reason: "r"},
	}}
	if _, ok := c.Match("f", "default", "pod/limitless-abc"); !ok {
		t.Error("namespace/name glob should match without kind prefix")
	}
}

func TestDiscover(t *testing.T) {
	if got := Discover("/explicit", ""); got != "/explicit" {
		t.Fatalf("flag must win, got %q", got)
	}
	if got := Discover("", ""); got != "" {
		t.Fatalf("empty without file, got %q", got)
	}
}
