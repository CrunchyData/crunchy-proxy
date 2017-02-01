#!/bin/bash 

# start up a master and a single replica for testing
CCP_IMAGE_TAG=centos7-9.6-1.2.7

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "starting master container..."

DATA_DIR=$DIR/.docker/master-data
sudo rm -rf $DATA_DIR
sudo mkdir -p $DATA_DIR
sudo chmod 777 $DATA_DIR
sudo chcon -Rt svirt_sandbox_file_t $DATA_DIR

CONFIG_DIR=$DIR/.docker/master-config
sudo rm -rf $CONFIG_DIR
sudo mkdir -p $CONFIG_DIR
sudo chmod 777 $CONFIG_DIR
sudo chcon -Rt svirt_sandbox_file_t $CONFIG_DIR


cp $DIR/master-config/postgresql.conf $CONFIG_DIR
cp $DIR/master-config/pg_hba.conf $CONFIG_DIR
sudo chmod 600 $CONFIG_DIR/pg_hba.conf $CONFIG_DIR/postgresql.conf
sudo chown postgres:postgres $CONFIG_DIR/pg_hba.conf $CONFIG_DIR/postgresql.conf

docker stop master
docker rm master

docker run \
	-p 12000:5432 \
	-v $CONFIG_DIR:/pgconf \
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

DATA_DIR=$DIR/.docker/replica-data
sudo rm -rf $DATA_DIR
sudo mkdir -p $DATA_DIR
sudo chown postgres:postgres $DATA_DIR
sudo chcon -Rt svirt_sandbox_file_t $DATA_DIR

CONFIG_DIR=$DIR/.docker/replica-config
sudo rm -rf $CONFIG_DIR
sudo mkdir -p $CONFIG_DIR
sudo chmod 777 $CONFIG_DIR
sudo chcon -Rt svirt_sandbox_file_t $CONFIG_DIR

cp $DIR/replica-config/postgresql.conf $CONFIG_DIR
cp $DIR/replica-config/pg_hba.conf $CONFIG_DIR
sudo chmod 600 $CONFIG_DIR/pg_hba.conf $CONFIG_DIR/postgresql.conf
sudo chown postgres:postgres $CONFIG_DIR/pg_hba.conf $CONFIG_DIR/postgresql.conf

sudo docker stop replica
sudo docker rm replica

sudo docker run \
	-p 12002:5432 \
	-v $CONFIG_DIR:/pgconf \
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

