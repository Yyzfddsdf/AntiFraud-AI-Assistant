const { createApp, ref, reactive, onMounted, onUnmounted, computed, watch } = Vue;

createApp({
    setup() {
        // State
        const isAuthenticated = ref(false);
        const token = ref(localStorage.getItem('token') || '');
        const user = ref({});
        const authMode = ref('login'); // login | register
        const activeTab = ref('tasks');
        const loading = ref(false);
        const analyzing = ref(false);
        const inviteCode = ref('');
        const captchaImage = ref('');
        const captchaId = ref('');
        const toasts = ref([]);
        const tasks = ref([]);
        const history = ref([]);
        const users = ref([]);
        const deletingHistory = reactive({});
        const selectedTask = ref(null);
        const userSearch = ref('');
        
        // Draggable Chat State
        const chatPosition = reactive({ left: 0, top: 0 });
        const isDragging = ref(false);
        const hasMoved = ref(false);

        // Sidebar State
        const isSidebarCollapsed = ref(false);
        const toggleSidebar = () => isSidebarCollapsed.value = !isSidebarCollapsed.value;

        const form = reactive({
            username: '',
            email: '',
            password: '',
            captchaCode: ''
        });

        const ageForm = reactive({ age: 28 });

        const analyzeForm = reactive({
            text: '',
            videos: [],
            audios: [],
            images: []
        });

        // Helpers
        const showToast = (message, type = 'success') => {
            const id = Date.now();
            toasts.value.push({ id, message, type });
            setTimeout(() => toasts.value = toasts.value.filter(t => t.id !== id), 3000);
        };

        const request = async (path, method = 'GET', body = null) => {
            const headers = { 'Accept': 'application/json', 'Content-Type': 'application/json' };
            if (token.value) headers['Authorization'] = `Bearer ${token.value}`;
            
            try {
                const res = await fetch(`/api${path}`, {
                    method,
                    headers,
                    body: body ? JSON.stringify(body) : undefined
                });
                
                if (res.status === 401 && isAuthenticated.value) {
                    logout();
                    return null;
                }
                
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || 'Request failed');
                return data;
            } catch (e) {
                showToast(e.message, 'error');
                return null;
            }
        };

        // Actions
        const fetchCaptcha = async () => {
            try {
                const res = await fetch('/api/auth/captcha');
                const data = await res.json();
                captchaId.value = data.captchaId;
                captchaImage.value = data.captchaImage;
            } catch (e) {
                showToast('验证码获取失败', 'error');
            }
        };

        const handleAuth = async () => {
            loading.value = true;
            const endpoint = authMode.value === 'login' ? '/auth/login' : '/auth/register';
            const payload = { ...form, captchaId: captchaId.value };
            
            const res = await request(endpoint, 'POST', payload);
            loading.value = false;
            
            if (res) {
                if (authMode.value === 'register') {
                    showToast('注册成功，请登录');
                    authMode.value = 'login';
                    fetchCaptcha();
                } else {
                    token.value = res.token;
                    localStorage.setItem('token', res.token);
                    isAuthenticated.value = true;
                    user.value = res.user;
                    if (res.user.age) ageForm.age = res.user.age;
                    else ageForm.age = 28;
                    showToast('登录成功');
                    startPolling();
                }
            } else {
                fetchCaptcha(); // Refresh captcha on fail
            }
        };

        const getUserInfo = async () => {
            const res = await request('/user');
            if (res) {
                user.value = res;
                if (res.age) ageForm.age = res.age;
                isAuthenticated.value = true;
                startPolling();
            } else {
                isAuthenticated.value = false;
            }
        };

        const logout = () => {
            token.value = '';
            localStorage.removeItem('token');
            isAuthenticated.value = false;
            user.value = {};
            stopPolling();
        };

        const updateAge = async () => {
            const res = await request('/scam/multimodal/user/age', 'PUT', { age: ageForm.age });
            if (res) {
                user.value.age = ageForm.age;
                showToast('年龄更新成功');
            }
        };

        const upgradeAccount = async () => {
            if (!inviteCode.value) return;
            loading.value = true;
            const res = await request('/upgrade', 'POST', { invite_code: inviteCode.value });
            loading.value = false;
            if (res) {
                showToast(res.message);
                user.value = res.user;
                inviteCode.value = '';
            }
        };

        const deleteAccount = async () => {
            if(!confirm('确定要删除账户吗？此操作不可逆！')) return;
            const res = await request('/user', 'DELETE');
            if (res) {
                showToast('账户已删除');
                logout();
            }
        };

        // File Handling
        const fileToBase64 = (file) => new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () => resolve(reader.result);
            reader.onerror = error => reject(error);
        });

        const handleFileSelect = async (event, type) => {
            const files = Array.from(event.target.files);
            if (files.length === 0) return;
            
            const base64Promises = files.map(file => fileToBase64(file));
            try {
                const results = await Promise.all(base64Promises);
                analyzeForm[type] = [...analyzeForm[type], ...results];
                showToast(`已添加 ${files.length} 个文件`);
            } catch (e) {
                showToast('文件读取失败', 'error');
            }
        };

        const submitAnalysis = async () => {
            if (!analyzeForm.text && analyzeForm.videos.length === 0 && analyzeForm.audios.length === 0 && analyzeForm.images.length === 0) {
                showToast('请至少提供一种输入（文本或文件）', 'error');
                return;
            }

            analyzing.value = true;
            const res = await request('/scam/multimodal/analyze', 'POST', analyzeForm);
            analyzing.value = false;

            if (res) {
                showToast('任务已提交');
                // Reset form
                analyzeForm.text = '';
                analyzeForm.videos = [];
                analyzeForm.audios = [];
                analyzeForm.images = [];
                activeTab.value = 'tasks';
                fetchTasks();
            }
        };

        const fetchTasks = async () => {
            if (!isAuthenticated.value) return;
            const res = await request('/scam/multimodal/tasks');
            if (res && res.tasks) {
                tasks.value = res.tasks.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
            }
        };

        const fetchHistory = async () => {
            const res = await request('/scam/multimodal/history');
            if (res && res.history) {
                history.value = res.history;
            }
        };

        const viewTaskDetail = async (taskId) => {
            const res = await request(`/scam/multimodal/tasks/${taskId}`);
            if (res && res.task) {
                selectedTask.value = res.task;
            }
        };

        const viewHistoryDetail = (item) => {
            viewTaskDetail(item.record_id);
        };

        const deleteHistoryCase = async (item) => {
            if (!item || !item.record_id) return;
            if (!confirm(`确定删除案件 ${item.record_id} 吗？此操作不可恢复。`)) return;

            deletingHistory[item.record_id] = true;
            try {
                const res = await request(`/scam/multimodal/history/${encodeURIComponent(item.record_id)}`, 'DELETE');
                if (!res) return;

                history.value = history.value.filter(h => h.record_id !== item.record_id);
                if (selectedTask.value && selectedTask.value.task_id === item.record_id) {
                    selectedTask.value = null;
                }
                showToast(res.message || '历史案件删除成功');
            } finally {
                deletingHistory[item.record_id] = false;
            }
        };

        // User Management
        const fetchUsers = async () => {
            if (!isAuthenticated.value) return;
            const query = userSearch.value ? `?query=${encodeURIComponent(userSearch.value)}` : '';
            const res = await request(`/users${query}`);
            if (res && res.users) {
                users.value = res.users;
            }
        };

        let debounceTimer;
        const debouncedFetchUsers = () => {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(fetchUsers, 300);
        };

        // Polling
        let pollInterval;
        const startPolling = () => {
            fetchTasks();
            fetchHistory();
            if (pollInterval) clearInterval(pollInterval);
            pollInterval = setInterval(() => {
                if (isAuthenticated.value && activeTab.value === 'tasks') fetchTasks();
            }, 5000);
        };
        
        const stopPolling = () => {
            if (pollInterval) clearInterval(pollInterval);
        };

        // Draggable Logic
        const initChatPosition = () => {
            // Initial position: bottom right
            chatPosition.left = window.innerWidth - 100;
            chatPosition.top = window.innerHeight - 100;
        };

        const handleResize = () => {
            const maxX = window.innerWidth - 60;
            const maxY = window.innerHeight - 60;
            if (chatPosition.left > maxX) chatPosition.left = maxX;
            if (chatPosition.top > maxY) chatPosition.top = maxY;
        };

        const startDrag = (e) => {
            if (e.button !== 0) return; // Only left mouse button
            
            isDragging.value = true;
            hasMoved.value = false;
            
            const startX = e.clientX;
            const startY = e.clientY;
            const initialLeft = chatPosition.left;
            const initialTop = chatPosition.top;
            
            const onMouseMove = (e) => {
                const dx = e.clientX - startX;
                const dy = e.clientY - startY;
                
                if (Math.abs(dx) > 5 || Math.abs(dy) > 5) {
                    hasMoved.value = true;
                }
                
                // Boundary check
                let newLeft = initialLeft + dx;
                let newTop = initialTop + dy;
                
                const maxX = window.innerWidth - 60;
                const maxY = window.innerHeight - 60;
                
                if (newLeft < 0) newLeft = 0;
                if (newLeft > maxX) newLeft = maxX;
                if (newTop < 0) newTop = 0;
                if (newTop > maxY) newTop = maxY;

                chatPosition.left = newLeft;
                chatPosition.top = newTop;
            };
            
            const onMouseUp = () => {
                window.removeEventListener('mousemove', onMouseMove);
                window.removeEventListener('mouseup', onMouseUp);
                isDragging.value = false;
            };
            
            window.addEventListener('mousemove', onMouseMove);
            window.addEventListener('mouseup', onMouseUp);
        };

        // Watchers
        watch(activeTab, (newTab) => {
            if (newTab === 'tasks') fetchTasks();
            if (newTab === 'history') fetchHistory();
            if (newTab === 'users') fetchUsers();
        });

        // Init
        onMounted(() => {
            fetchCaptcha();
            if (token.value) {
                getUserInfo();
            }
            initChatPosition();
            window.addEventListener('resize', handleResize);
        });

        onUnmounted(() => {
            stopPolling();
            window.removeEventListener('resize', handleResize);
        });

        // Formatting
        const formatTime = (iso) => {
            return new Date(iso).toLocaleString('zh-CN', { hour12: false });
        };

        const getStatusLabel = (status) => {
            const map = { 'pending': '等待中', 'processing': '分析中', 'completed': '已完成', 'failed': '失败' };
            return map[status] || status;
        };

        const getStatusClass = (status) => {
            const map = {
                'pending': 'bg-yellow-100 text-yellow-800 px-2 py-1 rounded-full text-xs font-bold',
                'processing': 'bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-xs font-bold animate-pulse',
                'completed': 'bg-green-100 text-green-800 px-2 py-1 rounded-full text-xs font-bold',
                'failed': 'bg-red-100 text-red-800 px-2 py-1 rounded-full text-xs font-bold'
            };
            return map[status] || 'bg-gray-100 text-gray-800 px-2 py-1 rounded-full text-xs font-bold';
        };
        
        const normalizeRiskLevelText = (level) => {
            const value = String(level || '').trim();
            if (value === '高') return '高';
            if (value === '低') return '低';
            return '中';
        };

        const getRiskClass = (level) => {
            const normalized = normalizeRiskLevelText(level);
            if (normalized === '高') return 'bg-red-100 text-red-800';
            if (normalized === '中') return 'bg-yellow-100 text-yellow-800';
            return 'bg-green-100 text-green-800';
        };

        const openImage = (src) => {
            const win = window.open("", "_blank");
            win.document.write(`<img src="${src}" style="max-width:100%; height:auto;">`);
        };

        // Chat State
        const showChat = ref(false);
        const chatMessages = ref([
            { type: 'ai', content: '你好！我是你的反诈骗智能助手。我可以帮你分析风险，解答疑问，或者总结最近的安全状况。' }
        ]);
        const chatInput = ref('');
        const isChatting = ref(false);
        const chatHistoryLoaded = ref(false);

        // Chat Actions
        const fetchChatHistory = async () => {
            if (!isAuthenticated.value) return;
            
            try {
                const data = await request('/chat/context');
                if (data.messages && Array.isArray(data.messages)) {
                    const history = [];
                    
                    for (const msg of data.messages) {
                        // Handle assistant messages with tool calls
                        if (msg.role === 'assistant') {
                            // If there are tool calls, add them as tool status messages
                            if (msg.tool_calls && Array.isArray(msg.tool_calls)) {
                                for (const call of msg.tool_calls) {
                                    const toolName = call.function?.name || 'unknown';
                                    history.push({
                                        type: 'tool',
                                        content: `正在调用工具: ${toolName}...`
                                    });
                                }
                            }
                            
                            // If there is content, add as AI message
                            if (msg.content) {
                                history.push({
                                    type: 'ai',
                                    content: msg.content
                                });
                            }
                        } 
                        // Handle tool result messages
                        else if (msg.role === 'tool') {
                            // Optionally show tool completion
                            history.push({
                                type: 'tool',
                                content: `工具调用完成`
                            });
                        }
                        // Handle user messages
                        else if (msg.role === 'user') {
                            history.push({
                                type: 'user',
                                content: msg.content
                            });
                        }
                    }
                    
                    if (history.length > 0) {
                        // Keep welcome message and append history
                        chatMessages.value = [
                            chatMessages.value[0], 
                            ...history
                        ];
                    }
                }
                chatHistoryLoaded.value = true;
                setTimeout(scrollToBottom, 100);
            } catch (e) {
                console.error('Fetch chat history failed:', e);
            }
        };

        const toggleChat = () => {
            if (hasMoved.value) return; // Prevent toggle if dragged
            showChat.value = !showChat.value;
            // Auto scroll to bottom when opening
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
                    currentSection = {
                        id: parseInt(match[1]),
                        title: match[2].trim(),
                        content: ''
                    };
                } else if (currentSection) {
                    currentSection.content += line + '\n';
                }
            }
            if (currentSection) {
                currentSection.content = currentSection.content.trim();
                sections.push(currentSection);
            }
            return sections;
        };

        const parseInsight = (text) => {
            if (!text) return [];
            const sections = [];
            // Regex to match titles like 【整体视觉感受】 or [关键信息提取] or 图片分析 #1
            const regex = /^[【\[](.+?)[】\]]\s*(.*)$/; 
            const lines = text.split('\n');
            let currentSection = null;

            // Check if the text starts with a header that we might be missing
            // For example "图片分析 #1" which is not in brackets
            
            for (const line of lines) {
                const match = line.trim().match(regex);
                if (match) {
                    if (currentSection) {
                        currentSection.content = currentSection.content.trim();
                        sections.push(currentSection);
                    }
                    currentSection = {
                        title: match[1].trim(),
                        content: match[2] ? match[2] + '\n' : ''
                    };
                } else if (currentSection) {
                    currentSection.content += line + '\n';
                } else if (line.trim()) {
                    // Handle content before the first title (if any) as a general introduction
                    // If we already have "概述" section, append to it
                    if (sections.length > 0 && sections[0].title === '概述') {
                         sections[0].content += line + '\n';
                    } else if (!currentSection) {
                        currentSection = { title: '概述', content: line + '\n' };
                    }
                }
            }
            if (currentSection) {
                currentSection.content = currentSection.content.trim();
                sections.push(currentSection);
            }
            return sections;
        };

        const scrollToBottom = () => {
            const container = document.getElementById('chat-container');
            if (container) container.scrollTop = container.scrollHeight;
        };

        const sendChatMessage = async () => {
            if (!chatInput.value.trim() || isChatting.value) return;
            
            const message = chatInput.value.trim();
            chatMessages.value.push({ type: 'user', content: message });
            chatInput.value = '';
            isChatting.value = true;
            scrollToBottom();

            try {
                const response = await fetch('/api/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token.value}`
                    },
                    body: JSON.stringify({ message })
                });

                if (!response.ok) throw new Error('Network response was not ok');

                const reader = response.body.getReader();
                const decoder = new TextDecoder();
                let aiMessageContent = '';
                
                // Add placeholder for AI response
                chatMessages.value.push({ type: 'ai', content: '' });
                
                // We can't use a fixed index anymore because tool calls might insert messages
                // So we'll always target the last message if it's type 'ai', or append a new one
                
                let buffer = '';

                while (true) {
                    const { done, value } = await reader.read();
                    if (done) break;
                    
                    const chunk = decoder.decode(value, { stream: true });
                    console.log('Chunk received:', chunk);
                    buffer += chunk;
                    const lines = buffer.split('\n');
                    
                    // Keep the last potentially incomplete line in the buffer
                    buffer = lines.pop();

                    for (const line of lines) {
                        const trimmedLine = line.trim();
                        if (!trimmedLine) continue;
                        
                        if (trimmedLine.startsWith('data:')) {
                            try {
                                const jsonStr = trimmedLine.slice(5).trim();
                                const data = JSON.parse(jsonStr);
                                
                                // Helper to get active AI message index
                                const getActiveAiIndex = () => {
                                    const lastIdx = chatMessages.value.length - 1;
                                    if (lastIdx >= 0 && chatMessages.value[lastIdx].type === 'ai') {
                                        return lastIdx;
                                    }
                                    // If last message is not AI (e.g. tool), push new AI placeholder
                                    chatMessages.value.push({ type: 'ai', content: '' });
                                    return chatMessages.value.length - 1;
                                };

                                if (data.type === 'content') {
                                    const idx = getActiveAiIndex();
                                    aiMessageContent += data.content;
                                    chatMessages.value[idx].content = aiMessageContent;
                                    scrollToBottom();
                                } else if (data.type === 'tool_call') {
                                    // Insert tool message
                                    const toolName = data.tool;
                                    // If current AI message is empty, replace it
                                    const lastIdx = chatMessages.value.length - 1;
                                    if (lastIdx >= 0 && chatMessages.value[lastIdx].type === 'ai' && !chatMessages.value[lastIdx].content) {
                                        chatMessages.value[lastIdx] = {
                                            type: 'tool',
                                            content: `正在调用工具: ${toolName}...`
                                        };
                                    } else {
                                        chatMessages.value.push({
                                            type: 'tool',
                                            content: `正在调用工具: ${toolName}...`
                                        });
                                    }
                                    // Reset aiMessageContent for next segment
                                    aiMessageContent = '';
                                    scrollToBottom();
                                } else if (data.type === 'tool_result') {
                                    // Insert tool completion message
                                    const toolName = data.tool;
                                    chatMessages.value.push({
                                        type: 'tool',
                                        content: `工具 ${toolName} 调用完成`
                                    });
                                    scrollToBottom();
                                } else if (data.type === 'done') {
                                    // Stream finished
                                }
                            } catch (e) {
                                console.error('Error parsing SSE data:', e, 'Line:', trimmedLine);
                            }
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
                await request('/chat/refresh', 'POST');
                chatMessages.value = [
                    { type: 'ai', content: '对话历史已清除。' }
                ];
                chatHistoryLoaded.value = true;
                showToast('对话历史已重置');
            } catch (e) {
                showToast('清除历史失败', 'error');
            }
        };

        const exportData = (type) => {
            if (!selectedTask.value) return;
            
            const task = selectedTask.value;
            const date = new Date().toISOString().slice(0, 10);
            const filename = `scam-report-${task.task_id}-${date}`;

            if (type === 'json') {
                const dataStr = JSON.stringify(task, null, 2);
                const blob = new Blob([dataStr], { type: "application/json" });
                const url = URL.createObjectURL(blob);
                const link = document.createElement('a');
                link.href = url;
                link.download = `${filename}.json`;
                document.body.appendChild(link);
                link.click();
                document.body.removeChild(link);
                URL.revokeObjectURL(url);
            } else if (type === 'md') {
                let content = `# 诈骗风险分析报告\n\n`;
                content += `**任务ID**: ${task.task_id}\n`;
                content += `**标题**: ${task.title}\n`;
                content += `**生成时间**: ${new Date(task.created_at).toLocaleString()}\n`;
                content += `**状态**: ${task.status}\n\n`;
                
                if (task.report) {
                    content += `## 综合分析报告\n${task.report}\n\n`;
                }
                
                // Add Insights if available
                if (task.payload) {
                        if (task.payload.video_insights && task.payload.video_insights.length) {
                            content += `## 视频分析洞察\n`;
                            task.payload.video_insights.forEach((insight, idx) => {
                                content += `### 视频 #${idx + 1}\n${insight}\n\n`;
                            });
                        }
                        if (task.payload.audio_insights && task.payload.audio_insights.length) {
                            content += `## 音频分析洞察\n`;
                            task.payload.audio_insights.forEach((insight, idx) => {
                                content += `### 音频 #${idx + 1}\n${insight}\n\n`;
                            });
                        }
                        if (task.payload.image_insights && task.payload.image_insights.length) {
                            content += `## 图片分析洞察\n`;
                            task.payload.image_insights.forEach((insight, idx) => {
                                content += `### 图片 #${idx + 1}\n${insight}\n\n`;
                            });
                        }
                        
                        if (task.payload.text) {
                            content += `## 原始文本证据\n${task.payload.text}\n\n`;
                        }
                }
                
                const blob = new Blob([content], { type: "text/markdown" });
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
            isAuthenticated, user, authMode, form, ageForm, analyzeForm, 
            captchaImage, fetchCaptcha, handleAuth, logout, loading,
            activeTab, tasks, history, users, selectedTask, userSearch, toasts, analyzing,
            deletingHistory, handleFileSelect, submitAnalysis, viewTaskDetail, viewHistoryDetail, deleteHistoryCase, debouncedFetchUsers,
            formatTime, getStatusLabel, getStatusClass, normalizeRiskLevelText, getRiskClass,
            updateAge, deleteAccount, upgradeAccount, inviteCode, openImage, exportData, printReport,
            showChat, chatMessages, chatInput, isChatting, toggleChat, sendChatMessage, clearChatHistory,
            chatPosition, startDrag, // Export drag handler and state
            isSidebarCollapsed, toggleSidebar,
            parseReport, parseInsight
        };
    }
}).mount('#app');
