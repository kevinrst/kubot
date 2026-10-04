package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"kubot/internal/render"
)

func newEventsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Show Warning events related to unhealthy workloads",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()

			rep, snap, err := gather(ctx)
			if err != nil {
				return err
			}
			if asJSON {
				return render.PrintReport(cmd.OutOrStdout(), rep, render.Options{NoColor: noColorFlag, JSON: true})
			}
			type row struct {
				ns, kind, name, reason string
				count                  int32
				msg                    string
			}
			var rows []row
			for _, e := range snap.Events {
				if e.Type != "Warning" {
					continue
				}
				rows = append(rows, row{
					ns: e.InvolvedObject.Namespace, kind: e.InvolvedObject.Kind,
					name: e.InvolvedObject.Name, reason: e.Reason,
					count: e.Count, msg: e.Message,
				})
			}
			sort.Slice(rows, func(i, j int) bool {
				if rows[i].count != rows[j].count {
					return rows[i].count > rows[j].count
				}
				return rows[i].name < rows[j].name
			})
			if len(rows) > 20 {
				rows = rows[:20]
			}
			out := cmd.OutOrStdout()
			if len(rows) == 0 {
				fmt.Fprintln(out, "No warning events.")
				return nil
			}
			for _, r := range rows {
				ns := r.ns
				if ns == "" {
					ns = "-"
				}
				fmt.Fprintf(out, "%-12s %-10s %-30s %-22s x%d\n  %s\n",
					ns, r.kind, r.name, r.reason, r.count, truncateMsg(r.msg, 220))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable JSON output")
	return cmd
}

func truncateMsg(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
