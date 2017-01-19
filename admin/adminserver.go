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
	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/golang/glog"
	"net/http"
)

const DEFAULT_ADMIN_HOST_PORT = "127.0.0.1:10000"

func Initialize() {
	if config.Cfg.AdminHostPort == "" {
		config.Cfg.AdminHostPort = DEFAULT_ADMIN_HOST_PORT
		glog.Infof("[adminserver] Admin Server host and port is not specified, using default: %s\n",
			DEFAULT_ADMIN_HOST_PORT)
	}

	glog.Infof("[adminserver] Initializing admin server on %s", config.Cfg.AdminHostPort)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	/*
	 * Setup the admin server HTTP routes.
	 */
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/config", GetConfig},
		&rest.Route{"GET", "/stats", GetStats},
		&rest.Route{"GET", "/stream", StreamEvents},
	)

	if err != nil {
		glog.Fatalln("An error occurred setting up the admin server routes, %s\n", err.Error())
	}

	api.SetApp(router)

	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))

	err = http.ListenAndServe(config.Cfg.AdminHostPort, nil)

	if err != nil {
		glog.Errorf("An error occurred starting up the admin server, %s\n", err.Error())
	}
}

func GetConfig(w rest.ResponseWriter, r *rest.Request) {
	glog.V(2).Infoln("[adminserver] /config requested")

	w.Header().Set("Content-Type", "text/json")
	w.WriteJson(config.Cfg)

	glog.V(2).Infoln("[adminserver] /config response sent")
}

type AdminStatsNode struct {
	HostPort string `json:"ipaddr"`
	Healthy  bool   `json:"healthy"`
	Queries  int    `json:"queries"`
}

type AdminStats struct {
	Nodes []AdminStatsNode `json:"nodes"`
}

func GetStats(w rest.ResponseWriter, r *rest.Request) {
	glog.V(2).Infoln("[adminserver] /stats requested")

	stats := AdminStats{}

	stats.Nodes = make([]AdminStatsNode, (1 + len(config.Cfg.Replicas)))

	// Add the master node statistics.
	stats.Nodes[0].HostPort = config.Cfg.Master.HostPort
	stats.Nodes[0].Queries = config.Cfg.Master.Stats.Queries
	stats.Nodes[0].Healthy = config.Cfg.Master.Healthy

	// Add the replica nodes statistics.
	for index, replica := range config.Cfg.Replicas {
		stats.Nodes[index+1].HostPort = replica.HostPort
		stats.Nodes[index+1].Queries = replica.Stats.Queries
		stats.Nodes[index+1].Healthy = replica.Healthy
	}

	w.Header().Set("Content-Type", "text/json")
	w.WriteJson(&stats)

	glog.V(2).Infoln("[adminserver] /stats response sent")
}
