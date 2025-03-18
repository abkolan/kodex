CHANGE MASTER TO 
    MASTER_HOST='mysql-master',
    MASTER_USER='replica_user',
    MASTER_PASSWORD='replica_password',
    MASTER_LOG_FILE='mysql-bin.000001',
    MASTER_LOG_POS=0;
START SLAVE;
SHOW SLAVE STATUS\G;