# Go 项目优化方案

> 目标：在现有 Go 框架内压缩二进制体积、提升运行效率。不重写，不改语言。

## 当前状态（2026-07-17）

| 指标 | 当前值 | 目标值 |
|------|--------|--------|
| 二进制体积 | 22MB（`-ldflags="-s -w"` 后） | 3-8MB |
| 代码量 | ~7,000 行 Go | 不变或略减 |
| 启动时间 | 即时 | 保持不变 |
| 命令响应 | < 100ms | < 50ms |

## 体积来源分析

| 依赖 | 用途 | 估计占比 |
|------|------|----------|
| `google.golang.org/api` (Drive v3 + Calendar v3) | Google Drive 上传、Google Calendar 集成 | **10-15MB** |
| `modernc.org/sqlite` | 纯 Go SQLite（日记 FTS5 搜索） | **5-8MB** |
| `golang.org/x/oauth2` + cloud auth 链 | Google OAuth 认证 | 2-3MB |
| cobra + lipgloss + 其他 | CLI 框架 + 终端渲染 | 1-2MB |

---

## 一、体积优化

### 方案 1：砍掉 Google API 重量级依赖（收益最大）

**问题**：[`internal/service/drive.go`](internal/service/drive.go) 和 [`internal/service/calendar.go`](internal/service/calendar.go) 导入了 `google.golang.org/api/drive/v3` 和 `google.golang.org/api/calendar/v3`。这些是 Google 自动生成的完整 protobuf 客户端，包含整个 Drive/Calendar API 的所有方法和消息类型，体积极大。

**方案 1A：手写 HTTP 替代 auto-generated 客户端（推荐）**

Google Drive 上传文件和 Calendar 创建事件实际只用到 2-3 个 REST API 端点。用标准库 `net/http` 手写几十行代码即可替代：

```
// 不用（拉 10MB+ 依赖）
import "google.golang.org/api/drive/v3"

// 用标准库就够了
func uploadToDrive(token *oauth2.Token, filePath string) error {
    // POST https://www.googleapis.com/upload/drive/v3/files
    // multipart upload: metadata + file content
    // ~50 行代码
}
```

- **预计收益**：22MB → 8-12MB
- **改动范围**：`internal/service/drive.go`、`internal/service/calendar.go`、`internal/auth/auth.go`
- **风险**：低。Google REST API 稳定，手写几个端点不会出问题
- **工作量**：1-2 小时

**方案 1B：build tags 条件编译（临时方案）**

如果暂时不想改代码结构，加 build tags 出两个版本：

```go
//go:build !minimal
// internal/service/drive.go  （完整实现）

//go:build minimal
// internal/service/drive_stub.go  （空实现，返回"功能未编译"）
```

```bash
# 日常用的精简版
go build -tags minimal -ldflags="-s -w" -o study_min.exe .

# 需要 Google 功能时
go build -ldflags="-s -w" -o study.exe .
```

- **预计收益**：日常版本 22MB → 8-12MB
- **改动范围**：加 `_stub.go` 文件和 build tags
- **风险**：低
- **工作量**：30 分钟

---

### 方案 2：替换 SQLite 驱动

**问题**：`modernc.org/sqlite` 是纯 Go 翻译的 SQLite（把 C 源码翻译成 Go），虽然零 CGO 依赖，但体积大。

**替代：`github.com/ncruces/go-sqlite3`**

这是目前 Go 社区评价最好的 SQLite 驱动，用 Wasm 沙箱运行 SQLite，体积比 modernc 小很多，且同样零 CGO：

| 驱动 | 二进制体积 | FTS5 支持 | 零 CGO |
|------|-----------|-----------|--------|
| modernc.org/sqlite | ~5-8MB | ✅ | ✅ |
| ncruces/go-sqlite3 | ~1-2MB | ✅ | ✅ |
| mattn/go-sqlite3 | ~0.5MB | ✅ | ❌ 需要 CGO |

`ncruces/go-sqlite3` 使用标准 `database/sql` 接口，API 层改动很小。

- **预计收益**：再减 3-5MB
- **改动范围**：`internal/storage/sqlite/diary.go`，替换 driver import + 适配差异 API
- **风险**：中。需要验证 FTS5 全文搜索行为一致、中文分词正常
- **工作量**：1-2 小时（含测试）

> ⚠️ 如果 `diary.go` 没有用 `database/sql` 标准接口，而是直接调用 modernc 的专有 API，改动会更大。需要先确认。

---

### 方案 3：UPX 可执行文件压缩（零代码改动，最快见效）

```bash
# 安装
winget install upx

# 压缩（--best 最高压缩率）
upx --best -o study_compressed.exe study.exe
```

Go 二进制通常能压到原来的 **35-50%**：

| 压缩前 | 压缩后 |
|--------|--------|
| 22MB | 8-10MB |

**优点**：5 分钟搞定，不动一行代码。

**缺点**：
- 启动时多花几十毫秒解压（对 CLI 工具基本无感）
- 部分杀毒软件可能误报 UPX 压缩的 exe（小众工具的通病）

---

### 方案 4：Go 编译器层面微调

```bash
go build \
  -ldflags="-s -w" \      # 去掉符号表和调试信息
  -trimpath \              # 去掉文件系统路径
  -gcflags="-l" \          # 禁用内联（牺牲少量性能换体积）
  -o study.exe .
```

`-gcflags="-l"` 禁用内联可能反而让二进制变大（内联消除了函数调用开销），一般不推荐。`-trimpath` 可减几百 KB，聊胜于无。

---

### 体积优化汇总

| 序号 | 方案 | 预计体积 | 改动量 | 风险 |
|------|------|----------|--------|------|
| ① | UPX 压缩 | 22MB → 8-10MB | 零 | 低（可能误报） |
| ② | 砍 Google API 依赖 | 22MB → 8-12MB | 中（写几百行 HTTP 调用） | 低 |
| ③ | 换 SQLite 驱动 | 再 -3~5MB | 小（换 import） | 中（需测试 FTS5） |
| ④ | `-trimpath` | 再 -0.5MB | 零 | 无 |
| **①+②+③ 叠加** | — | **22MB → 3-6MB** | — | — |

---

## 二、运行效率优化

### 1. SQLite 开启 WAL 模式

[`internal/storage/sqlite/diary.go`](internal/storage/sqlite/diary.go) 中加一行：

```go
db.Exec("PRAGMA journal_mode=WAL")
```

**效果**：读写不再互相阻塞。写日记时仍可搜索，搜索时仍可写日记。

- **改动量**：1 行
- **风险**：无。WAL 是 SQLite 默认推荐的模式

---

### 2. Dashboard 数据聚合避免重复 I/O

当前 [`internal/service/dashboard.go`](internal/service/dashboard.go) 中 `overview` 命令每次执行 8 项独立的数据收集（科目、考试、薄弱点、备忘录、近期记录、日记、连续天数、热力图），每一项都独立打开和读取文件。

**优化方向**：将 Dashboard 所需数据在一次文件遍历中收集完毕，或对独立文件做并发读取（goroutine + channel）。

```
当前：串行 8 次 → 优化后：并发读取 → 预计 50ms → 20ms
```

- **改动量**：中（重构 Dashboard 数据收集逻辑）
- **风险**：低

---

### 3. REPL 预加载低频数据

REPL 启动时科目列表、考试列表等低频变更数据可一次性加载到内存。

```go
type REPLState struct {
    subjects    []model.Subject    // 启动时加载，add/delete 时更新
    exams       []model.Exam       // 同上
    // ...
}
```

后续 REPL 命令直接读内存，不再碰磁盘。仅在 `add`/`delete` 操作后刷新对应缓存。

- **改动量**：中
- **风险**：低（需确保写操作后缓存一致性）

---

### 4. 热力图数据缓存

140 天热力图每次都重新遍历所有月份的 records 文件。需求规格决定了只有最近 30 天可能变化（用户不太可能改 3 个月前的记录）。

**策略**：缓存 30 天以前的历史数据（只算一次），每次只重新计算最近 30 天。

- **改动量**：小
- **风险**：低

---

## 三、建议执行顺序

```
第 1 步：UPX 压缩（5 分钟）
  └─ 立刻把 22MB 压到 ~8-10MB，不碰代码

第 2 步：SQLite WAL 模式（1 分钟）
  └─ 解决读写互锁，体验改善立即可感

第 3 步：砍 Google API 依赖（2-4 小时）
  └─ 收益最大，从根上解决体积问题

第 4 步：build tags 条件编译（30 分钟）
  └─ 把 Google 功能做成可选的，日常构建不带

第 5 步：换 SQLite 驱动（1-2 小时）
  └─ 进一步压体积，但需要充分测试

第 6 步：Dashboard 并发读取 + REPL 预加载（2-3 小时）
  └─ 优化运行效率，用户感知明显
```

---

## 四、不做的事情

- ❌ **不重写为 Rust**：投入产出比低，Go 版本在体积优化后可以做到 3-6MB，够用
- ❌ **不换用 CGO SQLite**：会破坏"零运行时依赖"的定位，Windows 上尤其麻烦
- ❌ **不微优化代码热点**：CLI 工具的瓶颈在 I/O，不在 CPU，代码层面的微优化无感
