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

package admin

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/crunchydata/crunchy-proxy/proxy/config"
	"log"
	"net/http"
	"strconv"
)

const DEFAULT_ADMIN_IPADDR = ":10000"

var globalconfig *config.Config

func Initialize(config *config.Config) {

	var ipaddr = DEFAULT_ADMIN_IPADDR
	log.Println("config.AdminIPAddr is [" + config.AdminIPAddr + "]")
	if config.AdminIPAddr != "" {
		ipaddr = config.AdminIPAddr
	}
	log.Println("adminserver: initializing on " + ipaddr)
	globalconfig = config

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/config", GetConfig},
		&rest.Route{"GET", "/stats", GetStats},
		&rest.Route{"GET", "/stream", StreamEvents},
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))

	log.Fatal(http.ListenAndServe(ipaddr, nil))
}

func GetConfig(w rest.ResponseWriter, r *rest.Request) {
	log.Println("adminserver: GetConfig called")

	w.Header().Set("Content-Type", "text/json")
	w.WriteJson(globalconfig)
	log.Println("adminserver: GetConfig report written")
}
func GetStats(w rest.ResponseWriter, r *rest.Request) {
	log.Println("adminserver: GetStats called")

	var somecontent = "something goes here"
	var thing http.ResponseWriter
	thing = w.(http.ResponseWriter)

	thing.Header().Set("Content-Type", "text/html")
	thing.Header().Set("Content-Length", strconv.Itoa(len(somecontent)))
	thing.Write([]byte(somecontent))
	log.Println("adminserver: GetStatus report written")
}
