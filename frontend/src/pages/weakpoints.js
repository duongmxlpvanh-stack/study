/**
 * 薄弱知识点管理页 — 按紧急程度分组 + 添加 + 删除
 */
registerPage('weakpoints', async function(container) {
    container.innerHTML = '<h2>🎯 薄弱知识点</h2><div id="wp-content"><p class="dim">加载中...</p></div>';
    await renderWeakPoints(container);
});

async function renderWeakPoints(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p class="error">后端未连接</p>'; return; }

    try {
        const [wps, stats] = await Promise.all([GoAPI.getWeakPoints(), GoAPI.getWeakPointStats()]);

        let html = '<h2>🎯 薄弱知识点</h2>';

        // 添加表单
        html += `
        <div class="section">
            <h3>➕ 添加薄弱点</h3>
            <form id="add-wp-form" class="inline-form" onsubmit="return false;">
                <input type="text" id="wp-content" placeholder="知识点内容" style="flex:2" required>
                <input type="text" id="wp-subject" placeholder="关联科目（可选）" style="flex:1">
                <select id="wp-urgency" required>
                    <option value="">-- 紧急程度 --</option>
                    <option value="紧急">🔴 紧急</option>
                    <option value="考前看">🟡 考前看</option>
                    <option value="不急">🟢 不急</option>
                </select>
                <button type="submit" class="btn btn-primary">添加</button>
            </form>
            <p id="wp-add-msg" class="dim" style="margin-top:8px"></p>
        </div>`;

        // 统计概览
        const total = (stats.Urgent || 0) + (stats.PreExam || 0) + (stats.Relaxed || 0);
        html += '<div class="section"><h3>📊 统计概览</h3>';
        html += '<div class="stat-summary">';
        html += `<span class="urgent">🔴 紧急: <strong>${stats.Urgent || 0}</strong></span>`;
        html += `<span class="warn">🟡 考前看: <strong>${stats.PreExam || 0}</strong></span>`;
        html += `<span class="dim">🟢 不急: <strong>${stats.Relaxed || 0}</strong></span>`;
        html += `<span>📋 总计: <strong>${total}</strong> 条</span>`;
        html += '</div></div>';

        // 按紧急程度分组
        html += '<div class="section"><h3>📋 薄弱点列表</h3>';
        if (!wps || wps.length === 0) {
            html += '<p class="empty">暂无薄弱点记录</p>';
        } else {
            const groups = { '紧急': [], '考前看': [], '不急': [] };
            wps.forEach(w => {
                const key = w.Urgency || '不急';
                if (groups[key]) groups[key].push(w);
                else groups['不急'].push(w);
            });

            for (const [urgency, items] of Object.entries(groups)) {
                if (items.length === 0) continue;
                let icon = '🟢';
                if (urgency === '紧急') icon = '🔴';
                else if (urgency === '考前看') icon = '🟡';

                html += `<h4 style="margin:12px 0 6px">${icon} ${urgency} (${items.length}条)</h4>`;
                html += '<div class="list">';
                items.forEach((w, i) => {
                    // 用全局序号：在全部 wps 数组中的索引 + 1
                    const globalIndex = wps.indexOf(w) + 1;
                    html += `<div class="list-item">
                        <span style="min-width:24px;text-align:right;color:var(--text-dim)">${globalIndex}.</span>
                        <span style="flex:1">${escapeHtml(w.Content || '')}</span>
                        <span class="subject">${escapeHtml(w.Subject || '')}</span>
                        <button class="btn btn-sm btn-danger" onclick="handleDeleteWP(${globalIndex}, '${escapeHtml(w.Content || '').replace(/'/g, "\\'")}')" title="删除">✕</button>
                    </div>`;
                });
                html += '</div>';
            }
        }
        html += '</div>';

        container.innerHTML = html;

        // 绑定添加事件
        const form = document.getElementById('add-wp-form');
        if (form) {
            form.onsubmit = async () => {
                const content = document.getElementById('wp-content').value.trim();
                const subject = document.getElementById('wp-subject').value.trim();
                const urgency = document.getElementById('wp-urgency').value;
                const msg = document.getElementById('wp-add-msg');
                if (!content || !urgency) {
                    msg.textContent = '请填写知识点内容和选择紧急程度';
                    msg.className = 'error';
                    return;
                }
                try {
                    await GoAPI.addWeakPoint(content, urgency, subject);
                    msg.textContent = `已添加: ${content}`;
                    msg.className = 'ok';
                    form.reset();
                    showToast('薄弱点添加成功', 'success');
                    setTimeout(() => renderWeakPoints(container), 400);
                } catch (err) {
                    msg.textContent = `添加失败: ${err}`;
                    msg.className = 'error';
                    showToast('添加失败', 'error');
                }
            };
        }

    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

// 删除薄弱点
async function handleDeleteWP(index, content) {
    const preview = (content || '').substring(0, 30);
    if (!confirm(`确定要删除薄弱点 #${index} "${preview}" 吗？`)) return;
    try {
        await GoAPI.deleteWeakPoint(index);
        showToast(`已删除: ${preview}`, 'success');
        const container = document.getElementById('content');
        renderWeakPoints(container);
    } catch (err) {
        showToast(`删除失败: ${err}`, 'error');
    }
}
