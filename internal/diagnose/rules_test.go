package diagnose

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kevinrst/kubot/internal/k8s"
)

func crashPod() corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "payments-api-7d8f9c6f8b-x2k4m", Namespace: "default"},
		Status: corev1.PodStatus{
			Phase: "Running",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "api",
					RestartCount: 14,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "back-off 5m0s restarting failed container"},
					},
					LastTerminationState: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"},
					},
				},
			},
		},
	}
}

func TestCrashLoop(t *testing.T) {
	s := &k8s.Snapshot{Pods: []corev1.Pod{crashPod()}}
	f := PodCrashLoopRule{}.Run(s)
	if len(f) != 1 || f[0].Reason != "pod_crashloop_backoff" || f[0].Severity != "critical" {
		t.Fatalf("expected crashloop finding, got %+v", f)
	}
}

func TestOOMKilled(t *testing.T) {
	p := crashPod()
	p.Status.ContainerStatuses[0].State.Waiting = nil
	p.Status.ContainerStatuses[0].State.Terminated = &corev1.ContainerStateTerminated{ExitCode: 137, Reason: "OOMKilled"}
	s := &k8s.Snapshot{Pods: []corev1.Pod{p}}
	f := PodOOMKilledRule{}.Run(s)
	if len(f) != 1 || f[0].Reason != "pod_oom_killed" {
		t.Fatalf("expected oom finding, got %+v", f)
	}
	if f[0].Evidence["exit_code"] != int32(137) && f[0].Evidence["exit_code"] != 137 {
		t.Fatalf("expected exit 137 evidence, got %+v", f[0].Evidence)
	}
}

func TestImagePull(t *testing.T) {
	p := crashPod()
	p.Status.ContainerStatuses[0].State.Waiting = &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff", Message: "Back-off pulling image"}
	p.Status.ContainerStatuses[0].LastTerminationState = corev1.ContainerState{}
	s := &k8s.Snapshot{Pods: []corev1.Pod{p}}
	got := (PodImagePullRule{}).Run(s)
	if len(got) != 1 {
		t.Fatalf("expected imagepull, got %+v", got)
	}
}

func TestPending(t *testing.T) {
	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "worker-0", Namespace: "default"},
		Status:     corev1.PodStatus{Phase: "Pending"},
	}
	ev := corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Namespace: "default"},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "worker-0", Namespace: "default"},
		Reason:         "FailedScheduling",
		Message:        "0/3 nodes are available: 3 Insufficient cpu.",
	}
	s := &k8s.Snapshot{Pods: []corev1.Pod{p}, Events: []corev1.Event{ev}}
	f := PodPendingRule{}.Run(s)
	if len(f) != 1 || f[0].Reason != "pod_pending" {
		t.Fatalf("expected pending, got %+v", f)
	}
}

func TestPendingSkipsScheduledPod(t *testing.T) {
	// Assigned to a node (image pulling): not a scheduling problem.
	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pulling-0", Namespace: "default"},
		Spec:       corev1.PodSpec{NodeName: "node-1"},
		Status:     corev1.PodStatus{Phase: "Pending"},
	}
	ev := corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Namespace: "default"},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "pulling-0", Namespace: "default"},
		Reason:         "Scheduled",
		Message:        "Successfully assigned default/pulling-0 to node-1",
	}
	s := &k8s.Snapshot{Pods: []corev1.Pod{p}, Events: []corev1.Event{ev}}
	if got := (PodPendingRule{}).Run(s); len(got) != 0 {
		t.Fatalf("scheduled pod must not be a pending finding, got %+v", got)
	}
}

func TestCrashLoopBackoffWindow(t *testing.T) {
	// Between backoffs: Error state, 3 restarts, non-zero last exit.
	p := crashPod()
	p.Status.ContainerStatuses[0].State.Waiting = &corev1.ContainerStateWaiting{Reason: "Error", Message: "some error"}
	p.Status.ContainerStatuses[0].RestartCount = 3
	p.Status.ContainerStatuses[0].LastTerminationState = corev1.ContainerState{
		Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"},
	}
	s := &k8s.Snapshot{Pods: []corev1.Pod{p}}
	got := (PodCrashLoopRule{}).Run(s)
	if len(got) != 1 || got[0].Reason != "pod_crashloop_backoff" {
		t.Fatalf("expected crashloop in backoff window, got %+v", got)
	}
}

func TestCrashLoopTerminatedWindow(t *testing.T) {
	// Crash landed, kubelet hasn't restarted yet: Terminated exit 1, 3 restarts.
	p := crashPod()
	p.Status.ContainerStatuses[0].State = corev1.ContainerState{
		Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"},
	}
	p.Status.ContainerStatuses[0].RestartCount = 3
	p.Status.ContainerStatuses[0].LastTerminationState = corev1.ContainerState{
		Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"},
	}
	s := &k8s.Snapshot{Pods: []corev1.Pod{p}}
	got := (PodCrashLoopRule{}).Run(s)
	if len(got) != 1 || got[0].Reason != "pod_crashloop_backoff" {
		t.Fatalf("expected crashloop in terminated window, got %+v", got)
	}
}

func TestProbeTransientIgnored(t *testing.T) {
	p := crashPod()
	p.Status.ContainerStatuses[0].State.Waiting = nil
	ev := corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Namespace: "default"},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: p.Name, Namespace: "default"},
		Type:           "Warning",
		Reason:         "Unhealthy",
		Message:        "Readiness probe failed: connection refused",
		Count:          1,
	}
	s := &k8s.Snapshot{Pods: []corev1.Pod{p}, Events: []corev1.Event{ev}}
	if got := (PodProbeFailingRule{}).Run(s); len(got) != 0 {
		t.Fatalf("transient probe blip must be silent, got %+v", got)
	}
}

func TestProbeFailing(t *testing.T) {
	p := crashPod()
	p.Status.ContainerStatuses[0].State.Waiting = nil
	ev := corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Namespace: "default"},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: p.Name, Namespace: "default"},
		Type:           "Warning",
		Reason:         "Unhealthy",
		Message:        "Readiness probe failed: HTTP probe failed with statuscode: 500",
		Count:          12,
	}
	s := &k8s.Snapshot{Pods: []corev1.Pod{p}, Events: []corev1.Event{ev}}
	f := PodProbeFailingRule{}.Run(s)
	if len(f) != 1 || f[0].Reason != "pod_probe_failing" {
		t.Fatalf("expected probe failing, got %+v", f)
	}
}

func TestDeploymentUnavailable(t *testing.T) {
	rep := int32(2)
	d := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "payments-api", Namespace: "default"},
		Spec:       appsv1.DeploymentSpec{Replicas: &rep},
		Status:     appsv1.DeploymentStatus{ReadyReplicas: 1, UnavailableReplicas: 1},
	}
	s := &k8s.Snapshot{Deployments: []appsv1.Deployment{d}}
	f := DeploymentUnavailableRule{}.Run(s)
	if len(f) != 1 || f[0].Reason != "deployment_unavailable" {
		t.Fatalf("expected deployment_unavailable, got %+v", f)
	}
}

func TestServiceNoEndpoints(t *testing.T) {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "shop", Namespace: "default"},
		Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "shop"}},
	}
	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "other-1", Namespace: "default", Labels: map[string]string{"app": "other"}},
	}
	s := &k8s.Snapshot{Services: []corev1.Service{svc}, Pods: []corev1.Pod{p}}
	f := ServiceNoEndpointsRule{}.Run(s)
	if len(f) != 1 || f[0].Reason != "service_no_endpoints" {
		t.Fatalf("expected service_no_endpoints, got %+v", f)
	}
}

func TestEngineSortDeterministic(t *testing.T) {
	s := &k8s.Snapshot{
		Pods: []corev1.Pod{crashPod()},
		Services: []corev1.Service{{
			ObjectMeta: metav1.ObjectMeta{Name: "shop", Namespace: "default"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "missing"}},
		}},
	}
	e := NewEngine()
	a, b := e.Run(s), e.Run(s)
	if len(a) != len(b) {
		t.Fatal("non-deterministic length")
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			t.Fatalf("non-deterministic finding %d: %+v vs %+v", i, a[i], b[i])
		}
	}
	if len(a) > 0 && a[0].Severity != "critical" {
		t.Fatalf("expected critical first, got %+v", a[0])
	}
}

func TestFilterByWorkload(t *testing.T) {
	s := &k8s.Snapshot{Pods: []corev1.Pod{crashPod()}}
	e := NewEngine()
	all := e.Run(s)
	filtered := FilterByWorkload(all, "payments-api", "", s.PodsByTopOwner())
	if len(filtered) == 0 {
		t.Fatal("expected workload match")
	}
	none := FilterByWorkload(all, "something-else-entirely", "", s.PodsByTopOwner())
	if len(none) != 0 {
		t.Fatalf("expected no match, got %+v", none)
	}
}
