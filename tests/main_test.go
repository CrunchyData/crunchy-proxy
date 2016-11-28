package tests

import (
	"flag"
	"os"
	"testing"
)

var rows, hostport, userid, password, database string

func TestMain(m *testing.M) {
	flag.StringVar(&rows, "rows", "onerow", "onerow or tworows")
	flag.StringVar(&hostport, "hostport", "localhost:5432", "host:port")
	flag.StringVar(&userid, "userid", "postgres", "postgres userid")
	flag.StringVar(&password, "password", "password", "postgres password")
	flag.StringVar(&database, "database", "postgres", "database")
	flag.Parse()
	os.Exit(m.Run())

}
