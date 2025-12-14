package main

import (
	"dfs-backend/dfs/client"
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

	// Configure seed nodes for master discovery
	// Primary seed is from config, additional seeds can be added for failover
	seeds := []client.NodeAddr{{IP: cfg.MasterHost, Port: cfg.MasterPort}}
	// TODO: Add additional seed nodes from config (e.g., cfg.SeedNodes)

	masterClient := client.NewMasterClientWithSeeds(seeds)
	log.Printf("Configured master client with %d seed node(s)", len(seeds))

	if err := masterClient.Ping(); err != nil {
		log.Printf("Warning: No master reachable at startup: %v", err)
		log.Printf("The API will attempt to discover master when handling requests")
	} else {
		log.Printf("Master node is reachable")
	}

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	fileHandler := handlers.NewFileHandler(db, masterClient)

	router := http.NewServeMux()

	router.HandleFunc("POST /api/auth/register", authHandler.RegisterUser)
	router.HandleFunc("POST /api/auth/login", authHandler.LoginUser)

	router.Handle("GET /api/auth/me", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.GetMe)))

	router.Handle("GET /api/files", authMiddleware.RequireAuth(http.HandlerFunc(fileHandler.ListFiles)))
	router.Handle("POST /api/files/upload/", authMiddleware.RequireAuth(http.HandlerFunc(fileHandler.UploadFile)))
	router.Handle("GET /api/files/{fileID}/", authMiddleware.RequireAuth(http.HandlerFunc(fileHandler.GetFile)))
	router.Handle("GET /api/files/{fileID}/metadata", authMiddleware.RequireAuth(http.HandlerFunc(fileHandler.GetFileMetadata)))
	router.Handle("DELETE /api/files/{fileID}/", authMiddleware.RequireAuth(http.HandlerFunc(fileHandler.DeleteFile)))

	handler := corsMiddleware(router)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	fmt.Printf("Server starting on http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Must be that strict, because of http-only cookies, otherwise won't work
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
