package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/portainer/portainer-cli/pkg/stacks"
	"github.com/spf13/cobra"
)

func newStackListCmd() *cobra.Command {
	var (
		endpointFilter int
		swarmFilter    string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List stacks",
		Long:    "List all stacks across all or specific environments.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return err
			}

			filter := &stacks.StackFilter{}
			if endpointFilter > 0 {
				filter.EndpointID = endpointFilter
			} else if cfg.EndpointID > 0 {
				filter.EndpointID = cfg.EndpointID
			}
			if swarmFilter != "" {
				filter.SwarmID = swarmFilter
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			stackList, err := c.Stacks().List(ctx, filter)
			if err != nil {
				return fmt.Errorf("failed to list stacks: %w", err)
			}

			return PrintFormatted(os.Stdout, cfg.Output, stackList, func() error {
				if len(stackList) == 0 {
					fmt.Println("No stacks found.")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 4, 3, ' ', 0)
				fmt.Fprintln(w, "ID\tNAME\tTYPE\tENDPOINT\tSTATUS\tCREATED BY\tUPDATED AT")
				for _, s := range stackList {
					updated := "-"
					if s.UpdateDate > 0 {
						updated = s.UpdatedAt().Format("2006-01-02 15:04")
					}
					fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%s\t%s\t%s\n",
						s.ID,
						s.Name,
						s.Type.String(),
						s.EndpointID,
						s.Status.String(),
						s.CreatedBy,
						updated,
					)
				}
				return w.Flush()
			})
		},
	}

	cmd.Flags().IntVarP(&endpointFilter, "endpoint", "e", 0, "Filter by environment(endpoint) ID")
	cmd.Flags().StringVar(&swarmFilter, "swarm", "", "Filter by Swarm cluster ID")

	return cmd
}
