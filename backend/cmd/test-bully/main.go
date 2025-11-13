package main

import (
	"context"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/node"
	"log"
	"time"
)

func printCoordinatorStatus(node1, node2, node3 *node.Node) {
	if node1 != nil {
		info1 := node1.GetNodeInfo()
		log.Printf("Node1 (Priority 10): Role=%s, Status=%s", info1.Role, info1.Status)

		peers := node1.GetAllPeers()
		for _, peer := range peers {
			log.Printf("  -> Peer %s: Priority=%d, Role=%s", peer.ID, peer.Priority, peer.Role)
		}
	}

	if node2 != nil {
		info2 := node2.GetNodeInfo()
		log.Printf("Node2 (Priority 5): Role=%s, Status=%s", info2.Role, info2.Status)

		peers := node2.GetAllPeers()
		for _, peer := range peers {
			log.Printf("  -> Peer %s: Priority=%d, Role=%s", peer.ID, peer.Priority, peer.Role)
		}
	}

	if node3 != nil {
		info3 := node3.GetNodeInfo()
		log.Printf("Node3 (Priority 1): Role=%s, Status=%s", info3.Role, info3.Status)

		peers := node3.GetAllPeers()
		for _, peer := range peers {
			log.Printf("  -> Peer %s: Priority=%d, Role=%s", peer.ID, peer.Priority, peer.Role)
		}
	}
}

func main() {
	log.Println("Bully algorithm test")
	log.Println()

	node1 := node.CreateNodeWithBully("127.0.0.1", 9000, common.RoleMaster, 10)
	node2 := node.CreateNodeWithBully("127.0.0.1", 9001, common.RoleStorage, 5)
	node3 := node.CreateNodeWithBully("127.0.0.1", 9002, common.RoleStorage, 1)

	ctx := context.Background()

	if err := node1.Start(ctx); err != nil {
		log.Fatalf("Failed to start node1: %v", err)
	}
	if err := node2.Start(ctx); err != nil {
		log.Fatalf("Failed to start node2: %v", err)
	}
	if err := node3.Start(ctx); err != nil {
		log.Fatalf("Failed to start node3: %v", err)
	}

	if err := node2.DiscoverPeers("127.0.0.1", 9000); err != nil {
		log.Printf("Warning: Discovery error on node2: %v", err)
	}
	if err := node3.DiscoverPeers("127.0.0.1", 9000); err != nil {
		log.Printf("Warning: Discovery error on node3: %v", err)
	}
	if err := node3.DiscoverPeers("127.0.0.1", 9001); err != nil {
		log.Printf("Warning: Discovery error on node3: %v", err)
	}

	log.Println("Waiting for heartbeats (6 seconds)...")
	time.Sleep(6 * time.Second)
	log.Println()

	peers1 := node1.GetAllPeers()
	peers2 := node2.GetAllPeers()
	peers3 := node3.GetAllPeers()

	if len(peers1) != 2 || len(peers2) != 2 || len(peers3) != 2 {
		log.Fatalf("Peer discovery failed: node1 has %d peers, node2 has %d peers, node3 has %d peers", len(peers1), len(peers2), len(peers3))
	}

	log.Println("Peer discovery successfull")
	printCoordinatorStatus(node1, node2, node3)

	log.Println("Stopping node1")
	if err := node1.Stop(); err != nil {
		log.Fatalf("Failed to stop node1: %v", err)
	}

	log.Println("Waiting for election (20 seconds)...")
	time.Sleep(20 * time.Second)

	log.Println("After election:")
	printCoordinatorStatus(nil, node2, node3)

	log.Println("Stopping node2")
	if err := node2.Stop(); err != nil {
		log.Fatalf("Failed to stop node2: %v", err)
	}

	log.Println("Waiting for election (20 seconds)...")
	time.Sleep(20 * time.Second)

	log.Println("After election:")
	printCoordinatorStatus(nil, nil, node3)
}
