#!/bin/bash
set -e

echo "Waiting for MySQL primary to be ready..."
# More robust wait mechanism with timeout
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if mysql -h mysql-primary -u root -prootpass -e "SELECT 1" &>/dev/null; then
        echo "Successfully connected to MySQL primary!"
        break
    fi
    attempt=$((attempt+1))
    echo "Attempt $attempt/$max_attempts: MySQL primary not ready yet, waiting 5 seconds..."
    sleep 5
done

if [ $attempt -eq $max_attempts ]; then
    echo "ERROR: Timed out waiting for MySQL primary to be ready after $max_attempts attempts."
    exit 1
fi

echo "Creating replication user..."
mysql -h mysql-primary -u root -prootpass -e "
CREATE USER 'repl'@'%' IDENTIFIED BY 'replpass';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
FLUSH PRIVILEGES;
"

# Wait for replicas to be ready
echo "Waiting for MySQL replicas to be ready..."
for replica in mysql-replica-1 mysql-replica-2; do
    until mysql -h  -u root -prootpass -e "SELECT 1"; do
        sleep 1
    done
done

# Configure replication on replica 1
echo "Configuring replication on mysql-replica-1..."
mysql -h mysql-replica-1 -u root -prootpass -e "
CHANGE MASTER TO
  MASTER_HOST='mysql-primary',
  MASTER_USER='repl',
  MASTER_PASSWORD='replpass',
  MASTER_AUTO_POSITION=1;
START SLAVE;
"

# Configure replication on replica 2
echo "Configuring replication on mysql-replica-2..."
mysql -h mysql-replica-2 -u root -prootpass -e "
CHANGE MASTER TO
  MASTER_HOST='mysql-primary',
  MASTER_USER='repl',
  MASTER_PASSWORD='replpass',
  MASTER_AUTO_POSITION=1;
START SLAVE;
"

echo "Replication setup complete."
