import { ref } from 'vue';

const adminChatBasePath = '/admin/chat';
const adminChatWelcomeMessage = '你好！我是管理助手。我可以帮你查询全国、省、市、区县案件分布、地区Top诈骗类型，以及辅助平台治理分析。';

export function useChatModule(deps) {
  const showChat = ref(false);
  const chatMessages = ref([
    { type: 'ai', content: adminChatWelcomeMessage }
  ]);
  const chatInput = ref('');
  const chatImages = ref([]);
  const isChatting = ref(false);
  const chatHistoryLoaded = ref(false);

  const scrollToBottom = () => {
    const container = document.getElementById('chat-container');
    if (container) container.scrollTop = container.scrollHeight;
  };

  const fetchChatHistory = async (force = false) => {
    if (!deps.isAuthenticated.value) return;
    if (chatHistoryLoaded.value && !force) return;
    try {
      const data = await deps.request(`${adminChatBasePath}/context`);
      chatMessages.value = [
        { type: 'ai', content: adminChatWelcomeMessage }
      ];
      if (data.messages && Array.isArray(data.messages)) {
        const history = [];
        for (const msg of data.messages) {
          if (msg.role === 'assistant') {
            if (msg.tool_calls && Array.isArray(msg.tool_calls)) {
              for (const call of msg.tool_calls) {
                const toolName = call.name || call.function?.name || 'unknown';
                history.push({ type: 'tool', content: `正在调用工具: ${toolName}...` });
              }
            }
            if (msg.content) {
              history.push({ type: 'ai', content: msg.content });
            }
          } else if (msg.role === 'tool') {
            history.push({ type: 'tool', content: '工具调用完成' });
          } else if (msg.role === 'user') {
            const imageUrls = Array.isArray(msg.image_urls) ? msg.image_urls.filter(item => typeof item === 'string' && item.trim()) : [];
            if (!msg.content && imageUrls.length === 0) continue;
            history.push({ type: 'user', content: msg.content || '', images: imageUrls });
          }
        }
        if (history.length > 0) {
          chatMessages.value = [chatMessages.value[0], ...history];
        }
      }
      chatHistoryLoaded.value = true;
      setTimeout(scrollToBottom, 100);
    } catch (e) {
      console.error('Fetch chat history failed:', e);
    }
  };

  const toggleChat = () => {
    if (deps.hasMoved.value) return;
    showChat.value = !showChat.value;
    if (showChat.value) {
      if (!chatHistoryLoaded.value) {
        fetchChatHistory();
      }
      setTimeout(scrollToBottom, 100);
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
        currentSection = { id: parseInt(match[1], 10), title: match[2].trim(), content: '' };
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
    const attackSection = parseReport(reportText).find((section) => String(section?.title || '').trim().includes('诈骗链路还原'));
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
    const keywordSection = parseReport(reportText).find((section) => String(section?.title || '').trim().includes('诈骗关键词句'));
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
        currentSection = { title: match[1].trim(), content: match[2] ? `${match[2]}\n` : '' };
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

  const triggerChatImagePicker = () => {
    if (isChatting.value) return;
    const input = document.getElementById('chat-image-input');
    if (input) input.click();
  };

  const handleChatImageSelect = async (event) => {
    const files = Array.from(event.target.files || []).filter(file => String(file.type || '').startsWith('image/'));
    event.target.value = '';
    if (files.length === 0) return;
    try {
      const results = await Promise.all(files.map(file => deps.fileToBase64(file)));
      chatImages.value = [...chatImages.value, ...results];
      deps.showToast(`已添加 ${files.length} 张图片`);
    } catch (e) {
      console.error('Read chat images failed:', e);
      deps.showToast('图片读取失败', 'error');
    }
  };

  const removeChatImage = (index) => {
    chatImages.value = chatImages.value.filter((_, idx) => idx !== index);
  };

  const sendChatMessage = async () => {
    if (isChatting.value) return;
    const message = chatInput.value.trim();
    const images = [...chatImages.value];
    if (!message && images.length === 0) return;

    chatMessages.value.push({ type: 'user', content: message, images });
    chatInput.value = '';
    chatImages.value = [];
    isChatting.value = true;
    scrollToBottom();

    try {
      const response = await fetch(`/api${adminChatBasePath}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${deps.token.value}`
        },
        body: JSON.stringify({ message, images })
      });

      if (!response.ok) throw new Error('Network response was not ok');

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let aiMessageContent = '';
      chatMessages.value.push({ type: 'ai', content: '' });
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop();

        for (const line of lines) {
          const trimmedLine = line.trim();
          if (!trimmedLine || !trimmedLine.startsWith('data:')) continue;

          try {
            const data = JSON.parse(trimmedLine.slice(5).trim());
            const getActiveAiIndex = () => {
              const lastIdx = chatMessages.value.length - 1;
              if (lastIdx >= 0 && chatMessages.value[lastIdx].type === 'ai') {
                return lastIdx;
              }
              chatMessages.value.push({ type: 'ai', content: '' });
              return chatMessages.value.length - 1;
            };

            if (data.type === 'content') {
              const idx = getActiveAiIndex();
              aiMessageContent += data.content;
              chatMessages.value[idx].content = aiMessageContent;
              scrollToBottom();
            } else if (data.type === 'tool_call') {
              const toolName = data.tool;
              const lastIdx = chatMessages.value.length - 1;
              if (lastIdx >= 0 && chatMessages.value[lastIdx].type === 'ai' && !chatMessages.value[lastIdx].content) {
                chatMessages.value[lastIdx] = { type: 'tool', content: `正在调用工具: ${toolName}...` };
              } else {
                chatMessages.value.push({ type: 'tool', content: `正在调用工具: ${toolName}...` });
              }
              aiMessageContent = '';
              scrollToBottom();
            } else if (data.type === 'tool_result') {
              chatMessages.value.push({ type: 'tool', content: `工具 ${data.tool} 调用完成` });
              scrollToBottom();
            }
          } catch (e) {
            console.error('Error parsing SSE data:', e, 'Line:', trimmedLine);
          }
        }
      }
    } catch (error) {
      console.error('Chat error:', error);
      chatMessages.value.push({ type: 'error', content: '抱歉，服务暂时不可用，请稍后再试。' });
    } finally {
      isChatting.value = false;
      scrollToBottom();
    }
  };

  const clearChatHistory = async () => {
    if (!confirm('确定要清除对话历史吗？')) return;
    try {
      await deps.request(`${adminChatBasePath}/refresh`, 'POST');
      chatMessages.value = [{ type: 'ai', content: '对话历史已清除。' }];
      chatInput.value = '';
      chatImages.value = [];
      chatHistoryLoaded.value = true;
      deps.showToast('对话历史已重置');
    } catch {
      deps.showToast('清除对话历史失败', 'error');
    }
  };

  const exportData = (type) => {
    if (!deps.selectedTask.value) return;
    const task = deps.selectedTask.value;
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
      if (task.report) content += `## 综合分析报告\n${task.report}\n\n`;
      if (task.payload) {
        if (task.payload.video_insights?.length) {
          content += '## 视频分析洞察\n';
          task.payload.video_insights.forEach((insight, idx) => { content += `### 视频 #${idx + 1}\n${insight}\n\n`; });
        }
        if (task.payload.audio_insights?.length) {
          content += '## 音频分析洞察\n';
          task.payload.audio_insights.forEach((insight, idx) => { content += `### 音频 #${idx + 1}\n${insight}\n\n`; });
        }
        if (task.payload.image_insights?.length) {
          content += '## 图片分析洞察\n';
          task.payload.image_insights.forEach((insight, idx) => { content += `### 图片 #${idx + 1}\n${insight}\n\n`; });
        }
        if (task.payload.text) content += `## 原始文本证据\n${task.payload.text}\n\n`;
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

  const printReport = () => window.print();

  return {
    showChat,
    chatMessages,
    chatInput,
    chatImages,
    isChatting,
    chatHistoryLoaded,
    fetchChatHistory,
    toggleChat,
    parseReport,
    extractAttackSteps,
    extractScamKeywordSentences,
    parseRiskSummary,
    parseInsight,
    triggerChatImagePicker,
    handleChatImageSelect,
    removeChatImage,
    sendChatMessage,
    clearChatHistory,
    exportData,
    printReport
  };
}
