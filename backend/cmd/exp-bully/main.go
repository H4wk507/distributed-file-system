package main

import (
	"context"
	"dfs-backend/dfs/client"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/node"
	"dfs-backend/internal/experiments"
	"flag"
	"fmt"
	"log"
	"time"
)

type startedNode struct {
	n    *node.Node
	port int
}

func main() {
	var (
		runs          = flag.Int("runs", 10, "number of runs")
		out           = flag.String("out", "experiments/bully.csv", "output CSV path (relative to backend/)")
		basePort      = flag.Int("base-port", 9300, "base TCP port for run 0")
		heartbeatWait = flag.Duration("heartbeat-wait", 6*time.Second, "wait for heartbeats/peer discovery")
		timeout       = flag.Duration("timeout", 45*time.Second, "max time to wait for failover")
	)
	flag.Parse()

	w, err := experiments.NewCSVWriter(*out, []string{"run", "nodes", "failover_ms", "new_master_priority", "new_master_port"}, false)
	if err != nil {
		log.Fatalf("csv: %v", err)
	}
	defer w.Close()

	for run := 0; run < *runs; run++ {
		ctx, cancel := context.WithCancel(context.Background())

		portsBase := *basePort + run*20
		nodes, seeds := startBullyCluster(ctx, portsBase)
		time.Sleep(*heartbeatWait)

		c := client.NewMasterClientWithSeeds(seeds)
		_ = c.Ping()

		// Stop original master
		var master *startedNode
		for i := range nodes {
			if nodes[i].n.GetRole() == common.RoleMaster {
				master = &nodes[i]
				break
			}
		}
		if master == nil {
			stopAll(nodes)
			cancel()
			log.Fatalf("run %d: no master found", run)
		}

		_ = master.n.Stop()

		start := time.Now()
		deadline := start.Add(*timeout)
		for time.Now().Before(deadline) {
			if err := c.Ping(); err == nil {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		failover := time.Since(start)

		newMasterPriority := -1
		newMasterPort := -1
		deadline2 := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline2) {
			for i := range nodes {
				if nodes[i].n != nil && nodes[i].port != master.port && nodes[i].n.GetRole() == common.RoleMaster {
					newMasterPriority = nodes[i].n.GetPriority()
					newMasterPort = nodes[i].port
					break
				}
			}
			if newMasterPort != -1 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		_ = w.Write([]string{
			fmt.Sprintf("%d", run),
			fmt.Sprintf("%d", len(nodes)),
			fmt.Sprintf("%d", failover.Milliseconds()),
			fmt.Sprintf("%d", newMasterPriority),
			fmt.Sprintf("%d", newMasterPort),
		})

		stopAll(nodes)
		cancel()
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("Wrote %d runs to %s", *runs, *out)
}

func startBullyCluster(ctx context.Context, basePort int) ([]startedNode, []client.NodeAddr) {
	master := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", basePort, common.RoleMaster, 10)
	n2 := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", basePort+1, common.RoleStorage, 5)
	n3 := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", basePort+2, common.RoleStorage, 1)

	nodes := []startedNode{{n: master, port: basePort}, {n: n2, port: basePort + 1}, {n: n3, port: basePort + 2}}
	for i := range nodes {
		if err := nodes[i].n.Start(ctx); err != nil {
			stopAll(nodes)
			log.Fatalf("start node on %d: %v", nodes[i].port, err)
		}
	}

	// Discovery: storage nodes discover master, then each other.
	_ = n2.DiscoverPeers("127.0.0.1", basePort)
	_ = n3.DiscoverPeers("127.0.0.1", basePort)
	_ = n3.DiscoverPeers("127.0.0.1", basePort+1)

	seeds := []client.NodeAddr{{IP: "127.0.0.1", Port: basePort}, {IP: "127.0.0.1", Port: basePort + 1}, {IP: "127.0.0.1", Port: basePort + 2}}
	return nodes, seeds
}

func stopAll(nodes []startedNode) {
	for i := range nodes {
		if nodes[i].n != nil {
			_ = nodes[i].n.Stop()
			nodes[i].n = nil
		}
	}
}
