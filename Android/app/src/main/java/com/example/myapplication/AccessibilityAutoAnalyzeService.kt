package com.example.myapplication

import android.accessibilityservice.AccessibilityService
import android.accessibilityservice.AccessibilityServiceInfo
import android.content.ComponentName
import android.content.Context
import android.content.SharedPreferences
import android.graphics.Bitmap
import android.os.Build
import android.provider.Settings
import android.view.Display
import android.view.accessibility.AccessibilityEvent
import android.view.accessibility.AccessibilityNodeInfo
import android.widget.Toast
import com.google.android.gms.tasks.Task
import com.google.mlkit.vision.common.InputImage
import com.google.mlkit.vision.text.TextRecognition
import com.google.mlkit.vision.text.TextRecognizer
import com.google.mlkit.vision.text.chinese.ChineseTextRecognizerOptions
import java.io.ByteArrayOutputStream
import java.util.ArrayDeque
import java.util.LinkedHashSet
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

class AccessibilityAutoAnalyzeService : AccessibilityService() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Default)

    private lateinit var config: AccessibilityAutoAnalyzeConfig
    private lateinit var repository: SentinelRepository
    private lateinit var preferences: SharedPreferences
    private lateinit var ocrRecognizer: TextRecognizer

    @Volatile
    private var analyzing = false
    private var lastTriggerAt = 0L
    private var lastFingerprint = ""
    private var lastOcrAttemptAt = 0L
    private var lastDebugToastAt = 0L
    private var lastDebugToastMessage = ""
    private var lastResolvedPackage = ""
    private lateinit var activeConfig: ResolvedAccessibilityAutoAnalyzeConfig

    override fun onCreate() {
        super.onCreate()
        config = AccessibilityAutoAnalyzeConfigLoader.load(this)
        activeConfig = config.resolveForPackage("")
        repository = SentinelRepository(AppConfigLoader.load(this).apiOrigin)
        preferences = getSharedPreferences("sentinel", Context.MODE_PRIVATE)
        ocrRecognizer = TextRecognition.getClient(ChineseTextRecognizerOptions.Builder().build())
        AppNotifier.ensureChannels(this)
    }

    override fun onServiceConnected() {
        super.onServiceConnected()
        serviceInfo = serviceInfo.apply {
            flags = flags or
                AccessibilityServiceInfo.FLAG_RETRIEVE_INTERACTIVE_WINDOWS or
                AccessibilityServiceInfo.FLAG_REPORT_VIEW_IDS or
                AccessibilityServiceInfo.FLAG_INCLUDE_NOT_IMPORTANT_VIEWS
        }
    }

    override fun onAccessibilityEvent(event: AccessibilityEvent?) {
        if (!preferences.getBoolean(PREF_ENABLED, false)) return
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.R) return
        if (analyzing) {
            showDebugTip("自动守护：正在分析中，忽略本次事件")
            return
        }

        val token = preferences.getString(BackgroundAlertService.PREF_TOKEN, "").orEmpty()
        if (token.isBlank()) {
            showDebugTip("自动守护：当前未登录，跳过分析")
            return
        }

        val packageName = event?.packageName?.toString().orEmpty()
        if (packageName.isBlank()) return
        if (packageName == applicationContext.packageName) return
        if (shouldIgnorePackage(packageName)) return
        refreshActiveConfig(packageName)
        val accessibilityTexts = rootInActiveWindow?.let(::collectTexts).orEmpty()
        val accessibilityCandidate = buildCandidate(packageName, accessibilityTexts)
        if (accessibilityCandidate == null && !shouldAttemptOcr(event)) {
            showDebugTip("自动守护：界面文字未命中，且本轮不触发 OCR")
            return
        }
        if (accessibilityCandidate != null && shouldThrottle(accessibilityCandidate.fingerprint)) {
            showDebugTip("自动守护：命中规则，但仍在冷却时间内")
            return
        }

        analyzing = true
        showDebugTip(
            if (accessibilityCandidate != null) {
                "自动守护：命中界面文字规则，开始快速分析"
            } else {
                "自动守护：开始 OCR 兜底识别"
            }
        )

        scope.launch {
            try {
                val analysisPayload = if (accessibilityCandidate != null) {
                    val screenshotBase64 = captureScreenshotBase64() ?: run {
                        showDebugTip("自动守护：截图失败，已取消本次分析")
                        return@launch
                    }
                    AnalyzePayload(
                        candidate = accessibilityCandidate,
                        screenshotBase64 = screenshotBase64,
                    )
                } else {
                    val screenshotBitmap = captureScreenshotBitmap() ?: run {
                        showDebugTip("自动守护：OCR 截图失败")
                        return@launch
                    }
                    try {
                        val ocrTexts = runCatching { recognizeOcrTexts(screenshotBitmap) }
                            .getOrElse { emptyList() }
                        val ocrCandidate = buildCandidate(packageName, ocrTexts) ?: run {
                            showDebugTip("自动守护：OCR 未识别到敏感场景")
                            return@launch
                        }
                        if (shouldThrottle(ocrCandidate.fingerprint)) {
                            showDebugTip("自动守护：OCR 命中规则，但仍在冷却时间内")
                            return@launch
                        }
                        showDebugTip("自动守护：OCR 命中敏感文字，开始快速分析")
                        AnalyzePayload(
                            candidate = ocrCandidate,
                            screenshotBase64 = bitmapToBase64(screenshotBitmap),
                        )
                    } finally {
                        screenshotBitmap.recycle()
                    }
                }
                lastTriggerAt = System.currentTimeMillis()
                lastFingerprint = analysisPayload.candidate.fingerprint
                runCatching {
                    repository.quickAnalyzeImage(token, analysisPayload.screenshotBase64)
                }.onSuccess { result ->
                    val presentation = buildQuickAnalyzePresentation(
                        result = result,
                        contextPrefix = "检测到屏幕存在${analysisPayload.candidate.triggerSummary}等敏感线索，已自动完成快速分析。"
                    )
                    if (!presentation.shouldNotifyForAutoTrigger()) {
                        showDebugTip("自动守护：API 判断为低风险，已静默处理")
                        return@onSuccess
                    }
                    showDebugTip("自动守护：API 判断为${presentation.riskLevel}风险，已发出提醒")
                    AppNotifier.showRiskAlert(
                        context = this@AccessibilityAutoAnalyzeService,
                        title = presentation.title,
                        body = presentation.body,
                        notificationId = analysisPayload.candidate.fingerprint.hashCode(),
                        riskLevel = presentation.riskLevel,
                    )
                }.onFailure { throwable ->
                    val exception = throwable as? ApiException
                    if (exception?.statusCode == 401) {
                        preferences.edit().putBoolean(PREF_ENABLED, false).apply()
                        showDebugTip("自动守护：登录失效，已暂停")
                        AppNotifier.showRiskAlert(
                            context = this@AccessibilityAutoAnalyzeService,
                            title = "自动屏幕守护已暂停",
                            body = "当前登录状态已失效，请重新进入应用登录后再开启自动守护。",
                            notificationId = AUTH_EXPIRED_NOTIFICATION_ID,
                            riskLevel = "中",
                        )
                    }
                }
            } finally {
                analyzing = false
            }
        }
    }

    override fun onInterrupt() = Unit

    override fun onDestroy() {
        ocrRecognizer.close()
        repository.close()
        scope.cancel()
        super.onDestroy()
    }

    private fun buildCandidate(
        packageName: String,
        texts: List<String>,
    ): SensitiveSceneCandidate? {
        if (texts.isEmpty()) return null

        val matchedGroups = activeConfig.keywordGroups.mapNotNull { group ->
            val matchedKeywords = group.keywords.filter { keyword ->
                texts.any { text -> text.contains(keyword) }
            }
            if (matchedKeywords.isEmpty()) {
                null
            } else {
                MatchedKeywordGroup(group = group, matchedKeywords = matchedKeywords)
            }
        }
        val matchedKeywords = matchedGroups.flatMap { it.matchedKeywords }.distinct()
        if (matchedKeywords.isEmpty()) return null

        val score = buildSensitivityScore(texts, matchedGroups)
        if (score < activeConfig.minSensitiveScore) return null

        val fingerprint = buildString {
            append(packageName)
            append('|')
            append(matchedKeywords.sorted().joinToString(","))
            append('|')
            append(texts.take(3).joinToString("|"))
        }
        return SensitiveSceneCandidate(
            fingerprint = fingerprint,
            triggerSummary = matchedKeywords.take(4).joinToString("、"),
        )
    }

    private fun buildSensitivityScore(
        texts: List<String>,
        matchedGroups: List<MatchedKeywordGroup>,
    ): Int {
        var score = matchedGroups.sumOf { it.group.weight }
        if (matchedGroups.size >= activeConfig.multiGroupBonusThreshold) {
            score += activeConfig.multiGroupBonus
        }
        if (activeConfig.linkIndicators.any { indicator -> texts.any { text -> text.contains(indicator, ignoreCase = true) } }) {
            score += activeConfig.linkIndicatorBonus
        }
        if (
            texts.any { VERIFICATION_CODE_REGEX.containsMatchIn(it) } &&
            matchedGroups.any { it.group.id.contains("verification", ignoreCase = true) }
        ) {
            score += activeConfig.verificationCodeBonus
        }
        return score
    }

    private fun collectTexts(root: AccessibilityNodeInfo): List<String> {
        val queue = ArrayDeque<AccessibilityNodeInfo>()
        val results = LinkedHashSet<String>()
        queue.add(root)
        while (queue.isNotEmpty() && results.size < activeConfig.maxCapturedTexts) {
            val node = queue.removeFirst()
            addNormalizedText(results, node.text)
            addNormalizedText(results, node.contentDescription)
            for (index in 0 until node.childCount) {
                node.getChild(index)?.let(queue::addLast)
            }
            node.recycle()
        }
        return results.toList()
    }

    private fun addNormalizedText(
        results: MutableSet<String>,
        source: CharSequence?,
    ) {
        val value = source?.toString()
            ?.replace(Regex("\\s+"), " ")
            ?.trim()
            .orEmpty()
        if (value.length < 2) return
        results += value.take(activeConfig.maxTextLength)
    }

    private fun shouldIgnorePackage(packageName: String): Boolean {
        return config.ignoredPackagePrefixes.any { packageName.startsWith(it) } ||
            config.ignoredPackageKeywords.any { packageName.contains(it) }
    }

    private fun shouldThrottle(fingerprint: String): Boolean {
        val now = System.currentTimeMillis()
        if (fingerprint == lastFingerprint && now - lastTriggerAt < activeConfig.sameScreenCooldownMs) return true
        if (now - lastTriggerAt < activeConfig.globalCooldownMs) return true
        return false
    }

    private fun shouldAttemptOcr(event: AccessibilityEvent?): Boolean {
        if (!activeConfig.ocrFallbackEnabled) return false
        val eventType = event?.eventType ?: return false
        val isEligibleEvent = eventType == AccessibilityEvent.TYPE_WINDOW_STATE_CHANGED ||
            eventType == AccessibilityEvent.TYPE_WINDOW_CONTENT_CHANGED
        if (!isEligibleEvent) return false
        val now = System.currentTimeMillis()
        if (now - lastOcrAttemptAt < activeConfig.ocrAttemptCooldownMs) return false
        lastOcrAttemptAt = now
        return true
    }

    private suspend fun captureScreenshotBase64(): String? {
        val bitmap = captureScreenshotBitmap() ?: return null
        return try {
            bitmapToBase64(bitmap)
        } finally {
            bitmap.recycle()
        }
    }

    private suspend fun captureScreenshotBitmap(): Bitmap? {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.R) return null
        return withContext(Dispatchers.Main.immediate) {
            suspendCancellableCoroutine { continuation ->
                takeScreenshot(
                    Display.DEFAULT_DISPLAY,
                    mainExecutor,
                    object : TakeScreenshotCallback {
                        override fun onSuccess(screenshot: ScreenshotResult) {
                            val buffer = screenshot.hardwareBuffer
                            val bitmap = Bitmap.wrapHardwareBuffer(buffer, screenshot.colorSpace)
                                ?.copy(Bitmap.Config.ARGB_8888, false)
                            buffer.close()
                            continuation.resume(bitmap)
                        }

                        override fun onFailure(errorCode: Int) {
                            continuation.resume(null)
                        }
                    },
                )
            }
        }
    }

    private suspend fun recognizeOcrTexts(bitmap: Bitmap): List<String> {
        val image = InputImage.fromBitmap(bitmap, 0)
        val result = ocrRecognizer.process(image).await()
        val lines = LinkedHashSet<String>()
        result.textBlocks.forEach { block ->
            block.lines.forEach { line ->
                val text = line.text.replace(Regex("\\s+"), " ").trim()
                if (text.length >= 2) {
                    lines += text.take(activeConfig.maxTextLength)
                }
            }
        }
        if (lines.isEmpty() && result.text.isNotBlank()) {
            result.text.lineSequence()
                .map { it.replace(Regex("\\s+"), " ").trim() }
                .filter { it.length >= 2 }
                .forEach { lines += it.take(activeConfig.maxTextLength) }
        }
        return lines.toList()
    }

    private fun bitmapToBase64(bitmap: Bitmap): String {
        val stream = ByteArrayOutputStream()
        bitmap.compress(Bitmap.CompressFormat.JPEG, 90, stream)
        return android.util.Base64.encodeToString(stream.toByteArray(), android.util.Base64.NO_WRAP)
    }

    private fun showDebugTip(message: String) {
        if (!activeConfig.debugToastEnabled) return
        val now = System.currentTimeMillis()
        if (
            message == lastDebugToastMessage &&
            now - lastDebugToastAt < activeConfig.debugToastMinIntervalMs
        ) {
            return
        }
        lastDebugToastAt = now
        lastDebugToastMessage = message
        mainExecutor.execute {
            Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
        }
    }

    private fun refreshActiveConfig(packageName: String) {
        if (::activeConfig.isInitialized && lastResolvedPackage == packageName) return
        activeConfig = config.resolveForPackage(packageName)
        lastResolvedPackage = packageName
    }

    companion object {
        const val PREF_ENABLED = "accessibility_auto_analyze_enabled"
        private const val AUTH_EXPIRED_NOTIFICATION_ID = 31_001
        private val VERIFICATION_CODE_REGEX = Regex("\\b\\d{4,6}\\b")

        fun isEnabledInSystem(context: Context): Boolean {
            val enabledServices = Settings.Secure.getString(
                context.contentResolver,
                Settings.Secure.ENABLED_ACCESSIBILITY_SERVICES,
            ).orEmpty()
            if (enabledServices.isBlank()) return false
            val expected = ComponentName(context, AccessibilityAutoAnalyzeService::class.java).flattenToString()
            return enabledServices.split(':').any { it.equals(expected, ignoreCase = true) }
        }
    }
}

private data class SensitiveSceneCandidate(
    val fingerprint: String,
    val triggerSummary: String,
)

private data class AnalyzePayload(
    val candidate: SensitiveSceneCandidate,
    val screenshotBase64: String,
)

private data class MatchedKeywordGroup(
    val group: AccessibilityKeywordGroup,
    val matchedKeywords: List<String>,
)

private suspend fun <T> Task<T>.await(): T = suspendCancellableCoroutine { continuation ->
    addOnSuccessListener { result ->
        if (continuation.isActive) continuation.resume(result)
    }
    addOnFailureListener { error ->
        if (continuation.isActive) continuation.resumeWithException(error)
    }
    addOnCanceledListener {
        if (continuation.isActive) continuation.cancel()
    }
}
