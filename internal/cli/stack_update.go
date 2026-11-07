package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/portainer/portainer-cli/pkg/stacks"
	"github.com/spf13/cobra"
)

func newStackUpdateCmd() *cobra.Command {
	var (
		file       string
		endpointID int
		envFlags   []string
		envFile    string
		prune      bool
		pullImage  bool
	)

	cmd := &cobra.Command{
		Use:   "update <stack-id>",
		Short: "Update stack configuration or compose file",
		Long:  "Updates the docker-compose content and/or environment variables for a stack and redeploys it.",
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

			ep := endpointID
			if ep == 0 {
				ep = cfg.EndpointID
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 180*time.Second)
			defer cancel()

			payload := &stacks.UpdateStackPayload{
				Prune:                  prune,
				RepullImageAndRedeploy: pullImage,
			}

			if file != "" {
				var fileBytes []byte
				if file == "-" {
					fileBytes, err = io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed reading from stdin: %w", err)
					}
				} else {
					fileBytes, err = os.ReadFile(file)
					if err != nil {
						return fmt.Errorf("failed reading file '%s': %w", file, err)
					}
				}
				payload.StackFileContent = string(fileBytes)
			} else {
				// If no file provided, fetch existing file content so Portainer doesn't clear it
				content, err := c.Stacks().GetFile(ctx, id)
				if err == nil && content != "" {
					payload.StackFileContent = content
				}
			}

			if len(envFlags) > 0 || envFile != "" {
				envPairs, err := parseEnvPairs(envFlags, envFile)
				if err != nil {
					return err
				}
				payload.Env = envPairs
			} else {
				// Retain existing env if not explicitly updated
				existing, err := c.Stacks().Inspect(ctx, id)
				if err == nil && existing != nil {
					payload.Env = existing.Env
					if ep == 0 {
						ep = existing.EndpointID
					}
				}
			}

			stack, err := c.Stacks().Update(ctx, id, ep, payload)
			if err != nil {
				return fmt.Errorf("failed to update stack %d: %w", id, err)
			}

			return PrintFormatted(os.Stdout, cfg.Output, stack, func() error {
				fmt.Printf("Stack '%s' (ID: %d) successfully updated.\n", stack.Name, stack.ID)
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "New compose file path, or '-' for stdin")
	cmd.Flags().IntVarP(&endpointID, "endpoint", "e", 0, "Environment(endpoint) ID")
	cmd.Flags().StringArrayVar(&envFlags, "env", nil, "Environment variables (KEY=VAL, repeatable)")
	cmd.Flags().StringVar(&envFile, "env-file", "", "Path to .env file for stack variables")
	cmd.Flags().BoolVar(&prune, "prune", false, "Prune services no longer referenced (Swarm stacks)")
	cmd.Flags().BoolVar(&pullImage, "pull-image", false, "Force re-pulling images and redeploying")

	return cmd
}
