package main

import (
	"context"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/node"
	"dfs-backend/internal/experiments"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type startedNode struct {
	n        *node.Node
	port     int
	priority int
}

func main() {
	var (
		runs     = flag.Int("runs", 10, "number of runs")
		nodesN   = flag.Int("nodes", 5, "number of nodes")
		out      = flag.String("out", "experiments/lock.csv", "output CSV path (relative to backend/)")
		basePort = flag.Int("base-port", 9400, "base TCP port for run 0")
		hold     = flag.Duration("hold", 200*time.Millisecond, "how long to hold lock in critical section")
	)
	flag.Parse()

	w, err := experiments.NewCSVWriter(*out, []string{"run", "node_idx", "priority", "wait_ms", "order"}, false)
	if err != nil {
		log.Fatalf("csv: %v", err)
	}
	defer w.Close()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for run := 0; run < *runs; run++ {
		ctx, cancel := context.WithCancel(context.Background())
		portsBase := *basePort + run*50

		nodes := startLockCluster(ctx, *nodesN, portsBase)
		connectAllPeers(nodes)
		time.Sleep(200 * time.Millisecond)

		resource := fmt.Sprintf("shared-%d.txt", rnd.Int())

		// All nodes request lock quickly.
		for i := range nodes {
			nodes[i].n.RequestLock(resource)
		}

		order := 0
		completed := make([]bool, len(nodes))
		startTimes := make([]time.Time, len(nodes))
		for i := range nodes {
			startTimes[i] = time.Now()
		}

		deadline := time.Now().Add(30 * time.Second)
		for order < len(nodes) && time.Now().Before(deadline) {
			progress := false
			for i := range nodes {
				if completed[i] {
					continue
				}
				if nodes[i].n.CanEnterCriticalSection(resource) {
					_ = nodes[i].n.EnterCriticalSection(resource)
					wait := time.Since(startTimes[i])
					order++
					completed[i] = true
					_ = w.Write([]string{
						fmt.Sprintf("%d", run),
						fmt.Sprintf("%d", i),
						fmt.Sprintf("%d", nodes[i].priority),
						fmt.Sprintf("%d", wait.Milliseconds()),
						fmt.Sprintf("%d", order),
					})
					time.Sleep(*hold)
					nodes[i].n.ReleaseLock(resource)
					progress = true
				}
			}
			if !progress {
				time.Sleep(10 * time.Millisecond)
			}
		}

		stopAll(nodes)
		cancel()
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("Wrote %d runs to %s", *runs, *out)
}

func startLockCluster(ctx context.Context, n int, basePort int) []startedNode {
	nodes := make([]startedNode, 0, n)
	for i := 0; i < n; i++ {
		priority := i + 1
		n := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", basePort+i, common.RoleStorage, priority)
		if err := n.Start(ctx); err != nil {
			stopAll(nodes)
			log.Fatalf("start node %d: %v", i, err)
		}
		nodes = append(nodes, startedNode{n: n, port: basePort + i, priority: priority})
	}
	return nodes
}

func connectAllPeers(nodes []startedNode) {
	for i := range nodes {
		for j := range nodes {
			if i == j {
				continue
			}
			nodes[i].n.AddPeer(nodes[j].n.GetNodeInfo())
		}
	}
}

func stopAll(nodes []startedNode) {
	for i := range nodes {
		if nodes[i].n != nil {
			_ = nodes[i].n.Stop()
			nodes[i].n = nil
		}
	}
}
