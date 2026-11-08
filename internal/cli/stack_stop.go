package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func newStackStopCmd() *cobra.Command {
	var endpointID int

	cmd := &cobra.Command{
		Use:   "stop <stack-id>",
		Short: "Stop a running stack",
		Long:  "Stops all running services and containers belonging to the stack.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid stack ID '%s': must be an integer", args[0])
			}

			c, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			ep, err := resolveEndpointID(ctx, c, id, endpointID)
			if err != nil {
				return err
			}

			stack, err := c.Stacks().Stop(ctx, id, ep)
			if err != nil {
				return fmt.Errorf("failed to stop stack %d: %w", id, err)
			}

			return PrintFormatted(os.Stdout, cfg.Output, stack, func() error {
				fmt.Printf("Stack %s (ID: %d) successfully stopped.\n", stack.Name, stack.ID)
				return nil
			})
		},
	}

	cmd.Flags().IntVarP(&endpointID, "endpoint", "e", 0, "Environment(endpoint) ID (overrides default)")

	return cmd
}
