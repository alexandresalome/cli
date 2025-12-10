package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/fleet"
)

func startGui(manager *fleet.FleetManager) {
	gui := fleet.New(manager)
	if err := gui.Start(); err != nil {
		exitWithError(err)
	}
}

func startDataOutput(manager *fleet.FleetManager) {

	// Organizations
	orgs, err := manager.Organization.ListAll()
	if err != nil {
		exitWithError(err)
	}
	for _, org := range orgs {
		fmt.Println(org.ToJson())
	}

	// Projects
	projects := manager.Project.SubscribeAll(context.Background())
	for project := range projects {
		fmt.Println(project.ToJson())
	}
}

func newFleetCommand(cnf *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fleet [flags]",
		Short: "Fleet command",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			manager := fleet.NewFleetManager(cnf, cmd)

			isAuthenticated, _, err := manager.Authentication()
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Internal error during the authentication:")
				exitWithError(err)
			} else if !isAuthenticated {
				fmt.Fprintln(cmd.OutOrStdout(), "You are not logged into Upsun.\nPlease run `upsun login`.")
				os.Exit(1)
			}

			data, _ := cmd.Flags().GetBool("data")
			if !data {
				startGui(manager)
			} else {
				startDataOutput(manager)
			}
		},
	}

	cmd.Flags().BoolP("data", "d", false, "Output fleet data")

	return cmd
}
