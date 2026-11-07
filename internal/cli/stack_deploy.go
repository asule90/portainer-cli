package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/portainer/portainer-cli/pkg/stacks"
	"github.com/spf13/cobra"
)

func newStackDeployCmd() *cobra.Command {
	var (
		name        string
		file        string
		endpointID  int
		envFlags    []string
		envFile     string
		isSwarm     bool
		swarmID     string
		fromAppTmpl bool
	)

	cmd := &cobra.Command{
		Use:     "deploy",
		Short:   "Deploy a stack from compose file or text",
		Long:    "Deploys a new Docker Compose (standalone) or Swarm stack from a file or stdin.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("--name is required")
			}
			if strings.TrimSpace(file) == "" {
				return fmt.Errorf("--file is required (path or - for stdin)")
			}

			var fileContent []byte
			var err error
			if file == "-" {
				fileContent, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed reading compose content from stdin: %w", err)
				}
			} else {
				fileContent, err = os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("failed reading compose file '%s': %w", file, err)
				}
			}

			if len(fileContent) == 0 {
				return fmt.Errorf("compose file content cannot be empty")
			}

			envPairs, err := parseEnvPairs(envFlags, envFile)
			if err != nil {
				return err
			}

			c, err := getClient()
			if err != nil {
				return err
			}

			ep := endpointID
			if ep == 0 {
				ep = cfg.EndpointID
			}
			if ep == 0 {
				return fmt.Errorf("endpoint ID is required (set via --endpoint or PORTAINER_ENDPOINT_ID)")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 120*time.Second)
			defer cancel()

			var stack *stacks.Stack
			if isSwarm {
				payload := &stacks.CreateSwarmStringPayload{
					Name:             name,
					StackFileContent: string(fileContent),
					Env:              envPairs,
					FromAppTemplate:  fromAppTmpl,
					SwarmID:          swarmID,
				}
				stack, err = c.Stacks().CreateSwarmFromString(ctx, ep, payload)
			} else {
				payload := &stacks.CreateComposeStringPayload{
					Name:             name,
					StackFileContent: string(fileContent),
					Env:              envPairs,
					FromAppTemplate:  fromAppTmpl,
				}
				stack, err = c.Stacks().CreateComposeFromString(ctx, ep, payload)
			}

			if err != nil {
				return fmt.Errorf("failed to deploy stack: %w", err)
			}

			return PrintFormatted(os.Stdout, cfg.Output, stack, func() error {
				fmt.Printf("Stack '%s' (ID: %d) successfully deployed on endpoint %d.\n", stack.Name, stack.ID, stack.EndpointID)
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Stack name (required)")
	cmd.Flags().StringVarP(&file, "file", "f", "", "Compose file path, or '-' for stdin (required)")
	cmd.Flags().IntVarP(&endpointID, "endpoint", "e", 0, "Environment(endpoint) ID")
	cmd.Flags().StringArrayVar(&envFlags, "env", nil, "Environment variables (KEY=VAL, repeatable)")
	cmd.Flags().StringVar(&envFile, "env-file", "", "Path to .env file for stack variables")
	cmd.Flags().BoolVar(&isSwarm, "swarm", false, "Deploy as Docker Swarm stack instead of Standalone Compose")
	cmd.Flags().StringVar(&swarmID, "swarm-id", "", "Target Swarm cluster ID (for Swarm stacks)")
	cmd.Flags().BoolVar(&fromAppTmpl, "from-template", false, "Mark stack as created from an app template")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
