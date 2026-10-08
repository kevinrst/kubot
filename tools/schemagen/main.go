package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubotdev/kubot/internal/model"
)

func main() {
	dir := "schema"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	name := fmt.Sprintf("kubot-inspect-%s.json", model.SchemaVersion)
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, model.ReportSchema(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote", p)
}
