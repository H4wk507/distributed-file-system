package main

import (
	"context"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/node"
	"dfs-backend/internal/config"
	"dfs-backend/internal/database"
	"dfs-backend/internal/handlers"
	"dfs-backend/internal/middleware"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(cfg.DatabaseURL)
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

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	fileHandler := handlers.NewFileHandler(db, master)

	router := http.NewServeMux()

	router.HandleFunc("POST /api/auth/register", authHandler.RegisterUser)
	router.HandleFunc("POST /api/auth/login", authHandler.LoginUser)

	router.Handle("GET /api/auth/me", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.GetMe)))

	router.Handle("POST /api/files/upload/", authMiddleware.RequireAuth(http.HandlerFunc(fileHandler.UploadFile)))
	router.Handle("GET /api/files/{filename}/", authMiddleware.RequireAuth(http.HandlerFunc(fileHandler.GetFile)))
	router.Handle("DELETE /api/files/{filename}/", authMiddleware.RequireAuth(http.HandlerFunc(fileHandler.DeleteFile)))

	handler := corsMiddleware(router)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	fmt.Printf("Server starting on http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Must be more strict, because of http-only cookies, otherwise won't work
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
