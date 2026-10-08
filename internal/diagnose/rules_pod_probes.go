package diagnose

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/kubotdev/kubot/internal/k8s"
	"github.com/kubotdev/kubot/internal/model"
)

// Fewer aggregated failures than this is startup noise, not a broken probe.
const minEventRepeats = 3

type PodProbeFailingRule struct{}

func (PodProbeFailingRule) Name() string { return "pod_probe_failing" }

func (PodProbeFailingRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pod := range s.Pods {
		if podReady(&pod) {
			continue
		}
		probeEvidence := probeFailureFromEvents(s, pod.Namespace, pod.Name)
		if probeEvidence == nil {
			continue
		}
		out = append(out, model.Finding{
			Severity:  "warning",
			Resource:  fmt.Sprintf("pod/%s", pod.Name),
			Namespace: pod.Namespace,
			Reason:    "pod_probe_failing",
			Message:   fmt.Sprintf("Pod has failing %s probe", probeEvidence["probe"]),
			Evidence:  probeEvidence,
			Recommendation: "Check the probe endpoint, port/path, initialDelaySeconds and failureThreshold; " +
				"verify the app is listening and logs show successful startup.",
		})
	}
	return out
}

func podReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func probeFailureFromEvents(s *k8s.Snapshot, namespace, podName string) map[string]any {
	var probe string
	var msg string
	var count int
	for _, e := range s.Events {
		if e.InvolvedObject.Kind != "Pod" || e.InvolvedObject.Name != podName || e.InvolvedObject.Namespace != namespace {
			continue
		}
		if e.Type != "Warning" {
			continue
		}
		m := strings.ToLower(e.Message)
		p := ""
		switch {
		case strings.Contains(m, "liveness"):
			p = "liveness"
		case strings.Contains(m, "readiness"):
			p = "readiness"
		case strings.Contains(m, "startup"):
			p = "startup"
		default:
			continue
		}
		if !strings.Contains(m, "fail") && !strings.Contains(m, "unhealthy") {
			continue
		}
		count += int(e.Count)
		if probe == "" {
			probe = p
			msg = e.Message
		}
	}
	if probe == "" {
		return nil
	}
	if count == 0 {
		count = 1
	}
	if count < minEventRepeats {
		return nil
	}
	return map[string]any{
		"probe":         probe,
		"failure_count": count,
		"event_message": truncate(msg, 300),
	}
}
