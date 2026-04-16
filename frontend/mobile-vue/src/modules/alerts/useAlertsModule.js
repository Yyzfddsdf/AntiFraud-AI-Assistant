import { computed, ref } from 'vue';
import { handleRealtimeHeartbeatMessage } from '../../app/realtimeHeartbeat.js';

export function useAlertsModule(deps) {
  const alertEvents = ref([]);
  const alertUnreadCount = ref(0);
  const alertModalVisible = ref(false);
  const activeAlertEvent = ref(null);
  const alertConnectionStatus = ref('disconnected');
  const alertDrawerVisible = ref(false);

  let alertSocket = null;
  let alertReconnectTimer = null;
  let alertReconnectAttempts = 0;
  const alertSeenRecordIDs = new Set();
  const maxAlertReconnectDelayMS = 30000;

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
        actionClass: 'bg-red-600'
      };
    }

    return {
      barClass: 'bg-amber-500',
      badgeClass: 'bg-amber-50 text-amber-700',
      modalBadgeClass: 'bg-amber-100 text-amber-700',
      panelClass: 'bg-amber-50',
      actionClass: 'bg-amber-500'
    };
  };

  const getAlertToastType = (level) => (normalizeAlertRiskLevel(level) === '高' ? 'error' : 'warning');

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
        .filter(Boolean)
    );

    const merged = new Map();
    const historyItems = Array.isArray(deps.history.value) ? deps.history.value : [];

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
      if (!recordID || !isWithinRecentAlertWindow(event.created_at)) continue;

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
      .sort((a, b) => new Date(b.created_at || b.sent_at || 0).getTime() - new Date(a.created_at || a.sent_at || 0).getTime())
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
      await deps.fetchHistory({ silent: true });
    }
  };

  const openAlertCaseDetail = async (item) => {
    if (!item || !item.record_id) return;
    markAlertReadByRecordID(item.record_id);
    alertDrawerVisible.value = false;
    alertModalVisible.value = false;
    activeAlertEvent.value = null;
    deps.activeTab.value = 'history';
    await deps.fetchHistory({ silent: true });
    await deps.viewTaskDetail(item.record_id);
  };

  const buildAlertWebSocketURL = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const base = `${protocol}://${window.location.host}/api/alert/ws`;
    const queryToken = encodeURIComponent(deps.token.value || '');
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

  const scheduleAlertReconnect = () => {
    if (!deps.isAuthenticated.value || !deps.token.value) return;
    if (alertReconnectTimer) return;

    const delay = Math.min(maxAlertReconnectDelayMS, 1000 * Math.pow(2, alertReconnectAttempts));
    alertReconnectAttempts += 1;
    alertReconnectTimer = setTimeout(() => {
      alertReconnectTimer = null;
      connectAlertWebSocket();
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

  const handleAlertMessage = (payload) => {
    if (!payload || !['risk_alert', 'high_risk_alert'].includes(String(payload.type || '').trim())) return;
    const recordID = String(payload.record_id || '').trim();
    if (!recordID || alertSeenRecordIDs.has(recordID)) return;

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
    deps.showToast(`${event.risk_level}风险预警：${event.title}`, getAlertToastType(event.risk_level));
    deps.fetchHistory({ silent: true });
  };

  const connectAlertWebSocket = () => {
    if (!deps.isAuthenticated.value || !deps.token.value) return;
    if (alertSocket && (alertSocket.readyState === WebSocket.OPEN || alertSocket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    let ws = null;
    try {
      alertConnectionStatus.value = 'connecting';
      ws = new WebSocket(buildAlertWebSocketURL());
    } catch {
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
        if (handleRealtimeHeartbeatMessage(ws, payload)) return;
        handleAlertMessage(payload);
      } catch {
        // ignore malformed payload
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
      if (!deps.isAuthenticated.value) {
        alertConnectionStatus.value = 'disconnected';
        return;
      }
      alertConnectionStatus.value = 'reconnecting';
      scheduleAlertReconnect();
    };
  };

  const resetAlertState = () => {
    alertDrawerVisible.value = false;
    alertEvents.value = [];
    alertUnreadCount.value = 0;
    alertModalVisible.value = false;
    activeAlertEvent.value = null;
    alertSeenRecordIDs.clear();
  };

  return {
    alertEvents,
    alertUnreadCount,
    alertModalVisible,
    activeAlertEvent,
    alertConnectionStatus,
    alertConnectionLabel,
    alertDrawerVisible,
    recentRiskAlerts,
    getAlertSeverityTheme,
    markAlertReadByRecordID,
    closeAlertDrawer,
    toggleAlertDrawer,
    openAlertCaseDetail,
    connectAlertWebSocket,
    disconnectAlertWebSocket,
    acknowledgeActiveAlert,
    openAlertHistory,
    resetAlertState
  };
}
