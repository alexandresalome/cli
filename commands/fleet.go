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

			fmt.Fprint(cmd.OutOrStdout(), "Authentication...")

			// 1. Verify that the user is authenticated
			isAuthenticated, email, err := manager.Authentication()
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), " failed (internal error)")
				exitWithError(err)
			}

			if !isAuthenticated {
				fmt.Fprintln(cmd.OutOrStdout(), " failed (not authenticated)")
				fmt.Fprintln(cmd.OutOrStdout(), "\n- You are not logged into Upsun.\n- Please run `upsun login`.")
				os.Exit(1)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			fmt.Fprintf(cmd.OutOrStdout(), "- Current user: %s\n", email)

			// 2. Fetch the list of projects
			fmt.Fprint(cmd.OutOrStdout(), "Fetching projects... ")

			projects, err := manager.Project.List()
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "failed")
				exitWithError(err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			fmt.Fprintf(cmd.OutOrStdout(), "- Found %d projects\n", len(projects))

			// 3. Display the list of projects
			for _, project := range projects {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s (ID: %s)\n", project.ProjectTitle, project.ProjectID)
			}
		},
	}

	return cmd
}
