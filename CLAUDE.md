# CLAUDE.md

## 项目概述

**study管理中心** — 面向大学生的个人学习管理工具（CLI）。
Go 语言，编译为单文件 exe（11MB），Windows 原生，零运行时依赖。

**一句话定位**：打开 Dashboard 看清全局，敲一条命令完成记录。

## 快速启动

```bash
cd c:/Users/A1881/OneDrive/Desktop/study
go build -o study.exe .    # 编译
./study.exe                # 进入 REPL 交互模式
./study.exe overview       # 查看仪表板
./study.exe --help         # 查看所有命令
```

## 技术栈

| 项 | 选型 | 原因 |
|---|------|------|
| 语言 | Go 1.26.5 | 编译为单文件，无运行时依赖 |
| CLI 框架 | cobra v1.10.2 | Go 社区标准 |
| SQLite | modernc.org/sqlite v1.53.0 | 纯 Go 实现，无需 CGO |
| 终端渲染 | 手写 ANSI 序列 | 零依赖，后续可升级 lipgloss |
| GUI | Wails v3 (WebView2) | Windows 原生桌面应用，复用 service 层 |

## 架构分层

```
main.go (入口, 7行)
  ├─ cli/ (命令定义 + REPL, 10个文件)
  └─ cmd/gui/main.go (GUI入口, 新增)
       ├─ service/ (业务逻辑, 14个服务 + bootstrap.go)
       │    ├─ storage/markdown/ (5个文件读写)
       │    ├─ storage/sqlite/ (日记 + FTS5全文搜索)
       │    └─ model/ (7个纯数据结构)
       ├─ cli/ (cobra 命令 + REPL)
       ├─ gui/ (Wails v3 app + handlers + systray + 锁)
       │    ├─ app.go, handlers.go, systray.go
       │    ├─ lock_windows.go, lock_other.go
       └─ frontend/ (纯 HTML/CSS/JS, 无框架)
            ├─ index.html, src/app.js, src/style.css
            └─ src/pages/ (7个页面模块)
```

**依赖方向**: CLI → Service → Storage → Model（Config 横向注入）

## 目录结构

```
study/
├── main.go              # 入口：cli.Init() → cli.Execute()
├── go.mod
├── study.exe            # 编译产物（不提交 Git）
├── study-gui.exe        # GUI 编译产物（不提交 Git）
├── 启动study.exe         # 启动器 — 双击无闪窗打开 Windows Terminal（不提交 Git）
├── CLAUDE.md            # 本文件
├── PROJECT_FEATURES_AND_PHILOSOPHY.md  # 需求规格文档
└── internal/
    ├── config/config.go       # Config struct, 路径管理, EnsureDirs()
    ├── model/                 # 纯数据: subject, record, exam, weakpoint, diary, memo, dashboard, heatmap
    ├── storage/
    │   ├── markdown/          # subjects/exams/weakpoints/memos/records 的 Markdown 读写
    │   └── sqlite/diary.go    # DiaryStore: SaveDiary, GetDiary, SearchDiaries (FTS5), ListRecentDiaries
    ├── service/               # 业务逻辑层
    │   ├── record.go          # 学习记录 (Log, ListRecent, GetAllRecords)
    │   ├── exam.go            # 考试管理 (List, Add, Delete) + 倒计时计算 + 三色标记
    │   ├── weakpoint.go       # 薄弱点 (List, Add, Delete, Stats)
    │   ├── subject.go         # 科目 (List, Add, ListWithMaterialCount)
    │   ├── diary.go           # 日记 (Open→外部编辑器, Search→FTS5, ListRecent)
    │   ├── memo.go            # 备忘 (List, Add, Search)
    │   ├── dashboard.go       # 聚合所有数据生成 Dashboard
    │   ├── streak.go          # 连续学习天数统计
    │   └── heatmap.go         # 140天热力图数据生成
    ├── cli/                   # cobra 命令定义 + REPL
    │   ├── root.go            # Init() 初始化所有服务, buildRootCmd() 注册命令
    │   ├── log.go, exams.go, weakpoints.go, subjects.go,
    │   ├── diary.go, memo.go, overview.go, heatmap.go,
    │   ├── streak.go, init.go, repl.go
    └── render/                # 终端输出渲染
        ├── style.go           # ANSI 颜色、HeatBlock、Section、Bold 等
        ├── dashboard.go       # Dashboard 完整渲染
        └── heatmap.go         # GitHub 风格热力图（7行×20周）
```

## 数据存储方案

| 数据 | 文件 | 格式 |
|------|------|------|
| 科目 | `~/.study/subjects.md` | `- 科目名` |
| 考试 | `~/.study/exams.md` | Markdown 表格 |
| 薄弱点 | `~/.study/weakpoints.md` | Markdown 表格 |
| 备忘 | `~/.study/memos.md` | `- 日期 内容` |
| 学习记录 | `~/.study/records/YYYY-MM.md` | 按月分文件，Markdown 表格 |
| 日记 | `~/.study/diary.db` | SQLite + FTS5 全文搜索 |
| 资料 | `~/.study/materials/<科目>/` | 用户自行放入文件 |

**环境变量**: `STUDY_DATA_DIR` 可覆盖数据目录（默认 `~/.study`）

## 关键设计决策

1. **DiarySvc 可为 nil** — SQLite 初始化失败时不阻断程序，日记功能降级不可用。所有使用 DiarySvc 的地方必须 nil 检查。
2. **Init() 只调用一次** — `cli.Init()` 在 main.go 中调用，初始化全局服务和 cobra 命令树。REPL 复用同一个 rootCmd。
3. **Markdown 存储无锁** — 单用户工具，不需要并发控制。
4. **中文原生** — 所有 UI 文本、紧急程度标签（紧急/不急/考前看）均为中文。
5. **紧急程度枚举** — `model.UrgencyUrgent = "紧急"`, `UrgencyRelaxed = "不急"`, `UrgencyPreExam = "考前看"`

## 当前状态 (2026-07-17)

### 已完成 (全部核心功能 + GUI)
- [x] 学习进度记录 (`study log`)
- [x] 薄弱知识点管理 (`study wp`)
- [x] 考试倒计时 (`study exam`)
- [x] 科目与资料管理 (`study subj`)
- [x] 学习日记 + 全文搜索 (`study diary`)
- [x] 仪表板 Dashboard (`study overview`)
- [x] 学习热力图 (`study heatmap`)
- [x] 连续学习统计 (`study streak`)
- [x] 行政事务备忘 (`study memo`)
- [x] REPL 交互模式 (`study`)
- [x] 首次引导向导 (`study init`)
- [x] GitHub 云端同步 (`study sync`)
- [x] **GUI 桌面应用 (`study-gui.exe`)** — Wails v3 + WebView2
  - 系统托盘（显示/隐藏 + 开机自启 + 退出）
  - 7 个子页面（仪表板/考试/薄弱点/科目/记录/日记/备忘）
  - 单实例检测 + 亮暗双主题
  - 纯 HTML/CSS/JS 前端，无框架依赖
- [x] 编译验证通过（CLI 15MB + GUI 22MB）

### 编译命令

```bash
go build -ldflags="-s -w" -o study.exe .                          # CLI
go build -ldflags="-H windowsgui -s -w" -o study-gui.exe ./cmd/gui/      # GUI（无终端窗口）
go build -ldflags="-H windowsgui -s -w" -o 启动study.exe ./cmd/launcher/  # 启动器（零闪窗）
wails3 dev                                            # GUI 开发模式（热重载）
```

### 明确不做
- 不开发移动端
- 不做社交/协作功能
- 不托管教学内容
- 不处理成绩/学分计算

## 常用模式

添加新功能的步骤：
1. `internal/model/` 定义数据结构（如需要新模型）
2. `internal/storage/markdown/` 或 `sqlite/` 实现读写
3. `internal/service/` 实现业务逻辑
4. `internal/cli/` 注册 cobra 命令（在 `buildRootCmd()` 中 `AddCommand`）
5. `internal/render/` 添加渲染函数（如需要终端输出）

## Git 工作流

**远程仓库**: `https://github.com/duongmxlpvanh-stack/study.git`（已关联 `origin`）

**规则：每次代码更改后必须提交并推送到远程仓库**，做好版本管理，保持提交历史清晰可追溯。

```bash
# 每次修改完代码后执行：
git add -A
git commit -m "<简短描述本次改动>"
git push
```

- 提交粒度：每完成一个有意义的改动就提交一次（一个功能点、一个修复、一个重构）
- 提交信息用中文，简洁描述改动内容
- 推送前先拉取远程变更：若 push 被拒绝，执行 `git pull --rebase` 后再 push
- 功能分支 → PR → squash merge

## 参考文档

- `PROJECT_FEATURES_AND_PHILOSOPHY.md` — 完整需求规格、设计哲学、用户场景
- 原项目代码 — 基本无复用价值，仅作功能参考


# 全程用中文交流
