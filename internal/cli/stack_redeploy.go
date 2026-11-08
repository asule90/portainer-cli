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

func newStackRedeployCmd() *cobra.Command {
	var (
		endpointID int
		ref        string
		username   string
		password   string
		prune      bool
		pullImage  bool
	)

	cmd := &cobra.Command{
		Use:   "redeploy <stack-id>",
		Short: "Pull from Git and redeploy stack",
		Long:  "Triggers a git pull and redeployment for a git-backed stack.",
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

			ctx, cancel := context.WithTimeout(cmd.Context(), 180*time.Second)
			defer cancel()

			ep, err := resolveEndpointID(ctx, c, id, endpointID)
			if err != nil {
				return err
			}

			auth := username != "" || password != ""
			payload := &stacks.RedeployGitPayload{
				RepositoryReferenceName:  ref,
				RepositoryAuthentication: auth,
				RepositoryUsername:       username,
				RepositoryPassword:       password,
				Prune:                    prune,
				RepullImageAndRedeploy:   pullImage,
			}

			stack, err := c.Stacks().RedeployGit(ctx, id, ep, payload)
			if err != nil {
				return fmt.Errorf("failed to redeploy stack %d: %w", id, err)
			}

			return PrintFormatted(os.Stdout, cfg.Output, stack, func() error {
				fmt.Printf("Stack '%s' (ID: %d) successfully redeployed from git.\n", stack.Name, stack.ID)
				return nil
			})
		},
	}

	cmd.Flags().IntVarP(&endpointID, "endpoint", "e", 0, "Environment(endpoint) ID")
	cmd.Flags().StringVar(&ref, "ref", "", "Git branch, tag, or reference (leave empty for default)")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Git authentication username")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Git authentication password or token")
	cmd.Flags().BoolVar(&prune, "prune", false, "Prune removed services")
	cmd.Flags().BoolVar(&pullImage, "pull-image", true, "Force re-pulling image before redeploying")

	return cmd
}
