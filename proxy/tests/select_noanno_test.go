package tests

import (
	"database/sql"
	"log"
	"testing"
	"time"
)

func TestSelectNoAnno(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("TestSelect was called")
	var startTime = time.Now()
	conn, err := Connect()
	defer conn.Close()
	if err != nil {
		t.FailNow()
	}

	var timestamp string
	err = conn.QueryRow("select text(now())").Scan(&timestamp)
	switch {
	case err == sql.ErrNoRows:
		log.Println("no rows returned")
		t.FailNow()
	case err != nil:
		log.Println(err.Error())
		t.FailNow()
	default:
		log.Println(timestamp + " was returned")
	}

	var endTime = time.Since(startTime)
	log.Printf("Duration %s\n", endTime)

}
