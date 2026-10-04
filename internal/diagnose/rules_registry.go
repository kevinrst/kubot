package diagnose

// Keep this list small: a few problems done well beats broad shallow coverage.
func DefaultRules() []Rule {
	return []Rule{
		PodCrashLoopRule{},
		PodOOMKilledRule{},
		PodImagePullRule{},
		PodPendingRule{},
		PodProbeFailingRule{},
		DeploymentUnavailableRule{},
		ServiceNoEndpointsRule{},
	}
}

func CheckedSubsystems() []string {
	return []string{"pods", "deployments", "services", "nodes", "events"}
}
