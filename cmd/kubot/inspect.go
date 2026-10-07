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

// Scriptable exit codes: 0 clean, 1 warning, 2 critical, 3 failure, 64 bad flags.
const (
	exitClean    = 0
	exitWarn     = 1
	exitCritical = 2
	exitFailure  = 3
	exitUsage    = 64
)

type inspectFlags struct {
	json     bool
	format   string // text|json
	failOn   string // critical|warn|info|none
	full     bool   // evidence + recommendations per finding
	workload string
}

func newInspectCmd() *cobra.Command {
	var f inspectFlags
	cmd := &cobra.Command{
		Use:   "inspect [workload]",
		Short: "Inspect the cluster or one workload",
		Long: "Collect cluster state read-only, run the deterministic diagnostic " +
			"engine, and print a findings-first report (or --json). " +
			"kubot never modifies the cluster.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.workload = argAt(args, 0)
			return runInspect(cmd, f)
		},
	}
	fl := cmd.Flags()
	fl.BoolVar(&f.json, "json", false, "emit the versioned Report as JSON (the agent/script contract)")
	fl.StringVar(&f.format, "format", "text", "output format: text|json")
	fl.StringVar(&f.failOn, "fail-on", "none", "exit non-zero on findings at/above this severity: critical|warn|info|none (default none; use check for CI gating)")
	fl.BoolVar(&f.full, "full", false, "show evidence and recommendations per finding")
	return cmd
}

func runInspect(cmd *cobra.Command, f inspectFlags) error {
	if !validFailOn(f.failOn) {
		return usageErrf("--fail-on must be critical|warn|info|none, got %q", f.failOn)
	}
	if f.json && f.format == "text" {
		f.format = "json" // --json is a shortcut for --format=json
	}
	if !validFormat(f.format) {
		return usageErrf("--format must be text|json, got %q", f.format)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
	defer cancel()

	rep, snap, err := gather(ctx)
	if err != nil {
		return err
	}
	if f.workload != "" {
		rep.Issues = diagnose.FilterByWorkload(rep.Issues, f.workload, namespaceFlag, snap)
		rep.Status = model.OverallStatus(rep.Issues)
	}

	if err := render.PrintReport(cmd.OutOrStdout(), rep, render.Options{NoColor: noColorFlag, JSON: f.format == "json", Full: f.full, Width: terminalWidth(), NoScore: f.workload != ""}); err != nil {
		return err
	}
	if f.workload != "" && len(rep.Issues) == 0 && f.format == "text" {
		if s := suggestWorkload(snap, f.workload); len(s) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Did you mean: %s?\n", strings.Join(s, ", "))
		}
	}
	return exitError{exitCode(rep.Issues, f.failOn)}
}

// Maps findings to the exit code, counting only severities at/above failOn.
func exitCode(fs []model.Finding, failOn string) int {
	if failOn == "none" {
		return exitClean
	}
	threshold := severityRank(failOn)
	hasCritical, hasAtThreshold := false, false
	for _, f := range fs {
		if f.Suppressed || severityRank(f.Severity) < threshold {
			continue
		}
		hasAtThreshold = true
		if f.Severity == model.SeverityCritical {
			hasCritical = true
		}
	}
	switch {
	case hasCritical:
		return exitCritical
	case hasAtThreshold:
		return exitWarn
	default:
		return exitClean
	}
}

func severityRank(s string) int {
	switch s {
	case model.SeverityCritical: // "critical"
		return 2
	case model.SeverityWarning, "warn": // "warning"
		return 1
	default: // note, info, ok
		return 0
	}
}

func validFailOn(s string) bool {
	switch s {
	case "critical", "warn", "info", "none":
		return true
	}
	return false
}

func validFormat(s string) bool {
	switch s {
	case "text", "json":
		return true
	}
	return false
}
