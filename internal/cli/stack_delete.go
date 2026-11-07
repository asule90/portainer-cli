package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func newStackDeleteCmd() *cobra.Command {
	var (
		endpointID int
		external   bool
	)

	cmd := &cobra.Command{
		Use:     "delete <stack-id>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a stack",
		Long:    "Removes a stack and its associated containers and resources from the environment.",
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

			ep := endpointID
			if ep == 0 {
				ep = cfg.EndpointID
			}

			// If endpoint not specified, try inspecting the stack to get its endpoint ID
			if ep == 0 {
				ctxLookup, cancelLookup := context.WithTimeout(cmd.Context(), 10*time.Second)
				if st, err := c.Stacks().Inspect(ctxLookup, id); err == nil && st != nil {
					ep = st.EndpointID
				}
				cancelLookup()
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			if err := c.Stacks().Delete(ctx, id, ep, external); err != nil {
				return fmt.Errorf("failed to delete stack %d: %w", id, err)
			}

			fmt.Printf("Stack %d successfully removed.\n", id)
			return nil
		},
	}

	cmd.Flags().IntVarP(&endpointID, "endpoint", "e", 0, "Environment(endpoint) ID")
	cmd.Flags().BoolVar(&external, "external", false, "Delete external Swarm stack")

	return cmd
}
