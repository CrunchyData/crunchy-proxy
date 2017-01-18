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
package tests

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"strings"
)

func Connect() (*sql.DB, error) {
	var conn *sql.DB
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	var err error
	//os.Setenv("PGCONNECTION_TIMEOUT", "20")
	var hostportarr = strings.Split(HostPort, ":")
	var dbHost = hostportarr[0]
	var dbPort = hostportarr[1]

	log.Println("connecting to host:" + dbHost + " port:" + dbPort + " user:" + userid + " password:" + password + " database:" + database)
	conn, err = GetDBConnection(dbHost, userid, dbPort, database, password)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func GetDBConnection(dbHost string, userid string, dbPort string, database string, password string) (*sql.DB, error) {

	var dbConn *sql.DB
	var err error

	if password == "" {
		//log.Println("a open db with dbHost=[" + dbHost + "] userid=[" + userid + "] dbPort=[" + dbPort + "] database=[" + database + "]")
		dbConn, err = sql.Open("postgres", "sslmode=disable user="+userid+" host="+dbHost+" port="+dbPort+" dbname="+database)
	} else {
		//log.Println("b open db with dbHost=[" + dbHost + "] userid=[" + userid + "] dbPort=[" + dbPort + "] database=[" + database + "] password=[" + password + "]")
		dbConn, err = sql.Open("postgres", "sslmode=disable user="+userid+" host="+dbHost+" port="+dbPort+" dbname="+database+" password="+password)
	}
	if err != nil {
		log.Println("error in getting connection :" + err.Error())
	}
	return dbConn, err
}
