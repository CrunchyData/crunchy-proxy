package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "",
	Long:    "",
	Example: "",
	RunE:    runVersion,
}

func init() {
	crunchyproxyCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Println("Version!")

	return nil
}
