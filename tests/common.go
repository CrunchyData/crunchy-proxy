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
	log.Println("Connect starting...")

	var err error
	//os.Setenv("PGCONNECTION_TIMEOUT", "20")
	var hostportarr = strings.Split(hostport, ":")
	var dbHost = hostportarr[0]
	var dbPort = hostportarr[1]

	log.Println("connecting to host:" + dbHost + " port:" + dbPort + " user:" + userid + " password:" + password + " database:" + database)
	conn, err = GetDBConnection(dbHost, userid, dbPort, database, password)
	if err != nil {
		return nil, err
	}

	log.Println("got a connection")

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
