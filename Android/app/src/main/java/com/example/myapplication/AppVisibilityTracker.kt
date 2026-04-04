package com.example.myapplication

object AppVisibilityTracker {
    @Volatile
    var isForeground: Boolean = false
}
