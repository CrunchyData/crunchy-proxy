package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "show version information for instance of proxy",
	Long:    "",
	Example: "",
	RunE:    runVersion,
}

func init() {
	flags := versionCmd.Flags()

	stringFlag(flags, &host, FlagAdminHost)
	stringFlag(flags, &port, FlagAdminPort)
	stringFlag(flags, &format, FlagOutputFormat)
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Println("Not Implemented")

	return nil
}
