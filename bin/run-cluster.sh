#!/bin/bash 

# start up a master and a single replica for testing
CCP_IMAGE_TAG=centos7-9.5-1.2.6

echo "starting master container..."

DATA_DIR=$HOME/master-data
sudo rm -rf $DATA_DIR
sudo mkdir -p $DATA_DIR
sudo chmod 777 $DATA_DIR
sudo chcon -Rt svirt_sandbox_file_t $DATA_DIR

docker stop master
docker rm master

docker run \
	-p 12000:5432 \
	-v $DATA_DIR:/pgdata \
	-e TEMP_BUFFERS=9MB \
	-e MAX_CONNECTIONS=101 \
	-e SHARED_BUFFERS=129MB \
	-e MAX_WAL_SENDERS=7 \
	-e WORK_MEM=5MB \
	-e PG_MODE=master \
	-e PG_MASTER_USER=masteruser \
	-e PG_MASTER_PASSWORD=password \
	-e PG_USER=testuser \
	-e PG_ROOT_PASSWORD=password \
	-e PG_PASSWORD=password \
	-e PG_DATABASE=userdb \
	-e PGHOST=/tmp \
	--name=master \
	--hostname=master \
	-d crunchydata/crunchy-postgres:$CCP_IMAGE_TAG

echo "sleeping a bit before starting replica..."
sleep 10
echo "starting replica container..."

DATA_DIR=$HOME/replica-data
sudo rm -rf $DATA_DIR
sudo mkdir -p $DATA_DIR
sudo chown postgres:postgres $DATA_DIR
sudo chcon -Rt svirt_sandbox_file_t $DATA_DIR

sudo docker stop replica
sudo docker rm replica

sudo docker run \
	-p 12002:5432 \
	-v $DATA_DIR:/pgdata \
	-e TEMP_BUFFERS=9MB \
	-e MAX_CONNECTIONS=101 \
	-e SHARED_BUFFERS=129MB \
	-e MAX_WAL_SENDERS=7 \
	-e WORK_MEM=5MB \
	-e PG_MODE=slave \
	-e PG_MASTER_USER=masteruser \
	-e PG_MASTER_PASSWORD=password \
	-e PG_MASTER_HOST=master \
	--link master:master \
	-e PG_MASTER_PORT=5432 \
	-e PG_USER=testuser \
	-e PG_ROOT_PASSWORD=password \
	-e PG_PASSWORD=password \
	-e PG_DATABASE=userdb \
	-e PGHOST=/tmp \
	--name=replica \
	--hostname=replica \
	-d crunchydata/crunchy-postgres:$CCP_IMAGE_TAG

