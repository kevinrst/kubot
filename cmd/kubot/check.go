package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	var failOn string
	cmd := &cobra.Command{
		Use:   "check",
		Short: "CI gate: exit non-zero when problems meet --fail-on",
		Long: "Minimal output plus the scriptable exit code for pipelines: " +
			"0 clean · 1 warning · 2 critical · 3 connection failure · 64 bad flags.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !validFailOn(failOn) {
				return usageErrf("--fail-on must be critical|warn|info|none, got %q", failOn)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()

			rep, _, err := gather(ctx)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "status: %s, issues: %d\n", rep.Status, len(rep.Issues))
			for _, i := range rep.Issues {
				muted := ""
				if i.Suppressed {
					muted = " (muted)"
				}
				fmt.Fprintf(out, "%s %s: %s%s\n", i.Severity, i.Resource, i.Message, muted)
			}
			return exitError{exitCode(rep.Issues, failOn)}
		},
	}
	cmd.Flags().StringVar(&failOn, "fail-on", "critical", "severity that exits non-zero: critical|warn|info|none")
	return cmd
}
