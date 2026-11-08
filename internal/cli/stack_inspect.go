package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func newStackInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect <stack-id>",
		Aliases: []string{"get", "show"},
		Short:   "Inspect a stack",
		Long:    "Retrieve detailed information about a specific stack by its ID.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid stack ID '%s': must be an integer", args[0])
			}

			c, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			stack, err := c.Stacks().Inspect(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to inspect stack %d: %w", id, err)
			}

			if endpointMap, err := c.Endpoints().GetMap(ctx); err == nil {
				stack.EndpointName = endpointMap[stack.EndpointID]
			}

			return PrintFormatted(os.Stdout, cfg.Output, stack, func() error {
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintf(w, "ID:\t%d\n", stack.ID)
				fmt.Fprintf(w, "Name:\t%s\n", stack.Name)
				fmt.Fprintf(w, "Type:\t%s\n", stack.Type.String())
				fmt.Fprintf(w, "Status:\t%s\n", stack.Status.String())
				fmt.Fprintf(w, "Endpoint ID:\t%d\n", stack.EndpointID)
				if stack.EndpointName != "" {
					fmt.Fprintf(w, "Environment:\t%s\n", stack.EndpointName)
				}
				if stack.SwarmID != "" {
					fmt.Fprintf(w, "Swarm ID:\t%s\n", stack.SwarmID)
				}
				if stack.EntryPoint != "" {
					fmt.Fprintf(w, "Entry Point:\t%s\n", stack.EntryPoint)
				}
				if stack.ProjectPath != "" {
					fmt.Fprintf(w, "Project Path:\t%s\n", stack.ProjectPath)
				}
				if stack.CreatedBy != "" {
					fmt.Fprintf(w, "Created By:\t%s (%s)\n", stack.CreatedBy, stack.CreatedAt().Format(time.RFC3339))
				}
				if stack.UpdatedBy != "" {
					fmt.Fprintf(w, "Updated By:\t%s (%s)\n", stack.UpdatedBy, stack.UpdatedAt().Format(time.RFC3339))
				}
				if stack.GitConfig != nil {
					fmt.Fprintf(w, "Git URL:\t%s\n", stack.GitConfig.URL)
					fmt.Fprintf(w, "Git Ref:\t%s\n", stack.GitConfig.ReferenceName)
				}
				if stack.AutoUpdate != nil {
					fmt.Fprintf(w, "Auto Update Interval:\t%s\n", stack.AutoUpdate.Interval)
					if stack.AutoUpdate.Webhook != "" {
						fmt.Fprintf(w, "Auto Update Webhook:\t%s\n", stack.AutoUpdate.Webhook)
					}
				}
				if len(stack.Env) > 0 {
					fmt.Fprintf(w, "Env Variables:\t(%d variables)\n", len(stack.Env))
					for _, env := range stack.Env {
						fmt.Fprintf(w, "  - %s:\t%s\n", env.Name, env.Value)
					}
				}
				return w.Flush()
			})
		},
	}

	return cmd
}
