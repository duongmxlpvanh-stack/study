package gui

import (
	"context"
	"fmt"

	"study/internal/model"
	"study/internal/service"
)

// Handler 暴露给前端的所有操作方法
// 每个方法通过 Wails v2 绑定自动转为 JS 函数：window.go.main.App.Handler.XXX()
type Handler struct {
	ctx context.Context
	svc *service.AllServices
}

// NewHandler 创建前端 API 处理器
func NewHandler(svc *service.AllServices) *Handler {
	return &Handler{svc: svc}
}

// ==================== 系统 ====================

// Ping 连通性测试
func (h *Handler) Ping() string {
	return "Go 后端已连接 ✓"
}

// GetDataDir 返回数据目录路径
func (h *Handler) GetDataDir() string {
	return h.svc.Config.DataDir
}

// ==================== Dashboard ====================

// GetDashboard 获取仪表板聚合数据（一次调用，后端 5 个 goroutine 并发加载）
func (h *Handler) GetDashboard() (*model.Dashboard, error) {
	return h.svc.Dash.Overview()
}

// ==================== 学习记录 ====================

// GetRecentRecords 获取最近 N 条学习记录
func (h *Handler) GetRecentRecords(limit int) ([]model.Record, error) {
	return h.svc.Record.ListRecent(limit)
}

// LogRecord 添加一条学习记录。格式："科目: 内容" 或 "科目 内容"
func (h *Handler) LogRecord(input string) error {
	return h.svc.Record.Log(input)
}

// ==================== 考试 ====================

// GetExams 获取所有考试（含倒计时和紧急程度）
func (h *Handler) GetExams() ([]model.ExamWithCountdown, error) {
	return h.svc.Exam.List()
}

// AddExam 添加考试，date 格式 YYYY-MM-DD
func (h *Handler) AddExam(name, date string) error {
	return h.svc.Exam.Add(name, date)
}

// DeleteExam 删除考试（序号从 1 开始）
func (h *Handler) DeleteExam(index int) error {
	return h.svc.Exam.Delete(index)
}

// ==================== 薄弱知识点 ====================

// GetWeakPoints 获取所有薄弱知识点
func (h *Handler) GetWeakPoints() ([]model.WeakPoint, error) {
	return h.svc.WP.List()
}

// GetWeakPointStats 获取薄弱点统计（紧急/不急/考前看 各多少）
func (h *Handler) GetWeakPointStats() (model.WeakPointStats, error) {
	return h.svc.WP.Stats()
}

// AddWeakPoint 添加薄弱知识点
// urgency: "紧急" / "不急" / "考前看"
func (h *Handler) AddWeakPoint(content, urgency, subject string) error {
	return h.svc.WP.Add(content, model.Urgency(urgency), subject)
}

// DeleteWeakPoint 删除薄弱点（序号从 1 开始）
func (h *Handler) DeleteWeakPoint(index int) error {
	return h.svc.WP.Delete(index)
}

// ==================== 科目 ====================

// GetSubjects 获取所有科目（含资料数量）
func (h *Handler) GetSubjects() ([]model.SubjectWithCount, error) {
	return h.svc.Subj.ListWithMaterialCount()
}

// AddSubject 添加科目（同时创建资料文件夹）
func (h *Handler) AddSubject(name string) error {
	return h.svc.Subj.Add(name)
}

// ==================== 日记 ====================

// GetRecentDiaries 获取最近 N 天日记摘要
func (h *Handler) GetRecentDiaries(limit int) ([]model.Diary, error) {
	if h.svc.Diary == nil {
		return nil, fmt.Errorf("日记功能不可用（数据库初始化失败）")
	}
	return h.svc.Diary.ListRecent(limit)
}

// SearchDiary 全文搜索日记
func (h *Handler) SearchDiary(keyword string) ([]model.Diary, error) {
	if h.svc.Diary == nil {
		return nil, fmt.Errorf("日记功能不可用（数据库初始化失败）")
	}
	return h.svc.Diary.Search(keyword)
}

// GetDiary 获取指定日期日记
func (h *Handler) GetDiary(date string) (*model.Diary, error) {
	if h.svc.Diary == nil {
		return nil, fmt.Errorf("日记功能不可用（数据库初始化失败）")
	}
	return h.svc.Diary.Get(date)
}

// ==================== 备忘 ====================

// GetMemos 获取所有备忘
func (h *Handler) GetMemos() ([]model.Memo, error) {
	return h.svc.Memo.List()
}

// AddMemo 添加备忘
func (h *Handler) AddMemo(content string) error {
	return h.svc.Memo.Add(content)
}

// DeleteMemo 删除备忘（序号从 1 开始）
func (h *Handler) DeleteMemo(index int) error {
	return h.svc.Memo.Delete(index)
}

// SearchMemo 搜索备忘
func (h *Handler) SearchMemo(keyword string) ([]model.Memo, error) {
	return h.svc.Memo.Search(keyword)
}

// ==================== 热力图 ====================

// GetHeatmap 获取热力图数据（过去 140 天）
// subject 为空时统计全部科目
func (h *Handler) GetHeatmap(subject string) ([]model.HeatMapDay, error) {
	return h.svc.Heat.Generate(subject)
}

// ==================== 连续统计 ====================

// GetStreak 获取连续学习统计
func (h *Handler) GetStreak() (model.StudyStats, error) {
	return h.svc.Streak.Compute()
}

// ==================== 云同步 ====================

// GetSyncStatus 获取云端同步状态
func (h *Handler) GetSyncStatus() *service.SyncStatus {
	return h.svc.Sync.Status()
}

// TriggerSync 手动触发云端同步
func (h *Handler) TriggerSync() (string, error) {
	return h.svc.Sync.ManualSync()
}
