package diagnose

func DefaultRules() []Rule {
	return []Rule{
		PodCrashLoopRule{},
		PodOOMKilledRule{},
		PodImagePullRule{},
		PodPendingRule{},
		PodProbeFailingRule{},
		DeploymentUnavailableRule{},
		ServiceNoEndpointsRule{},
		PodMissingResourcesRule{},
		PodLowLimitRule{},
		PodNearLimitRule{},
		DeploymentResourceRiskRule{},
		DeploymentStalledRule{},
		ReplicaSetLeftoverRule{},
		PVCPendingRule{},
		PodMountFailureRule{},
		NodePressureRule{},
	}
}

func CheckedSubsystems() []string {
	return []string{"pods", "deployments", "services", "nodes", "events"}
}
