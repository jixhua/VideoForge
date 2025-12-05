package models

import (
	"time"
)

type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusFinished TaskStatus = "finished"
	TaskStatusError    TaskStatus = "error"
)

type TaskType string

const (
	TaskTypeTranscode TaskType = "transcode"
	TaskTypeRemux     TaskType = "remux"
	TaskTypeTrim      TaskType = "trim"
	TaskTypeThumbnail TaskType = "thumbnail"
)

type Task struct {
	ID             int64      `json:"id"`
	InputPath      string     `json:"inputPath"`
	OutputPath     string     `json:"outputPath"`
	Type           TaskType   `json:"type"`
	Params         string     `json:"params"` // JSON string
	Status         TaskStatus `json:"status"`
	Progress       float64    `json:"progress"`
	ErrorLog       string     `json:"errorLog"`
	DeleteOriginal bool       `json:"deleteOriginal"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type ProgressUpdate struct {
	TaskID   int64   `json:"taskId"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
	FileName string  `json:"fileName"`
	Message  string  `json:"message"`
}
