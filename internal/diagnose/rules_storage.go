package diagnose

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/kubotdev/kubot/internal/k8s"
	"github.com/kubotdev/kubot/internal/model"
)

type PVCPendingRule struct{}

func (PVCPendingRule) Name() string { return "pvc_pending" }

func (PVCPendingRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pvc := range s.PVCs {
		if pvc.Status.Phase != corev1.ClaimPending {
			continue
		}
		class := "default"
		if pvc.Spec.StorageClassName != nil {
			class = *pvc.Spec.StorageClassName
		}
		size := ""
		if q, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			size = q.String()
		}
		out = append(out, model.Finding{
			Severity:  "warning",
			Resource:  fmt.Sprintf("persistentvolumeclaim/%s", pvc.Name),
			Namespace: pvc.Namespace,
			Reason:    "pvc_pending",
			Message:   "PVC is Pending and unbound — pods using it can't start",
			Evidence: map[string]any{
				"storage_class": class,
				"requested":     size,
				"access_modes":  accessModeNames(pvc.Spec.AccessModes),
			},
			Recommendation: "Check the StorageClass exists and has a provisioner for this cluster; check PVs and quota.",
		})
	}
	return out
}

type PodMountFailureRule struct{}

func (PodMountFailureRule) Name() string { return "pod_mount_failure" }

func (PodMountFailureRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pod := range s.Pods {
		msg, count := mountFailure(s, pod.Namespace, pod.Name)
		if count == 0 {
			continue
		}
		// Fresh event objects restart the counter from 1 after the 1h expiry,
		// so a low count alone can't prove transience: an old pod failing to
		// mount for longer than the backoff ceiling is sustained by definition.
		old := time.Since(pod.CreationTimestamp.Time) > 15*time.Minute
		if count < minEventRepeats && !old {
			continue
		}
		out = append(out, model.Finding{
			Severity:  "warning",
			Resource:  fmt.Sprintf("pod/%s", pod.Name),
			Namespace: pod.Namespace,
			Reason:    "pod_mount_failure",
			Message:   "Pod can't mount a volume",
			Evidence: map[string]any{
				"failure_count": count,
				"event_message": truncate(msg, 300),
			},
			Recommendation: "Check the volume source exists: PVC bound, StorageClass provisioner working, secret/configMap present, kubelet able to reach it.",
		})
	}
	return out
}

func mountFailure(s *k8s.Snapshot, namespace, podName string) (string, int) {
	msg := ""
	count := 0
	for _, e := range s.Events {
		if e.InvolvedObject.Kind != "Pod" || e.InvolvedObject.Name != podName || e.InvolvedObject.Namespace != namespace {
			continue
		}
		if e.Type != "Warning" || e.Reason != "FailedMount" {
			continue
		}
		count += int(e.Count)
		if msg == "" {
			msg = e.Message
		}
	}
	if count == 0 {
		count = 1
	}
	if msg == "" {
		return "", 0
	}
	return msg, count
}

func accessModeNames(modes []corev1.PersistentVolumeAccessMode) []string {
	out := make([]string, 0, len(modes))
	for _, m := range modes {
		out = append(out, string(m))
	}
	return out
}
