/**
 * 日记搜索页 — 搜索框 + 结果列表 + 全文预览弹窗
 */
registerPage('diary', async function(container) {
    container.innerHTML = '<h2>📖 日记</h2><div id="diary-content"><p class="dim">加载中...</p></div>';
    await renderDiary(container);
});

async function renderDiary(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p class="error">后端未连接</p>'; return; }

    try {
        const diaries = await GoAPI.getRecentDiaries(20);

        let html = '<h2>📖 学习日记</h2>';

        // 搜索框
        html += `
        <div class="section">
            <form id="diary-search-form" class="inline-form" onsubmit="return false;">
                <input type="text" id="diary-search-input" placeholder="全文搜索日记内容..." style="flex:2">
                <button type="submit" class="btn btn-primary">🔍 搜索</button>
                <button type="button" id="diary-clear-btn" class="btn btn-outline" style="display:none">清除</button>
            </form>
            <p id="diary-search-info" class="dim" style="margin-top:8px"></p>
        </div>`;

        // 日记列表
        html += '<div class="section"><h3>📋 最近日记</h3>';
        html += '<div id="diary-list-container"></div>';
        html += '</div>';

        // 日记详情弹出框
        html += `
        <div id="diary-modal" class="modal" style="display:none">
            <div class="modal-overlay" onclick="closeDiaryModal()"></div>
            <div class="modal-content">
                <div class="modal-header">
                    <span id="diary-modal-title">日记详情</span>
                    <button class="btn btn-sm" onclick="closeDiaryModal()">✕</button>
                </div>
                <div id="diary-modal-body" class="modal-body"></div>
            </div>
        </div>`;

        container.innerHTML = html;

        // 渲染列表
        renderDiaryList(diaries || [], '最近 20 篇日记');

        // 绑定搜索
        const searchForm = document.getElementById('diary-search-form');
        const searchInput = document.getElementById('diary-search-input');
        const clearBtn = document.getElementById('diary-clear-btn');
        const searchInfo = document.getElementById('diary-search-info');

        if (searchForm) {
            searchForm.onsubmit = async () => {
                const keyword = searchInput.value.trim();
                if (!keyword) {
                    searchInfo.textContent = '请输入搜索关键词';
                    searchInfo.className = 'error';
                    return;
                }
                try {
                    searchInfo.textContent = '搜索中...';
                    searchInfo.className = 'dim';
                    const results = await GoAPI.searchDiary(keyword);
                    if (!results || results.length === 0) {
                        searchInfo.textContent = `没有找到与 "${keyword}" 相关的日记`;
                        searchInfo.className = 'dim';
                        renderDiaryList([], '');
                    } else {
                        searchInfo.textContent = `找到 ${results.length} 条相关日记`;
                        searchInfo.className = 'ok';
                        renderDiaryList(results, `搜索结果: "${keyword}"`);
                        clearBtn.style.display = 'inline-block';
                    }
                } catch (err) {
                    searchInfo.textContent = `搜索失败: ${err}`;
                    searchInfo.className = 'error';
                }
            };
        }

        if (clearBtn) {
            clearBtn.onclick = async () => {
                searchInput.value = '';
                searchInfo.textContent = '';
                searchInfo.className = 'dim';
                clearBtn.style.display = 'none';
                const recent = await GoAPI.getRecentDiaries(20);
                renderDiaryList(recent || [], '最近 20 篇日记');
            };
        }

    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

function renderDiaryList(diaries, title) {
    const container = document.getElementById('diary-list-container');
    if (!container) return;

    if (!diaries || diaries.length === 0) {
        container.innerHTML = '<p class="empty">暂无日记</p>';
        return;
    }

    let html = '';
    if (title) {
        html += `<p class="dim" style="margin-bottom:10px">${title} (${diaries.length}篇)</p>`;
    }
    html += '<div class="list">';
    diaries.forEach(d => {
        const preview = (d.Content || '').substring(0, 100);
        html += `<div class="list-item diary-item" onclick="openDiaryDetail('${d.Date}', ${d.WordCount || 0})" style="cursor:pointer">
            <span class="dim">${d.Date}</span>
            <span class="dim">${d.WordCount || 0}字</span>
            <span style="flex:1">${escapeHtml(preview)}${(d.Content && d.Content.length > 100) ? '...' : ''}</span>
            <span class="dim" style="font-size:12px">点击查看 →</span>
        </div>`;
    });
    html += '</div>';
    container.innerHTML = html;
}

// 查看日记详情
async function openDiaryDetail(date, wordCount) {
    const modal = document.getElementById('diary-modal');
    const title = document.getElementById('diary-modal-title');
    const body = document.getElementById('diary-modal-body');
    if (!modal || !body) return;

    title.textContent = `📖 ${date} (${wordCount}字)`;
    body.innerHTML = '<p class="dim">加载中...</p>';
    modal.style.display = 'flex';

    try {
        const diary = await GoAPI.getDiary(date);
        if (diary && diary.Content) {
            // 将换行转为 <br>，简单 markdown 样式处理
            const formatted = escapeHtml(diary.Content)
                .replace(/\n/g, '<br>')
                .replace(/^# (.+)$/gm, '<h4>$1</h4>')
                .replace(/^- (.+)$/gm, '• $1');
            body.innerHTML = `<div style="line-height:1.8;white-space:pre-wrap">${formatted}</div>`;
        } else {
            body.innerHTML = '<p class="empty">日记内容为空</p>';
        }
    } catch (err) {
        body.innerHTML = `<p class="error">加载日记失败: ${err}</p>`;
    }
}

function closeDiaryModal() {
    const modal = document.getElementById('diary-modal');
    if (modal) modal.style.display = 'none';
}
