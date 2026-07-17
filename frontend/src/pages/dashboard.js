/**
 * Dashboard 页面渲染
 *
 * 依赖 app.js 中的 GoAPI 和 AppState
 */

async function renderDashboard(container) {
    // 先用缓存数据立即渲染，避免闪烁
    if (AppState.dashboard) {
        paintDashboard(container, AppState.dashboard);
    } else {
        container.innerHTML = '<div style="padding:40px;text-align:center;color:var(--text-dim)">加载仪表板...</div>';
    }

    // 异步刷新最新数据
    if (GoAPI.isReady()) {
        try {
            AppState.dashboard = await GoAPI.getDashboard();
            AppState.lastRefresh = Date.now();
            paintDashboard(container, AppState.dashboard);
        } catch (err) {
            if (!AppState.dashboard) {
                container.innerHTML = `<div class="error">加载失败: ${err}</div>`;
            }
        }
    }
}

function paintDashboard(container, d) {
    if (!d) {
        container.innerHTML = '<p class="empty">暂无数据</p>';
        return;
    }

    let html = '<h2>📋 学习仪表板</h2>';

    // === 1. 统计卡片 ===
    const stats = d.StudyStats || {};
    html += '<div class="stat-cards">';
    html += statCard('累计学习', `${stats.TotalDays || 0} 天`);
    html += statCard('总记录', `${stats.TotalRecords || 0} 条`);
    html += statCard('连续学习', `${stats.StreakDays || 0} 天`);
    html += statCard('日均记录', `${(stats.AvgPerDay || 0).toFixed(1)} 条`);
    html += '</div>';

    // === 2. 考试倒计时 ===
    html += '<div class="section"><h3>⏰ 考试倒计时</h3>';
    const exams = d.Exams || [];
    if (exams.length === 0) {
        html += '<p class="empty">暂无考试</p>';
    } else {
        exams.forEach(e => {
            let cls = 'ok';
            if (e.DaysLeft < 0) cls = 'dim';
            else if (e.DaysLeft <= 7) cls = 'urgent';
            else if (e.DaysLeft <= 30) cls = 'warn';
            html += `<div class="list-item">
                <span>${e.Name}</span>
                <span class="dim">${e.Date}</span>
                <span>剩余 <strong class="${cls}">${e.DaysLeft}</strong> 天</span>
                <span>${e.UrgencyStr}</span>
            </div>`;
        });
    }
    html += '</div>';

    // === 3. 薄弱点统计 ===
    html += '<div class="section"><h3>🎯 薄弱知识点</h3>';
    const wp = d.WeakPointStats || {};
    const total = (wp.Urgent || 0) + (wp.PreExam || 0) + (wp.Relaxed || 0);
    if (total === 0) {
        html += '<p class="empty">暂无薄弱点记录</p>';
    } else {
        html += '<div class="stat-summary">';
        html += `<span class="urgent">🔴 紧急: <strong>${wp.Urgent || 0}</strong></span>`;
        html += `<span class="warn">🟡 考前看: <strong>${wp.PreExam || 0}</strong></span>`;
        html += `<span class="dim">🟢 不急: <strong>${wp.Relaxed || 0}</strong></span>`;
        html += `<span>总计: <strong>${total}</strong> 条</span>`;
        html += '</div>';
    }
    html += '</div>';

    // === 4. 课程概览 ===
    html += '<div class="section"><h3>📚 课程概览</h3>';
    const subjects = d.Subjects || [];
    if (subjects.length === 0) {
        html += '<p class="empty">暂无课程</p>';
    } else {
        subjects.forEach(s => {
            html += `<div class="list-item">
                <span>${s.Name}</span>
                <span class="dim">${s.MaterialCount} 份资料</span>
            </div>`;
        });
    }
    html += '</div>';

    // === 5. 最近学习记录 ===
    html += '<div class="section"><h3>📝 最近学习</h3>';
    const records = d.RecentRecords || [];
    if (records.length === 0) {
        html += '<p class="empty">暂无学习记录</p>';
    } else {
        records.forEach(r => {
            html += `<div class="list-item">
                <span class="dim">${r.Date}</span>
                <span class="subject">${r.Subject}</span>
                <span>${r.Content}</span>
            </div>`;
        });
    }
    html += '</div>';

    // === 6. 最近日记 ===
    html += '<div class="section"><h3>📖 最近日记</h3>';
    const diaries = d.RecentDiaries || [];
    if (diaries.length === 0) {
        html += '<p class="empty">暂无日记</p>';
    } else {
        diaries.forEach(d => {
            const preview = (d.Content || '').substring(0, 80);
            html += `<div class="list-item">
                <span class="dim">${d.Date}</span>
                <span class="dim">${d.WordCount}字</span>
                <span>${preview}${d.Content && d.Content.length > 80 ? '...' : ''}</span>
            </div>`;
        });
    }
    html += '</div>';

    container.innerHTML = html;
}

function statCard(label, value) {
    return `<div class="stat-card">
        <div class="stat-label">${label}</div>
        <div class="stat-value">${value}</div>
    </div>`;
}
