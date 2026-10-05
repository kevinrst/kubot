package diagnose

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/kevinrst/kubot/internal/k8s"
	"github.com/kevinrst/kubot/internal/model"
)

type NodePressureRule struct{}

func (NodePressureRule) Name() string { return "node_pressure" }

func (NodePressureRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, n := range s.Nodes {
		var pressures []string
		for _, c := range n.Status.Conditions {
			switch c.Type {
			case corev1.NodeMemoryPressure, corev1.NodeDiskPressure, corev1.NodePIDPressure:
				if c.Status == corev1.ConditionTrue {
					pressures = append(pressures, string(c.Type))
				}
			}
		}
		if len(pressures) == 0 {
			continue
		}
		out = append(out, model.Finding{
			Severity:  "warning",
			Resource:  fmt.Sprintf("node/%s", n.Name),
			Namespace: "",
			Reason:    "node_pressure",
			Message:   fmt.Sprintf("Node under pressure: %s", joinList(pressures)),
			Evidence: map[string]any{
				"pressures": pressures,
			},
			Recommendation: "Pods on this node risk eviction and OOMKills; drain or add capacity, and check what's consuming the pressured resource.",
		})
	}
	return out
}
