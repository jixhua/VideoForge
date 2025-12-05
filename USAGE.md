# QOrder 使用指南

## 🚀 快速开始

### 1. 安装依赖

确保已安装：
- **Go 1.21+**
- **FFmpeg** (需在系统 PATH 或配置 config.json 中的路径)

在 Windows 上安装 FFmpeg：
```bash
# 下载 FFmpeg: https://ffmpeg.org/download.html
# 将 ffmpeg.exe 放到系统 PATH 或项目目录
```

在 Linux 上安装 FFmpeg：
```bash
sudo apt update
sudo apt install ffmpeg
```

### 2. 下载依赖

```bash
go mod download
```

### 3. 配置文件

编辑 `config.json`：

```json
{
  "server": {
    "host": "0.0.0.0",     // 服务器监听地址
    "port": 8080           // 服务器端口
  },
  "ffmpeg": {
    "path": "ffmpeg",      // Windows: "C:\\ffmpeg\\bin\\ffmpeg.exe"
    "defaultOutputDir": "./output"  // 默认输出目录
  },
  "database": {
    "path": "./qorder.db"  // SQLite 数据库路径
  },
  "videoRootDir": "./videos"  // 默认视频根目录
}
```

### 4. 运行服务

```bash
# 开发模式
go run main.go

# 编译后运行
go build -o qorder
./qorder        # Linux
qorder.exe      # Windows
```

### 5. 访问 Web 界面

打开浏览器访问：`http://localhost:8080`

---

## 📖 功能说明

### 目录浏览
1. 在左侧面板输入目录路径
2. 点击"浏览"按钮查看文件列表
3. 双击文件夹进入子目录
4. 视频文件显示为绿色，可直接预览

### 添加单个任务
1. 在文件列表中找到视频文件
2. 点击"添加任务"按钮
3. 选择任务类型（转码/转封装/裁剪/缩略图）
4. 配置参数后确认

### 批量添加任务
1. 浏览到包含视频的目录
2. 在左侧"批量操作"区域选择任务类型
3. 配置参数：
   - **转码**: 选择视频/音频编码、比特率、分辨率
   - **转封装**: 无需参数，快速转换容器格式
   - **裁剪**: 设置起始时间和持续时间
   - **缩略图**: 设置截图间隔和尺寸
4. 可选：
   - 勾选"递归扫描子目录"处理所有子文件夹
   - 勾选"处理后删除原文件"自动清理
5. 点击"批量添加任务"

### 监控进度
- **实时进度条**: 底部显示当前任务的处理进度
- **任务列表**: 右侧面板显示所有任务状态
- **WebSocket 推送**: 自动更新，无需刷新页面

### 预览视频
- 点击"预览"按钮在浏览器中播放视频
- 支持预览原始视频和处理后的视频

---

## 🎯 任务类型详解

### 1. 转码 (Transcode)
转换视频编码格式，适用于：
- 压缩视频体积
- 提高兼容性
- 优化播放性能

**参数**：
- `videoCodec`: H.264 / H.265 / VP9
- `audioCodec`: AAC / MP3
- `bitrate`: 2M (推荐), 5M (高质量)
- `resolution`: 1920x1080, 1280x720

### 2. 转封装 (Remux)
只改变容器格式，不重新编码：
- 速度快，无质量损失
- 例如：MKV → MP4

### 3. 裁剪 (Trim)
剪切视频片段：
- `startTime`: 起始时间 (HH:MM:SS)
- `duration`: 持续时间 (HH:MM:SS)

### 4. 生成缩略图 (Thumbnail)
批量截图：
- `interval`: 每 N 秒截取一张
- `scale`: 缩略图尺寸 (320x240)

---

## 🔧 API 文档

### HTTP API

#### 浏览目录
```
GET /api/browse?path=/path/to/directory
```

#### 获取所有任务
```
GET /api/tasks
```

#### 创建单个任务
```
POST /api/tasks
Content-Type: application/json

{
  "inputPath": "/path/to/input.mp4",
  "outputPath": "/path/to/output.mp4",
  "type": "transcode",
  "params": {
    "videoCodec": "libx264",
    "audioCodec": "aac",
    "bitrate": "2M"
  },
  "deleteOriginal": false
}
```

#### 批量创建任务
```
POST /api/tasks/batch
Content-Type: application/json

{
  "directory": "/path/to/videos",
  "recursive": true,
  "type": "transcode",
  "params": {...},
  "deleteOriginal": false,
  "outputDir": "./output"
}
```

#### 删除任务
```
DELETE /api/tasks/{id}
```

#### 访问文件
```
GET /api/files/{filepath}
```

### WebSocket

连接到 `ws://localhost:8080/ws` 接收实时进度推送：

```json
{
  "taskId": 1,
  "progress": 45.5,
  "status": "running",
  "fileName": "video.mp4",
  "message": "Processing..."
}
```

---

## 💡 使用技巧

1. **断点续传**: 服务重启后自动恢复未完成的任务
2. **顺序处理**: 任务按添加顺序依次执行，避免系统过载
3. **批量处理**: 使用"递归扫描"一次性处理整个目录树
4. **输出管理**: 所有输出文件默认保存到 `./output` 目录

---

## 🐛 常见问题

### FFmpeg 未找到
**错误**: `exec: "ffmpeg": executable file not found`

**解决**:
- Windows: 下载 FFmpeg 并配置到 PATH，或在 config.json 中指定完整路径
- Linux: `sudo apt install ffmpeg`

### 端口被占用
**错误**: `bind: address already in use`

**解决**: 修改 config.json 中的 `port` 配置

### 数据库锁定
**错误**: `database is locked`

**解决**: 确保只有一个 QOrder 实例在运行

---

## 📦 编译发布

### Windows
```bash
go build -ldflags="-s -w" -o qorder.exe
```

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o qorder
```

### 跨平台编译
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o qorder-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o qorder-windows.exe

# MacOS
GOOS=darwin GOARCH=amd64 go build -o qorder-macos
```

---

## 📄 许可证

MIT License

---

## 🙏 致谢

- [FFmpeg](https://ffmpeg.org/) - 强大的多媒体处理工具
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket 库
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite 驱动
