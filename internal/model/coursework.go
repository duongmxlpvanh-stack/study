package model

// GenIntent 是 LLM 从自然语言中解析出的生成意图
type GenIntent struct {
	Subject    string   `json:"subject"`    // 科目标识（约定文件名，如 "calculus"）
	Mode       string   `json:"mode"`       // "exam" 或 "study"
	Sections   []string `json:"sections"`   // 章节名列表
	Count      int      `json:"count"`      // 题目数量（exam 模式）
	Types      []string `json:"types"`      // 题型：["选择题","填空题","计算题"]
	Difficulty string   `json:"difficulty"` // "基础"/"期末"/"竞赛"
	Build      bool     `json:"build"`      // 是否编译 PDF
	Combined   bool     `json:"combined"`   // 试卷答案是否合订
	Figures    bool     `json:"figures"`    // 是否生成配图
	Title      string   `json:"title"`      // 自定义标题（可选）
	RawInput   string   `json:"-"`          // 用户原始自然语言输入（不序列化）
}

// GenResult 生成结果
type GenResult struct {
	Success  bool     // 是否成功
	OutPaths []string // 产物文件路径（.tex 或 .pdf）
	Summary  string   // 人类可读的摘要
	Errors   []string // 非致命警告/错误
}

// SubjectMeta 科目标识元信息
type SubjectMeta struct {
	Key  string // 约定文件名（无扩展名），如 "calculus"
	Name string // 中文课程名，如 "高等数学（多元微积分）"
}
