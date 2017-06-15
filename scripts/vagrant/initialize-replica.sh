#!/bin/bash

PG_BASEBACKUP=/usr/bin/pg_basebackup
PG_USER=rep
PG_MASTER_IP=192.168.56.100
PG_DATA_DIR=/var/lib/pgsql/9.6/data

SUDO='sudo -i PWD=/var/lib/pgsql -u postgres'

echo "Removing Data Directory..."
$SUDO rm -rf /var/lib/pgsql/9.6/data
echo "Done."


# Initialize the replica data directory based on the master.
#
# pg_basebackup parameters (provided for reference):
#
# * -P - Show progress information
# * -R - Write recovery.conf file
# * -c fast - set fast checkpointing
# * -X stream - set xlog method to streaming
# * -h - the ip of the master
# * -U - the replication user
# * -D - the directory to receive the base backup.
#
echo "Performing Base Backup..."
$SUDO $PG_BASEBACKUP -P -R -c fast -X stream -h $PG_MASTER_IP -U $PG_USER -D $PG_DATA_DIR

echo "Setting up recovery.conf..."
$SUDO "cat > /var/lib/pgsql/9.6/data/recovery.conf << 'EOF'
  trigger_file = '$PG_DATA_DIR/postgresql.trigger'
EOF
"
echo "Done."


