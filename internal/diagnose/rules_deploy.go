package diagnose

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"kubot/internal/k8s"
	"kubot/internal/model"
)

type DeploymentUnavailableRule struct{}

func (DeploymentUnavailableRule) Name() string { return "deployment_unavailable" }

func (DeploymentUnavailableRule) Run(s *k8s.Snapshot) []model.Finding {
	// Count unhealthy pods per top owner, for evidence.
	failingPods := map[string]int{}
	for _, pod := range s.Pods {
		if podUnhealthy(&pod) {
			failingPods[topOwnerOfPod(s, pod.Namespace, pod.Name)]++
		}
	}
	var out []model.Finding
	for _, d := range s.Deployments {
		unavail := int32(0)
		if d.Status.UnavailableReplicas > 0 {
			unavail = d.Status.UnavailableReplicas
		}
		desired := int32(0)
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		available := d.Status.AvailableReplicas
		condMsg := ""
		for _, c := range d.Status.Conditions {
			if c.Type == "Available" && string(c.Status) != "True" {
				condMsg = c.Message
			}
		}
		if unavail == 0 && condMsg == "" {
			continue
		}
		ev := map[string]any{
			"desired":     desired,
			"ready":       d.Status.ReadyReplicas,
			"available":   available,
			"unavailable": unavail,
			"updated":     d.Status.UpdatedReplicas,
		}
		if condMsg != "" {
			ev["condition"] = truncate(condMsg, 300)
		}
		if n := failingPods[d.Name]; n > 0 {
			ev["failing_pods"] = n
		}
		out = append(out, model.Finding{
			Severity:  "critical",
			Resource:  fmt.Sprintf("deployment/%s", d.Name),
			Namespace: d.Namespace,
			Reason:    "deployment_unavailable",
			Message:   fmt.Sprintf("Deployment has %d unavailable replica(s) (ready %d/%d)", unavail, d.Status.ReadyReplicas, desired),
			Evidence:  ev,
			Recommendation: "Inspect the failing pods owned by this deployment " +
				"(kubot why " + d.Name + "); check rollout status and pod events.",
		})
	}
	return out
}

func podUnhealthy(p *corev1.Pod) bool {
	if string(p.Status.Phase) == "Pending" || string(p.Status.Phase) == "Failed" {
		return true
	}
	for _, cs := range p.Status.ContainerStatuses {
		if cs.State.Waiting != nil || (cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0) {
			return true
		}
		if cs.RestartCount > 5 {
			return true
		}
	}
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status != corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func topOwnerOfPod(s *k8s.Snapshot, ns, podName string) string {
	m := s.PodsByTopOwner()
	for owner, pods := range m {
		for _, p := range pods {
			if p == podName {
				return owner
			}
		}
	}
	return podName
}
