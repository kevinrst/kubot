package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kevinrst/kubot/internal/model"
)

func TestOpenAIRequestShape(t *testing.T) {
	var gotAuth, gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotModel, _ = body["model"].(string)
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"it OOMed"}}]}`))
	}))
	defer srv.Close()

	p := OpenAIProvider{Key: "k", ModelID: "m", BaseURL: srv.URL}
	text, err := p.Complete(context.Background(), "sys", "user")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if text != "it OOMed" {
		t.Fatalf("got %q", text)
	}
	if gotAuth != "Bearer k" || gotModel != "m" {
		t.Fatalf("auth=%q model=%q", gotAuth, gotModel)
	}
}

func TestOpenAIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer srv.Close()

	p := OpenAIProvider{Key: "k", ModelID: "m", BaseURL: srv.URL}
	if _, err := p.Complete(context.Background(), "s", "u"); err == nil {
		t.Fatal("expected error")
	}
}

func TestAnthropicShape(t *testing.T) {
	var gotKey, gotVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		if r.URL.Path != "/v1/messages" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"content":[{"type":"text","text":"raise the limit"}]}`))
	}))
	defer srv.Close()

	p := AnthropicProvider{Key: "k", ModelID: "m", BaseURL: srv.URL}
	text, err := p.Complete(context.Background(), "sys", "user")
	if err != nil || text != "raise the limit" {
		t.Fatalf("got %q, %v", text, err)
	}
	if gotKey != "k" || gotVersion == "" {
		t.Fatalf("key=%q version=%q", gotKey, gotVersion)
	}
}

func TestGeminiShape(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"check memory"}]}}]}`))
	}))
	defer srv.Close()

	p := GeminiProvider{Key: "k", ModelID: "m", BaseURL: srv.URL}
	text, err := p.Complete(context.Background(), "sys", "user")
	if err != nil || text != "check memory" {
		t.Fatalf("got %q, %v", text, err)
	}
	if !strings.Contains(gotPath, "models/m:generateContent") || !strings.Contains(gotQuery, "key=k") {
		t.Fatalf("path=%q query=%q", gotPath, gotQuery)
	}
}

func TestResolve(t *testing.T) {
	t.Setenv("KUBOT_AI_PROVIDER", "")
	t.Setenv("KUBOT_AI_API_KEY", "")
	t.Setenv("KUBOT_AI_MODEL", "")
	t.Setenv("KUBOT_AI_BASE_URL", "")
	t.Setenv("OPENAI_API_KEY", "ok")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	p, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p.Name() != "openai" || p.Model() != "gpt-4o-mini" {
		t.Fatalf("got %s/%s", p.Name(), p.Model())
	}
}

func TestResolveUnknownProvider(t *testing.T) {
	t.Setenv("KUBOT_AI_PROVIDER", "nope")
	if _, err := Resolve(); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveModelOverride(t *testing.T) {
	t.Setenv("KUBOT_AI_PROVIDER", "")
	t.Setenv("KUBOT_AI_API_KEY", "")
	t.Setenv("KUBOT_AI_MODEL", "gemini-2.0-flash")
	t.Setenv("KUBOT_AI_BASE_URL", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "k")
	p, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p.Name() != "gemini" || p.Model() != "gemini-2.0-flash" {
		t.Fatalf("got %s/%s", p.Name(), p.Model())
	}
}

func TestUserPromptCarriesFindings(t *testing.T) {
	rep := model.Report{SchemaVersion: "0.1.0", Status: "critical",
		Issues: []model.Finding{{Severity: "critical", Resource: "pod/x", Reason: "pod_oom_killed", Message: "OOM"}}}
	u := UserPrompt("why?", rep)
	if !strings.Contains(u, "why?") || !strings.Contains(u, "pod_oom_killed") {
		t.Fatalf("prompt missing question or findings: %q", u)
	}
	if !strings.Contains(u, "relevant") {
		t.Fatalf("prompt must demand focus on relevant findings: %q", u)
	}
	if !strings.Contains(SystemPrompt(), "never add new ones") {
		t.Fatal("system prompt must forbid invention")
	}
	for _, want := range []string{"NO markdown", "at most three", "OMIT", "worst first"} {
		if !strings.Contains(SystemPrompt(), want) {
			t.Errorf("system prompt missing shape rule %q", want)
		}
	}
}
