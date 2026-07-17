package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"study/internal/config"
)

const (
	// Extended properties key for AI-generated study events
	propStudyAIGenerated = "study_ai_generated"
	propStudySubject     = "study_subject"
	propStudyFocus       = "study_focus"

	// Default timezone
	defaultTimeZone = "Asia/Shanghai"

	// Google Calendar REST API 端点
	calendarAPIBase = "https://www.googleapis.com/calendar/v3"
)

// ---------- 自定义类型 ----------

// CalendarEvent 日历事件。
// 字段名和路径与旧版 Calendar API 客户端兼容，不改变 CLI 层代码。
type CalendarEvent struct {
	Id                 string                       `json:"id"`
	Summary            string                       `json:"summary"`
	Description        string                       `json:"description"`
	Start              *CalendarEventDateTime       `json:"start"`
	End                *CalendarEventDateTime       `json:"end"`
	Status             string                       `json:"status"`
	HtmlLink           string                       `json:"htmlLink"`
	ExtendedProperties *CalendarExtendedProperties  `json:"extendedProperties"`
}

// CalendarEventDateTime 事件时间信息。
type CalendarEventDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

// CalendarExtendedProperties 扩展属性（私有属性映射）。
type CalendarExtendedProperties struct {
	Private map[string]string `json:"private"`
}

// calendarEventsListResponse 事件列表 API 的 JSON 响应。
type calendarEventsListResponse struct {
	Items []*CalendarEvent `json:"items"`
}

// CalendarEventParams 创建学习事件所需的参数。
type CalendarEventParams struct {
	Subject       string    // 科目名称
	Focus         string    // 学习重点
	StartTime     time.Time // 开始时间
	EndTime       time.Time // 结束时间
	CheckConflict bool      // 是否检查冲突
}

// CalendarStats 日历学习事件统计。
type CalendarStats struct {
	TotalEvents int       // 总学习事件数
	NextEvent   string    // 最近一个事件的摘要
	NextTime    time.Time // 最近一个事件的时间
	Subjects    []string  // 涉及的科目列表
}

// ---------- 服务结构 ----------

// GoogleCalendarService 提供 Google Calendar 事件管理功能。
type GoogleCalendarService struct {
	cfg    *config.Config
	httpClient *http.Client
}

// NewGoogleCalendarService 创建 Google Calendar 服务实例。
// httpClient 应为已通过 OAuth2 认证的 HTTP 客户端（自动附带 Bearer token）。
func NewGoogleCalendarService(cfg *config.Config, httpClient *http.Client) (*GoogleCalendarService, error) {
	return &GoogleCalendarService{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

// ---------- HTTP 辅助方法 ----------

// calendarDo 发送请求并解析 JSON 响应到 v。
func (s *GoogleCalendarService) calendarDo(req *http.Request, v interface{}) error {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API 错误 (%d): %s", resp.StatusCode, string(respBytes))
	}
	if v != nil {
		if err := json.Unmarshal(respBytes, v); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}
	return nil
}

// calendarDoNoContent 发送请求，不解析响应体（用于 DELETE 等操作）。
func (s *GoogleCalendarService) calendarDoNoContent(req *http.Request) error {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 错误 (%d): %s", resp.StatusCode, string(respBytes))
	}
	return nil
}

// ---------- 事件创建 ----------

// createEventBody 是创建事件时发送的 JSON 结构。
type createEventBody struct {
	Summary            string                       `json:"summary"`
	Description        string                       `json:"description"`
	Start              *CalendarEventDateTime       `json:"start"`
	End                *CalendarEventDateTime       `json:"end"`
	ExtendedProperties *CalendarExtendedProperties  `json:"extendedProperties,omitempty"`
}

// CreateStudyEvent 创建一个学习日历事件。
// 事件带有 study_ai_generated 扩展属性标记，方便后续批量管理。
func (s *GoogleCalendarService) CreateStudyEvent(ctx context.Context, params CalendarEventParams) (*CalendarEvent, error) {
	// 冲突检查
	if params.CheckConflict {
		conflicts, err := s.CheckConflicts(ctx, params.StartTime, params.EndTime)
		if err != nil {
			return nil, err
		}
		if len(conflicts) > 0 {
			var desc string
			for i, c := range conflicts {
				if i > 0 {
					desc += "、"
				}
				desc += c.Summary
			}
			return nil, fmt.Errorf("与现有日程冲突: %s", desc)
		}
	}

	summary := fmt.Sprintf("📚 %s - %s", params.Subject, params.Focus)

	body := createEventBody{
		Summary:     summary,
		Description: fmt.Sprintf("科目: %s\n重点: %s", params.Subject, params.Focus),
		Start: &CalendarEventDateTime{
			DateTime: params.StartTime.Format(time.RFC3339),
			TimeZone: defaultTimeZone,
		},
		End: &CalendarEventDateTime{
			DateTime: params.EndTime.Format(time.RFC3339),
			TimeZone: defaultTimeZone,
		},
		ExtendedProperties: &CalendarExtendedProperties{
			Private: map[string]string{
				propStudyAIGenerated: "true",
				propStudySubject:     params.Subject,
				propStudyFocus:       params.Focus,
			},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化事件失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		calendarAPIBase+"/calendars/primary/events",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var created CalendarEvent
	if err := s.calendarDo(req, &created); err != nil {
		return nil, fmt.Errorf("创建日历事件失败: %w", err)
	}
	return &created, nil
}

// ---------- 事件查询 ----------

// ListStudyEvents 列出 AI 生成的学习事件。
// subject 为空时列出所有科目的学习事件。
// maxDate 为最大日期范围（默认今天起 30 天）。
func (s *GoogleCalendarService) ListStudyEvents(ctx context.Context, subject string, maxDate time.Time) ([]*CalendarEvent, error) {
	now := time.Now()
	if maxDate.IsZero() {
		maxDate = now.AddDate(0, 0, 30)
	}

	// 构建查询参数
	queryParams := url.Values{}
	queryParams.Set("timeMin", now.Format(time.RFC3339))
	queryParams.Set("timeMax", maxDate.Format(time.RFC3339))
	queryParams.Set("privateExtendedProperty", propStudyAIGenerated+"=true")
	queryParams.Set("singleEvents", "true")
	queryParams.Set("orderBy", "startTime")
	queryParams.Set("maxResults", "250")

	reqURL := calendarAPIBase + "/calendars/primary/events?" + queryParams.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	var listResp calendarEventsListResponse
	if err := s.calendarDo(req, &listResp); err != nil {
		return nil, fmt.Errorf("查询学习事件失败: %w", err)
	}

	// 按科目筛选（客户端过滤）
	if subject != "" {
		var filtered []*CalendarEvent
		for _, evt := range listResp.Items {
			if evt.ExtendedProperties != nil && evt.ExtendedProperties.Private != nil {
				if evt.ExtendedProperties.Private[propStudySubject] == subject {
					filtered = append(filtered, evt)
				}
			}
		}
		return filtered, nil
	}

	return listResp.Items, nil
}

// ---------- 冲突检测 ----------

// CheckConflicts 检查指定时间段内是否有非学习事件冲突。
func (s *GoogleCalendarService) CheckConflicts(ctx context.Context, startTime, endTime time.Time) ([]*CalendarEvent, error) {
	queryParams := url.Values{}
	queryParams.Set("timeMin", startTime.Format(time.RFC3339))
	queryParams.Set("timeMax", endTime.Format(time.RFC3339))
	queryParams.Set("singleEvents", "true")
	queryParams.Set("maxResults", "100")

	reqURL := calendarAPIBase + "/calendars/primary/events?" + queryParams.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	var listResp calendarEventsListResponse
	if err := s.calendarDo(req, &listResp); err != nil {
		return nil, fmt.Errorf("查询日程冲突失败: %w", err)
	}

	// 过滤：非学习事件才算冲突
	var conflicts []*CalendarEvent
	for _, evt := range listResp.Items {
		// 跳过已取消的事件
		if evt.Status == "cancelled" {
			continue
		}
		// 跳过学习事件（允许重叠）
		if evt.ExtendedProperties != nil && evt.ExtendedProperties.Private != nil {
			if evt.ExtendedProperties.Private[propStudyAIGenerated] == "true" {
				continue
			}
		}
		conflicts = append(conflicts, evt)
	}
	return conflicts, nil
}

// ---------- 批量删除 ----------

// DeleteAllAIEvents 删除所有标记为 AI 生成的学习事件。
// 返回删除的事件数量。
func (s *GoogleCalendarService) DeleteAllAIEvents(ctx context.Context) (int, error) {
	// 先列出所有 AI 生成的事件
	queryParams := url.Values{}
	queryParams.Set("privateExtendedProperty", propStudyAIGenerated+"=true")
	queryParams.Set("maxResults", "2500")

	reqURL := calendarAPIBase + "/calendars/primary/events?" + queryParams.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, fmt.Errorf("构建请求失败: %w", err)
	}

	var listResp calendarEventsListResponse
	if err := s.calendarDo(req, &listResp); err != nil {
		return 0, fmt.Errorf("查询待删除事件失败: %w", err)
	}

	count := 0
	for _, evt := range listResp.Items {
		delURL := calendarAPIBase + "/calendars/primary/events/" + evt.Id
		delReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, delURL, nil)
		if err != nil {
			continue
		}
		if err := s.calendarDoNoContent(delReq); err != nil {
			// 继续删除其他事件，不因单个失败而中断
			continue
		}
		count++
	}
	return count, nil
}

// ---------- 状态查询 ----------

// GetStats 获取学习事件统计信息。
func (s *GoogleCalendarService) GetStats(ctx context.Context) (*CalendarStats, error) {
	events, err := s.ListStudyEvents(ctx, "", time.Time{})
	if err != nil {
		return nil, err
	}

	stats := &CalendarStats{
		TotalEvents: len(events),
	}

	subjectSet := make(map[string]bool)
	for i, evt := range events {
		if evt.ExtendedProperties != nil && evt.ExtendedProperties.Private != nil {
			if subj := evt.ExtendedProperties.Private[propStudySubject]; subj != "" {
				if !subjectSet[subj] {
					stats.Subjects = append(stats.Subjects, subj)
					subjectSet[subj] = true
				}
			}
		}
		if i == 0 {
			stats.NextEvent = evt.Summary
			if evt.Start != nil && evt.Start.DateTime != "" {
				stats.NextTime, _ = time.Parse(time.RFC3339, evt.Start.DateTime)
			}
		}
	}

	return stats, nil
}

// Close 关闭服务（当前无资源需释放，预留接口）。
func (s *GoogleCalendarService) Close() error {
	return nil
}
