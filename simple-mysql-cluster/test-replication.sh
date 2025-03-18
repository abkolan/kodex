#!/bin/bash
# Test replication by creating a table and inserting data on primary

echo "Creating a test table on primary..."
docker exec mysql-primary mysql -u root -prootpass testdb -e "
CREATE TABLE IF NOT EXISTS test_table (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_table (name) VALUES ('Test Data 1'), ('Test Data 2'), ('Test Data 3');
"

echo "Waiting for replication to sync..."
sleep 10 # Increased from 5 to 10 seconds

echo "Data on primary:"
docker exec mysql-primary mysql -u root -prootpass testdb -e "SELECT * FROM test_table;"

echo "Data on replica:"
docker exec mysql-replica mysql -u root -prootpass testdb -e "SELECT * FROM test_table;"

# Test read-only mode on replica
echo "Testing read-only mode on replica..."
docker exec mysql-replica mysql -u root -prootpass testdb -e "
INSERT INTO test_table (name) VALUES ('Should fail');
" || echo "Insert operation failed on replica as expected (read-only)"
