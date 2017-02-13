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
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//get the Config
	config.ReadConfig()

	glog.Infoln("main starting...")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		glog.Infoln(sig)
		glog.Infoln("caught signal, cleaning up and exiting...")
		os.Exit(0)
	}()

	proxy.SetupPools()

	go admin.StartHealthcheck()

	config.Cfg.SetupAdapters()
	config.Cfg.PrintNodeInfo()

	go admin.Initialize()

	proxy.ListenAndServe()

}
