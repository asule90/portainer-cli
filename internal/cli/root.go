package cli

import (
	"errors"
	"fmt"

	"github.com/portainer/portainer-cli/internal/config"
	"github.com/portainer/portainer-cli/pkg/client"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config

	rootCmd = &cobra.Command{
		Use:   "portainer-cli",
		Short: "portainer-cli is a CLI client for the Portainer API (v2.39.1)",
		Long: `A high-performance command line interface for interacting with Portainer CE / EE,
built according to the official Portainer API 2.39.1 specification.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Execute runs the root CLI command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cfg = config.Load()

	rootCmd.PersistentFlags().StringVar(&cfg.URL, "url", cfg.URL, "Portainer instance base URL (env: PORTAINER_URL)")
	rootCmd.PersistentFlags().StringVar(&cfg.APIKey, "api-key", cfg.APIKey, "Portainer API Key (env: PORTAINER_API_KEY, X_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&cfg.JWTToken, "jwt", cfg.JWTToken, "Portainer JWT authentication token (env: PORTAINER_JWT)")
	rootCmd.PersistentFlags().BoolVar(&cfg.Insecure, "insecure", cfg.Insecure, "Skip SSL/TLS certificate verification (env: PORTAINER_INSECURE)")
	rootCmd.PersistentFlags().IntVar(&cfg.EndpointID, "endpoint-id", cfg.EndpointID, "Target environment / endpoint ID (env: PORTAINER_ENDPOINT_ID)")
	rootCmd.PersistentFlags().StringVarP(&cfg.Output, "output", "o", cfg.Output, "Output format: table, json, yaml (env: PORTAINER_OUTPUT)")

	rootCmd.AddCommand(newStackCmd())
}

// getClient creates an initialized Portainer API client using current configuration.
func getClient() (*client.Client, error) {
	if cfg.URL == "" {
		return nil, errors.New("portainer URL is required (specify via --url or PORTAINER_URL)")
	}
	if cfg.APIKey == "" && cfg.JWTToken == "" {
		return nil, errors.New("authentication required: provide --api-key (or X_API_KEY / PORTAINER_API_KEY env) or --jwt")
	}

	opts := []client.Option{
		client.WithBaseURL(cfg.URL),
		client.WithInsecureTLS(cfg.Insecure),
	}

	if cfg.APIKey != "" {
		opts = append(opts, client.WithAPIKey(cfg.APIKey))
	} else if cfg.JWTToken != "" {
		opts = append(opts, client.WithJWT(cfg.JWTToken))
	}

	c, err := client.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	return c, nil
}
