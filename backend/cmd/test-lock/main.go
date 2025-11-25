package main

import (
	"context"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/node"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()

	fmt.Println("\ntest 1: single node")
	node1 := node.CreateNodeWithBully("localhost", 8001, common.RoleStorage, 1)
	node1.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	node1.RequestLock("file.txt")
	time.Sleep(200 * time.Millisecond)

	if node1.CanEnterCriticalSection("file.txt") {
		fmt.Println("[X] single node can enter")
		node1.ReleaseLock("file.txt")
	} else {
		fmt.Println("ERROR: single node cannot enter")
	}

	fmt.Println("\ntest 2: five nodes")
	node2 := node.CreateNodeWithBully("localhost", 8002, common.RoleStorage, 2)
	node3 := node.CreateNodeWithBully("localhost", 8003, common.RoleStorage, 3)
	node4 := node.CreateNodeWithBully("localhost", 8004, common.RoleStorage, 4)
	node5 := node.CreateNodeWithBully("localhost", 8005, common.RoleStorage, 5)
	node2.Start(ctx)
	node3.Start(ctx)
	node4.Start(ctx)
	node5.Start(ctx)
	time.Sleep(200 * time.Millisecond)

	nodes := []*node.Node{node1, node2, node3, node4, node5}
	for i, n := range nodes {
		for j, peer := range nodes {
			if i != j {
				n.AddPeer(peer.GetNodeInfo())
			}
		}
	}
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\nAll nodes requesting lock on 'shared.txt'...")
	node1.RequestLock("shared.txt")
	time.Sleep(50 * time.Millisecond)

	node3.RequestLock("shared.txt")
	time.Sleep(30 * time.Millisecond)

	node5.RequestLock("shared.txt")
	time.Sleep(20 * time.Millisecond)

	node2.RequestLock("shared.txt")
	time.Sleep(40 * time.Millisecond)

	node4.RequestLock("shared.txt")
	time.Sleep(1 * time.Second) // wait for acks

	fmt.Println("\n--- Expected order based on timestamps ---")
	fmt.Println("1. Node1 (requested first)")
	fmt.Println("2. Node3")
	fmt.Println("3. Node5")
	fmt.Println("4. Node2")
	fmt.Println("5. Node4 (requested last)")
	fmt.Println("------------------------------------------")

	if node1.CanEnterCriticalSection("shared.txt") {
		fmt.Println("[1/5] Node1 enters critical section")
		time.Sleep(500 * time.Millisecond)
		node1.ReleaseLock("shared.txt")
	} else {
		fmt.Println("ERROR: Node1 should enter first")
	}

	time.Sleep(300 * time.Millisecond)

	if node3.CanEnterCriticalSection("shared.txt") {
		fmt.Println("[2/5] Node3 enters critical section")
		time.Sleep(500 * time.Millisecond)
		node3.ReleaseLock("shared.txt")
	} else {
		fmt.Println("ERROR: Node3 should enter second")
	}

	time.Sleep(300 * time.Millisecond)

	if node5.CanEnterCriticalSection("shared.txt") {
		fmt.Println("[3/5] Node5 enters critical section")
		time.Sleep(500 * time.Millisecond)
		node5.ReleaseLock("shared.txt")
	} else {
		fmt.Println("ERROR: Node5 should enter third")
	}

	time.Sleep(300 * time.Millisecond)

	if node2.CanEnterCriticalSection("shared.txt") {
		fmt.Println("[4/5] Node2 enters critical section")
		time.Sleep(500 * time.Millisecond)
		node2.ReleaseLock("shared.txt")
	} else {
		fmt.Println("ERROR: Node2 should enter fourth")
	}

	time.Sleep(300 * time.Millisecond)

	if node4.CanEnterCriticalSection("shared.txt") {
		fmt.Println("[5/5] Node4 enters critical section (last)")
		time.Sleep(500 * time.Millisecond)
		node4.ReleaseLock("shared.txt")
	} else {
		fmt.Println("ERROR: Node4 should enter last")
	}

	time.Sleep(1 * time.Second)
	node1.Stop()
	node2.Stop()
	node3.Stop()
	node4.Stop()
	node5.Stop()
	fmt.Println("Done.")
}
