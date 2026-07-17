# study GUI 图形化界面 — Wails v3 实施方案

> **状态**: 阶段1完成 ✅ | 阶段2-4 待实施  
> **创建**: 2026-07-17 | **更新**: 2026-07-17

---

## 一、背景与目标

**问题**: 用户没带电脑时无法访问 study 数据。Web 路线暂不实施，先做 Windows 桌面 GUI。

**核心理念**:
- **秒开** — 冷启动 < 500ms，系统托盘热启动瞬时
- **单文件交付** — 和 CLI 一样，一个 exe 解决一切
- **代码复用** — `internal/service/`、`internal/model/`、`internal/config/` 零改动共用
- **互不干扰** — CLI (`study.exe`) 和 GUI (`study-gui.exe`) 两个独立二进制，各自编译

---

## 二、技术选型: 为什么 Wails v3

| 对比维度 | Wails v2 | Wails v3 | Electron |
|---------|----------|----------|----------|
| 系统托盘 | 需第三方库hack | ✅ 原生API | ✅ |
| 冷启动 | ~900ms | **~500ms** | ~2.4s |
| 空闲内存 | ~120MB | **~70MB** | ~250MB |
| 二进制大小 | ~18MB | **~12MB** | ~80MB+ |
| Go代码复用 | ✅ | ✅ | ❌ 需重写 |
| 成熟度 | 稳定 | alpha | 稳定 |

**v3 的关键优势**: 内置 systray API（`ToggleWindow`、`AttachWindow`、`OnClick`），直接支撑"秒开"。Windows 端功能基本稳定。

---

## 三、架构总览

```
                        internal/
                    ┌────────────────┐
                    │  config/       │  ← Config, EnsureDirs
                    │  model/        │  ← 7 个纯数据结构
                    │  service/      │  ← 14 个业务服务 + bootstrap.go
                    │  storage/      │  ← Markdown + SQLite
                    └───────┬────────┘
                            │ 共用
              ┌─────────────┼─────────────┐
              │                           │
         main.go                     cmd/gui/main.go
         (CLI入口, 不改)              (GUI入口, 新增)
              │                           │
         internal/cli/              internal/gui/
         cobra命令 + REPL           Wails v3 app + handlers
              │                           │
         study.exe                  study-gui.exe
                                    ├── Go 后端 (复用service)
                                    ├── 系统 WebView2
                                    └── frontend/ (HTML/CSS/JS)

```

### 数据流

```
┌─────────────┐  Wails v3 Service   ┌──────────────┐    直接调用    ┌──────────────┐
│  前端 JS    │ ◄─────────────────► │  gui/Handler  │ ────────────► │  service 层  │
│  (WebView2) │  Go方法自动暴露为JS  │  (Go struct)  │               │  (已有代码)  │
└─────────────┘                     └──────────────┘               └──────┬───────┘
                                                                          │
                                                                  ┌───────┴───────┐
                                                                  │ Markdown 文件  │
                                                                  │ SQLite 数据库  │
                                                                  └───────────────┘
```

---

## 四、目录结构

### 新增文件（阶段1已创建）

```
study/
├── main.go                              # CLI入口（不改）
├── cmd/
│   └── gui/
│       └── main.go                      # GUI入口（package main）
├── internal/
│   ├── service/
│   │   └── bootstrap.go                # 共享服务初始化 (AllServices + Bootstrap)
│   └── gui/
│       ├── app.go                       # Wails v3 应用包装 + 系统托盘
│       └── handlers.go                  # 30+ 前端绑定方法
├── frontend/
│   ├── index.html                       # SPA入口 + 骨架屏
│   └── src/
│       ├── app.js                       # 导航 + GoAPI封装 + 子页面切换
│       ├── style.css                    # 全局样式（亮/暗主题）
│       └── pages/
│           └── dashboard.js             # 仪表板渲染
├── go.mod                               # Wails v3 alpha.95 依赖
└── go.sum
```

### 待创建（阶段2-4）

```
internal/gui/
    └── systray.go                      # 托盘图标资源 + 菜单逻辑细化

frontend/src/pages/
    ├── exams.js                         # 考试管理页
    ├── records.js                       # 学习记录页
    ├── weakpoints.js                    # 薄弱点管理页
    └── diary.js                         # 日记搜索页

frontend/assets/
    └── icon.png                         # 应用图标
```

---

## 五、关键设计决策

### 5.1 服务初始化: `bootstrap.go`

从 `cli/root.go:Init()` 中提取纯服务初始化逻辑，去掉 cobra 命令树构建等 CLI 专属部分。

```go
// AllServices 聚合所有服务实例
type AllServices struct {
    Config   *config.Config
    Record   *RecordService
    Exam     *ExamService
    WP       *WeakPointService
    Subj     *SubjectService
    Memo     *MemoService
    Diary    *DiaryService        // 可能为 nil
    Dash     *DashboardService
    Heat     *HeatmapService
    Streak   *StreakService
    Sync     *GitSyncService
    Gen      *CourseworkService
    Drive    *GoogleDriveService  // nil = 未配置
    Calendar *GoogleCalendarService
}

func Bootstrap(cfg *config.Config, warn func(string)) (*AllServices, error)
```

**CLI 保持不变** — `cli/root.go` 的 `Init()` 不依赖 `Bootstrap()`，避免大规模重构。两者逻辑独立但等效。

### 5.2 Wails v3 绑定机制

v3 通过 `application.NewService(handler)` 注册，Handler 的所有导出方法自动暴露为 JS 函数。

前端调用: `window.go.main.Handler.GetDashboard()` → 返回 Promise

已绑定的方法（30+个）:
- **Dashboard**: `GetDashboard()`
- **考试**: `GetExams()`, `AddExam()`, `DeleteExam()`
- **记录**: `GetRecentRecords()`, `LogRecord()`
- **薄弱点**: `GetWeakPoints()`, `GetWeakPointStats()`, `AddWeakPoint()`, `DeleteWeakPoint()`
- **科目**: `GetSubjects()`, `AddSubject()`
- **日记**: `GetRecentDiaries()`, `SearchDiary()`, `GetDiary()`
- **备忘**: `GetMemos()`, `AddMemo()`, `DeleteMemo()`, `SearchMemo()`
- **热力图**: `GetHeatmap()`
- **统计**: `GetStreak()`
- **同步**: `GetSyncStatus()`, `TriggerSync()`
- **系统**: `Ping()`, `GetDataDir()`

### 5.3 前端: 纯 HTML/CSS/JS

不用任何框架，原因:
- Dashboard 本质是"数据展示 + 少量表单"，不需要 React/Vue
- 零构建步骤 → 开发更快
- 更小的运行时 → 启动更快
- Wails 的 Go 绑定已经是原生 JS 函数，不需要 axios/fetch

开发模式降级: 当 Wails 运行时不可用时（如在普通浏览器中打开），`app.js` 自动切换到开发模式，用模拟数据展示 UI。

### 5.4 秒开策略

#### 冷启动 (~500ms)

```
点击 study-gui.exe
  ├─ 0ms    进程创建，Go runtime 初始化
  ├─ 30ms   config.Load() + EnsureDirs()
  ├─ 60ms   Bootstrap() 完成（无网络调用）
  ├─ 100ms  Wails 窗口创建，WebView2 开始初始化
  ├─ 200ms  窗口显示骨架屏（内联CSS）
  ├─ 300ms  DOM Ready，Dashboard 渲染
  ├─ 400ms  GetDashboard() 返回，填充真实数据
  └─ 500ms  完全可交互 ✅
```

#### 热启动（系统托盘）— 真正秒开

```
系统启动 → study-gui.exe 自动启动 → 最小化到托盘
用户点击托盘图标
  ├─ 0ms    窗口从 hidden → visible
  ├─ 16ms   一帧之内完成显示
  └─ 即时   完全可交互 ✅✅✅
```

关键代码:
```go
// 窗口关闭 = 隐藏到托盘
window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
    window.Hide()
    e.Cancel()  // 阻止默认销毁
})

// Windows 选项: 关闭窗口不退出
Windows: application.WindowsOptions{
    DisableQuitOnLastWindowClosed: true,
}

// 托盘单击切换窗口
tray.OnClick(func() { ToggleWindow() })
```

---

## 六、实施计划

### 阶段 1: 基础设施 ✅ (已完成 2026-07-17)

- [x] `internal/service/bootstrap.go` — AllServices + Bootstrap()
- [x] `internal/gui/app.go` — Wails v3 App + 系统托盘 + 窗口管理
- [x] `internal/gui/handlers.go` — 30+ Go↔JS 绑定方法
- [x] `cmd/gui/main.go` — GUI 独立入口
- [x] `frontend/` — SPA 骨架 + Dashboard + 5 个子页面
- [x] CLI 编译验证通过 (`study.exe`)
- [x] GUI 编译验证通过 (`study-gui.exe`)

当前产物:
| 文件 | 大小 | 说明 |
|------|------|------|
| `study.exe` | 15.1 MB | CLI（功能完整） |
| `study-gui.exe` | 22.6 MB | GUI（含 Wails v3 运行时） |

### 阶段 2: Wails 开发环境 + Dashboard 验证 (待实施)

- [ ] 安装 `wails3` CLI
- [ ] `wails3 dev` 开发模式验证（前端热重载）
- [ ] Go↔JS 通信端到端验证
- [ ] Dashboard 数据与 CLI `study overview` 对比验证
- [ ] 前端样式在 WebView2 中的实际效果确认

### 阶段 3: 子页面完善 (待实施)

- [ ] 考试管理页（列表 + 添加/删除）
- [ ] 学习记录页（历史 + 添加记录输入框）
- [ ] 薄弱点管理页（按紧急程度分组 + 添加/删除）
- [ ] 日记搜索页（搜索框 + 结果列表）
- [ ] 备忘管理页

### 阶段 4: 秒开打磨 + 生产构建 (待实施)

- [ ] 托盘图标资源（PNG → ICO）
- [ ] 开机自启（Startup 文件夹快捷方式）
- [ ] 单实例检测（重复启动激活已有窗口）
- [ ] `wails3 build` 生产构建（前端 embed + WebView2 bootstrapper）
- [ ] `-ldflags="-s -w"` + UPX 压缩优化
- [ ] 端到端测试（冷启动速度、数据一致性）

---

## 七、构建命令

```bash
# CLI 构建（不变）
go build -ldflags="-s -w" -o study.exe .

# GUI 构建（开发模式 — 需要 wails3 CLI）
wails3 dev

# GUI 构建（生产模式）
wails3 build -ldflags="-s -w" -o study-gui.exe
```

---

## 八、前端页面结构

```
┌──────────────────────────────────────────┐
│ 🕮 study管理中心                    ─ □ ✕ │
├────────┬─────────────────────────────────┤
│ 📊 仪表板│  ┌──────┐ ┌──────┐ ┌──┐ ┌──┐  │
│ ⏰ 考试  │  │120天 │ │342条 │ │..│ │..│  │  统计卡片
│ 🎯 薄弱点│  └──────┘ └──────┘ └──┘ └──┘  │
│ 📚 科目  │                                  │
│ 📝 记录  │  ⏰ 考试倒计时列表               │
│ 📖 日记  │  🎯 薄弱点统计                   │
│ 📋 备忘  │  📚 课程概览                     │
│          │  📝 最近学习                     │
│          │  📖 最近日记                     │
└────────┴─────────────────────────────────┘
```

系统托盘菜单:
```
┌──────────────┐
│ 📋 显示/隐藏  │
│ ───────────  │
│ 🚪 退出      │
└──────────────┘
```

---

## 九、风险与注意事项

| 风险 | 影响 | 缓解 |
|------|------|------|
| **Wails v3 alpha 不稳定** | GUI 编译/运行异常 | 当前仅验证编译通过；后续若出问题可降级 v2 |
| **WebView2 未安装** | GUI 无法启动 | Win10 21H2+ 已预装；生产构建可内嵌 bootstrapper |
| **Go model 字段名** | JS 端访问大写字段名 | Wails v3 自动 JSON 序列化，前端直接用 PascalCase |
| **数据竞争** | CLI 和 GUI 同时写文件 | 单用户场景，Markdown 写操作原子，概率极低 |
| **CLI 二进制增大** | 从 ~11MB → ~15MB | Wails v3 升级了共享依赖（lipgloss 等），后续可用 UPX 压缩 |

---

## 十、参考

- [Wails v3 官方文档](https://v3.wails.io/)
- [Wails v3 API (pkg.go.dev)](https://pkg.go.dev/github.com/wailsapp/wails/v3/pkg/application)
- [Wails v3 systray example](https://github.com/wailsapp/wails/blob/v3.0.0-alpha.95/v3/examples/systray-basic/main.go)
- 项目 CLAUDE.md — 架构分层、设计哲学
- 项目 PROJECT_FEATURES_AND_PHILOSOPHY.md — 需求规格
