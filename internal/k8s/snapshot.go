package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
)

// The cluster state the engine reasons over. No diagnosis lives here.
type Snapshot struct {
	Context        string
	Pods           []corev1.Pod
	Deployments    []appsv1.Deployment
	ReplicaSets    []appsv1.ReplicaSet
	Services       []corev1.Service
	EndpointSlices []discoveryv1.EndpointSlice
	Events         []corev1.Event
	Nodes          []corev1.Node
	Usage          []ContainerUsage
	PVCs           []corev1.PersistentVolumeClaim

	// Collectors that failed without aborting the run (e.g. events RBAC denied).
	Degraded []string
}

// Events for one pod, most recent first.
func (s *Snapshot) EventsForPod(namespace, podName string) []corev1.Event {
	var out []corev1.Event
	for _, e := range s.Events {
		if e.InvolvedObject.Kind == "Pod" &&
			e.InvolvedObject.Name == podName &&
			e.InvolvedObject.Namespace == namespace {
			out = append(out, e)
		}
	}
	// Newest first.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].LastTimestamp.Time.After(out[j-1].LastTimestamp.Time); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// Ready/total backend addresses for a service, from its EndpointSlices.
func (s *Snapshot) ReadyEndpoints(namespace, svcName string) (ready, total int) {
	for _, es := range s.EndpointSlices {
		if es.Namespace != namespace {
			continue
		}
		svc, ok := es.Labels[discoveryv1.LabelServiceName]
		if !ok || svc != svcName {
			continue
		}
		for _, ep := range es.Endpoints {
			total++
			if ep.Conditions.Ready == nil || *ep.Conditions.Ready {
				ready++
			}
		}
	}
	return ready, total
}

// Pod names keyed by workload name (deployment, statefulset, …).
func (s *Snapshot) PodsByTopOwner() map[string][]string {
	m := map[string][]string{}
	for _, p := range s.Pods {
		owner := TopOwnerName(&p)
		m[owner] = append(m[owner], p.Name)
	}
	return m
}

// TopOwnerName guesses the workload owning a pod
func TopOwnerName(p *corev1.Pod) string {
	for _, o := range p.OwnerReferences {
		if o.Kind == "ReplicaSet" {
			if d := stripHashSuffix(o.Name); d != "" {
				return d
			}
			return o.Name
		}
		if o.Kind == "StatefulSet" || o.Kind == "DaemonSet" || o.Kind == "Job" {
			return o.Name
		}
	}
	// No owner: fall back to the pod name.
	return p.Name
}

func stripHashSuffix(rs string) string {
	i := lastDash(rs)
	if i <= 0 || len(rs)-i-1 < 5 {
		return rs
	}
	for _, c := range rs[i+1:] {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		return rs
	}
	return rs[:i]
}

func lastDash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '-' {
			return i
		}
	}
	return -1
}
