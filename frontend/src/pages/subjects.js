/**
 * 科目管理页 — 列表 + 添加
 */
registerPage('subjects', async function(container) {
    container.innerHTML = '<h2>📚 科目管理</h2><div id="subjects-content"><p class="dim">加载中...</p></div>';
    await renderSubjects(container);
});

async function renderSubjects(container) {
    if (!GoAPI.isReady()) { container.innerHTML = '<p class="error">后端未连接</p>'; return; }

    try {
        const subjects = await GoAPI.getSubjects();

        let html = '<h2>📚 科目管理</h2>';

        // 添加表单
        html += `
        <div class="section">
            <h3>➕ 添加科目</h3>
            <p class="dim" style="margin-bottom:8px;font-size:12px">添加后会自动在资料目录创建对应文件夹</p>
            <form id="add-subject-form" class="inline-form" onsubmit="return false;">
                <input type="text" id="subject-name" placeholder="科目名称（如：高等数学）" required>
                <button type="submit" class="btn btn-primary">添加</button>
            </form>
            <p id="subject-add-msg" class="dim" style="margin-top:8px"></p>
        </div>`;

        // 科目列表
        html += '<div class="section"><h3>📋 科目列表</h3>';
        if (!subjects || subjects.length === 0) {
            html += '<p class="empty">暂无科目，添加你的第一个科目吧！</p>';
        } else {
            html += `<p class="dim" style="margin-bottom:10px">共 ${subjects.length} 个科目</p>`;
            html += '<div class="list">';
            subjects.forEach(s => {
                html += `<div class="list-item">
                    <span>📖</span>
                    <span style="flex:1"><strong>${escapeHtml(s.Name)}</strong></span>
                    <span class="dim">📁 ${s.MaterialCount || 0} 份资料</span>
                    <span class="dim" style="font-size:12px">资料目录: materials/${escapeHtml(s.Name)}/</span>
                </div>`;
            });
            html += '</div>';
        }
        html += '</div>';

        container.innerHTML = html;

        // 绑定添加事件
        const form = document.getElementById('add-subject-form');
        if (form) {
            form.onsubmit = async () => {
                const name = document.getElementById('subject-name').value.trim();
                const msg = document.getElementById('subject-add-msg');
                if (!name) {
                    msg.textContent = '请输入科目名称';
                    msg.className = 'error';
                    return;
                }
                try {
                    await GoAPI.addSubject(name);
                    msg.textContent = `已添加科目: ${name}`;
                    msg.className = 'ok';
                    form.reset();
                    showToast('科目添加成功', 'success');
                    setTimeout(() => renderSubjects(container), 400);
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
