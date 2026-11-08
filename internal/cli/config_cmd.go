package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/portainer/portainer-cli/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or modify CLI configuration",
		Long:  "Inspect active settings or set keys in ~/.config/portainer-cli/config.json.",
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigPathCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display active configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := config.ConfigFilePath()
			maskedKey := "-"
			if cfg.APIKey != "" {
				if len(cfg.APIKey) > 10 {
					maskedKey = cfg.APIKey[:4] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
				} else {
					maskedKey = "***"
				}
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "Config File:\t%s\n", path)
			fmt.Fprintf(w, "URL:\t%s\n", cfg.URL)
			fmt.Fprintf(w, "API Key:\t%s\n", maskedKey)
			fmt.Fprintf(w, "Insecure TLS:\t%t\n", cfg.Insecure)
			fmt.Fprintf(w, "Default Endpoint:\t%d\n", cfg.EndpointID)
			fmt.Fprintf(w, "Output Format:\t%s\n", cfg.Output)
			return w.Flush()
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration property",
		Long: `Sets a property in ~/.config/portainer-cli/config.json.
Available keys: url, api-key, insecure, endpoint-id, output`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			val := args[1]
			if err := config.UpdateFileKey(key, val); err != nil {
				return err
			}
			path, _ := config.ConfigFilePath()
			fmt.Printf("Property '%s' updated in %s.\n", key, path)
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the path to the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.ConfigFilePath()
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		},
	}
}
