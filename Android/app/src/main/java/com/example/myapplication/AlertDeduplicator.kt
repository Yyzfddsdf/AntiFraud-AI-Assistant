package com.example.myapplication

import android.content.Context

class AlertDeduplicator(context: Context) {
    private val preferences = context.getSharedPreferences("sentinel_alert_dedup", Context.MODE_PRIVATE)

    fun wasAlertNotified(recordId: String): Boolean {
        if (recordId.isBlank()) return true
        return wasNotified("alert:$recordId")
    }

    fun wasFamilyNotified(notificationId: Int): Boolean {
        if (notificationId <= 0) return true
        return wasNotified("family:$notificationId")
    }

    fun markAlertNotified(recordId: String) {
        if (recordId.isBlank()) return
        markNotified("alert:$recordId")
    }

    fun markFamilyNotified(notificationId: Int) {
        if (notificationId <= 0) return
        markNotified("family:$notificationId")
    }

    private fun wasNotified(key: String): Boolean {
        val now = System.currentTimeMillis()
        val lastSeenAt = preferences.getLong(key, 0L)
        return lastSeenAt > 0L && now - lastSeenAt < DEDUP_WINDOW_MS
    }

    private fun markNotified(key: String) {
        val now = System.currentTimeMillis()
        preferences.edit()
            .putLong(key, now)
            .apply()
        cleanupIfNeeded(now)
    }

    private fun cleanupIfNeeded(now: Long) {
        val lastCleanupAt = preferences.getLong(PREF_LAST_CLEANUP_AT, 0L)
        if (now - lastCleanupAt < CLEANUP_INTERVAL_MS) return

        val cutoff = now - DEDUP_WINDOW_MS
        val staleKeys = preferences.all
            .filter { it.key != PREF_LAST_CLEANUP_AT }
            .mapNotNull { (key, value) -> key.takeIf { (value as? Long ?: 0L) < cutoff } }

        if (staleKeys.isEmpty()) {
            preferences.edit().putLong(PREF_LAST_CLEANUP_AT, now).apply()
            return
        }

        preferences.edit().apply {
            staleKeys.forEach(::remove)
            putLong(PREF_LAST_CLEANUP_AT, now)
        }.apply()
    }

    companion object {
        private const val PREF_LAST_CLEANUP_AT = "last_cleanup_at"
        private const val DEDUP_WINDOW_MS = 10 * 60 * 1000L
        private const val CLEANUP_INTERVAL_MS = 60 * 60 * 1000L
    }
}
