package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage:	%s host:port <onerow|tworows>", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]
	sqltype := os.Args[2]
	fmt.Println("service is " + service)
	fmt.Println("sql type is " + sqltype)

	var conn *sql.DB
	var err error
	//os.Setenv("PGCONNECTION_TIMEOUT", "20")
	var hostport = strings.Split(service, ":")
	var dbHost = hostport[0]
	var dbUser = "postgres"
	var dbPassword = "password"
	var dbPort = hostport[1]
	var database = "postgres"
	fmt.Println("connecting to host:" + dbHost + " port:" + dbPort + " user:" + dbUser + " password:" + dbPassword + " database:" + database)
	conn, err = GetDBConnection(dbHost, dbUser, dbPort, database, dbPassword)

	checkError(err)
	fmt.Println("got a connection")
	if conn != nil {
		fmt.Println("conn is not nil")
	}
	switch sqltype {
	case "onerow":
		OneRow(conn)
		break
	case "tworows":
		TwoRows(conn)
		break
	}

	conn.Close()
	os.Exit(0)

}

func OneRow(conn *sql.DB) {
	var timestamp string
	err := conn.QueryRow("select text(now())").Scan(&timestamp)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("no rows returned")
	case err != nil:
		fmt.Println(err.Error())
	default:
		fmt.Println(timestamp + " was returned")
	}
}
func TwoRows(conn *sql.DB) {
	var timestamp string
	rows, err := conn.Query("select text(generate_series(1,2))")
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&timestamp); err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(timestamp)
	}
	if err = rows.Err(); err != nil {
		fmt.Println(err.Error())
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal	error:	%s", err.Error())
		os.Exit(1)
	}
}

func GetDBConnection(dbHost string, dbUser string, dbPort string, database string, dbPassword string) (*sql.DB, error) {

	var dbConn *sql.DB
	var err error

	if dbPassword == "" {
		//fmt.Println("a open db with dbHost=[" + dbHost + "] dbUser=[" + dbUser + "] dbPort=[" + dbPort + "] database=[" + database + "]")
		dbConn, err = sql.Open("postgres", "sslmode=disable user="+dbUser+" host="+dbHost+" port="+dbPort+" dbname="+database)
	} else {
		//fmt.Println("b open db with dbHost=[" + dbHost + "] dbUser=[" + dbUser + "] dbPort=[" + dbPort + "] database=[" + database + "] password=[" + dbPassword + "]")
		dbConn, err = sql.Open("postgres", "sslmode=disable user="+dbUser+" host="+dbHost+" port="+dbPort+" dbname="+database+" password="+dbPassword)
	}
	if err != nil {
		fmt.Println("error in getting connection :" + err.Error())
	}
	return dbConn, err
}
