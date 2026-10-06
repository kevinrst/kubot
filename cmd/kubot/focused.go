package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kevinrst/kubot/internal/diagnose"
	"github.com/kevinrst/kubot/internal/k8s"
	"github.com/kevinrst/kubot/internal/model"
	"github.com/kevinrst/kubot/internal/render"
)

func newResourcesCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Per-container requests, limits, and live usage, plus resource problems",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()

			rep, snap, err := gather(ctx)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if !asJSON {
				render.PrintResourcesTable(out, resourceRows(snap), render.UseColor(noColorFlag), 50)
				fmt.Fprintln(out)
			}
			keep := map[string]bool{
				"pod_oom_killed": true, "pod_pending": true,
				"pod_missing_resources": true,
				"pod_low_limit":         true, "pod_near_limit": true,
				"deployment_resource_risk": true,
			}
			var f []model.Finding
			for _, i := range rep.Issues {
				if !keep[i.Reason] {
					continue
				}
				if i.Reason == "pod_pending" && !pendingCapacity(i) {
					continue
				}
				f = append(f, i)
			}
			rep.Issues = f
			rep.Status = model.OverallStatus(f)
			return render.PrintReport(out, rep, render.Options{NoColor: noColorFlag, JSON: asJSON, Width: terminalWidth(), NoScore: true})
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable JSON output")
	return cmd
}

func resourceRows(snap *k8s.Snapshot) []render.ResourceRow {
	var rows []render.ResourceRow
	for _, pod := range snap.Pods {
		name := pod.Name
		if pod.Namespace != "" {
			name = pod.Namespace + "/" + name
		}
		for _, c := range pod.Spec.Containers {
			r := render.ResourceRow{
				Pod: name, Container: c.Name, UseRatio: -1, MemUse: "-", MemLimit: "-",
				CPUReq: dash(c.Resources.Requests[corev1.ResourceCPU]),
				MemReq: dash(c.Resources.Requests[corev1.ResourceMemory]),
			}
			if q, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
				r.MemLimit = q.String()
				if used, ok := snap.UsageOf(pod.Namespace, pod.Name, c.Name); ok && q.Value() > 0 {
					r.MemUse = fmtMi(used)
					r.UseRatio = float64(used) / float64(q.Value())
					r.Hot = r.UseRatio >= diagnose.NearLimitRatio
				}
			}
			rows = append(rows, r)
		}
	}
	return rows
}

func dash(q resource.Quantity) string {
	if q.IsZero() {
		return "-"
	}
	return q.String()
}

func pendingCapacity(f model.Finding) bool {
	msg, _ := f.Evidence["scheduler_message"].(string)
	return strings.Contains(msg, "Insufficient cpu") || strings.Contains(msg, "Insufficient memory")
}

func fmtMi(b int64) string {
	if b < 1024*1024 {
		return "0Mi"
	}
	return fmt.Sprintf("%dMi", b/(1024*1024))
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
