package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"study/internal/model"

	_ "modernc.org/sqlite"
)

// DiaryStore 日记的 SQLite 存储
type DiaryStore struct {
	db *sql.DB
}

// New 创建或打开日记数据库
func New(dbPath string) (*DiaryStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 启用 WAL 模式提高并发
	db.Exec("PRAGMA journal_mode=WAL")

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	return &DiaryStore{db: db}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS diary (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL UNIQUE,
			content TEXT NOT NULL DEFAULT '',
			word_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// FTS5 全文搜索表（外部内容表，与 diary 同步）
		`CREATE VIRTUAL TABLE IF NOT EXISTS diary_fts USING fts5(
			date,
			content,
			content='diary',
			content_rowid='id'
		)`,
		// 触发器保持 FTS 同步
		`CREATE TRIGGER IF NOT EXISTS diary_ai AFTER INSERT ON diary BEGIN
			INSERT INTO diary_fts(rowid, date, content) VALUES (new.id, new.date, new.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS diary_ad AFTER DELETE ON diary BEGIN
			INSERT INTO diary_fts(diary_fts, rowid, date, content) VALUES('delete', old.id, old.date, old.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS diary_au AFTER UPDATE ON diary BEGIN
			INSERT INTO diary_fts(diary_fts, rowid, date, content) VALUES('delete', old.id, old.date, old.content);
			INSERT INTO diary_fts(rowid, date, content) VALUES (new.id, new.date, new.content);
		END`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// SaveDiary 保存日记（insert or update）
func (s *DiaryStore) SaveDiary(date string, content string) error {
	wordCount := len([]rune(content)) // 中文字数统计
	_, err := s.db.Exec(
		`INSERT INTO diary (date, content, word_count, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(date) DO UPDATE SET
		 content = excluded.content,
		 word_count = excluded.word_count,
		 updated_at = CURRENT_TIMESTAMP`,
		date, content, wordCount,
	)
	return err
}

// GetDiary 获取指定日期的日记
func (s *DiaryStore) GetDiary(date string) (*model.Diary, error) {
	row := s.db.QueryRow(
		"SELECT id, date, content, word_count, created_at, updated_at FROM diary WHERE date = ?",
		date,
	)
	var d model.Diary
	var createdAt, updatedAt string
	err := row.Scan(&d.ID, &d.Date, &d.Content, &d.WordCount, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	d.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	d.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &d, nil
}

// DeleteDiary 删除指定日期的日记
func (s *DiaryStore) DeleteDiary(date string) error {
	_, err := s.db.Exec("DELETE FROM diary WHERE date = ?", date)
	return err
}

// ListRecentDiaries 列出最近 N 天的日记
func (s *DiaryStore) ListRecentDiaries(limit int) ([]model.Diary, error) {
	rows, err := s.db.Query(
		"SELECT id, date, content, word_count, created_at, updated_at FROM diary ORDER BY date DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDiaries(rows)
}

// SearchDiaries 全文搜索日记
func (s *DiaryStore) SearchDiaries(keyword string) ([]model.Diary, error) {
	// 使用 FTS5 搜索
	rows, err := s.db.Query(
		`SELECT d.id, d.date, snippet(diary_fts, 2, '…', '…', '', 30) as snippet,
		 d.word_count, d.created_at, d.updated_at
		 FROM diary_fts
		 JOIN diary d ON d.id = diary_fts.rowid
		 WHERE diary_fts MATCH ?
		 ORDER BY rank
		 LIMIT 50`,
		keyword,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diaries []model.Diary
	for rows.Next() {
		var d model.Diary
		var createdAt, updatedAt string
		if err := rows.Scan(&d.ID, &d.Date, &d.Content, &d.WordCount, &createdAt, &updatedAt); err != nil {
			continue
		}
		d.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		d.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		diaries = append(diaries, d)
	}
	return diaries, nil
}

// ListDiaryDates 列出所有有日记的日期
func (s *DiaryStore) ListDiaryDates() ([]string, error) {
	rows, err := s.db.Query("SELECT date FROM diary ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			continue
		}
		dates = append(dates, d)
	}
	return dates, nil
}

func scanDiaries(rows *sql.Rows) ([]model.Diary, error) {
	var diaries []model.Diary
	for rows.Next() {
		var d model.Diary
		var createdAt, updatedAt string
		if err := rows.Scan(&d.ID, &d.Date, &d.Content, &d.WordCount, &createdAt, &updatedAt); err != nil {
			continue
		}
		d.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		d.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		diaries = append(diaries, d)
	}
	return diaries, rows.Err()
}

// CountWords 统计中文总字数
func CountWords(text string) int {
	// 统计非空白字符数（适合中文）
	return len([]rune(strings.TrimSpace(text)))
}

// Close 关闭数据库
func (s *DiaryStore) Close() error {
	return s.db.Close()
}
