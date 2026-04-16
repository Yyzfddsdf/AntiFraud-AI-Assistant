import { reactive, ref } from 'vue';

export function useTasksModule(deps) {
  const analyzing = ref(false);
  const tasks = ref([]);
  const history = ref([]);
  const selectedTask = ref(null);
  const deletingHistory = reactive({});
  const analyzeForm = reactive({
    text: '',
    videos: [],
    audios: [],
    images: []
  });

  const handleFileSelect = async (event, type) => {
    const files = Array.from(event.target.files || []);
    if (files.length === 0) return;

    try {
      const results = await Promise.all(files.map((file) => deps.fileToBase64(file)));
      analyzeForm[type] = [...analyzeForm[type], ...results];
      deps.showToast(`已添加 ${files.length} 个文件`);
    } catch (error) {
      console.error('read file failed:', error);
      deps.showToast('文件读取失败', 'error');
    }
  };

  const submitAnalysis = async () => {
    if (!analyzeForm.text && analyzeForm.videos.length === 0 && analyzeForm.audios.length === 0 && analyzeForm.images.length === 0) {
      deps.showToast('请至少提供一种输入（文本或文件）', 'error');
      return;
    }

    analyzing.value = true;
    const res = await deps.request('/scam/multimodal/analyze', 'POST', analyzeForm);
    analyzing.value = false;

    if (res) {
      deps.showToast('任务已提交');
      analyzeForm.text = '';
      analyzeForm.videos = [];
      analyzeForm.audios = [];
      analyzeForm.images = [];
      deps.activeTab.value = 'tasks';
      fetchTasks();
    }
  };

  const fetchTasks = async ({ silent = false } = {}) => {
    if (!deps.isAuthenticated.value) return;
    const res = await deps.request('/scam/multimodal/tasks', 'GET', null, { silent });
    if (res && Array.isArray(res.tasks)) {
      const nextTasks = [...res.tasks].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
      deps.replaceListIfChanged(tasks, nextTasks);
    }
  };

  const fetchHistory = async ({ silent = false } = {}) => {
    const res = await deps.request('/scam/multimodal/history', 'GET', null, { silent });
    if (res && Array.isArray(res.history)) {
      deps.replaceListIfChanged(history, res.history);
    }
  };

  const viewTaskDetail = async (taskId) => {
    const res = await deps.request(`/scam/multimodal/tasks/${taskId}`);
    if (res && res.task) {
      if (!res.task.risk_level && res.task.risk_summary) {
        try {
          const parsed = JSON.parse(res.task.risk_summary);
          if (parsed && parsed.risk_level) {
            res.task.risk_level = parsed.risk_level;
          }
        } catch (e) {
          // ignore parsing error
        }
      }
      selectedTask.value = res.task;
    }
  };

  const viewHistoryDetail = async (item) => {
    if (!item || !item.record_id) return;
    await viewTaskDetail(item.record_id);
    if (selectedTask.value) {
      if (!selectedTask.value.risk_level && item.risk_level) {
        selectedTask.value.risk_level = item.risk_level;
      }
      if (!selectedTask.value.scam_type && item.scam_type) {
        selectedTask.value.scam_type = item.scam_type;
      }
    }
  };

  const deleteHistoryCase = async (item) => {
    if (!item || !item.record_id) return;
    if (!confirm(`确定删除案件 ${item.record_id} 吗？此操作不可恢复。`)) return;

    deletingHistory[item.record_id] = true;
    try {
      const res = await deps.request(`/scam/multimodal/history/${encodeURIComponent(item.record_id)}`, 'DELETE');
      if (!res) return;

      fetchHistory();
      if (selectedTask.value && selectedTask.value.task_id === item.record_id) {
        selectedTask.value = null;
      }
      deps.showToast(res.message || '历史案件删除成功');
    } finally {
      deletingHistory[item.record_id] = false;
    }
  };

  const formatTime = (iso) => new Date(iso).toLocaleString('zh-CN', { hour12: false });

  const getStatusLabel = (status) => {
    const map = {
      pending: '等待中',
      processing: '分析中',
      completed: '已完成',
      failed: '失败'
    };
    return map[status] || status;
  };

  const getStatusClass = (status) => {
    const map = {
      pending: 'bg-yellow-100 text-yellow-800 px-2 py-1 rounded-full text-xs font-bold',
      processing: 'bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-xs font-bold',
      completed: 'bg-green-100 text-green-800 px-2 py-1 rounded-full text-xs font-bold',
      failed: 'bg-red-100 text-red-800 px-2 py-1 rounded-full text-xs font-bold'
    };
    return map[status] || 'bg-gray-100 text-gray-800 px-2 py-1 rounded-full text-xs font-bold';
  };

  const normalizeRiskLevelText = (level) => {
    if (!level) return '';
    return String(level).trim();
  };

  const getRiskClass = (level) => {
    const rawValue = String(level || '').trim().toLowerCase();
    
    // 如果是高风险相关的词，就标红
    if (['高', 'high', 'severe', 'critical'].includes(rawValue) || rawValue.includes('高')) return 'bg-red-100 text-red-800 border-red-200';
    // 中风险标黄
    if (['中', 'medium', 'warning'].includes(rawValue) || rawValue.includes('中')) return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    // 低风险/安全标绿
    if (['低', 'low', 'safe', 'none', '安全'].includes(rawValue) || rawValue.includes('低')) return 'bg-green-100 text-green-800 border-green-200';
    
    // 默认样式，对于一些自定义文本保持中性样式
    return 'bg-slate-100 text-slate-700 border-slate-200';
  };

  const openImage = (src) => {
    const win = window.open('', '_blank');
    if (!win) return;
    win.document.write(`<img src="${src}" style="max-width:100%; height:auto;">`);
  };

  const escapeHtml = (text) => String(text || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');

  const renderPlainText = (text) => escapeHtml(text).replace(/\r?\n/g, '<br>');

  const renderMarkdown = (text) => {
    if (!text) return '';

    const parser = window.marked && typeof window.marked.parse === 'function'
      ? window.marked
      : (typeof marked !== 'undefined' && marked && typeof marked.parse === 'function' ? marked : null);

    if (!parser) {
      return renderPlainText(text);
    }

    try {
      return parser.parse(text, {
        breaks: true,
        gfm: true
      });
    } catch (error) {
      console.error('Markdown parse error:', error);
      return renderPlainText(text);
    }
  };

  const parseReport = (text) => {
    if (!text) return [];
    const sections = [];
    const lines = text.split('\n');
    let currentSection = null;

    for (const line of lines) {
      const match = line.trim().match(/^(\d+)\.\s+(.+)$/);
      if (match) {
        if (currentSection) {
          currentSection.content = currentSection.content.trim();
          sections.push(currentSection);
        }
        currentSection = {
          id: parseInt(match[1], 10),
          title: match[2].trim(),
          content: ''
        };
      } else if (currentSection) {
        currentSection.content += `${line}\n`;
      }
    }

    if (currentSection) {
      currentSection.content = currentSection.content.trim();
      sections.push(currentSection);
    }

    return sections;
  };

  const extractAttackSteps = (reportText) => {
    if (!reportText) return [];
    const reportSections = parseReport(reportText);
    const attackSection = reportSections.find((section) => String(section?.title || '').trim().includes('诈骗链路还原'));
    if (!attackSection || !attackSection.content) return [];

    return attackSection.content
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => line.replace(/^[-*•]\s+/, '').replace(/^\d+[.)、]\s*/, '').trim())
      .filter(Boolean);
  };

  const extractScamKeywordSentences = (reportText) => {
    if (!reportText) return [];
    const reportSections = parseReport(reportText);
    const keywordSection = reportSections.find((section) => String(section?.title || '').trim().includes('诈骗关键词句'));
    if (!keywordSection || !keywordSection.content) return [];

    return keywordSection.content
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => line.replace(/^[-*•]\s+/, '').replace(/^\d+[.)、]\s*/, '').trim())
      .filter(Boolean);
  };

  const parseRiskSummary = (raw) => {
    if (!raw || !String(raw).trim()) return null;
    try {
      const parsed = JSON.parse(raw);
      return parsed && typeof parsed === 'object' ? parsed : null;
    } catch {
      return null;
    }
  };

  const parseInsight = (text) => {
    if (!text) return [];
    const sections = [];
    const regex = /^[【\[](.+?)[】\]]\s*(.*)$/;
    const lines = text.split('\n');
    let currentSection = null;

    for (const line of lines) {
      const trimmedLine = line.trim();
      const match = trimmedLine.match(regex);
      if (match) {
        if (currentSection) {
          currentSection.content = currentSection.content.trim();
          sections.push(currentSection);
        }
        currentSection = {
          title: match[1].trim(),
          content: match[2] ? `${match[2]}\n` : ''
        };
      } else if (currentSection) {
        currentSection.content += `${line}\n`;
      } else if (trimmedLine) {
        if (sections.length > 0 && sections[0].title === '概述') {
          sections[0].content += `${line}\n`;
        } else if (!currentSection) {
          currentSection = { title: '概述', content: `${line}\n` };
        }
      }
    }

    if (currentSection) {
      currentSection.content = currentSection.content.trim();
      sections.push(currentSection);
    }

    return sections;
  };

  const exportData = (type) => {
    if (!selectedTask.value) return;

    const task = selectedTask.value;
    const date = new Date().toISOString().slice(0, 10);
    const filename = `scam-report-${task.task_id}-${date}`;

    if (type === 'json') {
      const blob = new Blob([JSON.stringify(task, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${filename}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      return;
    }

    if (type === 'md') {
      let content = '# 诈骗风险分析报告\n\n';
      content += `**任务ID**: ${task.task_id}\n`;
      content += `**标题**: ${task.title}\n`;
      content += `**诈骗类型**: ${task.scam_type || '未识别'}\n`;
      content += `**生成时间**: ${new Date(task.created_at).toLocaleString()}\n`;
      content += `**状态**: ${task.status}\n\n`;

      if (task.report) {
        content += `## 综合分析报告\n${task.report}\n\n`;
      }

      if (task.payload) {
        if (task.payload.video_insights?.length) {
          content += '## 视频分析洞察\n';
          task.payload.video_insights.forEach((insight, idx) => {
            content += `### 视频 #${idx + 1}\n${insight}\n\n`;
          });
        }

        if (task.payload.audio_insights?.length) {
          content += '## 音频分析洞察\n';
          task.payload.audio_insights.forEach((insight, idx) => {
            content += `### 音频 #${idx + 1}\n${insight}\n\n`;
          });
        }

        if (task.payload.image_insights?.length) {
          content += '## 图片分析洞察\n';
          task.payload.image_insights.forEach((insight, idx) => {
            content += `### 图片 #${idx + 1}\n${insight}\n\n`;
          });
        }

        if (task.payload.text) {
          content += `## 原始文本证据\n${task.payload.text}\n\n`;
        }
      }

      const blob = new Blob([content], { type: 'text/markdown' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${filename}.md`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    }
  };

  const printReport = () => {
    window.print();
  };

  return {
    analyzing,
    tasks,
    history,
    selectedTask,
    deletingHistory,
    analyzeForm,
    handleFileSelect,
    submitAnalysis,
    fetchTasks,
    fetchHistory,
    viewTaskDetail,
    viewHistoryDetail,
    deleteHistoryCase,
    formatTime,
    getStatusLabel,
    getStatusClass,
    normalizeRiskLevelText,
    getRiskClass,
    openImage,
    renderMarkdown,
    parseReport,
    extractAttackSteps,
    extractScamKeywordSentences,
    parseRiskSummary,
    parseInsight,
    exportData,
    printReport
  };
}
