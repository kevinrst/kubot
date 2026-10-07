package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/kevinrst/kubot/internal/ai"
	"github.com/kevinrst/kubot/internal/diagnose"
	"github.com/kevinrst/kubot/internal/model"
	"github.com/kevinrst/kubot/internal/render"
)

func newAskCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Ask about cluster health, answered from deterministic findings",
		Long: "Gather findings read-only, send them to a model, and print a " +
			"plain-language answer. The findings are computed in Go; the model " +
			"only narrates them. Keys come from env, never flags.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()

			rep, _, err := gather(ctx)
			if err != nil {
				return err
			}
			return explainWithModel(cmd, rep, strings.Join(args, " "), yes)
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "skip the data-disclosure confirmation")
	return cmd
}

func newExplainCmd() *cobra.Command {
	var yes bool
	var workload string
	cmd := &cobra.Command{
		Use:   "explain [workload]",
		Short: "Print the deterministic report plus an AI reading of it",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				workload = args[0]
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeoutFlag)
			defer cancel()

			rep, snap, err := gather(ctx)
			if err != nil {
				return err
			}
			if workload != "" {
				rep.Issues = diagnose.FilterByWorkload(rep.Issues, workload, namespaceFlag, snap)
			}
			if err := render.PrintReport(cmd.OutOrStdout(), rep, render.Options{NoColor: noColorFlag, Width: terminalWidth()}); err != nil {
				return err
			}
			return explainWithModel(cmd, rep, "", yes)
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "skip the data-disclosure confirmation")
	return cmd
}

// explainWithModel sends one report to the configured model, after telling
// the user exactly what leaves the machine. Local endpoints skip confirmation;
// without a key the deterministic report above still stands.
func explainWithModel(cmd *cobra.Command, rep model.Report, question string, yes bool) error {
	p, err := ai.Resolve()
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if !yes && !p.Local() {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf("not a terminal: pass --yes to send findings to %s", p.Name())
		}
		fmt.Fprintf(os.Stderr, "Send the findings to %s (%s) for explanation? [y/N] ", p.Name(), p.Model())
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.ToLower(strings.TrimSpace(line)) != "y" {
			fmt.Fprintln(os.Stderr, "Aborted. The deterministic report above still stands.")
			return nil
		}
	}
	text, err := p.Complete(cmd.Context(), ai.SystemPrompt(), ai.UserPrompt(question, rep))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return fmt.Errorf("model: %w (deterministic findings above are unaffected; the model id may be retired — retry with KUBOT_AI_MODEL=<id>)", err)
		}
		return fmt.Errorf("model: %w (deterministic findings above are unaffected)", err)
	}
	printAIHeader(out, p.Name(), p.Model(), render.UseColor(noColorFlag))
	fmt.Fprintf(out, "%s\n", strings.TrimSpace(text))
	return nil
}

func printAIHeader(out io.Writer, provider, model string, color bool) {
	label := fmt.Sprintf("AI reading (%s/%s — verify before acting)", provider, model)
	if !color {
		fmt.Fprintf(out, "\n%s:\n\n", label)
		return
	}
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(strings.Repeat("─", 60))
	head := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(label)
	fmt.Fprintf(out, "\n%s\n%s\n\n", bar, head)
}
