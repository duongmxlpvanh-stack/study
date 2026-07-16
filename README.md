# 📚 study 管理中心

> 面向大学生的个人学习管理工具（CLI）—— 打开 Dashboard 看清全局，敲一条命令完成记录。

[![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Windows-原生-blue?logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

---

## ✨ 功能一览

### 核心功能（零依赖，永远可用）

| 命令 | 别名 | 功能 | 说明 |
|------|------|------|------|
| `study overview` | `ov` | 📊 仪表板 | 全局视图：考试倒计时、科目资料、薄弱点统计、学习统计、近期记录与日记 |
| `study log` | `l` | ✏️ 学习记录 | 一句话记录每日学习，按月归档，支持按科目筛选 |
| `study exam` | `e` | 📅 考试管理 | 倒计时 + 三色标记（🔴紧急 / 🟡考前看 / 🟢不急） |
| `study wp` | | 🎯 薄弱点 | 追踪知识薄弱环节，紧急程度标签管理 |
| `study subj` | | 📖 科目管理 | 科目列表 + 本地资料文件夹，一键打开 |
| `study diary` | `dj` | 📝 学习日记 | 外部编辑器打开 + FTS5 全文搜索 + 字数统计 |
| `study memo` | `mm` | 📋 行政备忘 | 快速记录待办事项，支持检索 |
| `study heatmap` | `hm` | 🔥 学习热力图 | GitHub 风格 7×20 日历热力图，140 天可视化，可按科目筛选 |
| `study streak` | `sk` | 🔥 连续天数 | 连续学习天数、累计天数、总记录数、日均统计 |
| `study init` | | 🧭 首次向导 | 交互式引导，对话式完成初始设置 |
| `study` | | 💬 交互 REPL | 进入后直接输入子命令，可持续操作，支持引号参数 |

### 可选功能（需额外配置）

| 命令 | 别名 | 功能 | 前置条件 |
|------|------|------|----------|
| `study gen` | `g` | 🤖 AI 试题/讲义生成 | Python 3.10+ + LaTeX 环境 |
| `study drive` | `dr` | ☁️ Google Drive 上传 | Google OAuth 授权 |
| `study calendar` | `cal` | 📆 Google Calendar 学习事件 | Google OAuth 授权 |
| `study google-auth` | | 🔑 Google 授权管理 | — |
| `study path` | | 📂 数据路径管理 | — |

---

## 🚀 快速开始

### 环境要求

- Windows 10/11
- Go 1.26+（如需自行编译）

### 下载使用

从 [Releases](https://github.com/duongmxlpvanh-stack/study/releases) 下载 `study.exe`（约 11MB），放入任意目录即可运行。

### 从源码编译

```bash
git clone https://github.com/duongmxlpvanh-stack/study.git
cd study
go build -o study.exe .
./study.exe
```

### 首次使用

```bash
./study.exe init    # 初始化向导，创建数据目录与默认配置
./study.exe --help  # 查看所有命令
```

---

## 📁 数据存储

所有数据存放在 `~/.study/` 下，纯文本 Markdown + SQLite：

```
~/.study/
├── subjects.md          # 科目列表
├── exams.md             # 考试信息
├── weakpoints.md        # 薄弱知识点
├── memos.md             # 备忘事项
├── diary.db             # 日记 (SQLite + FTS5 全文搜索)
├── records/
│   ├── 2026-07.md       # 学习记录，按月分文件
│   └── ...
└── materials/
    ├── 高等数学/         # 按科目存放资料
    └── ...
```

> 设置环境变量 `STUDY_DATA_DIR` 可自定义数据目录。

---

## 🏗️ 技术栈

| 项 | 选型 | 亮点 |
|---|------|------|
| 语言 | Go 1.26.5 | 编译为单文件 exe，零运行时依赖 |
| CLI 框架 | [cobra](https://github.com/spf13/cobra) v1.10.2 | Go CLI 标准 |
| 数据库 | [modernc.org/sqlite](https://modernc.org/sqlite) v1.53.0 | 纯 Go SQLite，无需 CGO |
| 终端渲染 | 手写 ANSI 序列 | 零第三方依赖 |
| 终端检测 | [go-isatty](https://github.com/mattn/go-isatty) | 判断交互/管道模式 |

---

## 📐 架构

```
main.go (入口, 7 行)
  └─ cli/ (命令定义 + REPL, 20 个文件)
       ├─ service/ (业务逻辑, 12 个服务)
       │    ├─ auth/          # Google OAuth 认证
       │    ├─ storage/       # markdown + sqlite 读写
       │    └─ model/         # 纯数据结构
       └─ render/ (终端渲染, 3 个文件)
```

**依赖方向**：CLI → Service → Storage → Model（单向依赖，Config 横向注入）

### 服务清单

| 服务 | 职责 | 必需 |
|------|------|------|
| `RecordService` | 学习记录读写、统计 | ✅ |
| `ExamService` | 考试 CRUD + 倒计时 | ✅ |
| `WeakPointService` | 薄弱点 CRUD + 统计 | ✅ |
| `SubjectService` | 科目 CRUD + 资料计数 | ✅ |
| `MemoService` | 备忘 CRUD + 搜索 | ✅ |
| `DiaryService` | 日记读写 + FTS5 全文搜索 | ✅ |
| `DashboardService` | 聚合所有数据生成仪表板 | ✅ |
| `HeatmapService` | 140 天热力图数据 | ✅ |
| `StreakService` | 连续学习天数统计 | ✅ |
| `CourseworkService` | AI 试题/讲义生成（调用 Python 管线） | 可选 |
| `GoogleDriveService` | Google Drive 上传管理 | 可选 |
| `GoogleCalendarService` | Google Calendar 学习事件 | 可选 |

---

## 🎨 设计哲学

> **把精力留给学习本身，管理交给工具。**

- **极简操作**：学习记录就是一句话，不做表单、不填字段
- **单一入口**：Dashboard 是所有信息的汇聚点，日常只需要看它
- **单文件交付**：一个 exe 复制到任意电脑即可使用，无需安装器
- **本地优先**：所有数据存储在本地，云端功能是可选的增强
- **中文原生**：界面语言、提示信息全面中文，中文路径和科目名正常处理
- **数据可移植**：Markdown 纯文本 + SQLite 单文件，整个目录复制即可迁移
- **出错有解**：错误提示用中文、说人话、给出修复建议

详见 [PROJECT_FEATURES_AND_PHILOSOPHY.md](PROJECT_FEATURES_AND_PHILOSOPHY.md)

---

## 📋 当前状态 (2026-07-16)

### ✅ 已完成
- [x] 全部核心功能（学习记录、考试、薄弱点、科目、日记、备忘、仪表板、热力图、连续统计）
- [x] REPL 交互模式 + 首次引导向导
- [x] AI 试题/讲义生成（`study gen`）
- [x] Google Drive 上传 + Google Calendar 集成
- [x] Google OAuth 授权（PKCE 流程）
- [x] 编译验证通过，端到端测试通过

### 🔧 待修复
- [ ] 连续学习天数（streak）计算逻辑需验证
- [ ] 热力图星期起始对齐
- [ ] `study diary open` 在纯 CMD 环境下的编辑器检测

### 📌 明确不做
- 不开发移动端
- 不做社交/协作功能
- 不托管教学内容
- 不处理成绩/学分计算
- 不替代专业笔记软件

---

## 🤝 贡献

```bash
# 修改代码后提交
git add -A
git commit -m "<改动描述>"
git push
```

- 提交信息用中文，简洁描述改动内容
- 功能分支 → PR → squash merge
- 推送前先拉取：若 push 被拒绝，执行 `git pull --rebase` 后再 push

---

## 📄 许可证

[MIT](LICENSE)

## 👤 作者

Lappland · [GitHub](https://github.com/duongmxlpvanh-stack)
