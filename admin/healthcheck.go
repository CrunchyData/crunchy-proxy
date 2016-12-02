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
	"github.com/golang/glog"
	_ "github.com/lib/pq"
	"strings"
	"sync"
	"time"
)

func StartHealthcheck(c *config.Config) {

	if c.Healthcheck.Query == "" {
		c.Healthcheck.Query = "select now()"
	}
	if c.Healthcheck.Delay == 0 {
		c.Healthcheck.Delay = 10
	}
	//log.Printf("[hc]: delay: %d query:%s\n", c.Healthcheck.Delay, c.Healthcheck.Query)
	var result bool
	var mutex = &sync.Mutex{}
	var event ProxyEvent

	for true {
		result = HealthcheckQuery(c.Credentials, c.Healthcheck, c.Master)
		//log.Printf("[hc] master: %t ", result)
		event = ProxyEvent{
			Name:    "hc",
			Message: fmt.Sprintf("master is %t", result),
		}
		for j := range EventChannel {
			EventChannel[j] <- event
		}

		mutex.Lock()
		c.Master.Healthy = result
		mutex.Unlock()

		for i := range c.Replicas {
			result = HealthcheckQuery(c.Credentials, c.Healthcheck, c.Replicas[i])
			//log.Printf("[hc] replica: %d %t ", i, result)
			event = ProxyEvent{
				Name:    "hc",
				Message: fmt.Sprintf("replica is %t", result),
			}
			for j := range EventChannel {
				EventChannel[j] <- event
			}
			mutex.Lock()
			c.Replicas[i].Healthy = result
			mutex.Unlock()
		}
		time.Sleep(time.Duration(c.Healthcheck.Delay) * time.Second)
	}
}

func HealthcheckQuery(cred config.PGCredentials, hc config.Healthcheck, node config.Node) bool {

	var conn *sql.DB
	var err error
	var hostport = strings.Split(node.IPAddr, ":")
	var dbHost = hostport[0]
	var dbUser = cred.Username
	var dbPassword = cred.Password
	var dbPort = hostport[1]
	var database = cred.Database
	//log.Println("[hc] connecting to host:" + dbHost + " port:" + dbPort + " user:" + dbUser + " password:" + dbPassword + " database:" + database)
	conn, err = GetDBConnection(dbHost, dbUser, dbPort, database, dbPassword)
	defer conn.Close()

	if err != nil {
		glog.Errorln("[hc] healthcheck failed: error: " + err.Error())
		return false
	}
	//log.Println("[hc] got a connection")
	//var rows *sql.Rows
	_, err = conn.Query(hc.Query)
	if err != nil {
		//log.Println("[hc] failed: error: " + err.Error())
		return false
	}
	//log.Println("healthcheck passed")
	return true
}

func GetDBConnection(dbHost string, dbUser string, dbPort string, database string, dbPassword string) (*sql.DB, error) {

	var dbConn *sql.DB
	var err error

	if dbPassword == "" {
		//log.Println("a open db with dbHost=[" + dbHost + "] dbUser=[" + dbUser + "] dbPort=[" + dbPort + "] database=[" + database + "]")
		dbConn, err = sql.Open("postgres", "sslmode=disable user="+dbUser+" host="+dbHost+" port="+dbPort+" dbname="+database)
	} else {
		//log.Println("b open db with dbHost=[" + dbHost + "] dbUser=[" + dbUser + "] dbPort=[" + dbPort + "] database=[" + database + "] password=[" + dbPassword + "]")
		dbConn, err = sql.Open("postgres", "sslmode=disable user="+dbUser+" host="+dbHost+" port="+dbPort+" dbname="+database+" password="+dbPassword)
	}
	if err != nil {
		glog.Errorln("error in getting connection :" + err.Error())
	}
	return dbConn, err
}
