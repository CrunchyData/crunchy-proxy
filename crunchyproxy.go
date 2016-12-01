/*
Copyright 2016 Crunchy Data Solutions, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"github.com/crunchydata/crunchy-proxy/admin"
	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/crunchydata/crunchy-proxy/proxy"
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
