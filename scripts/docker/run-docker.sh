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

PROXY_TAG=centos7-1.0.0-beta
PROXY_PORT=5432
PROXY_ADMIN_PORT=10000

CONTAINER_NAME=crunchy-proxy

CONFIG=$(readlink -f ./config.yaml)

if [ -f /etc/redhat-release ]; then
	sudo chcon -Rt svirt_sandbox_file_t $(readlink -f $CONFIG)
fi

docker rm $CONTAINER_NAME
docker run -d --name=$CONTAINER_NAME \
	-p 127.0.0.1:$PROXY_PORT:$PROXY_PORT \
	-p 127.0.0.1:$PROXY_ADMIN_PORT:$PROXY_ADMIN_PORT \
	-v $CONFIG:/config/config.yaml \
	crunchydata/crunchy-proxy:$PROXY_TAG
