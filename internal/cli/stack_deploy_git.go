package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/portainer/portainer-cli/pkg/stacks"
	"github.com/spf13/cobra"
)

func newStackDeployGitCmd() *cobra.Command {
	var (
		name         string
		repoURL      string
		ref          string
		composeFile  string
		endpointID   int
		username     string
		password     string
		skipTLS      bool
		isSwarm      bool
		swarmID      string
		envFlags     []string
		envFile      string
		autoInterval string
		webhook      string
	)

	cmd := &cobra.Command{
		Use:   "deploy-git",
		Short: "Deploy a stack from a Git repository",
		Long:  "Deploys a new Docker Compose or Swarm stack pulled from a Git repository (GitOps).",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("--name is required")
			}
			if strings.TrimSpace(repoURL) == "" {
				return fmt.Errorf("--repo is required")
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

			ctx, cancel := context.WithTimeout(cmd.Context(), 180*time.Second)
			defer cancel()

			var autoUpdate *stacks.AutoUpdateSettings
			if autoInterval != "" || webhook != "" {
				autoUpdate = &stacks.AutoUpdateSettings{
					Interval: autoInterval,
					Webhook:  webhook,
				}
			}

			auth := username != "" || password != ""

			var stack *stacks.Stack
			if isSwarm {
				payload := &stacks.CreateSwarmGitPayload{
					Name:                     name,
					RepositoryURL:            repoURL,
					RepositoryReferenceName:  ref,
					ComposeFile:              composeFile,
					RepositoryAuthentication: auth,
					RepositoryUsername:       username,
					RepositoryPassword:       password,
					TLSSkipVerify:            skipTLS,
					Env:                      envPairs,
					AutoUpdate:               autoUpdate,
					SwarmID:                  swarmID,
				}
				stack, err = c.Stacks().CreateSwarmFromGit(ctx, ep, payload)
			} else {
				payload := &stacks.CreateComposeGitPayload{
					Name:                     name,
					RepositoryURL:            repoURL,
					RepositoryReferenceName:  ref,
					ComposeFile:              composeFile,
					RepositoryAuthentication: auth,
					RepositoryUsername:       username,
					RepositoryPassword:       password,
					TLSSkipVerify:            skipTLS,
					Env:                      envPairs,
					AutoUpdate:               autoUpdate,
				}
				stack, err = c.Stacks().CreateComposeFromGit(ctx, ep, payload)
			}

			if err != nil {
				return fmt.Errorf("failed to deploy stack from git: %w", err)
			}

			return PrintFormatted(os.Stdout, cfg.Output, stack, func() error {
				fmt.Printf("Git-backed stack '%s' (ID: %d) successfully deployed on endpoint %d.\n", stack.Name, stack.ID, stack.EndpointID)
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Stack name (required)")
	cmd.Flags().StringVarP(&repoURL, "repo", "r", "", "Git repository URL (required)")
	cmd.Flags().StringVar(&ref, "ref", "refs/heads/main", "Git branch, tag, or reference")
	cmd.Flags().StringVarP(&composeFile, "compose-file", "c", "docker-compose.yml", "Path to compose file in repository")
	cmd.Flags().IntVarP(&endpointID, "endpoint", "e", 0, "Environment(endpoint) ID")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Git authentication username")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Git authentication password or token")
	cmd.Flags().BoolVar(&skipTLS, "skip-tls", false, "Skip Git SSL verification")
	cmd.Flags().BoolVar(&isSwarm, "swarm", false, "Deploy as Docker Swarm stack")
	cmd.Flags().StringVar(&swarmID, "swarm-id", "", "Target Swarm cluster ID")
	cmd.Flags().StringArrayVar(&envFlags, "env", nil, "Environment variables (KEY=VAL, repeatable)")
	cmd.Flags().StringVar(&envFile, "env-file", "", "Path to .env file for stack variables")
	cmd.Flags().StringVar(&autoInterval, "auto-update-interval", "", "GitOps poll interval (e.g. 5m, 1h)")
	cmd.Flags().StringVar(&webhook, "webhook", "", "GitOps webhook trigger URL/ID")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}
