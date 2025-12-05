// WebSocket 连接
let ws = null;
let currentPath = '';
let tasks = [];

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    connectWebSocket();
    refreshTasks();
    updateTaskTypeParams();
    browsePath(); // 加载上次浏览的目录
    
    // 任务类型切换时更新参数表单
    document.getElementById('batchTaskType').addEventListener('change', updateTaskTypeParams);
});

// WebSocket 连接
function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        console.log('WebSocket 连接成功');
        updateConnectionStatus(true);
    };
    
    ws.onclose = () => {
        console.log('WebSocket 连接断开');
        updateConnectionStatus(false);
        // 5秒后重连
        setTimeout(connectWebSocket, 5000);
    };
    
    ws.onerror = (error) => {
        console.error('WebSocket 错误:', error);
    };
    
    ws.onmessage = (event) => {
        const update = JSON.parse(event.data);
        handleProgressUpdate(update);
    };
}

function updateConnectionStatus(connected) {
    const statusDot = document.getElementById('wsStatus');
    const statusText = document.getElementById('wsStatusText');
    
    if (connected) {
        statusDot.classList.remove('offline');
        statusDot.classList.add('online');
        statusText.textContent = '已连接';
    } else {
        statusDot.classList.remove('online');
        statusDot.classList.add('offline');
        statusText.textContent = '未连接';
    }
}

// 处理进度更新
function handleProgressUpdate(update) {
    console.log('进度更新:', update);
    
    // 更新当前进度条
    if (update.status === 'running') {
        document.getElementById('currentFile').textContent = update.fileName || '处理中...';
        document.getElementById('currentPercent').textContent = `${update.progress.toFixed(1)}%`;
        document.getElementById('progressBarFill').style.width = `${update.progress}%`;
        document.getElementById('progressMessage').textContent = update.message || '';
    } else if (update.status === 'finished') {
        document.getElementById('progressBarFill').style.width = '100%';
        document.getElementById('currentPercent').textContent = '100%';
        document.getElementById('progressMessage').textContent = '任务完成！';
    } else if (update.status === 'error') {
        document.getElementById('progressMessage').textContent = `错误: ${update.message}`;
    }
    
    // 刷新任务列表
    refreshTasks();
}

// 浏览目录
async function browsePath() {
    const path = document.getElementById('directoryPath').value || '';
    
    try {
        const response = await fetch(`/api/browse?path=${encodeURIComponent(path)}`);
        const data = await response.json();
        
        if (response.ok) {
            currentPath = data.path;
            document.getElementById('directoryPath').value = currentPath;
            displayFileList(data.files);
        } else {
            alert(`错误: ${data.error}`);
        }
    } catch (error) {
        console.error('浏览目录失败:', error);
        alert('浏览目录失败');
    }
}

// 显示文件列表
function displayFileList(files) {
    const fileList = document.getElementById('fileList');
    fileList.innerHTML = '';
    
    if (!files || files.length === 0) {
        fileList.innerHTML = '<p style="text-align: center; color: #999;">目录为空</p>';
        return;
    }
    
    files.forEach(file => {
        const item = document.createElement('div');
        item.className = `file-item ${file.isDir ? 'directory' : ''} ${file.isVideo ? 'video' : ''}`;
        
        const icon = file.isDir ? '📁' : (file.isVideo ? '🎬' : '📄');
        const size = file.isDir ? '' : ` (${formatFileSize(file.size)})`;
        
        item.innerHTML = `
            <span class="file-name">${icon} ${file.name}${size}</span>
            <div class="file-actions">
                ${file.isDir ? `<button class="open-dir" data-path="${escapeHtml(file.path)}">打开</button>` : ''}
                ${file.isVideo ? `<button class="preview-video" data-path="${escapeHtml(file.path)}" data-name="${escapeHtml(file.name)}">预览</button>` : ''}
                ${file.isVideo ? `<button class="add-task" data-path="${escapeHtml(file.path)}">添加任务</button>` : ''}
            </div>
        `;
        
        if (file.isDir) {
            const openBtn = item.querySelector('.open-dir');
            if (openBtn) {
                openBtn.addEventListener('click', () => {
                    const dirPath = openBtn.getAttribute('data-path');
                    navigateToDir(dirPath);
                });
            }

            item.addEventListener('dblclick', () => navigateToDir(file.path));
        }

        if (file.isVideo) {
            const previewBtn = item.querySelector('.preview-video');
            if (previewBtn) {
                previewBtn.addEventListener('click', () => {
                    const videoPath = previewBtn.getAttribute('data-path');
                    const videoName = previewBtn.getAttribute('data-name');
                    previewVideo(videoPath, videoName);
                });
            }

            const addTaskBtn = item.querySelector('.add-task');
            if (addTaskBtn) {
                addTaskBtn.addEventListener('click', () => {
                    const videoPath = addTaskBtn.getAttribute('data-path');
                    addSingleTask(videoPath);
                });
            }
        }
        
        fileList.appendChild(item);
    });
}

function navigateToDir(path) {
    document.getElementById('directoryPath').value = path;
    browsePath();
}

function goParentDirectory() {
    const input = document.getElementById('directoryPath');
    let path = input.value || '';

    // 兼容 Windows 和 Linux/Mac，统一把反斜杠转成正斜杠再处理
    let normalized = path.replace(/\\/g, '/');

    // 去掉末尾的斜杠
    normalized = normalized.replace(/\/+$/, '');

    const parts = normalized.split('/');
    if (parts.length <= 1) {
        return; // 已经是根了
    }

    parts.pop();
    let parent = parts.join('/');

    // 对 Windows 盘符（如 C:）恢复为 C:\ 形式
    if (/^[A-Za-z]:$/.test(parent)) {
        parent += '\\';
    }

    input.value = parent;
    browsePath();
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i];
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML.replace(/'/g, "\\'");
}

// 更新任务参数表单
function updateTaskTypeParams() {
    const taskType = document.getElementById('batchTaskType').value;
    const paramsForm = document.getElementById('taskParamsForm');
    
    let html = '';
    
    if (taskType === 'transcode') {
        html = `
            <div class="param-input">
                <label>视频编码:</label>
                <select id="videoCodec">
                    <option value="libx264">H.264</option>
                    <option value="libx265">H.265</option>
                    <option value="libvpx-vp9">VP9</option>
                </select>
            </div>
            <div class="param-input">
                <label>音频编码:</label>
                <select id="audioCodec">
                    <option value="aac">AAC</option>
                    <option value="libmp3lame">MP3</option>
                </select>
            </div>
            <div class="param-input">
                <label>比特率:</label>
                <input type="text" id="bitrate" placeholder="例如: 2M" value="2M">
            </div>
            <div class="param-input">
                <label>分辨率:</label>
                <input type="text" id="resolution" placeholder="例如: 1920x1080">
            </div>
        `;
    } else if (taskType === 'remux') {
        html = `
            <div class="param-input">
                <label>目标封装格式:</label>
                <select id="outputExtension">
                    <option value="mp4">MP4</option>
                    <option value="flv">FLV</option>
                    <option value="m3u8">M3U8(HLS)</option>
                </select>
            </div>
        `;
    } else if (taskType === 'trim') {
        html = `
            <div class="param-input">
                <label>起始时间 (HH:MM:SS):</label>
                <input type="text" id="startTime" placeholder="00:00:00" value="00:00:00">
            </div>
            <div class="param-input">
                <label>持续时间 (HH:MM:SS):</label>
                <input type="text" id="duration" placeholder="00:05:00" value="00:05:00">
            </div>
        `;
    } else if (taskType === 'thumbnail') {
        html = `
            <div class="param-input">
                <label>截图间隔 (秒):</label>
                <input type="number" id="interval" value="5" min="1">
            </div>
            <div class="param-input">
                <label>缩略图尺寸:</label>
                <input type="text" id="scale" placeholder="320x240" value="320x240">
            </div>
        `;
    }
    
    paramsForm.innerHTML = html;
}

// 获取任务参数
function getTaskParams() {
    const taskType = document.getElementById('batchTaskType').value;
    const params = {};
    
    if (taskType === 'transcode') {
        params.videoCodec = document.getElementById('videoCodec').value;
        params.audioCodec = document.getElementById('audioCodec').value;
        params.bitrate = document.getElementById('bitrate').value;
        params.resolution = document.getElementById('resolution').value;
    } else if (taskType === 'remux') {
        params.outputExtension = document.getElementById('outputExtension').value;
    } else if (taskType === 'trim') {
        params.startTime = document.getElementById('startTime').value;
        params.duration = document.getElementById('duration').value;
    } else if (taskType === 'thumbnail') {
        params.interval = parseInt(document.getElementById('interval').value);
        params.scale = document.getElementById('scale').value;
    }
    
    return params;
}

// 批量添加任务
async function batchAddTasks() {
    const directory = currentPath || document.getElementById('directoryPath').value;
    if (!directory) {
        alert('请先浏览一个目录');
        return;
    }
    
    const taskType = document.getElementById('batchTaskType').value;
    const recursive = document.getElementById('batchRecursive').checked;
    const deleteOriginal = document.getElementById('batchDeleteOriginal').checked;
    const params = getTaskParams();
    
    try {
        const response = await fetch('/api/tasks/batch', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                directory,
                recursive,
                type: taskType,
                params,
                deleteOriginal,
                outputDir: ''
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            alert(`成功添加 ${data.count} 个任务！`);
            refreshTasks();
        } else {
            alert(`错误: ${data.error}`);
        }
    } catch (error) {
        console.error('批量添加任务失败:', error);
        alert('批量添加任务失败');
    }
}

// 添加单个任务
async function addSingleTask(inputPath) {
    const taskType = document.getElementById('batchTaskType').value;
    const deleteOriginal = document.getElementById('batchDeleteOriginal').checked;
    const params = getTaskParams();

    // 生成输出路径
    let outputPath;
    if (taskType === 'remux') {
        const fileName = inputPath.split(/[\\/]/).pop();
        const nameWithoutExt = fileName.replace(/\.[^/.]+$/, '');
        const outputExt = params.outputExtension ? `.${params.outputExtension}` : '.mp4';
        outputPath = `./output/${nameWithoutExt}_remuxed${outputExt}`;
    } else {
        const fileName = inputPath.split(/[\\/]/).pop();
        outputPath = `./output/${fileName}`;
    }

    try {
        const response = await fetch('/api/tasks', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                inputPath,
                outputPath,
                type: taskType,
                params,
                deleteOriginal
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            alert('任务已添加！');
            refreshTasks();
        } else {
            alert(`错误: ${data.error}`);
        }
    } catch (error) {
        console.error('添加任务失败:', error);
        alert('添加任务失败');
    }
}

// 刷新任务列表
async function refreshTasks() {
    try {
        const response = await fetch('/api/tasks');
        const data = await response.json();
        
        if (response.ok) {
            tasks = data || [];
            displayTasks();
        }
    } catch (error) {
        console.error('刷新任务列表失败:', error);
    }
}

// 显示任务列表
function displayTasks() {
    const tasksList = document.getElementById('tasksList');
    
    if (!tasks || tasks.length === 0) {
        tasksList.innerHTML = '<p style="text-align: center; color: #999; padding: 20px;">暂无任务</p>';
        return;
    }
    
    tasksList.innerHTML = tasks.map(task => {
        const statusText = {
            'pending': '等待中',
            'running': '处理中',
            'finished': '已完成',
            'error': '失败'
        }[task.status] || task.status;
        
        const typeText = {
            'transcode': '转码',
            'remux': '转封装',
            'trim': '裁剪',
            'thumbnail': '缩略图'
        }[task.type] || task.type;
        
        return `
            <div class="task-item ${task.status}">
                <div class="task-header">
                    <span class="task-type">${typeText}</span>
                    <span class="task-status">${statusText}</span>
                </div>
                <div class="task-path">
                    <strong>输入:</strong> ${task.inputPath}<br>
                    <strong>输出:</strong> ${task.outputPath}
                </div>
                ${task.status === 'running' || task.status === 'finished' ? `
                    <div class="task-progress">
                        <div class="task-progress-bar">
                            <div class="task-progress-fill" style="width: ${task.progress}%"></div>
                        </div>
                        <small>${task.progress.toFixed(1)}%</small>
                    </div>
                ` : ''}
                ${task.status === 'error' ? `
                    <div style="color: #ef4444; font-size: 12px; margin-top: 5px;">
                        ${task.errorLog}
                    </div>
                ` : ''}
                <div class="task-actions">
                    ${task.status === 'finished' ? `
                        <button onclick="previewVideo('${escapeHtml(task.outputPath)}', '处理后的视频')">预览结果</button>
                    ` : ''}
                    <button onclick="deleteTask(${task.id})">删除</button>
                </div>
            </div>
        `;
    }).join('');
}

// 删除任务
async function deleteTask(taskId) {
    if (!confirm('确定要删除此任务吗？')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/tasks/${taskId}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            refreshTasks();
        } else {
            alert('删除失败');
        }
    } catch (error) {
        console.error('删除任务失败:', error);
        alert('删除任务失败');
    }
}

// 清除已完成的任务
async function clearFinishedTasks() {
    const finishedTasks = tasks.filter(t => t.status === 'finished');
    
    if (finishedTasks.length === 0) {
        alert('没有已完成的任务');
        return;
    }
    
    if (!confirm(`确定要删除 ${finishedTasks.length} 个已完成的任务吗？`)) {
        return;
    }
    
    for (const task of finishedTasks) {
        await deleteTask(task.id);
    }
}

// 预览视频
function previewVideo(filePath, title) {
    const modal = document.getElementById('videoModal');
    const videoPlayer = document.getElementById('videoPlayer');
    const videoTitle = document.getElementById('videoTitle');
    
    videoTitle.textContent = title;
    videoPlayer.src = `/api/files/${encodeURIComponent(filePath)}`;
    modal.style.display = 'block';
}

function closeVideoModal() {
    const modal = document.getElementById('videoModal');
    const videoPlayer = document.getElementById('videoPlayer');
    
    videoPlayer.pause();
    videoPlayer.src = '';
    modal.style.display = 'none';
}

// 点击模态框外部关闭
window.onclick = function(event) {
    const modal = document.getElementById('videoModal');
    if (event.target === modal) {
        closeVideoModal();
    }
}
