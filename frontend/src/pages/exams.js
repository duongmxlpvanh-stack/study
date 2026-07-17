/**
 * 考试管理页 — 列表 + 添加 + 删除 + 倒计时颜色
 */
registerPage('exams', async function(container) {
    container.innerHTML = '<h2>⏰ 考试管理</h2><div id="exams-content"><p class="dim">加载中...</p></div>';
    await renderExams(container);
});

async function renderExams(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p class="error">后端未连接</p>'; return; }

    try {
        const exams = await GoAPI.getExams();
        let html = '<h2>⏰ 考试管理</h2>';

        // 添加表单
        html += `
        <div class="section">
            <h3>➕ 添加考试</h3>
            <form id="add-exam-form" class="inline-form" onsubmit="return false;">
                <input type="text" id="exam-name" placeholder="考试名称（如：期末考试）" required>
                <input type="date" id="exam-date" required>
                <button type="submit" class="btn btn-primary">添加</button>
            </form>
            <p id="exam-add-msg" class="dim" style="margin-top:8px"></p>
        </div>`;

        // 考试列表
        html += '<div class="section"><h3>📋 考试列表</h3>';
        if (!exams || exams.length === 0) {
            html += '<p class="empty">暂无考试</p>';
        } else {
            html += '<div class="list">';
            exams.forEach((e, i) => {
                let cls = 'ok';
                let icon = '🟢';
                if (e.DaysLeft < 0) { cls = 'dim'; icon = '⚫'; }
                else if (e.DaysLeft <= 7) { cls = 'urgent'; icon = '🔴'; }
                else if (e.DaysLeft <= 30) { cls = 'warn'; icon = '🟡'; }

                const daysText = e.DaysLeft < 0 ? '已结束' : `剩余 <strong class="${cls}">${e.DaysLeft}</strong> 天`;

                html += `<div class="list-item">
                    <span>${icon}</span>
                    <span style="flex:1"><strong>${escapeHtml(e.Name)}</strong></span>
                    <span class="dim">${e.Date}</span>
                    <span>${daysText}</span>
                    <span>${e.UrgencyStr || ''}</span>
                    <button class="btn btn-sm btn-danger" onclick="handleDeleteExam(${i + 1}, '${escapeHtml(e.Name)}')" title="删除考试">✕</button>
                </div>`;
            });
            html += '</div>';
        }
        html += '</div>';

        container.innerHTML = html;

        // 绑定添加事件
        const form = document.getElementById('add-exam-form');
        if (form) {
            form.onsubmit = async () => {
                const name = document.getElementById('exam-name').value.trim();
                const date = document.getElementById('exam-date').value;
                const msg = document.getElementById('exam-add-msg');
                if (!name || !date) {
                    msg.textContent = '请填写考试名称和日期';
                    msg.className = 'error';
                    return;
                }
                try {
                    await GoAPI.addExam(name, date);
                    msg.textContent = `已添加: ${name} (${date})`;
                    msg.className = 'ok';
                    form.reset();
                    showToast('考试添加成功', 'success');
                    // 刷新列表
                    setTimeout(() => renderExams(container), 400);
                } catch (err) {
                    msg.textContent = `添加失败: ${err}`;
                    msg.className = 'error';
                    showToast('添加考试失败', 'error');
                }
            };
        }

    } catch (err) {
        container.innerHTML = `<p class="error">加载失败: ${err}</p>`;
    }
}

// 删除考试（由列表中的按钮调用）
async function handleDeleteExam(index, name) {
    if (!confirm(`确定要删除考试 "${name}" 吗？此操作不可撤销。`)) return;
    try {
        await GoAPI.deleteExam(index);
        showToast(`已删除: ${name}`, 'success');
        // 刷新当前页面
        const content = document.getElementById('content');
        renderExams(content);
    } catch (err) {
        showToast(`删除失败: ${err}`, 'error');
    }
}
