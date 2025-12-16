package main

import (
	"context"
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
		runs     = flag.Int("runs", 10, "number of runs")
		out      = flag.String("out", "experiments/deadlock.csv", "output CSV path (relative to backend/)")
		basePort = flag.Int("base-port", 9500, "base TCP port for run 0")
		timeout  = flag.Duration("timeout", 60*time.Second, "max time to wait for resolution")
	)
	flag.Parse()

	w, err := experiments.NewCSVWriter(*out, []string{"run", "resolve_ms"}, false)
	if err != nil {
		log.Fatalf("csv: %v", err)
	}
	defer w.Close()

	for run := 0; run < *runs; run++ {
		ctx, cancel := context.WithCancel(context.Background())
		portsBase := *basePort + run*30

		master := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", portsBase, common.RoleMaster, 10)
		a := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", portsBase+1, common.RoleStorage, 5)
		b := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", portsBase+2, common.RoleStorage, 1)

		nodes := []startedNode{{n: master, port: portsBase}, {n: a, port: portsBase + 1}, {n: b, port: portsBase + 2}}
		for i := range nodes {
			if err := nodes[i].n.Start(ctx); err != nil {
				stopAll(nodes)
				cancel()
				log.Fatalf("start: %v", err)
			}
		}

		// Fully connect peers (no discovery needed for local harness)
		for i := range nodes {
			for j := range nodes {
				if i == j {
					continue
				}
				nodes[i].n.AddPeer(nodes[j].n.GetNodeInfo())
			}
		}
		time.Sleep(500 * time.Millisecond)

		// Create deadlock:
		// A holds r1, B holds r2, then A requests r2 and B requests r1.
		r1, r2 := fmt.Sprintf("r1-%d", run), fmt.Sprintf("r2-%d", run)

		a.RequestLock(r1)
		waitUntil(func() bool { return a.CanEnterCriticalSection(r1) }, 5*time.Second)
		_ = a.EnterCriticalSection(r1)

		b.RequestLock(r2)
		waitUntil(func() bool { return b.CanEnterCriticalSection(r2) }, 5*time.Second)
		_ = b.EnterCriticalSection(r2)

		time.Sleep(200 * time.Millisecond)

		start := time.Now()
		a.RequestLock(r2)
		time.Sleep(100 * time.Millisecond)
		b.RequestLock(r1)

		deadline := time.Now().Add(*timeout)
		resolved := false
		for time.Now().Before(deadline) {
			// Victim should be B (lower priority), so A should eventually get r2.
			if a.CanEnterCriticalSection(r2) {
				_ = a.EnterCriticalSection(r2)
				resolved = true
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		resolveMs := int64(-1)
		if resolved {
			resolveMs = time.Since(start).Milliseconds()
		}

		_ = w.Write([]string{fmt.Sprintf("%d", run), fmt.Sprintf("%d", resolveMs)})

		// Cleanup
		a.ReleaseLock(r2)
		a.ReleaseLock(r1)
		b.ReleaseLock(r2)
		b.ReleaseLock(r1)
		stopAll(nodes)
		cancel()
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("Wrote %d runs to %s", *runs, *out)
}

func waitUntil(cond func() bool, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
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
