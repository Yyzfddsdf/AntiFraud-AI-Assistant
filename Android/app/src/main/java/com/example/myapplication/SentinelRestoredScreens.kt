package com.example.myapplication

import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawing
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.windowInsetsBottomHeight
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.ArrowBack
import androidx.compose.material.icons.outlined.DeleteOutline
import androidx.compose.material.icons.outlined.Image
import androidx.compose.material.icons.outlined.KeyboardArrowRight
import androidx.compose.material.icons.outlined.Mic
import androidx.compose.material.icons.outlined.Security
import androidx.compose.material.icons.outlined.ShowChart
import androidx.compose.material.icons.outlined.Videocam
import androidx.compose.material3.AssistChip
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import java.time.Instant
import java.time.ZoneId
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

val MobileSurface = Color(0xFFF8FAFC)
val MobileBorder = Color(0xFFE2E8F0)
val MobileText = Color(0xFF0F172A)
val MobileSubtle = Color(0xFF64748B)
val MobileGreen = Color(0xFF10B981)
val MobileTeal = Color(0xFF14B8A6)
val MobileRose = Color(0xFFEF4444)

@Composable
private fun ScreenEnterItem(
    visible: Boolean,
    index: Int,
    content: @Composable () -> Unit,
) {
    AnimatedVisibility(
        visible = visible,
        enter = fadeIn(
            animationSpec = tween(
                durationMillis = 360,
                delayMillis = index * 70,
            ),
        ) + slideInVertically(
            animationSpec = tween(
                durationMillis = 420,
                delayMillis = index * 70,
            ),
            initialOffsetY = { fullHeight -> fullHeight / 6 },
        ),
    ) {
        content()
    }
}

@Composable
fun MobileDashboardScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
) {
    val todayHistory = remember(state.history) {
        val today = Instant.now().atZone(ZoneId.systemDefault()).toLocalDate()
        state.history.filter { parseInstant(it.created_at).atZone(ZoneId.systemDefault()).toLocalDate() == today }
    }
    val recentOngoingTasks = remember(state.tasks) {
        state.tasks.filter { it.status == "pending" || it.status == "processing" }.take(3)
    }
    val overallTrendLabel = homeOverallTrendLabel(state.riskData?.analysis?.overall_trend)
    val overallTrendColor = homeOverallTrendColor(state.riskData?.analysis?.overall_trend)
    val currentRiskLabel = homeCurrentRiskLabel(state.riskData?.stats)
    val currentRiskColor = homeCurrentRiskColor(currentRiskLabel)

    LazyColumn(
        modifier = Modifier.fillMaxSize().background(Color.White),
        contentPadding = PaddingValues(
            start = 20.dp,
            end = 20.dp,
            top = padding.calculateTopPadding() + 16.dp,
            bottom = padding.calculateBottomPadding() + 36.dp,
        ),
        verticalArrangement = Arrangement.spacedBy(20.dp),
    ) {
        item {
            SecurityOverviewCard(totalCount = state.riskData?.stats?.total ?: state.history.size)
        }
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                DashboardAction("AI助手", "bot", MobileGreen, Modifier.weight(1f)) {
                    viewModel.openScreen(AppScreen.Chat)
                }
                DashboardAction("历史", "calendar", Color(0xFF2563EB), Modifier.weight(1f)) {
                    viewModel.openScreen(AppScreen.History)
                }
                DashboardAction("趋势", "line-chart", Color(0xFFD97706), Modifier.weight(1f)) {
                    viewModel.openScreen(AppScreen.RiskTrend)
                }
                DashboardAction("守护", "heart", Color(0xFF4F46E5), Modifier.weight(1f)) {
                    viewModel.openScreen(AppScreen.Family)
                }
                DashboardAction("演练", "play-circle", Color(0xFFE11D48), Modifier.weight(1f)) {
                    viewModel.openScreen(AppScreen.SimulationQuiz)
                }
            }
        }
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                MetricCard("今日检测数量", todayHistory.size.toString(), MobileGreen, "search", Modifier.weight(1f))
                MetricCard("今日高危预警", todayHistory.count { normalizeRiskLevel(it.risk_level) == "高" }.toString(), MobileRose, "alert-octagon", Modifier.weight(1f))
            }
        }
        item {
            SectionTitle(title = "风险趋势", action = "详情 >") { viewModel.openScreen(AppScreen.RiskTrend) }
        }
        item {
            AiInsightCard(
                headline = currentInsightHeadline(state.riskData?.analysis),
                body = state.riskData?.analysis?.summary.orEmpty().ifBlank { "近期暂无风险分析摘要" },
                overallLabel = overallTrendLabel,
                overallColor = overallTrendColor,
                currentRiskLabel = currentRiskLabel,
                currentRiskColor = currentRiskColor,
            )
        }
        item {
            SectionTitle(title = "最近任务", action = "全部", actionColor = Color(0xFF94A3B8)) { viewModel.openScreen(AppScreen.History) }
        }
        if (recentOngoingTasks.isEmpty()) {
            item { HomeRecentTasksEmptyCard() }
        } else {
            items(recentOngoingTasks, key = { it.task_id }) { task ->
                MobileTaskRow(task = task) { viewModel.openTaskDetail(task.task_id) }
            }
        }
    }
}

@Composable
fun MobileHistoryScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
) {
    var contentVisible by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) {
        contentVisible = true
    }
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Brush.verticalGradient(listOf(Color(0xFFF8FAFC), Color(0xFFFBFDFF))))
            .padding(top = padding.calculateTopPadding()),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White.copy(alpha = 0.92f)),
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                IconButton(onClick = { viewModel.openScreen(AppScreen.Dashboard) }) {
                    Icon(Icons.Outlined.ArrowBack, contentDescription = "返回", tint = MobileText)
                }
                Text(
                    "历史档案",
                    modifier = Modifier.weight(1f),
                    textAlign = TextAlign.Center,
                    fontSize = 17.sp,
                    fontWeight = FontWeight.Bold,
                    color = MobileText,
                )
                Box(
                    modifier = Modifier
                        .background(Color(0xFFF1F5F9), RoundedCornerShape(999.dp))
                        .padding(horizontal = 8.dp, vertical = 5.dp),
                ) {
                    Text(state.history.size.toString(), fontSize = 10.sp, fontWeight = FontWeight.Black, color = Color(0xFF64748B))
                }
            }
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(1.dp)
                    .background(MobileBorder.copy(alpha = 0.55f)),
            )
        }
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(start = 14.dp, end = 14.dp, top = 12.dp, bottom = padding.calculateBottomPadding() + 20.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            item {
                ScreenEnterItem(visible = contentVisible, index = 0) {
                    HistoryOverviewCard(state.history)
                }
            }
            if (state.history.isEmpty()) {
                item {
                    ScreenEnterItem(visible = contentVisible, index = 1) {
                        HistoryEmptyStateCard()
                    }
                }
            } else {
                itemsIndexed(state.history, key = { _, item -> item.record_id }) { index, item ->
                    ScreenEnterItem(visible = contentVisible, index = index + 1) {
                        HistoryArchiveCard(
                            item = item,
                            deleting = state.deletingHistoryIds.contains(item.record_id),
                            onOpen = { viewModel.openTaskDetail(item.record_id) },
                            onDelete = { viewModel.deleteHistoryRecord(item.record_id) },
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun MobileRiskTrendScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
) {
    var contentVisible by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) {
        contentVisible = true
    }
    val rows = remember(state.riskData) { recentTrendRows(state.riskData?.trend.orEmpty(), 7) }
    val recentTrendCards = remember(rows) { rows.take(3) }
    val latestRow = rows.firstOrNull()
    val previousRow = rows.getOrNull(1)
    val regionSignal = state.regionCaseStats?.let(::regionSignalSummary)
    val heroTitle = state.riskData?.analysis?.let(::currentInsightHeadline)
        ?: regionSignal?.title
        ?: "AI 研判：风险走势仍需持续观察"
    val heroDetail = state.riskData?.analysis?.summary.orEmpty().ifBlank {
        regionSignal?.detail ?: "后端趋势分析暂未返回摘要，请继续关注最新检测结果。"
    }
    val heroFooter = state.regionCaseStats?.let(::regionTopScamHint)?.removePrefix("最近高发：")
        ?: "暂未形成明显高发类型"
    val totalDelta = buildTrendDeltaText(latestRow?.total ?: 0, previousRow?.total ?: 0)
    val highDelta = buildTrendDeltaText(latestRow?.high ?: 0, previousRow?.high ?: 0)
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MobileSurface)
            .padding(top = padding.calculateTopPadding()),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White.copy(alpha = 0.92f)),
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                IconButton(onClick = { viewModel.openScreen(AppScreen.Dashboard) }) {
                    Icon(Icons.Outlined.ArrowBack, contentDescription = "返回", tint = MobileText)
                }
                Text(
                    "风险趋势分析",
                    modifier = Modifier.weight(1f),
                    textAlign = TextAlign.Center,
                    fontSize = 17.sp,
                    fontWeight = FontWeight.Bold,
                    color = MobileText,
                )
                Spacer(Modifier.width(40.dp))
            }
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(1.dp)
                    .background(MobileBorder.copy(alpha = 0.55f)),
            )
        }
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(start = 14.dp, end = 14.dp, top = 16.dp, bottom = padding.calculateBottomPadding() + 12.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            item {
                ScreenEnterItem(visible = contentVisible, index = 0) {
                    RiskHeroCard(
                        title = heroTitle,
                        detail = heroDetail,
                        footer = heroFooter,
                    )
                }
            }
            item {
                ScreenEnterItem(visible = contentVisible, index = 1) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
                        RiskSummaryFeatureCard(
                            label = "总检测次数",
                            value = state.riskData?.stats?.total ?: 0,
                            deltaText = totalDelta,
                            deltaColor = Color(0xFF10B981),
                            accent = MobileGreen,
                            icon = Icons.Outlined.ShowChart,
                            modifier = Modifier.weight(1f),
                        )
                        RiskSummaryFeatureCard(
                            label = "高风险预警",
                            value = state.riskData?.stats?.high ?: 0,
                            deltaText = highDelta,
                            deltaColor = Color(0xFFFB7185),
                            accent = MobileRose,
                            icon = Icons.Outlined.Security,
                            modifier = Modifier.weight(1f),
                        )
                    }
                }
            }
            state.regionCaseStats?.let { stats ->
                item {
                    ScreenEnterItem(visible = contentVisible, index = 2) {
                        RegionStatsCard(stats)
                    }
                }
            }
            item {
                ScreenEnterItem(visible = contentVisible, index = 3) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            "近期检测走势",
                            fontSize = 22.sp,
                            fontWeight = FontWeight.Black,
                            color = MobileText,
                            modifier = Modifier.weight(1f),
                        )
                        Box(
                            modifier = Modifier
                                .background(Color(0x1A6366F1), RoundedCornerShape(999.dp))
                                .padding(horizontal = 10.dp, vertical = 6.dp),
                        ) {
                            Text("近7天", fontSize = 11.sp, fontWeight = FontWeight.Black, color = Color(0xFF5B4CF0))
                        }
                    }
                }
            }
            if (rows.isEmpty()) {
                item {
                    ScreenEnterItem(visible = contentVisible, index = 4) {
                        RiskTrendEmptyCard()
                    }
                }
            } else {
                itemsIndexed(recentTrendCards, key = { _, row -> row.time_bucket }) { index, row ->
                    ScreenEnterItem(visible = contentVisible, index = index + 4) {
                        TrendRowCard(row)
                    }
                }
            }
        }
    }
}

@Composable
fun MobileSubmitScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
    onPickImages: () -> Unit,
    onPickAudios: () -> Unit,
    onPickVideos: () -> Unit,
) {
    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(MobileSurface),
        contentPadding = PaddingValues(
            start = 16.dp,
            end = 16.dp,
            top = padding.calculateTopPadding() + 16.dp,
            bottom = padding.calculateBottomPadding() + 28.dp,
        ),
        verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        item {
            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Text("智能检测", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.ExtraBold, color = MobileText)
                Text("提交可疑信息，AI 护航实时为您排查风险", style = MaterialTheme.typography.bodySmall, color = MobileSubtle)
            }
        }
        item {
            Card(
                shape = RoundedCornerShape(24.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
            ) {
                Column(
                    modifier = Modifier.padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(14.dp),
                ) {
                    Text("可疑内容描述", fontWeight = FontWeight.Bold, color = MobileText)
                    OutlinedTextField(
                        value = state.analyzeForm.text,
                        onValueChange = viewModel::updateAnalyzeText,
                        modifier = Modifier.fillMaxWidth().heightIn(min = 140.dp),
                        placeholder = { Text("请输入聊天记录、短信、链接或事件描述") },
                        shape = RoundedCornerShape(20.dp),
                    )
                    Text("上传附件", fontWeight = FontWeight.Bold, color = MobileText)
                    Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                        UploadActionTile("图片", state.analyzeForm.images.size, Icons.Outlined.Image, Modifier.weight(1f), onPickImages)
                        UploadActionTile("音频", state.analyzeForm.audios.size, Icons.Outlined.Mic, Modifier.weight(1f), onPickAudios)
                        UploadActionTile("视频", state.analyzeForm.videos.size, Icons.Outlined.Videocam, Modifier.weight(1f), onPickVideos)
                    }
                }
            }
        }
        item {
            Button(
                onClick = viewModel::submitAnalysis,
                modifier = Modifier.fillMaxWidth().height(56.dp),
                shape = RoundedCornerShape(18.dp),
                colors = ButtonDefaults.buttonColors(containerColor = MobileText),
                enabled = !state.analyzing,
            ) {
                Text(if (state.analyzing) "正在深度分析..." else "开始全面检测", fontWeight = FontWeight.Bold)
            }
        }
    }
}

@Composable
fun MobileAlertsScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
) {
    val alerts = remember(state.alertItems) { combinedAlerts(state) }
    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(MobileSurface),
        contentPadding = PaddingValues(
            start = 16.dp,
            end = 16.dp,
            top = padding.calculateTopPadding() + 12.dp,
            bottom = padding.calculateBottomPadding() + 28.dp,
        ),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            Column(verticalArrangement = Arrangement.spacedBy(4.dp), modifier = Modifier.padding(horizontal = 4.dp, vertical = 4.dp)) {
                Text("消息中心", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.ExtraBold, color = MobileText)
                Text(
                    "您有 ${state.alertItems.count { !it.read }} 条未读风险预警",
                    style = MaterialTheme.typography.bodySmall,
                    color = MobileSubtle,
                )
            }
        }
        if (alerts.isEmpty()) {
            item { EmptyCard("暂无风险预警") }
        } else {
            items(alerts, key = { it.record_id }) { item ->
                AlertListCard(item = item) { viewModel.openAlertTask(item.record_id) }
            }
        }
    }
}

@Composable
fun MobileSimulationQuizScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
) {
    if (state.simulationViewMode == SimulationViewMode.Exam) {
        MobileSimulationExamScreen(state = state, viewModel = viewModel)
        return
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MobileSurface)
            .padding(top = padding.calculateTopPadding()),
    ) {
        MobileBackHeader(title = "诈骗演练") { viewModel.openScreen(AppScreen.Dashboard) }
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 12.dp, bottom = padding.calculateBottomPadding() + 24.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            item {
                Card(
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                ) {
                    Column(
                        modifier = Modifier.padding(16.dp),
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
                        Text("定制演练场景", fontWeight = FontWeight.ExtraBold, color = MobileText)
                        OutlinedTextField(
                            value = state.simulationForm.caseType,
                            onValueChange = viewModel::updateSimulationCaseType,
                            modifier = Modifier.fillMaxWidth(),
                            placeholder = { Text("场景，例如冒充公检法、投资荐股") },
                            shape = RoundedCornerShape(18.dp),
                        )
                        OutlinedTextField(
                            value = state.simulationForm.targetPersona,
                            onValueChange = viewModel::updateSimulationTargetPersona,
                            modifier = Modifier.fillMaxWidth(),
                            placeholder = { Text("目标身份，例如老人、学生") },
                            shape = RoundedCornerShape(18.dp),
                        )
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                            DifficultyChip("简单", "easy", state.simulationForm.difficulty, Modifier.weight(1f), viewModel::updateSimulationDifficulty)
                            DifficultyChip("中等", "medium", state.simulationForm.difficulty, Modifier.weight(1f), viewModel::updateSimulationDifficulty)
                            DifficultyChip("困难", "hard", state.simulationForm.difficulty, Modifier.weight(1f), viewModel::updateSimulationDifficulty)
                        }
                        Button(
                            onClick = viewModel::generateSimulationPack,
                            modifier = Modifier.fillMaxWidth().height(50.dp),
                            shape = RoundedCornerShape(18.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = MobileText),
                            enabled = !state.simulationGenerating,
                        ) {
                            Text(if (state.simulationGenerating) "正在生成..." else "生成专属演练", fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
            item {
                SectionTitle(title = "待挑战题包", action = "刷新") { viewModel.fetchSimulationPacks(silent = false) }
            }
            if (state.simulationPackList.isEmpty()) {
                item { EmptyCard("暂无可挑战的演练题包") }
            } else {
                items(state.simulationPackList, key = { "pack-${it.pack_id}" }) { pack ->
                    SimulationPackCard(pack = pack, onStart = { viewModel.startSimulationSession(pack.pack_id) })
                }
            }
            item {
                SectionTitle(title = "演练记录", action = "刷新") { viewModel.fetchSimulationSessions(silent = false) }
            }
            if (state.simulationSessionList.isEmpty()) {
                item { EmptyCard("暂无演练记录") }
            } else {
                items(state.simulationSessionList, key = { "session-${it.pack_id}" }) { session ->
                    SimulationSessionCard(
                        session = session,
                        deleting = state.deletingSimulationSessionIds.contains(session.pack_id),
                        onResume = { viewModel.startSimulationSession(session.pack_id) },
                        onDelete = { viewModel.deleteSimulationSession(session.pack_id) },
                    )
                }
            }
        }
    }
}

@Composable
private fun MobileSimulationExamScreen(
    state: MainUiState,
    viewModel: SentinelViewModel,
) {
    val progress by animateFloatAsState(targetValue = examProgress(state), label = "simulationProgress")
    val coroutineScope = rememberCoroutineScope()
    var selectedOptionKey by remember { mutableStateOf("") }
    BackHandler(enabled = true, onBack = viewModel::closeSimulationExamView)
    LaunchedEffect(state.simulationCurrentStep?.step_id, state.simulationStatus) {
        selectedOptionKey = ""
    }
    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MobileSurface),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 16.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                IconButton(onClick = viewModel::closeSimulationExamView) {
                    Icon(Icons.Outlined.ArrowBack, contentDescription = "返回")
                }
                Text("模拟演练", fontWeight = FontWeight.ExtraBold, color = MobileText, modifier = Modifier.weight(1f))
                Text("${state.simulationCurrentScore}", fontWeight = FontWeight.ExtraBold, color = MobileGreen)
            }
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(8.dp)
                    .background(Color(0xFFE2E8F0), RoundedCornerShape(999.dp)),
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth(progress)
                        .fillMaxHeight()
                        .background(MobileGreen, RoundedCornerShape(999.dp)),
                )
            }
        }
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            if (state.simulationStatus == "completed" && state.simulationResult != null) {
                item { SimulationResultCard(result = state.simulationResult, onBack = viewModel::closeSimulationExamView) }
            } else {
                state.simulationPack?.takeIf { state.simulationAnswers.isEmpty() }?.let { pack ->
                    item { IntroExamCard(pack = pack) }
                }
                item {
                    AnimatedContent(
                        targetState = state.simulationCurrentStep?.step_id ?: "simulation-empty",
                        transitionSpec = {
                            (fadeIn() + slideInVertically { fullHeight -> fullHeight / 5 }) togetherWith
                                (fadeOut() + slideOutVertically { fullHeight -> -fullHeight / 10 })
                        },
                        label = "simulationStepTransition",
                    ) {
                        val step = state.simulationCurrentStep
                        if (step == null) {
                            EmptyCard("当前暂无可继续的题目")
                        } else {
                            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                                ScenarioCard(step)
                                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                                    step.options.forEach { option ->
                                        AnswerOptionCard(
                                            option = option,
                                            enabled = !state.simulationSubmitting && state.simulationStatus == "in_progress",
                                            selected = selectedOptionKey == option.key,
                                        ) {
                                            if (state.simulationSubmitting || selectedOptionKey.isNotBlank()) return@AnswerOptionCard
                                            coroutineScope.launch {
                                                selectedOptionKey = option.key
                                                delay(140)
                                                viewModel.submitSimulationAnswer(option.key)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun MobileChatScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
    onPickImages: () -> Unit,
) {
    val listState = rememberLazyListState()
    val lastMessage = state.chatMessages.lastOrNull()
    val hasComposerContent = state.chatInput.isNotBlank() || state.chatImages.isNotEmpty()
    val uploadAlpha by animateFloatAsState(targetValue = if (hasComposerContent) 0f else 1f, label = "chatUploadAlpha")
    val inputLeadingPadding by animateDpAsState(targetValue = if (hasComposerContent) 18.dp else 52.dp, label = "chatInputLeadingPadding")
    val sendScale by animateFloatAsState(targetValue = if (hasComposerContent && !state.isChatting) 1f else 0.25f, label = "chatSendScale")
    val sendAlpha by animateFloatAsState(targetValue = if (hasComposerContent && !state.isChatting) 1f else 0f, label = "chatSendAlpha")
    val sendEnabled = hasComposerContent && !state.isChatting
    val currentSendChat by rememberUpdatedState(viewModel::sendChatMessage)

    LaunchedEffect(
        state.chatMessages.size,
        lastMessage?.id,
        lastMessage?.content?.length,
        state.isChatting,
    ) {
        if (state.chatMessages.isNotEmpty()) {
            listState.scrollToItem(state.chatMessages.lastIndex)
            delay(48)
            listState.scrollToItem(state.chatMessages.lastIndex)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White)
            .padding(top = padding.calculateTopPadding()),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 16.dp, vertical = 10.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .size(32.dp)
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                    ) { viewModel.openScreen(AppScreen.Dashboard) },
                contentAlignment = Alignment.Center,
            ) {
                LucideSvgIcon(
                    iconName = "arrow-left",
                    contentDescription = "返回首页",
                    tint = Color(0xFF64748B),
                    size = 18.dp,
                )
            }
            Column(modifier = Modifier.weight(1f), horizontalAlignment = Alignment.CenterHorizontally) {
                Text(
                    "用户问题助手回应",
                    fontSize = 9.sp,
                    color = Color(0xFF94A3B8),
                    fontWeight = FontWeight.Bold,
                    letterSpacing = 1.2.sp,
                )
                Text("Sentinel AI", fontWeight = FontWeight.Bold, color = Color(0xFF1E293B), fontSize = 12.sp)
            }
            Box(
                modifier = Modifier
                    .size(32.dp)
                    .clickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                    ) { viewModel.clearChatHistory() },
                contentAlignment = Alignment.Center,
            ) {
                LucideSvgIcon(
                    iconName = "trash-2",
                    contentDescription = "清空对话",
                    tint = Color(0xFF64748B),
                    size = 18.dp,
                )
            }
        }
        LazyColumn(
            modifier = Modifier.weight(1f),
            state = listState,
            contentPadding = PaddingValues(horizontal = 16.dp, vertical = 16.dp),
            verticalArrangement = Arrangement.spacedBy(18.dp),
        ) {
            items(state.chatMessages, key = { it.id }) { message ->
                MobileChatBubble(message = message)
            }
            if (state.isChatting) {
                item {
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Start) {
                        Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                            Box(modifier = Modifier.size(8.dp).background(Color(0xFF94A3B8), CircleShape))
                            Box(modifier = Modifier.size(8.dp).background(Color(0xFF94A3B8), CircleShape))
                            Box(modifier = Modifier.size(8.dp).background(Color(0xFF94A3B8), CircleShape))
                        }
                    }
                }
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(horizontal = 16.dp, vertical = 10.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            if (state.chatImages.isNotEmpty()) {
                LazyRow(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    items(state.chatImages.size) { index ->
                        val image = state.chatImages[index]
                        Box(
                            modifier = Modifier
                                .size(64.dp)
                                .clip(RoundedCornerShape(16.dp))
                                .border(1.dp, MobileBorder, RoundedCornerShape(16.dp)),
                        ) {
                            Base64Thumbnail(dataUrl = image, size = 64.dp)
                            Box(
                                modifier = Modifier
                                    .align(Alignment.TopEnd)
                                    .padding(4.dp)
                                    .size(18.dp)
                                    .background(Color(0x990F172A), CircleShape)
                                    .clickable { viewModel.removeChatImage(index) },
                                contentAlignment = Alignment.Center,
                            ) {
                                Text("×", color = Color.White, fontSize = 10.sp, fontWeight = FontWeight.Bold)
                            }
                        }
                    }
                }
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 2.dp),
            ) {
                Box(
                    modifier = Modifier
                        .align(Alignment.CenterStart)
                        .padding(start = 4.dp)
                        .alpha(uploadAlpha)
                        .clickable(enabled = !state.isChatting) { onPickImages() },
                    contentAlignment = Alignment.Center,
                ) {
                    LucideSvgIcon(
                        iconName = "image",
                        contentDescription = "添加图片",
                        tint = Color(0xFFAAAAAA),
                        size = 22.dp,
                    )
                }

                BasicTextField(
                    value = state.chatInput,
                    onValueChange = viewModel::updateChatInput,
                    modifier = Modifier
                        .fillMaxWidth()
                        .heightIn(min = 50.dp)
                        .padding(start = inputLeadingPadding, end = 0.dp),
                    textStyle = androidx.compose.ui.text.TextStyle(
                        color = Color(0xFF4C4C4C),
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Medium,
                    ),
                    enabled = !state.isChatting,
                    maxLines = 4,
                    keyboardOptions = KeyboardOptions(
                        keyboardType = KeyboardType.Text,
                        imeAction = ImeAction.Send,
                    ),
                    keyboardActions = KeyboardActions(
                        onSend = {
                            if (sendEnabled) currentSendChat()
                        },
                    ),
                    decorationBox = { innerTextField ->
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(Color(0xFFE9E9E9), RoundedCornerShape(50.dp))
                                .padding(start = 16.dp, end = 48.dp, top = 12.dp, bottom = 12.dp),
                        ) {
                            if (state.chatInput.isBlank()) {
                                Text("Ask Anything...", color = Color(0xFF959595), fontSize = 14.sp, fontWeight = FontWeight.Medium)
                            }
                            innerTextField()
                        }
                    },
                )

                Box(
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .padding(end = 4.dp)
                        .alpha(sendAlpha)
                        .size(36.dp)
                        .scale(sendScale)
                        .clip(CircleShape)
                        .background(Brush.linearGradient(listOf(Color(0xFF9147FF), Color(0xFFFF4141))))
                        .clickable(enabled = sendEnabled) { currentSendChat() },
                    contentAlignment = Alignment.Center,
                ) {
                    LucideSvgIcon(
                        iconName = "send",
                        contentDescription = "发送",
                        tint = Color(0xFFE9E9E9),
                        size = 18.dp,
                    )
                }
            }

            Text(
                "内容由 AI 生成，请仔细甄别",
                modifier = Modifier.fillMaxWidth(),
                textAlign = TextAlign.Center,
                color = Color(0xFFC0C7D2),
                fontSize = 9.sp,
            )
            Spacer(modifier = Modifier.windowInsetsBottomHeight(WindowInsets.safeDrawing))
        }
    }
}

@Composable
fun MobileFamilyScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
) {
    var isMembersExpanded by remember { mutableStateOf(false) }
    val membersArrowRotation by animateFloatAsState(if (isMembersExpanded) 180f else 0f, label = "familyMembersArrow")

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.White)
            .padding(top = padding.calculateTopPadding()),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White.copy(alpha = 0.82f))
                .padding(vertical = 12.dp),
            contentAlignment = Alignment.Center,
        ) {
            Text("家庭守护", fontSize = 16.sp, fontWeight = FontWeight.Bold, color = Color(0xFF1E293B))
        }

        if (state.familyOverview.family == null) {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 16.dp, bottom = padding.calculateBottomPadding() + 24.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                if (state.loading && state.familyReceivedInvitations.isEmpty()) {
                    item {
                        Card(
                            shape = RoundedCornerShape(20.dp),
                            colors = CardDefaults.cardColors(containerColor = Color(0xFFF9FAFB)),
                            border = BorderStroke(1.dp, Color(0x12CBD5E1)),
                        ) {
                            Column(
                                modifier = Modifier.fillMaxWidth().padding(vertical = 24.dp),
                                horizontalAlignment = Alignment.CenterHorizontally,
                                verticalArrangement = Arrangement.spacedBy(8.dp),
                            ) {
                                Text("正在加载数据...", fontSize = 12.sp, color = Color(0xFF94A3B8), fontWeight = FontWeight.Medium)
                            }
                        }
                    }
                }

                item {
                    Column(horizontalAlignment = Alignment.CenterHorizontally, modifier = Modifier.fillMaxWidth()) {
                        Box(
                            modifier = Modifier
                                .size(56.dp)
                                .background(Color(0xFFF9FAFB), RoundedCornerShape(16.dp))
                                .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(16.dp)),
                            contentAlignment = Alignment.Center,
                        ) {
                            LucideSvgIcon(iconName = "users", contentDescription = null, tint = Color(0xFF334155), size = 24.dp)
                        }
                    }
                }

                item {
                    Card(
                        shape = RoundedCornerShape(20.dp),
                        colors = CardDefaults.cardColors(containerColor = Color(0xFFF9FAFB)),
                        border = BorderStroke(1.dp, Color(0x12CBD5E1)),
                    ) {
                        Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text("收到的邀请", fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 15.sp, modifier = Modifier.weight(1f))
                                Box(
                                    modifier = Modifier
                                        .background(Color(0x80E5E7EB), RoundedCornerShape(8.dp))
                                        .padding(horizontal = 8.dp, vertical = 3.dp),
                                ) {
                                    Text(state.familyReceivedInvitations.size.toString(), fontSize = 10.sp, color = Color(0xFF64748B), fontWeight = FontWeight.Medium)
                                }
                            }
                            if (state.familyReceivedInvitations.isEmpty()) {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color.White, RoundedCornerShape(16.dp))
                                        .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(16.dp))
                                        .padding(vertical = 24.dp),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Text("暂无邀请", fontSize = 11.sp, color = Color(0xFF94A3B8))
                                }
                            } else {
                                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                                    state.familyReceivedInvitations.forEach { invitation ->
                                        FamilyInvitationCard(
                                            invitation = invitation,
                                            accepting = state.acceptingInvitationIds.contains(invitation.id),
                                        ) {
                                            viewModel.acceptFamilyInvitation(invitation.invite_code, invitation.id)
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                item {
                    Card(
                        shape = RoundedCornerShape(20.dp),
                        colors = CardDefaults.cardColors(containerColor = Color(0xFFF9FAFB)),
                        border = BorderStroke(1.dp, Color(0x12CBD5E1)),
                    ) {
                        Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                            Text("创建新家庭", fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 15.sp)
                            OutlinedTextField(
                                value = state.familyCreateName,
                                onValueChange = viewModel::updateFamilyCreateName,
                                modifier = Modifier.fillMaxWidth(),
                                placeholder = { Text("输入家庭名称") },
                                shape = RoundedCornerShape(12.dp),
                            )
                            Button(
                                onClick = viewModel::createFamily,
                                modifier = Modifier.fillMaxWidth().height(48.dp),
                                shape = RoundedCornerShape(12.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF0F172A), contentColor = Color.White),
                            ) {
                                Text("创建", fontWeight = FontWeight.Bold, fontSize = 13.sp)
                            }
                        }
                    }
                }

                item {
                    Card(
                        shape = RoundedCornerShape(20.dp),
                        colors = CardDefaults.cardColors(containerColor = Color(0xFFF9FAFB)),
                        border = BorderStroke(1.dp, Color(0x12CBD5E1)),
                    ) {
                        Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            Text("邀请码加入", fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 15.sp)
                            Text("输入家人发来的邀请码，快速加入", fontSize = 11.sp, color = Color(0xFF94A3B8))
                            OutlinedTextField(
                                value = state.familyAcceptCode,
                                onValueChange = viewModel::updateFamilyAcceptCode,
                                modifier = Modifier.fillMaxWidth(),
                                placeholder = { Text("输入邀请码") },
                                shape = RoundedCornerShape(12.dp),
                            )
                            Button(
                                onClick = { viewModel.acceptFamilyInvitation() },
                                modifier = Modifier.fillMaxWidth().height(48.dp),
                                shape = RoundedCornerShape(12.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF0F172A), contentColor = Color.White),
                            ) {
                                Text("加入", fontWeight = FontWeight.Bold, fontSize = 13.sp)
                            }
                        }
                    }
                }
            }
        } else {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 16.dp, bottom = padding.calculateBottomPadding() + 24.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                item { FamilySummaryHero(state) { viewModel.openScreen(AppScreen.FamilyManage) } }

                item {
                    Card(
                        shape = RoundedCornerShape(24.dp),
                        colors = CardDefaults.cardColors(containerColor = Color(0xFFF0FCFA)),
                        border = BorderStroke(1.dp, Color(0xFFE0F8F5)),
                    ) {
                        Column(modifier = Modifier.padding(8.dp)) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clickable { isMembersExpanded = !isMembersExpanded }
                                    .padding(horizontal = 8.dp, vertical = 8.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(12.dp),
                            ) {
                                Box(
                                    modifier = Modifier
                                        .size(36.dp)
                                        .background(Color(0xFFDFF7F4), RoundedCornerShape(12.dp)),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    LucideSvgIcon(iconName = "users", contentDescription = null, tint = Color(0xFF14B8A6), size = 16.dp)
                                }
                                Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                    Text("成员列表", fontSize = 14.sp, fontWeight = FontWeight.ExtraBold, color = MobileText)
                                    Text("共 ${state.familyOverview.members.size} 名成员", fontSize = 10.sp, color = Color(0xFF64748B))
                                }
                                Box(
                                    modifier = Modifier
                                        .size(28.dp)
                                        .background(Color.White, CircleShape),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Icon(
                                        Icons.Outlined.KeyboardArrowRight,
                                        contentDescription = null,
                                        tint = Color(0xFF94A3B8),
                                        modifier = Modifier.rotate(membersArrowRotation).size(16.dp),
                                    )
                                }
                            }
                            AnimatedVisibility(visible = isMembersExpanded) {
                                Column(
                                    modifier = Modifier.padding(start = 8.dp, end = 8.dp, bottom = 8.dp, top = 4.dp),
                                    verticalArrangement = Arrangement.spacedBy(8.dp),
                                ) {
                                    if (state.familyOverview.members.isEmpty()) {
                                        Box(
                                            modifier = Modifier
                                                .fillMaxWidth()
                                                .background(Color.White, RoundedCornerShape(16.dp))
                                                .padding(vertical = 20.dp),
                                            contentAlignment = Alignment.Center,
                                        ) {
                                            Text("暂无家庭成员", fontSize = 11.sp, color = Color(0xFF94A3B8))
                                        }
                                    } else {
                                        state.familyOverview.members.forEach { member ->
                                            FamilyMemberCard(member)
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                item {
                    Card(
                        shape = RoundedCornerShape(24.dp),
                        colors = CardDefaults.cardColors(containerColor = Color(0xFFF9FAFB)),
                        border = BorderStroke(1.dp, Color(0x12CBD5E1)),
                    ) {
                        Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text("守护关系", fontSize = 15.sp, fontWeight = FontWeight.ExtraBold, color = MobileText, modifier = Modifier.weight(1f))
                                Box(
                                    modifier = Modifier
                                        .background(Color.White, RoundedCornerShape(8.dp))
                                        .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(8.dp))
                                        .padding(horizontal = 8.dp, vertical = 3.dp),
                                ) {
                                    Text(state.familyOverview.guardian_links.size.toString(), fontSize = 10.sp, color = Color(0xFF64748B))
                                }
                            }
                            if (state.familyOverview.guardian_links.isEmpty()) {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color.White, RoundedCornerShape(16.dp))
                                        .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(16.dp))
                                        .padding(vertical = 24.dp),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Text("暂无守护关系", fontSize = 11.sp, color = Color(0xFF94A3B8))
                                }
                            } else {
                                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                                    state.familyOverview.guardian_links.forEach { link ->
                                        GuardianRelationCard(link)
                                    }
                                }
                            }
                        }
                    }
                }

                item {
                    Card(
                        shape = RoundedCornerShape(24.dp),
                        colors = CardDefaults.cardColors(containerColor = Color(0xFFF9FAFB)),
                        border = BorderStroke(1.dp, Color(0x12CBD5E1)),
                    ) {
                        Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                            Text("最新动态", fontSize = 15.sp, fontWeight = FontWeight.ExtraBold, color = MobileText)
                            if (state.familyNotifications.isEmpty()) {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color.White, RoundedCornerShape(16.dp))
                                        .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(16.dp))
                                        .padding(vertical = 24.dp),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Text("无最新动态", fontSize = 11.sp, color = Color(0xFF94A3B8))
                                }
                            } else {
                                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                                    state.familyNotifications.forEach { note ->
                                        FamilyNotificationCard(note)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun MobileFamilyManageScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
) {
    val roleOptions = listOf(
        DropdownOption("member", "普通成员", "家庭中的普通成员"),
        DropdownOption("guardian", "守护人", "会接收高风险提醒并协助处理"),
    )
    val guardianOptions = state.familyOverview.members
        .filter { it.role == "owner" || it.role == "guardian" }
        .map {
            DropdownOption(
                value = it.user_id.toString(),
                label = it.username,
                hint = it.email.ifBlank { it.phone.ifBlank { it.role } },
            )
        }
    val protectedOptions = state.familyOverview.members
        .filter { it.role != "owner" }
        .map {
            DropdownOption(
                value = it.user_id.toString(),
                label = it.username,
                hint = it.relation.ifBlank { it.email.ifBlank { it.phone.ifBlank { it.role } } },
            )
        }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MobileSurface)
            .padding(top = padding.calculateTopPadding()),
    ) {
        MobileBackHeader(title = "家庭管理") { viewModel.openScreen(AppScreen.Family) }
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 12.dp, bottom = padding.calculateBottomPadding() + 24.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            item {
                Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
                    Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Text("邀请成员", fontWeight = FontWeight.ExtraBold, color = MobileText)
                        OutlinedTextField(
                            value = state.familyInviteForm.inviteeEmail,
                            onValueChange = viewModel::updateFamilyInviteEmail,
                            modifier = Modifier.fillMaxWidth(),
                            placeholder = { Text("邮箱（选填）") },
                            shape = RoundedCornerShape(18.dp),
                        )
                        OutlinedTextField(
                            value = state.familyInviteForm.inviteePhone,
                            onValueChange = viewModel::updateFamilyInvitePhone,
                            modifier = Modifier.fillMaxWidth(),
                            placeholder = { Text("手机号（选填）") },
                            shape = RoundedCornerShape(18.dp),
                        )
                        OutlinedTextField(
                            value = state.familyInviteForm.relation,
                            onValueChange = viewModel::updateFamilyInviteRelation,
                            modifier = Modifier.fillMaxWidth(),
                            placeholder = { Text("关系，例如父亲、配偶") },
                            shape = RoundedCornerShape(18.dp),
                        )
                        SelectionDropdownField(
                            label = "成员角色",
                            valueLabel = roleOptions.firstOrNull { it.value == state.familyInviteForm.role }?.label.orEmpty(),
                            placeholder = "选择角色",
                            options = roleOptions,
                            selectedValue = state.familyInviteForm.role,
                            hint = roleOptions.firstOrNull { it.value == state.familyInviteForm.role }?.hint.orEmpty().ifBlank { "设置成员在家庭中的职责" },
                            accentColor = Color(0xFF3B82F6),
                        ) { value ->
                            viewModel.updateFamilyInviteRole(value)
                        }
                        Button(onClick = viewModel::createFamilyInvitation, modifier = Modifier.fillMaxWidth().height(48.dp), shape = RoundedCornerShape(18.dp), colors = ButtonDefaults.buttonColors(containerColor = MobileText)) {
                            Text("发送邀请", fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
            item {
                Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
                    Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Text("配置守护", fontWeight = FontWeight.ExtraBold, color = MobileText)
                        SelectionDropdownField(
                            label = "守护人",
                            valueLabel = guardianOptions.firstOrNull { it.value == state.familyGuardianForm.guardianUserId }?.label.orEmpty(),
                            placeholder = "选择守护人",
                            options = guardianOptions,
                            selectedValue = state.familyGuardianForm.guardianUserId,
                            hint = guardianOptions.firstOrNull { it.value == state.familyGuardianForm.guardianUserId }?.hint.orEmpty().ifBlank { "守护人会收到高风险提醒" },
                            accentColor = MobileGreen,
                        ) { value ->
                            viewModel.updateGuardianUser(value)
                        }
                        SelectionDropdownField(
                            label = "被守护人",
                            valueLabel = protectedOptions.firstOrNull { it.value == state.familyGuardianForm.memberUserId }?.label.orEmpty(),
                            placeholder = "选择被守护人",
                            options = protectedOptions,
                            selectedValue = state.familyGuardianForm.memberUserId,
                            hint = protectedOptions.firstOrNull { it.value == state.familyGuardianForm.memberUserId }?.hint.orEmpty().ifBlank { "出现高风险时会触发通知" },
                            accentColor = MobileGreen,
                        ) { value ->
                            viewModel.updateProtectedUser(value)
                        }
                        Button(onClick = viewModel::createGuardianLink, modifier = Modifier.fillMaxWidth().height(48.dp), shape = RoundedCornerShape(18.dp)) {
                            Text("保存关系", fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
            item { SectionTitle("成员列表") }
            if (state.familyOverview.members.isEmpty()) {
                item { EmptyCard("暂无家庭成员") }
            } else {
                items(state.familyOverview.members, key = { it.member_id }) { member ->
                    FamilyManageMemberCard(
                        member = member,
                        deleting = state.deletingFamilyMemberIds.contains(member.member_id),
                        canDelete = state.familyOverview.current_member?.role == "owner" && member.role != "owner",
                    ) { viewModel.deleteFamilyMember(member.member_id) }
                }
            }
            item { SectionTitle("邀请记录") }
            if (state.familyOverview.invitations.isEmpty()) {
                item { EmptyCard("暂无邀请记录") }
            } else {
                items(state.familyOverview.invitations, key = { it.id }) { invitation ->
                    FamilyManageInvitationCard(invitation)
                }
            }
        }
    }
}

@Composable
fun MobileProfileScreen(
    state: MainUiState,
    padding: PaddingValues,
    onOpenPrivacy: () -> Unit,
    onQuickAnalyzeBubbleChange: (Boolean) -> Unit,
    onAccessibilityAutoAnalyzeChange: (Boolean) -> Unit,
    onLogout: () -> Unit,
) {
    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(MobileSurface),
        contentPadding = PaddingValues(
            start = 16.dp,
            end = 16.dp,
            top = padding.calculateTopPadding() + 12.dp,
            bottom = padding.calculateBottomPadding() + 32.dp,
        ),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        item {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(32.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
            ) {
                Box {
                    Box(
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .size(120.dp)
                            .offset(x = 20.dp, y = (-20).dp)
                            .background(Brush.linearGradient(listOf(Color(0xFFF0FDF4), Color(0xFFE2FBEF))), CircleShape),
                    )
                    Box(
                        modifier = Modifier
                            .align(Alignment.BottomStart)
                            .size(90.dp)
                            .offset(x = (-18).dp, y = 18.dp)
                            .background(Brush.linearGradient(listOf(Color(0xFFEFF6FF), Color(0xFFEEF2FF))), CircleShape),
                    )
                    Row(modifier = Modifier.padding(24.dp), verticalAlignment = Alignment.CenterVertically) {
                        Box(
                            modifier = Modifier
                                .size(80.dp)
                                .background(Brush.linearGradient(listOf(Color(0xFFF1F5F9), Color(0xFFF8FAFC))), RoundedCornerShape(24.dp))
                                .border(1.dp, Color.White, RoundedCornerShape(24.dp)),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(state.user.username.take(1).ifBlank { "U" }, fontSize = 30.sp, fontWeight = FontWeight.ExtraBold, color = MobileText)
                        }
                        Column(modifier = Modifier.padding(start = 16.dp).weight(1f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                            Text(state.user.username.ifBlank { "未设置用户名" }, style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.ExtraBold, color = MobileText)
                            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                                Box(modifier = Modifier.size(8.dp).background(MobileGreen, CircleShape))
                                Text("欢迎使用反诈护航 AI", fontSize = 13.sp, color = Color(0xFF94A3B8), fontWeight = FontWeight.Bold)
                            }
                        }
                        Box(
                            modifier = Modifier
                                .background(Brush.horizontalGradient(listOf(Color(0xFF10B981), Color(0xFF34D399))), RoundedCornerShape(10.dp))
                                .padding(horizontal = 10.dp, vertical = 4.dp),
                        ) {
                            Text(state.user.role.ifBlank { "User" }, color = Color.White, fontSize = 10.sp, fontWeight = FontWeight.Black)
                        }
                    }
                }
            }
        }
        item {
            NavigationCell("隐私资料", "查看与管理您的个人画像数据", Icons.Outlined.Security, Color(0xFF3B82F6)) {
                onOpenPrivacy()
            }
        }
        item {
            Text("设备守护", fontSize = 12.sp, fontWeight = FontWeight.Black, color = Color(0xFF94A3B8), modifier = Modifier.padding(horizontal = 4.dp))
        }
        item {
            ToggleCell(
                title = "悬浮球快捷分析",
                description = if (state.quickAnalyzeBubbleEnabled) "已开启，可快速截图并分析风险" else "开启后可在系统层快速触发风险分析",
                checked = state.quickAnalyzeBubbleEnabled,
                iconKind = GuardFeatureIconKind.QuickAnalyze,
                onCheckedChange = onQuickAnalyzeBubbleChange,
            )
        }
        item {
            ToggleCell(
                title = "无障碍自动守护",
                description = if (state.accessibilityAutoAnalyzeEnabled) "已开启，在敏感场景中自动触发分析" else "开启后可自动识别敏感场景并分析",
                checked = state.accessibilityAutoAnalyzeEnabled,
                iconKind = GuardFeatureIconKind.Accessibility,
                onCheckedChange = onAccessibilityAutoAnalyzeChange,
            )
        }
        item {
            Card(
                modifier = Modifier.fillMaxWidth().clickable(onClick = onLogout),
                shape = RoundedCornerShape(24.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
                border = BorderStroke(1.dp, Color(0x12CBD5E1)),
            ) {
                Box(modifier = Modifier.fillMaxWidth().padding(vertical = 16.dp), contentAlignment = Alignment.Center) {
                    Text("退出登录", color = Color(0xFFEF4444), fontWeight = FontWeight.Bold, fontSize = 15.sp)
                }
            }
        }
    }
}

@Composable
private fun ToggleCell(
    title: String,
    description: String,
    checked: Boolean,
    iconKind: GuardFeatureIconKind,
    onCheckedChange: (Boolean) -> Unit,
) {
    val badgeText = if (checked) "已开启" else "未开启"
    val badgeBg = if (checked) Color(0xFFF0FDF4) else Color(0xFFF8FAFC)
    val badgeTextColor = if (checked) Color(0xFF16A34A) else Color(0xFF64748B)
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(28.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = BorderStroke(1.dp, Color(0x12CBD5E1)),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 18.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            GuardFeatureIconTile(kind = iconKind)
            Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Text(title, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 15.sp)
                    Box(
                        modifier = Modifier
                            .background(badgeBg, RoundedCornerShape(999.dp))
                            .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(999.dp))
                            .padding(horizontal = 8.dp, vertical = 4.dp),
                    ) {
                        Text(badgeText, fontSize = 10.sp, fontWeight = FontWeight.Bold, color = badgeTextColor)
                    }
                }
                Text(description, fontSize = 12.sp, color = MobileSubtle, lineHeight = 18.sp, maxLines = 2, overflow = TextOverflow.Ellipsis)
            }
            androidx.compose.material3.Switch(checked = checked, onCheckedChange = onCheckedChange)
        }
    }
}

@Composable
fun MobileProfilePrivacyScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
    onRequestLocation: () -> Unit,
) {
    val occupationOptions = buildList {
        add(DropdownOption("", "未设置", "清空职业信息"))
        addAll(state.occupationOptions.map { DropdownOption(it, it, "") })
    }
    val provinceDropdownOptions = state.provinceOptions.map { option ->
        DropdownOption(option.code, option.name, "")
    }
    val cityDropdownOptions = state.cityOptions.map { option ->
        DropdownOption(option.code, option.name, "")
    }
    val districtDropdownOptions = state.districtOptions.map { option ->
        DropdownOption(option.code, option.name, "")
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MobileSurface)
            .padding(top = padding.calculateTopPadding()),
    ) {
        MobileBackHeader(title = "隐私资料") { viewModel.openScreen(AppScreen.Profile) }
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 12.dp, bottom = padding.calculateBottomPadding() + 24.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(32.dp),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    border = BorderStroke(1.dp, Color(0x12CBD5E1)),
                ) {
                    Column(modifier = Modifier.padding(24.dp), verticalArrangement = Arrangement.spacedBy(20.dp)) {
                        Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                            Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                Text("手机号", fontSize = 11.sp, fontWeight = FontWeight.Black, color = Color(0xFF94A3B8))
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color(0xFFF8FAFC), RoundedCornerShape(16.dp))
                                        .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(16.dp))
                                        .padding(horizontal = 16.dp, vertical = 14.dp),
                                ) {
                                    Text(state.user.phone.ifBlank { "未设置" }, color = Color(0xFF334155), fontWeight = FontWeight.ExtraBold, fontSize = 15.sp)
                                }
                            }
                            Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                Text("邮箱", fontSize = 11.sp, fontWeight = FontWeight.Black, color = Color(0xFF94A3B8))
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color(0xFFF8FAFC), RoundedCornerShape(16.dp))
                                        .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(16.dp))
                                        .padding(horizontal = 16.dp, vertical = 14.dp),
                                ) {
                                    Text(state.user.email.ifBlank { "未设置" }, color = Color(0xFF334155), fontWeight = FontWeight.ExtraBold, fontSize = 15.sp)
                                }
                            }
                        }

                        Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text("画像资料", fontSize = 13.sp, fontWeight = FontWeight.Black, color = MobileText, modifier = Modifier.weight(1f))
                                Box(
                                    modifier = Modifier
                                        .background(Color(0xFFF0FDF4), RoundedCornerShape(999.dp))
                                        .clickable(onClick = viewModel::toggleAgeEditor)
                                        .padding(horizontal = 14.dp, vertical = 7.dp),
                                ) {
                                    Text(
                                        if (state.ageEditorVisible) "收起编辑" else "编辑资料",
                                        fontSize = 12.sp,
                                        fontWeight = FontWeight.Bold,
                                        color = Color(0xFF059669),
                                    )
                                }
                            }

                            AnimatedVisibility(visible = !state.ageEditorVisible) {
                                Column(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color(0xFFF8FAFC), RoundedCornerShape(24.dp))
                                        .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(24.dp))
                                        .padding(18.dp),
                                    verticalArrangement = Arrangement.spacedBy(14.dp),
                                ) {
                                    Row(verticalAlignment = Alignment.Top, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                                        Text("年龄", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = Color(0xFF64748B), modifier = Modifier.width(44.dp))
                                        Text(
                                            (state.user.age ?: state.profileForm.age).takeIf { it > 0 }?.toString() ?: "未设置",
                                            fontSize = 14.sp,
                                            fontWeight = FontWeight.ExtraBold,
                                            color = Color(0xFF334155),
                                            modifier = Modifier.weight(1f),
                                            textAlign = TextAlign.End,
                                        )
                                    }
                                    Row(verticalAlignment = Alignment.Top, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                                        Text("职业", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = Color(0xFF64748B), modifier = Modifier.width(44.dp))
                                        Text(
                                            state.user.occupation.ifBlank { "未设置" },
                                            fontSize = 14.sp,
                                            fontWeight = FontWeight.ExtraBold,
                                            color = Color(0xFF334155),
                                            modifier = Modifier.weight(1f),
                                            textAlign = TextAlign.End,
                                        )
                                    }
                                    Row(verticalAlignment = Alignment.Top, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                                        Text("位置", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = Color(0xFF64748B), modifier = Modifier.width(44.dp))
                                        Text(
                                            listOf(state.user.province_name, state.user.city_name, state.user.district_name)
                                                .filter { it.isNotBlank() }
                                                .joinToString(" / ")
                                                .ifBlank { "未设置" },
                                            fontSize = 14.sp,
                                            fontWeight = FontWeight.ExtraBold,
                                            color = Color(0xFF334155),
                                            modifier = Modifier.weight(1f),
                                            textAlign = TextAlign.End,
                                        )
                                    }
                                }
                            }

                            AnimatedVisibility(visible = state.ageEditorVisible) {
                                Column(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color(0xFFF8FAFC), RoundedCornerShape(24.dp))
                                        .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(24.dp))
                                        .padding(18.dp),
                                    verticalArrangement = Arrangement.spacedBy(12.dp),
                                ) {
                                OutlinedTextField(
                                    value = state.profileForm.age.toString(),
                                    onValueChange = { input -> input.toIntOrNull()?.let(viewModel::updateProfileAge) },
                                    modifier = Modifier.fillMaxWidth(),
                                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                                    label = { Text("年龄") },
                                    shape = RoundedCornerShape(14.dp),
                                )
                                SelectionDropdownField(
                                    label = "职业",
                                    valueLabel = state.profileForm.occupation,
                                    placeholder = "选择职业",
                                    options = occupationOptions,
                                    selectedValue = state.profileForm.occupation,
                                    hint = "从配置列表中选择当前职业",
                                    accentColor = MobileGreen,
                                ) { value ->
                                    viewModel.updateProfileOccupation(value)
                                }
                                SelectionDropdownField(
                                    label = "省份",
                                    valueLabel = state.profileForm.provinceName,
                                    placeholder = "选择省份",
                                    options = provinceDropdownOptions,
                                    selectedValue = state.profileForm.provinceCode,
                                    hint = "选择您当前常驻的省份",
                                    accentColor = MobileGreen,
                                ) { value ->
                                    viewModel.selectProvinceValue(value)
                                }
                                SelectionDropdownField(
                                    label = "城市",
                                    valueLabel = state.profileForm.cityName,
                                    placeholder = "选择城市",
                                    options = cityDropdownOptions,
                                    selectedValue = state.profileForm.cityCode,
                                    hint = if (state.profileForm.provinceCode.isBlank()) "请先选择省份" else "选择所在城市",
                                    accentColor = MobileGreen,
                                    enabled = state.profileForm.provinceCode.isNotBlank(),
                                ) { value ->
                                    viewModel.selectCityValue(value)
                                }
                                SelectionDropdownField(
                                    label = "区县",
                                    valueLabel = state.profileForm.districtName,
                                    placeholder = "选择区县",
                                    options = districtDropdownOptions,
                                    selectedValue = state.profileForm.districtCode,
                                    hint = if (state.profileForm.cityCode.isBlank()) "请先选择城市" else "选择所在区县",
                                    accentColor = MobileGreen,
                                    enabled = state.profileForm.cityCode.isNotBlank(),
                                ) { value ->
                                    viewModel.selectDistrictValue(value)
                                }
                                Button(
                                    onClick = onRequestLocation,
                                    modifier = Modifier.fillMaxWidth().height(48.dp),
                                    shape = RoundedCornerShape(14.dp),
                                    enabled = !state.locationResolving,
                                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFF0FDF4), contentColor = Color(0xFF059669)),
                                ) {
                                    Text(if (state.locationResolving) "定位中..." else "自动定位当前位置", fontWeight = FontWeight.Bold)
                                }
                                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                                    Button(
                                        onClick = viewModel::updateUserProfile,
                                        modifier = Modifier.weight(1f).height(52.dp),
                                        enabled = !state.profileSaving,
                                        shape = RoundedCornerShape(14.dp),
                                        colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF10B981), contentColor = Color.White),
                                    ) {
                                        Text(if (state.profileSaving) "保存中..." else "保存修改", fontWeight = FontWeight.Bold)
                                    }
                                    Button(
                                        onClick = viewModel::cancelAgeEditor,
                                        modifier = Modifier.weight(1f).height(52.dp),
                                        shape = RoundedCornerShape(14.dp),
                                        colors = ButtonDefaults.buttonColors(containerColor = Color.White, contentColor = Color(0xFF64748B)),
                                        border = BorderStroke(1.dp, Color(0x12CBD5E1)),
                                    ) {
                                        Text("取消", fontWeight = FontWeight.Bold)
                                    }
                                }
                                }
                            }
                        }

                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(1.dp)
                                .background(Color(0xFFF1F5F9)),
                        )

                        Button(
                            onClick = viewModel::deleteAccount,
                            modifier = Modifier.fillMaxWidth().height(52.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFFFF1F2), contentColor = Color(0xFFDC2626)),
                            shape = RoundedCornerShape(14.dp),
                        ) {
                            Text("永久注销账户", fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ChoiceButton(title: String, selected: Boolean, modifier: Modifier = Modifier, onClick: () -> Unit) {
    Button(
        onClick = onClick,
        modifier = modifier.height(44.dp),
        shape = RoundedCornerShape(14.dp),
        colors = ButtonDefaults.buttonColors(
            containerColor = if (selected) MobileText else Color(0xFFF8FAFC),
            contentColor = if (selected) Color.White else MobileText,
        ),
    ) {
        Text(title, fontWeight = FontWeight.Bold)
    }
}

@Composable
private fun SimpleSelectionRow(title: String, selected: Boolean, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .background(if (selected) Color(0xFFF0FDF4) else Color.White, RoundedCornerShape(14.dp))
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(title, modifier = Modifier.weight(1f), color = MobileText, fontWeight = if (selected) FontWeight.ExtraBold else FontWeight.Medium)
        if (selected) {
            Text("已选", color = MobileGreen, fontSize = 12.sp, fontWeight = FontWeight.Bold)
        }
    }
}
