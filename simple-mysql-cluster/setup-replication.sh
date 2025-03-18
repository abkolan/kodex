#!/bin/bash
# Wait for MySQL servers to be up and running
echo "Waiting for MySQL servers to start..."
sleep 30 # Increased wait time

# Create replication user on primary
echo "Creating replication user..."
docker exec mysql-primary mysql -u root -prootpass -e "
CREATE USER 'replica'@'%' IDENTIFIED WITH mysql_native_password BY 'replicapass';
GRANT REPLICATION SLAVE ON *.* TO 'replica'@'%';
FLUSH PRIVILEGES;
"

# Configure replication on replica
echo "Configuring replication..."
docker exec mysql-replica mysql -u root -prootpass -e "
CHANGE MASTER TO
  MASTER_HOST='mysql-primary',
  MASTER_USER='replica',
  MASTER_PASSWORD='replicapass',
  MASTER_AUTO_POSITION=1,
  GET_MASTER_PUBLIC_KEY=1;
START SLAVE;
"

# Check replication status
echo "Checking replication status:"
docker exec mysql-replica mysql -u root -prootpass -e "SHOW SLAVE STATUS\G"

echo "Setup complete!"
echo "Primary MySQL: localhost:3306 (user: root, password: rootpass)"
echo "Replica MySQL: localhost:3307 (user: root, password: rootpass)"
