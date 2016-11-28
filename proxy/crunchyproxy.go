package main

import (
	"github.com/crunchydata/crunchy-proxy/proxy/admin"
	"github.com/crunchydata/crunchy-proxy/proxy/config"
	"github.com/crunchydata/crunchy-proxy/proxy/proxy"
	"log"
)

var cfg config.Config

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("main starting...")

	//get the Config
	cfg = config.ReadConfig()
	go admin.StartHealthcheck(&cfg)
	cfg.SetupAdapters()
	cfg.PrintNodeInfo("after SetupAdapters")

	go admin.Initialize(&cfg)

	if cfg.Pool.Enabled {
		proxy.SetupPools(&cfg)
	}

	proxy.ListenAndServe(&cfg)

}
