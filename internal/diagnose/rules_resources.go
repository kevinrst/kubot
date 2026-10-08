package diagnose

import (
	"fmt"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kubotdev/kubot/internal/k8s"
	"github.com/kubotdev/kubot/internal/model"
)

// Below this, a memory limit is more likely a typo than a budget.
var suspiciousMemoryFloor = resource.MustParse("32Mi")

// Usage at/above this share of the memory limit means OOM risk.
const NearLimitRatio = 0.8

type PodMissingResourcesRule struct{}

func (PodMissingResourcesRule) Name() string { return "pod_missing_resources" }

func (PodMissingResourcesRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pod := range s.Pods {
		if isMirrorPod(&pod) || pod.Namespace == "kube-system" {
			continue
		}
		var bare []string
		for _, c := range pod.Spec.Containers {
			missing := []string{}
			if _, ok := c.Resources.Limits[corev1.ResourceMemory]; !ok {
				missing = append(missing, "memory limit")
			}
			if _, ok := c.Resources.Requests[corev1.ResourceCPU]; !ok {
				missing = append(missing, "cpu request")
			}
			if _, ok := c.Resources.Requests[corev1.ResourceMemory]; !ok {
				missing = append(missing, "memory request")
			}
			if len(missing) > 0 {
				bare = append(bare, c.Name+" ("+joinList(missing)+")")
			}
		}
		if len(bare) == 0 {
			continue
		}
		out = append(out, model.Finding{
			Severity:  "note",
			Resource:  fmt.Sprintf("pod/%s", pod.Name),
			Namespace: pod.Namespace,
			Reason:    "pod_missing_resources",
			Message:   fmt.Sprintf("Container(s) missing settings: %s", joinList(bare)),
			Evidence: map[string]any{
				"containers": bare,
			},
			Recommendation: "Set memory limits (no limit = OOMKill by node pressure) and cpu/memory requests (scheduling and QoS depend on them).",
		})
	}
	return out
}

type PodLowLimitRule struct{}

func (PodLowLimitRule) Name() string { return "pod_low_limit" }

func (PodLowLimitRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pod := range s.Pods {
		for _, c := range pod.Spec.Containers {
			q, ok := c.Resources.Limits[corev1.ResourceMemory]
			if !ok || q.Cmp(suspiciousMemoryFloor) >= 0 {
				continue
			}
			out = append(out, model.Finding{
				Severity:  "warning",
				Resource:  fmt.Sprintf("pod/%s", pod.Name),
				Namespace: pod.Namespace,
				Reason:    "pod_low_limit",
				Message:   fmt.Sprintf("Container %q memory limit %s is suspiciously low", c.Name, q.String()),
				Evidence: map[string]any{
					"container":    c.Name,
					"memory_limit": q.String(),
					"floor":        suspiciousMemoryFloor.String(),
				},
				Recommendation: "Verify the limit is intentional; most runtimes exceed this under any real load. Raise it or remove it and observe.",
			})
		}
	}
	return out
}

type PodNearLimitRule struct{}

func (PodNearLimitRule) Name() string { return "pod_near_limit" }

func (PodNearLimitRule) Run(s *k8s.Snapshot) []model.Finding {
	usage := map[string]int64{}
	for _, u := range s.Usage {
		usage[u.Namespace+"/"+u.Pod+"/"+u.Container] = u.MemoryBytes
	}
	var out []model.Finding
	for _, pod := range s.Pods {
		for _, c := range pod.Spec.Containers {
			limit, ok := c.Resources.Limits[corev1.ResourceMemory]
			if !ok || limit.Value() <= 0 {
				continue
			}
			used, ok := usage[pod.Namespace+"/"+pod.Name+"/"+c.Name]
			if !ok {
				continue
			}
			ratio := float64(used) / float64(limit.Value())
			if ratio < NearLimitRatio {
				continue
			}
			out = append(out, model.Finding{
				Severity:  "warning",
				Resource:  fmt.Sprintf("pod/%s", pod.Name),
				Namespace: pod.Namespace,
				Reason:    "pod_near_limit",
				Message:   fmt.Sprintf("Container %q uses %.0f%% of its memory limit", c.Name, ratio*100),
				Evidence: map[string]any{
					"container":    c.Name,
					"memory_limit": limit.String(),
					"memory_usage": joinBytes(used),
					"usage_ratio":  round2(ratio),
				},
				Recommendation: "Raise the limit or cut usage before the next spike OOMKills it.",
			})
		}
	}
	return out
}

type DeploymentResourceRiskRule struct{}

func (DeploymentResourceRiskRule) Name() string { return "deployment_resource_risk" }

func (DeploymentResourceRiskRule) Run(s *k8s.Snapshot) []model.Finding {
	byOwner := s.PodsByTopOwner()
	var out []model.Finding
	for _, d := range s.Deployments {
		if d.Namespace == "kube-system" {
			continue
		}
		pods := byOwner[d.Name]
		if len(pods) == 0 {
			continue
		}
		bare := 0
		for _, pod := range s.Pods {
			if pod.Namespace != d.Namespace || !inList(pods, pod.Name) {
				continue
			}
			for _, c := range pod.Spec.Containers {
				if _, ok := c.Resources.Limits[corev1.ResourceMemory]; !ok {
					bare++
					break
				}
			}
		}
		if bare*2 < len(pods) {
			continue
		}
		out = append(out, model.Finding{
			Severity:  "note",
			Resource:  fmt.Sprintf("deployment/%s", d.Name),
			Namespace: d.Namespace,
			Reason:    "deployment_resource_risk",
			Message:   fmt.Sprintf("%d/%d pod(s) run without a memory limit", bare, len(pods)),
			Evidence: map[string]any{
				"pods_total":         len(pods),
				"pods_without_limit": bare,
			},
			Recommendation: "Set memory limits so one hungry pod can't take the node (and its neighbors) down with it.",
		})
	}
	return out
}

func isMirrorPod(pod *corev1.Pod) bool {
	for _, o := range pod.OwnerReferences {
		if o.Kind == "Node" {
			return true
		}
	}
	return false
}

func joinList(ss []string) string {
	return strings.Join(ss, ", ")
}

func inList(ss []string, v string) bool {
	return slices.Contains(ss, v)
}

func joinBytes(b int64) string {
	const unit = 1024 * 1024
	if b < unit {
		return fmt.Sprintf("%dKi", b/1024)
	}
	return fmt.Sprintf("%dMi", b/unit)
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
