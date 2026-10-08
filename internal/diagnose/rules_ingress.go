package diagnose

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"

	"github.com/kubotdev/kubot/internal/k8s"
	"github.com/kubotdev/kubot/internal/model"
)

type IngressNoBackendsRule struct{}

func (IngressNoBackendsRule) Name() string { return "ingress_no_backends" }

func (IngressNoBackendsRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, ing := range s.Ingresses {
		for _, problem := range backendProblems(s, &ing) {
			out = append(out, model.Finding{
				Severity:  "warning",
				Resource:  fmt.Sprintf("ingress/%s", ing.Name),
				Namespace: ing.Namespace,
				Reason:    "ingress_no_backends",
				Message:   problem.message,
				Evidence:  problem.evidence,
				Recommendation: "Point the backend at a Service that exists and has ready pods; " +
					"check the service name, port number vs port name, and namespace.",
			})
		}
	}
	return out
}

type backendProblem struct {
	message  string
	evidence map[string]any
}

func backendProblems(s *k8s.Snapshot, ing *networkingv1.Ingress) []backendProblem {
	seen := map[string]bool{}
	var out []backendProblem
	check := func(svcName string, port string) {
		if svcName == "" || seen[svcName] {
			return
		}
		seen[svcName] = true
		if !serviceExists(s, ing.Namespace, svcName) {
			out = append(out, backendProblem{
				message: fmt.Sprintf("Backend service %q does not exist", svcName),
				evidence: map[string]any{
					"service": svcName,
					"port":    port,
					"missing": true,
				},
			})
			return
		}
		if ready, _ := s.ReadyEndpoints(ing.Namespace, svcName); ready == 0 {
			out = append(out, backendProblem{
				message: fmt.Sprintf("Backend service %q has no ready endpoints", svcName),
				evidence: map[string]any{
					"service":         svcName,
					"port":            port,
					"ready_endpoints": ready,
				},
			})
		}
	}
	if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
		b := ing.Spec.DefaultBackend.Service
		check(b.Name, portString(b.Port))
	}
	for _, r := range ing.Spec.Rules {
		if r.HTTP == nil {
			continue
		}
		for _, p := range r.HTTP.Paths {
			if p.Backend.Service != nil {
				check(p.Backend.Service.Name, portString(p.Backend.Service.Port))
			}
		}
	}
	return out
}

func portString(p networkingv1.ServiceBackendPort) string {
	if p.Name != "" {
		return p.Name
	}
	return fmt.Sprintf("%d", p.Number)
}

func serviceExists(s *k8s.Snapshot, ns, name string) bool {
	for _, svc := range s.Services {
		if svc.Namespace == ns && svc.Name == name {
			return true
		}
	}
	return false
}

type IngressClassRule struct{}

func (IngressClassRule) Name() string { return "ingress_unknown_class" }

func (IngressClassRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, ing := range s.Ingresses {
		if ing.Spec.IngressClassName == nil {
			if !hasDefaultClass(s) {
				out = append(out, model.Finding{
					Severity:       "warning",
					Resource:       fmt.Sprintf("ingress/%s", ing.Name),
					Namespace:      ing.Namespace,
					Reason:         "ingress_unknown_class",
					Message:        "Ingress sets no class and the cluster has no default",
					Evidence:       map[string]any{"ingress_class": ""},
					Recommendation: "Set spec.ingressClassName to an installed controller's class, or define a default IngressClass.",
				})
			}
			continue
		}
		class := *ing.Spec.IngressClassName
		if !classExists(s, class) {
			out = append(out, model.Finding{
				Severity:       "warning",
				Resource:       fmt.Sprintf("ingress/%s", ing.Name),
				Namespace:      ing.Namespace,
				Reason:         "ingress_unknown_class",
				Message:        fmt.Sprintf("Ingress class %q matches no installed IngressClass", class),
				Evidence:       map[string]any{"ingress_class": class},
				Recommendation: "Install the controller for this class or fix the class name; check `kubectl get ingressclass`.",
			})
		}
	}
	return out
}

func classExists(s *k8s.Snapshot, name string) bool {
	for _, c := range s.IngressClasses {
		if c.Name == name {
			return true
		}
	}
	return false
}

func hasDefaultClass(s *k8s.Snapshot) bool {
	for _, c := range s.IngressClasses {
		if c.Annotations["ingressclass.kubernetes.io/is-default-class"] == "true" {
			return true
		}
	}
	return false
}

type IngressTLSSecretRule struct{}

func (IngressTLSSecretRule) Name() string { return "ingress_tls_secret_missing" }

func (IngressTLSSecretRule) Run(s *k8s.Snapshot) []model.Finding {
	present := map[string]bool{}
	for _, sec := range s.Secrets {
		present[sec.Namespace+"/"+sec.Name] = true
	}
	var out []model.Finding
	for _, ing := range s.Ingresses {
		for _, tls := range ing.Spec.TLS {
			if tls.SecretName == "" || present[ing.Namespace+"/"+tls.SecretName] {
				continue
			}
			hosts := make([]string, 0, len(tls.Hosts))
			hosts = append(hosts, tls.Hosts...)
			out = append(out, model.Finding{
				Severity:  "warning",
				Resource:  fmt.Sprintf("ingress/%s", ing.Name),
				Namespace: ing.Namespace,
				Reason:    "ingress_tls_secret_missing",
				Message:   fmt.Sprintf("TLS secret %q does not exist — TLS terminates nowhere", tls.SecretName),
				Evidence: map[string]any{
					"secret": tls.SecretName,
					"hosts":  hosts,
				},
				Recommendation: "Create the TLS secret in the ingress namespace (kubectl create secret tls) or fix the name.",
			})
		}
	}
	return out
}
