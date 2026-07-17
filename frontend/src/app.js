/**
 * study GUI — 主应用逻辑
 *
 * 纯 JS，无框架依赖。通过 Wails v3 自动生成的绑定调用 Go 后端。
 * 开发时：Go 方法通过 window.go.main.Handler.XXX() 暴露
 */

// ==================== 应用状态 ====================

const AppState = {
    currentPage: 'dashboard',
    dashboard: null,      // 缓存的 Dashboard 数据
    lastRefresh: 0,       // 上次刷新时间戳
};

// ==================== Go 后端调用封装 ====================

// GoAPI 封装所有对 Go 后端的调用
// 在 Wails v3 环境中，这些方法通过 window.go 对象自动注入
const GoAPI = {
    // 检测 Wails 运行时是否就绪
    isReady() {
        return typeof window.go !== 'undefined' && window.go.main && window.go.main.Handler;
    },

    // 获取 Handler 引用
    get handler() {
        return window.go.main.Handler;
    },

    // Dashboard
    async getDashboard() {
        return await this.handler.GetDashboard();
    },

    // 学习记录
    async getRecentRecords(limit) {
        return await this.handler.GetRecentRecords(limit);
    },
    async logRecord(input) {
        return await this.handler.LogRecord(input);
    },

    // 考试
    async getExams() {
        return await this.handler.GetExams();
    },
    async addExam(name, date) {
        return await this.handler.AddExam(name, date);
    },
    async deleteExam(index) {
        return await this.handler.DeleteExam(index);
    },

    // 薄弱点
    async getWeakPoints() {
        return await this.handler.GetWeakPoints();
    },
    async getWeakPointStats() {
        return await this.handler.GetWeakPointStats();
    },
    async addWeakPoint(content, urgency, subject) {
        return await this.handler.AddWeakPoint(content, urgency, subject);
    },
    async deleteWeakPoint(index) {
        return await this.handler.DeleteWeakPoint(index);
    },

    // 科目
    async getSubjects() {
        return await this.handler.GetSubjects();
    },
    async addSubject(name) {
        return await this.handler.AddSubject(name);
    },

    // 日记
    async getRecentDiaries(limit) {
        return await this.handler.GetRecentDiaries(limit);
    },
    async searchDiary(keyword) {
        return await this.handler.SearchDiary(keyword);
    },

    // 备忘
    async getMemos() {
        return await this.handler.GetMemos();
    },
    async addMemo(content) {
        return await this.handler.AddMemo(content);
    },
    async deleteMemo(index) {
        return await this.handler.DeleteMemo(index);
    },

    // 热力图
    async getHeatmap(subject) {
        return await this.handler.GetHeatmap(subject || '');
    },

    // 统计
    async getStreak() {
        return await this.handler.GetStreak();
    },

    // 系统
    ping() {
        return this.handler.Ping();
    },
    getDataDir() {
        return this.handler.GetDataDir();
    },
};

// ==================== 导航 ====================

function initNavigation() {
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', () => {
            const page = item.dataset.page;
            navigateTo(page);
        });
    });
}

function navigateTo(page) {
    // 更新导航高亮
    document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
    const target = document.querySelector(`.nav-item[data-page="${page}"]`);
    if (target) target.classList.add('active');

    AppState.currentPage = page;
    renderPage(page);
}

function renderPage(page) {
    const content = document.getElementById('content');
    switch (page) {
        case 'dashboard':
            renderDashboard(content);
            break;
        case 'exams':
            content.innerHTML = '<h2>⏰ 考试管理</h2><p>加载中...</p>';
            loadExamsPage(content);
            break;
        case 'weakpoints':
            content.innerHTML = '<h2>🎯 薄弱知识点</h2><p>加载中...</p>';
            loadWeakPointsPage(content);
            break;
        case 'subjects':
            content.innerHTML = '<h2>📚 科目管理</h2><p>加载中...</p>';
            loadSubjectsPage(content);
            break;
        case 'records':
            content.innerHTML = '<h2>📝 学习记录</h2><p>加载中...</p>';
            loadRecordsPage(content);
            break;
        case 'diary':
            content.innerHTML = '<h2>📖 日记</h2><p>加载中...</p>';
            loadDiaryPage(content);
            break;
        case 'memos':
            content.innerHTML = '<h2>📋 备忘</h2><p>加载中...</p>';
            loadMemosPage(content);
            break;
        default:
            content.innerHTML = '<p>未知页面</p>';
    }
}

// ==================== 子页面加载（占位，阶段3实现） ====================

async function loadExamsPage(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p>后端未连接</p>'; return; }
    try {
        const exams = await GoAPI.getExams();
        let html = '<h2>⏰ 考试管理</h2>';
        if (!exams || exams.length === 0) {
            html += '<p class="empty">暂无考试</p>';
        } else {
            html += '<div class="list">';
            exams.forEach((e, i) => {
                const cls = e.DaysLeft <= 7 ? 'urgent' : e.DaysLeft <= 30 ? 'warn' : 'ok';
                html += `<div class="list-item ${cls}">
                    <span>${e.Name}</span>
                    <span>${e.Date}</span>
                    <span>剩余 ${e.DaysLeft} 天</span>
                    <span>${e.UrgencyStr}</span>
                </div>`;
            });
            html += '</div>';
        }
        container.innerHTML = html;
    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

async function loadWeakPointsPage(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p>后端未连接</p>'; return; }
    try {
        const [wps, stats] = await Promise.all([GoAPI.getWeakPoints(), GoAPI.getWeakPointStats()]);
        let html = '<h2>🎯 薄弱知识点</h2>';
        html += `<div class="stat-summary">
            <span class="urgent">紧急: ${stats.Urgent}</span>
            <span class="warn">考前看: ${stats.PreExam}</span>
            <span class="dim">不急: ${stats.Relaxed}</span>
            <span>总计: ${stats.Urgent + stats.PreExam + stats.Relaxed}</span>
        </div>`;
        if (!wps || wps.length === 0) {
            html += '<p class="empty">暂无薄弱点</p>';
        } else {
            html += '<div class="list">';
            wps.forEach((w, i) => {
                html += `<div class="list-item"><span>${i + 1}.</span><span>[${w.Urgency}]</span><span>${w.Content}</span><span class="dim">${w.Subject || ''}</span></div>`;
            });
            html += '</div>';
        }
        container.innerHTML = html;
    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

async function loadSubjectsPage(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p>后端未连接</p>'; return; }
    try {
        const subjects = await GoAPI.getSubjects();
        let html = '<h2>📚 科目管理</h2>';
        if (!subjects || subjects.length === 0) {
            html += '<p class="empty">暂无科目</p>';
        } else {
            html += '<div class="list">';
            subjects.forEach(s => {
                html += `<div class="list-item"><span>${s.Name}</span><span>${s.MaterialCount} 份资料</span></div>`;
            });
            html += '</div>';
        }
        container.innerHTML = html;
    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

async function loadRecordsPage(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p>后端未连接</p>'; return; }
    try {
        const records = await GoAPI.getRecentRecords(50);
        let html = '<h2>📝 学习记录</h2>';
        if (!records || records.length === 0) {
            html += '<p class="empty">暂无学习记录</p>';
        } else {
            html += '<div class="list">';
            records.forEach(r => {
                html += `<div class="list-item"><span class="dim">${r.Date}</span><span class="subject">${r.Subject}</span><span>${r.Content}</span></div>`;
            });
            html += '</div>';
        }
        container.innerHTML = html;
    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

async function loadDiaryPage(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p>后端未连接</p>'; return; }
    try {
        const diaries = await GoAPI.getRecentDiaries(20);
        let html = '<h2>📖 日记</h2>';
        if (!diaries || diaries.length === 0) {
            html += '<p class="empty">暂无日记</p>';
        } else {
            html += '<div class="list">';
            diaries.forEach(d => {
                const preview = (d.Content || '').substring(0, 60);
                html += `<div class="list-item"><span>${d.Date}</span><span class="dim">${d.WordCount}字</span><span>${preview}...</span></div>`;
            });
            html += '</div>';
        }
        container.innerHTML = html;
    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

async function loadMemosPage(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p>后端未连接</p>'; return; }
    try {
        const memos = await GoAPI.getMemos();
        let html = '<h2>📋 行政备忘</h2>';
        if (!memos || memos.length === 0) {
            html += '<p class="empty">暂无误</p>';
        } else {
            html += '<div class="list">';
            memos.forEach(m => {
                html += `<div class="list-item"><span>${m.Content}</span></div>`;
            });
            html += '</div>';
        }
        container.innerHTML = html;
    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

// ==================== 启动 ====================

function hideSkeleton() {
    const skel = document.getElementById('skeleton');
    if (skel) skel.style.display = 'none';
    const app = document.getElementById('app');
    if (app) app.style.display = 'flex';
}

async function initApp() {
    initNavigation();

    // 检查 Wails 运行时
    if (GoAPI.isReady()) {
        // 真实 Wails 环境
        hideSkeleton();
        await refreshDashboard();
        navigateTo('dashboard');
    } else {
        // 开发/降级模式：用模拟数据显示 UI
        console.log('[dev] Wails 运行时未检测到，使用开发模式');
        hideSkeleton();
        useDevMode();
        navigateTo('dashboard');
    }
}

// 开发模式：用静态数据展示 UI（无需 Go 后端）
function useDevMode() {
    // 覆盖 GoAPI 方法，返回模拟数据
    GoAPI.isReady = () => true; // 假装就绪
    GoAPI.getDashboard = async () => ({
        Exams: [
            { Name: '期末考试', Date: '2026-07-20', DaysLeft: 3, UrgencyStr: '🔴 临近' },
            { Name: '六级', Date: '2026-12-15', DaysLeft: 151, UrgencyStr: '🟢 充裕' },
        ],
        Subjects: [
            { Name: '高等数学', MaterialCount: 5 },
            { Name: '大学物理', MaterialCount: 3 },
        ],
        WeakPointStats: { Urgent: 2, PreExam: 1, Relaxed: 3 },
        StudyStats: { TotalDays: 120, TotalRecords: 342, StreakDays: 15, AvgPerDay: 2.85 },
        RecentRecords: [
            { Date: '2026-07-16', Subject: '高数', Content: '多元函数微分习题' },
            { Date: '2026-07-15', Subject: '物理', Content: '电磁感应复习' },
        ],
        RecentDiaries: [
            { Date: '2026-07-16', WordCount: 120, Content: '今天复习了多元函数微分，对链式法则有了更深的理解...' },
        ],
    });
    GoAPI.getExams = async () => [
        { Name: '期末考试', Date: '2026-07-20', DaysLeft: 3, UrgencyStr: '🔴 临近' },
    ];
    GoAPI.getWeakPoints = async () => [
        { Content: '泰勒展开的误差估计', Urgency: '紧急', Subject: '高数' },
        { Content: '麦克斯韦方程组推导', Urgency: '考前看', Subject: '大物' },
    ];
    GoAPI.getWeakPointStats = async () => ({ Urgent: 1, PreExam: 1, Relaxed: 0 });
    GoAPI.getSubjects = async () => [
        { Name: '高等数学', MaterialCount: 5 },
        { Name: '大学物理', MaterialCount: 3 },
    ];
    GoAPI.getRecentRecords = async () => [
        { Date: '2026-07-16', Subject: '高数', Content: '多元函数微分习题' },
    ];
    GoAPI.getRecentDiaries = async () => [
        { Date: '2026-07-16', WordCount: 120, Content: '今天复习了多元函数微分...' },
    ];
    GoAPI.getMemos = async () => [
        { Content: '学分认定申请截止 7月30日' },
    ];
    GoAPI.ping = () => 'dev-mode';
    GoAPI.getDataDir = () => '~/.study (dev)';
}

// 全局刷新
async function refreshDashboard() {
    if (!GoAPI.isReady()) return;
    try {
        AppState.dashboard = await GoAPI.getDashboard();
        AppState.lastRefresh = Date.now();
    } catch (err) {
        console.error('刷新 Dashboard 失败:', err);
    }
}

// 启动应用
document.addEventListener('DOMContentLoaded', initApp);
