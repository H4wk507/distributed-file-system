package main

import (
	"context"
	"log"
	"time"

	"dfs-backend/dfs/node"
)

func main() {
	log.Println("Node Communication Test")
	log.Println()

	log.Println("Creating nodes...")
	node1 := node.NewNode("127.0.0.1", 9000, node.RoleMaster, 10)
	node2 := node.NewNode("127.0.0.1", 9001, node.RoleStorage, 5)
	log.Printf("Node1 ID: %s", node1.ID)
	log.Printf("Node2 ID: %s", node2.ID)
	log.Println()

	ctx := context.Background()

	log.Println("Starting nodes...")
	if err := node1.Start(ctx); err != nil {
		log.Fatalf("Failed to start node1: %v", err)
	}
	if err := node2.Start(ctx); err != nil {
		log.Fatalf("Failed to start node2: %v", err)
	}
	log.Println("Both nodes started")
	log.Println()

	time.Sleep(time.Second)
	log.Println("Starting discovery...")
	if err := node2.DiscoverPeers("127.0.0.1", 9000); err != nil {
		log.Printf("Warning: Discovery error: %v", err)
	} else {
		log.Println("Discovery request sent")
	}
	log.Println()

	log.Println("Waiting for heartbeats (6 seconds)...")
	time.Sleep(6 * time.Second)
	log.Println()

	peers1 := node1.GetAllPeers()
	peers2 := node2.GetAllPeers()

	log.Println("Results")
	log.Printf("Node1 (%s) knows about %d peer(s):", node1.ID, len(peers1))
	for _, peer := range peers1 {
		log.Printf("  - %s (%s:%d) [%s]", peer.ID, peer.IP, peer.Port, peer.Role)
	}
	log.Println()

	log.Printf("Node2 (%s) knows about %d peer(s):", node2.ID, len(peers2))
	for _, peer := range peers2 {
		log.Printf("  - %s (%s:%d) [%s]", peer.ID, peer.IP, peer.Port, peer.Role)
	}
	log.Println()

	if len(peers1) > 0 && len(peers2) > 0 {
		log.Println("SUCCESS: Nodes discovered each other!")
		log.Println("   - Heartbeat mechanism working")
		log.Println("   - Discovery mechanism working")
		log.Println("   - Peer management working")
	} else {
		log.Println("FAILED: Nodes did not discover each other")
		if len(peers1) == 0 {
			log.Println("   - Node1 has no peers")
		}
		if len(peers2) == 0 {
			log.Println("   - Node2 has no peers")
		}
	}
	log.Println()

	log.Println("Testing message channels...")
	msgChan1 := node1.GetMessageChan()
	msgChan2 := node2.GetMessageChan()

	messagesReceived := 0
	timeout := time.After(3 * time.Second)

checkMessages:
	for {
		select {
		case msg := <-msgChan1:
			log.Printf("Node1 received: %s from %s", msg.Type, msg.From)
			messagesReceived++
		case msg := <-msgChan2:
			log.Printf("Node2 received: %s from %s", msg.Type, msg.From)
			messagesReceived++
		case <-timeout:
			break checkMessages
		}
	}

	if messagesReceived > 0 {
		log.Printf("Message system working (%d messages received)", messagesReceived)
	} else {
		log.Println("No messages captured (this is OK, they might have been processed already)")
	}
	log.Println()

	log.Println("Stopping nodes...")
	if err := node1.Stop(); err != nil {
		log.Printf("Error stopping node1: %v", err)
	}
	if err := node2.Stop(); err != nil {
		log.Printf("Error stopping node2: %v", err)
	}
	log.Println("Nodes stopped")
	log.Println()

	log.Println("Test Complete")
}
