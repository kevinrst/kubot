package diagnose

import (
	"sort"

	"github.com/kevinrst/kubot/internal/k8s"
	"github.com/kevinrst/kubot/internal/model"
)

// A single diagnostic check. Rules are pure: same snapshot in, same findings out.
type Rule interface {
	Name() string
	Run(s *k8s.Snapshot) []model.Finding
}

type Engine struct {
	rules []Rule
}

func NewEngine() *Engine {
	return &Engine{rules: DefaultRules()}
}

// Runs all rules. Output order is fixed (severity, resource, reason) so the
// same snapshot always yields the same report.
func (e *Engine) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, r := range e.rules {
		out = append(out, r.Run(s)...)
	}
	sort.Slice(out, func(i, j int) bool {
		oi, oj := model.SeverityOrder[out[i].Severity], model.SeverityOrder[out[j].Severity]
		if oi != oj {
			return oi < oj
		}
		if out[i].Resource != out[j].Resource {
			return out[i].Resource < out[j].Resource
		}
		if out[i].Namespace != out[j].Namespace {
			return out[i].Namespace < out[j].Namespace
		}
		return out[i].Reason < out[j].Reason
	})
	return out
}

// Narrows findings to one workload: bare name, "type/name", owned pods, and
// matching services. Empty workload returns everything (namespace-filtered).
func FilterByWorkload(findings []model.Finding, workload, namespace string, podsByOwner map[string][]string) []model.Finding {
	if workload == "" {
		if namespace == "" {
			return findings
		}
		var out []model.Finding
		for _, f := range findings {
			if f.Namespace == namespace || f.Namespace == "" {
				out = append(out, f)
			}
		}
		return out
	}
	name := workload
	if i := indexSlash(workload); i >= 0 {
		name = workload[i+1:] // "type/name" -> name
	}
	related := map[string]bool{name: true}
	for owner, pods := range podsByOwner {
		if owner == name {
			for _, p := range pods {
				related[p] = true
			}
		}
	}
	// Pods named "<workload>-<replicaset-hash>-<id>" belong to it too.
	var out []model.Finding
	for _, f := range findings {
		if namespace != "" && f.Namespace != "" && f.Namespace != namespace {
			continue
		}
		resName := resourceName(f.Resource)
		if related[resName] || hasPrefixMatch(resName, name) {
			out = append(out, f)
		}
	}
	return out
}

func indexSlash(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}

func resourceName(resource string) string {
	for i := len(resource) - 1; i >= 0; i-- {
		if resource[i] == '/' {
			return resource[i+1:]
		}
	}
	return resource
}

func hasPrefixMatch(candidate, prefix string) bool {
	if len(candidate) < len(prefix) {
		return false
	}
	if candidate[:len(prefix)] != prefix {
		return false
	}
	// Boundary-aware: "payments-api-xxx" matches "payments-api".
	if len(candidate) == len(prefix) {
		return true
	}
	c := candidate[len(prefix)]
	return c == '-' || c == '/' || c == '.'
}
