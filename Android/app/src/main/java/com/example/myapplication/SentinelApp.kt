package com.example.myapplication

import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.Intent
import android.media.projection.MediaProjectionManager
import android.net.Uri
import android.os.Build
import android.os.PowerManager
import android.provider.Settings
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawing
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Add
import androidx.compose.material.icons.outlined.Campaign
import androidx.compose.material.icons.outlined.CheckCircle
import androidx.compose.material.icons.outlined.ErrorOutline
import androidx.compose.material.icons.outlined.Layers
import androidx.compose.material.icons.outlined.NotificationsNone
import androidx.compose.material.icons.outlined.PowerSettingsNew
import androidx.compose.material.icons.outlined.RocketLaunch
import androidx.compose.material.icons.outlined.ShowChart
import androidx.compose.material.icons.outlined.TipsAndUpdates
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Badge
import androidx.compose.material3.BadgedBox
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Snackbar
import androidx.compose.material3.SnackbarDuration
import androidx.compose.material3.SnackbarData
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.lifecycle.viewmodel.compose.viewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SentinelApp(viewModel: SentinelViewModel = viewModel()) {
    SentinelTheme {
        val state = viewModel.uiState
        val context = LocalContext.current
        val snackbarHostState = remember { SnackbarHostState() }
        val mediaProjectionManager = remember(context) {
            context.getSystemService(Context.MEDIA_PROJECTION_SERVICE) as MediaProjectionManager
        }
        val accessibilityAutoAnalyzeGranted = rememberAccessibilityAutoAnalyzeState(context)

        val imagePicker = rememberLauncherForActivityResult(ActivityResultContracts.OpenMultipleDocuments()) { uris ->
            viewModel.addAnalyzeAssets(AnalyzeAssetKind.Images, uris)
        }
        val audioPicker = rememberLauncherForActivityResult(ActivityResultContracts.OpenMultipleDocuments()) { uris ->
            viewModel.addAnalyzeAssets(AnalyzeAssetKind.Audios, uris)
        }
        val videoPicker = rememberLauncherForActivityResult(ActivityResultContracts.OpenMultipleDocuments()) { uris ->
            viewModel.addAnalyzeAssets(AnalyzeAssetKind.Videos, uris)
        }
        val chatImagePicker = rememberLauncherForActivityResult(ActivityResultContracts.OpenMultipleDocuments()) { uris ->
            viewModel.addChatImages(uris)
        }
        val notificationPermissionLauncher = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) { granted ->
            if (granted) {
                AppNotifier.ensureChannels(context)
                BackgroundAlertService.start(context)
            }
        }
        val locationPermissionLauncher = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) { granted ->
            if (granted) {
                viewModel.requestCurrentRegion()
            } else {
                viewModel.emitUiMessage("请先授予定位权限", isError = true)
            }
        }
        val screenCaptureLauncher = rememberLauncherForActivityResult(ActivityResultContracts.StartActivityForResult()) { result ->
            val permissionData = result.data
            if (result.resultCode == Activity.RESULT_OK && permissionData != null) {
                QuickAnalyzeOverlayService.start(context, result.resultCode, permissionData)
                viewModel.setQuickAnalyzeBubbleEnabled(true)
                viewModel.emitUiMessage("悬浮球已开启，点击屏幕气泡即可快速分析")
            } else {
                viewModel.setQuickAnalyzeBubbleEnabled(false)
                viewModel.emitUiMessage("未授予截屏权限，悬浮球未开启", isError = true)
            }
        }

        LaunchedEffect(viewModel.latestMessage?.id) {
            viewModel.latestMessage?.let { message ->
                snackbarHostState.showSnackbar(message.text, duration = SnackbarDuration.Short)
            }
        }

        LaunchedEffect(accessibilityAutoAnalyzeGranted) {
            viewModel.setAccessibilityAutoAnalyzePermissionGranted(accessibilityAutoAnalyzeGranted)
        }

        LaunchedEffect(state.screen, state.isAuthenticated) {
            if (!state.isAuthenticated) return@LaunchedEffect
            when (state.screen) {
                AppScreen.Dashboard -> {
                    viewModel.fetchTasks(silent = true)
                    viewModel.fetchHistory(silent = true)
                    viewModel.fetchRiskOverview(silent = true)
                    viewModel.fetchCurrentRegionCaseStatsIfConfigured(silent = true)
                }
                AppScreen.History -> viewModel.fetchHistory(silent = false)
                AppScreen.RiskTrend -> {
                    viewModel.fetchRiskOverview(silent = false)
                    viewModel.fetchCurrentRegionCaseStatsIfConfigured(silent = false)
                }
                AppScreen.SimulationQuiz -> {
                    viewModel.fetchSimulationPacks(silent = false)
                    viewModel.fetchSimulationSessions(silent = false)
                }
                AppScreen.Chat -> if (!state.chatHistoryLoaded) viewModel.fetchChatHistory()
                AppScreen.Family -> {
                    viewModel.fetchFamilyOverview(silent = false)
                    if (state.familyOverview.family == null) {
                        viewModel.fetchReceivedInvitations(silent = false)
                    }
                }
                AppScreen.FamilyManage -> viewModel.fetchFamilyOverview(silent = false)
                AppScreen.ProfilePrivacy -> viewModel.fetchProvinceOptions(silent = true)
                else -> Unit
            }
        }

        Surface(
            modifier = Modifier.fillMaxSize(),
            color = MaterialTheme.colorScheme.background,
        ) {
            Scaffold(
                snackbarHost = {
                    SnackbarHost(
                        hostState = snackbarHostState,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                    ) { data ->
                        AppMessageSnackbar(
                            data = data,
                            isError = viewModel.latestMessage?.isError == true,
                        )
                    }
                },
                contentWindowInsets = WindowInsets.safeDrawing,
                topBar = {
                    if (state.isAuthenticated && state.screen !in setOf(
                            AppScreen.Chat,
                            AppScreen.FamilyManage,
                            AppScreen.History,
                            AppScreen.RiskTrend,
                            AppScreen.ProfilePrivacy,
                            AppScreen.Submit,
                            AppScreen.SimulationQuiz,
                        )
                    ) {
                        SentinelTopBar(
                            unreadCount = state.alertItems.count { !it.read },
                            onOpenAlerts = { viewModel.openScreen(AppScreen.Alerts) },
                        )
                    }
                },
                bottomBar = {
                    if (state.isAuthenticated && state.screen !in setOf(AppScreen.Chat, AppScreen.FamilyManage, AppScreen.SimulationQuiz)) {
                        SentinelBottomBar(
                            current = state.screen,
                            onScreenChange = viewModel::openScreen,
                        )
                    }
                },
            ) { padding ->
                if (!state.isAuthenticated) {
                    AuthScreen(
                        state = state,
                        viewModel = viewModel,
                        modifier = Modifier.padding(padding),
                    )
                } else {
                    when (state.screen) {
                        AppScreen.Dashboard -> MobileDashboardScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                        )
                        AppScreen.History -> MobileHistoryScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                        )
                        AppScreen.RiskTrend -> MobileRiskTrendScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                        )
                        AppScreen.Alerts -> MobileAlertsScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                        )
                        AppScreen.Submit -> MobileSubmitScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                            onPickImages = { imagePicker.launch(arrayOf("image/*")) },
                            onPickAudios = { audioPicker.launch(arrayOf("audio/*")) },
                            onPickVideos = { videoPicker.launch(arrayOf("video/*")) },
                        )
                        AppScreen.SimulationQuiz -> MobileSimulationQuizScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                        )
                        AppScreen.Chat -> MobileChatScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                            onPickImages = { chatImagePicker.launch(arrayOf("image/*")) },
                        )
                        AppScreen.Family -> MobileFamilyScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                        )
                        AppScreen.FamilyManage -> MobileFamilyManageScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                        )
                        AppScreen.Profile -> MobileProfileScreen(
                            state = state,
                            padding = padding,
                            onOpenPrivacy = { viewModel.openScreen(AppScreen.ProfilePrivacy) },
                            onQuickAnalyzeBubbleChange = { enabled ->
                                if (enabled) {
                                    if (!hasOverlayPermission(context)) {
                                        viewModel.setQuickAnalyzeBubbleEnabled(false)
                                        openOverlayPermissionSettings(context)
                                        viewModel.emitUiMessage("请先授予悬浮窗权限，再开启悬浮球快捷分析", isError = true)
                                    } else {
                                        screenCaptureLauncher.launch(mediaProjectionManager.createScreenCaptureIntent())
                                    }
                                } else {
                                    QuickAnalyzeOverlayService.stop(context)
                                    viewModel.setQuickAnalyzeBubbleEnabled(false)
                                    viewModel.emitUiMessage("悬浮球已关闭")
                                }
                            },
                            onAccessibilityAutoAnalyzeChange = { enabled ->
                                if (!enabled) {
                                    viewModel.setAccessibilityAutoAnalyzeEnabled(false)
                                    viewModel.emitUiMessage("屏幕自动守护已关闭")
                                } else if (Build.VERSION.SDK_INT < Build.VERSION_CODES.R) {
                                    viewModel.setAccessibilityAutoAnalyzeEnabled(false)
                                    viewModel.emitUiMessage("当前系统版本不支持自动截图分析，需要 Android 11 及以上", isError = true)
                                } else {
                                    viewModel.setAccessibilityAutoAnalyzeEnabled(true)
                                    if (!accessibilityAutoAnalyzeGranted) {
                                        openAccessibilitySettings(context)
                                        viewModel.emitUiMessage("请在系统无障碍设置中开启“反诈屏幕守护”", isError = true)
                                    } else {
                                        viewModel.emitUiMessage("屏幕自动守护已开启，将在敏感场景下自动快速分析")
                                    }
                                }
                            },
                            onLogout = viewModel::logout,
                        )
                        AppScreen.ProfilePrivacy -> MobileProfilePrivacyScreen(
                            state = state,
                            padding = padding,
                            viewModel = viewModel,
                            onRequestLocation = {
                                if (ContextCompat.checkSelfPermission(context, Manifest.permission.ACCESS_FINE_LOCATION) == android.content.pm.PackageManager.PERMISSION_GRANTED ||
                                    ContextCompat.checkSelfPermission(context, Manifest.permission.ACCESS_COARSE_LOCATION) == android.content.pm.PackageManager.PERMISSION_GRANTED
                                ) {
                                    viewModel.requestCurrentRegion()
                                } else {
                                    locationPermissionLauncher.launch(Manifest.permission.ACCESS_FINE_LOCATION)
                                }
                            },
                        )
                    }
                }
            }
        }

        PermissionOnboardingGate(
            isAuthenticated = state.isAuthenticated,
            onRequestNotifications = {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                    notificationPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
                } else {
                    AppNotifier.ensureChannels(context)
                }
            },
        )

        state.selectedTask?.let { task ->
            MobileTaskDetailSheet(
                task = task,
                onDismiss = viewModel::closeTaskDetail,
                onMessage = { message, isError -> viewModel.emitUiMessage(message, isError) },
            )
        }
        state.activeAlert?.let { alert ->
            MobileAlertSheet(
                event = alert,
                onDismiss = viewModel::dismissAlert,
                onOpenCase = { viewModel.openAlertTask(alert.record_id) },
            )
        }
        state.activeFamilyNotification?.let { notification ->
            MobileFamilyAlertSheet(
                notification = notification,
                onDismiss = viewModel::dismissFamilyAlert,
                onOpenCenter = viewModel::openFamilyNotificationCenter,
            )
        }
    }
}

@Composable
private fun PermissionOnboardingGate(
    isAuthenticated: Boolean,
    onRequestNotifications: () -> Unit,
) {
    val context = LocalContext.current
    var dismissed by remember { mutableStateOf(false) }
    val notificationsGranted = rememberNotificationPermissionState(context)
    val batteryIgnored = rememberBatteryOptimizationState(context)
    val overlayGranted = rememberOverlayPermissionState(context)

    if (!isAuthenticated || dismissed) return
    if (notificationsGranted && batteryIgnored && overlayGranted) return

    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp, vertical = 20.dp),
        contentAlignment = Alignment.TopCenter,
    ) {
        Card(
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(containerColor = Color(0xFFFCFFFE)),
            modifier = Modifier
                .fillMaxWidth()
                .widthIn(max = 560.dp)
                .border(1.dp, Color(0xFFD1FAE5), RoundedCornerShape(16.dp)),
        ) {
            Column(
                modifier = Modifier.padding(20.dp),
                verticalArrangement = Arrangement.spacedBy(14.dp),
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(20.dp))
                        .background(
                            Brush.linearGradient(
                                listOf(Color(0xFFECFDF5), Color(0xFFF0F9FF)),
                            ),
                        )
                        .padding(16.dp),
                ) {
                    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Row(horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                            Box(
                                modifier = Modifier
                                    .size(42.dp)
                                    .clip(RoundedCornerShape(14.dp))
                                    .background(Color.White.copy(alpha = 0.82f)),
                                contentAlignment = Alignment.Center,
                            ) {
                                Icon(Icons.Outlined.TipsAndUpdates, contentDescription = null, tint = Color(0xFF059669))
                            }
                            Column {
                                Text("开启后台守护", fontWeight = FontWeight.Bold, fontSize = 19.sp, color = Color(0xFF0F172A))
                                Text("补齐关键授权后，顶部提醒会更稳定。", color = Color(0xFF475569), fontSize = 13.sp)
                            }
                        }
                        Text("建议优先完成通知、悬浮窗和电池优化设置；自启动入口作为不同机型的补充项。", color = Color(0xFF334155), fontSize = 13.sp)
                    }
                }
                PermissionGuideRow(
                    title = "通知提醒",
                    description = "允许中高风险事件在后台以顶部通知方式弹出",
                    ready = notificationsGranted,
                    buttonText = "去开启",
                    icon = Icons.Outlined.NotificationsNone,
                    onClick = onRequestNotifications,
                )
                PermissionGuideRow(
                    title = "电池优化白名单",
                    description = "减少后台连接被系统休眠或清理的概率",
                    ready = batteryIgnored,
                    buttonText = "去设置",
                    icon = Icons.Outlined.PowerSettingsNew,
                    onClick = { openBatteryOptimizationSettings(context) },
                )
                PermissionGuideRow(
                    title = "悬浮窗权限",
                    description = "允许悬浮球常驻屏幕，并支持点击后快速截屏分析",
                    ready = overlayGranted,
                    buttonText = "去设置",
                    icon = Icons.Outlined.Layers,
                    onClick = { openOverlayPermissionSettings(context) },
                )
                PermissionGuideRow(
                    title = "厂商自启动管理",
                    description = "不同品牌入口不同，开启后应用更不容易被后台限制",
                    ready = false,
                    buttonText = "去设置",
                    icon = Icons.Outlined.RocketLaunch,
                    onClick = { openAutoStartSettings(context) },
                )
                Spacer(Modifier.size(2.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text("只有缺少关键授权时才会显示这里", color = Color(0xFF94A3B8), fontSize = 12.sp)
                    TextButton(
                        onClick = {
                            dismissed = true
                        },
                    ) {
                        Text("稍后再说")
                    }
                }
            }
        }
    }
}

@Composable
private fun PermissionGuideRow(
    title: String,
    description: String,
    ready: Boolean,
    buttonText: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    onClick: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(18.dp))
            .background(if (ready) Color(0xFFF0FDF4) else Color(0xFFF8FAFC))
            .border(1.dp, if (ready) Color(0xFFA7F3D0) else Color(0xFFE2E8F0), RoundedCornerShape(18.dp))
            .padding(horizontal = 14.dp, vertical = 12.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(38.dp)
                .clip(RoundedCornerShape(12.dp))
                .background(if (ready) Color(0xFFDCFCE7) else Color.White),
            contentAlignment = Alignment.Center,
        ) {
            Icon(icon, contentDescription = null, tint = if (ready) Color(0xFF059669) else Color(0xFF475569))
        }
        Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
            Text(title, color = Color(0xFF0F172A), fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
            Text(description, color = Color(0xFF64748B), fontSize = 12.sp)
            Text(if (ready) "已完成" else "建议开启", color = if (ready) Color(0xFF059669) else Color(0xFF94A3B8), fontSize = 12.sp)
        }
        Button(
            onClick = onClick,
            shape = RoundedCornerShape(14.dp),
            colors = ButtonDefaults.buttonColors(containerColor = if (ready) Color(0xFFE2E8F0) else Color(0xFF059669)),
        ) {
            Text(if (ready) "已开启" else buttonText, color = if (ready) Color(0xFF475569) else Color.White)
        }
    }
}

@Composable
private fun rememberNotificationPermissionState(context: Context): Boolean {
    return if (Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU) {
        true
    } else {
        ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) == android.content.pm.PackageManager.PERMISSION_GRANTED
    }
}

@Composable
private fun rememberBatteryOptimizationState(context: Context): Boolean {
    val powerManager = context.getSystemService(Context.POWER_SERVICE) as PowerManager
    return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
        powerManager.isIgnoringBatteryOptimizations(context.packageName)
    } else {
        true
    }
}

@Composable
private fun rememberOverlayPermissionState(context: Context): Boolean {
    return hasOverlayPermission(context)
}

@Composable
private fun rememberAccessibilityAutoAnalyzeState(context: Context): Boolean {
    return AccessibilityAutoAnalyzeService.isEnabledInSystem(context)
}

private fun openBatteryOptimizationSettings(context: Context) {
    val intent = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
        Intent(Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS).apply {
            data = Uri.parse("package:${context.packageName}")
        }
    } else {
        Intent(Settings.ACTION_IGNORE_BATTERY_OPTIMIZATION_SETTINGS)
    }
    intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
    runCatching { context.startActivity(intent) }
}

private fun openAutoStartSettings(context: Context) {
    val fallbackIntent = Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS).apply {
        data = Uri.parse("package:${context.packageName}")
        addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
    }
    val candidates = listOf(
        Intent().setClassName("com.miui.securitycenter", "com.miui.permcenter.autostart.AutoStartManagementActivity"),
        Intent().setClassName("com.huawei.systemmanager", "com.huawei.systemmanager.startupmgr.ui.StartupNormalAppListActivity"),
        Intent().setClassName("com.coloros.safecenter", "com.coloros.safecenter.permission.startup.StartupAppListActivity"),
        Intent().setClassName("com.vivo.permissionmanager", "com.vivo.permissionmanager.activity.BgStartUpManagerActivity"),
    ).onEach { it.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK) }

    val launched = candidates.firstOrNull { candidate ->
        runCatching {
            context.packageManager.resolveActivity(candidate, 0) != null
        }.getOrDefault(false)
    }
    runCatching { context.startActivity(launched ?: fallbackIntent) }
}

private fun openOverlayPermissionSettings(context: Context) {
    val intent = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
        Intent(Settings.ACTION_MANAGE_OVERLAY_PERMISSION, Uri.parse("package:${context.packageName}"))
    } else {
        Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS).apply {
            data = Uri.parse("package:${context.packageName}")
        }
    }
    intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
    runCatching { context.startActivity(intent) }
}

private fun openAccessibilitySettings(context: Context) {
    val intent = Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS).apply {
        addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
    }
    runCatching { context.startActivity(intent) }
}

private fun hasOverlayPermission(context: Context): Boolean {
    return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
        Settings.canDrawOverlays(context)
    } else {
        true
    }
}

@Composable
private fun AppMessageSnackbar(
    data: SnackbarData,
    isError: Boolean,
) {
    val background = if (isError) Color(0xFFFFF1F2) else Color(0xFFF0FDF4)
    val border = if (isError) Color(0xFFFDA4AF) else Color(0xFFA7F3D0)
    val accent = if (isError) Color(0xFFDC2626) else Color(0xFF059669)
    val icon = if (isError) Icons.Outlined.ErrorOutline else Icons.Outlined.CheckCircle

    Card(
        shape = RoundedCornerShape(18.dp),
        colors = CardDefaults.cardColors(containerColor = background),
        modifier = Modifier
            .fillMaxWidth()
            .widthIn(max = 520.dp)
            .border(1.dp, border, RoundedCornerShape(18.dp)),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .size(32.dp)
                    .clip(CircleShape)
                    .background(accent.copy(alpha = 0.12f)),
                contentAlignment = Alignment.Center,
            ) {
                Icon(icon, contentDescription = null, tint = accent, modifier = Modifier.size(18.dp))
            }
            Text(
                text = data.visuals.message,
                color = Color(0xFF0F172A),
                fontSize = 14.sp,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.weight(1f),
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun SentinelTopBar(
    unreadCount: Int,
    onOpenAlerts: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .statusBarsPadding()
            .padding(horizontal = 20.dp, vertical = 16.dp),
    ) {
        Row(
            modifier = Modifier.align(Alignment.CenterStart),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            BrandShieldIconTile(modifier = Modifier.size(36.dp))
            Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Text(
                    text = "反诈卫士",
                    fontWeight = FontWeight.ExtraBold,
                    fontSize = 16.sp,
                    color = Color(0xFF0F172A),
                )
                Text(
                    text = "SENTINEL AI",
                    fontWeight = FontWeight.Black,
                    fontSize = 9.sp,
                    color = Color(0xFF94A3B8),
                    letterSpacing = 1.2.sp,
                )
            }
        }

        Box(
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .size(36.dp)
                .clip(CircleShape)
                .background(Color(0xFFF8FAFC))
                .border(1.dp, Color(0xFFF1F5F9), CircleShape)
                .clickable(onClick = onOpenAlerts),
            contentAlignment = Alignment.Center,
        ) {
            BadgedBox(
                badge = {
                    if (unreadCount > 0) {
                        Badge { Text(unreadCount.coerceAtMost(9).toString()) }
                    }
                }
            ) {
                LucideSvgIcon(
                    iconName = "bell",
                    contentDescription = "消息",
                    tint = Color(0xFF64748B),
                    size = 18.dp,
                )
            }
        }
    }
}

@Composable
private fun SentinelBottomBar(
    current: AppScreen,
    onScreenChange: (AppScreen) -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .navigationBarsPadding()
            .padding(start = 20.dp, end = 20.dp, top = 8.dp, bottom = 24.dp),
    ) {
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = Color.White.copy(alpha = 0.92f),
            shadowElevation = 18.dp,
            shape = RoundedCornerShape(16.dp),
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(76.dp)
                    .border(1.dp, Color.White.copy(alpha = 0.72f), RoundedCornerShape(16.dp))
                    .padding(horizontal = 24.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                BottomBarItem(
                    label = "首页",
                    iconName = "home",
                    selected = current in setOf(AppScreen.Dashboard, AppScreen.History, AppScreen.RiskTrend),
                    onClick = { onScreenChange(AppScreen.Dashboard) },
                )
                BottomBarItem(
                    label = "消息",
                    iconName = "message-square",
                    selected = current == AppScreen.Alerts,
                    onClick = { onScreenChange(AppScreen.Alerts) },
                )
                Spacer(modifier = Modifier.widthIn(min = 56.dp))
                BottomBarItem(
                    label = "家庭",
                    iconName = "users",
                    selected = current in setOf(AppScreen.Family, AppScreen.FamilyManage),
                    onClick = { onScreenChange(AppScreen.Family) },
                )
                BottomBarItem(
                    label = "我的",
                    iconName = "circle-user-round",
                    selected = current in setOf(AppScreen.Profile, AppScreen.ProfilePrivacy),
                    onClick = { onScreenChange(AppScreen.Profile) },
                )
            }
        }

        Box(
            modifier = Modifier
                .align(Alignment.Center)
                .size(56.dp)
                .clip(RoundedCornerShape(16.dp))
                .background(if (current == AppScreen.Submit) brandPrimary() else Color(0xFF0F172A))
                .clickable { onScreenChange(AppScreen.Submit) },
            contentAlignment = Alignment.Center,
        ) {
            LucideSvgIcon(
                iconName = "maximize",
                contentDescription = "检测",
                tint = Color.White,
                size = 24.dp,
            )
        }
    }
}

@Composable
private fun BottomBarItem(
    label: String,
    iconName: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    val contentColor = if (selected) brandPrimary() else Color(0xFF94A3B8)

    Column(
        modifier = Modifier.clickable(onClick = onClick),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(2.dp),
    ) {
        LucideSvgIcon(
            iconName = iconName,
            contentDescription = label,
            tint = contentColor,
            size = 20.dp,
        )
        Text(
            text = label,
            color = contentColor,
            fontSize = 10.sp,
            fontWeight = FontWeight.Bold,
        )
    }
}

private fun brandPrimary(): Color = Color(0xFF059669)

private fun screenTitle(screen: AppScreen): String = when (screen) {
    AppScreen.Dashboard, AppScreen.History, AppScreen.RiskTrend -> "概览"
    AppScreen.Alerts -> "消息"
    AppScreen.Submit -> "分析"
    AppScreen.SimulationQuiz -> "演练"
    AppScreen.Chat -> "AI 助手"
    AppScreen.Family, AppScreen.FamilyManage -> "家庭"
    AppScreen.Profile, AppScreen.ProfilePrivacy -> "我的"
}
