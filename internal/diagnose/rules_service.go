package diagnose

import (
	"fmt"

	"github.com/kevinrst/kubot/internal/k8s"
	"github.com/kevinrst/kubot/internal/model"
)

type ServiceNoEndpointsRule struct{}

func (ServiceNoEndpointsRule) Name() string { return "service_no_endpoints" }

func (ServiceNoEndpointsRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, svc := range s.Services {
		if svc.Spec.ClusterIP == "None" {
			continue
		}
		if svc.Spec.Type == "ExternalName" {
			continue
		}
		if len(svc.Spec.Selector) == 0 {
			continue
		}
		ready, total := s.ReadyEndpoints(svc.Namespace, svc.Name)
		if ready > 0 {
			continue
		}
		matched := countSelectorMatches(s, svc.Namespace, svc.Spec.Selector)
		ev := map[string]any{
			"selector":        svc.Spec.Selector,
			"matched_pods":    matched,
			"ready_endpoints": ready,
			"total_endpoints": total,
		}
		msg := "Service has no ready endpoints"
		if matched == 0 {
			msg = "Service selector matches no pods"
		} else if total == 0 {
			msg = "Service has no endpoints despite matching pods (backends not ready?)"
		}
		out = append(out, model.Finding{
			Severity:  "warning",
			Resource:  fmt.Sprintf("service/%s", svc.Name),
			Namespace: svc.Namespace,
			Reason:    "service_no_endpoints",
			Message:   msg,
			Evidence:  ev,
			Recommendation: "Compare the service selector against pod labels " +
				"(kubectl get pods --show-labels); check that backing pods are Ready and ports match.",
		})
	}
	return out
}

func countSelectorMatches(s *k8s.Snapshot, ns string, sel map[string]string) int {
	n := 0
	for _, p := range s.Pods {
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
			n++
		}
	}
	return n
}
