#!/bin/bash

# MySQL Cluster Failure Simulation Script
# This script helps to simulate different failure scenarios and test replication

# Color codes for better output readability
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if docker-compose is running
check_cluster_status() {
    echo -e "${YELLOW}Checking cluster status...${NC}"
    docker-compose ps

    # Check replication status
    echo -e "\n${YELLOW}Checking replication status for mysql-replica-1:${NC}"
    docker-compose exec mysql-replica-1 mysql -u root -prootpass -e "SHOW SLAVE STATUS\G"

    echo -e "\n${YELLOW}Checking replication status for mysql-replica-2:${NC}"
    docker-compose exec mysql-replica-2 mysql -u root -prootpass -e "SHOW SLAVE STATUS\G"
}

# Create test data
create_test_data() {
    echo -e "${YELLOW}Creating test data on primary node...${NC}"
    docker-compose exec mysql-primary mysql -u root -prootpass testdb -e "
  CREATE TABLE IF NOT EXISTS test_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100),
    value INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  
  -- Insert some initial data
  INSERT INTO test_data (name, value) VALUES 
    ('test1', 100),
    ('test2', 200),
    ('test3', 300);
  "

    echo -e "${GREEN}Test data created successfully!${NC}"
}

# Check data consistency
check_data_consistency() {
    echo -e "${YELLOW}Checking data consistency across nodes...${NC}"

    echo -e "\n${YELLOW}Data on primary:${NC}"
    docker-compose exec mysql-primary mysql -u root -prootpass testdb -e "SELECT * FROM test_data;"

    echo -e "\n${YELLOW}Data on replica-1:${NC}"
    docker-compose exec mysql-replica-1 mysql -u root -prootpass testdb -e "SELECT * FROM test_data;"

    echo -e "\n${YELLOW}Data on replica-2:${NC}"
    docker-compose exec mysql-replica-2 mysql -u root -prootpass testdb -e "SELECT * FROM test_data;"
}

# Function to simulate primary node failure
simulate_primary_failure() {
    echo -e "${RED}Simulating primary node failure...${NC}"
    docker-compose stop mysql-primary
    echo -e "${YELLOW}Primary node is now down. You should see the ProxySQL redirecting read queries to replicas.${NC}"
    echo -e "${YELLOW}You can try connecting to the cluster through ProxySQL:${NC}"
    echo "mysql -h localhost -P 6033 -u testuser -ptestpass testdb"
}

# Function to recover the primary node
recover_primary() {
    echo -e "${YELLOW}Recovering primary node...${NC}"
    docker-compose start mysql-primary
    sleep 10 # Wait for primary to be fully up
    echo -e "${GREEN}Primary node is back online!${NC}"
}

# Function to simulate replica failure
simulate_replica_failure() {
    echo -e "${YELLOW}Which replica do you want to simulate failure for? (1/2)${NC}"
    read replica_num

    if [ "$replica_num" == "1" ]; then
        echo -e "${RED}Simulating replica-1 failure...${NC}"
        docker-compose stop mysql-replica-1
    elif [ "$replica_num" == "2" ]; then
        echo -e "${RED}Simulating replica-2 failure...${NC}"
        docker-compose stop mysql-replica-2
    else
        echo -e "${RED}Invalid selection. Please enter 1 or 2.${NC}"
        return
    fi

    echo -e "${YELLOW}Replica-$replica_num is now down. The cluster should continue to function.${NC}"
}

# Function to recover a replica
recover_replica() {
    echo -e "${YELLOW}Which replica do you want to recover? (1/2)${NC}"
    read replica_num

    if [ "$replica_num" == "1" ]; then
        echo -e "${YELLOW}Recovering replica-1...${NC}"
        docker-compose start mysql-replica-1
        sleep 10 # Wait for replica to be fully up

        # Reset replication if needed
        docker-compose exec mysql-replica-1 mysql -u root -prootpass -e "
    STOP SLAVE;
    RESET SLAVE;
    CHANGE MASTER TO
      MASTER_HOST='mysql-primary',
      MASTER_USER='repl',
      MASTER_PASSWORD='replpass',
      MASTER_AUTO_POSITION=1;
    START SLAVE;
    "
    elif [ "$replica_num" == "2" ]; then
        echo -e "${YELLOW}Recovering replica-2...${NC}"
        docker-compose start mysql-replica-2
        sleep 10 # Wait for replica to be fully up

        # Reset replication if needed
        docker-compose exec mysql-replica-2 mysql -u root -prootpass -e "
    STOP SLAVE;
    RESET SLAVE;
    CHANGE MASTER TO
      MASTER_HOST='mysql-primary',
      MASTER_USER='repl',
      MASTER_PASSWORD='replpass',
      MASTER_AUTO_POSITION=1;
    START SLAVE;
    "
    else
        echo -e "${RED}Invalid selection. Please enter 1 or 2.${NC}"
        return
    fi

    echo -e "${GREEN}Replica-$replica_num is back online and replication has been reset!${NC}"
}

# Function to simulate network partition
simulate_network_partition() {
    echo -e "${RED}Simulating network partition by disconnecting replicas from primary...${NC}"
    docker network disconnect mysql-cluster-network mysql-replica-1
    docker network disconnect mysql-cluster-network mysql-replica-2

    echo -e "${YELLOW}Network partition simulated. Replicas are now disconnected from primary.${NC}"
}

# Function to heal network partition
heal_network_partition() {
    echo -e "${YELLOW}Healing network partition...${NC}"
    docker network connect mysql-cluster-network mysql-replica-1
    docker network connect mysql-cluster-network mysql-replica-2

    sleep 5

    # Reset replication on both replicas
    echo -e "${YELLOW}Resetting replication on replicas...${NC}"

    docker-compose exec mysql-replica-1 mysql -u root -prootpass -e "
  STOP SLAVE;
  RESET SLAVE;
  CHANGE MASTER TO
    MASTER_HOST='mysql-primary',
    MASTER_USER='repl',
    MASTER_PASSWORD='replpass',
    MASTER_AUTO_POSITION=1;
  START SLAVE;
  "

    docker-compose exec mysql-replica-2 mysql -u root -prootpass -e "
  STOP SLAVE;
  RESET SLAVE;
  CHANGE MASTER TO
    MASTER_HOST='mysql-primary',
    MASTER_USER='repl',
    MASTER_PASSWORD='replpass',
    MASTER_AUTO_POSITION=1;
  START SLAVE;
  "

    echo -e "${GREEN}Network partition healed and replication reset!${NC}"
}

# Function to promote a replica to primary
promote_replica_to_primary() {
    echo -e "${YELLOW}Which replica do you want to promote to primary? (1/2)${NC}"
    read replica_num

    if [ "$replica_num" != "1" ] && [ "$replica_num" != "2" ]; then
        echo -e "${RED}Invalid selection. Please enter 1 or 2.${NC}"
        return
    fi

    echo -e "${RED}Promoting replica-$replica_num to primary. This simulates a failover scenario.${NC}"

    # Stop the current primary
    docker-compose stop mysql-primary

    # Get the chosen replica container name
    replica_container="mysql-replica-$replica_num"

    # Disable read-only on the chosen replica
    docker-compose exec $replica_container mysql -u root -prootpass -e "
  STOP SLAVE;
  RESET SLAVE ALL;
  SET GLOBAL read_only = 0;
  SET GLOBAL super_read_only = 0;
  "

    echo -e "${GREEN}Replica-$replica_num has been promoted to primary!${NC}"
    echo -e "${YELLOW}Note: In a production environment, you would also need to update ProxySQL configuration to route write traffic to the new primary.${NC}"

    # You can update ProxySQL configuration here if you want to implement that
}

# Main menu
show_menu() {
    echo -e "\n${GREEN}MySQL Cluster Management and Failure Simulation${NC}"
    echo "-------------------------------------------"
    echo "1. Check cluster status"
    echo "2. Create test data"
    echo "3. Check data consistency"
    echo "4. Simulate primary node failure"
    echo "5. Recover primary node"
    echo "6. Simulate replica failure"
    echo "7. Recover replica"
    echo "8. Simulate network partition"
    echo "9. Heal network partition"
    echo "10. Promote replica to primary (manual failover)"
    echo "11. Exit"
    echo -e "${YELLOW}Enter your choice [1-11]:${NC} "
}

# Main loop
while true; do
    show_menu
    read choice

    case $choice in
    1) check_cluster_status ;;
    2) create_test_data ;;
    3) check_data_consistency ;;
    4) simulate_primary_failure ;;
    5) recover_primary ;;
    6) simulate_replica_failure ;;
    7) recover_replica ;;
    8) simulate_network_partition ;;
    9) heal_network_partition ;;
    10) promote_replica_to_primary ;;
    11)
        echo -e "${GREEN}Goodbye!${NC}"
        exit 0
        ;;
    *) echo -e "${RED}Invalid option. Please try again.${NC}" ;;
    esac

    echo -e "\nPress Enter to continue..."
    read
done
