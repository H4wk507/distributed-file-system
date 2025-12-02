package main

import (
	"context"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/node"
	"dfs-backend/internal/database"
	"dfs-backend/internal/handlers"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db, err := database.Init("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	master := node.CreateNodeWithBully("127.0.0.1", 9000, common.RoleMaster, 100)
	if err := master.Start(ctx); err != nil {
		log.Fatalf("failed to start master node: %v", err)
	}

	storage1 := node.CreateNodeWithBully("127.0.0.1", 9001, common.RoleStorage, 1)
	if err := storage1.Start(ctx); err != nil {
		log.Fatalf("failed to start storage node 1: %v", err)
	}

	storage2 := node.CreateNodeWithBully("127.0.0.1", 9002, common.RoleStorage, 2)
	if err := storage2.Start(ctx); err != nil {
		log.Fatalf("failed to start storage node 2: %v", err)
	}

	master.RegisterStorageNode(storage1.GetNodeInfo())
	master.RegisterStorageNode(storage2.GetNodeInfo())

	storage1.AddPeer(master.GetNodeInfo())
	storage2.AddPeer(master.GetNodeInfo())

	authHandler := handlers.NewAuthHandler(db)
	fileHandler := handlers.NewFileHandler(db, master)

	router := http.NewServeMux()
	router.HandleFunc("POST /api/auth/register", authHandler.RegisterUser)
	router.HandleFunc("POST /api/auth/login", authHandler.LoginUser)

	router.HandleFunc("POST /api/files/upload/", fileHandler.UploadFile)
	router.HandleFunc("GET /api/files/{filename}/", fileHandler.GetFile)
	router.HandleFunc("DELETE /api/files/{filename}/", fileHandler.DeleteFile)

	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
