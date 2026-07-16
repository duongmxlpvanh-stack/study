package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"study/internal/config"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	// Extended properties key for AI-generated study events
	propStudyAIGenerated = "study_ai_generated"
	propStudySubject     = "study_subject"
	propStudyFocus       = "study_focus"

	// Default timezone
	defaultTimeZone = "Asia/Shanghai"
)

// CalendarEventParams 创建学习事件所需的参数。
type CalendarEventParams struct {
	Subject       string    // 科目名称
	Focus         string    // 学习重点
	StartTime     time.Time // 开始时间
	EndTime       time.Time // 结束时间
	CheckConflict bool      // 是否检查冲突
}

// GoogleCalendarService 提供 Google Calendar 事件管理功能。
type GoogleCalendarService struct {
	cfg    *config.Config
	calSvc *calendar.Service
}

// NewGoogleCalendarService 创建 Google Calendar 服务实例。
// httpClient 应为已通过 OAuth2 认证的 HTTP 客户端。
func NewGoogleCalendarService(cfg *config.Config, httpClient *http.Client) (*GoogleCalendarService, error) {
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("初始化 Google Calendar 服务失败: %w", err)
	}
	return &GoogleCalendarService{
		cfg:    cfg,
		calSvc: srv,
	}, nil
}

// ---------- 事件创建 ----------

// CreateStudyEvent 创建一个学习日历事件。
// 事件带有 study_ai_generated 扩展属性标记，方便后续批量管理。
func (s *GoogleCalendarService) CreateStudyEvent(ctx context.Context, params CalendarEventParams) (*calendar.Event, error) {
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

	event := &calendar.Event{
		Summary:     summary,
		Description: fmt.Sprintf("科目: %s\n重点: %s", params.Subject, params.Focus),
		Start: &calendar.EventDateTime{
			DateTime: params.StartTime.Format(time.RFC3339),
			TimeZone: defaultTimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: params.EndTime.Format(time.RFC3339),
			TimeZone: defaultTimeZone,
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				propStudyAIGenerated: "true",
				propStudySubject:     params.Subject,
				propStudyFocus:       params.Focus,
			},
		},
	}

	created, err := s.calSvc.Events.Insert("primary", event).Do()
	if err != nil {
		return nil, fmt.Errorf("创建日历事件失败: %w", err)
	}
	return created, nil
}

// ---------- 事件查询 ----------

// ListStudyEvents 列出 AI 生成的学习事件。
// subject 为空时列出所有科目的学习事件。
// maxDate 为最大日期范围（默认今天起 30 天）。
func (s *GoogleCalendarService) ListStudyEvents(ctx context.Context, subject string, maxDate time.Time) ([]*calendar.Event, error) {
	now := time.Now()
	if maxDate.IsZero() {
		maxDate = now.AddDate(0, 0, 30)
	}

	call := s.calSvc.Events.List("primary").
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(maxDate.Format(time.RFC3339)).
		PrivateExtendedProperty(fmt.Sprintf("%s=true", propStudyAIGenerated)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(250)

	events, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("查询学习事件失败: %w", err)
	}

	// 按科目筛选（客户端过滤）
	if subject != "" {
		var filtered []*calendar.Event
		for _, evt := range events.Items {
			if evt.ExtendedProperties != nil && evt.ExtendedProperties.Private != nil {
				if evt.ExtendedProperties.Private[propStudySubject] == subject {
					filtered = append(filtered, evt)
				}
			}
		}
		return filtered, nil
	}

	return events.Items, nil
}

// ---------- 冲突检测 ----------

// CheckConflicts 检查指定时间段内是否有非学习事件冲突。
func (s *GoogleCalendarService) CheckConflicts(ctx context.Context, startTime, endTime time.Time) ([]*calendar.Event, error) {
	events, err := s.calSvc.Events.List("primary").
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		SingleEvents(true).
		MaxResults(100).
		Do()
	if err != nil {
		return nil, fmt.Errorf("查询日程冲突失败: %w", err)
	}

	// 过滤：非学习事件才算冲突
	var conflicts []*calendar.Event
	for _, evt := range events.Items {
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
	events, err := s.calSvc.Events.List("primary").
		PrivateExtendedProperty(fmt.Sprintf("%s=true", propStudyAIGenerated)).
		MaxResults(2500).
		Do()
	if err != nil {
		return 0, fmt.Errorf("查询待删除事件失败: %w", err)
	}

	count := 0
	for _, evt := range events.Items {
		if err := s.calSvc.Events.Delete("primary", evt.Id).Do(); err != nil {
			// 继续删除其他事件，不因单个失败而中断
			continue
		}
		count++
	}
	return count, nil
}

// ---------- 状态查询 ----------

// CalendarStats 日历学习事件统计。
type CalendarStats struct {
	TotalEvents int       // 总学习事件数
	NextEvent   string    // 最近一个事件的摘要
	NextTime    time.Time // 最近一个事件的时间
	Subjects    []string  // 涉及的科目列表
}

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
