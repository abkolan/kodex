# Keepalived HAProxy High Availability Demo

This repository demonstrates a high availability setup using HAProxy and Keepalived in Docker containers. The setup shows how a virtual IP address automatically moves between servers when one fails.

## Architecture

This demo includes:

- Two HAProxy servers in an active-passive configuration
- Keepalived managing a virtual IP (VIP) between them
- Automatic failover when the active server goes down
- Everything running in Docker containers
- A TypeScript monitoring script to visualize the failover

## How Keepalived Works

Keepalived uses VRRP (Virtual Router Redundancy Protocol) to provide high availability:

1. Both servers run Keepalived with the same virtual_router_id (51)
2. The MASTER server (node1) has higher priority (101 vs 100)
3. Keepalived assigns the virtual IP (172.20.0.100) to the MASTER server
4. Keepalived monitors HAProxy on both servers
5. If the MASTER fails, the BACKUP takes over the virtual IP
6. Client requests are automatically redirected to the new active server

## Prerequisites

- Docker and Docker Compose
- Node.js and npm (for the monitoring script)

## Running the Demo

### Step 1: Start the Docker containers

```bash
# Create the required directories
mkdir -p haproxy keepalived scripts

# Copy all files from this repository to their corresponding locations
# ...

# Build and start the containers
docker-compose up -d
```

### Step 2: Run the monitoring script

```bash
# Install dependencies
cd scripts
npm install

# Start the monitoring
npm start
```

You should see output like:
```
Monitoring HAProxy service. Press Ctrl+C to stop.
Sending requests to Virtual IP: http://172.20.0.100
[2023-10-30T12:00:00.000Z] Response from HAProxy NODE 1 (MASTER)
[2023-10-30T12:00:01.000Z] Response from HAProxy NODE 1 (MASTER)
...
```

### Step 3: Demonstrate failover

In a new terminal window, stop the active (master) server:

```bash
docker stop haproxy_node1
```

You should see the monitoring script output change from node1 to node2:

```
[2023-10-30T12:01:00.000Z] Response from HAProxy NODE 1 (MASTER)
[2023-10-30T12:01:01.000Z] Response from HAProxy NODE 1 (MASTER)
[2023-10-30T12:01:02.000Z] Error: connect ECONNREFUSED 172.20.0.100:80
[2023-10-30T12:01:03.000Z] Response from HAProxy NODE 2 (BACKUP)
[2023-10-30T12:01:04.000Z] Response from HAProxy NODE 2 (BACKUP)
```

This demonstrates the failover process. The virtual IP has moved from node1 to node2.

### Step 4: Restart the failed node

```bash
docker start haproxy_node1
```

After a while, node1 will come back online, but the virtual IP will remain on node2 (this is standard behavior in Keepalived to prevent unnecessary flapping).

## Understanding the Components

### HAProxy Configuration

- Each HAProxy instance is configured to return a unique response
- This makes it easy to see which server is handling requests
- In a real-world scenario, HAProxy would forward requests to backend servers

### Keepalived Configuration

- Uses VRRP to manage the virtual IP
- Monitors HAProxy health with the `check_haproxy` script
- Assigns higher priority to the master node
- Automatically moves the VIP when the master node fails

## Monitoring and Administration

- HAProxy stats are available at:
  - Master node: http://localhost:8081/stats
  - Backup node: http://localhost:8082/stats

## Cleanup

```bash
# Stop and remove all containers
docker-compose down
```