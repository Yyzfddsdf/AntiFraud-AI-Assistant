package com.example.myapplication

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            SentinelApp()
        }
    }

    override fun onStart() {
        super.onStart()
        AppVisibilityTracker.isForeground = true
    }

    override fun onStop() {
        AppVisibilityTracker.isForeground = false
        super.onStop()
    }
}
