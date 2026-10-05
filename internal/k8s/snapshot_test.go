package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventsForPod_filtersAndOrders(t *testing.T) {
	old := corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Name: "e1", Namespace: "default"},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "p1", Namespace: "default"},
		Reason:         "Scheduled",
	}
	old.LastTimestamp = metav1.Unix(1000, 0)
	recent := old
	recent.Name = "e2"
	recent.Reason = "FailedScheduling"
	recent.LastTimestamp = metav1.Unix(2000, 0)
	other := old
	other.Name = "e3"
	other.InvolvedObject.Name = "p2"

	s := &Snapshot{Events: []corev1.Event{old, other, recent}}
	got := s.EventsForPod("default", "p1")
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Name != "e2" {
		t.Fatalf("expected newest first, got %s", got[0].Name)
	}
	if len(s.EventsForPod("other", "p1")) != 0 {
		t.Fatal("namespace must filter")
	}
}

func TestReadyEndpoints_slices(t *testing.T) {
	ready := true
	s := &Snapshot{EndpointSlices: []discoveryv1.EndpointSlice{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "shop-1", Namespace: "default",
				Labels: map[string]string{discoveryv1.LabelServiceName: "shop"},
			},
			Endpoints: []discoveryv1.Endpoint{
				{Conditions: discoveryv1.EndpointConditions{Ready: &ready}},
				{Conditions: discoveryv1.EndpointConditions{Ready: &ready}},
			},
		},
	}}
	r, total := s.ReadyEndpoints("default", "shop")
	if r != 2 || total != 2 {
		t.Fatalf("got ready=%d total=%d", r, total)
	}
	if r2, _ := s.ReadyEndpoints("default", "missing"); r2 != 0 {
		t.Fatalf("missing service must have 0 ready, got %d", r2)
	}
}

func TestPodsByTopOwner_stripsReplicaSetHash(t *testing.T) {
	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "payments-api-7d8f9c6f8b-x2k4m",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "payments-api-7d8f9c6f8b"},
			},
		},
	}
	s := &Snapshot{Pods: []corev1.Pod{p}}
	m := s.PodsByTopOwner()
	if _, ok := m["payments-api"]; !ok {
		t.Fatalf("expected payments-api owner, got %v", m)
	}
}
