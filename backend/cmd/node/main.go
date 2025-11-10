package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dfs-backend/dfs/node"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "Node IP address")
	port := flag.Int("port", 9000, "Node port")
	role := flag.String("role", "storage", "Node role (master/storage)")
	priority := flag.Int("priority", 1, "Node priority (higher = stronger)")
	seedIP := flag.String("seed-ip", "", "Seed node IP for discovery")
	seedPort := flag.Int("seed-port", 9000, "Seed node port for discovery")
	flag.Parse()

	var nodeRole node.NodeRole
	switch *role {
	case "master":
		nodeRole = node.RoleMaster
	case "storage":
		nodeRole = node.RoleStorage
	default:
		log.Fatalf("Invalid role: %s (must be 'master' or 'storage')", *role)
	}

	n := node.NewNode(*ip, *port, nodeRole, *priority)
	log.Printf("Created node: ID=%s, Role=%s, Priority=%d", n.ID, nodeRole, *priority)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := n.Start(ctx); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}
	log.Printf("Node started successfully on %s:%d", *ip, *port)

	if *seedIP != "" {
		log.Printf("Attempting discovery from seed node %s:%d", *seedIP, *seedPort)
		if err := n.DiscoverPeers(*seedIP, *seedPort); err != nil {
			log.Printf("Warning: Discovery failed: %v", err)
		} else {
			peers := n.GetAllPeers()
			log.Printf("Discovery successful! Found %d peers", len(peers))
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigChan:
				return
			}
		}
	}()

	fmt.Println("\n=== Node Running ===")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Printf("Node ID: %s\n", n.ID)
	fmt.Printf("Address: %s:%d\n", *ip, *port)
	fmt.Printf("Role: %s\n", nodeRole)
	fmt.Printf("Priority: %d\n", *priority)
	fmt.Println("====================")

	<-sigChan
	log.Println("\nReceived shutdown signal")

	cancel()
	if err := n.Stop(); err != nil {
		log.Fatalf("Error stopping node: %v", err)
	}

	log.Println("Node stopped successfully")
}
