package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Main() {
	if err := Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Failed running %q\n", os.Args[1])
		os.Exit(1)
	}
}

var crunchyproxyCmd = &cobra.Command{
	Use:          "crunchy-proxy",
	Short:        "A simple connection pool based SQL routing proxy",
	SilenceUsage: true,
}

func init() {
	cobra.EnableCommandSorting = false

	crunchyproxyCmd.AddCommand(
		startCmd,
		stopCmd,
		nodeCmd,
		statsCmd,
		healthCmd,
		versionCmd,
	)
}

func Run(args []string) error {
	crunchyproxyCmd.SetArgs(args)
	return crunchyproxyCmd.Execute()
}
