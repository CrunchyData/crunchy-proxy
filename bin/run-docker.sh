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
PROXY=/home/jeffmc/gdev/src/github.com/crunchydata/crunchy-proxy
CONFIG=$PROXY/config.json
sudo chcon -Rt svirt_sandbox_file_t $CONFIG
PROXY_TAG=centos7-0.0.1
CONTAINER=crunchyproxy
docker rm $CONTAINER
docker run -d --name=$CONTAINER \
	-p 5432:5432 \
	-v $CONFIG:/config/config.json \
	-d crunchy-proxy:latest
