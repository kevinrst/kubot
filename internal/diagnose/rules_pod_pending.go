package diagnose

import (
	"fmt"

	"kubot/internal/k8s"
	"kubot/internal/model"
)

type PodImagePullRule struct{}

func (PodImagePullRule) Name() string { return "pod_imagepull_backoff" }

func (PodImagePullRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pod := range s.Pods {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting == nil {
				continue
			}
			r := cs.State.Waiting.Reason
			if r != "ImagePullBackOff" && r != "ErrImagePull" {
				continue
			}
			image := ""
			for _, c := range pod.Spec.Containers {
				if c.Name == cs.Name {
					image = c.Image
				}
			}
			out = append(out, model.Finding{
				Severity:  "critical",
				Resource:  fmt.Sprintf("pod/%s", pod.Name),
				Namespace: pod.Namespace,
				Reason:    "pod_imagepull_backoff",
				Message:   fmt.Sprintf("Container %q cannot pull image %q (%s)", cs.Name, image, r),
				Evidence: map[string]any{
					"container": cs.Name,
					"image":     image,
					"reason":    r,
					"message":   cs.State.Waiting.Message,
				},
				Recommendation: "Verify the image tag exists and registry credentials (imagePullSecrets) are correct.",
			})
		}
		// Init containers waiting on image pull too.
		for _, cs := range pod.Status.InitContainerStatuses {
			if cs.State.Waiting == nil {
				continue
			}
			r := cs.State.Waiting.Reason
			if r != "ImagePullBackOff" && r != "ErrImagePull" {
				continue
			}
			out = append(out, model.Finding{
				Severity:  "critical",
				Resource:  fmt.Sprintf("pod/%s", pod.Name),
				Namespace: pod.Namespace,
				Reason:    "pod_imagepull_backoff",
				Message:   fmt.Sprintf("Init container %q cannot pull image (%s)", cs.Name, r),
				Evidence: map[string]any{
					"container":     cs.Name,
					"initContainer": true,
					"reason":        r,
					"message":       cs.State.Waiting.Message,
				},
				Recommendation: "Verify the image tag exists and registry credentials (imagePullSecrets) are correct.",
			})
		}
	}
	return out
}

type PodPendingRule struct{}

func (PodPendingRule) Name() string { return "pod_pending" }

func (PodPendingRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pod := range s.Pods {
		if string(pod.Status.Phase) != "Pending" {
			continue
		}
		// Assigned to a node = not a scheduling problem (image pull, volume, …).
		if pod.Spec.NodeName != "" {
			continue
		}
		schedMsg := ""
		schedReason := ""
		for _, e := range s.EventsForPod(pod.Namespace, pod.Name) {
			if e.Reason == "FailedScheduling" {
				schedMsg = e.Message
				schedReason = e.Reason
				break
			}
		}
		var cpuReq, memReq string
		for _, c := range pod.Spec.Containers {
			if q, ok := c.Resources.Requests["cpu"]; ok && cpuReq == "" {
				cpuReq = q.String()
			}
			if q, ok := c.Resources.Requests["memory"]; ok && memReq == "" {
				memReq = q.String()
			}
		}
		sev := "warning"
		msg := fmt.Sprintf("Pod is Pending and unscheduled%s", reasonSuffix(schedReason))
		ev := map[string]any{
			"phase": "Pending",
		}
		if schedReason != "" {
			ev["scheduler_reason"] = schedReason
		}
		if schedMsg != "" {
			ev["scheduler_message"] = truncate(schedMsg, 300)
		}
		if cpuReq != "" {
			ev["cpu_request"] = cpuReq
		}
		if memReq != "" {
			ev["memory_request"] = memReq
		}
		if pod.Spec.NodeSelector != nil {
			ev["nodeSelector"] = pod.Spec.NodeSelector
		}
		out = append(out, model.Finding{
			Severity:  sev,
			Resource:  fmt.Sprintf("pod/%s", pod.Name),
			Namespace: pod.Namespace,
			Reason:    "pod_pending",
			Message:   msg,
			Evidence:  ev,
			Recommendation: "Describe the pod events for FailedScheduling detail; check requests vs node capacity, " +
				"nodeSelector/affinity/tolerations, and unschedulable nodes.",
		})
	}
	return out
}

func reasonSuffix(r string) string {
	if r == "" {
		return ""
	}
	return ": " + r
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
