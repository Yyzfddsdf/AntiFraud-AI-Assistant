const HEARTBEAT_TYPE_PING = 'ping';
const HEARTBEAT_TYPE_PONG = 'pong';

export function handleRealtimeHeartbeatMessage(ws, payload) {
  const type = String(payload?.type || '').trim();
  if (type === HEARTBEAT_TYPE_PING) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      try {
        ws.send(JSON.stringify({
          type: HEARTBEAT_TYPE_PONG,
          sent_at: String(payload?.sent_at || '').trim(),
          received_at: new Date().toISOString()
        }));
      } catch {
        // ignore heartbeat reply failures and let reconnect logic recover
      }
    }
    return true;
  }

  return type === HEARTBEAT_TYPE_PONG;
}
