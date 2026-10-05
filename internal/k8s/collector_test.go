package k8s

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// No live cluster needed.
func TestCollect_fakeClient(t *testing.T) {
	cs := fake.NewClientset(
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "default"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "s1", Namespace: "default"}},
		&corev1.Event{ObjectMeta: metav1.ObjectMeta{Name: "e1", Namespace: "default"}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}},
		&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "data", Namespace: "default"}},
	)
	snap, err := Collect(context.Background(), cs, "", 10*time.Second)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(snap.Pods) != 1 || len(snap.Services) != 1 || len(snap.Events) != 1 || len(snap.Nodes) != 1 || len(snap.PVCs) != 1 {
		t.Fatalf("incomplete snapshot: %+v", snap)
	}
	if len(snap.Degraded) != 0 {
		t.Fatalf("unexpected degradations: %v", snap.Degraded)
	}
}

// -n filters namespaced resources; nodes (cluster-scoped) still come back.
func TestCollect_namespaceScope(t *testing.T) {
	cs := fake.NewClientset(
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "a"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "b"}},
	)
	snap, err := Collect(context.Background(), cs, "a", 10*time.Second)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(snap.Pods) != 1 || snap.Pods[0].Name != "p1" {
		t.Fatalf("namespace filter broken: %+v", snap.Pods)
	}
}

func TestResolveContextName_explicit(t *testing.T) {
	if got := ResolveContextName(Options{Context: "my-ctx"}); got != "my-ctx" {
		t.Fatalf("got %q", got)
	}
}
