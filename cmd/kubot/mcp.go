package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kevinrst/kubot/internal/diagnose"
	"github.com/kevinrst/kubot/internal/mcp"
	"github.com/kevinrst/kubot/internal/model"
)

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run as an MCP server over stdio (for AI agents like Claude)",
		Long: "Speaks the Model Context Protocol on stdin/stdout. Configure it in an MCP\n" +
			"client (Claude Desktop/Code, Cursor, …) and the agent gains read-only tools:\n" +
			"  inspect  — cluster or workload findings as JSON\n" +
			"  why      — explain why a workload is unhealthy\n\n" +
			"kubot never modifies the cluster. Point KUBECONFIG at the cluster in the\n" +
			"server's env so tools need no connection argument.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			srv := &mcp.Server{
				Name:    "kubot",
				Version: appVersion(),
				Instructions: "kubot gives read-only Kubernetes health findings. Call `inspect` and " +
					"explain its findings to the user — the findings are computed deterministically; " +
					"treat them as facts and carry any caveats into your advice.",
				Tools:     kubotTools(),
				Prompts:   kubotPrompts(),
				Resources: kubotResources(),
			}
			fmt.Fprintln(os.Stderr, "kubot mcp: serving on stdio (ctrl-c to stop)")
			return srv.Serve(cmd.Context(), os.Stdin, os.Stdout)
		},
	}
}

func kubotTools() []mcp.Tool {
	scopeSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"workload": map[string]any{
				"type":        "string",
				"description": "Deployment, service, or pod name to focus on. Omit for the whole cluster.",
			},
			"namespace": map[string]any{
				"type":        "string",
				"description": "Limit to one namespace. Omit for all namespaces.",
			},
		},
	}
	return []mcp.Tool{
		{
			Name: "inspect",
			Description: "Run a read-only Kubernetes health inspection and return the " +
				"findings (critical/warning/note) with evidence and recommendations as JSON. " +
				"Deterministic — computed in Go, not by a model. kubot never writes to the cluster.",
			InputSchema: scopeSchema,
			Handler:     inspectTool,
		},
		{
			Name: "why",
			Description: "Explain why a Kubernetes workload is unhealthy: the deterministic " +
				"findings for one deployment, service, or pod with evidence. Read-only.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"workload"},
				"properties": map[string]any{
					"workload": map[string]any{
						"type":        "string",
						"description": "Deployment, service, or pod name.",
					},
					"namespace": map[string]any{
						"type":        "string",
						"description": "Limit to one namespace. Omit for all namespaces.",
					},
				},
			},
			Handler: whyTool,
		},
	}
}

func kubotPrompts() []mcp.Prompt {
	return []mcp.Prompt{
		{
			Name:        "diagnose",
			Description: "Inspect the cluster and give a prioritized diagnosis.",
			Arguments: []mcp.PromptArg{
				{Name: "workload", Description: "Optional workload to focus on."},
			},
			Build: func(_ context.Context, args map[string]string) ([]mcp.PromptMessage, error) {
				wl := args["workload"]
				text := "Inspect the Kubernetes cluster with kubot's `inspect` tool"
				if wl != "" {
					text += " focused on workload `" + wl + "`"
				}
				return []mcp.PromptMessage{{
					Role: "user",
					Text: text + " and give me a prioritized diagnosis: what is wrong, why, " +
						"and what to check next. Treat kubot's findings as facts.",
				}}, nil
			},
		},
	}
}

func kubotResources() []mcp.Resource {
	return []mcp.Resource{
		{
			URI:         "kubot://schema",
			Name:        "inspect schema",
			Description: "The versioned JSON Schema for kubot's inspect report.",
			MimeType:    "application/json",
			Read: func(_ context.Context) (string, error) {
				return string(model.ReportSchema()), nil
			},
		},
	}
}

func diagnoseForTool(ctx context.Context, namespace, workload string) (model.Report, error) {
	rep, snap, err := gatherScoped(ctx, namespace)
	if err != nil {
		return model.Report{}, err
	}
	rep.Issues = diagnose.FilterByWorkload(rep.Issues, workload, namespace, snap)
	if rep.Issues == nil {
		rep.Issues = []model.Finding{}
	}
	rep.Status = model.OverallStatus(rep.Issues)
	return rep, nil
}

func inspectTool(ctx context.Context, raw json.RawMessage) (string, error) {
	var args struct {
		Workload  string `json:"workload"`
		Namespace string `json:"namespace"`
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid params: %w", err)
		}
	}
	rep, err := diagnoseForTool(ctx, args.Namespace, args.Workload)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func whyTool(ctx context.Context, raw json.RawMessage) (string, error) {
	var args struct {
		Workload  string `json:"workload"`
		Namespace string `json:"namespace"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if args.Workload == "" {
		return "", fmt.Errorf("why requires workload")
	}
	return inspectTool(ctx, raw)
}
