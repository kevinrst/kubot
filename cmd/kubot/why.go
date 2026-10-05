package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kevinrst/kubot/internal/diagnose"
	"github.com/kevinrst/kubot/internal/model"
	"github.com/kevinrst/kubot/internal/render"
)

func newWhyCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "why <workload>",
		Short: "Explain why a workload is unhealthy",
		Long: "Filter the deterministic diagnosis to one workload (deployment, " +
			"service, or pod name) and show the symptom ← mechanism ← cause chain " +
			"with evidence. The model explains; kubot computes.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()

			rep, snap, err := gather(ctx)
			if err != nil {
				return err
			}
			rep.Issues = diagnose.FilterByWorkload(rep.Issues, args[0], namespaceFlag, snap.PodsByTopOwner())
			rep.Status = model.OverallStatus(rep.Issues)
			if len(rep.Issues) == 0 && !asJSON {
				fmt.Fprintf(cmd.OutOrStdout(), "No problems detected for %q.\n", args[0])
				if s := suggestWorkload(snap, args[0]); len(s) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "Did you mean: %s?\n", strings.Join(s, ", "))
				}
				return nil
			}
			return render.PrintReport(cmd.OutOrStdout(), rep, render.Options{NoColor: noColorFlag, JSON: asJSON})
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable JSON output")
	return cmd
}
