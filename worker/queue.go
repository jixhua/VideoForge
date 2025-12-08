package worker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"videoforge/database"
	"videoforge/ffmpeg"
	"videoforge/models"

	"os/exec"
)

type TaskQueue struct {
	db         *database.DB
	ffmpeg     *ffmpeg.FFmpeg
	taskChan   chan *models.Task
	isRunning  bool
	mu         sync.Mutex
	progressCb func(update models.ProgressUpdate)

	currentTask *models.Task
	currentCmd  *exec.Cmd
	cancelMu    sync.Mutex
	canceledTasks map[int64]bool
}

func NewTaskQueue(db *database.DB, ffmpegPath string, threads int, progressCallback func(models.ProgressUpdate)) *TaskQueue {
	return &TaskQueue{
		db:            db,
		ffmpeg:        ffmpeg.NewFFmpeg(ffmpegPath, threads),
		taskChan:      make(chan *models.Task, 100),
		isRunning:     false,
		progressCb:    progressCallback,
		canceledTasks: make(map[int64]bool),
	}
}

// Start 启动任务队列处理器
func (tq *TaskQueue) Start() {
	tq.mu.Lock()
	if tq.isRunning {
		tq.mu.Unlock()
		return
	}
	tq.isRunning = true
	tq.mu.Unlock()

	// 恢复未完成的任务
	go func() {
		if err := tq.recoverPendingTasks(); err != nil {
			log.Printf("Failed to recover pending tasks: %v", err)
		}
	}()

	// 启动任务处理循环
	go tq.processLoop()
}

// recoverPendingTasks 恢复未完成的任务
func (tq *TaskQueue) recoverPendingTasks() error {
	tasks, err := tq.db.GetPendingTasks()
	if err != nil {
		return err
	}

	log.Printf("Recovering %d pending tasks", len(tasks))

	for _, task := range tasks {
		// 重置 running 状态为 pending
		if task.Status == models.TaskStatusRunning {
			task.Status = models.TaskStatusPending
			task.Progress = 0
			tq.db.UpdateTaskStatus(task.ID, models.TaskStatusPending, 0, "")
		}
		tq.taskChan <- task
	}

	return nil
}

// AddTask 添加任务到队列
func (tq *TaskQueue) AddTask(task *models.Task) error {
	task.Status = models.TaskStatusPending
	if err := tq.db.CreateTask(task); err != nil {
		return err
	}

	tq.taskChan <- task

	tq.notifyProgress(models.ProgressUpdate{
		TaskID:  task.ID,
		Status:  string(models.TaskStatusPending),
		Message: "Task added to queue",
	})

	return nil
}

// processLoop 任务处理循环
func (tq *TaskQueue) processLoop() {
	for task := range tq.taskChan {
		tq.processTask(task)
	}
}

// processTask 处理单个任务
func (tq *TaskQueue) processTask(task *models.Task) {
	log.Printf("Processing task %d: %s (%s)", task.ID, task.InputPath, task.Type)

	// 如果任务已被取消（在队列中但尚未开始），直接跳过
	tq.cancelMu.Lock()
	if tq.canceledTasks[task.ID] {
		delete(tq.canceledTasks, task.ID)
		tq.cancelMu.Unlock()
		log.Printf("Task %d canceled before start, skipping", task.ID)
		return
	}
	tq.cancelMu.Unlock()

	// 更新状态为运行中
	tq.db.UpdateTaskStatus(task.ID, models.TaskStatusRunning, 0, "")
	tq.notifyProgress(models.ProgressUpdate{
		TaskID:   task.ID,
		Status:   string(models.TaskStatusRunning),
		FileName: filepath.Base(task.InputPath),
		Message:  "Processing started",
	})

	// 确保输出目录存在
	outputDir := filepath.Dir(task.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		tq.handleTaskError(task, err)
		return
	}

	// 进度回调
	progressCallback := func(progress float64, message string) {
		tq.db.UpdateTaskProgress(task.ID, progress)
		tq.notifyProgress(models.ProgressUpdate{
			TaskID:   task.ID,
			Progress: progress,
			Status:   string(models.TaskStatusRunning),
			FileName: filepath.Base(task.InputPath),
			Message:  message,
		})
	}

	// 根据任务类型执行，获取已启动的 FFmpeg 进程
	var err error
	var cmd *exec.Cmd
	switch task.Type {
	case models.TaskTypeTranscode:
		cmd, err = tq.ffmpeg.Transcode(task.InputPath, task.OutputPath, task.Params, progressCallback)
	case models.TaskTypeRemux:
		cmd, err = tq.ffmpeg.Remux(task.InputPath, task.OutputPath, task.Params, progressCallback)
	case models.TaskTypeTrim:
		cmd, err = tq.ffmpeg.Trim(task.InputPath, task.OutputPath, task.Params, progressCallback)
	case models.TaskTypeThumbnail:
		cmd, err = tq.ffmpeg.GenerateThumbnails(task.InputPath, task.OutputPath, task.Params, progressCallback)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	if err != nil {
		log.Printf("Task %d failed to start ffmpeg: %v", task.ID, err)
		tq.handleTaskError(task, err)
		return
	}

	// 立即设置当前任务和进程，便于删除时 Kill
	tq.cancelMu.Lock()
	tq.currentTask = task
	tq.currentCmd = cmd
	tq.cancelMu.Unlock()

	log.Printf("Task %d ffmpeg started, PID: %d", task.ID, cmd.Process.Pid)

	// 等待 FFmpeg 进程结束
	waitErr := cmd.Wait()

	// 清理当前任务/进程引用
	tq.cancelMu.Lock()
	tq.currentTask = nil
	tq.currentCmd = nil
	tq.cancelMu.Unlock()

	if waitErr != nil {
		log.Printf("Task %d ffmpeg process error: %v", task.ID, waitErr)
		tq.handleTaskError(task, waitErr)
		return
	}

	// 任务成功完成
	tq.db.UpdateTaskStatus(task.ID, models.TaskStatusFinished, 100, "")
	tq.notifyProgress(models.ProgressUpdate{
		TaskID:   task.ID,
		Progress: 100,
		Status:   string(models.TaskStatusFinished),
		FileName: filepath.Base(task.InputPath),
		Message:  "Task completed successfully",
	})

	// 如果设置了删除原文件
	if task.DeleteOriginal {
		if err := os.Remove(task.InputPath); err != nil {
			log.Printf("Failed to delete original file %s: %v", task.InputPath, err)
		} else {
			log.Printf("Deleted original file: %s", task.InputPath)
		}
	}

	log.Printf("Task %d completed successfully", task.ID)
}

// handleTaskError 处理任务错误
func (tq *TaskQueue) handleTaskError(task *models.Task, err error) {
	log.Printf("Task %d failed: %v", task.ID, err)

	tq.db.UpdateTaskStatus(task.ID, models.TaskStatusError, task.Progress, err.Error())
	tq.cancelMu.Lock()
	tq.currentTask = nil
	tq.currentCmd = nil
	tq.cancelMu.Unlock()
	tq.notifyProgress(models.ProgressUpdate{
		TaskID:   task.ID,
		Status:   string(models.TaskStatusError),
		FileName: filepath.Base(task.InputPath),
		Message:  fmt.Sprintf("Error: %v", err),
	})
}

// notifyProgress 通知进度更新
func (tq *TaskQueue) notifyProgress(update models.ProgressUpdate) {
	if tq.progressCb != nil {
		tq.progressCb(update)
	}
}

// CancelTask 取消当前正在执行的任务（如果任务 ID 匹配）
func (tq *TaskQueue) CancelTask(id int64) error {
	tq.cancelMu.Lock()
	defer tq.cancelMu.Unlock()

	// 标记为取消，以便尚未开始的任务被跳过
	tq.canceledTasks[id] = true

	// 如果当前正在运行的是该任务，尝试 Kill 进程
	if tq.currentTask != nil && tq.currentTask.ID == id {
		if tq.currentCmd != nil && tq.currentCmd.Process != nil {
			return tq.currentCmd.Process.Kill()
		}
	}
	return nil
}
