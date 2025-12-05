# VideoForge - 视频处理服务

基于 Go + FFmpeg 的跨平台视频处理服务，支持转码、转封装、裁剪和缩略图生成。

## 功能特性

- ✅ HTTP/WebSocket 服务
- ✅ FFmpeg 视频处理（转码、转封装、裁剪、缩略图）
- ✅ SQLite 任务持久化
- ✅ 实时进度推送
- ✅ Web UI 界面
- ✅ 跨平台支持（Windows/Linux）

## 快速开始

### 前置要求

- Go 1.21+
- FFmpeg（需在系统 PATH 中或配置路径）

### 安装

```bash
go mod download
```

### 运行

```bash
go run main.go
```

访问 http://localhost:8080

## 配置

编辑 `config.json` 文件配置服务参数。

## API 文档

### HTTP API

- `GET /api/browse?path=xxx` - 浏览目录
- `POST /api/tasks` - 添加任务
- `GET /api/tasks` - 获取任务列表
- `DELETE /api/tasks/:id` - 删除任务
- `GET /api/files/*filepath` - 访问文件

### WebSocket

- `WS /ws` - 实时进度推送

## 许可证

MIT
