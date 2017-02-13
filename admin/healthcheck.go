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
	"database/sql"
	"fmt"
	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/crunchydata/crunchy-proxy/proxy"
	"github.com/golang/glog"
	_ "github.com/lib/pq"
	"net"
	"sync"
	"time"
)

const DEFAULT_HEALTHCHECK_QUERY = "SELECT now();"
const DEFAULT_HEALTHCHECK_DELAY = 10

func StartHealthcheck() {

	var result bool
	var mutex = &sync.Mutex{}
	var event ProxyEvent

	// If a healthcheck query is not provided, then use the default.
	if config.Cfg.Healthcheck.Query == "" {
		config.Cfg.Healthcheck.Query = DEFAULT_HEALTHCHECK_QUERY
		glog.Infof("[hc] Healthcheck query is not specified, using default: %s\n",
			config.Cfg.Healthcheck.Query)
	}

	// If a healthcheck delay is not provided, then use the default.
	if config.Cfg.Healthcheck.Delay == 0 {
		config.Cfg.Healthcheck.Delay = DEFAULT_HEALTHCHECK_DELAY
		glog.Infof("[hc] Healthcheck delay is not specified, using default: %d\n",
			config.Cfg.Healthcheck.Delay)
	}

	// Start healthcheck of all nodes.
	for true {

		// Check master node.
		glog.V(2).Info("[hc] Checking Master")
		result = healthcheckQuery(config.Cfg.Master)

		event = ProxyEvent{
			Name:    "hc",
			Message: fmt.Sprintf("master is healthy: %t", result),
		}

		for j := range EventChannel {
			EventChannel[j] <- event
		}

		mutex.Lock()

		if !config.Cfg.Master.Healthy && result == true {
			glog.V(2).Info("[hc] Master going healthy after being down")
			glog.V(2).Info("[hc] Rebuilding connection pool for master")
			proxy.SetupPoolForNode(&config.Cfg.Master)
		}
		config.Cfg.Master.Healthy = result
		mutex.Unlock()

		// Check replica nodes.
		for i := range config.Cfg.Replicas {
			glog.V(2).Infof("[hc] Checking Replica %d\n", i)
			result = healthcheckQuery(config.Cfg.Replicas[i])

			event = ProxyEvent{
				Name:    "hc",
				Message: fmt.Sprintf("replica [%d] is healthy: %t", i, result),
			}

			for j := range EventChannel {
				EventChannel[j] <- event
			}

			mutex.Lock()
			if !config.Cfg.Replicas[i].Healthy && result == true {
				glog.V(2).Info("[hc] Replica going healthy after being down")
				glog.V(2).Info("[hc] Rebuilding connection pool for replica")
				proxy.SetupPoolForNode(&config.Cfg.Replicas[i])
			}
			config.Cfg.Replicas[i].Healthy = result
			mutex.Unlock()
		}

		// Wait specified delay period before checking again.
		time.Sleep(time.Duration(config.Cfg.Healthcheck.Delay) * time.Second)
	}
}

func healthcheckQuery(node config.Node) bool {
	connection, err := getDBConnection(node)

	defer func() {
		if connection != nil {
			connection.Close()
		}
	}()

	if err != nil {
		glog.Errorln("[hc] healthcheck failed: error: " + err.Error())
		return false
	}

	rows, err := connection.Query(config.Cfg.Healthcheck.Query)

	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	if err != nil {
		glog.Errorln("[hc] failed: error: " + err.Error())
		return false
	}

	return true
}

func getDBConnection(node config.Node) (*sql.DB, error) {
	host, port, _ := net.SplitHostPort(node.HostPort)

	connectionString := fmt.Sprintf("host=%s port=%s ", host, port)
	connectionString += fmt.Sprintf(" user=%s", config.Cfg.Credentials.Username)
	connectionString += fmt.Sprintf(" database=%s", config.Cfg.Credentials.Database)

	connectionString += fmt.Sprintf(" sslmode=%s", config.Cfg.Credentials.SSL.SSLMode)
	connectionString += " application_name=proxy_healthcheck"

	if config.Cfg.Credentials.Password != "" {
		connectionString += fmt.Sprintf(" password=%s", config.Cfg.Credentials.Password)
	}

	if config.Cfg.Credentials.SSL.Enable {
		connectionString += fmt.Sprintf(" sslcert=%s", config.Cfg.Credentials.SSL.SSLCert)
		connectionString += fmt.Sprintf(" sslkey=%s", config.Cfg.Credentials.SSL.SSLKey)
		connectionString += fmt.Sprintf(" sslrootcert=%s", config.Cfg.Credentials.SSL.SSLRootCA)
	}

	/* Build connection string. */
	for key, value := range config.Cfg.Credentials.Options {
		connectionString += fmt.Sprintf(" %s=%s", key, value)
	}

	glog.V(2).Infof("[hc] Opening connection with parameters: %s\n", connectionString)

	dbConn, err := sql.Open("postgres", connectionString)

	if err != nil {
		glog.Errorf("[hc] Error creating connection : %s\n", err.Error())
	}

	return dbConn, err
}
