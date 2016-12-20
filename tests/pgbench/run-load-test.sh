#!/bin/bash 

# Copyright 2016 Crunchy Data Solutions, Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "starting pgbench load test..."

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export PGPASSFILE=$DIR/pgpass

PORT=5432
HOST=localhost

echo "refresh the proxydb database.."
psql -h $HOST -p $PORT -U postgres -c 'drop database proxydb;' postgres
psql -h $HOST -p $PORT -U postgres -c 'create database proxydb;' postgres
pgbench -h localhost -p 12000 -U postgres -i proxydb
psql -h $HOST -p $PORT -U postgres -c 'create table proxytest (id int, name varchar(20), value varchar(20));' proxydb

echo "start the load test..."

pgbench -h $HOST -p $PORT \
	-U postgres -f $DIR/load-test.sql \
	-t 1 proxydb

echo "load test ends."
