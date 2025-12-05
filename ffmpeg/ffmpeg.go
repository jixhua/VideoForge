package ffmpeg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"log"
)

type FFmpeg struct {
	BinaryPath string
}

func NewFFmpeg(binaryPath string) *FFmpeg {
	return &FFmpeg{BinaryPath: binaryPath}
}

// TaskParams 任务参数
type TranscodeParams struct {
	VideoCodec string `json:"videoCodec"` // h264, h265, vp9
	AudioCodec string `json:"audioCodec"` // aac, mp3
	Bitrate    string `json:"bitrate"`    // 2M, 5M
	Resolution string `json:"resolution"` // 1920x1080, 1280x720
}

type TrimParams struct {
	StartTime string `json:"startTime"` // 00:00:10
	Duration  string `json:"duration"`  // 00:05:00
}

type ThumbnailParams struct {
	Interval int    `json:"interval"` // seconds
	Scale    string `json:"scale"`    // 320x240
}

type RemuxParams struct {
	OutputExtension string `json:"outputExtension"` // 例如: "mp4", "flv", "ts", "wav"
}

// ProgressCallback 进度回调函数
type ProgressCallback func(progress float64, message string)

// GetVideoDuration 获取视频时长（秒）
func (f *FFmpeg) GetVideoDuration(inputPath string) (float64, error) {
	cmd := exec.Command(f.BinaryPath, "-i", inputPath)
	output, _ := cmd.CombinedOutput()

	// 解析 Duration: 00:05:24.75
	re := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2}\.\d{2})`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 4 {
		return 0, fmt.Errorf("could not parse duration")
	}

	hours, _ := strconv.ParseFloat(matches[1], 64)
	minutes, _ := strconv.ParseFloat(matches[2], 64)
	seconds, _ := strconv.ParseFloat(matches[3], 64)

	return hours*3600 + minutes*60 + seconds, nil
}

// Transcode 转码
func (f *FFmpeg) Transcode(inputPath, outputPath, paramsJSON string, callback ProgressCallback) (*exec.Cmd, error) {
	var params TranscodeParams
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, err
	}

	args := []string{"-i", inputPath, "-y"}

	if params.VideoCodec != "" {
		args = append(args, "-c:v", params.VideoCodec)
	}
	if params.AudioCodec != "" {
		args = append(args, "-c:a", params.AudioCodec)
	}
	if params.Bitrate != "" {
		args = append(args, "-b:v", params.Bitrate)
	}
	if params.Resolution != "" {
		args = append(args, "-s", params.Resolution)
	}

	args = append(args, outputPath)

	return f.runWithProgress(inputPath, args, callback)
}

// Remux 转封装
func (f *FFmpeg) Remux(inputPath, outputPath, paramsJSON string, callback ProgressCallback) (*exec.Cmd, error) {
	var params RemuxParams
	if paramsJSON != "" {
		_ = json.Unmarshal([]byte(paramsJSON), &params)
	}

	ext := strings.ToLower(filepath.Ext(outputPath))

	var args []string

	switch ext {
	case ".mp4", ".m4v":
		args = []string{"-i", inputPath, "-c:v", "libx264", "-c:a", "aac", "-y", outputPath}
	case ".flv":
		args = []string{"-i", inputPath, "-c:v", "libx264", "-c:a", "aac", "-y", outputPath}
	case ".m3u8":
		args = []string{"-i", inputPath, "-c:v", "libx264", "-c:a", "aac", "-f", "hls", "-y", outputPath}
	default:
		args = []string{"-i", inputPath, "-c", "copy", "-y", outputPath}
	}

	return f.runWithProgress(inputPath, args, callback)
}

// Trim 裁剪
func (f *FFmpeg) Trim(inputPath, outputPath, paramsJSON string, callback ProgressCallback) (*exec.Cmd, error) {
	var params TrimParams
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, err
	}

	args := []string{"-i", inputPath}

	if params.StartTime != "" {
		args = append(args, "-ss", params.StartTime)
	}
	if params.Duration != "" {
		args = append(args, "-t", params.Duration)
	}

	args = append(args, "-c", "copy", "-y", outputPath)

	return f.runWithProgress(inputPath, args, callback)
}

// GenerateThumbnails 生成缩略图
func (f *FFmpeg) GenerateThumbnails(inputPath, outputDir, paramsJSON string, callback ProgressCallback) (*exec.Cmd, error) {
	var params ThumbnailParams
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, err
	}

	if params.Interval <= 0 {
		params.Interval = 5
	}
	if params.Scale == "" {
		params.Scale = "320x240"
	}

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	baseName := filepath.Base(inputPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outputPattern := filepath.Join(outputDir, baseName+"_thumb_%04d.jpg")

	args := []string{
		"-i", inputPath,
		"-vf", fmt.Sprintf("fps=1/%d,scale=%s", params.Interval, params.Scale),
		"-y",
		outputPattern,
	}

	return f.runWithProgress(inputPath, args, callback)
}

// runWithProgress 启动 FFmpeg 进程并开始异步解析进度，立即返回 cmd 供调用方 Wait
func (f *FFmpeg) runWithProgress(inputPath string, args []string, callback ProgressCallback) (*exec.Cmd, error) {
	// 获取视频总时长
	duration, err := f.GetVideoDuration(inputPath)
	if err != nil {
		duration = 0
	}

	args = append([]string{
		"-progress", "pipe:2",
		"-nostats",
		"-loglevel", "error",
	}, args...)

	cmd := exec.Command(f.BinaryPath, args...)

	// 同时捕获 stdout 和 stderr，便于调试
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// 独立 goroutine 实时读取 stderr，解析进度
	go func(totalDuration float64) {
		scanner := bufio.NewScanner(stderr)
		// 匹配行中任意位置的 time=HH:MM:SS.xx
		progressRe := regexp.MustCompile(`time=([0-9]{2}):([0-9]{2}):([0-9]{2}\.[0-9]{2})`)

		for scanner.Scan() {
			line := scanner.Text()

			// 调试输出 FFmpeg 日志，便于确认 stderr 被正确捕获
			log.Printf("ffmpeg stderr: %s", line)

			if matches := progressRe.FindStringSubmatch(line); len(matches) >= 4 {
				hours, _ := strconv.ParseFloat(matches[1], 64)
				minutes, _ := strconv.ParseFloat(matches[2], 64)
				seconds, _ := strconv.ParseFloat(matches[3], 64)
				currentTime := hours*3600 + minutes*60 + seconds

				progress := 0.0
				if totalDuration > 0 {
					progress = (currentTime / totalDuration) * 100
					if progress > 100 {
						progress = 100
					}
				}

				if callback != nil {
					callback(progress, line)
				}
			}
		}
		// stderr 读取结束后，发送 100% 完成信号
		if callback != nil {
			callback(100, "Completed")
		}
	}(duration)

	// 可选：把 stdout 也打到日志中，便于排查
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("ffmpeg stdout: %s", line)
		}
	}()

	// 立即返回已启动的 cmd，由调用方负责 Wait()
	return cmd, nil
}

// IsVideoFile 检查是否为视频文件
func IsVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	videoExts := []string{".mp4", ".mkv", ".avi", ".mov", ".flv", ".wmv", ".webm", ".m4v", ".mpg", ".mpeg"}
	for _, e := range videoExts {
		if ext == e {
			return true
		}
	}
	return false
}
