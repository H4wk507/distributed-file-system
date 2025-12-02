package main

import (
	"bytes"
	"context"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/node"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func printStatus(nodes ...*node.Node) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
	for i, n := range nodes {
		if n == nil {
			fmt.Printf("  Node %d: STOPPED\n", i+1)
			continue
		}
		info := n.GetNodeInfo()
		fmt.Printf("  Node %d: Role=%s, Priority=%d, Peers=%d\n",
			i+1, info.Role, info.Priority, len(n.GetPeers()))
	}
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
}

func main() {
	log.Println("═══════════════════════════════════════════════════════")
	log.Println("        METADATA SYNC AFTER ELECTION TEST")
	log.Println("═══════════════════════════════════════════════════════")
	log.Println()

	ctx := context.Background()

	// Create nodes
	log.Println("Step 1: Creating nodes...")
	master := node.CreateNodeWithBully("127.0.0.1", 9100, common.RoleMaster, 10)
	storage1 := node.CreateNodeWithBully("127.0.0.1", 9101, common.RoleStorage, 5)
	storage2 := node.CreateNodeWithBully("127.0.0.1", 9102, common.RoleStorage, 3)
	storage3 := node.CreateNodeWithBully("127.0.0.1", 9103, common.RoleStorage, 1)

	// Start all nodes
	log.Println("Step 2: Starting nodes...")
	if err := master.Start(ctx); err != nil {
		log.Fatalf("Failed to start master: %v", err)
	}
	if err := storage1.Start(ctx); err != nil {
		log.Fatalf("Failed to start storage1: %v", err)
	}
	if err := storage2.Start(ctx); err != nil {
		log.Fatalf("Failed to start storage2: %v", err)
	}
	if err := storage3.Start(ctx); err != nil {
		log.Fatalf("Failed to start storage3: %v", err)
	}

	// Discover peers - do it in stages to ensure complete peer lists
	log.Println("Step 3: Discovering peers...")
	time.Sleep(1 * time.Second)

	// First round: all storage nodes discover from master
	storage1.DiscoverPeers("127.0.0.1", 9100)
	storage2.DiscoverPeers("127.0.0.1", 9100)
	storage3.DiscoverPeers("127.0.0.1", 9100)

	time.Sleep(1 * time.Second)

	// Second round: discover from each other to complete peer lists
	storage1.DiscoverPeers("127.0.0.1", 9102) // storage1 from storage2
	storage1.DiscoverPeers("127.0.0.1", 9103) // storage1 from storage3
	storage2.DiscoverPeers("127.0.0.1", 9101) // storage2 from storage1
	storage2.DiscoverPeers("127.0.0.1", 9103) // storage2 from storage3
	storage3.DiscoverPeers("127.0.0.1", 9101) // storage3 from storage1

	// Wait for discovery and heartbeats
	log.Println("Step 4: Waiting for cluster to stabilize (6s)...")
	time.Sleep(6 * time.Second)

	printStatus(master, storage1, storage2, storage3)

	// Upload some test files through master
	log.Println("Step 5: Uploading test files to storage nodes...")
	testFiles := []struct {
		name    string
		content string
	}{
		{"test-file-1.txt", "Content of file 1"},
		{"test-file-2.txt", "Content of file 2"},
		{"test-file-3.txt", "Content of file 3"},
	}

	for _, tf := range testFiles {
		fileID := uuid.New()
		data := []byte(tf.content)
		master.StoreFile(fileID, tf.name, "text/plain", data)
		log.Printf("  Uploaded: %s (ID: %s)", tf.name, fileID)
	}

	// Wait for files to be stored
	log.Println("Step 6: Waiting for files to be replicated (3s)...")
	time.Sleep(3 * time.Second)

	// Manually store a file directly on storage1 (to simulate pre-existing data)
	log.Println("Step 7: Storing file directly on storage1 (simulate pre-existing data)...")
	directFileID := uuid.New()
	directData := bytes.NewReader([]byte("Direct content on storage1"))
	_, err := storage1.GetStorage().SaveFile(directFileID, "direct-file.txt", "text/plain", directData)
	if err != nil {
		log.Printf("  Warning: Failed to save direct file: %v", err)
	} else {
		log.Printf("  Stored: direct-file.txt (ID: %s)", directFileID)
	}

	// Print storage status
	log.Println()
	log.Println("Storage node file counts:")
	log.Printf("  Storage1: %d files", len(storage1.GetStorage().GetAllMetadata()))
	log.Printf("  Storage2: %d files", len(storage2.GetStorage().GetAllMetadata()))
	log.Printf("  Storage3: %d files", len(storage3.GetStorage().GetAllMetadata()))
	log.Println()

	// Kill the master
	log.Println("Step 8: Killing the master node...")
	if err := master.Stop(); err != nil {
		log.Fatalf("Failed to stop master: %v", err)
	}
	master = nil

	printStatus(master, storage1, storage2, storage3)

	// Wait for election and metadata sync
	log.Println("Step 9: Waiting for election + metadata sync (20s)...")
	log.Println("        (storage1 should become new master and sync metadata)")
	time.Sleep(20 * time.Second)

	printStatus(master, storage1, storage2, storage3)

	// Check who became master
	log.Println("Step 10: Checking election results...")
	newMaster := findMaster(storage1, storage2, storage3)
	if newMaster == nil {
		log.Println("  ERROR: No new master was elected!")
	} else {
		info := newMaster.GetNodeInfo()
		log.Printf("  SUCCESS: New master elected (Priority=%d)", info.Priority)

		// Check if metadata was synced
		log.Println()
		log.Println("Step 11: Verifying metadata sync...")
		globalIndex := newMaster.GetGlobalFileIndex()
		if globalIndex == nil {
			log.Println("  WARNING: globalFileIndex is nil")
		} else {
			log.Printf("  Global file index has %d files:", len(globalIndex))
			for hash, fileInfo := range globalIndex {
				log.Printf("    - %s (hash: %s..., replicas: %d)",
					fileInfo.Filename, hash[:8], len(fileInfo.Replicas))
			}
		}
	}

	// Cleanup
	log.Println()
	log.Println("Step 12: Stopping remaining nodes...")
	if storage1 != nil {
		storage1.Stop()
	}
	if storage2 != nil {
		storage2.Stop()
	}
	if storage3 != nil {
		storage3.Stop()
	}

	log.Println()
	log.Println("═══════════════════════════════════════════════════════")
	log.Println("                    TEST COMPLETE")
	log.Println("═══════════════════════════════════════════════════════")
}

func findMaster(nodes ...*node.Node) *node.Node {
	for _, n := range nodes {
		if n != nil && n.GetRole() == common.RoleMaster {
			return n
		}
	}
	return nil
}
