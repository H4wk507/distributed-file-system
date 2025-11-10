#!/bin/bash

# Start a test cluster with 1 master and 3 storage nodes

echo "Building node binary..."
cd "$(dirname "$0")/.."
go build -o bin/node ./cmd/node
echo "Build complete"

echo "Starting Master Node (Port 9000, Priority 10)..."
./bin/node -ip 127.0.0.1 -port 9000 -role master -priority 10 &
MASTER_PID=$!
echo "  PID: $MASTER_PID"

sleep 2

echo ""
echo "Starting Storage Node 1 (Port 9001, Priority 5)..."
./bin/node -ip 127.0.0.1 -port 9001 -role storage -priority 5 \
    -seed-ip 127.0.0.1 -seed-port 9000 &
NODE1_PID=$!
echo "  PID: $NODE1_PID"

sleep 1

echo ""
echo "Starting Storage Node 2 (Port 9002, Priority 3)..."
./bin/node -ip 127.0.0.1 -port 9002 -role storage -priority 3 \
    -seed-ip 127.0.0.1 -seed-port 9000 &
NODE2_PID=$!
echo "  PID: $NODE2_PID"

sleep 1

echo ""
echo "Starting Storage Node 3 (Port 9003, Priority 1)..."
./bin/node -ip 127.0.0.1 -port 9003 -role storage -priority 1 \
    -seed-ip 127.0.0.1 -seed-port 9000 &
NODE3_PID=$!
echo "  PID: $NODE3_PID"

echo ""
echo "Cluster started successfully!"
echo ""
echo "═══════════════════════════════════════════════════════"
echo "  Master Node:    127.0.0.1:9000 (Priority 10)"
echo "  Storage Node 1: 127.0.0.1:9001 (Priority 5)"
echo "  Storage Node 2: 127.0.0.1:9002 (Priority 3)"
echo "  Storage Node 3: 127.0.0.1:9003 (Priority 1)"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "Process IDs:"
echo "  Master: $MASTER_PID"
echo "  Node 1: $NODE1_PID"
echo "  Node 2: $NODE2_PID"
echo "  Node 3: $NODE3_PID"
echo ""

echo "$MASTER_PID" > /tmp/dfs-master.pid
echo "$NODE1_PID" > /tmp/dfs-node1.pid
echo "$NODE2_PID" > /tmp/dfs-node2.pid
echo "$NODE3_PID" > /tmp/dfs-node3.pid

trap "echo ''; echo 'Stopping cluster...'; kill $MASTER_PID $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null; rm /tmp/dfs-*.pid 2>/dev/null; exit 0" INT

wait