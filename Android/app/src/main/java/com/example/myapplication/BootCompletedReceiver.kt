package com.example.myapplication

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

class BootCompletedReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent?) {
        if (intent?.action != Intent.ACTION_BOOT_COMPLETED) return
        val preferences = context.getSharedPreferences("sentinel", Context.MODE_PRIVATE)
        if (preferences.getString(BackgroundAlertService.PREF_TOKEN, "").isNullOrBlank()) return
        BackgroundAlertService.start(context)
    }
}
