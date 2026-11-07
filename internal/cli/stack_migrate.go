package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/portainer/portainer-cli/pkg/stacks"
	"github.com/spf13/cobra"
)

func newStackMigrateCmd() *cobra.Command {
	var (
		sourceEndpointID int
		targetEndpointID int
		newName          string
		swarmID          string
	)

	cmd := &cobra.Command{
		Use:   "migrate <stack-id>",
		Short: "Migrate stack to another environment",
		Long:  "Relocates a stack to a different Portainer environment (endpoint).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid stack ID '%s': must be an integer", args[0])
			}

			if targetEndpointID <= 0 {
				return fmt.Errorf("--target-endpoint is required and must be > 0")
			}

			c, err := getClient()
			if err != nil {
				return err
			}

			srcEP := sourceEndpointID
			if srcEP == 0 {
				srcEP = cfg.EndpointID
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 180*time.Second)
			defer cancel()

			payload := &stacks.MigrateStackPayload{
				EndpointID: targetEndpointID,
				Name:       newName,
				SwarmID:    swarmID,
			}

			stack, err := c.Stacks().Migrate(ctx, id, srcEP, payload)
			if err != nil {
				return fmt.Errorf("failed to migrate stack %d: %w", id, err)
			}

			return PrintFormatted(os.Stdout, cfg.Output, stack, func() error {
				fmt.Printf("Stack '%s' (ID: %d) successfully migrated to endpoint %d.\n", stack.Name, stack.ID, stack.EndpointID)
				return nil
			})
		},
	}

	cmd.Flags().IntVar(&sourceEndpointID, "source-endpoint", 0, "Source environment(endpoint) ID")
	cmd.Flags().IntVarP(&targetEndpointID, "target-endpoint", "t", 0, "Target environment(endpoint) ID (required)")
	cmd.Flags().StringVarP(&newName, "name", "n", "", "New name for the stack on destination")
	cmd.Flags().StringVar(&swarmID, "swarm-id", "", "Target Swarm cluster ID")

	_ = cmd.MarkFlagRequired("target-endpoint")

	return cmd
}
