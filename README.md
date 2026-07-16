# 📚 study 管理中心

> 面向大学生的个人学习管理工具（CLI）—— 打开 Dashboard 看清全局，敲一条命令完成记录。

[![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Windows-原生-blue?logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

---

## ✨ 功能一览

| 命令 | 功能 | 说明 |
|------|------|------|
| `study overview` | 📊 仪表板 | 全局视图：今日记录、考试倒计时、薄弱点、备忘 |
| `study log` | ✏️ 学习记录 | 记录每日学习时长与科目，按月归档 |
| `study exam` | 📅 考试管理 | 倒计时 + 三色标记（紧急/考前看/不急） |
| `study wp` | 🎯 薄弱点 | 追踪知识薄弱环节，打标签管理 |
| `study subj` | 📖 科目管理 | 科目列表 + 本地资料文件夹 |
| `study diary` | 📝 学习日记 | 外部编辑器打开 + FTS5 全文搜索 |
| `study memo` | 📋 行政备忘 | 快速记录待办事项 |
| `study heatmap` | 🔥 学习热力图 | GitHub 风格，140 天可视化 |
| `study streak` | 🔥 连续天数 | 连续学习天数统计 |
| `study` | 💬 交互 REPL | 进入后直接输入子命令，可持续操作 |

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

## 📁 数据存储

所有数据存放在 `~/.study/` 下，纯文本 Markdown + SQLite：

```
~/.study/
├── subjects.md          # 科目列表
├── exams.md             # 考试信息
├── weakpoints.md        # 薄弱知识点
├── memos.md             # 备忘事项
├── diary.db             # 日记 (SQLite + FTS5)
├── records/
│   ├── 2026-07.md       # 按月分文件
│   └── ...
└── materials/
    ├── 高等数学/         # 按科目存放资料
    └── ...
```

> 设置环境变量 `STUDY_DATA_DIR` 可自定义数据目录。

## 🏗️ 技术栈

| 项 | 选型 | 亮点 |
|---|------|------|
| 语言 | Go | 编译为单文件 exe，零运行时依赖 |
| CLI 框架 | [cobra](https://github.com/spf13/cobra) | Go CLI 标准 |
| 数据库 | [modernc.org/sqlite](https://modernc.org/sqlite) | 纯 Go SQLite，无需 CGO |
| 终端渲染 | 手写 ANSI 序列 | 零第三方依赖 |

## 📐 架构

```
main.go
  └─ cli/ (命令定义 + REPL)
       └─ service/ (业务逻辑)
            └─ storage/ (markdown + sqlite)
                 └─ model/ (数据结构)
       └─ render/ (终端渲染)
```

依赖方向：CLI → Service → Storage → Model（单向依赖，Config 横向注入）

## 🤝 贡献

```bash
# 修改代码后提交
git add -A
git commit -m "<改动描述>"
git push
```

- 提交信息用中文
- 功能分支 → PR → squash merge

## 📄 许可证

[MIT](LICENSE)

## 👤 作者

Lappland · [GitHub](https://github.com/duongmxlpvanh-stack)
