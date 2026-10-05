package k8s

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Collect(ctx context.Context, cs kubernetes.Interface, namespace string, timeout time.Duration) *Snapshot {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	snap := &Snapshot{}
	listOpts := metav1.ListOptions{Limit: 500}

	collect := func(name string, fn func() error) {
		if err := fn(); err != nil {
			if ctx.Err() != nil {
				return
			}
			snap.Degraded = append(snap.Degraded, name+": "+err.Error())
		}
	}

	collect("pods", func() error { return collectPods(ctx, listOpts, snap, cs, namespace) })
	collect("deployments", func() error { return collectDeployments(ctx, listOpts, snap, cs, namespace) })
	collect("replicasets", func() error { return collectReplicaSets(ctx, listOpts, snap, cs, namespace) })
	collect("services", func() error { return collectServices(ctx, listOpts, snap, cs, namespace) })
	collect("endpointslices", func() error { return collectEndpointSlices(ctx, listOpts, snap, cs, namespace) })
	collect("events", func() error { return collectEvents(ctx, listOpts, snap, cs, namespace) })
	collect("pvcs", func() error { return collectPVCs(ctx, listOpts, snap, cs, namespace) })
	collect("nodes", func() error { return collectNodes(ctx, listOpts, snap, cs, namespace) })

	if ctx.Err() == context.DeadlineExceeded {
		snap.Degraded = append(snap.Degraded, "collection timed out; results partial")
	}
	return snap
}

func collectPods(ctx context.Context, opts metav1.ListOptions, snap *Snapshot, cs kubernetes.Interface, ns string) error {
	res, err := cs.CoreV1().Pods(ns).List(ctx, opts)
	if err != nil {
		return err
	}
	snap.Pods = res.Items
	return nil
}

func collectDeployments(ctx context.Context, opts metav1.ListOptions, snap *Snapshot, cs kubernetes.Interface, ns string) error {
	res, err := cs.AppsV1().Deployments(ns).List(ctx, opts)
	if err != nil {
		return err
	}
	snap.Deployments = res.Items
	return nil
}

func collectReplicaSets(ctx context.Context, opts metav1.ListOptions, snap *Snapshot, cs kubernetes.Interface, ns string) error {
	res, err := cs.AppsV1().ReplicaSets(ns).List(ctx, opts)
	if err != nil {
		return err
	}
	snap.ReplicaSets = res.Items
	return nil
}

func collectServices(ctx context.Context, opts metav1.ListOptions, snap *Snapshot, cs kubernetes.Interface, ns string) error {
	res, err := cs.CoreV1().Services(ns).List(ctx, opts)
	if err != nil {
		return err
	}
	snap.Services = res.Items
	return nil
}

func collectEndpointSlices(ctx context.Context, opts metav1.ListOptions, snap *Snapshot, cs kubernetes.Interface, ns string) error {
	res, err := cs.DiscoveryV1().EndpointSlices(ns).List(ctx, opts)
	if err != nil {
		return err
	}
	snap.EndpointSlices = res.Items
	return nil
}

func collectEvents(ctx context.Context, opts metav1.ListOptions, snap *Snapshot, cs kubernetes.Interface, ns string) error {
	res, err := cs.CoreV1().Events(ns).List(ctx, opts)
	if err != nil {
		return err
	}
	snap.Events = res.Items
	return nil
}

func collectPVCs(ctx context.Context, opts metav1.ListOptions, snap *Snapshot, cs kubernetes.Interface, ns string) error {
	res, err := cs.CoreV1().PersistentVolumeClaims(ns).List(ctx, opts)
	if err != nil {
		return err
	}
	snap.PVCs = res.Items
	return nil
}

func collectNodes(ctx context.Context, opts metav1.ListOptions, snap *Snapshot, cs kubernetes.Interface, _ string) error {
	res, err := cs.CoreV1().Nodes().List(ctx, opts)
	if err != nil {
		return err
	}
	snap.Nodes = res.Items
	return nil
}
