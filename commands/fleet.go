package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/fleet"
)

func newFleetCommand(cnf *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fleet [flags] [namespace]",
		Short: "Fleet command",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			manager := fleet.NewFleetManager(cnf, cmd)

			// 1. Verify that the user is authenticated
			isAuthenticated, email, err := manager.Authentication()
			if err != nil {
				exitWithError(err)
			}

			if !isAuthenticated {
				fmt.Fprintln(cmd.OutOrStdout(), "You are not logged in to Upsun. Please log in to continue.")
				os.Exit(1)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Authenticated as %s\n", email)
		},
	}

	return cmd
}
