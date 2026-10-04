package render

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"kubot/internal/model"
)

// Options controls rendering.
type Options struct {
	NoColor bool
	JSON    bool
}

func UseColor(noColorFlag bool) bool {
	if noColorFlag {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func PrintReport(w io.Writer, rep model.Report, opt Options) error {
	if opt.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}
	return PrintText(w, rep, UseColor(opt.NoColor))
}

func PrintText(w io.Writer, rep model.Report, color bool) error {
	bold, red, yellow, reset := "", "", "", ""
	if color {
		bold, red, yellow, reset = "\033[1m", "\033[31m", "\033[33m", "\033[0m"
	}
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "%sKubernetes Health Report%s\n", bold, reset)
	if rep.Cluster.Context != "" {
		fmt.Fprintf(w, "context %s", rep.Cluster.Context)
		if rep.Cluster.Namespace != "" {
			fmt.Fprintf(w, " · namespace %s", rep.Cluster.Namespace)
		} else {
			fmt.Fprintf(w, " · all namespaces")
		}
		fmt.Fprintln(w, "")
	}
	fmt.Fprintln(w, "")
	if len(rep.Issues) == 0 {
		fmt.Fprintf(w, "  %sOK%s — no problems detected.\n", bold, reset)
	} else {
		groups := map[string][]model.Finding{}
		for _, f := range rep.Issues {
			groups[strings.ToUpper(f.Severity)] = append(groups[strings.ToUpper(f.Severity)], f)
		}
		for _, sev := range []string{"CRITICAL", "WARNING", "NOTE"} {
			fs := groups[sev]
			if len(fs) == 0 {
				continue
			}
			h := sev
			if color {
				if sev == "CRITICAL" {
					h = red + sev + reset
				} else if sev == "WARNING" {
					h = yellow + sev + reset
				}
			}
			fmt.Fprintln(w, h)
			for _, f := range fs {
				printFinding(w, f, color)
			}
			fmt.Fprintln(w, "")
		}
	}
	if len(rep.Checked) > 0 {
		cp := append([]string{}, rep.Checked...)
		sort.Strings(cp)
		fmt.Fprintf(w, "checked · %s\n", strings.Join(cp, " · "))
	}
	fmt.Fprintln(w, "")
	if len(rep.Issues) > 0 {
		fmt.Fprintln(w, "Details: kubot inspect <workload>  ·  Machine-readable: kubot inspect --json")
	}
	return nil
}

func printFinding(w io.Writer, f model.Finding, color bool) {
	bold, reset := "", ""
	if color {
		bold, reset = "\033[1m", "\033[0m"
	}
	res := f.Resource
	if f.Namespace != "" {
		res = fmt.Sprintf("%s (%s)", f.Resource, f.Namespace)
	}
	fmt.Fprintf(w, "  %s%s%s\n", bold, res, reset)
	fmt.Fprintf(w, "    %s\n", f.Message)
	if len(f.Evidence) > 0 {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "    Evidence:")
		keys := make([]string, 0, len(f.Evidence))
		for k := range f.Evidence {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "      %s: %v\n", k, f.Evidence[k])
		}
	}
	if f.Recommendation != "" {
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "    Recommendation:\n      %s\n", f.Recommendation)
	}
	fmt.Fprintln(w, "")
}
