package com.example.myapplication

import android.app.Activity
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.SharedPreferences
import android.content.pm.ServiceInfo
import android.graphics.Bitmap
import android.graphics.Color
import android.graphics.PixelFormat
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.hardware.display.DisplayManager
import android.hardware.display.VirtualDisplay
import android.media.Image
import android.media.ImageReader
import android.media.projection.MediaProjection
import android.media.projection.MediaProjectionManager
import android.os.Build
import android.os.Handler
import android.os.IBinder
import android.os.Looper
import android.provider.Settings
import android.util.Base64
import android.view.Gravity
import android.view.MotionEvent
import android.view.View
import android.view.ViewConfiguration
import android.view.WindowManager
import android.widget.FrameLayout
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.core.content.ContextCompat
import java.io.ByteArrayOutputStream
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import kotlin.math.abs

class QuickAnalyzeOverlayService : Service() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private val mainHandler = Handler(Looper.getMainLooper())

    private lateinit var repository: SentinelRepository
    private lateinit var preferences: SharedPreferences
    private lateinit var windowManager: WindowManager
    private lateinit var mediaProjectionManager: MediaProjectionManager

    private var mediaProjection: MediaProjection? = null
    private var virtualDisplay: VirtualDisplay? = null
    private var imageReader: ImageReader? = null
    private var bubbleView: FrameLayout? = null
    private var bubbleIconView: ImageView? = null
    private var bubbleLoadingView: LinearLayout? = null
    private var bubbleLoadingDots: List<View> = emptyList()
    private var bubbleLayoutParams: WindowManager.LayoutParams? = null
    private var bannerView: View? = null
    private var bannerHideRunnable: Runnable? = null
    private var bubbleTipView: View? = null
    private var bubbleTipHideRunnable: Runnable? = null
    private var bubbleLoadingRunnable: Runnable? = null
    private var bubbleLoadingStep = 0
    private var analyzing = false

    private val projectionCallback = object : MediaProjection.Callback() {
        override fun onStop() {
            stopSelf()
        }
    }

    override fun onCreate() {
        super.onCreate()
        running = true
        repository = SentinelRepository(AppConfigLoader.load(this).apiOrigin)
        preferences = getSharedPreferences("sentinel", Context.MODE_PRIVATE)
        windowManager = getSystemService(Context.WINDOW_SERVICE) as WindowManager
        mediaProjectionManager = getSystemService(Context.MEDIA_PROJECTION_SERVICE) as MediaProjectionManager
        AppNotifier.ensureChannels(this)
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        return when (intent?.action) {
            ACTION_START, null -> {
                if (!Settings.canDrawOverlays(this)) {
                    stopSelf()
                    START_NOT_STICKY
                } else {
                    val resultCode = intent?.getIntExtra(EXTRA_RESULT_CODE, Activity.RESULT_CANCELED)
                        ?: Activity.RESULT_CANCELED
                    val permissionData = intent?.readPermissionData()
                    if (resultCode != Activity.RESULT_OK || permissionData == null) {
                        stopSelf()
                        START_NOT_STICKY
                    } else {
                        startAsForeground()
                        if (!startProjection(resultCode, permissionData)) {
                            stopSelf()
                            START_NOT_STICKY
                        } else {
                            showBubble()
                            persistEnabled(true)
                            START_NOT_STICKY
                        }
                    }
                }
            }

            else -> START_NOT_STICKY
        }
    }

    override fun onDestroy() {
        removeBanner()
        removeBubble()
        stopProjection()
        repository.close()
        scope.cancel()
        persistEnabled(false)
        running = false
        super.onDestroy()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    private fun startAsForeground() {
        val notification = AppNotifier.buildQuickAnalyzeServiceNotification(this)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            startForeground(
                AppNotifier.quickAnalyzeServiceNotificationId(),
                notification,
                ServiceInfo.FOREGROUND_SERVICE_TYPE_MEDIA_PROJECTION,
            )
        } else {
            startForeground(AppNotifier.quickAnalyzeServiceNotificationId(), notification)
        }
    }

    private fun startProjection(resultCode: Int, permissionData: Intent): Boolean {
        stopProjection()
        return runCatching {
            val metrics = resources.displayMetrics
            val size = displaySize()
            val reader = ImageReader.newInstance(size.first, size.second, PixelFormat.RGBA_8888, 2)
            val projection = mediaProjectionManager.getMediaProjection(resultCode, permissionData)
                ?: error("media projection unavailable")
            projection.registerCallback(projectionCallback, mainHandler)
            val display = projection.createVirtualDisplay(
                "sentinel-quick-analyze",
                size.first,
                size.second,
                metrics.densityDpi,
                DisplayManager.VIRTUAL_DISPLAY_FLAG_AUTO_MIRROR,
                reader.surface,
                null,
                mainHandler,
            )
            imageReader = reader
            mediaProjection = projection
            virtualDisplay = display
        }.isSuccess
    }

    private fun stopProjection() {
        virtualDisplay?.release()
        virtualDisplay = null
        imageReader?.close()
        imageReader = null
        mediaProjection?.unregisterCallback(projectionCallback)
        mediaProjection?.stop()
        mediaProjection = null
    }

    private fun showBubble() {
        if (bubbleView != null) return
        val icon = ImageView(this).apply {
            setImageDrawable(ContextCompat.getDrawable(this@QuickAnalyzeOverlayService, R.drawable.ic_overlay_magnifier))
            scaleType = ImageView.ScaleType.CENTER_INSIDE
            imageAlpha = 245
        }
        val loading = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER
            visibility = View.GONE
        }
        val dots = List(3) { index ->
            View(this).apply {
                background = GradientDrawable().apply {
                    shape = GradientDrawable.OVAL
                    setColor(Color.parseColor("#0F172A"))
                }
                alpha = 0.28f
                scaleX = 0.82f
                scaleY = 0.82f
            }.also { dot ->
                loading.addView(
                    dot,
                    LinearLayout.LayoutParams(dp(6), dp(6)).apply {
                        if (index < 2) marginEnd = dp(4)
                    },
                )
            }
        }
        val view = FrameLayout(this).apply {
            background = GradientDrawable().apply {
                shape = GradientDrawable.OVAL
                setColor(Color.argb(120, 255, 255, 255))
                setStroke(dp(1), Color.parseColor("#5C10B981"))
            }
            elevation = dp(14).toFloat()
            setPadding(dp(12), dp(12), dp(12), dp(12))
            addView(
                icon,
                FrameLayout.LayoutParams(
                    FrameLayout.LayoutParams.MATCH_PARENT,
                    FrameLayout.LayoutParams.MATCH_PARENT,
                ),
            )
            addView(
                loading,
                FrameLayout.LayoutParams(
                    FrameLayout.LayoutParams.WRAP_CONTENT,
                    FrameLayout.LayoutParams.WRAP_CONTENT,
                    Gravity.CENTER,
                ),
            )
        }
        val params = WindowManager.LayoutParams(
            dp(56),
            dp(56),
            overlayWindowType(),
            WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE or
                WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN,
            PixelFormat.TRANSLUCENT,
        ).apply {
            gravity = Gravity.TOP or Gravity.START
            x = displaySize().first - dp(76)
            y = (displaySize().second * 0.45f).toInt()
        }
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            params.layoutInDisplayCutoutMode =
                WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES
        }
        attachBubbleTouch(view, params)
        runCatching { windowManager.addView(view, params) }
            .onSuccess {
                bubbleView = view
                bubbleIconView = icon
                bubbleLoadingView = loading
                bubbleLoadingDots = dots
                bubbleLayoutParams = params
            }
            .onFailure { stopSelf() }
    }

    private fun attachBubbleTouch(
        view: View,
        params: WindowManager.LayoutParams,
    ) {
        val touchSlop = ViewConfiguration.get(this).scaledTouchSlop
        view.setOnTouchListener(object : View.OnTouchListener {
            private var initialX = 0
            private var initialY = 0
            private var initialTouchX = 0f
            private var initialTouchY = 0f
            private var moved = false

            override fun onTouch(v: View, event: MotionEvent): Boolean {
                when (event.actionMasked) {
                    MotionEvent.ACTION_DOWN -> {
                        initialX = params.x
                        initialY = params.y
                        initialTouchX = event.rawX
                        initialTouchY = event.rawY
                        moved = false
                        return true
                    }

                    MotionEvent.ACTION_MOVE -> {
                        val deltaX = (event.rawX - initialTouchX).toInt()
                        val deltaY = (event.rawY - initialTouchY).toInt()
                        if (!moved && (abs(deltaX) > touchSlop || abs(deltaY) > touchSlop)) {
                            moved = true
                        }
                        params.x = initialX + deltaX
                        params.y = initialY + deltaY
                        bubbleLayoutParams = params
                        runCatching { windowManager.updateViewLayout(v, params) }
                        return true
                    }

                    MotionEvent.ACTION_UP -> {
                        if (!moved) {
                            v.performClick()
                        }
                        return true
                    }
                }
                return false
            }
        })
        view.setOnClickListener { onBubbleClicked() }
    }

    private fun onBubbleClicked() {
        if (analyzing) return
        val token = preferences.getString(BackgroundAlertService.PREF_TOKEN, "").orEmpty()
        if (token.isBlank()) {
            showBanner("快速分析暂不可用", "当前登录状态已失效，请重新进入应用登录。", "高")
            return
        }
        analyzing = true
        updateBubbleState(true)
        scope.launch {
            val imageBase64 = captureScreenBase64()
            if (imageBase64.isNullOrBlank()) {
                showBanner("快速分析失败", "未获取到当前屏幕画面，请稍后重试。", "高")
                analyzing = false
                updateBubbleState(false)
                return@launch
            }
            runCatching {
                repository.quickAnalyzeImage(token, imageBase64)
            }.onSuccess { result ->
                showQuickAnalyzeResult(result)
            }.onFailure { throwable ->
                val message = (throwable as? ApiException)?.message ?: "网络异常，请稍后重试。"
                showBanner("快速分析失败", message, "高")
            }
            analyzing = false
            updateBubbleState(false)
        }
    }

    private suspend fun captureScreenBase64(): String? {
        withContext(Dispatchers.Main) {
            // Temporarily hide overlay UI so the captured frame only contains the user's screen.
            removeBanner()
            removeBubbleTip()
            bubbleView?.visibility = View.INVISIBLE
        }
        delay(140)
        return try {
            repeat(12) {
                val image = imageReader?.acquireLatestImage()
                if (image != null) {
                    return try {
                        imageToBase64(image)
                    } finally {
                        image.close()
                    }
                }
                delay(80)
            }
            null
        } finally {
            withContext(Dispatchers.Main) {
                bubbleView?.visibility = View.VISIBLE
            }
        }
    }

    private fun imageToBase64(image: Image): String? {
        val plane = image.planes.firstOrNull() ?: return null
        val buffer = plane.buffer
        val width = image.width
        val height = image.height
        val pixelStride = plane.pixelStride
        val rowStride = plane.rowStride
        val rowPadding = rowStride - pixelStride * width
        val bitmap = Bitmap.createBitmap(
            width + rowPadding / pixelStride,
            height,
            Bitmap.Config.ARGB_8888,
        )
        bitmap.copyPixelsFromBuffer(buffer)
        val cropped = Bitmap.createBitmap(bitmap, 0, 0, width, height)
        if (cropped != bitmap) {
            bitmap.recycle()
        }
        val outputStream = ByteArrayOutputStream()
        cropped.compress(Bitmap.CompressFormat.JPEG, 90, outputStream)
        cropped.recycle()
        return Base64.encodeToString(outputStream.toByteArray(), Base64.NO_WRAP)
    }

    private fun showQuickAnalyzeResult(result: QuickAnalyzeResponse) {
        val presentation = buildQuickAnalyzePresentation(result)
        if (presentation.riskLevel == "低") {
            showBubbleTip(presentation.title)
        } else {
            showBanner(presentation.title, presentation.body, presentation.riskLevel)
        }
    }

    private fun showBanner(
        title: String,
        reason: String,
        riskLevel: String,
    ) {
        mainHandler.post {
            removeBubbleTip()
            removeBanner()
            val accent = Color.parseColor("#DC2626")
            val card = LinearLayout(this).apply {
                gravity = Gravity.CENTER_VERTICAL
                orientation = LinearLayout.VERTICAL
                background = GradientDrawable().apply {
                    cornerRadius = dp(20).toFloat()
                    setColor(Color.parseColor("#DC2626"))
                    setStroke(dp(1), Color.parseColor("#F87171"))
                }
                elevation = dp(18).toFloat()
                setPadding(dp(16), dp(14), dp(16), dp(14))
                setOnClickListener {
                    openApp()
                    removeBanner()
                }
            }
            val headerRow = LinearLayout(this).apply {
                orientation = LinearLayout.HORIZONTAL
                gravity = Gravity.CENTER_VERTICAL
            }
            val iconWrap = FrameLayout(this).apply {
                background = GradientDrawable().apply {
                    shape = GradientDrawable.OVAL
                    setColor(Color.parseColor("#FEE2E2"))
                }
                layoutParams = LinearLayout.LayoutParams(dp(36), dp(36)).apply {
                    marginEnd = dp(12)
                }
            }
            val iconView = ImageView(this).apply {
                setImageDrawable(ContextCompat.getDrawable(this@QuickAnalyzeOverlayService, android.R.drawable.ic_dialog_alert))
                setColorFilter(accent)
                scaleType = ImageView.ScaleType.CENTER_INSIDE
            }
            iconWrap.addView(
                iconView,
                FrameLayout.LayoutParams(dp(18), dp(18), Gravity.CENTER),
            )
            val contentColumn = LinearLayout(this).apply {
                orientation = LinearLayout.VERTICAL
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            }
            val eyebrowView = TextView(this).apply {
                text = "后台检测提醒"
                setTextColor(Color.parseColor("#FECACA"))
                textSize = 10f
                typeface = Typeface.DEFAULT_BOLD
                letterSpacing = 0.08f
            }
            val titleView = TextView(this).apply {
                text = title
                setTextColor(Color.WHITE)
                textSize = 15f
                typeface = Typeface.DEFAULT_BOLD
                setPadding(0, dp(2), 0, 0)
            }
            val reasonView = TextView(this).apply {
                text = reason
                setTextColor(Color.parseColor("#FEE2E2"))
                textSize = 12f
                setLineSpacing(dp(3).toFloat(), 1f)
                setPadding(0, dp(8), 0, 0)
            }
            contentColumn.addView(eyebrowView)
            contentColumn.addView(titleView)
            contentColumn.addView(reasonView)
            val riskChipView = TextView(this).apply {
                text = "${riskLevel}风险"
                setTextColor(accent)
                textSize = 10f
                typeface = Typeface.DEFAULT_BOLD
                background = GradientDrawable().apply {
                    cornerRadius = dp(999).toFloat()
                    setColor(Color.WHITE)
                }
                setPadding(dp(10), dp(5), dp(10), dp(5))
            }
            headerRow.addView(iconWrap)
            headerRow.addView(contentColumn)
            headerRow.addView(riskChipView)
            val footerView = TextView(this).apply {
                text = "点击查看详情"
                setTextColor(Color.parseColor("#FECACA"))
                textSize = 10f
                typeface = Typeface.DEFAULT_BOLD
                setPadding(0, dp(10), 0, 0)
            }
            card.addView(headerRow)
            card.addView(footerView)
            val params = WindowManager.LayoutParams(
                displaySize().first - dp(24),
                WindowManager.LayoutParams.WRAP_CONTENT,
                overlayWindowType(),
                WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE or
                    WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN,
                PixelFormat.TRANSLUCENT,
            ).apply {
                gravity = Gravity.TOP or Gravity.CENTER_HORIZONTAL
                y = statusBarHeight() + dp(12)
            }
            runCatching { windowManager.addView(card, params) }
                .onSuccess {
                    bannerView = card
                    val hideRunnable = Runnable { removeBanner() }
                    bannerHideRunnable = hideRunnable
                    mainHandler.postDelayed(hideRunnable, 4_500L)
                }
        }
    }

    private fun showBubbleTip(message: String) {
        mainHandler.post {
            removeBanner()
            removeBubbleTip()
            val bubbleParams = bubbleLayoutParams ?: return@post
            val textView = TextView(this).apply {
                text = message
                setTextColor(Color.parseColor("#065F46"))
                textSize = 12f
                typeface = Typeface.DEFAULT_BOLD
                background = GradientDrawable().apply {
                    cornerRadius = dp(999).toFloat()
                    setColor(Color.parseColor("#ECFDF5"))
                    setStroke(dp(1), Color.parseColor("#A7F3D0"))
                }
                setPadding(dp(12), dp(8), dp(12), dp(8))
                elevation = dp(8).toFloat()
            }
            val bubbleSize = bubbleView?.width?.takeIf { it > 0 } ?: dp(56)
            val displayWidth = displaySize().first
            val showOnLeft = bubbleParams.x > displayWidth / 2
            val params = WindowManager.LayoutParams(
                WindowManager.LayoutParams.WRAP_CONTENT,
                WindowManager.LayoutParams.WRAP_CONTENT,
                overlayWindowType(),
                WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE or
                    WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN,
                PixelFormat.TRANSLUCENT,
            ).apply {
                gravity = Gravity.TOP or Gravity.START
                x = if (showOnLeft) {
                    (bubbleParams.x - dp(184)).coerceAtLeast(dp(12))
                } else {
                    (bubbleParams.x + bubbleSize + dp(10)).coerceAtMost(displayWidth - dp(140))
                }
                y = (bubbleParams.y + bubbleSize / 2 - dp(18)).coerceAtLeast(statusBarHeight() + dp(8))
            }
            runCatching { windowManager.addView(textView, params) }
                .onSuccess {
                    bubbleTipView = textView
                    val hideRunnable = Runnable { removeBubbleTip() }
                    bubbleTipHideRunnable = hideRunnable
                    mainHandler.postDelayed(hideRunnable, 2_600L)
                }
        }
    }

    private fun removeBubble() {
        stopBubbleLoadingAnimation()
        removeBubbleTip()
        bubbleView?.let { view ->
            runCatching { windowManager.removeView(view) }
        }
        bubbleView = null
        bubbleIconView = null
        bubbleLoadingView = null
        bubbleLoadingDots = emptyList()
        bubbleLayoutParams = null
    }

    private fun removeBanner() {
        bannerHideRunnable?.let(mainHandler::removeCallbacks)
        bannerHideRunnable = null
        bannerView?.let { view ->
            runCatching { windowManager.removeView(view) }
        }
        bannerView = null
    }

    private fun removeBubbleTip() {
        bubbleTipHideRunnable?.let(mainHandler::removeCallbacks)
        bubbleTipHideRunnable = null
        bubbleTipView?.let { view ->
            runCatching { windowManager.removeView(view) }
        }
        bubbleTipView = null
    }

    private fun updateBubbleState(busy: Boolean) {
        mainHandler.post {
            bubbleIconView?.visibility = if (busy) View.GONE else View.VISIBLE
            bubbleLoadingView?.visibility = if (busy) View.VISIBLE else View.GONE
            bubbleView?.alpha = if (busy) 0.78f else 1f
            bubbleView?.scaleX = if (busy) 0.94f else 1f
            bubbleView?.scaleY = if (busy) 0.94f else 1f
            if (busy) {
                startBubbleLoadingAnimation()
            } else {
                stopBubbleLoadingAnimation()
            }
        }
    }

    private fun startBubbleLoadingAnimation() {
        if (bubbleLoadingDots.isEmpty() || bubbleLoadingRunnable != null) return
        bubbleLoadingStep = 0
        val runnable = object : Runnable {
            override fun run() {
                bubbleLoadingDots.forEachIndexed { index, dot ->
                    val active = index == bubbleLoadingStep
                    dot.alpha = if (active) 1f else 0.28f
                    dot.scaleX = if (active) 1.08f else 0.82f
                    dot.scaleY = if (active) 1.08f else 0.82f
                }
                bubbleLoadingStep = (bubbleLoadingStep + 1) % bubbleLoadingDots.size
                mainHandler.postDelayed(this, 220L)
            }
        }
        bubbleLoadingRunnable = runnable
        mainHandler.post(runnable)
    }

    private fun stopBubbleLoadingAnimation() {
        bubbleLoadingRunnable?.let(mainHandler::removeCallbacks)
        bubbleLoadingRunnable = null
        bubbleLoadingStep = 0
        bubbleLoadingDots.forEach { dot ->
            dot.alpha = 0.28f
            dot.scaleX = 0.82f
            dot.scaleY = 0.82f
        }
    }

    private fun displaySize(): Pair<Int, Int> {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
            val bounds = windowManager.currentWindowMetrics.bounds
            bounds.width() to bounds.height()
        } else {
            val metrics = resources.displayMetrics
            metrics.widthPixels to metrics.heightPixels
        }
    }

    private fun overlayWindowType(): Int {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY
        } else {
            @Suppress("DEPRECATION")
            WindowManager.LayoutParams.TYPE_PHONE
        }
    }

    private fun statusBarHeight(): Int {
        val resourceId = resources.getIdentifier("status_bar_height", "dimen", "android")
        return if (resourceId > 0) resources.getDimensionPixelSize(resourceId) else dp(8)
    }

    private fun dp(value: Int): Int = (value * resources.displayMetrics.density).toInt()

    private fun openApp() {
        val intent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        startActivity(intent)
    }

    private fun persistEnabled(enabled: Boolean) {
        preferences.edit().putBoolean(PREF_ENABLED, enabled).apply()
    }

    private fun Intent.readPermissionData(): Intent? {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            getParcelableExtra(EXTRA_RESULT_DATA, Intent::class.java)
        } else {
            @Suppress("DEPRECATION")
            getParcelableExtra(EXTRA_RESULT_DATA)
        }
    }

    companion object {
        private const val ACTION_START = "com.example.myapplication.action.START_QUICK_ANALYZE_OVERLAY"
        private const val EXTRA_RESULT_CODE = "result_code"
        private const val EXTRA_RESULT_DATA = "result_data"
        const val PREF_ENABLED = "quick_analyze_bubble_enabled"

        @Volatile
        private var running = false

        fun start(
            context: Context,
            resultCode: Int,
            permissionData: Intent,
        ) {
            val intent = Intent(context, QuickAnalyzeOverlayService::class.java).apply {
                action = ACTION_START
                putExtra(EXTRA_RESULT_CODE, resultCode)
                putExtra(EXTRA_RESULT_DATA, permissionData)
            }
            ContextCompat.startForegroundService(context, intent)
        }

        fun stop(context: Context) {
            context.stopService(Intent(context, QuickAnalyzeOverlayService::class.java))
        }

        fun isRunning(): Boolean = running
    }
}
