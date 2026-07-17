/**
 * 学习记录页 — 历史 + 添加记录表单
 */
registerPage('records', async function(container) {
    container.innerHTML = '<h2>📝 学习记录</h2><div id="records-content"><p class="dim">加载中...</p></div>';
    await renderRecords(container);
});

async function renderRecords(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p class="error">后端未连接</p>'; return; }

    try {
        const records = await GoAPI.getRecentRecords(50);
        let html = '<h2>📝 学习记录</h2>';

        // 添加快捷记录
        html += `
        <div class="section">
            <h3>✏️ 快速记录</h3>
            <p class="dim" style="margin-bottom:8px;font-size:12px">格式: "科目 内容" 或 "科目: 内容"（如: 高数 多元函数微分习题）</p>
            <form id="add-record-form" class="inline-form" onsubmit="return false;">
                <input type="text" id="record-input" placeholder="科目 学习内容" style="flex:2" required>
                <button type="submit" class="btn btn-primary">📝 记录</button>
            </form>
            <p id="record-add-msg" class="dim" style="margin-top:8px"></p>
        </div>`;

        // 记录列表
        html += '<div class="section"><h3>📋 最近学习记录</h3>';
        if (!records || records.length === 0) {
            html += '<p class="empty">暂无学习记录，开始你的第一条记录吧！</p>';
        } else {
            html += `<p class="dim" style="margin-bottom:12px">共 ${records.length} 条记录</p>`;
            html += '<div class="list">';
            records.forEach(r => {
                html += `<div class="list-item">
                    <span class="dim">${r.Date}</span>
                    <span class="subject">${escapeHtml(r.Subject || '—')}</span>
                    <span style="flex:1">${escapeHtml(r.Content || '')}</span>
                </div>`;
            });
            html += '</div>';
        }
        html += '</div>';

        container.innerHTML = html;

        // 绑定添加事件
        const form = document.getElementById('add-record-form');
        if (form) {
            form.onsubmit = async () => {
                const input = document.getElementById('record-input');
                const msg = document.getElementById('record-add-msg');
                const value = input.value.trim();
                if (!value) {
                    msg.textContent = '请输入学习内容';
                    msg.className = 'error';
                    return;
                }
                try {
                    await GoAPI.logRecord(value);
                    msg.textContent = '记录成功！';
                    msg.className = 'ok';
                    input.value = '';
                    showToast('学习记录已保存', 'success');
                    // 刷新列表
                    setTimeout(() => renderRecords(container), 400);
                } catch (err) {
                    msg.textContent = `记录失败: ${err}`;
                    msg.className = 'error';
                    showToast('记录失败', 'error');
                }
            };
        }

    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}
