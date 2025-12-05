package database

import (
	"database/sql"
	"time"
	"videoforge/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		input_path TEXT NOT NULL,
		output_path TEXT NOT NULL,
		type TEXT NOT NULL,
		params TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		progress REAL DEFAULT 0,
		error_log TEXT,
		delete_original INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_status ON tasks(status);
	
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) CreateTask(task *models.Task) error {
	result, err := db.conn.Exec(`
		INSERT INTO tasks (input_path, output_path, type, params, status, delete_original, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, task.InputPath, task.OutputPath, task.Type, task.Params, task.Status, task.DeleteOriginal, time.Now(), time.Now())

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.ID = id
	return nil
}

func (db *DB) GetTask(id int64) (*models.Task, error) {
	task := &models.Task{}
	err := db.conn.QueryRow(`
		SELECT id, input_path, output_path, type, params, status, progress, error_log, delete_original, created_at, updated_at
		FROM tasks WHERE id = ?
	`, id).Scan(&task.ID, &task.InputPath, &task.OutputPath, &task.Type, &task.Params,
		&task.Status, &task.Progress, &task.ErrorLog, &task.DeleteOriginal, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return task, nil
}

func (db *DB) GetAllTasks() ([]*models.Task, error) {
	rows, err := db.conn.Query(`
		SELECT id, input_path, output_path, type, params, status, progress, error_log, delete_original, created_at, updated_at
		FROM tasks ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(&task.ID, &task.InputPath, &task.OutputPath, &task.Type, &task.Params,
			&task.Status, &task.Progress, &task.ErrorLog, &task.DeleteOriginal, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (db *DB) GetPendingTasks() ([]*models.Task, error) {
	rows, err := db.conn.Query(`
		SELECT id, input_path, output_path, type, params, status, progress, error_log, delete_original, created_at, updated_at
		FROM tasks WHERE status IN ('pending', 'running') ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(&task.ID, &task.InputPath, &task.OutputPath, &task.Type, &task.Params,
			&task.Status, &task.Progress, &task.ErrorLog, &task.DeleteOriginal, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (db *DB) UpdateTaskStatus(id int64, status models.TaskStatus, progress float64, errorLog string) error {
	_, err := db.conn.Exec(`
		UPDATE tasks SET status = ?, progress = ?, error_log = ?, updated_at = ? WHERE id = ?
	`, status, progress, errorLog, time.Now(), id)
	return err
}

func (db *DB) UpdateTaskProgress(id int64, progress float64) error {
	_, err := db.conn.Exec(`
		UPDATE tasks SET progress = ?, updated_at = ? WHERE id = ?
	`, progress, time.Now(), id)
	return err
}

func (db *DB) DeleteTask(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}

// GetSetting 获取设置值
func (db *DB) GetSetting(key string) (string, error) {
	var value string
	err := db.conn.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetSetting 设置配置值
func (db *DB) SetSetting(key, value string) error {
	_, err := db.conn.Exec(`
		INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`, key, value, time.Now(), value, time.Now())
	return err
}
