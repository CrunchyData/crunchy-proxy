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
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("main starting...")

	var rows, hostport, userid, password, database string
	var count int
	flag.StringVar(&rows, "rows", "onerow", "onerow or tworows")
	flag.StringVar(&hostport, "hostport", "localhost:5432", "host:port")
	flag.StringVar(&userid, "userid", "postgres", "postgres userid")
	flag.StringVar(&password, "password", "password", "postgres password")
	flag.StringVar(&database, "database", "postgres", "database")
	flag.IntVar(&count, "count", 1, "number of executions")
	flag.Parse()

	var conn *sql.DB
	var err error
	//os.Setenv("PGCONNECTION_TIMEOUT", "20")
	var hostportarr = strings.Split(hostport, ":")
	var dbHost = hostportarr[0]
	var dbPort = hostportarr[1]

	log.Println("connecting to host:" + dbHost +
		" port:" + dbPort +
		" user:" + userid +
		" password:" + password +
		" database:" + database)

	conn, err = GetDBConnection(dbHost, userid, dbPort, database, password)

	checkError(err)
	log.Println("got a connection")
	if conn != nil {
		log.Println("conn is not nil")
	}

	log.Println("execution starts")
	var startTime = time.Now()

	switch rows {
	case "onerow":
		for i := 0; i < count; i++ {
			OneRow(conn)
		}
		break
	case "tworows":
		TwoRows(conn)
		break
	}

	log.Println("execution ends")

	var endTime = time.Since(startTime)

	log.Printf("Duration %s\n", endTime)

	conn.Close()
	os.Exit(0)
}

func OneRow(conn *sql.DB) {
	var timestamp string
	err := conn.QueryRow("/* read */ select text(now())").Scan(&timestamp)
	switch {
	case err == sql.ErrNoRows:
		log.Println("no rows returned")
	case err != nil:
		log.Println(err.Error())
	default:
		log.Println(timestamp + " was returned")
	}
}

func TwoRows(conn *sql.DB) {
	var timestamp string
	rows, err := conn.Query("/* read */ select text(generate_series(1,2))")
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&timestamp); err != nil {
			log.Println(err.Error())
		}
		log.Println(timestamp)
	}
	if err = rows.Err(); err != nil {
		log.Println(err.Error())
	}
}

func checkError(err error) {
	if err != nil {
		log.Println("Fatal error:" + err.Error())
		os.Exit(1)
	}
}

func GetDBConnection(dbHost string, userid string, dbPort string, database string, password string) (*sql.DB, error) {

	var dbConn *sql.DB
	var err error

	if password == "" {
		dbConn, err = sql.Open("postgres",
			"sslmode=disable user="+userid+
				" host="+dbHost+
				" port="+dbPort+
				" dbname="+database+
				" sslmode=disable")
	} else {
		dbConn, err = sql.Open("postgres",
			"sslmode=disable user="+userid+
				" host="+dbHost+
				" port="+dbPort+
				" dbname="+database+
				" password="+password+
				" sslmode=disable")
	}
	if err != nil {
		log.Println("error in getting connection :" + err.Error())
	}
	return dbConn, err
}
