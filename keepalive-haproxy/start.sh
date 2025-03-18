#!/bin/bash
set -e

# Make sure haproxy config file has proper line endings
echo "" >> /usr/local/etc/haproxy/haproxy.cfg

# Verify HAProxy configuration
echo "Checking HAProxy configuration..."
haproxy -c -f /usr/local/etc/haproxy/haproxy.cfg

echo "Starting Keepalived..."
keepalived --dont-fork --log-console &
KEEPALIVED_PID=$!

echo "Starting HAProxy..."
haproxy -f /usr/local/etc/haproxy/haproxy.cfg -d

# If HAProxy exits, kill keepalived too
kill $KEEPALIVED_PID || true

# Keep container running
tail -f /dev/null