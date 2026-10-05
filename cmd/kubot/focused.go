package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/kevinrst/kubot/internal/model"
	"github.com/kevinrst/kubot/internal/render"
)

func newResourcesCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Show resource-related problems (OOM, pending, limits)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()

			rep, _, err := gather(ctx)
			if err != nil {
				return err
			}
			keep := map[string]bool{
				"pod_oom_killed": true, "pod_pending": true,
				"deployment_unavailable": true, "pod_missing_resources": true,
				"pod_low_limit": true, "pod_near_limit": true,
				"deployment_resource_risk": true,
			}
			var f []model.Finding
			for _, i := range rep.Issues {
				if keep[i.Reason] {
					f = append(f, i)
				}
			}
			rep.Issues = f
			rep.Status = model.OverallStatus(f)
			return render.PrintReport(cmd.OutOrStdout(), rep, render.Options{NoColor: noColorFlag, JSON: asJSON, Width: terminalWidth(), NoScore: true})
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable JSON output")
	return cmd
}

func newNetworkingCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "networking",
		Short: "Check Service and endpoint problems",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()

			rep, _, err := gather(ctx)
			if err != nil {
				return err
			}
			var f []model.Finding
			for _, i := range rep.Issues {
				if i.Reason == "service_no_endpoints" || i.Reason == "pod_probe_failing" {
					f = append(f, i)
				}
			}
			rep.Issues = f
			rep.Status = model.OverallStatus(f)
			return render.PrintReport(cmd.OutOrStdout(), rep, render.Options{NoColor: noColorFlag, JSON: asJSON, Width: terminalWidth(), NoScore: true})
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable JSON output")
	return cmd
}
