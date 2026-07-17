/**
 * 备忘管理页 — 添加 + 删除 + 搜索
 */
registerPage('memos', async function(container) {
    container.innerHTML = '<h2>📋 备忘</h2><div id="memos-content"><p class="dim">加载中...</p></div>';
    await renderMemos(container);
});

async function renderMemos(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p class="error">后端未连接</p>'; return; }

    try {
        const memos = await GoAPI.getMemos();

        let html = '<h2>📋 行政备忘</h2>';

        // 添加 + 搜索
        html += `
        <div class="section">
            <h3>➕ 添加备忘</h3>
            <form id="add-memo-form" class="inline-form" onsubmit="return false;">
                <input type="text" id="memo-content" placeholder="备忘内容..." style="flex:2" required>
                <button type="submit" class="btn btn-primary">添加</button>
            </form>
            <p id="memo-add-msg" class="dim" style="margin-top:8px"></p>
            <hr style="margin:12px 0;border-color:var(--border)">
            <form id="memo-search-form" class="inline-form" onsubmit="return false;">
                <input type="text" id="memo-search-input" placeholder="搜索备忘..." style="flex:2">
                <button type="submit" class="btn btn-primary">🔍</button>
                <button type="button" id="memo-clear-btn" class="btn btn-outline" style="display:none">清除</button>
            </form>
            <p id="memo-search-info" class="dim" style="margin-top:8px"></p>
        </div>`;

        // 备忘列表
        html += '<div class="section"><h3>📋 备忘列表</h3>';
        html += '<div id="memo-list-container"></div>';
        html += '</div>';

        container.innerHTML = html;

        // 渲染列表
        renderMemoList(memos || [], `共 ${(memos || []).length} 条备忘`);

        // 绑定添加
        const addForm = document.getElementById('add-memo-form');
        if (addForm) {
            addForm.onsubmit = async () => {
                const content = document.getElementById('memo-content').value.trim();
                const msg = document.getElementById('memo-add-msg');
                if (!content) {
                    msg.textContent = '请输入备忘内容';
                    msg.className = 'error';
                    return;
                }
                try {
                    await GoAPI.addMemo(content);
                    msg.textContent = '备忘已添加';
                    msg.className = 'ok';
                    document.getElementById('memo-content').value = '';
                    showToast('备忘添加成功', 'success');
                    // 刷新
                    const memos = await GoAPI.getMemos();
                    renderMemoList(memos || [], `共 ${(memos || []).length} 条备忘`);
                } catch (err) {
                    msg.textContent = `添加失败: ${err}`;
                    msg.className = 'error';
                    showToast('添加失败', 'error');
                }
            };
        }

        // 绑定搜索
        const searchForm = document.getElementById('memo-search-form');
        const searchInput = document.getElementById('memo-search-input');
        const clearBtn = document.getElementById('memo-clear-btn');
        const searchInfo = document.getElementById('memo-search-info');

        if (searchForm) {
            searchForm.onsubmit = async () => {
                const keyword = searchInput.value.trim();
                if (!keyword) {
                    searchInfo.textContent = '请输入搜索关键词';
                    searchInfo.className = 'error';
                    return;
                }
                try {
                    const results = await GoAPI.searchMemo(keyword);
                    if (!results || results.length === 0) {
                        searchInfo.textContent = `没有找到与 "${keyword}" 相关的备忘`;
                        searchInfo.className = 'dim';
                        renderMemoList([], '');
                    } else {
                        searchInfo.textContent = `找到 ${results.length} 条备忘`;
                        searchInfo.className = 'ok';
                        renderMemoList(results, `搜索结果: "${keyword}"`);
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
                const memos = await GoAPI.getMemos();
                renderMemoList(memos || [], `共 ${(memos || []).length} 条备忘`);
            };
        }

    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

function renderMemoList(memos, title) {
    const container = document.getElementById('memo-list-container');
    if (!container) return;

    if (!memos || memos.length === 0) {
        container.innerHTML = '<p class="empty">暂无备忘，添加你的第一条备忘吧！</p>';
        return;
    }

    let html = '';
    if (title) {
        html += `<p class="dim" style="margin-bottom:10px">${title}</p>`;
    }
    html += '<div class="list">';
    memos.forEach((m, i) => {
        html += `<div class="list-item">
            <span style="min-width:20px;color:var(--text-dim)">${i + 1}.</span>
            <span style="flex:1">${escapeHtml(m.Content || '')}</span>
            <button class="btn btn-sm btn-danger" onclick="handleDeleteMemo(${i + 1}, '${escapeHtml((m.Content || '').substring(0, 30))}')" title="删除备忘">✕</button>
        </div>`;
    });
    html += '</div>';
    container.innerHTML = html;
}

// 删除备忘
async function handleDeleteMemo(index, preview) {
    if (!confirm(`确定要删除备忘 "${preview}..." 吗？`)) return;
    try {
        await GoAPI.deleteMemo(index);
        showToast('备忘已删除', 'success');
        // 刷新
        const memos = await GoAPI.getMemos();
        renderMemoList(memos || [], `共 ${(memos || []).length} 条备忘`);
    } catch (err) {
        showToast(`删除失败: ${err}`, 'error');
    }
}
