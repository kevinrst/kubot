package diagnose

import (
	"os"
	"path/filepath"
	"testing"
)

// Every registered rule needs a catalogue page. A finding id with no docs
// is a dead end for the user holding it.
func TestCatalogueCoversRules(t *testing.T) {
	for _, r := range DefaultRules() {
		p := filepath.Join("..", "..", "docs", "findings", r.Name()+".md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("rule %q has no catalogue page (expected %s)", r.Name(), p)
		}
	}
}
