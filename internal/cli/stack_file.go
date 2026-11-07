package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func newStackFileCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:     "file <stack-id>",
		Aliases: []string{"cat"},
		Short:   "Get stack compose file",
		Long:    "Retrieve the docker-compose or stack file content for a specified stack.",
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

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			content, err := c.Stacks().GetFile(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get stack file: %w", err)
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, []byte(content), 0600); err != nil {
					return fmt.Errorf("failed writing file %s: %w", outputFile, err)
				}
				fmt.Printf("Stack file saved to %s\n", outputFile)
				return nil
			}

			fmt.Print(content)
			if len(content) > 0 && content[len(content)-1] != '\n' {
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "out", "w", "", "Write output to file instead of stdout")

	return cmd
}
