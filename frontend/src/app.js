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
    async getDiary(date) {
        return await this.handler.GetDiary(date);
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
    async searchMemo(keyword) {
        return await this.handler.SearchMemo(keyword);
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

    // 同步
    async getSyncStatus() {
        return await this.handler.GetSyncStatus();
    },
    async triggerSync() {
        return await this.handler.TriggerSync();
    },
};

// ==================== 页面注册表 ====================

// 每个页面模块在此注册自己的 render 函数
const PageRegistry = {};

function registerPage(name, renderFn) {
    PageRegistry[name] = renderFn;
}

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
    const renderFn = PageRegistry[page];

    if (renderFn) {
        renderFn(content);
    } else {
        content.innerHTML = `<p class="empty">页面 "${page}" 未注册</p>`;
    }
}

// ==================== 开发模式 ====================

// 开发模式：用静态数据展示 UI（无需 Go 后端运行时）
function useDevMode() {
    // 覆盖 GoAPI 方法，返回模拟数据
    GoAPI.isReady = () => true;
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
        { Name: '六级', Date: '2026-12-15', DaysLeft: 151, UrgencyStr: '🟢 充裕' },
    ];
    GoAPI.getWeakPoints = async () => [
        { Content: '泰勒展开的误差估计', Urgency: '紧急', Subject: '高数' },
        { Content: '麦克斯韦方程组推导', Urgency: '考前看', Subject: '大物' },
        { Content: '傅里叶变换应用', Urgency: '不急', Subject: '信号' },
    ];
    GoAPI.getWeakPointStats = async () => ({ Urgent: 1, PreExam: 1, Relaxed: 1 });
    GoAPI.getSubjects = async () => [
        { Name: '高等数学', MaterialCount: 5 },
        { Name: '大学物理', MaterialCount: 3 },
    ];
    GoAPI.getRecentRecords = async () => [
        { Date: '2026-07-16', Subject: '高数', Content: '多元函数微分习题' },
        { Date: '2026-07-15', Subject: '物理', Content: '电磁感应复习' },
        { Date: '2026-07-14', Subject: '英语', Content: '六级阅读练习' },
    ];
    GoAPI.getRecentDiaries = async () => [
        { Date: '2026-07-16', WordCount: 120, Content: '今天复习了多元函数微分...' },
    ];
    GoAPI.getDiary = async (date) => ({ Date: date, WordCount: 120, Content: `这是 ${date} 的日记内容。\n\n今天学习了多元函数的链式法则，感觉对方向导数的理解更深了。还做了一些练习题巩固。` });
    GoAPI.getMemos = async () => [
        { Content: '学分认定申请截止 7月30日' },
        { Content: '教材费退费到账确认' },
    ];
    GoAPI.getHeatmap = async () => {
        const days = [];
        for (let i = 139; i >= 0; i--) {
            const d = new Date();
            d.setDate(d.getDate() - i);
            const dateStr = d.toISOString().split('T')[0];
            days.push({ Date: dateStr, Count: Math.random() > 0.4 ? Math.floor(Math.random() * 5) + 1 : 0 });
        }
        return days;
    };
    GoAPI.getStreak = async () => ({ TotalDays: 120, TotalRecords: 342, StreakDays: 15, AvgPerDay: 2.85 });
    GoAPI.ping = () => 'dev-mode';
    GoAPI.getDataDir = () => '~/.study (dev)';
    GoAPI.getSyncStatus = async () => ({ Synced: true, LastSync: '2026-07-16 10:30', Ahead: 0, Behind: 0 });
    GoAPI.triggerSync = async () => '开发模式，无需同步';
    GoAPI.addExam = async (name, date) => { console.log('[dev] addExam:', name, date); };
    GoAPI.deleteExam = async (index) => { console.log('[dev] deleteExam:', index); };
    GoAPI.logRecord = async (input) => { console.log('[dev] logRecord:', input); };
    GoAPI.addWeakPoint = async (content, urgency, subject) => { console.log('[dev] addWeakPoint:', content, urgency, subject); };
    GoAPI.deleteWeakPoint = async (index) => { console.log('[dev] deleteWeakPoint:', index); };
    GoAPI.addSubject = async (name) => { console.log('[dev] addSubject:', name); };
    GoAPI.addMemo = async (content) => { console.log('[dev] addMemo:', content); };
    GoAPI.deleteMemo = async (index) => { console.log('[dev] deleteMemo:', index); };
    GoAPI.searchMemo = async (keyword) => [
        { Content: `搜索结果: ${keyword} 相关备忘` },
    ];
    GoAPI.searchDiary = async (keyword) => [
        { Date: '2026-07-16', WordCount: 120, Content: `关于 "${keyword}" 的日记内容...` },
    ];
}

// ==================== 全局刷新 ====================

async function refreshDashboard() {
    if (!GoAPI.isReady()) return;
    try {
        AppState.dashboard = await GoAPI.getDashboard();
        AppState.lastRefresh = Date.now();
    } catch (err) {
        console.error('刷新 Dashboard 失败:', err);
    }
}

// ==================== 工具函数 ====================

// 显示提示消息
function showToast(message, type) {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type || 'info'}`;
    toast.textContent = message;
    toast.style.cssText = `
        position: fixed; bottom: 20px; right: 20px; z-index: 9999;
        padding: 10px 20px; border-radius: 8px; font-size: 14px;
        color: #fff; animation: slideIn 0.3s ease;
        ${type === 'error' ? 'background: var(--urgent);' : type === 'success' ? 'background: var(--ok); color: #1e1e2e;' : 'background: var(--accent);'}
    `;
    document.body.appendChild(toast);
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transition = 'opacity 0.3s';
        setTimeout(() => toast.remove(), 300);
    }, 2500);
}

// HTML 转义
function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
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

// 启动应用
document.addEventListener('DOMContentLoaded', initApp);
