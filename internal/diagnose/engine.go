package diagnose

import (
	"sort"

	networkingv1 "k8s.io/api/networking/v1"

	"github.com/kevinrst/kubot/internal/k8s"
	"github.com/kevinrst/kubot/internal/model"
)

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

func (e *Engine) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, r := range e.rules {
		out = append(out, r.Run(s)...)
	}
	sort.Slice(out, func(i, j int) bool {
		oi, oj := model.RankSeverity(out[i].Severity), model.RankSeverity(out[j].Severity)
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

// Narrows findings to one workload: bare/type name, owned pods, and services linked via selectors.
// Empty workload returns everything (namespace-filtered).
func FilterByWorkload(findings []model.Finding, workload, namespace string, snap *k8s.Snapshot) []model.Finding {
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
	inScope := func(ns string) bool { return namespace == "" || ns == namespace }

	ownerOf := map[podID]string{}
	for _, p := range snap.Pods {
		if inScope(p.Namespace) {
			ownerOf[podID{p.Namespace, p.Name}] = k8s.TopOwnerName(&p)
		}
	}
	related := map[string]bool{name: true}
	nameHit := func(obj string) bool { return obj == name || hasPrefixMatch(obj, name) }
	for _, p := range snap.Pods {
		if inScope(p.Namespace) && nameHit(p.Name) {
			related[p.Name] = true
		}
	}
	for _, svc := range snap.Services {
		if inScope(svc.Namespace) && nameHit(svc.Name) {
			related[svc.Name] = true
		}
	}
	for _, d := range snap.Deployments {
		if inScope(d.Namespace) && nameHit(d.Name) {
			related[d.Name] = true
		}
	}
	for _, st := range snap.StatefulSets {
		if inScope(st.Namespace) && nameHit(st.Name) {
			related[st.Name] = true
		}
	}
	for _, ds := range snap.DaemonSets {
		if inScope(ds.Namespace) && nameHit(ds.Name) {
			related[ds.Name] = true
		}
	}
	for _, ing := range snap.Ingresses {
		if inScope(ing.Namespace) && nameHit(ing.Name) {
			related[ing.Name] = true
		}
	}
	for _, j := range snap.Jobs {
		if inScope(j.Namespace) && nameHit(j.Name) {
			related[j.Name] = true
		}
	}
	for _, cj := range snap.CronJobs {
		if inScope(cj.Namespace) && nameHit(cj.Name) {
			related[cj.Name] = true
		}
	}
	for changed := true; changed; {
		changed = false
		add := func(n string) {
			if n != "" && !related[n] {
				related[n] = true
				changed = true
			}
		}
		for id, owner := range ownerOf {
			if related[owner] {
				add(id.name)
			}
			if related[id.name] {
				add(owner)
			}
		}
		for _, j := range snap.Jobs {
			if !inScope(j.Namespace) {
				continue
			}
			for _, o := range j.OwnerReferences {
				if o.Kind != "CronJob" {
					continue
				}
				if related[o.Name] {
					add(j.Name)
				}
				if related[j.Name] {
					add(o.Name)
				}
			}
		}
		for _, svc := range snap.Services {
			if !inScope(svc.Namespace) {
				continue
			}
			matched := podsMatchingSelector(snap, svc.Namespace, svc.Spec.Selector)
			if related[svc.Name] {
				for _, p := range matched {
					add(p)
					add(ownerOf[podID{svc.Namespace, p}])
				}
			} else {
				for _, p := range matched {
					if related[p] {
						add(svc.Name)
						break
					}
				}
			}
		}
		for _, ing := range snap.Ingresses {
			if !inScope(ing.Namespace) {
				continue
			}
			backends := ingressBackends(&ing)
			if related[ing.Name] {
				for _, b := range backends {
					add(b)
				}
			} else {
				for _, b := range backends {
					if related[b] {
						add(ing.Name)
						break
					}
				}
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

type podID struct {
	ns   string
	name string
}

// lists backend service names referenced by an ingress.
func ingressBackends(ing *networkingv1.Ingress) []string {
	var out []string
	seen := map[string]bool{}
	add := func(n string) {
		if n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
		add(ing.Spec.DefaultBackend.Service.Name)
	}
	for _, r := range ing.Spec.Rules {
		if r.HTTP == nil {
			continue
		}
		for _, p := range r.HTTP.Paths {
			if p.Backend.Service != nil {
				add(p.Backend.Service.Name)
			}
		}
	}
	return out
}

func podsMatchingSelector(snap *k8s.Snapshot, ns string, sel map[string]string) []string {
	if len(sel) == 0 {
		return nil
	}
	var out []string
	for _, p := range snap.Pods {
		if p.Namespace != ns {
			continue
		}
		ok := true
		for k, v := range sel {
			if p.Labels[k] != v {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, p.Name)
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
	if len(candidate) == len(prefix) {
		return true
	}
	c := candidate[len(prefix)]
	return c == '-' || c == '/' || c == '.'
}
