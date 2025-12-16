package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"dfs-backend/dfs/client"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/node"
	"dfs-backend/internal/experiments"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
)

type startedNode struct {
	n    *node.Node
	port int
}

func main() {
	var (
		runs      = flag.Int("runs", 10, "number of runs")
		out       = flag.String("out", "experiments/streaming.csv", "output CSV path (relative to backend/)")
		basePort  = flag.Int("base-port", 9600, "base TCP port for run 0")
		sizeMB    = flag.Int("size-mb", 16, "file size in MB")
		stableWait = flag.Duration("stable-wait", 4*time.Second, "wait for cluster to stabilize")
	)
	flag.Parse()

	w, err := experiments.NewCSVWriter(*out, []string{"run", "size_bytes", "upload_ms", "upload_mbps", "download_ms", "download_mbps"}, false)
	if err != nil {
		log.Fatalf("csv: %v", err)
	}
	defer w.Close()

	size := int64(*sizeMB) * 1024 * 1024

	for run := 0; run < *runs; run++ {
		ctx, cancel := context.WithCancel(context.Background())
		portsBase := *basePort + run*40

		master := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", portsBase, common.RoleMaster, 10)
		s1 := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", portsBase+1, common.RoleStorage, 5)
		s2 := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", portsBase+2, common.RoleStorage, 3)
		s3 := node.CreateNodeWithBully("127.0.0.1", "127.0.0.1", portsBase+3, common.RoleStorage, 1)

		nodes := []startedNode{{n: master, port: portsBase}, {n: s1, port: portsBase + 1}, {n: s2, port: portsBase + 2}, {n: s3, port: portsBase + 3}}
		for i := range nodes {
			if err := nodes[i].n.Start(ctx); err != nil {
				stopAll(nodes)
				cancel()
				log.Fatalf("start: %v", err)
			}
		}

		// IMPORTANT: use the real discovery/join flow so the master registers storage nodes
		// and adds them to the hash ring. Manual AddPeer() doesn't populate the master's ring.
		_ = s1.DiscoverPeers("127.0.0.1", portsBase)
		_ = s2.DiscoverPeers("127.0.0.1", portsBase)
		_ = s3.DiscoverPeers("127.0.0.1", portsBase)

		time.Sleep(*stableWait)

		seeds := []client.NodeAddr{{IP: "127.0.0.1", Port: portsBase}, {IP: "127.0.0.1", Port: portsBase + 1}, {IP: "127.0.0.1", Port: portsBase + 2}, {IP: "127.0.0.1", Port: portsBase + 3}}
		c := client.NewMasterClientWithSeeds(seeds)

		data := make([]byte, size)
		if _, err := io.ReadFull(rand.Reader, data); err != nil {
			stopAll(nodes)
			cancel()
			log.Fatalf("random: %v", err)
		}

		fileID := uuid.New()
		startUp := time.Now()
		res, err := c.UploadFileStream(fileID, fmt.Sprintf("bench-%d.bin", run), "application/octet-stream", size, bytes.NewReader(data))
		uploadDur := time.Since(startUp)
		if err != nil {
			log.Printf("upload failed: %v", err)
			_ = w.Write([]string{fmt.Sprintf("%d", run), fmt.Sprintf("%d", size), "-1", "-1", "-1", "-1"})
			stopAll(nodes)
			cancel()
			continue
		}

		uploadMbps := mbps(size, uploadDur)

		startDown := time.Now()
		_ = c.DownloadFileStream(fileID, res.Hash, io.Discard)
		downloadDur := time.Since(startDown)
		downloadMbps := mbps(size, downloadDur)

		_ = w.Write([]string{
			fmt.Sprintf("%d", run),
			fmt.Sprintf("%d", size),
			fmt.Sprintf("%d", uploadDur.Milliseconds()),
			fmt.Sprintf("%.2f", uploadMbps),
			fmt.Sprintf("%d", downloadDur.Milliseconds()),
			fmt.Sprintf("%.2f", downloadMbps),
		})

		stopAll(nodes)
		cancel()
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("Wrote %d runs to %s", *runs, *out)
}

func mbps(bytes int64, d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return (float64(bytes) * 8 / 1_000_000) / d.Seconds()
}

func stopAll(nodes []startedNode) {
	for i := range nodes {
		if nodes[i].n != nil {
			_ = nodes[i].n.Stop()
			nodes[i].n = nil
		}
	}
}
