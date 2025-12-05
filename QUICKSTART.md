# 🚀 QOrder 快速开始指南

> 5 分钟内启动你的视频处理服务！

---

## 📋 前置要求

1. **Go 1.21+** - [下载安装](https://go.dev/dl/)
2. **FFmpeg** - [下载安装](https://ffmpeg.org/download.html)

### Windows 上安装 FFmpeg
```bash
# 方法 1: 使用 Chocolatey
choco install ffmpeg

# 方法 2: 手动安装
# 1. 从 https://ffmpeg.org/download.html 下载
# 2. 解压到 C:\ffmpeg
# 3. 添加 C:\ffmpeg\bin 到系统 PATH
```

### Linux 上安装 FFmpeg
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg

# CentOS/RHEL
sudo yum install ffmpeg

# Arch Linux
sudo pacman -S ffmpeg
```

---

## ⚡ 三步启动

### 步骤 1: 下载依赖
```bash
cd QOrder
go mod download
```

### 步骤 2: 编译（可选）
```bash
# Windows
build.bat

# Linux/Mac
chmod +x build.sh
./build.sh
```

### 步骤 3: 启动服务
```bash
# 方式 1: 直接运行（推荐开发时使用）
# Windows
start.bat

# Linux/Mac
chmod +x start.sh
./start.sh

# 方式 2: 编译后运行
# Windows
qorder.exe

# Linux/Mac
./qorder
```

服务启动后，访问: **http://localhost:8888**

---

## 🎬 第一个任务

### 1. 准备测试视频
将一些视频文件放到 `./videos` 目录（自动创建）

### 2. 浏览目录
1. 打开浏览器访问 `http://localhost:8888`
2. 在左侧输入路径：`./videos`（或完整路径）
3. 点击"浏览"按钮

### 3. 添加任务
**方式 A: 单个文件**
- 在文件列表中找到视频
- 点击"添加任务"按钮

**方式 B: 批量处理**
1. 选择任务类型（转码/转封装/裁剪/缩略图）
2. 配置参数（如果需要）
3. 点击"批量添加任务"

### 4. 查看进度
- 底部进度条显示当前任务
- 右侧列表显示所有任务
- 实时更新（WebSocket）

### 5. 预览结果
- 任务完成后点击"预览结果"
- 在浏览器中直接播放

---

## 📝 常用配置

编辑 `config.json`:

```json
{
  "server": {
    "host": "0.0.0.0",    // 监听地址
    "port": 8888          // 端口（如果 8888 被占用，改成其他的）
  },
  "ffmpeg": {
    "path": "ffmpeg",     // FFmpeg 路径
    "defaultOutputDir": "./output"  // 输出目录
  },
  "database": {
    "path": "./qorder.db" // 数据库文件
  },
  "videoRootDir": "./videos"  // 默认视频目录
}
```

### Windows 自定义 FFmpeg 路径
```json
"ffmpeg": {
  "path": "C:\\ffmpeg\\bin\\ffmpeg.exe"
}
```

---

## 🎯 任务类型速查

### 1️⃣ 转码 (Transcode)
**用途**: 改变视频编码格式，压缩体积

**参数**:
- 视频编码: H.264 / H.265 / VP9
- 音频编码: AAC / MP3
- 比特率: 2M (推荐) / 5M (高质量)
- 分辨率: 1920x1080 / 1280x720

**示例**: 将高清视频压缩为适合网络播放的格式

---

### 2️⃣ 转封装 (Remux)
**用途**: 改变容器格式，不重新编码（速度快）

**示例**: MKV → MP4

---

### 3️⃣ 裁剪 (Trim)
**用途**: 提取视频片段

**参数**:
- 起始时间: 00:00:10 (从第 10 秒开始)
- 持续时间: 00:05:00 (截取 5 分钟)

**示例**: 提取视频的精彩片段

---

### 4️⃣ 生成缩略图 (Thumbnail)
**用途**: 批量截图

**参数**:
- 间隔: 5 秒（每 5 秒截取一张）
- 尺寸: 320x240

**示例**: 为视频生成预览图集

---

## 💡 常见问题

### ❓ FFmpeg 未找到
```
错误: exec: "ffmpeg": executable file not found
```
**解决**: 
1. 确认 FFmpeg 已安装: `ffmpeg -version`
2. 如果未安装，参考上方安装说明
3. 或在 config.json 中指定完整路径

---

### ❓ 端口被占用
```
错误: bind: address already in use
```
**解决**: 修改 `config.json` 中的 port 为其他值（如 9999）

---

### ❓ WebSocket 未连接
**解决**: 
1. 检查浏览器控制台错误
2. 确认服务器正常运行
3. 刷新页面重新连接

---

### ❓ 找不到视频文件
**解决**:
1. 确认路径正确（可以使用绝对路径）
2. 检查文件扩展名是否支持（.mp4, .mkv, .avi 等）
3. Windows 路径使用 `\` 或 `/` 都可以

---

## 🔥 高级用法

### 批量转码整个目录
1. 浏览到视频目录
2. 选择"转码"
3. ✅ 勾选"递归扫描子目录"
4. 配置参数（如 H.264, 2M 比特率）
5. 点击"批量添加任务"

### 处理后自动删除原文件
⚠️ **危险操作，谨慎使用！**

1. ✅ 勾选"处理后删除原文件"
2. 确认输出文件正确
3. 添加任务

### 断点续传
服务意外中断？没问题！
1. 重新启动服务
2. 未完成的任务自动恢复
3. 继续处理

---

## 📱 API 使用

### 通过 API 添加任务
```bash
curl -X POST "http://localhost:8888/api/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "inputPath": "/path/to/input.mp4",
    "outputPath": "/path/to/output.mp4",
    "type": "transcode",
    "params": {
      "videoCodec": "libx264",
      "bitrate": "2M"
    }
  }'
```

### 查询任务状态
```bash
curl "http://localhost:8888/api/tasks"
```

---

## 🎓 最佳实践

1. ✅ **先测试小文件**: 用小视频测试参数是否正确
2. ✅ **检查磁盘空间**: 确保输出目录有足够空间
3. ✅ **备份重要文件**: 处理前备份原始视频
4. ✅ **合理设置比特率**: 
   - 2M 适合一般网络播放
   - 5M 适合高质量保存
   - 太低会严重损失画质
5. ✅ **使用转封装**: 如果只需要改格式，用转封装最快

---

## 🆘 获取帮助

1. 查看完整文档: `USAGE.md`
2. 项目结构说明: `STRUCTURE.md`
3. 测试报告: `TEST_REPORT.md`

---

## 🎉 开始使用吧！

现在你已经准备好了！

```bash
# 启动服务
start.bat  # Windows
./start.sh # Linux/Mac

# 访问
http://localhost:8888
```

享受高效的视频处理体验！🚀
