package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func newStackRestartCmd() *cobra.Command {
	var endpointID int

	cmd := &cobra.Command{
		Use:   "restart <stack-id>",
		Short: "Restart a stack (stops then starts)",
		Long: `Restarts a stack by stopping all running services and starting them again.
Portainer API does not provide a direct restart endpoint, so this performs a sequential stop & start.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid stack ID '%s': must be an integer", args[0])
			}

			c, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 120*time.Second)
			defer cancel()

			ep, err := resolveEndpointID(ctx, c, id, endpointID)
			if err != nil {
				return err
			}

			if cfg.Output == "table" || cfg.Output == "" {
				fmt.Printf("Restarting stack %d (environment: %d)...\n", id, ep)
			}

			stack, err := c.Stacks().Restart(ctx, id, ep)
			if err != nil {
				return fmt.Errorf("failed to restart stack %d: %w", id, err)
			}

			return PrintFormatted(os.Stdout, cfg.Output, stack, func() error {
				fmt.Printf("Stack %s (ID: %d) successfully restarted.\n", stack.Name, stack.ID)
				return nil
			})
		},
	}

	cmd.Flags().IntVarP(&endpointID, "endpoint", "e", 0, "Environment(endpoint) ID (overrides default)")

	return cmd
}
