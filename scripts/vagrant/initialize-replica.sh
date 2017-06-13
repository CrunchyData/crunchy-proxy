#!/bin/bash

PG_BASEBACKUP=/usr/bin/pg_basebackup
PG_USER=rep
PG_MASTER_IP=192.168.56.100

SUDO='sudo -i PWD=/var/lib/pgsql -u postgres'

echo "Removing Data Directory..."
$SUDO rm -rf /var/lib/pgsql/9.6/data
echo "Done."

echo "Performing Base Backup..."
$SUDO $PG_BASEBACKUP -h $PG_MASTER_IP -D /var/lib/pgsql/9.6/data -U $PG_USER -v -P

echo "Setting up recovery.conf..."
$SUDO "cat > /var/lib/pgsql/9.6/data/recovery.conf << 'EOF'
  standby_mode = 'on'
  primary_conninfo = 'host=$PG_MASTER_IP port=5432 user=$PG_USER sslmode=require'
  trigger_file = '/tmp/postgresql.trigger'
EOF
"
echo "Done."


