package model

import (
	"os"
	"path/filepath"
	"testing"
)

// Committed schema must equal a fresh generation. Fix: `go run ./tools/schemagen`.
func TestSchema_matchesModel(t *testing.T) {
	want := ReportSchema()
	path := filepath.Join("..", "..", "schema", "kubot-inspect-"+SchemaVersion+".json")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed schema: %v (run `go run ./tools/schemagen`)", err)
	}
	if string(got) != string(want) {
		t.Fatalf("committed schema %s drifts from model types — run `go run ./tools/schemagen`", path)
	}
}
