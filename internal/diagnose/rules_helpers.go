package diagnose

import (
	corev1 "k8s.io/api/core/v1"
)

// The terminated state if the container was OOMKilled, else nil.
func oomTerm(cs corev1.ContainerStatus) *corev1.ContainerStateTerminated {
	if cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled" {
		return cs.State.Terminated
	}
	if cs.LastTerminationState.Terminated != nil &&
		cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
		return cs.LastTerminationState.Terminated
	}
	return nil
}
