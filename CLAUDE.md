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
| GUI | 暂无 | 先只做 CLI |

## 架构分层

```
main.go (入口, 7行)
  └─ cli/ (命令定义 + REPL, 10个文件)
       └─ service/ (业务逻辑, 8个服务)
            ├─ storage/markdown/ (5个文件读写)
            ├─ storage/sqlite/ (日记 + FTS5全文搜索)
            └─ model/ (7个纯数据结构)
       └─ render/ (终端渲染, 3个文件)
```

**依赖方向**: CLI → Service → Storage → Model（Config 横向注入）

## 目录结构

```
study/
├── main.go              # 入口：cli.Init() → cli.Execute()
├── go.mod
├── study.exe            # 编译产物（不提交 Git）
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

## 当前状态 (2026-07-16)

### 已完成 (MVP 全部核心功能)
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
- [x] 编译验证通过，端到端测试通过

### 待修复
- [ ] 连续学习天数（streak）计算逻辑需验证：当天没有记录时是否正确归零
- [ ] 热力图星期起始对齐（首周前面的空白填充）
- [ ] `study diary open` 在纯 CMD 环境下的编辑器检测

### 待开发（可选功能，按需启用）
- [ ] PDF 试题与讲义生成（AI + LaTeX）
- [ ] Google Drive 云端上传
- [ ] AI 学习规划 + Google Calendar 集成
- [ ] Windows GUI 桌面客户端

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

```bash
git init && git add -A && git commit -m "初始化 MVP 骨架"
# 后续：功能分支 → PR → squash merge
```

## 参考文档

- `PROJECT_FEATURES_AND_PHILOSOPHY.md` — 完整需求规格、设计哲学、用户场景
- 原项目代码 — 基本无复用价值，仅作功能参考
