package com.example.myapplication

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat

object AppNotifier {
    const val ALERT_CHANNEL_ID = "sentinel_alerts"
    private const val SERVICE_CHANNEL_ID = "sentinel_service"
    private const val QUICK_ANALYZE_SERVICE_CHANNEL_ID = "sentinel_quick_analyze_service"
    private const val ALERT_GROUP = "sentinel_alert_group"
    private const val SERVICE_NOTIFICATION_ID = 1001
    private const val QUICK_ANALYZE_SERVICE_NOTIFICATION_ID = 1002

    fun ensureChannels(context: Context) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val manager = context.getSystemService(NotificationManager::class.java)

        val alertChannel = NotificationChannel(
            ALERT_CHANNEL_ID,
            "风险预警",
            NotificationManager.IMPORTANCE_HIGH,
        ).apply {
            description = "中高风险案件与家庭风险提醒"
            enableVibration(true)
            setShowBadge(true)
            lockscreenVisibility = android.app.Notification.VISIBILITY_PUBLIC
        }

        val serviceChannel = NotificationChannel(
            SERVICE_CHANNEL_ID,
            "后台守护",
            NotificationManager.IMPORTANCE_LOW,
        ).apply {
            description = "保持风险预警后台连接"
            setShowBadge(false)
        }

        val quickAnalyzeChannel = NotificationChannel(
            QUICK_ANALYZE_SERVICE_CHANNEL_ID,
            "快捷分析悬浮窗",
            NotificationManager.IMPORTANCE_LOW,
        ).apply {
            description = "保持悬浮球与截屏快速分析能力"
            setShowBadge(false)
        }

        manager.createNotificationChannel(alertChannel)
        manager.createNotificationChannel(serviceChannel)
        manager.createNotificationChannel(quickAnalyzeChannel)
    }

    fun buildServiceNotification(context: Context) =
        NotificationCompat.Builder(context, SERVICE_CHANNEL_ID)
            .setSmallIcon(android.R.drawable.stat_notify_sync)
            .setContentTitle("反诈卫士后台守护中")
            .setContentText("正在持续监听中高风险预警与家庭通知")
            .setOngoing(true)
            .setSilent(true)
            .setContentIntent(mainPendingIntent(context))
            .build()

    fun serviceNotificationId(): Int = SERVICE_NOTIFICATION_ID

    fun buildQuickAnalyzeServiceNotification(context: Context) =
        NotificationCompat.Builder(context, QUICK_ANALYZE_SERVICE_CHANNEL_ID)
            .setSmallIcon(android.R.drawable.ic_menu_camera)
            .setContentTitle("快捷分析悬浮窗已开启")
            .setContentText("点击屏幕气泡即可截屏并快速识别当前风险")
            .setOngoing(true)
            .setSilent(true)
            .setContentIntent(mainPendingIntent(context))
            .build()

    fun quickAnalyzeServiceNotificationId(): Int = QUICK_ANALYZE_SERVICE_NOTIFICATION_ID

    fun showRiskAlert(
        context: Context,
        title: String,
        body: String,
        notificationId: Int,
        riskLevel: String = "高",
    ) {
        ensureChannels(context)
        val normalizedRiskLevel = normalizeRiskLevel(riskLevel)
        val defaultTitle = if (normalizedRiskLevel == "高") "高风险预警" else "中风险提醒"
        val defaultBody = if (normalizedRiskLevel == "高") "检测到新的高风险事件，请尽快查看。" else "检测到新的中风险事件，请留意核查。"
        val accentColor = if (normalizedRiskLevel == "高") 0xFFDC2626.toInt() else 0xFFD97706.toInt()
        val contentTitle = title.ifBlank { defaultTitle }
        val contentBody = body.ifBlank { defaultBody }
        val notification = NotificationCompat.Builder(context, ALERT_CHANNEL_ID)
            .setSmallIcon(android.R.drawable.stat_sys_warning)
            .setColor(accentColor)
            .setContentTitle(contentTitle)
            .setContentText(contentBody)
            .setSubText("后台检测提醒")
            .setStyle(NotificationCompat.BigTextStyle().bigText(contentBody))
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setCategory(NotificationCompat.CATEGORY_ALARM)
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
            .setAutoCancel(true)
            .setGroup(ALERT_GROUP)
            .setDefaults(NotificationCompat.DEFAULT_ALL)
            .setContentIntent(mainPendingIntent(context))
            .build()

        runCatching {
            NotificationManagerCompat.from(context).notify(notificationId, notification)
        }
    }

    private fun normalizeRiskLevel(value: String): String = when {
        value.contains("高") -> "高"
        value.contains("中") -> "中"
        value.contains("低") -> "低"
        else -> "高"
    }

    private fun mainPendingIntent(context: Context): PendingIntent {
        val intent = Intent(context, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        return PendingIntent.getActivity(
            context,
            0,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )
    }
}
