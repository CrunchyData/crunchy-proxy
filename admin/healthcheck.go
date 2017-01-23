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
	"strings"
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
		result = healthcheckQuery(config.Cfg.Credentials, config.Cfg.Healthcheck.Query, config.Cfg.Master)

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
			result = healthcheckQuery(config.Cfg.Credentials, config.Cfg.Healthcheck.Query, config.Cfg.Replicas[i])

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

func healthcheckQuery(cred config.PGCredentials, query string, node config.Node) bool {

	var conn *sql.DB
	var rows *sql.Rows
	var err error
	var hostport = strings.Split(node.HostPort, ":")
	var dbHost = hostport[0]
	var dbUser = cred.Username
	var dbPassword = cred.Password
	var dbPort = hostport[1]
	var database = cred.Database

	conn, err = getDBConnection(dbHost, dbUser, dbPort, database, dbPassword)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	if err != nil {
		glog.Errorln("[hc] healthcheck failed: error: " + err.Error())
		return false
	}

	rows, err = conn.Query(query)
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

func getDBConnection(dbHost string, dbUser string, dbPort string, database string, dbPassword string) (*sql.DB, error) {

	var dbConn *sql.DB
	var err error
	var connectionString string

	if dbPassword == "" {
		connectionString = fmt.Sprintf("host=%s port=%s user=%s database=%s sslmode=disable",
			dbHost, dbPort, dbUser, database)
	} else {
		connectionString = fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, database)
	}

	glog.V(5).Infof("[hc] Opening connection with parameters: %s\n", connectionString)

	dbConn, err = sql.Open("postgres", connectionString)

	if err != nil {
		glog.Errorf("[hc] Error creating connection : %s\n", err.Error())
	}

	return dbConn, err
}
