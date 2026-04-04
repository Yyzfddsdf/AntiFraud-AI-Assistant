package com.example.myapplication

import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.IBinder
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import okhttp3.WebSocket

class BackgroundAlertService : Service() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private lateinit var repository: SentinelRepository
    private lateinit var preferences: android.content.SharedPreferences
    private lateinit var deduplicator: AlertDeduplicator

    private var alertSocket: WebSocket? = null
    private var familySocket: WebSocket? = null
    private var alertReconnectJob: Job? = null
    private var familyReconnectJob: Job? = null

    override fun onCreate() {
        super.onCreate()
        repository = SentinelRepository(AppConfigLoader.load(this).apiOrigin)
        preferences = getSharedPreferences("sentinel", Context.MODE_PRIVATE)
        deduplicator = AlertDeduplicator(this)
        AppNotifier.ensureChannels(this)
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_STOP -> {
                stopSelf()
                return START_NOT_STICKY
            }
            ACTION_REFRESH, ACTION_START, null -> {
                startForeground(AppNotifier.serviceNotificationId(), AppNotifier.buildServiceNotification(this))
                reconnectAll()
            }
        }
        return START_STICKY
    }

    override fun onTaskRemoved(rootIntent: Intent?) {
        val restartIntent = Intent(applicationContext, BackgroundAlertService::class.java).apply {
            action = ACTION_START
        }
        val pendingIntent = android.app.PendingIntent.getService(
            applicationContext,
            99,
            restartIntent,
            android.app.PendingIntent.FLAG_ONE_SHOT or android.app.PendingIntent.FLAG_IMMUTABLE,
        )
        val alarmManager = getSystemService(Context.ALARM_SERVICE) as android.app.AlarmManager
        alarmManager.setExactAndAllowWhileIdle(
            android.app.AlarmManager.RTC_WAKEUP,
            System.currentTimeMillis() + 1_000L,
            pendingIntent,
        )
        super.onTaskRemoved(rootIntent)
    }

    override fun onDestroy() {
        disconnectAll()
        repository.close()
        scope.cancel()
        super.onDestroy()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    private fun reconnectAll() {
        disconnectAll()
        val token = preferences.getString(PREF_TOKEN, "").orEmpty()
        if (token.isBlank()) {
            stopSelf()
            return
        }
        connectAlertSocket(token)
        if (preferences.getBoolean(PREF_HAS_FAMILY, false)) {
            connectFamilySocket(token)
        }
    }

    private fun disconnectAll() {
        alertReconnectJob?.cancel()
        familyReconnectJob?.cancel()
        alertSocket?.close(1000, "service refresh")
        familySocket?.close(1000, "service refresh")
        alertSocket = null
        familySocket = null
    }

    private fun connectAlertSocket(token: String) {
        alertSocket = repository.connectAlertSocket(
            token = token,
            onStateChange = { state ->
                if (state == RealtimeState.Reconnecting) scheduleAlertReconnect()
            },
            onMessage = { event ->
                if (event.type != "risk_alert" && event.type != "high_risk_alert") return@connectAlertSocket
                if (AppVisibilityTracker.isForeground) return@connectAlertSocket
                if (deduplicator.wasAlertNotified(event.record_id)) return@connectAlertSocket
                val riskLevel = normalizeRiskLevel(event.risk_level)
                AppNotifier.showRiskAlert(
                    context = this,
                    title = event.title.ifBlank { if (riskLevel == "高") "高风险预警" else "中风险提醒" },
                    body = event.case_summary.ifBlank { if (riskLevel == "高") "检测到新的高风险事件，请尽快查看。" else "检测到新的中风险事件，请留意核查。" },
                    notificationId = event.record_id.hashCode(),
                    riskLevel = riskLevel,
                )
                deduplicator.markAlertNotified(event.record_id)
            },
        )
    }

    private fun connectFamilySocket(token: String) {
        familySocket = repository.connectFamilyNotificationSocket(
            token = token,
            onStateChange = { state ->
                if (state == RealtimeState.Reconnecting) scheduleFamilyReconnect()
            },
            onMessage = { event ->
                if (event.type != "family_high_risk_alert") return@connectFamilyNotificationSocket
                if (AppVisibilityTracker.isForeground) return@connectFamilyNotificationSocket
                if (deduplicator.wasFamilyNotified(event.notification_id)) return@connectFamilyNotificationSocket
                AppNotifier.showRiskAlert(
                    context = this,
                    title = event.title.ifBlank { "家庭高风险提醒" },
                    body = event.summary.ifBlank { event.case_summary.ifBlank { "家庭成员触发新的高风险提醒，请尽快查看。" } },
                    notificationId = 20_000 + event.notification_id,
                )
                deduplicator.markFamilyNotified(event.notification_id)
            },
        )
    }

    private fun scheduleAlertReconnect() {
        if (alertReconnectJob?.isActive == true) return
        alertReconnectJob = scope.launch {
            delay(5_000L)
            val token = preferences.getString(PREF_TOKEN, "").orEmpty()
            if (token.isNotBlank()) connectAlertSocket(token)
        }
    }

    private fun scheduleFamilyReconnect() {
        if (familyReconnectJob?.isActive == true) return
        familyReconnectJob = scope.launch {
            delay(5_000L)
            val token = preferences.getString(PREF_TOKEN, "").orEmpty()
            if (token.isNotBlank() && preferences.getBoolean(PREF_HAS_FAMILY, false)) {
                connectFamilySocket(token)
            }
        }
    }

    companion object {
        private const val ACTION_START = "com.example.myapplication.action.START_ALERT_SERVICE"
        private const val ACTION_STOP = "com.example.myapplication.action.STOP_ALERT_SERVICE"
        private const val ACTION_REFRESH = "com.example.myapplication.action.REFRESH_ALERT_SERVICE"
        const val PREF_TOKEN = "token"
        const val PREF_HAS_FAMILY = "has_family_group"

        fun start(context: Context) {
            val intent = Intent(context, BackgroundAlertService::class.java).apply {
                action = ACTION_START
            }
            androidx.core.content.ContextCompat.startForegroundService(context, intent)
        }

        fun stop(context: Context) {
            val intent = Intent(context, BackgroundAlertService::class.java).apply {
                action = ACTION_STOP
            }
            context.startService(intent)
        }

        fun refresh(context: Context) {
            val intent = Intent(context, BackgroundAlertService::class.java).apply {
                action = ACTION_REFRESH
            }
            androidx.core.content.ContextCompat.startForegroundService(context, intent)
        }
    }

    private fun normalizeRiskLevel(value: String): String = when {
        value.contains("高") -> "高"
        value.contains("中") -> "中"
        else -> "高"
    }
}
