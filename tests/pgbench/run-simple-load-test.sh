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
PROXY_HOST=localhost
MASTER_HOST=master.crunchy.lab

echo "Initialize the 'proxydb' database..."
pgbench -h $MASTER_HOST -p $PORT  -U postgres -i proxydb

echo "start the load test..."

pgbench \
	-h $PROXY_HOST \
	-p $PORT \
	-U postgres \
	-f $DIR/simple-load-test.sql \
	-c 10 \
	--no-vacuum \
	-t 10000 proxydb

echo "load test ends."
