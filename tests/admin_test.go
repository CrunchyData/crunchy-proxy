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
	"log"
	"net/http"
	"testing"
	"time"
)

func TestAPIStream(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("TestAPIStream was called")
	var startTime = time.Now()

	req, err := http.NewRequest("GET", "http://localhost:10000/api/stream", nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		t.FailNow()
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		t.FailNow()
	}

	defer resp.Body.Close()

	var endTime = time.Since(startTime)
	log.Printf("Duration %s\n", endTime)

}

func TestAPIConfig(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("TestAPIConfig was called")
	var startTime = time.Now()

	req, err := http.NewRequest("GET", "http://localhost:10000/api/config", nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		t.FailNow()
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		t.FailNow()
	}

	defer resp.Body.Close()

	var endTime = time.Since(startTime)
	log.Printf("Duration %s\n", endTime)

}
func TestAPIStats(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("TestAPIStats was called")
	var startTime = time.Now()

	req, err := http.NewRequest("GET", "http://localhost:10000/api/stats", nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		t.FailNow()
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		t.FailNow()
	}

	defer resp.Body.Close()

	var endTime = time.Since(startTime)
	log.Printf("Duration %s\n", endTime)

}
