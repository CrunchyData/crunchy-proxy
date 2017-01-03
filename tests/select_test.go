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
	"log"
	"testing"
	"time"
)

func TestSelect(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("TestSelect was called")
	var startTime = time.Now()
	conn, err := Connect()
	defer conn.Close()
	if err != nil {
		t.FailNow()
	}

	var timestamp string
	err = conn.QueryRow("/* read */ select text(now())").Scan(&timestamp)
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
