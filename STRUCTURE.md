# QOrder 项目结构

```
QOrder/
├── main.go                 # 主程序入口
├── go.mod                  # Go 模块定义
├── go.sum                  # Go 依赖锁定
├── config.json             # 配置文件
├── README.md               # 项目说明
├── USAGE.md                # 使用指南
├── .gitignore              # Git 忽略文件
├── build.bat               # Windows 编译脚本
├── build.sh                # Linux/Mac 编译脚本
├── start.bat               # Windows 启动脚本
├── start.sh                # Linux/Mac 启动脚本
│
├── config/                 # 配置管理
│   └── config.go           # 配置加载逻辑
│
├── models/                 # 数据模型
│   └── task.go             # 任务模型定义
│
├── database/               # 数据库层
│   └── db.go               # SQLite 数据库操作
│
├── ffmpeg/                 # FFmpeg 调用
│   └── ffmpeg.go           # FFmpeg 封装，进度解析
│
├── worker/                 # 任务队列
│   └── queue.go            # 任务队列管理器
│
├── websocket/              # WebSocket 服务
│   └── hub.go              # WebSocket Hub 和客户端管理
│
├── api/                    # HTTP API
│   └── handlers.go         # API 处理器
│
├── web/                    # 前端资源
│   ├── index.html          # 主页面
│   ├── style.css           # 样式文件
│   └── app.js              # 前端逻辑
│
├── output/                 # 输出目录（自动创建）
│   └── (处理后的视频文件)
│
├── videos/                 # 视频根目录（自动创建）
│   └── (待处理的视频文件)
│
└── qorder.db               # SQLite 数据库（运行时创建）
```

## 模块说明

### 后端模块

#### `main.go`
- 程序入口点
- 初始化数据库、WebSocket Hub、任务队列
- 配置路由和启动 HTTP 服务器

#### `config/`
- 加载和管理 JSON 配置文件
- 全局配置访问

#### `models/`
- 定义数据结构
- 任务模型（Task）
- 进度更新模型（ProgressUpdate）

#### `database/`
- SQLite 数据库操作
- 任务 CRUD（创建、读取、更新、删除）
- 数据库模式初始化

#### `ffmpeg/`
- FFmpeg 命令行调用封装
- 支持转码、转封装、裁剪、缩略图生成
- 实时解析 FFmpeg 输出获取进度
- 视频时长检测

#### `worker/`
- 任务队列管理
- 顺序执行任务
- 断点续跑（服务重启后恢复未完成任务）
- 进度回调和状态更新

#### `websocket/`
- WebSocket Hub 管理所有客户端连接
- 实时广播进度更新到所有连接的客户端
- 自动处理连接和断开

#### `api/`
- HTTP API 处理器
- 目录浏览接口
- 任务管理接口（创建、查询、删除）
- 批量任务创建
- 文件访问接口

### 前端模块

#### `web/index.html`
- 主页面结构
- 目录浏览面板
- 任务列表面板
- 进度显示区域
- 视频预览模态框

#### `web/style.css`
- 响应式布局
- 现代化 UI 样式
- 任务状态颜色标识
- 进度条动画

#### `web/app.js`
- WebSocket 客户端
- API 调用封装
- 动态 UI 更新
- 任务参数表单生成
- 视频预览功能

## 数据流

```
1. 用户操作 (浏览器)
   ↓
2. HTTP POST /api/tasks (添加任务)
   ↓
3. 数据库写入 (SQLite)
   ↓
4. 任务队列 (worker.TaskQueue)
   ↓
5. FFmpeg 执行 (ffmpeg.FFmpeg)
   ↓
6. 进度回调 (progressCallback)
   ↓
7. WebSocket 广播 (websocket.Hub)
   ↓
8. 浏览器实时更新 (app.js)
```

## 关键技术点

### 1. 跨平台支持
- 使用 `filepath` 包处理路径
- 配置文件支持不同平台的 FFmpeg 路径

### 2. 并发安全
- WebSocket Hub 使用 mutex 保护客户端映射
- 任务队列使用 channel 进行通信

### 3. 进度解析
- 正则表达式解析 FFmpeg stderr 输出
- 实时计算百分比进度

### 4. 断点续跑
- 启动时查询数据库中未完成的任务
- 重置 running 状态为 pending
- 重新加入任务队列

### 5. 实时通信
- WebSocket 长连接
- 自动重连机制
- JSON 消息序列化

## 扩展建议

1. **用户认证**: 添加登录功能保护系统
2. **多队列**: 支持并行处理多个任务
3. **优先级**: 任务优先级排序
4. **预设模板**: 保存常用的转码参数
5. **历史记录**: 保留已完成任务的记录
6. **日志系统**: 更详细的操作日志
7. **资源限制**: CPU/内存使用限制
8. **通知系统**: 任务完成邮件/推送通知
