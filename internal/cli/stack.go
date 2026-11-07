package cli

import (
	"github.com/spf13/cobra"
)

func newStackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stack",
		Aliases: []string{"stacks"},
		Short:   "Manage Portainer stacks (compose, swarm, git)",
		Long:    "Deploy, inspect, start, stop, update, redeploy, and remove stacks in Portainer environments.",
	}

	cmd.AddCommand(newStackListCmd())
	cmd.AddCommand(newStackInspectCmd())
	cmd.AddCommand(newStackFileCmd())
	cmd.AddCommand(newStackStartCmd())
	cmd.AddCommand(newStackStopCmd())
	cmd.AddCommand(newStackDeployCmd())
	cmd.AddCommand(newStackDeployGitCmd())
	cmd.AddCommand(newStackUpdateCmd())
	cmd.AddCommand(newStackRedeployCmd())
	cmd.AddCommand(newStackDeleteCmd())
	cmd.AddCommand(newStackMigrateCmd())

	return cmd
}
