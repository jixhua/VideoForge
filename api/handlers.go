package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"videoforge/config"
	"videoforge/database"
	"videoforge/ffmpeg"
	"videoforge/models"
	"videoforge/worker"
)

type Server struct {
	db    *database.DB
	queue *worker.TaskQueue
}

func NewServer(db *database.DB, queue *worker.TaskQueue) *Server {
	return &Server{
		db:    db,
		queue: queue,
	}
}

// BrowseDirectory 浏览目录结构
func (s *Server) BrowseDirectory(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		// 尝试从数据库获取上次浏览的目录
		lastPath, err := s.db.GetSetting("last_browsed_directory")
		if err == nil && lastPath != "" {
			path = lastPath
		} else {
			path = config.GlobalConfig.VideoRootDir
		}
	}

	// 安全检查：防止目录遍历攻击
	absPath, err := filepath.Abs(path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		respondError(w, http.StatusNotFound, "Path not found")
		return
	}

	if !fileInfo.IsDir() {
		respondError(w, http.StatusBadRequest, "Path is not a directory")
		return
	}

	// 保存当前浏览的目录
	_ = s.db.SetSetting("last_browsed_directory", absPath)

	entries, err := os.ReadDir(absPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read directory")
		return
	}

	type FileEntry struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		IsDir   bool   `json:"isDir"`
		Size    int64  `json:"size"`
		IsVideo bool   `json:"isVideo"`
	}

	var files []FileEntry
	for _, entry := range entries {
		info, _ := entry.Info()
		fullPath := filepath.Join(absPath, entry.Name())

		isVideo := false
		if !entry.IsDir() {
			isVideo = ffmpeg.IsVideoFile(entry.Name())
		}

		files = append(files, FileEntry{
			Name:    entry.Name(),
			Path:    fullPath,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			IsVideo: isVideo,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"path":  absPath,
		"files": files,
	})
}

// CreateTask 创建任务
func (s *Server) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InputPath      string          `json:"inputPath"`
		OutputPath     string          `json:"outputPath"`
		Type           models.TaskType `json:"type"`
		Params         interface{}     `json:"params"`
		DeleteOriginal bool            `json:"deleteOriginal"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 验证输入文件存在
	if _, err := os.Stat(req.InputPath); err != nil {
		respondError(w, http.StatusBadRequest, "Input file not found")
		return
	}

	paramsJSON, _ := json.Marshal(req.Params)

	outputPath := req.OutputPath
	if outputPath == "" {
		outputPath = generateOutputPath(req.InputPath, req.Type, config.GlobalConfig.FFmpeg.DefaultOutputDir, req.Params)
	}

	task := &models.Task{
		InputPath:      req.InputPath,
		OutputPath:     req.OutputPath,
		Type:           req.Type,
		Params:         string(paramsJSON),
		DeleteOriginal: req.DeleteOriginal,
		Status:         models.TaskStatusPending,
	}

	if err := s.queue.AddTask(task); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	respondJSON(w, http.StatusCreated, task)
}

// BatchCreateTasks 批量创建任务
func (s *Server) BatchCreateTasks(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Directory      string          `json:"directory"`
		Recursive      bool            `json:"recursive"`
		Type           models.TaskType `json:"type"`
		Params         interface{}     `json:"params"`
		DeleteOriginal bool            `json:"deleteOriginal"`
		OutputDir      string          `json:"outputDir"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 查找所有视频文件
	videoFiles, err := findVideoFiles(req.Directory, req.Recursive)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to scan directory")
		return
	}

	if len(videoFiles) == 0 {
		respondError(w, http.StatusBadRequest, "No video files found")
		return
	}

	paramsJSON, _ := json.Marshal(req.Params)

	var createdTasks []*models.Task
	for _, videoFile := range videoFiles {
		outputPath := generateOutputPath(videoFile, req.Type, req.OutputDir, req.Params)

		task := &models.Task{
			InputPath:      videoFile,
			OutputPath:     outputPath,
			Type:           req.Type,
			Params:         string(paramsJSON),
			DeleteOriginal: req.DeleteOriginal,
			Status:         models.TaskStatusPending,
		}

		if err := s.queue.AddTask(task); err != nil {
			log.Printf("Failed to add task for %s: %v", videoFile, err)
			continue
		}

		createdTasks = append(createdTasks, task)
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"count": len(createdTasks),
		"tasks": createdTasks,
	})
}

// GetTasks 获取所有任务
func (s *Server) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.GetAllTasks()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get tasks")
		return
	}

	respondJSON(w, http.StatusOK, tasks)
}

// GetTask 获取单个任务
func (s *Server) GetTask(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := s.db.GetTask(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Task not found")
		return
	}

	respondJSON(w, http.StatusOK, task)
}

// DeleteTask 删除任务
func (s *Server) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// 获取任务信息，用于删除输出文件
	task, err := s.db.GetTask(id)
	if err != nil {
		// 任务不存在也继续尝试取消
		log.Printf("Failed to get task %d: %v", id, err)
	}

	// 先尝试取消并杀掉 FFmpeg 进程（如果正在运行）
	if cancelErr := s.queue.CancelTask(id); cancelErr != nil {
		log.Printf("Failed to cancel task %d: %v", id, cancelErr)
	}

	// 等待一小段时间，确保进程已完全终止并释放文件句柄
	time.Sleep(100 * time.Millisecond)

	// 删除输出文件（如果存在）
	if task != nil && task.OutputPath != "" {
		if err := os.RemoveAll(task.OutputPath); err != nil {
			log.Printf("Failed to delete output file %s: %v", task.OutputPath, err)
		} else {
			log.Printf("Deleted output file: %s", task.OutputPath)
		}
	}

	if err := s.db.DeleteTask(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Task deleted"})
}

// ServeFile 提供文件访问
func (s *Server) ServeFile(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/api/files/")

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if _, err := os.Stat(absPath); err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, absPath)
}

// 辅助函数

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func findVideoFiles(directory string, recursive bool) ([]string, error) {
	var videoFiles []string

	if recursive {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && ffmpeg.IsVideoFile(path) {
				videoFiles = append(videoFiles, path)
			}
			return nil
		})
		return videoFiles, err
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			fullPath := filepath.Join(directory, entry.Name())
			if ffmpeg.IsVideoFile(fullPath) {
				videoFiles = append(videoFiles, fullPath)
			}
		}
	}

	return videoFiles, nil
}

func generateOutputPath(inputPath string, taskType models.TaskType, outputDir string, params interface{}) string {
	baseName := filepath.Base(inputPath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// 尝试从参数中解析 outputExtension（目前只在 remux 中使用）
	var outputExtFromParams string
	if params != nil {
		if m, ok := params.(map[string]interface{}); ok {
			if v, ok := m["outputExtension"]; ok {
				if s, ok := v.(string); ok && s != "" {
					if !strings.HasPrefix(s, ".") {
						outputExtFromParams = "." + s
					} else {
						outputExtFromParams = s
					}
				}
			}
		}
	}

	var suffix string
	var ext string

	switch taskType {
	case models.TaskTypeTranscode:
		suffix = "_transcoded"
		ext = ".mp4"
	case models.TaskTypeRemux:
		suffix = "_remuxed"
		if outputExtFromParams != "" {
			ext = outputExtFromParams
		} else {
			ext = ".mp4"
		}
	case models.TaskTypeTrim:
		suffix = "_trimmed"
		ext = filepath.Ext(baseName)
	case models.TaskTypeThumbnail:
		suffix = "_thumbs"
		ext = "" // 缩略图目录
	default:
		suffix = "_processed"
		ext = filepath.Ext(baseName)
	}

	if outputDir == "" {
		outputDir = config.GlobalConfig.FFmpeg.DefaultOutputDir
	}

	if taskType == models.TaskTypeThumbnail {
		return filepath.Join(outputDir, nameWithoutExt+suffix)
	}

	return filepath.Join(outputDir, nameWithoutExt+suffix+ext)
}
