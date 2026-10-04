package main

import (
	"context"

	"github.com/kevinrst/kubot/internal/diagnose"
	"github.com/kevinrst/kubot/internal/k8s"
	"github.com/kevinrst/kubot/internal/model"
)

// Shared pipeline every command and MCP tool runs: connect, collect, diagnose.
func gather(ctx context.Context) (model.Report, *k8s.Snapshot, error) {
	cs, err := k8s.NewClientset(k8s.Options{Kubeconfig: kubeconfigFlag, Context: contextFlag})
	if err != nil {
		return model.Report{}, nil, err
	}
	snap, err := k8s.Collect(ctx, cs, namespaceFlag, timeoutFlag)
	if err != nil {
		return model.Report{}, nil, err
	}
	snap.Context = k8s.ResolveContextName(k8s.Options{Kubeconfig: kubeconfigFlag, Context: contextFlag})
	engine := diagnose.NewEngine()
	findings := engine.Run(snap)
	if findings == nil {
		findings = []model.Finding{}
	}
	return model.Report{
		SchemaVersion: model.SchemaVersion,
		Cluster:       model.ClusterInfo{Context: snap.Context, Namespace: namespaceFlag},
		Status:        model.OverallStatus(findings),
		Issues:        findings,
		Checked:       diagnose.CheckedSubsystems(),
	}, snap, nil
}
