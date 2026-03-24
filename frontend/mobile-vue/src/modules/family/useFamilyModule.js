import { computed, reactive, ref } from 'vue';

export function useFamilyModule(deps) {
  const familyOverview = ref(null);
  const familyMembers = ref([]);
  const familyInvitations = ref([]);
  const familyReceivedInvitations = ref([]);
  const familyGuardianLinks = ref([]);
  const familyNotifications = ref([]);
  const familyLoading = ref(false);
  const familyReceivedLoading = ref(false);
  const familyNotificationConnectionStatus = ref('disconnected');
  const familyAlertModalVisible = ref(false);
  const activeFamilyNotification = ref(null);

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

  const familyDeletingMembers = reactive({});
  const familyDeletingGuardianLinks = reactive({});
  const familyAcceptingInvitations = reactive({});
  const familyMarkingNotifications = reactive({});

  let familyNotificationSocket = null;
  let familyNotificationReconnectTimer = null;
  let familyNotificationReconnectAttempts = 0;
  const familyNotificationSeenIDs = new Set();
  const maxFamilyNotificationReconnectDelayMS = 30000;

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

  const familyUnreadCount = computed(() => (familyNotifications.value || []).filter((item) => item && !item.read_at).length);
  const familyHasGroup = computed(() => !!familyOverview.value?.family);
  const familyGuardianCandidates = computed(() => (familyMembers.value || []).filter((item) => item && (item.role === 'owner' || item.role === 'guardian')));
  const familyProtectedCandidates = computed(() => (familyMembers.value || []).filter((item) => item && item.role !== 'owner'));
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

  const pruneFamilyNotificationsByMembers = (members) => {
    const activeUserIDs = new Set(
      (Array.isArray(members) ? members : [])
        .map((item) => Number(item?.user_id || 0))
        .filter((userID) => Number.isInteger(userID) && userID > 0)
    );

    if (activeUserIDs.size === 0) {
      familyNotifications.value = [];
      familyNotificationSeenIDs.clear();
      activeFamilyNotification.value = null;
      familyAlertModalVisible.value = false;
      return;
    }

    familyNotifications.value = familyNotifications.value.filter((item) => {
      const targetUserID = Number(item?.target_user_id || 0);
      return Number.isInteger(targetUserID) && activeUserIDs.has(targetUserID);
    });

    familyNotificationSeenIDs.clear();
    familyNotifications.value.forEach((item) => {
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
    if (!deps.isAuthenticated.value) return;
    familyLoading.value = !silent;
    const res = await deps.request('/families/me', 'GET', null, { silent });
    familyLoading.value = false;
    if (res) {
      hydrateFamilyOverview(res);
      if (res.family) {
        deps.replaceListIfChanged(familyReceivedInvitations, []);
        familyReceivedLoading.value = false;
      }
    }
  };

  const fetchReceivedFamilyInvitations = async ({ silent = false } = {}) => {
    if (!deps.isAuthenticated.value) return;
    familyReceivedLoading.value = !silent;
    const res = await deps.request('/families/invitations/received', 'GET', null, { silent });
    familyReceivedLoading.value = false;
    if (res && Array.isArray(res.invitations)) {
      deps.replaceListIfChanged(familyReceivedInvitations, res.invitations);
    }
  };

  const buildFamilyNotificationWebSocketURL = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const base = `${protocol}://${window.location.host}/api/families/notifications/ws`;
    const queryToken = encodeURIComponent(deps.token.value || '');
    return `${base}?token=${queryToken}`;
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

  const scheduleFamilyNotificationReconnect = () => {
    if (!deps.isAuthenticated.value || !deps.token.value) return;
    if (familyNotificationReconnectTimer) return;

    const delay = Math.min(maxFamilyNotificationReconnectDelayMS, 1000 * Math.pow(2, familyNotificationReconnectAttempts));
    familyNotificationReconnectAttempts += 1;
    familyNotificationReconnectTimer = setTimeout(() => {
      familyNotificationReconnectTimer = null;
      connectFamilyNotificationWebSocket();
    }, delay);
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
    deps.showToast(`家庭通知：${notification.summary}`, 'error');
  };

  const connectFamilyNotificationWebSocket = () => {
    if (!deps.isAuthenticated.value || !deps.token.value) return;
    if (familyNotificationSocket && (familyNotificationSocket.readyState === WebSocket.OPEN || familyNotificationSocket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    let ws = null;
    try {
      familyNotificationConnectionStatus.value = 'connecting';
      ws = new WebSocket(buildFamilyNotificationWebSocketURL());
    } catch {
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
        handleFamilyNotificationMessage(JSON.parse(event.data));
      } catch {
        // ignore malformed payload
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
      if (!deps.isAuthenticated.value) {
        familyNotificationConnectionStatus.value = 'disconnected';
        return;
      }
      familyNotificationConnectionStatus.value = 'reconnecting';
      scheduleFamilyNotificationReconnect();
    };
  };

  const createFamily = async () => {
    const payload = { name: familyCreateForm.name.trim() };
    const res = await deps.request('/families', 'POST', payload);
    if (res) {
      hydrateFamilyOverview(res);
      familyCreateForm.name = '';
      deps.replaceListIfChanged(familyReceivedInvitations, []);
      familyReceivedLoading.value = false;
      deps.showToast('家庭创建成功');
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
    const res = await deps.request('/families/invitations', 'POST', payload);
    if (res && res.invitation) {
      deps.showToast('家庭邀请已创建');
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
      deps.showToast('请输入家庭邀请码', 'error');
      return;
    }

    if (invitationID) {
      familyAcceptingInvitations[invitationID] = true;
    }

    try {
      const res = await deps.request('/families/invitations/accept', 'POST', { invite_code: inviteCode });
      if (res) {
        hydrateFamilyOverview(res);
        familyAcceptForm.invite_code = '';
        deps.replaceListIfChanged(familyReceivedInvitations, []);
        familyReceivedLoading.value = false;
        deps.showToast('已加入家庭');
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
      deps.showToast('请选择守护人和被守护成员', 'error');
      return;
    }

    const res = await deps.request('/families/guardian-links', 'POST', {
      guardian_user_id: guardianUserID,
      member_user_id: memberUserID
    });

    if (res && res.guardian_link) {
      deps.showToast('守护关系配置成功');
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
      const res = await deps.request(`/families/members/${encodeURIComponent(member.member_id)}`, 'DELETE');
      if (res) {
        deps.showToast(res.message || '成员已移除');
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
      const res = await deps.request(`/families/guardian-links/${encodeURIComponent(link.id)}`, 'DELETE');
      if (res) {
        deps.showToast(res.message || '守护关系已移除');
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
      const res = await deps.request(`/families/notifications/${encodeURIComponent(notification.id)}/read`, 'POST');
      if (res) {
        const readAt = new Date().toISOString();
        familyNotifications.value = familyNotifications.value.map((item) => (
          item && item.id === notification.id ? { ...item, read_at: readAt } : item
        ));
      }
    } finally {
      familyMarkingNotifications[notification.id] = false;
    }
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
    deps.activeTab.value = 'family';
    if (current) {
      await markFamilyNotificationRead(current);
    }
    await fetchFamilyOverview({ silent: true });
    connectFamilyNotificationWebSocket();
  };

  const resetFamilyState = () => {
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

  return {
    familyOverview,
    familyMembers,
    familyInvitations,
    familyReceivedInvitations,
    familyGuardianLinks,
    familyNotifications,
    familyLoading,
    familyReceivedLoading,
    familyNotificationConnectionStatus,
    familyNotificationConnectionLabel,
    familyAlertModalVisible,
    activeFamilyNotification,
    familyCreateForm,
    familyInviteForm,
    familyAcceptForm,
    familyGuardianForm,
    familyDeletingMembers,
    familyDeletingGuardianLinks,
    familyAcceptingInvitations,
    familyMarkingNotifications,
    familyUnreadCount,
    familyHasGroup,
    familyGuardianCandidates,
    familyProtectedCandidates,
    familyRoleSelectOptions,
    familyGuardianSelectOptions,
    familyProtectedSelectOptions,
    fetchFamilyOverview,
    fetchReceivedFamilyInvitations,
    connectFamilyNotificationWebSocket,
    disconnectFamilyNotificationWebSocket,
    createFamily,
    createFamilyInvitation,
    acceptFamilyInvitation,
    createGuardianLink,
    deleteFamilyMember,
    deleteGuardianLink,
    markFamilyNotificationRead,
    acknowledgeFamilyAlert,
    openFamilyNotificationCenter,
    resetFamilyState
  };
}
