package cli

import (
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/crunchydata/crunchy-proxy/server"
	"github.com/crunchydata/crunchy-proxy/util/log"
)

var background bool
var configPath string
var logLevel string

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "start a proxy instance",
	Long:    "",
	Example: "",
	RunE:    runStart,
}

func init() {
	startCmd.Flags().BoolVarP(&background, "background", "b", false, "")
	startCmd.Flags().StringVarP(&configPath, "config", "c", "", "")
	startCmd.Flags().StringVarP(&logLevel, "log-level", "", "info", "")
}

func runStart(cmd *cobra.Command, args []string) error {
	if background {
		args = make([]string, 0, len(os.Args))

		for _, arg := range os.Args {
			if strings.HasPrefix(arg, "--background") {
				continue
			}
			args = append(args, arg)
		}

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Start()
	}

	log.SetLevel(logLevel)

	if configPath != "" {
		config.SetConfigPath(configPath)
	}

	config.ReadConfig()

	s := server.NewServer()

	s.Start()

	return nil
}
