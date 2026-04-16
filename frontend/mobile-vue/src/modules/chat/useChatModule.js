import { nextTick, ref } from 'vue';

export function useChatModule(deps) {
  const buildChatMessage = (message) => {
    const normalizedType = String(message?.type || 'ai').trim() || 'ai';
    const normalizedContent = typeof message?.content === 'string' ? message.content : '';
    const normalizedImages = Array.isArray(message?.images)
      ? message.images.filter((item) => typeof item === 'string' && item.trim())
      : [];

    return {
      ...message,
      type: normalizedType,
      content: normalizedContent,
      images: normalizedImages,
      rendered_content: normalizedType === 'ai' ? deps.renderMarkdown(normalizedContent) : ''
    };
  };

  const chatMessages = ref([
    buildChatMessage({
      type: 'ai',
      content: '你好！我是你的反诈骗智能助手。我可以帮你分析风险、解答疑问，或者总结最近的安全情况。'
    })
  ]);
  const chatInput = ref('');
  const chatImages = ref([]);
  const isChatting = ref(false);
  const chatHistoryLoaded = ref(false);

  const appendChatMessage = (message) => {
    chatMessages.value.push(buildChatMessage(message));
    return chatMessages.value.length - 1;
  };

  const replaceChatMessage = (index, messagePatch) => {
    if (!Number.isInteger(index) || index < 0 || index >= chatMessages.value.length) {
      return -1;
    }

    const currentMessage = chatMessages.value[index] || {};
    const nextMessage = buildChatMessage({
      ...currentMessage,
      ...messagePatch
    });
    chatMessages.value.splice(index, 1, nextMessage);
    return index;
  };

  const scrollToBottom = () => {
    nextTick(() => {
      const container = document.getElementById('chat-container');
      if (!container) return;
      window.requestAnimationFrame(() => {
        container.scrollTop = container.scrollHeight;
      });
    });
  };

  const fetchChatHistory = async () => {
    if (!deps.isAuthenticated.value) return;

    try {
      const data = await deps.request('/chat/context');
      if (data && Array.isArray(data.messages)) {
        const history = [];
        for (const msg of data.messages) {
          if (msg.role === 'assistant') {
            if (Array.isArray(msg.tool_calls)) {
              for (const call of msg.tool_calls) {
                const toolName = call.name || call.function?.name || 'unknown';
                history.push(buildChatMessage({
                  type: 'tool',
                  content: `正在调用工具: ${toolName}...`
                }));
              }
            }

            if (msg.content) {
              history.push(buildChatMessage({
                type: 'ai',
                content: msg.content
              }));
            }
          } else if (msg.role === 'tool') {
            history.push(buildChatMessage({
              type: 'tool',
              content: '工具调用完成'
            }));
          } else if (msg.role === 'user') {
            const imageUrls = Array.isArray(msg.image_urls)
              ? msg.image_urls.filter((item) => typeof item === 'string' && item.trim())
              : [];
            if (!msg.content && imageUrls.length === 0) continue;
            history.push(buildChatMessage({
              type: 'user',
              content: msg.content || '',
              images: imageUrls
            }));
          }
        }

        if (history.length > 0) {
          chatMessages.value = [
            chatMessages.value[0],
            ...history
          ];
        }
      }

      chatHistoryLoaded.value = true;
      setTimeout(scrollToBottom, 100);
    } catch (error) {
      console.error('Fetch chat history failed:', error);
    }
  };

  const triggerChatImagePicker = () => {
    if (isChatting.value) return;
    const input = document.getElementById('chat-image-input');
    if (input) input.click();
  };

  const handleChatImageSelect = async (event) => {
    const files = Array.from(event.target.files || []).filter((file) => String(file.type || '').startsWith('image/'));
    event.target.value = '';
    if (files.length === 0) return;

    try {
      const results = await Promise.all(files.map((file) => deps.fileToBase64(file)));
      chatImages.value = [...chatImages.value, ...results];
      deps.showToast(`已添加 ${files.length} 张图片`);
    } catch (error) {
      console.error('Read chat images failed:', error);
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

    appendChatMessage({ type: 'user', content: message, images });
    chatInput.value = '';
    chatImages.value = [];
    isChatting.value = true;
    scrollToBottom();

    try {
      const response = await fetch('/api/chat', {
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
      appendChatMessage({ type: 'ai', content: '' });
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

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
              return appendChatMessage({ type: 'ai', content: '' });
            };

            if (data.type === 'content') {
              const idx = getActiveAiIndex();
              aiMessageContent += data.content;
              replaceChatMessage(idx, { type: 'ai', content: aiMessageContent });
              scrollToBottom();
            } else if (data.type === 'tool_call') {
              const toolName = data.tool;
              const lastIdx = chatMessages.value.length - 1;
              if (lastIdx >= 0 && chatMessages.value[lastIdx].type === 'ai' && !chatMessages.value[lastIdx].content) {
                replaceChatMessage(lastIdx, {
                  type: 'tool',
                  content: `正在调用工具: ${toolName}...`
                });
              } else {
                appendChatMessage({
                  type: 'tool',
                  content: `正在调用工具: ${toolName}...`
                });
              }
              aiMessageContent = '';
              scrollToBottom();
            } else if (data.type === 'tool_result') {
              appendChatMessage({
                type: 'tool',
                content: `工具 ${data.tool} 调用完成`
              });
              scrollToBottom();
            }
          } catch (error) {
            console.error('Error parsing SSE data:', error, 'Line:', trimmedLine);
          }
        }
      }
    } catch (error) {
      console.error('Chat error:', error);
      appendChatMessage({ type: 'error', content: '抱歉，服务暂时不可用，请稍后再试。' });
    } finally {
      isChatting.value = false;
      scrollToBottom();
    }
  };

  const clearChatHistory = async () => {
    if (!confirm('确定要清除对话历史吗？')) return;

    try {
      await deps.request('/chat/refresh', 'POST');
      chatMessages.value = [
        buildChatMessage({ type: 'ai', content: '对话历史已清除。' })
      ];
      chatInput.value = '';
      chatImages.value = [];
      chatHistoryLoaded.value = true;
      deps.showToast('对话历史已重置');
    } catch {
      deps.showToast('清除对话历史失败', 'error');
    }
  };

  const resetChatState = () => {
    chatMessages.value = [
      buildChatMessage({
        type: 'ai',
        content: '你好！我是你的反诈骗智能助手。我可以帮你分析风险、解答疑问，或者总结最近的安全情况。'
      })
    ];
    chatInput.value = '';
    chatImages.value = [];
    chatHistoryLoaded.value = false;
  };

  return {
    chatMessages,
    chatInput,
    chatImages,
    isChatting,
    chatHistoryLoaded,
    scrollToBottom,
    fetchChatHistory,
    triggerChatImagePicker,
    handleChatImageSelect,
    removeChatImage,
    sendChatMessage,
    clearChatHistory,
    resetChatState
  };
}
