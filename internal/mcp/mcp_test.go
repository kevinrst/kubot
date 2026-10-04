package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

var errTest = errors.New("boom")

func serveOne(t *testing.T, srv *Server, lines ...string) []map[string]any {
	t.Helper()
	in := strings.NewReader(strings.Join(lines, "\n") + "\n")
	var out strings.Builder
	if err := srv.Serve(context.Background(), in, &out); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	var resps []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("decode response %q: %v", line, err)
		}
		resps = append(resps, m)
	}
	return resps
}

func TestServe_initializeAndToolsList(t *testing.T) {
	srv := &Server{Name: "kubot", Version: "dev", Tools: []Tool{
		{Name: "inspect", Description: "d", Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			return `{}`, nil
		}},
	}}
	resps := serveOne(t, srv,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
	)
	if len(resps) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(resps))
	}
	initRes := resps[0]["result"].(map[string]any)
	if initRes["protocolVersion"] != "2025-06-18" {
		t.Fatalf("protocol not negotiated: %+v", initRes)
	}
	tools := resps[1]["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %+v", tools)
	}
}

func TestServe_unknownToolIsToolError(t *testing.T) {
	srv := &Server{Name: "kubot", Tools: nil}
	resps := serveOne(t, srv,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"nope"}}`,
	)
	if _, ok := resps[0]["error"]; !ok {
		t.Fatalf("expected JSON-RPC error for unknown tool, got %+v", resps[0])
	}
}

func TestServe_toolHandlerErrorIsResult(t *testing.T) {
	srv := &Server{Name: "kubot", Tools: []Tool{
		{Name: "inspect", Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
			return "", errTest
		}},
	}}
	resps := serveOne(t, srv,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"inspect","arguments":{}}}`,
	)
	res := resps[0]["result"].(map[string]any)
	if res["isError"] != true {
		t.Fatalf("handler error must surface as tool result with isError, got %+v", res)
	}
}
