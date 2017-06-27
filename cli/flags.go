package cli

import (
	"github.com/spf13/pflag"
)

var host string
var port string

var format string

type FlagInfo struct {
	Name        string
	Shorthand   string
	Description string
	Default     string
}

var (
	FlagAdminHost = FlagInfo{
		Name:        "host",
		Description: "proxy admin server address",
		Default:     "localhost",
	}

	FlagAdminPort = FlagInfo{
		Name:        "port",
		Description: "proxy admin server port",
		Default:     "8000",
	}

	FlagOutputFormat = FlagInfo{
		Name:        "format",
		Description: "the output format",
		Default:     "plain",
	}
)

func stringFlag(f *pflag.FlagSet, valPtr *string, flagInfo FlagInfo) {
	f.StringVarP(valPtr,
		flagInfo.Name,
		flagInfo.Shorthand,
		flagInfo.Default,
		flagInfo.Description)
}
