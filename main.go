package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"videoforge/api"
	"videoforge/config"
	"videoforge/database"
	"videoforge/models"
	"videoforge/websocket"
	"videoforge/worker"
)

func main() {
	// 加载配置
	if err := config.LoadConfig("config.json"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := database.NewDB(config.GlobalConfig.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 创建 WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 创建任务队列
	queue := worker.NewTaskQueue(db, config.GlobalConfig.FFmpeg.Path, config.GlobalConfig.FFmpeg.Threads, func(update models.ProgressUpdate) {
		hub.Broadcast(update)
	})
	queue.Start()

	// 创建 API 服务器
	apiServer := api.NewServer(db, queue)

	// 设置路由
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/browse", apiServer.BrowseDirectory)
	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			apiServer.GetTasks(w, r)
		case http.MethodPost:
			apiServer.CreateTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/tasks/batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			apiServer.BatchCreateTasks(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			apiServer.GetTask(w, r)
		case http.MethodDelete:
			apiServer.DeleteTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/files/", apiServer.ServeFile)

	// WebSocket 路由
	mux.HandleFunc("/ws", hub.HandleWebSocket)

	// 静态文件服务
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.GlobalConfig.Server.Host, config.GlobalConfig.Server.Port)
	log.Printf("Starting VideoForge server on http://%s", addr)
	log.Printf("FFmpeg path: %s", config.GlobalConfig.FFmpeg.Path)
	log.Printf("FFmpeg threads: %d", config.GlobalConfig.FFmpeg.Threads)
	log.Printf("Database: %s", config.GlobalConfig.Database.Path)

	// 确保必要的目录存在
	os.MkdirAll(config.GlobalConfig.FFmpeg.DefaultOutputDir, 0755)
	os.MkdirAll(config.GlobalConfig.VideoRootDir, 0755)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
