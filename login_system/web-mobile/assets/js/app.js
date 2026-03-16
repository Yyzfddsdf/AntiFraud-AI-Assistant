const { createApp, ref, reactive, onMounted, onUnmounted, computed, watch, nextTick } = Vue;

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
        const ageEditorVisible = ref(false);
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
        const deletingHistory = reactive({});
        const familyDeletingMembers = reactive({});
        const familyDeletingGuardianLinks = reactive({});
        const familyAcceptingInvitations = reactive({});
        const familyMarkingNotifications = reactive({});
        const selectedTask = ref(null);
        
        // Risk Trend State
        const riskInterval = ref('day');
        const riskData = ref(null);
        // Cache for risk trend data: { 'day': data, 'week': data, 'month': data }
        const riskCache = reactive({});
        let pieChartInstance = null;
        let lineChartInstance = null;
        let lineDetailChartInstance = null;
        let hasWarnedMissingChartLibrary = false;
        const openDropdownKey = ref('');

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
                    barClass: 'bg-red-500',
                    badgeClass: 'bg-red-50 text-red-600',
                    modalBadgeClass: 'bg-red-100 text-red-600',
                    panelClass: 'bg-red-50',
                    actionClass: 'bg-red-600',
                };
            }
            return {
                barClass: 'bg-amber-500',
                badgeClass: 'bg-amber-50 text-amber-700',
                modalBadgeClass: 'bg-amber-100 text-amber-700',
                panelClass: 'bg-amber-50',
                actionClass: 'bg-amber-500',
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
            activeAlertEvent.value = null;
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
            activeAlertEvent.value = null;
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
            activeFamilyNotification.value = null;
        };

        const openFamilyNotificationCenter = async () => {
            const current = activeFamilyNotification.value;
            familyAlertModalVisible.value = false;
            activeFamilyNotification.value = null;
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
                case_summary: String(payload.case_summary || '').trim(),
                summary: String(payload.summary || '').trim() || '家庭成员触发高风险案件，请及时核查。',
                scam_type: String(payload.scam_type || '').trim(),
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
        const familyRoleSelectOptions = [
            { value: 'member', label: '普通成员', hint: '接收家庭通知，不承担守护职责' },
            { value: 'guardian', label: '守护人', hint: '接收被守护成员的高风险提醒' }
        ];
        const familyGuardianSelectOptions = computed(() => (familyGuardianCandidates.value || []).map((item) => ({
            value: item.user_id,
            label: String(item.username || '').trim() || '未命名成员',
            hint: String(item.role || '').trim() || '守护人候选'
        })));
        const familyProtectedSelectOptions = computed(() => (familyProtectedCandidates.value || []).map((item) => ({
            value: item.user_id,
            label: String(item.username || '').trim() || '未命名成员',
            hint: String(item.relation || item.role || '').trim() || '被守护成员候选'
        })));
        const riskStatsSummary = computed(() => {
            const stats = riskData.value && riskData.value.stats ? riskData.value.stats : {};
            return {
                total: Number(stats.total) || 0,
                high: Number(stats.high) || 0,
                medium: Number(stats.medium) || 0,
                low: Number(stats.low) || 0
            };
        });
        const normalizeOptionValue = (value) => String(value ?? '').trim();
        const getSelectedOption = (options, value) => (options || []).find((option) => normalizeOptionValue(option.value) === normalizeOptionValue(value)) || null;
        const getSelectedOptionLabel = (options, value, placeholder = '请选择') => {
            const matched = getSelectedOption(options, value);
            return matched ? matched.label : placeholder;
        };
        const getSelectedOptionHint = (options, value, fallback = '') => {
            const matched = getSelectedOption(options, value);
            return matched ? String(matched.hint || '').trim() : fallback;
        };
        const toggleDropdown = (dropdownKey) => {
            openDropdownKey.value = openDropdownKey.value === dropdownKey ? '' : dropdownKey;
        };
        const closeDropdown = () => {
            openDropdownKey.value = '';
        };
        const selectDropdownValue = (dropdownKey, target, field, value) => {
            if (target && typeof target === 'object' && field) {
                target[field] = value;
            }
            openDropdownKey.value = '';
        };
        const handleDocumentClick = (event) => {
            const target = event && event.target;
            if (!(target instanceof Element)) {
                closeDropdown();
                return;
            }
            if (!target.closest('[data-custom-dropdown]')) {
                closeDropdown();
            }
        };

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
        const toggleAgeEditor = () => {
            ageEditorVisible.value = !ageEditorVisible.value;
        };
        const cancelAgeEditor = () => {
            ageEditorVisible.value = false;
            syncProfileForm(user.value || {});
        };
        const openProfilePrivacyPage = () => {
            ageEditorVisible.value = false;
            activeTab.value = 'profile_privacy';
        };
        const closeProfilePrivacyPage = () => {
            cancelAgeEditor();
            activeTab.value = 'profile';
        };

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
            chatMessages.value = [
                buildChatMessage({ type: 'ai', content: '你好！我是你的反诈骗智能助手。我可以帮你分析风险、解答疑问，或者总结最近的安全情况。' })
            ];
            chatInput.value = '';
            chatImages.value = [];
            chatHistoryLoaded.value = false;
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
                ageEditorVisible.value = false;
                showToast(res.message || '用户画像更新成功');
            }
        };
        const updateAge = updateUserProfile;

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
            pruneFamilyNotificationsByMembers(familyMembers.value);
        };

        const fetchFamilyOverview = async ({ silent = false } = {}) => {
            if (!isAuthenticated.value) return;
            familyLoading.value = !silent;
            const res = await request('/families/me', 'GET', null, { silent });
            familyLoading.value = false;
            if (res) {
                hydrateFamilyOverview(res);
                if (res.family) {
                    replaceListIfChanged(familyReceivedInvitations, []);
                    familyReceivedLoading.value = false;
                }
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
                replaceListIfChanged(familyReceivedInvitations, []);
                familyReceivedLoading.value = false;
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

        const acceptFamilyInvitation = async (rawInviteCode = '', invitationID = 0) => {
            const inviteCode = String(rawInviteCode || familyAcceptForm.invite_code || '').trim();
            if (!inviteCode) {
                showToast('请输入家庭邀请码', 'error');
                return;
            }

            if (invitationID) {
                familyAcceptingInvitations[invitationID] = true;
            }
            try {
                const payload = { invite_code: inviteCode };
                const res = await request('/families/invitations/accept', 'POST', payload);
                if (res) {
                    hydrateFamilyOverview(res);
                    familyAcceptForm.invite_code = '';
                    replaceListIfChanged(familyReceivedInvitations, []);
                    familyReceivedLoading.value = false;
                    showToast('已加入家庭');
                    familyNotifications.value = [];
                    familyNotificationSeenIDs.clear();
                    connectFamilyNotificationWebSocket();
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

        // Risk Trend Logic
        const fetchRiskTrend = async (forceRefresh = false) => {
            if (!isAuthenticated.value) return;

            riskInterval.value = 'day';
            const interval = 'day';
            
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
        const getRecentRiskTrendRows = (limit = 7) => {
            if (!riskData.value || !Array.isArray(riskData.value.trend)) return [];
            const normalizedLimit = Number(limit) > 0 ? Number(limit) : 7;
            return fillTrendGaps(riskData.value.trend, riskInterval.value)
                .slice(-normalizedLimit)
                .reverse();
        };

        const disposeUserRiskCharts = () => {
            if (pieChartInstance && typeof pieChartInstance.dispose === 'function') {
                pieChartInstance.dispose();
            }
            if (lineChartInstance && typeof lineChartInstance.dispose === 'function') {
                lineChartInstance.dispose();
            }
            if (lineDetailChartInstance && typeof lineDetailChartInstance.dispose === 'function') {
                lineDetailChartInstance.dispose();
            }
            pieChartInstance = null;
            lineChartInstance = null;
            lineDetailChartInstance = null;
        };

        const buildRiskLineOption = (trend, compact = false) => ({
            tooltip: {
                trigger: 'axis',
                backgroundColor: 'rgba(15, 23, 42, 0.9)',
                textStyle: { color: '#fff' }
            },
            legend: compact ? undefined : { bottom: 0, textStyle: { fontSize: 11 } },
            grid: compact
                ? { left: '2%', right: '2%', top: '8%', bottom: '6%', containLabel: false }
                : { left: '3%', right: '4%', top: '10%', bottom: '15%', containLabel: true },
            xAxis: {
                type: 'category',
                boundaryGap: false,
                data: trend.map(item => formatChartLabel(item.time_bucket)),
                axisLabel: compact ? { show: false } : { color: '#64748b' },
                axisLine: compact ? { show: false } : undefined,
                axisTick: compact ? { show: false } : undefined
            },
            yAxis: {
                type: 'value',
                axisLabel: compact ? { show: false } : { color: '#64748b' },
                axisLine: compact ? { show: false } : undefined,
                axisTick: compact ? { show: false } : undefined,
                splitLine: compact
                    ? { show: false }
                    : { lineStyle: { type: 'dashed', color: 'rgba(148, 163, 184, 0.1)' } }
            },
            series: [
                {
                    name: '高风险',
                    type: 'line',
                    smooth: true,
                    showSymbol: !compact,
                    symbolSize: compact ? 0 : 6,
                    data: trend.map(item => item.high),
                    itemStyle: { color: '#ef4444' },
                    lineStyle: { width: compact ? 2.5 : 3 },
                    areaStyle: compact ? { color: 'rgba(239, 68, 68, 0.08)' } : undefined
                },
                {
                    name: '中风险',
                    type: 'line',
                    smooth: true,
                    showSymbol: !compact,
                    symbolSize: compact ? 0 : 6,
                    data: trend.map(item => item.medium),
                    itemStyle: { color: '#f59e0b' },
                    lineStyle: { width: compact ? 2.5 : 3 },
                    areaStyle: compact ? { color: 'rgba(245, 158, 11, 0.08)' } : undefined
                },
                {
                    name: '低风险',
                    type: 'line',
                    smooth: true,
                    showSymbol: !compact,
                    symbolSize: compact ? 0 : 6,
                    data: trend.map(item => item.low),
                    itemStyle: { color: '#10b981' },
                    lineStyle: { width: compact ? 2.5 : 3 },
                    areaStyle: compact ? { color: 'rgba(16, 185, 129, 0.08)' } : undefined
                }
            ]
        });

        const renderCharts = () => {
            if (!riskData.value) {
                disposeUserRiskCharts();
                return;
            }

            if (typeof window.echarts === 'undefined') {
                if (!hasWarnedMissingChartLibrary) {
                    console.warn('ECharts 未加载，已跳过用户侧图表渲染。');
                    hasWarnedMissingChartLibrary = true;
                }
                return;
            }

            hasWarnedMissingChartLibrary = false;
            const stats = riskData.value.stats || { high: 0, medium: 0, low: 0, total: 0 };
            const trend = fillTrendGaps(riskData.value.trend, riskInterval.value);

            disposeUserRiskCharts();

            const pieDom = document.getElementById('riskPieChart');
            if (pieDom) {
                pieChartInstance = echarts.init(pieDom);
                pieChartInstance.setOption({
                    tooltip: {
                        trigger: 'item',
                        backgroundColor: 'rgba(15, 23, 42, 0.9)',
                        textStyle: { color: '#fff' }
                    },
                    graphic: [
                        {
                            type: 'text',
                            left: 'center',
                            top: '43%',
                            style: {
                                text: String(Number(stats.total) || 0),
                                fill: '#0f172a',
                                fontSize: 26,
                                fontWeight: '700'
                            }
                        },
                        {
                            type: 'text',
                            left: 'center',
                            top: '58%',
                            style: {
                                text: '总检测',
                                fill: '#64748b',
                                fontSize: 12
                            }
                        }
                    ],
                    series: [{
                        name: '风险分布',
                        type: 'pie',
                        radius: ['56%', '82%'],
                        avoidLabelOverlap: false,
                        itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
                        label: { show: false },
                        emphasis: { label: { show: true, fontSize: 13, fontWeight: 'bold' } },
                        data: [
                            { value: Number(stats.high) || 0, name: '高风险', itemStyle: { color: '#ef4444' } },
                            { value: Number(stats.medium) || 0, name: '中风险', itemStyle: { color: '#f59e0b' } },
                            { value: Number(stats.low) || 0, name: '低风险', itemStyle: { color: '#10b981' } }
                        ]
                    }]
                });
            }

            const lineDom = document.getElementById('riskLineChart');
            if (lineDom) {
                lineChartInstance = echarts.init(lineDom);
                lineChartInstance.setOption(buildRiskLineOption(trend, true));
            }

            const lineDetailDom = document.getElementById('riskLineChartDetail');
            if (lineDetailDom) {
                lineDetailChartInstance = echarts.init(lineDetailDom);
                lineDetailChartInstance.setOption(buildRiskLineOption(trend, false));
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
        const formatRiskTrendDescriptor = (trendText, dimension = 'overall') => {
            const normalized = String(trendText || '').trim();
            if (dimension === 'high') {
                if (normalized === '上升') return '高危暴露增强';
                if (normalized === '下降') return '高危暴露收敛';
                if (normalized === '平稳') return '高危信号平稳';
                return '高危信号待观察';
            }
            if (normalized === '上升') return '风险热度抬升';
            if (normalized === '下降') return '风险热度回落';
            if (normalized === '平稳') return '风险热度平稳';
            return '风险热度待观察';
        };
        const getRiskTrendHeadline = (analysis) => {
            const overall = String(analysis?.overall_trend || '').trim();
            const high = String(analysis?.high_risk_trend || '').trim();
            if (overall === '上升' && high === '上升') return 'AI 研判：风险热度正在拉升';
            if (overall === '下降' && high === '下降') return 'AI 研判：风险热度出现回落';
            if (high === '上升') return 'AI 研判：高危暴露正在增强';
            if (overall === '上升') return 'AI 研判：整体风险有所抬头';
            if (overall === '下降') return 'AI 研判：整体风险趋于缓和';
            if (overall === '平稳' && high === '平稳') return 'AI 研判：当前波动较为平稳';
            return 'AI 研判：风险走势仍需持续观察';
        };

        watch(activeTab, (newTab) => {
            closeDropdown();
            if (newTab === 'family') {
                fetchFamilyOverview().then(() => {
                    if (familyHasGroup.value) {
                        connectFamilyNotificationWebSocket();
                    } else {
                        fetchReceivedFamilyInvitations();
                    }
                });
            }
            if (newTab === 'chat') {
                if (!chatHistoryLoaded.value) {
                    fetchChatHistory();
                } else {
                    scrollToBottom();
                }
            }
            if (newTab === 'history') fetchHistory();
            if (newTab === 'tasks') {
                fetchTasks();
                fetchRiskTrend();
            }
            if (newTab === 'risk_trend') fetchRiskTrend();
        });

        // Polling
        let pollInterval;
        const startPolling = () => {
            fetchTasks({ silent: true });
            fetchHistory({ silent: true });
            fetchFamilyOverview({ silent: true });
            fetchRiskTrend();
            connectAlertWebSocket();
            connectFamilyNotificationWebSocket();
            if (pollInterval) clearInterval(pollInterval);
            pollInterval = setInterval(() => {
                if (isAuthenticated.value && activeTab.value === 'tasks') {
                    fetchTasks({ silent: true });
                    fetchRiskTrend();
                }
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

        const handleWindowResize = () => {
            handleResize();
            if (pieChartInstance && typeof pieChartInstance.resize === 'function') pieChartInstance.resize();
            if (lineChartInstance && typeof lineChartInstance.resize === 'function') lineChartInstance.resize();
            if (lineDetailChartInstance && typeof lineDetailChartInstance.resize === 'function') lineDetailChartInstance.resize();
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
            window.addEventListener('resize', handleWindowResize);
            document.addEventListener('click', handleDocumentClick);
            startBannerCarousel();
        });

        onUnmounted(() => {
            stopPolling();
            disposeUserRiskCharts();
            clearSMSCodeCooldownTimer();
            window.removeEventListener('resize', handleWindowResize);
            document.removeEventListener('click', handleDocumentClick);
            stopBannerCarousel();
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

        const escapeHtml = (text) => String(text || '')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');

        const renderPlainText = (text) => {
            const escaped = escapeHtml(text);
            return escaped.replace(/\r?\n/g, '<br>');
        };

        // Markdown Renderer
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
            } catch (e) {
                console.error('Markdown parse error:', e);
                return renderPlainText(text);
            }
        };

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
                rendered_content: normalizedType === 'ai' ? renderMarkdown(normalizedContent) : ''
            };
        };

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

            // Replace the whole message object so streaming updates always trigger markdown re-render.
            chatMessages.value.splice(index, 1, nextMessage);
            return index;
        };

        // Chat State
        const showChat = ref(false);
        const chatMessages = ref([
            buildChatMessage({ type: 'ai', content: '你好！我是你的反诈骗智能助手。我可以帮你分析风险、解答疑问，或者总结最近的安全情况。' })
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
                                    history.push(buildChatMessage({
                                        type: 'tool',
                                        content: `正在调用工具: ${toolName}...`
                                    }));
                                }
                            }
                            
                            // If there is content, add as AI message
                            if (msg.content) {
                                history.push(buildChatMessage({
                                    type: 'ai',
                                    content: msg.content
                                }));
                            }
                        } 
                        // Handle tool result messages
                        else if (msg.role === 'tool') {
                            // Optionally show tool completion
                            history.push(buildChatMessage({
                                type: 'tool',
                                content: `工具调用完成`
                            }));
                        }
                        // Handle user messages
                        else if (msg.role === 'user') {
                            const imageUrls = Array.isArray(msg.image_urls)
                                ? msg.image_urls.filter(item => typeof item === 'string' && item.trim())
                                : [];
                            if (!msg.content && imageUrls.length === 0) {
                                continue;
                            }
                            history.push(buildChatMessage({
                                type: 'user',
                                content: msg.content || '',
                                images: imageUrls
                            }));
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
            nextTick(() => {
                const container = document.getElementById('chat-container');
                if (!container) return;
                window.requestAnimationFrame(() => {
                    container.scrollTop = container.scrollHeight;
                });
            });
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
                        'Authorization': `Bearer ${token.value}`
                    },
                    body: JSON.stringify({ message, images })
                });

                if (!response.ok) throw new Error('Network response was not ok');

                const reader = response.body.getReader();
                const decoder = new TextDecoder();
                let aiMessageContent = '';
                
                // Add placeholder for AI response
                appendChatMessage({ type: 'ai', content: '' });
                
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
                                    return appendChatMessage({ type: 'ai', content: '' });
                                };

                                if (data.type === 'content') {
                                    const idx = getActiveAiIndex();
                                    aiMessageContent += data.content;
                                    replaceChatMessage(idx, { type: 'ai', content: aiMessageContent });
                                    scrollToBottom();
                                } else if (data.type === 'tool_call') {
                                    // Insert tool message
                                    const toolName = data.tool;
                                    // If current AI message is empty, replace it
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
                                    // Reset aiMessageContent for next segment
                                    aiMessageContent = '';
                                    scrollToBottom();
                                } else if (data.type === 'tool_result') {
                                    // Insert tool completion message
                                    const toolName = data.tool;
                                    appendChatMessage({
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
                appendChatMessage({ type: 'error', content: '抱歉，服务暂时不可用，请稍后再试。' });
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
                    buildChatMessage({ type: 'ai', content: '对话历史已清除。' })
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

        // Banner Carousel State
        const currentBannerIndex = ref(0);
        let bannerCarouselTimer = null;
        
        const startBannerCarousel = () => {
            bannerCarouselTimer = setInterval(() => {
                currentBannerIndex.value = (currentBannerIndex.value + 1) % 2;
            }, 5000);
        };
        
        const stopBannerCarousel = () => {
            if (bannerCarouselTimer) {
                clearInterval(bannerCarouselTimer);
                bannerCarouselTimer = null;
            }
        };

        return {
            isAuthenticated, user, authMode, loginMethod, form, ageForm, profileForm, occupationOptions, profileSaving, analyzeForm, 
            captchaImage, requiresGraphCaptcha, shouldShowSMSCodeSection, authSubmitLabel, smsCodeButtonText, canSendSMSCode, demoSMSCode,
            fetchCaptcha, sendSMSCode, handleAuth, logout, loading,
            activeTab, tasks, history, selectedTask, toasts, analyzing,
            deletingHistory, handleFileSelect, submitAnalysis, viewTaskDetail, viewHistoryDetail, deleteHistoryCase,
            formatTime, getStatusLabel, getStatusClass, normalizeRiskLevelText, getRiskClass, getAlertSeverityTheme, renderMarkdown,
            updateAge, updateUserProfile, deleteAccount, openImage, exportData, printReport,
            getUserDisplayName, getUserEmailText, getUserPhoneText, getUserAvatarText,
            ageEditorVisible, toggleAgeEditor, cancelAgeEditor, openProfilePrivacyPage, closeProfilePrivacyPage,
            familyOverview, familyMembers, familyInvitations, familyReceivedInvitations, familyGuardianLinks, familyNotifications,
            familyLoading, familyReceivedLoading, familyNotificationConnectionStatus, familyNotificationConnectionLabel, familyAlertModalVisible, activeFamilyNotification, familyCreateForm, familyInviteForm, familyAcceptForm, familyGuardianForm,
            familyUnreadCount, familyHasGroup, familyGuardianCandidates, familyProtectedCandidates,
            openDropdownKey, familyRoleSelectOptions, familyGuardianSelectOptions, familyProtectedSelectOptions,
            getSelectedOptionLabel, getSelectedOptionHint, toggleDropdown, closeDropdown, selectDropdownValue,
            createFamily, createFamilyInvitation, acceptFamilyInvitation, fetchReceivedFamilyInvitations, createGuardianLink, deleteFamilyMember, deleteGuardianLink, markFamilyNotificationRead, acknowledgeFamilyAlert, openFamilyNotificationCenter,
            familyDeletingMembers, familyDeletingGuardianLinks, familyAcceptingInvitations, familyMarkingNotifications,
            showChat, chatMessages, chatInput, chatImages, isChatting, toggleChat, sendChatMessage, clearChatHistory,
            triggerChatImagePicker, handleChatImageSelect, removeChatImage,
            chatPosition, startDrag, // Export drag handler and state
            isSidebarCollapsed, toggleSidebar,
            parseReport, extractAttackSteps, extractScamKeywordSentences, parseInsight,
            riskInterval, fetchRiskTrend, riskData, riskStatsSummary, getRiskTrendAnalysisClass, formatRiskTrendDescriptor, getRiskTrendHeadline, getRecentRiskTrendRows, formatChartLabel,
            alertEvents, alertUnreadCount, alertModalVisible, activeAlertEvent, alertConnectionStatus, alertConnectionLabel,
            alertDrawerVisible, recentRiskAlerts, toggleAlertDrawer, closeAlertDrawer, openAlertCaseDetail,
            acknowledgeActiveAlert, openAlertHistory,
            currentBannerIndex, startBannerCarousel, stopBannerCarousel
        };
    }
}).mount('#app');
