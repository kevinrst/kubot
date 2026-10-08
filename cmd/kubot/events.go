package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubotdev/kubot/internal/render"
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
				return render.PrintReport(cmd.OutOrStdout(), rep, render.Options{NoColor: noColorFlag, JSON: true, Width: terminalWidth()})
			}
			var rows []render.EventRow
			for _, e := range snap.Events {
				if e.Type != "Warning" {
					continue
				}
				ns := e.InvolvedObject.Namespace
				if ns == "" {
					ns = "-"
				}
				rows = append(rows, render.EventRow{
					Namespace: ns,
					Object:    fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
					Reason:    e.Reason,
					Count:     e.Count,
					Message:   e.Message,
				})
			}
			out := cmd.OutOrStdout()
			if len(rows) == 0 {
				fmt.Fprintln(out, "No warning events.")
				return nil
			}
			render.PrintEventsTable(out, rows, render.UseColor(noColorFlag), 20, terminalWidth())
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable JSON output")
	return cmd
}
