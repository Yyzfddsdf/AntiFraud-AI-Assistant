const { createApp, ref, reactive, onMounted, onUnmounted, computed, watch } = Vue;

createApp({
    setup() {
        // State
        const isAuthenticated = ref(false);
        const token = ref(localStorage.getItem('token') || '');
        const user = ref({});
        const authMode = ref('login'); // login | register
        const loginMethod = ref('password'); // password | sms
        const activeTab = ref('tasks');
        const loading = ref(false);
        const analyzing = ref(false);
        const inviteCode = ref('');
        const captchaImage = ref('');
        const captchaId = ref('');
        const toasts = ref([]);
        const alertEvents = ref([]);
        const alertUnreadCount = ref(0);
        const alertModalVisible = ref(false);
        const activeAlertEvent = ref(null);
        const alertConnectionStatus = ref('disconnected'); // disconnected | connecting | connected | reconnecting
        const alertDrawerVisible = ref(false);
        const tasks = ref([]);
        const history = ref([]);
        const users = ref([]);
        const familyOverview = ref(null);
        const familyMembers = ref([]);
        const familyInvitations = ref([]);
        const familyReceivedInvitations = ref([]);
        const familyGuardianLinks = ref([]);
        const familyNotifications = ref([]);
        const familyLoading = ref(false);
        const familyReceivedLoading = ref(false);
        const familyNotificationConnectionStatus = ref('disconnected'); // disconnected | connecting | connected | reconnecting
        const familyAlertModalVisible = ref(false);
        const activeFamilyNotification = ref(null);
        const caseLibrary = ref([]); // Admin Case Library
        const pendingReviews = ref([]); // Admin Pending Reviews
        const startingCaseCollection = ref(false);
        const selectedReview = ref(null);
        const showReviewDetailModal = ref(false);
        const scamTypeOptions = ref([]);
        const targetGroupOptions = ref([]);
        const selectedCase = ref(null); // For admin view details
        const showCaseModal = ref(false);
        const submittingCase = ref(false);
        const caseForm = reactive({
            title: '',
            target_group: '',
            risk_level: '',
            scam_type: '',
            case_description: '',
            typical_scripts_raw: '',
            keywords_raw: '',
            violated_law: '',
            suggestion: ''
        });
        const caseCollectionForm = reactive({
            query: '',
            case_count: 5
        });
        const deletingHistory = reactive({});
        const familyDeletingMembers = reactive({});
        const familyDeletingGuardianLinks = reactive({});
        const familyAcceptingInvitations = reactive({});
        const familyMarkingNotifications = reactive({});
        const selectedTask = ref(null);
        const userSearch = ref('');
        
        // Risk Trend State
        const riskInterval = ref('day');
        const riskData = ref(null);
        // Cache for risk trend data: { 'day': data, 'week': data, 'month': data }
        const riskCache = reactive({});
        let pieChartInstance = null;
        let lineChartInstance = null;
        let hasWarnedMissingChartLibrary = false;

        // Admin Stats State
        const adminStatsInterval = ref('day');
        const adminStatsData = ref(null);
        // Cache for admin stats: { 'day': data, 'week': data, 'month': data }
        const adminStatsCache = reactive({});
        const adminGraphData = ref(null);
        const adminGraphCache = ref(null);
        const adminTargetGroupChartData = ref(null);
        const adminTargetGroupChartCache = reactive({});
        const selectedGraphTargetGroup = ref('');
        const showGraphModal = ref(false);
        let adminTrendChart = null;
        let adminTypeChart = null;
        let adminTargetChart = null;
        let adminTargetGroupBarChart = null;
        let adminNetworkInstance = null;

        // Draggable Chat State
        const chatPosition = reactive({ left: 0, top: 0 });
        const isDragging = ref(false);
        const hasMoved = ref(false);

        // Sidebar State
        const isSidebarCollapsed = ref(false);
        const toggleSidebar = () => isSidebarCollapsed.value = !isSidebarCollapsed.value;

        const form = reactive({
            account: '',
            username: '',
            email: '',
            phone: '',
            password: '',
            captchaCode: '',
            smsCode: ''
        });

        const ageForm = reactive({ age: 28 });
        const profileForm = reactive({
            occupation: '',
            recentTagsText: ''
        });
        const occupationOptions = ref([]);
        const profileSaving = ref(false);
        const demoSMSCode = '000000';
        const smsCodeCooldown = ref(0);
        const smsCodeSending = ref(false);
        let smsCodeCooldownTimer = null;
        const familyCreateForm = reactive({ name: '' });
        const familyInviteForm = reactive({
            invitee_email: '',
            invitee_phone: '',
            role: 'member',
            relation: ''
        });
        const familyAcceptForm = reactive({ invite_code: '' });
        const familyGuardianForm = reactive({
            guardian_user_id: '',
            member_user_id: ''
        });

        const analyzeForm = reactive({
            text: '',
            videos: [],
            audios: [],
            images: []
        });

        // Alert WebSocket Runtime
        let alertSocket = null;
        let alertReconnectTimer = null;
        let alertReconnectAttempts = 0;
        const alertSeenRecordIDs = new Set();
        const maxAlertReconnectDelayMS = 30000;
        let familyNotificationSocket = null;
        let familyNotificationReconnectTimer = null;
        let familyNotificationReconnectAttempts = 0;
        const familyNotificationSeenIDs = new Set();
        const maxFamilyNotificationReconnectDelayMS = 30000;

        // Helpers
        const showToast = (message, type = 'success') => {
            const id = Date.now();
            toasts.value.push({ id, message, type });
            setTimeout(() => toasts.value = toasts.value.filter(t => t.id !== id), 3000);
        };

        const parseRecentTagsInput = (raw) => {
            return String(raw || '')
                .split(/\r?\n|,|，|;|；/)
                .map(item => item.trim())
                .filter(Boolean);
        };

        const syncProfileForm = (profile) => {
            const normalizedAge = Number(profile && profile.age);
            ageForm.age = Number.isFinite(normalizedAge) && normalizedAge > 0 ? normalizedAge : 28;
            profileForm.occupation = String(profile && profile.occupation ? profile.occupation : '').trim();
            profileForm.recentTagsText = Array.isArray(profile && profile.recent_tags)
                ? profile.recent_tags.filter(item => typeof item === 'string' && item.trim()).join('\n')
                : '';
        };

        const fetchOccupationOptions = async () => {
            const res = await request('/user/profile/options/occupations', 'GET', null, { silent: true });
            if (res && Array.isArray(res.occupations)) {
                occupationOptions.value = res.occupations.filter(item => typeof item === 'string' && item.trim());
            }
        };

        const alertConnectionLabel = computed(() => {
            switch (alertConnectionStatus.value) {
                case 'connected':
                    return '风险预警通道已连接';
                case 'connecting':
                    return '风险预警通道连接中';
                case 'reconnecting':
                    return '风险预警通道重连中';
                default:
                    return '风险预警通道未连接';
            }
        });
        const familyNotificationConnectionLabel = computed(() => {
            switch (familyNotificationConnectionStatus.value) {
                case 'connected':
                    return '家庭通知通道已连接';
                case 'connecting':
                    return '家庭通知连接中';
                case 'reconnecting':
                    return '家庭通知重连中';
                default:
                    return '家庭通知未连接';
            }
        });

        const normalizeAlertRiskLevel = (level) => {
            const value = String(level || '').trim();
            if (value === '高') return '高';
            if (value === '中') return '中';
            if (value === '低') return '低';
            return '';
        };

        const isPersonalAlertRiskLevel = (level) => {
            const normalized = normalizeAlertRiskLevel(level);
            return normalized === '高' || normalized === '中';
        };

        const getAlertSeverityTheme = (level) => {
            const normalized = normalizeAlertRiskLevel(level) || '中';
            if (normalized === '高') {
                return {
                    hoverClass: 'hover:border-red-200 hover:bg-red-50/40',
                    pillClass: 'text-red-600 bg-red-50 border-red-100',
                    unreadClass: 'bg-red-500',
                    modalBorderClass: 'border-red-100',
                    modalHeaderClass: 'from-red-600 to-rose-600',
                    modalPanelClass: 'border-red-100 bg-red-50/70',
                    actionClass: 'bg-red-600 hover:bg-red-700 shadow-red-600/20',
                };
            }
            return {
                hoverClass: 'hover:border-amber-200 hover:bg-amber-50/50',
                pillClass: 'text-amber-700 bg-amber-50 border-amber-100',
                unreadClass: 'bg-amber-500',
                modalBorderClass: 'border-amber-100',
                modalHeaderClass: 'from-amber-500 to-orange-500',
                modalPanelClass: 'border-amber-100 bg-amber-50/80',
                actionClass: 'bg-amber-500 hover:bg-amber-600 shadow-amber-500/20',
            };
        };

        const getAlertToastType = (level) => normalizeAlertRiskLevel(level) === '高' ? 'error' : 'warning';

        const recentRiskAlerts = computed(() => {
            const recentAlertWindowMS = 60 * 60 * 1000;
            const nowMS = Date.now();
            const isWithinRecentAlertWindow = (rawTime) => {
                const parsedTime = new Date(rawTime || '').getTime();
                if (!Number.isFinite(parsedTime) || parsedTime <= 0) return false;
                return parsedTime >= nowMS - recentAlertWindowMS && parsedTime <= nowMS;
            };

            const unreadRecordIDs = new Set(
                (alertEvents.value || [])
                    .filter((item) => item && !item.read)
                    .map((item) => String(item.record_id || '').trim())
                    .filter((item) => item !== '')
            );

            const merged = new Map();
            const historyItems = Array.isArray(history.value) ? history.value : [];
            for (const item of historyItems) {
                if (!item) continue;
                const recordID = String(item.record_id || '').trim();
                if (!recordID) continue;
                const riskLevel = normalizeAlertRiskLevel(item.risk_level);
                if (!isPersonalAlertRiskLevel(riskLevel)) continue;
                if (!isWithinRecentAlertWindow(item.created_at)) continue;

                merged.set(recordID, {
                    record_id: recordID,
                    title: String(item.title || '').trim() || '风险预警',
                    case_summary: String(item.case_summary || '').trim() || '暂无摘要',
                    scam_type: String(item.scam_type || '').trim() || '未知类型',
                    risk_level: riskLevel,
                    created_at: String(item.created_at || '').trim(),
                    sent_at: '',
                    unread: unreadRecordIDs.has(recordID)
                });
            }

            const alertItems = Array.isArray(alertEvents.value) ? alertEvents.value : [];
            for (const event of alertItems) {
                if (!event) continue;
                const recordID = String(event.record_id || '').trim();
                if (!recordID) continue;
                if (!isWithinRecentAlertWindow(event.created_at)) continue;
                if (!merged.has(recordID)) {
                    merged.set(recordID, {
                        record_id: recordID,
                        title: String(event.title || '').trim() || '风险预警',
                        case_summary: String(event.case_summary || '').trim() || '暂无摘要',
                        scam_type: String(event.scam_type || '').trim() || '未知类型',
                        risk_level: normalizeAlertRiskLevel(event.risk_level) || '中',
                        created_at: String(event.created_at || '').trim(),
                        sent_at: String(event.sent_at || '').trim(),
                        unread: !event.read
                    });
                } else if (!event.read) {
                    const existing = merged.get(recordID);
                    if (existing) existing.unread = true;
                }
            }

            return Array.from(merged.values())
                .sort((a, b) => {
                    const timeA = new Date(a.created_at || a.sent_at || 0).getTime();
                    const timeB = new Date(b.created_at || b.sent_at || 0).getTime();
                    return timeB - timeA;
                })
                .slice(0, 20);
        });

        const markAlertReadByRecordID = (recordID) => {
            const targetID = String(recordID || '').trim();
            if (!targetID) return;

            let reduced = 0;
            for (const event of alertEvents.value) {
                if (!event) continue;
                if (String(event.record_id || '').trim() !== targetID) continue;
                if (!event.read) {
                    event.read = true;
                    reduced += 1;
                }
            }
            if (reduced > 0) {
                alertUnreadCount.value = Math.max(0, alertUnreadCount.value - reduced);
            }
            if (activeAlertEvent.value && String(activeAlertEvent.value.record_id || '').trim() === targetID) {
                activeAlertEvent.value.read = true;
            }
        };

        const closeAlertDrawer = () => {
            alertDrawerVisible.value = false;
        };

        const toggleAlertDrawer = async () => {
            alertDrawerVisible.value = !alertDrawerVisible.value;
            if (alertDrawerVisible.value) {
                await fetchHistory({ silent: true });
            }
        };

        const openAlertCaseDetail = async (item) => {
            if (!item || !item.record_id) return;
            markAlertReadByRecordID(item.record_id);
            alertDrawerVisible.value = false;
            alertModalVisible.value = false;
            activeTab.value = 'history';
            await fetchHistory({ silent: true });
            await viewTaskDetail(item.record_id);
        };

        const buildAlertWebSocketURL = () => {
            const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
            const base = `${protocol}://${window.location.host}/api/alert/ws`;
            const queryToken = encodeURIComponent(token.value || '');
            return `${base}?token=${queryToken}`;
        };

        const buildFamilyNotificationWebSocketURL = () => {
            const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
            const base = `${protocol}://${window.location.host}/api/families/notifications/ws`;
            const queryToken = encodeURIComponent(token.value || '');
            return `${base}?token=${queryToken}`;
        };

        const disconnectAlertWebSocket = () => {
            if (alertReconnectTimer) {
                clearTimeout(alertReconnectTimer);
                alertReconnectTimer = null;
            }
            alertReconnectAttempts = 0;

            if (alertSocket) {
                const ws = alertSocket;
                alertSocket = null;
                ws.onopen = null;
                ws.onmessage = null;
                ws.onerror = null;
                ws.onclose = null;
                if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
                    ws.close(1000, 'client logout');
                }
            }
            alertConnectionStatus.value = 'disconnected';
        };

        const disconnectFamilyNotificationWebSocket = () => {
            if (familyNotificationReconnectTimer) {
                clearTimeout(familyNotificationReconnectTimer);
                familyNotificationReconnectTimer = null;
            }
            familyNotificationReconnectAttempts = 0;

            if (familyNotificationSocket) {
                const ws = familyNotificationSocket;
                familyNotificationSocket = null;
                ws.onopen = null;
                ws.onmessage = null;
                ws.onerror = null;
                ws.onclose = null;
                if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
                    ws.close(1000, 'client logout');
                }
            }
            familyNotificationConnectionStatus.value = 'disconnected';
        };

        const scheduleAlertReconnect = () => {
            if (!isAuthenticated.value || !token.value) return;
            if (alertReconnectTimer) return;

            const delay = Math.min(maxAlertReconnectDelayMS, 1000 * Math.pow(2, alertReconnectAttempts));
            alertReconnectAttempts += 1;
            alertReconnectTimer = setTimeout(() => {
                alertReconnectTimer = null;
                connectAlertWebSocket();
            }, delay);
        };

        const scheduleFamilyNotificationReconnect = () => {
            if (!isAuthenticated.value || !token.value) return;
            if (familyNotificationReconnectTimer) return;

            const delay = Math.min(maxFamilyNotificationReconnectDelayMS, 1000 * Math.pow(2, familyNotificationReconnectAttempts));
            familyNotificationReconnectAttempts += 1;
            familyNotificationReconnectTimer = setTimeout(() => {
                familyNotificationReconnectTimer = null;
                connectFamilyNotificationWebSocket();
            }, delay);
        };

        const acknowledgeActiveAlert = () => {
            if (activeAlertEvent.value) {
                markAlertReadByRecordID(activeAlertEvent.value.record_id);
            }
            alertModalVisible.value = false;
        };

        const openAlertHistory = async () => {
            const current = activeAlertEvent.value;
            if (!current) return;
            await openAlertCaseDetail(current);
        };

        const acknowledgeFamilyAlert = () => {
            const current = activeFamilyNotification.value;
            if (!current) {
                familyAlertModalVisible.value = false;
                return;
            }
            markFamilyNotificationRead(current);
            familyAlertModalVisible.value = false;
        };

        const openFamilyNotificationCenter = async () => {
            const current = activeFamilyNotification.value;
            familyAlertModalVisible.value = false;
            activeTab.value = 'family';
            if (current) {
                await markFamilyNotificationRead(current);
            }
            await fetchFamilyOverview({ silent: true });
            connectFamilyNotificationWebSocket();
        };

        const handleAlertMessage = (payload) => {
            if (!payload || !['risk_alert', 'high_risk_alert'].includes(String(payload.type || '').trim())) return;
            const recordID = String(payload.record_id || '').trim();
            if (!recordID) return;
            if (alertSeenRecordIDs.has(recordID)) return;
            alertSeenRecordIDs.add(recordID);
            const riskLevel = normalizeAlertRiskLevel(payload.risk_level) || '高';

            const event = {
                id: `${recordID}-${Date.now()}`,
                record_id: recordID,
                title: String(payload.title || '').trim() || '风险预警',
                case_summary: String(payload.case_summary || '').trim() || `${riskLevel}风险事件已触发预警，请及时核查。`,
                scam_type: String(payload.scam_type || '').trim() || '未知类型',
                risk_level: riskLevel,
                created_at: String(payload.created_at || '').trim(),
                sent_at: String(payload.sent_at || '').trim(),
                read: false
            };

            alertEvents.value = [event, ...alertEvents.value].slice(0, 30);
            alertUnreadCount.value += 1;
            activeAlertEvent.value = event;
            alertModalVisible.value = true;
            showToast(`${event.risk_level}风险预警：${event.title}`, getAlertToastType(event.risk_level));
            fetchHistory({ silent: true });
        };

        const handleFamilyNotificationMessage = (payload) => {
            if (!payload || payload.type !== 'family_high_risk_alert') return;
            const notificationID = Number(payload.notification_id || 0);
            if (!Number.isInteger(notificationID) || notificationID <= 0) return;
            if (familyNotificationSeenIDs.has(notificationID)) return;
            familyNotificationSeenIDs.add(notificationID);

            const notification = {
                id: notificationID,
                family_id: Number(payload.family_id || 0),
                target_user_id: Number(payload.target_user_id || 0),
                target_name: String(payload.target_name || '').trim() || '家庭成员',
                event_type: String(payload.event_type || '').trim() || 'high_risk_case',
                record_id: String(payload.record_id || '').trim(),
                title: String(payload.title || '').trim() || '家庭高风险通知',
                summary: String(payload.summary || '').trim() || '家庭成员触发高风险案件，请及时核查。',
                risk_level: String(payload.risk_level || '').trim() || '高',
                event_at: String(payload.event_at || '').trim(),
                read_at: String(payload.read_at || '').trim()
            };
            familyNotifications.value = [notification, ...familyNotifications.value].slice(0, 50);
            activeFamilyNotification.value = notification;
            familyAlertModalVisible.value = true;
            showToast(`家庭通知：${notification.summary}`, 'error');
        };

        const connectAlertWebSocket = () => {
            if (!isAuthenticated.value || !token.value) return;
            if (alertSocket && (alertSocket.readyState === WebSocket.OPEN || alertSocket.readyState === WebSocket.CONNECTING)) {
                return;
            }

            let ws = null;
            try {
                alertConnectionStatus.value = 'connecting';
                ws = new WebSocket(buildAlertWebSocketURL());
            } catch (error) {
                alertConnectionStatus.value = 'reconnecting';
                scheduleAlertReconnect();
                return;
            }

            alertSocket = ws;

            ws.onopen = () => {
                if (alertSocket !== ws) return;
                alertReconnectAttempts = 0;
                alertConnectionStatus.value = 'connected';
            };

            ws.onmessage = (event) => {
                if (!event || typeof event.data !== 'string') return;
                try {
                    const payload = JSON.parse(event.data);
                    handleAlertMessage(payload);
                } catch (_) {
                    // ignore malformed payload to avoid UI cascade failure
                }
            };

            ws.onerror = () => {
                if (alertSocket !== ws) return;
                if (alertConnectionStatus.value !== 'connected') {
                    alertConnectionStatus.value = 'reconnecting';
                }
            };

            ws.onclose = () => {
                if (alertSocket !== ws) return;
                alertSocket = null;
                if (!isAuthenticated.value) {
                    alertConnectionStatus.value = 'disconnected';
                    return;
                }
                alertConnectionStatus.value = 'reconnecting';
                scheduleAlertReconnect();
            };
        };

        const connectFamilyNotificationWebSocket = () => {
            if (!isAuthenticated.value || !token.value) return;
            if (familyNotificationSocket && (familyNotificationSocket.readyState === WebSocket.OPEN || familyNotificationSocket.readyState === WebSocket.CONNECTING)) {
                return;
            }

            let ws = null;
            try {
                familyNotificationConnectionStatus.value = 'connecting';
                ws = new WebSocket(buildFamilyNotificationWebSocketURL());
            } catch (error) {
                familyNotificationConnectionStatus.value = 'reconnecting';
                scheduleFamilyNotificationReconnect();
                return;
            }

            familyNotificationSocket = ws;

            ws.onopen = () => {
                if (familyNotificationSocket !== ws) return;
                familyNotificationReconnectAttempts = 0;
                familyNotificationConnectionStatus.value = 'connected';
            };

            ws.onmessage = (event) => {
                if (!event || typeof event.data !== 'string') return;
                try {
                    const payload = JSON.parse(event.data);
                    handleFamilyNotificationMessage(payload);
                } catch (_) {
                    // ignore malformed payload to avoid UI cascade failure
                }
            };

            ws.onerror = () => {
                if (familyNotificationSocket !== ws) return;
                if (familyNotificationConnectionStatus.value !== 'connected') {
                    familyNotificationConnectionStatus.value = 'reconnecting';
                }
            };

            ws.onclose = () => {
                if (familyNotificationSocket !== ws) return;
                familyNotificationSocket = null;
                if (!isAuthenticated.value) {
                    familyNotificationConnectionStatus.value = 'disconnected';
                    return;
                }
                familyNotificationConnectionStatus.value = 'reconnecting';
                scheduleFamilyNotificationReconnect();
            };
        };

        const stableJSONStringify = (value) => {
            try {
                return JSON.stringify(value);
            } catch (_) {
                return '';
            }
        };

        const replaceListIfChanged = (targetRef, nextList) => {
            const normalized = Array.isArray(nextList) ? nextList : [];
            if (stableJSONStringify(targetRef.value) !== stableJSONStringify(normalized)) {
                targetRef.value = normalized;
            }
        };

        const requiresGraphCaptcha = computed(() => authMode.value === 'register' || loginMethod.value === 'password');
        const shouldShowSMSCodeSection = computed(() => authMode.value === 'register' || loginMethod.value === 'sms');
        const authSubmitLabel = computed(() => {
            if (authMode.value === 'register') return '创建账户';
            return loginMethod.value === 'sms' ? '短信登录' : '立即登录';
        });
        const smsCodeButtonLabel = computed(() => authMode.value === 'register' ? '发送注册短信码' : '发送登录短信码');
        const smsCodeButtonText = computed(() => smsCodeCooldown.value > 0 ? `${smsCodeCooldown.value}s后重试` : smsCodeButtonLabel.value);
        const canSendSMSCode = computed(() => !smsCodeSending.value && smsCodeCooldown.value === 0);
        const familyUnreadCount = computed(() => (familyNotifications.value || []).filter(item => item && !item.read_at).length);
        const familyHasGroup = computed(() => !!familyOverview.value?.family);
        const familyGuardianCandidates = computed(() => (familyMembers.value || []).filter(item => item && (item.role === 'owner' || item.role === 'guardian')));
        const familyProtectedCandidates = computed(() => (familyMembers.value || []).filter(item => item && item.role !== 'owner'));
        const getUserDisplayName = (userInfo) => {
            const candidate = userInfo && typeof userInfo === 'object' ? userInfo : {};
            return String(candidate.username || '').trim()
                || String(candidate.email || '').trim()
                || String(candidate.phone || '').trim()
                || '未设置';
        };

        const getUserEmailText = (userInfo) => {
            const candidate = userInfo && typeof userInfo === 'object' ? userInfo : {};
            return String(candidate.email || '').trim();
        };

        const getUserPhoneText = (userInfo) => {
            const candidate = userInfo && typeof userInfo === 'object' ? userInfo : {};
            return String(candidate.phone || '').trim();
        };

        const getUserAvatarText = (userInfo) => getUserDisplayName(userInfo).slice(0, 2).toUpperCase();

        const clearSMSCodeCooldownTimer = () => {
            if (smsCodeCooldownTimer) {
                clearInterval(smsCodeCooldownTimer);
                smsCodeCooldownTimer = null;
            }
        };

        const startSMSCodeCooldown = () => {
            clearSMSCodeCooldownTimer();
            smsCodeCooldown.value = 60;
            smsCodeCooldownTimer = setInterval(() => {
                if (smsCodeCooldown.value <= 1) {
                    smsCodeCooldown.value = 0;
                    clearSMSCodeCooldownTimer();
                    return;
                }
                smsCodeCooldown.value -= 1;
            }, 1000);
        };

        const request = async (path, method = 'GET', body = null, options = {}) => {
            const { silent = false } = options || {};
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
                if (!silent) {
                    showToast(e.message, 'error');
                }
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

        const sendSMSCode = async () => {
            if (!canSendSMSCode.value) return;
            const phone = form.phone.trim();
            if (!phone) {
                showToast('请输入手机号', 'error');
                return;
            }

            smsCodeSending.value = true;
            const res = await request('/auth/sms-code', 'POST', { phone });
            smsCodeSending.value = false;
            if (res) {
                showToast(res.message || `短信验证码已发送，请使用 ${demoSMSCode}`);
                startSMSCodeCooldown();
            }
        };

        const buildAuthPayload = () => {
            if (authMode.value === 'register') {
                return {
                    username: form.username.trim(),
                    email: form.email.trim(),
                    phone: form.phone.trim(),
                    password: form.password,
                    captchaId: captchaId.value,
                    captchaCode: form.captchaCode.trim(),
                    smsCode: form.smsCode.trim()
                };
            }

            if (loginMethod.value === 'sms') {
                return {
                    phone: form.phone.trim(),
                    smsCode: form.smsCode.trim()
                };
            }

            return {
                account: form.account.trim(),
                password: form.password,
                captchaId: captchaId.value,
                captchaCode: form.captchaCode.trim()
            };
        };

        const handleAuth = async () => {
            loading.value = true;
            const endpoint = authMode.value === 'login' ? '/auth/login' : '/auth/register';
            const payload = buildAuthPayload();
            
            const res = await request(endpoint, 'POST', payload);
            loading.value = false;
            
            if (res) {
                if (authMode.value === 'register') {
                    showToast('注册成功，请登录');
                    authMode.value = 'login';
                    loginMethod.value = 'password';
                    form.account = form.email.trim();
                    form.password = '';
                    form.captchaCode = '';
                    form.smsCode = '';
                    fetchCaptcha();
                } else {
                    token.value = res.token;
                    localStorage.setItem('token', res.token);
                    isAuthenticated.value = true;
                    user.value = res.user;
                    syncProfileForm(res.user);
                    fetchOccupationOptions();
                    showToast('登录成功');
                    startPolling();
                }
            } else if (requiresGraphCaptcha.value) {
                fetchCaptcha(); // Refresh captcha on fail
            }
        };

        const getUserInfo = async () => {
            const res = await request('/user');
            if (res) {
                user.value = res;
                syncProfileForm(res);
                fetchOccupationOptions();
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
            syncProfileForm({});
            stopPolling();
            alertDrawerVisible.value = false;
            alertEvents.value = [];
            alertUnreadCount.value = 0;
            alertModalVisible.value = false;
            activeAlertEvent.value = null;
            alertSeenRecordIDs.clear();
            familyNotifications.value = [];
            familyNotificationSeenIDs.clear();
            familyOverview.value = null;
            familyMembers.value = [];
            familyInvitations.value = [];
            familyReceivedInvitations.value = [];
            familyGuardianLinks.value = [];
            familyReceivedLoading.value = false;
            Object.keys(familyAcceptingInvitations).forEach((key) => delete familyAcceptingInvitations[key]);
            familyAlertModalVisible.value = false;
            activeFamilyNotification.value = null;
        };

        const updateUserProfile = async () => {
            const normalizedAge = Number(ageForm.age);
            if (!Number.isFinite(normalizedAge) || normalizedAge < 1 || normalizedAge > 150) {
                showToast('年龄请输入 1 到 150 之间的数字', 'error');
                return;
            }

            profileSaving.value = true;
            const res = await request('/user/profile', 'PUT', {
                age: normalizedAge,
                occupation: profileForm.occupation
            });
            profileSaving.value = false;
            if (res && res.user) {
                user.value = res.user;
                syncProfileForm(res.user);
                showToast(res.message || '用户画像更新成功');
            }
        };
        const updateAge = updateUserProfile;

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

        watch([authMode, loginMethod], () => {
            if (requiresGraphCaptcha.value) {
                fetchCaptcha();
            } else {
                form.captchaCode = '';
            }
        });

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

        const fetchTasks = async ({ silent = false } = {}) => {
            if (!isAuthenticated.value) return;
            const res = await request('/scam/multimodal/tasks', 'GET', null, { silent });
            if (res && res.tasks) {
                const nextTasks = [...res.tasks].sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
                replaceListIfChanged(tasks, nextTasks);
            }
        };

        const fetchHistory = async ({ silent = false } = {}) => {
            const res = await request('/scam/multimodal/history', 'GET', null, { silent });
            if (res && res.history) {
                replaceListIfChanged(history, res.history);
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

                fetchHistory();
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
                replaceListIfChanged(users, res.users);
            }
        };

        let debounceTimer;
        const debouncedFetchUsers = () => {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(fetchUsers, 300);
        };

        // Family System
        const pruneFamilyNotificationsByMembers = (members) => {
            const activeUserIDs = new Set(
                (Array.isArray(members) ? members : [])
                    .map(item => Number(item?.user_id || 0))
                    .filter(userID => Number.isInteger(userID) && userID > 0)
            );

            if (activeUserIDs.size === 0) {
                familyNotifications.value = [];
                familyNotificationSeenIDs.clear();
                activeFamilyNotification.value = null;
                familyAlertModalVisible.value = false;
                return;
            }

            familyNotifications.value = familyNotifications.value.filter(item => {
                const targetUserID = Number(item?.target_user_id || 0);
                return Number.isInteger(targetUserID) && activeUserIDs.has(targetUserID);
            });

            familyNotificationSeenIDs.clear();
            familyNotifications.value.forEach(item => {
                const notificationID = Number(item?.id || 0);
                if (Number.isInteger(notificationID) && notificationID > 0) {
                    familyNotificationSeenIDs.add(notificationID);
                }
            });

            const activeTargetUserID = Number(activeFamilyNotification.value?.target_user_id || 0);
            if (!Number.isInteger(activeTargetUserID) || !activeUserIDs.has(activeTargetUserID)) {
                activeFamilyNotification.value = null;
                familyAlertModalVisible.value = false;
            }
        };

        const hydrateFamilyOverview = (overview) => {
            familyOverview.value = overview || null;
            familyMembers.value = Array.isArray(overview?.members) ? overview.members : [];
            familyInvitations.value = Array.isArray(overview?.invitations) ? overview.invitations : [];
            familyGuardianLinks.value = Array.isArray(overview?.guardian_links) ? overview.guardian_links : [];
            if (overview?.family) {
                replaceListIfChanged(familyReceivedInvitations, []);
                familyReceivedLoading.value = false;
            }
            pruneFamilyNotificationsByMembers(familyMembers.value);
        };

        const fetchFamilyOverview = async ({ silent = false } = {}) => {
            if (!isAuthenticated.value) return;
            familyLoading.value = !silent;
            const res = await request('/families/me', 'GET', null, { silent });
            familyLoading.value = false;
            if (res) {
                hydrateFamilyOverview(res);
            }
        };

        const fetchReceivedFamilyInvitations = async ({ silent = false } = {}) => {
            if (!isAuthenticated.value) return;
            familyReceivedLoading.value = !silent;
            const res = await request('/families/invitations/received', 'GET', null, { silent });
            familyReceivedLoading.value = false;
            if (res && Array.isArray(res.invitations)) {
                replaceListIfChanged(familyReceivedInvitations, res.invitations);
            }
        };

        const createFamily = async () => {
            const payload = { name: familyCreateForm.name.trim() };
            const res = await request('/families', 'POST', payload);
            if (res) {
                hydrateFamilyOverview(res);
                familyCreateForm.name = '';
                showToast('家庭创建成功');
                familyNotifications.value = [];
                familyNotificationSeenIDs.clear();
                connectFamilyNotificationWebSocket();
            }
        };

        const createFamilyInvitation = async () => {
            const payload = {
                invitee_email: familyInviteForm.invitee_email.trim(),
                invitee_phone: familyInviteForm.invitee_phone.trim(),
                role: familyInviteForm.role,
                relation: familyInviteForm.relation.trim()
            };
            const res = await request('/families/invitations', 'POST', payload);
            if (res && res.invitation) {
                showToast('家庭邀请已创建');
                familyInviteForm.invitee_email = '';
                familyInviteForm.invitee_phone = '';
                familyInviteForm.role = 'member';
                familyInviteForm.relation = '';
                fetchFamilyOverview({ silent: true });
            }
        };

        const acceptFamilyInvitation = async () => {
            const payload = { invite_code: familyAcceptForm.invite_code.trim() };
            const res = await request('/families/invitations/accept', 'POST', payload);
            if (res) {
                hydrateFamilyOverview(res);
                familyAcceptForm.invite_code = '';
                showToast('已加入家庭');
                familyNotifications.value = [];
                familyNotificationSeenIDs.clear();
                connectFamilyNotificationWebSocket();
            }
        };

        const acceptReceivedFamilyInvitation = async (rawInviteCode = '', invitationID = 0) => {
            const inviteCode = String(rawInviteCode || familyAcceptForm.invite_code || '').trim();
            if (!inviteCode) {
                showToast('璇疯緭鍏ュ搴個璇风爜', 'error');
                return;
            }

            if (invitationID) {
                familyAcceptingInvitations[invitationID] = true;
            }
            try {
                familyAcceptForm.invite_code = inviteCode;
                await acceptFamilyInvitation();
                if (familyHasGroup.value) {
                    replaceListIfChanged(familyReceivedInvitations, []);
                    familyReceivedLoading.value = false;
                }
            } finally {
                if (invitationID) {
                    familyAcceptingInvitations[invitationID] = false;
                }
            }
        };

        const createGuardianLink = async () => {
            const guardianUserID = Number(familyGuardianForm.guardian_user_id);
            const memberUserID = Number(familyGuardianForm.member_user_id);
            if (!Number.isInteger(guardianUserID) || !Number.isInteger(memberUserID)) {
                showToast('请选择守护人和被守护成员', 'error');
                return;
            }
            const payload = {
                guardian_user_id: guardianUserID,
                member_user_id: memberUserID
            };
            const res = await request('/families/guardian-links', 'POST', payload);
            if (res && res.guardian_link) {
                showToast('守护关系配置成功');
                familyGuardianForm.guardian_user_id = '';
                familyGuardianForm.member_user_id = '';
                fetchFamilyOverview({ silent: true });
            }
        };

        const deleteFamilyMember = async (member) => {
            if (!member || !member.member_id) return;
            if (!confirm(`确定移除成员 ${member.username} 吗？`)) return;
            familyDeletingMembers[member.member_id] = true;
            try {
                const res = await request(`/families/members/${encodeURIComponent(member.member_id)}`, 'DELETE');
                if (res) {
                    showToast(res.message || '成员已移除');
                    fetchFamilyOverview({ silent: true });
                }
            } finally {
                familyDeletingMembers[member.member_id] = false;
            }
        };

        const deleteGuardianLink = async (link) => {
            if (!link || !link.id) return;
            if (!confirm(`确定取消 ${link.guardian_name} -> ${link.member_name} 的守护关系吗？`)) return;
            familyDeletingGuardianLinks[link.id] = true;
            try {
                const res = await request(`/families/guardian-links/${encodeURIComponent(link.id)}`, 'DELETE');
                if (res) {
                    showToast(res.message || '守护关系已移除');
                    fetchFamilyOverview({ silent: true });
                }
            } finally {
                familyDeletingGuardianLinks[link.id] = false;
            }
        };

        const markFamilyNotificationRead = async (notification) => {
            if (!notification || !notification.id || notification.read_at) return;
            familyMarkingNotifications[notification.id] = true;
            try {
                const res = await request(`/families/notifications/${encodeURIComponent(notification.id)}/read`, 'POST');
                if (res) {
                    const readAt = new Date().toISOString();
                    familyNotifications.value = familyNotifications.value.map(item => item && item.id === notification.id ? { ...item, read_at: readAt } : item);
                }
            } finally {
                familyMarkingNotifications[notification.id] = false;
            }
        };

        // Case Library Management
        const fetchCaseLibrary = async () => {
            if (!isAuthenticated.value || (user.value.role !== 'admin')) return;
            const res = await request('/scam/case-library/cases');
            if (res && res.cases) {
                replaceListIfChanged(caseLibrary, res.cases);
            }
        };

        // Pending Review Management
        const fetchPendingReviews = async () => {
            if (!isAuthenticated.value || (user.value.role !== 'admin')) return;
            const res = await request('/scam/review/cases');
            if (res && res.cases) {
                replaceListIfChanged(pendingReviews, res.cases);
            }
        };

        const viewReviewDetail = async (recordId) => {
            const res = await request(`/scam/review/cases/${recordId}`);
            if (res && res.case) {
                selectedReview.value = res.case;
                showReviewDetailModal.value = true;
            }
        };

        const approveReview = async (recordId) => {
            if (!confirm('确认通过该案件审核并入库知识库？')) return;
            const res = await request(`/scam/review/cases/${recordId}/approve`, 'POST');
            if (res && res.case_id) {
                showReviewDetailModal.value = false;
                selectedReview.value = null;
                fetchPendingReviews();
            }
        };

        const submitCaseCollection = async () => {
            const query = String(caseCollectionForm.query || '').trim();
            const caseCount = Number(caseCollectionForm.case_count);
            if (!query) {
                showToast('采集主题不能为空', 'error');
                return;
            }
            if (!Number.isInteger(caseCount) || caseCount < 1 || caseCount > 20) {
                showToast('案件数量取值范围应为 1-20', 'error');
                return;
            }

            startingCaseCollection.value = true;
            try {
                const res = await request('/scam/case-collection/search', 'POST', {
                    query,
                    case_count: caseCount
                });
                showToast((res && res.message) || '案件采集任务已在后台启动');
                caseCollectionForm.query = '';
                setTimeout(() => fetchPendingReviews(), 1200);
            } catch (e) {
                showToast('启动失败: ' + e.message, 'error');
            } finally {
                startingCaseCollection.value = false;
            }
        };

        const fetchCaseOptionLists = async () => {
            if (!isAuthenticated.value || (user.value.role !== 'admin')) return;
            const [scamTypeRes, targetGroupRes] = await Promise.all([
                request('/scam/case-library/options/scam-types'),
                request('/scam/case-library/options/target-groups')
            ]);
            if (scamTypeRes && Array.isArray(scamTypeRes.options)) {
                replaceListIfChanged(scamTypeOptions, scamTypeRes.options);
            }
            if (targetGroupRes && Array.isArray(targetGroupRes.options)) {
                replaceListIfChanged(targetGroupOptions, targetGroupRes.options);
            }
        };

        const openCaseModal = () => {
            Object.assign(caseForm, {
                title: '',
                target_group: '',
                risk_level: '',
                scam_type: '',
                case_description: '',
                typical_scripts_raw: '',
                keywords_raw: '',
                violated_law: '',
                suggestion: ''
            });
            showCaseModal.value = true;
        };

        const minCaseDescriptionRunes = 12;
        const maxCaseDescriptionRunes = 400;
        const randomLikeAlnumChunkLimit = 16;

        const validateCaseDescriptionQualityClient = (description) => {
            const normalized = String(description || '').trim().replace(/\s+/g, ' ');
            const runes = Array.from(normalized);

            if (!normalized) {
                return { ok: false, message: '案件描述不能为空' };
            }
            if (runes.length < minCaseDescriptionRunes) {
                return { ok: false, message: `案件描述过短，至少 ${minCaseDescriptionRunes} 个字符` };
            }

            if (runes.length > maxCaseDescriptionRunes) {
                return { ok: false, message: `案件描述过长，最多 ${maxCaseDescriptionRunes} 个字符` };
            }

            const uniqueChars = new Set(runes);
            if (uniqueChars.size <= 2) {
                return { ok: false, message: '案件描述疑似无效，请填写有语义的内容' };
            }

            const hasHan = /[\u4e00-\u9fff]/.test(normalized);
            const hasSeparator = /[^A-Za-z0-9]/.test(normalized);

            let maxAlnumChunk = 0;
            let currentAlnumChunk = 0;
            let alnumCount = 0;
            let digitCount = 0;

            for (const ch of runes) {
                if (/[A-Za-z0-9]/.test(ch)) {
                    currentAlnumChunk += 1;
                    alnumCount += 1;
                    if (/[0-9]/.test(ch)) {
                        digitCount += 1;
                    }
                } else {
                    currentAlnumChunk = 0;
                }

                if (currentAlnumChunk > maxAlnumChunk) {
                    maxAlnumChunk = currentAlnumChunk;
                }
            }

            if (!hasHan && !hasSeparator && maxAlnumChunk >= randomLikeAlnumChunkLimit) {
                return { ok: false, message: '案件描述疑似随机字符串，请补充有效描述' };
            }

            if (!hasHan && alnumCount >= minCaseDescriptionRunes) {
                const digitRatio = digitCount / alnumCount;
                if (maxAlnumChunk >= minCaseDescriptionRunes && digitRatio > 0.35) {
                    return { ok: false, message: '案件描述疑似随机字符串，请补充有效描述' };
                }
            }

            return { ok: true, message: '' };
        };

        const submitCase = async () => {
            if (!String(caseForm.title || '').trim()) {
                showToast('案件标题不能为空', 'error');
                return;
            }
            if (!String(caseForm.target_group || '').trim()) {
                showToast('目标人群不能为空', 'error');
                return;
            }
            if (!String(caseForm.risk_level || '').trim()) {
                showToast('风险等级不能为空', 'error');
                return;
            }
            if (!String(caseForm.scam_type || '').trim()) {
                showToast('诈骗类型不能为空', 'error');
                return;
            }

            const descriptionValidation = validateCaseDescriptionQualityClient(caseForm.case_description);
            if (!descriptionValidation.ok) {
                showToast(descriptionValidation.message, 'error');
                return;
            }

            submittingCase.value = true;
            try {
                const payload = {
                    title: String(caseForm.title || '').trim(),
                    target_group: String(caseForm.target_group || '').trim(),
                    risk_level: String(caseForm.risk_level || '').trim(),
                    scam_type: String(caseForm.scam_type || '').trim(),
                    case_description: String(caseForm.case_description || '').trim(),
                    typical_scripts: caseForm.typical_scripts_raw.split('\n').filter(s => s.trim()),
                    keywords: caseForm.keywords_raw.split(/[,，]/).map(s => s.trim()).filter(s => s),
                    violated_law: String(caseForm.violated_law || '').trim(),
                    suggestion: String(caseForm.suggestion || '').trim()
                };

                const res = await request('/scam/case-library/cases', 'POST', payload);
                if (res) {
                    showToast('案件录入成功');
                    showCaseModal.value = false;
                    fetchCaseLibrary();
                    fetchCaseOptionLists();
                    fetchAdminStats(true);
                }
            } catch (e) {
                showToast('录入失败: ' + e.message, 'error');
            } finally {
                submittingCase.value = false;
            }
        };

        const viewCaseDetail = async (caseId) => {
            const res = await request(`/scam/case-library/cases/${caseId}`);
            if (res && res.case) {
                selectedCase.value = res.case;
            }
        };

        const deleteCase = async (item) => {
             if (!item || !item.case_id) return;
             if (!confirm(`确定删除案件 ${item.title} 吗？此操作不可恢复。`)) return;

             try {
                 const res = await request(`/scam/case-library/cases/${item.case_id}`, 'DELETE');
                 if (res) {
                     showToast(res.message || '案件已删除');
                     fetchCaseLibrary();
                     fetchAdminStats(true);
                     if (selectedCase.value && selectedCase.value.case_id === item.case_id) {
                         selectedCase.value = null;
                     }
                 }
             } catch (e) {
                 showToast('删除失败: ' + e.message, 'error');
             }
        };

        // Risk Trend Logic
        const fetchRiskTrend = async (forceRefresh = false) => {
            if (!isAuthenticated.value) return;

            const interval = riskInterval.value;
            
            // 1. 如果有缓存，先立即渲染缓存数据（Stale-While-Revalidate 策略）
            if (riskCache[interval]) {
                riskData.value = riskCache[interval];
                setTimeout(() => renderCharts(), 0);
                
                // 如果不是强制刷新，且距离上次更新很近（例如10秒内），可以考虑不发请求
                // 但为了响应用户需求“点击日周年会访问API”，我们这里总是继续发送请求
            }

            // 2. 静默请求最新数据
            try {
                const res = await request(`/scam/multimodal/history/overview?interval=${interval}`, 'GET', null, { silent: true });
                if (res) {
                    // 3. 比较数据是否有变化
                    const cachedData = riskCache[interval];
                    const hasChanged = !cachedData || stableJSONStringify(cachedData) !== stableJSONStringify(res);

                    if (hasChanged || forceRefresh) {
                        riskData.value = res;
                        riskCache[interval] = res;
                        setTimeout(() => renderCharts(), 100);
                        if (forceRefresh) showToast('数据已更新');
                    }
                }
            } catch (e) {
                console.error('Fetch risk trend failed:', e);
            }
        };

        const formatChartLabel = (label, intervalType) => {
            const interval = intervalType || riskInterval.value;
            if (interval === 'week' && label.includes('-W')) {
                try {
                    const [yearStr, weekStr] = label.split('-W');
                    const year = parseInt(yearStr);
                    const week = parseInt(weekStr);
                    
                    // ISO Week calculation: 1st week contains Jan 4th
                    const jan4 = new Date(year, 0, 4);
                    const jan4Day = jan4.getDay() || 7; // Mon=1, Sun=7
                    const week1Start = new Date(year, 0, 4 - jan4Day + 1);
                    const start = new Date(week1Start.getTime() + (week - 1) * 7 * 86400000);
                    const end = new Date(start.getTime() + 6 * 86400000);
                    
                    const fmt = d => `${d.getMonth() + 1}.${d.getDate()}`;
                    return `${year}年第${week}周 (${fmt(start)}-${fmt(end)})`;
                } catch (e) {
                    return label;
                }
            }
            if (interval === 'month' && /^\d{4}-\d{2}$/.test(label)) {
                const [y, m] = label.split('-');
                return `${y}年${parseInt(m)}月`;
            }
            return label;
        };

        const fillTrendGaps = (sparseTrend, interval) => {
            if (!sparseTrend || sparseTrend.length === 0) return [];
            
            // Sort by time_bucket first
            const sorted = [...sparseTrend].sort((a, b) => a.time_bucket.localeCompare(b.time_bucket));
            const startBucket = sorted[0].time_bucket;
            const endBucket = sorted[sorted.length - 1].time_bucket;
            
            const filled = [];
            const dataMap = new Map(sorted.map(item => [item.time_bucket, item]));
            
            let current = startBucket;
            const maxIterations = 500; // Safety break
            let count = 0;

            while (current <= endBucket && count < maxIterations) {
                count++;
                if (dataMap.has(current)) {
                    filled.push(dataMap.get(current));
                } else {
                    // Create a zeroed point
                    const zeroPoint = { time_bucket: current, total: 0 };
                    // For user risk trend, it has high/medium/low fields
                    if ('high' in sorted[0]) {
                        zeroPoint.high = 0;
                        zeroPoint.medium = 0;
                        zeroPoint.low = 0;
                    }
                    // For admin trend, it has count field
                    if ('count' in sorted[0]) {
                        zeroPoint.count = 0;
                    }
                    filled.push(zeroPoint);
                }

                // Increment current bucket
                if (interval === 'day') {
                    const date = new Date(current);
                    date.setDate(date.getDate() + 1);
                    current = date.toISOString().split('T')[0];
                } else if (interval === 'week') {
                    const [y, w] = current.split('-W').map(Number);
                    let nextW = w + 1;
                    let nextY = y;
                    // Simplified week overflow (actual ISO week max is 52 or 53, 
                    // but since we only compare string current <= endBucket, it's safer)
                    if (nextW > 53) { nextW = 1; nextY++; }
                    current = `${nextY}-W${String(nextW).padStart(2, '0')}`;
                } else if (interval === 'month') {
                    const [y, m] = current.split('-').map(Number);
                    let nextM = m + 1;
                    let nextY = y;
                    if (nextM > 12) { nextM = 1; nextY++; }
                    current = `${nextY}-${String(nextM).padStart(2, '0')}`;
                } else {
                    break;
                }
            }
            return filled;
        };

        const renderCharts = () => {
            if (!riskData.value) return;
            if (typeof window.echarts === 'undefined') {
                console.warn('ECharts 未加载，已跳过用户侧图表渲染。');
                return;
            }
            const stats = riskData.value.stats;
            const trend = fillTrendGaps(riskData.value.trend, riskInterval.value);

            // Destroy old charts
            if (pieChartInstance && typeof pieChartInstance.dispose === 'function') pieChartInstance.dispose();
            if (lineChartInstance && typeof lineChartInstance.dispose === 'function') lineChartInstance.dispose();

            // Pie Chart
            const pieDom = document.getElementById('riskPieChart');
            if (pieDom) {
                pieChartInstance = echarts.init(pieDom);
                pieChartInstance.setOption({
                    tooltip: { trigger: 'item', backgroundColor: 'rgba(15, 23, 42, 0.9)', textStyle: { color: '#fff' } },
                    series: [{
                        name: '风险分布',
                        type: 'pie',
                        radius: ['50%', '80%'],
                        avoidLabelOverlap: false,
                        itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
                        label: { show: false },
                        emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
                        data: [
                            { value: stats.high, name: '高风险', itemStyle: { color: '#ef4444' } },
                            { value: stats.medium, name: '中风险', itemStyle: { color: '#f59e0b' } },
                            { value: stats.low, name: '低风险', itemStyle: { color: '#10b981' } }
                        ]
                    }]
                });
            }

            // Line Chart
            const lineDom = document.getElementById('riskLineChart');
            if (lineDom) {
                lineChartInstance = echarts.init(lineDom);
                lineChartInstance.setOption({
                    tooltip: {
                        trigger: 'axis',
                        backgroundColor: 'rgba(15, 23, 42, 0.9)',
                        textStyle: { color: '#fff' }
                    },
                    legend: { bottom: 0, textStyle: { fontSize: 11 } },
                    grid: { left: '3%', right: '4%', top: '10%', bottom: '15%', containLabel: true },
                    xAxis: {
                        type: 'category',
                        boundaryGap: false,
                        data: trend.map(item => formatChartLabel(item.time_bucket)),
                        axisLabel: { color: '#64748b' }
                    },
                    yAxis: {
                        type: 'value',
                        axisLabel: { color: '#64748b' },
                        splitLine: { lineStyle: { type: 'dashed', color: 'rgba(148, 163, 184, 0.1)' } }
                    },
                    series: [
                        {
                            name: '高风险',
                            type: 'line',
                            smooth: true,
                            data: trend.map(item => item.high),
                            itemStyle: { color: '#ef4444' },
                            lineStyle: { width: 3 }
                        },
                        {
                            name: '中风险',
                            type: 'line',
                            smooth: true,
                            data: trend.map(item => item.medium),
                            itemStyle: { color: '#f59e0b' },
                            lineStyle: { width: 3 }
                        },
                        {
                            name: '低风险',
                            type: 'line',
                            smooth: true,
                            data: trend.map(item => item.low),
                            itemStyle: { color: '#10b981' },
                            lineStyle: { width: 3 }
                        }
                    ]
                });
            }
        };

        const getRiskTrendAnalysisClass = (trendText) => {
            switch ((trendText || '').trim()) {
                case '上升':
                    return 'bg-red-50 text-red-700 ring-1 ring-red-200';
                case '下降':
                    return 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200';
                case '平稳':
                    return 'bg-amber-50 text-amber-700 ring-1 ring-amber-200';
                default:
                    return 'bg-slate-100 text-slate-600 ring-1 ring-slate-200';
            }
        };

        const formatAdminChartLabel = (label) => {
            const interval = adminStatsInterval.value;
            if (interval === 'week' && label.includes('-W')) {
                try {
                    const [yearStr, weekStr] = label.split('-W');
                    const year = parseInt(yearStr);
                    const week = parseInt(weekStr);
                    const jan4 = new Date(year, 0, 4);
                    const jan4Day = jan4.getDay() || 7; 
                    const week1Start = new Date(year, 0, 4 - jan4Day + 1);
                    const start = new Date(week1Start.getTime() + (week - 1) * 7 * 86400000);
                    const end = new Date(start.getTime() + 6 * 86400000);
                    const fmt = d => `${d.getMonth() + 1}.${d.getDate()}`;
                    return `${year}年第${week}周 (${fmt(start)}-${fmt(end)})`;
                } catch (e) {
                    return label;
                }
            }
            if (interval === 'month' && /^\d{4}-\d{2}$/.test(label)) {
                const [y, m] = label.split('-');
                return `${y}年${parseInt(m)}月`;
            }
            return label;
        };

        const fetchAdminStats = async (forceRefresh = false) => {
            if (!isAuthenticated.value || user.value.role !== 'admin') return;

            const interval = adminStatsInterval.value;
            
            // 1. Check Cache
            if (adminStatsCache[interval]) {
                adminStatsData.value = adminStatsCache[interval];
                setTimeout(() => renderAdminCharts(), 0);
            }

            // 2. Silent Update
            try {
                const res = await request(`/scam/case-library/cases/overview?interval=${interval}`, 'GET', null, { silent: true });
                if (res) {
                    const cachedData = adminStatsCache[interval];
                    const hasChanged = !cachedData || stableJSONStringify(cachedData) !== stableJSONStringify(res);

                    if (hasChanged || forceRefresh) {
                        adminStatsData.value = res;
                        adminStatsCache[interval] = res;
                        setTimeout(() => renderAdminCharts(), 100);
                        if (forceRefresh) showToast('全景数据已更新');
                    }
                }
            } catch (e) {
                console.error('Fetch admin stats failed:', e);
            }

            await fetchAdminCaseGraph(forceRefresh);
        };

        const fetchAdminCaseGraph = async (forceRefresh = false) => {
            if (!isAuthenticated.value || user.value.role !== 'admin') return;

            if (adminGraphCache.value) {
                adminGraphData.value = adminGraphCache.value;
            }

            try {
                const res = await request('/scam/case-library/cases/graph?top_k=5', 'GET', null, { silent: true });
                if (res) {
                    const cachedData = adminGraphCache.value;
                    const hasChanged = !cachedData || stableJSONStringify(cachedData) !== stableJSONStringify(res);
                    if (hasChanged || forceRefresh) {
                        adminGraphData.value = res;
                        adminGraphCache.value = res;
                    }
                }

                if (selectedGraphTargetGroup.value) {
                    await fetchAdminTargetGroupChart(selectedGraphTargetGroup.value, forceRefresh);
                }
            } catch (e) {
                console.error('Fetch admin graph failed:', e);
            }
        };

        const fetchAdminTargetGroupChart = async (targetGroup, forceRefresh = false) => {
            if (!isAuthenticated.value || user.value.role !== 'admin') return;

            const normalizedTargetGroup = String(targetGroup || '').trim();
            if (!normalizedTargetGroup) {
                clearAdminTargetGroupFocus();
                return;
            }

            selectedGraphTargetGroup.value = normalizedTargetGroup;

            if (adminTargetGroupChartCache[normalizedTargetGroup] && !forceRefresh) {
                adminTargetGroupChartData.value = adminTargetGroupChartCache[normalizedTargetGroup];
                setTimeout(() => renderAdminTargetGroupBarChart(), 0);
                return;
            }

            // Show loading if instance exists
            if (adminTargetGroupBarChart) {
                adminTargetGroupBarChart.showLoading({
                    text: '分析中...',
                    color: '#6366f1',
                    textColor: '#6366f1',
                    maskColor: 'rgba(255, 255, 255, 0.2)',
                    zlevel: 0
                });
            }

            try {
                const query = `/scam/case-library/cases/graph?top_k=5&focus_group=${encodeURIComponent(normalizedTargetGroup)}`;
                const res = await request(query, 'GET', null, { silent: true });
                const nextData = res && Array.isArray(res.target_group_top_scam_types)
                    ? (res.target_group_top_scam_types[0] || null)
                    : null;

                adminTargetGroupChartData.value = nextData;
                if (nextData) {
                    adminTargetGroupChartCache[normalizedTargetGroup] = nextData;
                    setTimeout(() => {
                        renderAdminTargetGroupBarChart();
                        if (adminTargetGroupBarChart) adminTargetGroupBarChart.hideLoading();
                    }, 0);
                } else if (adminTargetGroupBarChart) {
                    adminTargetGroupBarChart.hideLoading();
                    adminTargetGroupBarChart.destroy();
                    adminTargetGroupBarChart = null;
                }
            } catch (e) {
                console.error('Fetch admin target group chart failed:', e);
                if (adminTargetGroupBarChart) adminTargetGroupBarChart.hideLoading();
            }
        };

        const clearAdminTargetGroupFocus = () => {
            selectedGraphTargetGroup.value = '';
            adminTargetGroupChartData.value = null;
            if (adminTargetGroupBarChart) {
                adminTargetGroupBarChart.destroy();
                adminTargetGroupBarChart = null;
            }
        };

        const openGraphModal = () => {
            showGraphModal.value = true;
            setTimeout(() => renderAdminGraphNetwork(), 300);
        };

        const renderAdminGraphNetwork = () => {
            if (!adminGraphData.value || !adminGraphData.value.graph) return;
            const container = document.getElementById('adminGraphNetwork');
            if (!container) return;

            const { nodes, edges } = adminGraphData.value.graph;

            const visNodes = new vis.DataSet(nodes.map(node => {
                let background = '#6366f1'; // Default Indigo-500
                let border = '#818cf8'; // Indigo-400
                let size = 20;

                if (node.node_type === 'scam_type') {
                    background = '#d946ef'; // Fuchsia-500
                    border = '#f0abfc'; // Fuchsia-300
                    size = 35;
                } else if (node.node_type === 'target_group') {
                    background = '#10b981'; // Emerald-500
                    border = '#6ee7b7'; // Emerald-300
                    size = 42; // Larger to act as an anchor
                } else if (node.node_type === 'keyword') {
                    background = '#818cf8'; // Indigo-400
                    border = '#c7d2fe'; // Indigo-200
                    size = 18;
                }

                return {
                    id: node.id,
                    label: node.label,
                    nodeType: node.node_type, // Custom property for interaction
                    title: `类型: ${node.node_type}\n名称: ${node.label}${node.properties?.case_count ? '\n案件数: ' + node.properties.case_count : ''}`,
                    color: {
                        background: background,
                        border: border,
                        highlight: { background: background, border: '#000' },
                        hover: { background: background, border: border }
                    },
                    font: { color: '#475569', size: 14, face: 'Plus Jakarta Sans', weight: '800' },
                    size: size,
                    shape: node.node_type === 'target_group' ? 'diamond' : 'dot', // Diamond shape for people anchors
                    borderWidth: node.node_type === 'target_group' ? 6 : 4,
                    shadow: {
                        enabled: true,
                        color: 'rgba(0,0,0,0.1)',
                        size: 12,
                        x: 0,
                        y: 4
                    }
                };
            }));

            const visEdges = new vis.DataSet(edges.map(edge => {
                let label = '';
                let dashes = false;
                let color = '#e2e8f0'; // Slate-200
                const relation = String(edge.relation || edge.relation_type || '').trim();

                if (relation === 'similar' || relation === 'similar_to') {
                    label = '相似';
                    dashes = true;
                    color = '#fbcfe8'; // Pink-200
                } else if (relation === 'targets' || relation === 'target_of') {
                    label = '针对';
                    color = '#d1fae5'; // Emerald-100
                } else if (relation === 'keyword' || relation === 'has_keyword') {
                    label = '关键词';
                    color = '#e0e7ff'; // Indigo-100
                }

                return {
                    from: edge.source,
                    to: edge.target,
                    label: label,
                    arrows: { to: { enabled: true, scaleFactor: 0.5, type: 'arrow' } },
                    font: { size: 11, align: 'middle', color: '#94a3b8', strokeWidth: 0 },
                    color: { color: color, highlight: color, hover: color },
                    width: edge.weight ? Math.max(1.5, edge.weight * 4) : 1.5,
                    dashes: dashes,
                    smooth: { type: 'curvedCW', roundness: 0.2 }
                };
            }));

            const data = { nodes: visNodes, edges: visEdges };
            const options = {
                nodes: {
                    borderWidthSelected: 2
                },
                edges: {
                    hoverWidth: 1.5,
                    selectionWidth: 2
                },
                physics: {
                    forceAtlas2Based: {
                        gravitationalConstant: -80,
                        centralGravity: 0.005,
                        springLength: 180,
                        springConstant: 0.04,
                        avoidOverlap: 1
                    },
                    maxVelocity: 45,
                    solver: 'forceAtlas2Based',
                    timestep: 0.35,
                    stabilization: { iterations: 200, updateInterval: 25 }
                },
                interaction: {
                    hover: true,
                    tooltipDelay: 100,
                    navigationButtons: false,
                    keyboard: true,
                    zoomView: true,
                    dragView: true
                }
            };

            if (adminNetworkInstance) {
                adminNetworkInstance.destroy();
            }
            adminNetworkInstance = new vis.Network(container, data, options);

            // Click interaction for "People-oriented" perspective
            adminNetworkInstance.on('click', (params) => {
                if (params.nodes.length > 0) {
                    const nodeId = params.nodes[0];
                    const node = visNodes.get(nodeId);
                    
                    if (node && node.nodeType === 'target_group') {
                        // If it's a "Target Group", select all associated "Scam Types"
                        const connectedNodeIds = adminNetworkInstance.getConnectedNodes(nodeId);
                        const allToSelect = [nodeId, ...connectedNodeIds];
                        adminNetworkInstance.selectNodes(allToSelect);
                        fetchAdminTargetGroupChart(node.label);
                        
                        showToast(`已切换至【${node.label}】人群视角，关联 ${connectedNodeIds.length} 种诈骗手法`, 'success');
                    }
                }
            });
        };

        const resetGraphZoom = () => {
            if (adminNetworkInstance) {
                adminNetworkInstance.fit({ animation: true });
            }
        };

        const formatGraphScore = (score) => {
            const value = Number(score);
            if (!Number.isFinite(value)) return '--';
            return `${(value * 100).toFixed(1)}%`;
        };

        const availableGraphTargetGroups = computed(() => {
            if (!adminGraphData.value || !Array.isArray(adminGraphData.value.target_group_top_scam_types)) {
                return [];
            }
            return adminGraphData.value.target_group_top_scam_types;
        });

        const renderAdminTargetGroupBarChart = () => {
            if (!adminTargetGroupChartData.value) return;
            if (typeof window.echarts === 'undefined') {
                console.warn('ECharts 未加载，已跳过人群案件柱状图渲染。');
                return;
            }

            const items = Array.isArray(adminTargetGroupChartData.value.top_scam_types)
                ? adminTargetGroupChartData.value.top_scam_types
                : [];
            
            // Calculate max value for scaling to fill space
            const rawScores = items.map(item => Number((Number(item.score || 0) * 100).toFixed(2)));
            const maxScore = rawScores.length > 0 ? Math.max(...rawScores) : 100;
            const displayMax = Math.ceil(maxScore * 1.15 / 10) * 10; // Add some margin and round up to nearest 10

            const targetDom = document.getElementById('adminTargetGroupBarChart');
            if (!targetDom) return;

            // Dispose old instance if exists
            if (adminTargetGroupBarChart) {
                adminTargetGroupBarChart.dispose();
            }

            adminTargetGroupBarChart = echarts.init(targetDom);
            
            const option = {
                tooltip: {
                    trigger: 'axis',
                    axisPointer: { type: 'shadow' },
                    backgroundColor: 'rgba(15, 23, 42, 0.9)',
                    borderColor: 'rgba(255, 255, 255, 0.1)',
                    borderWidth: 1,
                    textStyle: { color: '#fff', fontSize: 12 },
                    formatter: (params) => {
                        const data = params.find(p => p.seriesName === '案件占比');
                        if (!data) return '';
                        return `<div class="font-bold mb-1">${data.name}</div>
                                <div class="flex items-center gap-2">
                                    <span class="w-2 h-2 rounded-full" style="background:${data.color}"></span>
                                    <span>占比: ${data.value}%</span>
                                </div>`;
                    }
                },
                grid: {
                    left: '3%',
                    right: '15%', // Leave space for labels
                    bottom: '3%',
                    top: '3%',
                    containLabel: true
                },
                xAxis: {
                    type: 'value',
                    max: displayMax, // Dynamic max to fill horizontal space
                    axisLabel: {
                        formatter: '{value}%',
                        color: '#64748b',
                        fontWeight: 'bold'
                    },
                    splitLine: {
                        lineStyle: { color: 'rgba(148, 163, 184, 0.1)' }
                    }
                },
                yAxis: {
                    type: 'category',
                    data: items.map(item => item.scam_type).reverse(),
                    axisLabel: {
                        color: '#334155',
                        fontWeight: 'bold',
                        fontSize: 12
                    },
                    axisLine: { show: false },
                    axisTick: { show: false }
                },
                series: [
                    {
                        name: 'placeholder',
                        type: 'bar',
                        itemStyle: {
                            color: 'rgba(148, 163, 184, 0.05)',
                            borderRadius: [0, 20, 20, 0]
                        },
                        barGap: '-100%',
                        barWidth: 32,
                        data: items.map(() => displayMax), 
                        animation: false,
                        tooltip: { show: false }
                    },
                    {
                        name: '案件占比',
                        type: 'bar',
                        data: items.map((item, index) => {
                            // Define a set of beautiful gradients
                            const gradients = [
                                ['#3b82f6', '#6366f1', '#d946ef'], // Blue-Indigo-Purple
                                ['#10b981', '#3b82f6', '#6366f1'], // Emerald-Blue-Indigo
                                ['#f59e0b', '#ef4444', '#d946ef'], // Amber-Red-Purple
                                ['#6366f1', '#a855f7', '#ec4899'], // Indigo-Purple-Pink
                                ['#0ea5e9', '#2dd4bf', '#10b981']  // Sky-Teal-Emerald
                            ];
                            const colors = gradients[index % gradients.length];
                            
                            return {
                                value: Number((Number(item.score || 0) * 100).toFixed(2)),
                                itemStyle: {
                                    borderRadius: [0, 20, 20, 0],
                                    color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
                                        { offset: 0, color: colors[0] },
                                        { offset: 0.5, color: colors[1] },
                                        { offset: 1, color: colors[2] }
                                    ])
                                }
                            };
                        }).reverse(),
                        barWidth: 32,
                        label: {
                            show: true,
                            position: 'right',
                            formatter: '{c}%',
                            color: '#475569',
                            fontWeight: 'bold',
                            distance: 10
                        },
                        emphasis: {
                            itemStyle: {
                                shadowBlur: 15,
                                shadowColor: 'rgba(0,0,0,0.1)',
                                shadowOffsetX: 5
                            }
                        }
                    }
                ]
            };

            adminTargetGroupBarChart.setOption(option);
        };

        const renderAdminCharts = () => {
            if (!adminStatsData.value) return;
            if (typeof window.echarts === 'undefined') {
                console.warn('ECharts 未加载，已跳过管理侧图表渲染。');
                return;
            }
            const sparseTrend = adminStatsData.value.trend;
            const trend = fillTrendGaps(sparseTrend, adminStatsInterval.value);
            const { by_scam_type, by_target_group } = adminStatsData.value;

            // Dispose old instances
            if (adminTrendChart && typeof adminTrendChart.dispose === 'function') adminTrendChart.dispose();
            if (adminTypeChart && typeof adminTypeChart.dispose === 'function') adminTypeChart.dispose();
            if (adminTargetChart && typeof adminTargetChart.dispose === 'function') adminTargetChart.dispose();

            // 1. Trend Line Chart
            const trendDom = document.getElementById('adminTrendChart');
            if (trendDom) {
                adminTrendChart = echarts.init(trendDom);
                adminTrendChart.setOption({
                    tooltip: {
                        trigger: 'axis',
                        backgroundColor: 'rgba(15, 23, 42, 0.9)',
                        textStyle: { color: '#fff' }
                    },
                    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
                    xAxis: {
                        type: 'category',
                        boundaryGap: false,
                        data: trend.map(item => formatAdminChartLabel(item.time_bucket)),
                        axisLabel: { color: '#64748b' }
                    },
                    yAxis: {
                        type: 'value',
                        axisLabel: { color: '#64748b' },
                        splitLine: { lineStyle: { type: 'dashed', color: 'rgba(148, 163, 184, 0.1)' } }
                    },
                    series: [{
                        name: '新增案件数',
                        type: 'line',
                        smooth: true,
                        data: trend.map(item => item.count),
                        symbol: 'circle',
                        symbolSize: 8,
                        itemStyle: { color: '#6366f1' },
                        areaStyle: {
                            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                                { offset: 0, color: 'rgba(99, 102, 241, 0.3)' },
                                { offset: 1, color: 'rgba(99, 102, 241, 0)' }
                            ])
                        },
                        lineStyle: { width: 3 }
                    }]
                });
            }

            // 2. Scam Type Pie Chart
            const typeDom = document.getElementById('adminTypeChart');
            if (typeDom) {
                adminTypeChart = echarts.init(typeDom);
                adminTypeChart.setOption({
                    tooltip: { trigger: 'item', backgroundColor: 'rgba(15, 23, 42, 0.9)', textStyle: { color: '#fff' } },
                    legend: { orient: 'vertical', right: 10, top: 'center', itemWidth: 10, itemHeight: 10, textStyle: { fontSize: 11, color: '#64748b' } },
                    series: [{
                        name: '诈骗类型',
                        type: 'pie',
                        radius: ['40%', '70%'],
                        center: ['40%', '50%'],
                        avoidLabelOverlap: false,
                        itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
                        label: { show: false },
                        emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
                        data: by_scam_type.map(i => ({ value: i.count, name: i.name }))
                    }]
                });
            }

            // 3. Target Group Pie Chart
            const targetDom = document.getElementById('adminTargetChart');
            if (targetDom) {
                adminTargetChart = echarts.init(targetDom);
                adminTargetChart.setOption({
                    tooltip: { trigger: 'item', backgroundColor: 'rgba(15, 23, 42, 0.9)', textStyle: { color: '#fff' } },
                    legend: { orient: 'vertical', right: 10, top: 'center', itemWidth: 10, itemHeight: 10, textStyle: { fontSize: 11, color: '#64748b' } },
                    series: [{
                        name: '目标人群',
                        type: 'pie',
                        radius: '70%',
                        center: ['40%', '50%'],
                        itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
                        label: { show: false },
                        data: by_target_group.map(i => ({ value: i.count, name: i.name }))
                    }]
                });
            }
        };

        watch(activeTab, (newTab) => {
            if (newTab === 'case_review') {
                fetchPendingReviews();
            }
            if (newTab === 'case_library') {
                fetchCaseLibrary();
                fetchCaseOptionLists();
            }
            if (newTab === 'family') {
                fetchFamilyOverview().then(() => {
                    if (familyHasGroup.value) {
                        connectFamilyNotificationWebSocket();
                    } else {
                        fetchReceivedFamilyInvitations();
                    }
                });
            }
            if (newTab === 'users') fetchUsers();
            if (newTab === 'history') fetchHistory();
            if (newTab === 'tasks') fetchTasks();
            if (newTab === 'risk_trend') fetchRiskTrend();
            if (newTab === 'admin_stats') fetchAdminStats();
        });

        // Polling
        let pollInterval;
        const startPolling = () => {
            fetchTasks({ silent: true });
            fetchHistory({ silent: true });
            fetchFamilyOverview({ silent: true });
            connectAlertWebSocket();
            connectFamilyNotificationWebSocket();
            if (pollInterval) clearInterval(pollInterval);
            pollInterval = setInterval(() => {
                if (isAuthenticated.value && activeTab.value === 'tasks') fetchTasks({ silent: true });
                if (isAuthenticated.value && activeTab.value === 'family') {
                    fetchFamilyOverview({ silent: true }).then(() => {
                        if (!familyHasGroup.value) {
                            fetchReceivedFamilyInvitations({ silent: true });
                        }
                    });
                }
            }, 5000);
        };
        
        const stopPolling = () => {
            if (pollInterval) clearInterval(pollInterval);
            disconnectAlertWebSocket();
            disconnectFamilyNotificationWebSocket();
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

        // Init
        onMounted(() => {
            fetchCaptcha();
            if (token.value) {
                getUserInfo();
            }
            initChatPosition();
            window.addEventListener('resize', handleResize);
            window.addEventListener('resize', () => {
                if (adminTargetGroupBarChart && typeof adminTargetGroupBarChart.resize === 'function') adminTargetGroupBarChart.resize();
                if (adminTrendChart && typeof adminTrendChart.resize === 'function') adminTrendChart.resize();
                if (adminTypeChart && typeof adminTypeChart.resize === 'function') adminTypeChart.resize();
                if (adminTargetChart && typeof adminTargetChart.resize === 'function') adminTargetChart.resize();
                if (pieChartInstance && typeof pieChartInstance.resize === 'function') pieChartInstance.resize();
                if (lineChartInstance && typeof lineChartInstance.resize === 'function') lineChartInstance.resize();
            });
        });

        onUnmounted(() => {
            stopPolling();
            clearSMSCodeCooldownTimer();
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
                'processing': 'bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-xs font-bold',
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
            { type: 'ai', content: '你好！我是你的反诈骗智能助手。我可以帮你分析风险、解答疑问，或者总结最近的安全情况。' }
        ]);
        const chatInput = ref('');
        const chatImages = ref([]);
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
                                    const toolName = call.name || call.function?.name || 'unknown';
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
                            const imageUrls = Array.isArray(msg.image_urls)
                                ? msg.image_urls.filter(item => typeof item === 'string' && item.trim())
                                : [];
                            if (!msg.content && imageUrls.length === 0) {
                                continue;
                            }
                            history.push({
                                type: 'user',
                                content: msg.content || '',
                                images: imageUrls
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

        const extractAttackSteps = (reportText) => {
            if (!reportText) return [];
            const reportSections = parseReport(reportText);
            const attackSection = reportSections.find((section) => {
                const title = String(section?.title || '').trim();
                return title.includes('诈骗链路还原');
            });
            if (!attackSection || !attackSection.content) return [];

            const stepLines = attackSection.content
                .split('\n')
                .map((line) => line.trim())
                .filter((line) => line !== '')
                .map((line) => line.replace(/^[-*•]\s+/, '').replace(/^\d+[.)、]\s*/, '').trim())
                .filter((line) => line !== '');

            return stepLines;
        };

        const extractScamKeywordSentences = (reportText) => {
            if (!reportText) return [];
            const reportSections = parseReport(reportText);
            const keywordSection = reportSections.find((section) => {
                const title = String(section?.title || '').trim();
                return title.includes('诈骗关键词句');
            });
            if (!keywordSection || !keywordSection.content) return [];

            const keywords = keywordSection.content
                .split('\n')
                .map((line) => line.trim())
                .filter((line) => line !== '')
                .map((line) => line.replace(/^[-*•]\s+/, '').replace(/^\d+[.)、]\s*/, '').trim())
                .filter((line) => line !== '');

            return keywords;
        };

        const parseRiskSummary = (raw) => {
            if (!raw || !String(raw).trim()) return null;
            try {
                const parsed = JSON.parse(raw);
                return parsed && typeof parsed === 'object' ? parsed : null;
            } catch (e) {
                return null;
            }
        };

        const parseInsight = (text) => {
            if (!text) return [];
            const sections = [];
            // Regex to match titles like 【整体视觉感受】 or [关键信息提取]
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
                        content: match[2] ? match[2] + '\n' : ''
                    };
                } else if (currentSection) {
                    currentSection.content += line + '\n';
                } else if (trimmedLine) {
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
                const results = await Promise.all(files.map(file => fileToBase64(file)));
                chatImages.value = [...chatImages.value, ...results];
                showToast(`已添加 ${files.length} 张图片`);
            } catch (e) {
                console.error('Read chat images failed:', e);
                showToast('图片读取失败', 'error');
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
                const response = await fetch('/api/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token.value}`
                    },
                    body: JSON.stringify({ message, images })
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
                chatInput.value = '';
                chatImages.value = [];
                chatHistoryLoaded.value = true;
                showToast('对话历史已重置');
            } catch (e) {
                showToast('娓呴櫎鍘嗗彶澶辫触', 'error');
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
                content += `**诈骗类型**: ${task.scam_type || '未识别'}\n`;
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
            isAuthenticated, user, authMode, loginMethod, form, ageForm, profileForm, occupationOptions, profileSaving, analyzeForm, 
            captchaImage, requiresGraphCaptcha, shouldShowSMSCodeSection, authSubmitLabel, smsCodeButtonText, canSendSMSCode, demoSMSCode,
            fetchCaptcha, sendSMSCode, handleAuth, logout, loading,
            activeTab, tasks, history, users, selectedTask, userSearch, toasts, analyzing,
            deletingHistory, handleFileSelect, submitAnalysis, viewTaskDetail, viewHistoryDetail, deleteHistoryCase, debouncedFetchUsers,
            formatTime, getStatusLabel, getStatusClass, normalizeRiskLevelText, getRiskClass, getAlertSeverityTheme,
            updateAge, updateUserProfile, deleteAccount, upgradeAccount, inviteCode, openImage, exportData, printReport,
            getUserDisplayName, getUserEmailText, getUserPhoneText, getUserAvatarText,
            familyOverview, familyMembers, familyInvitations, familyReceivedInvitations, familyGuardianLinks, familyNotifications,
            familyLoading, familyReceivedLoading, familyNotificationConnectionStatus, familyNotificationConnectionLabel, familyAlertModalVisible, activeFamilyNotification, familyCreateForm, familyInviteForm, familyAcceptForm, familyGuardianForm,
            familyUnreadCount, familyHasGroup, familyGuardianCandidates, familyProtectedCandidates,
            createFamily, createFamilyInvitation, acceptFamilyInvitation, acceptReceivedFamilyInvitation, fetchReceivedFamilyInvitations, createGuardianLink, deleteFamilyMember, deleteGuardianLink, markFamilyNotificationRead, acknowledgeFamilyAlert, openFamilyNotificationCenter,
            familyDeletingMembers, familyDeletingGuardianLinks, familyAcceptingInvitations, familyMarkingNotifications,
            showChat, chatMessages, chatInput, chatImages, isChatting, toggleChat, sendChatMessage, clearChatHistory,
            triggerChatImagePicker, handleChatImageSelect, removeChatImage,
            chatPosition, startDrag, // Export drag handler and state
            isSidebarCollapsed, toggleSidebar,
            parseReport, extractAttackSteps, extractScamKeywordSentences, parseRiskSummary, parseInsight,
            caseLibrary, scamTypeOptions, targetGroupOptions, selectedCase, showCaseModal, submittingCase, caseForm, submitCase, openCaseModal, fetchCaseLibrary, viewCaseDetail, deleteCase,
            pendingReviews, selectedReview, showReviewDetailModal, fetchPendingReviews, viewReviewDetail, approveReview,
            caseCollectionForm, startingCaseCollection, submitCaseCollection,
            riskInterval, fetchRiskTrend, riskData, getRiskTrendAnalysisClass,
            adminStatsInterval, fetchAdminStats, adminStatsData, adminGraphData, formatGraphScore,
            showGraphModal, openGraphModal, resetGraphZoom,
            adminTargetGroupChartData, availableGraphTargetGroups, selectedGraphTargetGroup,
            fetchAdminTargetGroupChart, clearAdminTargetGroupFocus,
            alertEvents, alertUnreadCount, alertModalVisible, activeAlertEvent, alertConnectionStatus, alertConnectionLabel,
            alertDrawerVisible, recentRiskAlerts, toggleAlertDrawer, closeAlertDrawer, openAlertCaseDetail,
            acknowledgeActiveAlert, openAlertHistory
        };
    }
}).mount('#app');
