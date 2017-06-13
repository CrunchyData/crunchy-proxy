package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:     "stats [options]",
	Short:   "",
	Long:    "",
	Example: "",
	RunE:    runStats,
}

func init() {
	flags := statsCmd.Flags()

	flags.StringVarP(&host, "host", "", "localhost", "")
	flags.StringVarP(&host, "port", "", "8000", "")
}

func runStats(cmd *cobra.Command, args []string) error {
	fmt.Printf("Stats - host: %s, port: %s\n", host, port)
	return nil
}
