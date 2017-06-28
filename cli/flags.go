package cli

import (
	"github.com/spf13/pflag"
)

var host string
var port string

var format string

type flagInfoString struct {
	Name        string
	Shorthand   string
	Description string
	Default     string
}

type flagInfoBool struct {
	Name        string
	Shorthand   string
	Description string
	Default     bool
}

var (
	FlagAdminHost = flagInfoString{
		Name:        "host",
		Description: "proxy admin server address",
		Default:     "localhost",
	}

	FlagAdminPort = flagInfoString{
		Name:        "port",
		Description: "proxy admin server port",
		Default:     "8000",
	}

	FlagOutputFormat = flagInfoString{
		Name:        "format",
		Description: "the output format",
		Default:     "plain",
	}

	FlagConfigPath = flagInfoString{
		Name:        "config",
		Shorthand:   "c",
		Description: "path to configuration file",
	}

	FlagLogLevel = flagInfoString{
		Name:        "log-level",
		Description: "logging level",
		Default:     "info",
	}

	FlagBackground = flagInfoBool{
		Name:        "background",
		Description: "run process in background",
		Default:     false,
	}
)

func stringFlag(f *pflag.FlagSet, valPtr *string, flagInfo flagInfoString) {
	f.StringVarP(valPtr,
		flagInfo.Name,
		flagInfo.Shorthand,
		flagInfo.Default,
		flagInfo.Description)
}

func boolFlag(f *pflag.FlagSet, valPtr *bool, flagInfo flagInfoBool) {
	f.BoolVarP(valPtr,
		flagInfo.Name,
		flagInfo.Shorthand,
		flagInfo.Default,
		flagInfo.Description)
}
