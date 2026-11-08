package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/portainer/portainer-cli/internal/config"
	"github.com/portainer/portainer-cli/pkg/client"
	"github.com/spf13/cobra"
)

type userResponse struct {
	ID       int    `json:"Id"`
	Username string `json:"Username"`
	Role     int    `json:"Role"`
}

func newLoginCmd() *cobra.Command {
	var (
		urlFlag      string
		apiKeyFlag   string
		insecureFlag bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Portainer and persist credentials",
		Long: `Authenticate with your Portainer instance using an API key.
Once logged in, credentials are saved to ~/.config/portainer-cli/config.json
so you do not need to enter or paste the API key on subsequent commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetURL := urlFlag
			if targetURL == "" {
				targetURL = cfg.URL
			}
			if targetURL == "" {
				targetURL = "https://localhost:9443"
			}

			apiKey := strings.TrimSpace(apiKeyFlag)
			if apiKey == "" {
				// Prompt user for API key
				fmt.Printf("Portainer URL [%s]: ", targetURL)
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					inputURL := strings.TrimSpace(scanner.Text())
					if inputURL != "" {
						targetURL = inputURL
					}
				}

				fmt.Print("Enter Portainer API Key: ")
				if scanner.Scan() {
					apiKey = strings.TrimSpace(scanner.Text())
				}
			}

			if apiKey == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			// Validate credentials against Portainer /api/users/me
			opts := []client.Option{
				client.WithBaseURL(targetURL),
				client.WithAPIKey(apiKey),
				client.WithInsecureTLS(insecureFlag),
			}

			c, err := client.NewClient(opts...)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()

			var user userResponse
			if err := c.Get(ctx, "/users/me", nil, &user); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			// Save to persistent file
			fc, err := config.LoadFileConfig()
			if err != nil {
				fc = &config.FileConfig{}
			}
			fc.URL = targetURL
			fc.APIKey = apiKey
			fc.Insecure = &insecureFlag

			if err := config.SaveFileConfig(fc); err != nil {
				return fmt.Errorf("failed saving config: %w", err)
			}

			path, _ := config.ConfigFilePath()
			fmt.Printf("Successfully authenticated as '%s' at %s.\n", user.Username, targetURL)
			fmt.Printf("Credentials saved to %s\n", path)
			fmt.Println("You can now run commands (e.g. 'portainer-cli stack list') from anywhere without passing --api-key.")
			return nil
		},
	}

	cmd.Flags().StringVar(&urlFlag, "url", "", "Portainer base URL (default: https://localhost:9443)")
	cmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "Portainer API Key")
	cmd.Flags().BoolVar(&insecureFlag, "insecure", true, "Skip TLS verification")

	return cmd
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear saved Portainer credentials",
		Long:  "Removes saved credentials from ~/.config/portainer-cli/config.json.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.DeleteFileConfig(); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}
			path, _ := config.ConfigFilePath()
			fmt.Printf("Successfully logged out. Removed %s\n", path)
			return nil
		},
	}
}
