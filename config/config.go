package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	FFmpeg struct {
		Path             string `json:"path"`
		DefaultOutputDir string `json:"defaultOutputDir"`
		Threads          int    `json:"threads"`
	} `json:"ffmpeg"`
	Database struct {
		Path string `json:"path"`
	} `json:"database"`
	VideoRootDir string `json:"videoRootDir"`
}

var GlobalConfig Config

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &GlobalConfig)
}
