# Create directories for configuration files
mkdir -p config scripts

# Primary MySQL Server Configuration
cat >config/primary.cnf <<EOF
[mysqld]
default_authentication_plugin=mysql_native_password
skip-host-cache
skip-name-resolve
datadir=/var/lib/mysql
socket=/var/run/mysqld/mysqld.sock
secure-file-priv=/var/lib/mysql-files
user=mysql

pid-file=/var/run/mysqld/mysqld.pid

# Replication settings
server-id=1
log_bin=mysql-bin
binlog_format=ROW
gtid_mode=ON
enforce_gtid_consistency=ON
binlog_expire_logs_seconds=604800
max_binlog_size=100M
sync_binlog=1
# Set the binary log replication ignored databases
binlog-ignore-db=performance_schema
binlog-ignore-db=information_schema
binlog-ignore-db=mysql
binlog-ignore-db=sys

# Performance settings
innodb_buffer_pool_size=256M
innodb_log_file_size=64M
innodb_flush_log_at_trx_commit=1
innodb_flush_method=O_DIRECT

# Connection settings
max_connections=500
EOF

# Replica MySQL Server Configuration
cat >config/replica.cnf <<EOF
[mysqld]
default_authentication_plugin=mysql_native_password
skip-host-cache
skip-name-resolve
datadir=/var/lib/mysql
socket=/var/run/mysqld/mysqld.sock
secure-file-priv=/var/lib/mysql-files
user=mysql

pid-file=/var/run/mysqld/mysqld.pid

# Replication settings (server-id set in docker-compose.yml)
log_bin=mysql-bin
binlog_format=ROW
gtid_mode=ON
enforce_gtid_consistency=ON
log_slave_updates=ON
relay_log=/var/lib/mysql/mysql-relay-bin
read_only=ON
# Replica should skip auto-increment conflicts
slave_skip_errors=1396
# Set the binary log replication ignored databases
binlog-ignore-db=performance_schema
binlog-ignore-db=information_schema
binlog-ignore-db=mysql
binlog-ignore-db=sys

# Performance settings
innodb_buffer_pool_size=256M
innodb_log_file_size=64M
innodb_flush_log_at_trx_commit=1
innodb_flush_method=O_DIRECT

# Connection settings
max_connections=500
EOF

# ProxySQL Configuration
cat >config/proxysql.cnf <<EOF
datadir="/var/lib/proxysql"

admin_variables=
{
    admin_credentials="admin:admin"
    mysql_ifaces="0.0.0.0:6032"
    refresh_interval=2000
}

mysql_variables=
{
    threads=4
    max_connections=2048
    default_query_delay=0
    default_query_timeout=36000000
    have_compress=true
    poll_timeout=2000
    interfaces="0.0.0.0:6033"
    default_schema="information_schema"
    stacksize=1048576
    server_version="8.0.28"
    connect_timeout_server=3000
    monitor_username="root"
    monitor_password="rootpass"
    monitor_history=600000
    monitor_connect_interval=60000
    monitor_ping_interval=10000
    monitor_read_only_interval=1500
    monitor_read_only_timeout=500
    ping_interval_server_msec=120000
    ping_timeout_server=500
    commands_stats=true
    sessions_sort=true
    connect_retries_on_failure=10
}

mysql_servers =
(
    {
        address="mysql-primary"
        port=3306
        hostgroup=0
        max_connections=100
        weight=1
        status="ONLINE"
        max_replication_lag=0
    },
    {
        address="mysql-replica-1"
        port=3306
        hostgroup=1
        max_connections=100
        weight=1
        status="ONLINE"
        max_replication_lag=5
    },
    {
        address="mysql-replica-2"
        port=3306
        hostgroup=1
        max_connections=100
        weight=1
        status="ONLINE"
        max_replication_lag=5
    }
)

mysql_users =
(
    {
        username="root"
        password="rootpass"
        default_hostgroup=0
        active=1
        max_connections=1000
    },
    {
        username="testuser"
        password="testpass"
        default_hostgroup=0
        active=1
        max_connections=1000
    }
)

mysql_query_rules =
(
    {
        rule_id=1
        active=1
        match_digest="^SELECT.*"
        destination_hostgroup=1
        apply=1
    },
    {
        rule_id=2
        active=1
        match_digest="^(INSERT|UPDATE|DELETE|REPLACE).*"
        destination_hostgroup=0
        apply=1
    }
)
EOF

# Create initialization script for setting up replication
cat >scripts/init_replication.sh <<EOF
#!/bin/bash
set -e

echo "Waiting for MySQL primary to be ready..."
# More robust wait mechanism with timeout
max_attempts=30
attempt=0
while [ \$attempt -lt \$max_attempts ]; do
    if mysql -h mysql-primary -u root -prootpass -e "SELECT 1" &>/dev/null; then
        echo "Successfully connected to MySQL primary!"
        break
    fi
    attempt=\$((attempt+1))
    echo "Attempt \$attempt/\$max_attempts: MySQL primary not ready yet, waiting 5 seconds..."
    sleep 5
done

if [ \$attempt -eq \$max_attempts ]; then
    echo "ERROR: Timed out waiting for MySQL primary to be ready after \$max_attempts attempts."
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
    until mysql -h $replica -u root -prootpass -e "SELECT 1"; do
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
EOF

# Make the init script executable
chmod +x scripts/init_replication.sh

# Create Prometheus configuration
mkdir -p config
cat >config/prometheus.yml <<EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'mysql-nodes'
    static_configs:
      - targets: ['mysql-exporter-primary:9104', 'mysql-exporter-replica1:9104', 'mysql-exporter-replica2:9104']
        labels:
          group: 'mysql-cluster'
EOF
