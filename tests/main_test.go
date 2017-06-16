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
	"flag"
	"os"
	"testing"
)

var HostPort string
var rows, userid, password, database string

func TestMain(m *testing.M) {
	flag.StringVar(&rows, "rows", "onerow", "onerow or tworows")
	flag.StringVar(&HostPort, "hostport", "localhost:5432", "host:port")
	flag.StringVar(&userid, "userid", "postgres", "postgres userid")
	flag.StringVar(&password, "password", "password", "postgres password")
	flag.StringVar(&database, "database", "postgres", "database")
	flag.Parse()
	os.Exit(m.Run())

}
