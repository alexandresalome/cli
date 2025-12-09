package commands

import (
	"github.com/spf13/cobra"

	"github.com/platformsh/cli/internal/config"
)

func newFleetCommand(cnf *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fleet [flags] [namespace]",
		Short: "Fleet command",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			arguments := []string{"multi", "--help"}

			c := makeLegacyCLIWrapper(cnf, cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
			if err := c.Exec(cmd.Context(), arguments...); err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}
