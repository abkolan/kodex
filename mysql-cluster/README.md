# MySQL Cluster with Docker Compose

This project helps you set up a MySQL cluster using Docker Compose and experiment with different failure scenarios and replication modes.

## Architecture Overview

This setup creates the following components:

1. **mysql-primary**: The primary MySQL server that accepts write operations
2. **mysql-replica-1**: A replica MySQL server that replicates data from the primary
3. **mysql-replica-2**: Another replica MySQL server that replicates data from the primary
4. **proxysql**: A MySQL proxy that handles load balancing and routing queries
5. **pmm-server**: Percona Monitoring and Management for monitoring the cluster

The architecture uses MySQL's built-in replication capabilities with GTID-based replication for improved fault tolerance and reliability.

## Prerequisites

- Docker and Docker Compose installed on your system
- Basic understanding of MySQL and database concepts
- At least 4GB of RAM available for the containers

## Setup Instructions

1. **Create the project structure**

   Create a new directory for your project and place all the files from this setup in it:

   ```bash
   mkdir mysql-cluster
   cd mysql-cluster
   ```

2. **Prepare configuration files**

   Run the config generation script to create all necessary configuration files:

   ```bash
   bash mysql-config-files.sh
   ```

   This will create the following directory structure:

   ```
   mysql-cluster/
   ├── docker-compose.yml
   ├── mysql-config-files.sh
   ├── mysql-failure-simulation.sh
   ├── config/
   │   ├── primary.cnf
   │   ├── replica.cnf
   │   └── proxysql.cnf
   └── scripts/
       └── init_replication.sh
   ```

3. **Start the cluster**

   Launch the MySQL cluster using Docker Compose:

   ```bash
   docker-compose up -d
   ```

   This will start all the containers defined in the docker-compose.yml file.

4. **Initialize replication**

   The replication setup script should run automatically, but you can run it manually if needed:

   ```bash
   docker-compose exec mysql-primary /docker-entrypoint-initdb.d/init_replication.sh
   ```

5. **Verify the cluster is working**

   Check the status of the cluster:

   ```bash
   docker-compose ps
   ```

   All containers should be in the "Up" state.

## Testing Replication

1. **Connect to the primary server**

   ```bash
   docker-compose exec mysql-primary mysql -u root -prootpass
   ```

2. **Create a test database and table on the primary**

   ```sql
   USE testdb;
   CREATE TABLE test (id INT AUTO_INCREMENT PRIMARY KEY, data VARCHAR(100));
   INSERT INTO test (data) VALUES ('test1'), ('test2'), ('test3');
   ```

3. **Verify data is replicated to the replicas**

   Connect to a replica:

   ```bash
   docker-compose exec mysql-replica-1 mysql -u root -prootpass
   ```

   Check the data:

   ```sql
   USE testdb;
   SELECT * FROM test;
   ```

   You should see the same data that you inserted on the primary.

## Using the Failure Simulation Tool

The `mysql-failure-simulation.sh` script allows you to simulate various failure scenarios:

1. **Make the script executable**

   ```bash
   chmod +x mysql-failure-simulation.sh
   ```

2. **Run the script**

   ```bash
   ./mysql-failure-simulation.sh
   ```

3. **Choose from the menu options**

   The script provides various options to:

   - Check cluster status
   - Create test data
   - Check data consistency across nodes
   - Simulate primary node failure
   - Recover primary node
   - Simulate replica failure
   - Recover replica
   - Simulate network partition
   - Heal network partition
   - Promote replica to primary (manual failover)

## Common MySQL Cluster Failure Scenarios

1. **Primary Node Failure**

   - What happens: The primary node becomes unavailable.
   - Expected behavior: Writes will fail, but reads can be redirected to replicas through ProxySQL.
   - Recovery: Promote a replica to become the new primary or restore the original primary.

2. **Replica Node Failure**

   - What happens: One or more replicas become unavailable.
   - Expected behavior: The cluster continues to function with reduced read capacity.
   - Recovery: Restart the failed replica and re-establish replication.

3. **Network Partition**

   - What happens: Network connectivity between nodes is lost.
   - Expected behavior: Replicas lose connection to the primary and stop receiving updates.
   - Recovery: Re-establish network connectivity and reset replication.

4. **Data Corruption**
   - What happens: Data becomes corrupted on one or more nodes.
   - Expected behavior: Depends on the type of corruption, but might result in replication errors.
   - Recovery: Restore from backup or rebuild the affected nodes.

## Using ProxySQL for Load Balancing

ProxySQL is configured to route:

- Read queries to replicas (hostgroup 1)
- Write queries to the primary (hostgroup 0)

Connect to ProxySQL:

```bash
mysql -h localhost -P 6033 -u testuser -ptestpass testdb
```

## Monitoring with Percona Monitoring and Management (PMM)

Access the PMM dashboard:

1. Open your browser and navigate to http://localhost
2. Login with default credentials (admin/admin)
3. You can now monitor your MySQL cluster's performance

## Understanding MySQL Replication Modes

This setup uses the following replication features:

1. **GTID-based Replication**

   - Global Transaction Identifiers uniquely identify transactions across the cluster
   - Provides better failover capabilities compared to traditional binary log position-based replication

2. **Row-based Replication**

   - The binary log contains the actual data changes rather than the SQL statements
   - More reliable for certain operations and avoids many edge cases

3. **Semi-synchronous Replication**
   - In this mode, the primary waits for at least one replica to acknowledge receipt of the transaction's events
   - Improves data consistency but can impact performance

## Advanced Configurations

You can modify this setup for different replication scenarios:

1. **Multi-source Replication**

   - Allow replicas to replicate from multiple sources
   - Good for aggregating data from multiple systems

2. **Delayed Replication**

   - Configure a replica to lag behind the primary
   - Useful for recovery from user errors (e.g., accidental DROP TABLE)

3. **Group Replication**
   - For multi-primary or single-primary with automatic failover
   - Would require modifications to the Docker Compose setup

## Troubleshooting

1. **Replication errors**

   - Check the replica status: `SHOW SLAVE STATUS\G`
   - Common issues include duplicate data and schema differences

2. **Container startup issues**

   - Check container logs: `docker-compose logs mysql-primary`
   - Ensure volumes have correct permissions

3. **ProxySQL routing issues**
   - Check ProxySQL admin interface: `mysql -h localhost -P 6032 -u admin -padmin`
   - Verify host groups and routing rules: `SELECT * FROM mysql_servers; SELECT * FROM mysql_query_rules;`

## Cleanup

To stop and remove all containers, networks, and volumes:

```bash
docker-compose down -v
```

## Reference Documentation

- [MySQL Replication Documentation](https://dev.mysql.com/doc/refman/8.0/en/replication.html)
- [ProxySQL Documentation](https://proxysql.com/documentation/)
- [Percona PMM Documentation](https://www.percona.com/doc/percona-monitoring-and-management/index.html)

## Next Steps

1. Experiment with different replication modes
2. Try implementing a multi-primary setup with Group Replication
3. Implement automated failover using tools like Orchestrator
4. Set up regular backups using tools like Percona XtraBackup
