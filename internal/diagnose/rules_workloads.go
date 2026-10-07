package diagnose

import (
	"fmt"

	"github.com/kevinrst/kubot/internal/k8s"
	"github.com/kevinrst/kubot/internal/model"
)

type StatefulSetUnavailableRule struct{}

func (StatefulSetUnavailableRule) Name() string { return "statefulset_unavailable" }

func (StatefulSetUnavailableRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, st := range s.StatefulSets {
		desired := int32(1)
		if st.Spec.Replicas != nil {
			desired = *st.Spec.Replicas
		}
		if st.Status.ReadyReplicas >= desired {
			continue
		}
		unavail := desired - st.Status.ReadyReplicas
		out = append(out, model.Finding{
			Severity:  "critical",
			Resource:  fmt.Sprintf("statefulset/%s", st.Name),
			Namespace: st.Namespace,
			Reason:    "statefulset_unavailable",
			Message:   fmt.Sprintf("StatefulSet has %d unavailable replica(s) (ready %d/%d)", unavail, st.Status.ReadyReplicas, desired),
			Evidence: map[string]any{
				"desired": desired,
				"ready":   st.Status.ReadyReplicas,
				"current": st.Status.CurrentReplicas,
				"updated": st.Status.UpdatedReplicas,
			},
			Recommendation: "Inspect the failing pods owned by this StatefulSet " +
				"(kubot why " + st.Name + "); stateful pods often block on storage or identity — check PVCs too.",
		})
	}
	return out
}

type DaemonSetUnavailableRule struct{}

func (DaemonSetUnavailableRule) Name() string { return "daemonset_unavailable" }

func (DaemonSetUnavailableRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, ds := range s.DaemonSets {
		if ds.Status.NumberUnavailable == 0 {
			continue
		}
		out = append(out, model.Finding{
			Severity:  "critical",
			Resource:  fmt.Sprintf("daemonset/%s", ds.Name),
			Namespace: ds.Namespace,
			Reason:    "daemonset_unavailable",
			Message:   fmt.Sprintf("DaemonSet has %d unavailable pod(s) (ready %d/%d)", ds.Status.NumberUnavailable, ds.Status.NumberReady, ds.Status.DesiredNumberScheduled),
			Evidence: map[string]any{
				"desired":      ds.Status.DesiredNumberScheduled,
				"ready":        ds.Status.NumberReady,
				"available":    ds.Status.NumberAvailable,
				"unavailable":  ds.Status.NumberUnavailable,
				"misscheduled": ds.Status.NumberMisscheduled,
			},
			Recommendation: "A DaemonSet runs on every node, so one bad pod spec breaks everywhere at once — check node affinity, tolerations, and the pod logs on an affected node.",
		})
	}
	return out
}
