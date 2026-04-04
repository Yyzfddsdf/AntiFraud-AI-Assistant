package com.example.myapplication

import android.content.ContentValues
import android.os.Environment
import android.provider.MediaStore
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material.icons.outlined.DeleteOutline
import androidx.compose.material.icons.outlined.Description
import androidx.compose.material.icons.outlined.Image
import androidx.compose.material.icons.outlined.Mic
import androidx.compose.material.icons.outlined.Videocam
import androidx.compose.material.icons.outlined.WarningAmber
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun MobileTaskDetailSheet(
    task: TaskDetail,
    onDismiss: () -> Unit,
    onMessage: (String, Boolean) -> Unit,
) {
    val context = LocalContext.current
    var menuExpanded by remember { mutableStateOf(false) }
    val parsedRiskSummary = remember(task.risk_summary) { parseRiskSummary(task.risk_summary) }
    val displayRiskLevel = task.risk_level.ifBlank {
        parsedRiskSummary?.riskLevel
            ?.takeIf { it.isNotBlank() }
            ?: extractRiskLevelFromReport(task.report).takeIf { it.isNotBlank() }
            ?: extractTaskRiskLevel(task.risk_summary).takeIf { it.isNotBlank() }
            ?: extractTaskRiskScore(task.risk_summary)?.let(::inferRiskLevelFromScore)
            ?: inferRiskLevelFromScore(task.risk_score)
    }
    val riskTheme = overlayRiskTheme(displayRiskLevel)
    val attackSteps = remember(task.report) { extractAttackSteps(task.report) }
    val keywordSentences = remember(task.report) { extractScamKeywordSentences(task.report) }
    val reportSections = remember(task.report) { parseReport(task.report).ifEmpty { listOf(ReportSection(0, "报告内容", task.report)) } }
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.4f))
            .clickable(onClick = onDismiss),
        contentAlignment = Alignment.BottomCenter,
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .fillMaxSize(0.90f)
                .clickable(enabled = false) {},
            shape = RoundedCornerShape(topStart = 32.dp, topEnd = 32.dp),
            colors = CardDefaults.cardColors(containerColor = MobileSurface),
        ) {
            Column {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White.copy(alpha = 0.92f))
                        .padding(top = 12.dp, bottom = 12.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Box(
                        modifier = Modifier
                            .width(48.dp)
                            .height(6.dp)
                            .background(Color(0xFFE2E8F0), RoundedCornerShape(999.dp)),
                    )
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 10.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Column(modifier = Modifier.weight(1f)) {
                            Text("案件分析详情", fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 17.sp)
                            Text("ID: ${task.task_id}", color = Color(0xFF94A3B8), fontSize = 11.sp)
                        }
                        IconButton(
                            onClick = onDismiss,
                            modifier = Modifier
                                .size(32.dp)
                                .background(Color(0xFFF1F5F9), CircleShape),
                        ) {
                            Icon(Icons.Outlined.Close, contentDescription = "关闭", tint = Color(0xFF64748B), modifier = Modifier.size(16.dp))
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
                    contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 16.dp, bottom = 120.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                ) {
                    item {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Row(
                                modifier = Modifier.weight(1f),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                            ) {
                                OverlayStatusBadge(task.status)
                                Box(modifier = Modifier.size(4.dp).background(Color(0xFFCBD5E1), CircleShape))
                                Text(formatDateTime(task.created_at), color = Color(0xFF64748B), fontSize = 11.sp, fontWeight = FontWeight.Medium)
                            }
                            Box {
                                IconButton(
                                    onClick = { menuExpanded = true },
                                    modifier = Modifier
                                        .size(32.dp)
                                        .background(if (menuExpanded) Color(0xFFECFDF5) else Color.White, CircleShape)
                                        .border(1.dp, if (menuExpanded) Color(0xFFA7F3D0) else MobileBorder, CircleShape),
                                ) {
                                    Icon(
                                        Icons.Outlined.Description,
                                        contentDescription = "导出",
                                        tint = if (menuExpanded) MobileGreen else Color(0xFF475569),
                                        modifier = Modifier.size(16.dp),
                                    )
                                }
                                DropdownMenu(expanded = menuExpanded, onDismissRequest = { menuExpanded = false }) {
                                    DropdownMenuItem(
                                        text = {
                                            Column {
                                                Text("Markdown", fontWeight = FontWeight.Bold)
                                                Text("导出结构化文本报告", fontSize = 10.sp, color = Color(0xFF64748B))
                                            }
                                        },
                                        onClick = {
                                            saveTextToDownloads(context, buildExportFilename(task, "md"), "text/markdown", buildTaskMarkdown(task), onMessage)
                                            menuExpanded = false
                                        },
                                    )
                                    DropdownMenuItem(
                                        text = {
                                            Column {
                                                Text("JSON", fontWeight = FontWeight.Bold)
                                                Text("导出完整结构化数据", fontSize = 10.sp, color = Color(0xFF64748B))
                                            }
                                        },
                                        onClick = {
                                            saveTextToDownloads(context, buildExportFilename(task, "json"), "application/json", Json { prettyPrint = true }.encodeToString(TaskDetail.serializer(), task), onMessage)
                                            menuExpanded = false
                                        },
                                    )
                                }
                            }
                        }
                    }
                    if (task.risk_score > 0 || task.risk_summary.isNotBlank()) {
                        item {
                            Card(
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(16.dp),
                                colors = CardDefaults.cardColors(containerColor = Color.White),
                            ) {
                                Box(modifier = Modifier.fillMaxWidth()) {
                                    Box(
                                        modifier = Modifier
                                            .align(Alignment.TopEnd)
                                            .size(128.dp)
                                            .background(riskTheme.first.copy(alpha = 0.10f), RoundedCornerShape(bottomStart = 128.dp)),
                                    )
                                    Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
                                        Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                                            Text("风险评估", fontSize = 11.sp, color = Color(0xFF94A3B8), fontWeight = FontWeight.Black)
                                            Row(verticalAlignment = Alignment.Bottom, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                                Text("${task.risk_score}", fontWeight = FontWeight.Black, fontSize = 36.sp, color = MobileText)
                                                if (displayRiskLevel.isNotBlank()) {
                                                    OverlayRiskLevelBadge(displayRiskLevel)
                                                }
                                            }
                                        }
                                        if (parsedRiskSummary != null) {
                                            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                                                    OverlayMetricCell("社工话术", parsedRiskSummary.dimensions["social_engineering"] ?: 0, Modifier.weight(1f))
                                                    OverlayMetricCell("诱导动作", parsedRiskSummary.dimensions["requested_actions"] ?: 0, Modifier.weight(1f))
                                                }
                                                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                                                    OverlayMetricCell("证据强度", parsedRiskSummary.dimensions["evidence_strength"] ?: 0, Modifier.weight(1f))
                                                    OverlayMetricCell("受害暴露", parsedRiskSummary.dimensions["loss_exposure"] ?: 0, Modifier.weight(1f))
                                                }
                                                if (parsedRiskSummary.hitRules.isNotEmpty()) {
                                                    FlowRow(
                                                        modifier = Modifier.fillMaxWidth(),
                                                        horizontalArrangement = Arrangement.spacedBy(6.dp),
                                                        verticalArrangement = Arrangement.spacedBy(6.dp),
                                                    ) {
                                                        parsedRiskSummary.hitRules.forEach { rule ->
                                                            Box(
                                                                modifier = Modifier
                                                                    .background(Color(0xFFFFF1F2), RoundedCornerShape(10.dp))
                                                                    .border(1.dp, Color(0xFFFFE4E6), RoundedCornerShape(10.dp))
                                                                    .padding(horizontal = 10.dp, vertical = 6.dp),
                                                            ) {
                                                                Text(rule, fontSize = 10.sp, fontWeight = FontWeight.Bold, color = Color(0xFFE11D48))
                                                            }
                                                        }
                                                    }
                                                }
                                            }
                                        } else if (task.risk_summary.isNotBlank()) {
                                            Text(task.risk_summary, fontSize = 12.sp, color = Color(0xFF475569), lineHeight = 19.sp)
                                        }
                                    }
                                }
                            }
                        }
                    }
                    if (task.summary.isNotBlank()) {
                        item { OverlaySectionCard(eyebrow = "案件摘要", title = "", body = task.summary) }
                    }
                    if (task.report.isNotBlank()) {
                        if (attackSteps.isNotEmpty()) {
                            item {
                                Card(
                                    modifier = Modifier.fillMaxWidth(),
                                    shape = RoundedCornerShape(16.dp),
                                    colors = CardDefaults.cardColors(containerColor = Color(0xFFFFFBFB)),
                                ) {
                                    Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                                        Text("诈骗链路时间线", fontWeight = FontWeight.Black, color = Color(0xFFFB7185), fontSize = 11.sp)
                                        attackSteps.forEachIndexed { index, step ->
                                            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.Top) {
                                                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                                    Box(
                                                        modifier = Modifier
                                                            .size(18.dp)
                                                            .background(Color.White, CircleShape)
                                                            .border(2.dp, Color(0xFFFB7185), CircleShape),
                                                        contentAlignment = Alignment.Center,
                                                    ) {
                                                        Box(modifier = Modifier.size(6.dp).background(Color(0xFFF43F5E), CircleShape))
                                                    }
                                                    if (index < attackSteps.lastIndex) {
                                                        Box(modifier = Modifier.width(2.dp).height(34.dp).background(Color(0xFFFECDD3)))
                                                    }
                                                }
                                                Text(
                                                    step,
                                                    modifier = Modifier
                                                        .weight(1f)
                                                        .background(Color.White.copy(alpha = 0.65f), RoundedCornerShape(14.dp))
                                                        .border(1.dp, Color(0xFFFFE4E6), RoundedCornerShape(14.dp))
                                                        .padding(horizontal = 12.dp, vertical = 10.dp),
                                                    color = MobileText,
                                                    fontSize = 13.sp,
                                                    lineHeight = 19.sp,
                                                    fontWeight = FontWeight.Medium,
                                                )
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        if (keywordSentences.isNotEmpty()) {
                            item {
                                Card(
                                    modifier = Modifier.fillMaxWidth(),
                                    shape = RoundedCornerShape(16.dp),
                                    colors = CardDefaults.cardColors(containerColor = Color(0xFFFDF4FF)),
                                ) {
                                    Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                                        Text("高危关键词句", fontWeight = FontWeight.Black, color = Color(0xFFC026D3), fontSize = 11.sp)
                                        keywordSentences.forEach { keyword ->
                                            Box(
                                                modifier = Modifier
                                                    .background(Color.White, RoundedCornerShape(12.dp))
                                                    .border(1.dp, Color(0xFFF5D0FE), RoundedCornerShape(12.dp))
                                                    .padding(horizontal = 12.dp, vertical = 10.dp),
                                            ) {
                                                Text(keyword, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFFA21CAF))
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        item {
                            Card(
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(16.dp),
                                colors = CardDefaults.cardColors(containerColor = Color.White),
                            ) {
                                Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                                    Text("综合分析报告", fontWeight = FontWeight.Black, color = Color(0xFF94A3B8), fontSize = 11.sp)
                                    reportSections.forEach { section ->
                                        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                            Text(section.title, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 13.sp)
                                            Text(section.content, color = MobileSubtle, fontSize = 13.sp, lineHeight = 20.sp)
                                        }
                                    }
                                }
                            }
                        }
                    }
                    item {
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(16.dp),
                            colors = CardDefaults.cardColors(containerColor = Color.White),
                        ) {
                            Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                                Text("输入概览", fontWeight = FontWeight.Black, color = Color(0xFF94A3B8), fontSize = 11.sp)
                                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                                    OverlayInputMetricCell("文本", if (task.payload.text.isNotBlank()) "已提交" else "无", Icons.Outlined.Description, Modifier.weight(1f))
                                    OverlayInputMetricCell("图片", "${task.payload.images.size} 份", Icons.Outlined.Image, Modifier.weight(1f))
                                }
                                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                                    OverlayInputMetricCell("音频", "${task.payload.audios.size} 份", Icons.Outlined.Mic, Modifier.weight(1f))
                                    OverlayInputMetricCell("视频", "${task.payload.videos.size} 份", Icons.Outlined.Videocam, Modifier.weight(1f))
                                }
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color(0xFFF8FAFC), RoundedCornerShape(14.dp))
                                        .border(1.dp, MobileBorder, RoundedCornerShape(14.dp))
                                        .padding(horizontal = 14.dp, vertical = 12.dp),
                                ) {
                                    Text(
                                        "原始多模态材料受限于移动端屏幕尺寸，不在此处展开展示。",
                                        fontSize = 11.sp,
                                        lineHeight = 17.sp,
                                        fontWeight = FontWeight.Medium,
                                        color = Color(0xFF94A3B8),
                                    )
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
fun MobileAlertSheet(
    event: AlertEvent,
    onDismiss: () -> Unit,
    onOpenCase: () -> Unit,
) {
    val theme = alertSeverityPalette(event.risk_level)
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0x990F172A))
            .clickable(onClick = onDismiss),
        contentAlignment = Alignment.BottomCenter,
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .clickable(enabled = false) {},
            shape = RoundedCornerShape(topStart = 32.dp, topEnd = 32.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White),
        ) {
            Column {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 12.dp),
                    contentAlignment = Alignment.Center,
                ) {
                    Box(
                        modifier = Modifier
                            .width(48.dp)
                            .height(6.dp)
                            .background(Color(0xFFE5E7EB), RoundedCornerShape(999.dp)),
                    )
                }
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White.copy(alpha = 0.95f))
                        .padding(horizontal = 20.dp, vertical = 16.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(10.dp), modifier = Modifier.weight(1f)) {
                        Box(
                            modifier = Modifier
                                .size(32.dp)
                                .background(theme.badgeBackground, CircleShape),
                            contentAlignment = Alignment.Center,
                        ) {
                            Icon(Icons.Outlined.WarningAmber, contentDescription = null, tint = theme.badgeColor, modifier = Modifier.size(16.dp))
                        }
                        Text("风险预警", fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 20.sp)
                    }
                    IconButton(
                        onClick = onDismiss,
                        modifier = Modifier
                            .size(36.dp)
                            .background(Color(0xFFF8FAFC), CircleShape),
                    ) {
                        Icon(Icons.Outlined.Close, contentDescription = "关闭", tint = Color(0xFF64748B))
                    }
                }
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 20.dp, vertical = 20.dp),
                    verticalArrangement = Arrangement.spacedBy(18.dp),
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                        Box(
                            modifier = Modifier
                                .background(theme.badgeBackground, RoundedCornerShape(10.dp))
                                .border(1.dp, theme.badgeColor.copy(alpha = 0.20f), RoundedCornerShape(10.dp))
                                .padding(horizontal = 10.dp, vertical = 6.dp),
                        ) {
                            Text("${normalizeRiskLevel(event.risk_level)} 风险", color = theme.badgeColor, fontSize = 11.sp, fontWeight = FontWeight.Bold)
                        }
                        Text(formatDateTime(event.created_at.ifBlank { event.sent_at }), color = Color(0xFF94A3B8), fontSize = 12.sp)
                    }
                    Text(event.title.ifBlank { "风险预警" }, fontWeight = FontWeight.ExtraBold, fontSize = 26.sp, color = MobileText)
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Brush.linearGradient(listOf(theme.panelStart, theme.panelEnd)), RoundedCornerShape(16.dp))
                            .border(1.dp, theme.panelBorder, RoundedCornerShape(16.dp))
                            .padding(16.dp),
                    ) {
                        Text(event.case_summary.ifBlank { "请及时核查相关风险事件。" }, color = Color(0xFF334155), fontSize = 15.sp, lineHeight = 22.sp)
                    }
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp), modifier = Modifier.fillMaxWidth()) {
                        AlertInfoMiniCard("诈骗类型", event.scam_type.ifBlank { "待分析" }, Modifier.weight(1f))
                        AlertInfoMiniCard("案件ID", event.record_id.ifBlank { "-" }, Modifier.weight(1f), mono = true)
                    }
                }
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .border(1.dp, Color(0xFFF1F5F9), RoundedCornerShape(1.dp))
                        .padding(horizontal = 20.dp, vertical = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    OutlinedButton(
                        onClick = onDismiss,
                        modifier = Modifier.weight(0.4f),
                        shape = RoundedCornerShape(18.dp),
                        border = null,
                        colors = ButtonDefaults.outlinedButtonColors(containerColor = Color(0xFFF1F5F9), contentColor = Color(0xFF475569)),
                    ) {
                        Text("稍后处理", fontWeight = FontWeight.Bold)
                    }
                    Button(
                        onClick = onOpenCase,
                        modifier = Modifier.weight(0.6f),
                        shape = RoundedCornerShape(18.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = theme.actionColor),
                    ) {
                        Text("查看案件", fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}

@Composable
fun MobileFamilyAlertSheet(
    notification: FamilyNotification,
    onDismiss: () -> Unit,
    onOpenCenter: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0x990F172A))
            .clickable(onClick = onDismiss),
        contentAlignment = Alignment.BottomCenter,
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .clickable(enabled = false) {},
            shape = RoundedCornerShape(topStart = 32.dp, topEnd = 32.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White),
        ) {
            Column {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 12.dp),
                    contentAlignment = Alignment.Center,
                ) {
                    Box(
                        modifier = Modifier
                            .width(48.dp)
                            .height(6.dp)
                            .background(Color(0xFFE5E7EB), RoundedCornerShape(999.dp)),
                    )
                }
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White.copy(alpha = 0.95f))
                        .padding(horizontal = 20.dp, vertical = 16.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(10.dp), modifier = Modifier.weight(1f)) {
                        Box(
                            modifier = Modifier
                                .size(32.dp)
                                .background(Color(0xFFFFE4E6), CircleShape),
                            contentAlignment = Alignment.Center,
                        ) {
                            Icon(Icons.Outlined.WarningAmber, contentDescription = null, tint = Color(0xFFE11D48), modifier = Modifier.size(16.dp))
                        }
                        Text("家庭联防通知", fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 20.sp)
                    }
                    IconButton(
                        onClick = onDismiss,
                        modifier = Modifier
                            .size(36.dp)
                            .background(Color(0xFFF8FAFC), CircleShape),
                    ) {
                        Icon(Icons.Outlined.Close, contentDescription = "关闭", tint = Color(0xFF64748B))
                    }
                }
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 20.dp, vertical = 20.dp),
                    verticalArrangement = Arrangement.spacedBy(18.dp),
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                        Box(
                            modifier = Modifier
                                .background(Color(0xFFFFE4E6), RoundedCornerShape(10.dp))
                                .border(1.dp, Color(0xFFFDA4AF), RoundedCornerShape(10.dp))
                                .padding(horizontal = 10.dp, vertical = 6.dp),
                        ) {
                            Text("${notification.risk_level.ifBlank { "高" }} 风险", color = Color(0xFFE11D48), fontSize = 11.sp, fontWeight = FontWeight.Bold)
                        }
                        Text(formatDateTime(notification.event_at), color = Color(0xFF94A3B8), fontSize = 12.sp)
                    }
                    Text(notification.title.ifBlank { "高风险案件预警" }, fontWeight = FontWeight.ExtraBold, fontSize = 26.sp, color = MobileText)
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(
                                Brush.linearGradient(listOf(Color(0xFFFFF1F2), Color(0xFFFFFBFB))),
                                RoundedCornerShape(16.dp),
                            )
                            .border(1.dp, Color(0xFFFECACA), RoundedCornerShape(16.dp))
                            .padding(16.dp),
                    ) {
                        Text(
                            notification.case_summary.ifBlank { notification.summary.ifBlank { "家庭成员触发高风险事件，请及时核查。" } },
                            color = Color(0xFF334155),
                            fontSize = 15.sp,
                            lineHeight = 22.sp,
                        )
                    }
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp), modifier = Modifier.fillMaxWidth()) {
                        AlertInfoMiniCard("家庭成员", notification.target_name.ifBlank { "待确认" }, Modifier.weight(1f))
                        AlertInfoMiniCard("诈骗类型", notification.scam_type.ifBlank { "待分析" }, Modifier.weight(1f))
                    }
                }
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .border(1.dp, Color(0xFFF1F5F9), RoundedCornerShape(1.dp))
                        .padding(horizontal = 20.dp, vertical = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    OutlinedButton(
                        onClick = onDismiss,
                        modifier = Modifier.weight(0.4f),
                        shape = RoundedCornerShape(18.dp),
                        border = null,
                        colors = ButtonDefaults.outlinedButtonColors(containerColor = Color(0xFFF1F5F9), contentColor = Color(0xFF475569)),
                    ) {
                        Text("稍后处理", fontWeight = FontWeight.Bold)
                    }
                    Button(
                        onClick = onOpenCenter,
                        modifier = Modifier.weight(0.6f),
                        shape = RoundedCornerShape(18.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = MobileRose),
                    ) {
                        Text("进入家庭中心", fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}

private data class AlertSeverityPalette(
    val badgeBackground: Color,
    val badgeColor: Color,
    val panelStart: Color,
    val panelEnd: Color,
    val panelBorder: Color,
    val actionColor: Color,
)

@Composable
private fun AlertInfoMiniCard(
    label: String,
    value: String,
    modifier: Modifier = Modifier,
    mono: Boolean = false,
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(18.dp),
        colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC)),
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(label, fontSize = 11.sp, fontWeight = FontWeight.Black, color = Color(0xFF94A3B8))
            Text(value, fontSize = if (mono) 13.sp else 15.sp, fontWeight = FontWeight.Bold, color = MobileText)
        }
    }
}

private fun alertSeverityPalette(level: String): AlertSeverityPalette = when (normalizeRiskLevel(level)) {
    "高" -> AlertSeverityPalette(
        badgeBackground = Color(0xFFFEE2E2),
        badgeColor = Color(0xFFDC2626),
        panelStart = Color(0xFFFEF2F2),
        panelEnd = Color(0xFFFFFBFB),
        panelBorder = Color(0xFFFECACA),
        actionColor = Color(0xFFDC2626),
    )
    "低" -> AlertSeverityPalette(
        badgeBackground = Color(0xFFD1FAE5),
        badgeColor = Color(0xFF059669),
        panelStart = Color(0xFFECFDF5),
        panelEnd = Color(0xFFF0FDF4),
        panelBorder = Color(0xFFA7F3D0),
        actionColor = Color(0xFF059669),
    )
    else -> AlertSeverityPalette(
        badgeBackground = Color(0xFFFEF3C7),
        badgeColor = Color(0xFFD97706),
        panelStart = Color(0xFFFFFBEB),
        panelEnd = Color(0xFFFFF7ED),
        panelBorder = Color(0xFFFDE68A),
        actionColor = Color(0xFFD97706),
    )
}

@Composable
private fun OverlaySectionCard(eyebrow: String, title: String, body: String) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
    ) {
        Column(modifier = Modifier.padding(20.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Text(eyebrow, fontSize = 11.sp, fontWeight = FontWeight.Black, color = Color(0xFF94A3B8))
            if (title.isNotBlank()) {
                Text(title, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 13.sp)
            }
            Text(body, color = MobileSubtle, fontSize = 13.sp, lineHeight = 20.sp, fontWeight = FontWeight.Medium)
        }
    }
}

private data class OverlayParsedRiskSummary(
    val riskLevel: String,
    val dimensions: Map<String, Int>,
    val hitRules: List<String>,
)

private fun parseRiskSummary(raw: String): OverlayParsedRiskSummary? {
    if (raw.isBlank()) return null
    return runCatching {
        val normalized = raw.trim()
            .removeSurrounding("\"")
            .replace("\\\"", "\"")
            .replace("\\n", "\n")
            .replace("\\r", "\r")
            .replace("\\t", "\t")
        val root = Json.parseToJsonElement(normalized).jsonObject
        val riskLevel = extractTaskRiskLevel(normalized)
        val dimensions = root["dimensions"]?.jsonObject?.mapValues { (_, value) -> value.toString().trim('"').toIntOrNull() ?: 0 } ?: emptyMap()
        val hitRules = root["hit_rules"]?.jsonArray?.mapNotNull { value ->
            value.toString().trim('"').takeIf { it.isNotBlank() }
        }.orEmpty()
        OverlayParsedRiskSummary(riskLevel, dimensions, hitRules)
    }.getOrNull()
}

private fun parseReport(text: String): List<ReportSection> {
    val sections = mutableListOf<ReportSection>()
    var currentTitle = ""
    val buffer = StringBuilder()
    var currentId = 0
    text.lines().forEach { line ->
        val match = Regex("""^(\d+)\.\s+(.+)$""").matchEntire(line.trim())
        if (match != null) {
            if (currentTitle.isNotBlank()) sections += ReportSection(currentId, currentTitle, buffer.toString().trim())
            currentId = match.groupValues[1].toIntOrNull() ?: currentId + 1
            currentTitle = match.groupValues[2]
            buffer.clear()
        } else {
            buffer.appendLine(line)
        }
    }
    if (currentTitle.isNotBlank()) sections += ReportSection(currentId, currentTitle, buffer.toString().trim())
    return sections
}

private fun extractAttackSteps(text: String): List<String> = parseReport(text)
    .firstOrNull { it.title.contains("链路") }
    ?.content
    ?.lines()
    ?.map { it.trim().replace(Regex("""^[-*•]\s+|^\d+[.)、]\s*"""), "") }
    ?.filter { it.isNotBlank() }
    .orEmpty()

private fun extractScamKeywordSentences(text: String): List<String> = parseReport(text)
    .firstOrNull { it.title.contains("关键词句") }
    ?.content
    ?.lines()
    ?.map { it.trim().replace(Regex("""^[-*•]\s+|^\d+[.)、]\s*"""), "") }
    ?.filter { it.isNotBlank() }
    .orEmpty()

@Composable
private fun OverlayMetricCell(label: String, value: Int, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier
            .background(Color(0xFFF8FAFC), RoundedCornerShape(16.dp))
            .border(1.dp, MobileBorder, RoundedCornerShape(16.dp))
            .padding(horizontal = 12.dp, vertical = 12.dp),
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Text(label, fontSize = 10.sp, fontWeight = FontWeight.Bold, color = Color(0xFF94A3B8))
        Text(value.toString(), fontSize = 15.sp, fontWeight = FontWeight.Black, color = MobileText)
    }
}

@Composable
private fun OverlayInputMetricCell(
    label: String,
    value: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier
            .background(Color(0xFFF8FAFC), RoundedCornerShape(16.dp))
            .border(1.dp, MobileBorder, RoundedCornerShape(16.dp))
            .padding(horizontal = 12.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Box(
            modifier = Modifier
                .size(28.dp)
                .background(Color.White, RoundedCornerShape(10.dp)),
            contentAlignment = Alignment.Center,
        ) {
            Icon(icon, contentDescription = null, tint = Color(0xFF64748B), modifier = Modifier.size(14.dp))
        }
        Column {
            Text(label, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFF64748B))
            Text(value, fontSize = 13.sp, fontWeight = FontWeight.Black, color = MobileText)
        }
    }
}

@Composable
private fun OverlayStatusBadge(status: String) {
    val (text, bg, fg) = when (status) {
        "pending" -> Triple("等待中", Color(0xFFFEF3C7), Color(0xFFD97706))
        "processing" -> Triple("分析中", Color(0xFFDBEAFE), Color(0xFF2563EB))
        "completed" -> Triple("已完成", Color(0xFFD1FAE5), Color(0xFF059669))
        "failed" -> Triple("失败", Color(0xFFFEE2E2), Color(0xFFDC2626))
        else -> Triple(status, Color(0xFFF1F5F9), Color(0xFF475569))
    }
    Box(modifier = Modifier.background(bg, RoundedCornerShape(10.dp)).padding(horizontal = 8.dp, vertical = 4.dp)) {
        Text(text, fontSize = 10.sp, fontWeight = FontWeight.Black, color = fg)
    }
}

@Composable
private fun OverlayRiskLevelBadge(level: String) {
    val (bg, fg, text) = when (normalizeRiskLevel(level)) {
        "高" -> Triple(Color(0xFFFEE2E2), Color(0xFFDC2626), "高")
        "低" -> Triple(Color(0xFFD1FAE5), Color(0xFF059669), "低")
        else -> Triple(Color(0xFFFEF3C7), Color(0xFFD97706), "中")
    }
    Box(
        modifier = Modifier
            .background(bg, RoundedCornerShape(8.dp))
            .border(1.dp, fg.copy(alpha = 0.20f), RoundedCornerShape(8.dp))
            .padding(horizontal = 8.dp, vertical = 4.dp),
    ) {
        Text(text, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = fg)
    }
}

private fun overlayRiskTheme(level: String): Triple<Color, Color, Color> {
    if (level.isBlank()) return Triple(Color(0xFF94A3B8), Color(0xFFF1F5F9), Color(0xFF64748B))
    return when (normalizeRiskLevel(level)) {
        "高" -> Triple(Color(0xFFEF4444), Color(0xFFFEE2E2), Color(0xFFDC2626))
        "低" -> Triple(Color(0xFF10B981), Color(0xFFD1FAE5), Color(0xFF059669))
        else -> Triple(Color(0xFFF59E0B), Color(0xFFFEF3C7), Color(0xFFD97706))
    }
}

private fun buildExportFilename(task: TaskDetail, extension: String): String {
    val date = SimpleDateFormat("yyyy-MM-dd", Locale.getDefault()).format(Date())
    return "scam-report-${task.task_id}-$date.$extension"
}

private fun buildTaskMarkdown(task: TaskDetail): String {
    return buildString {
        append("# 诈骗风险分析报告\n\n")
        append("**任务ID**: ${task.task_id}\n")
        append("**标题**: ${task.title.ifBlank { "未命名任务" }}\n")
        append("**诈骗类型**: ${task.scam_type.ifBlank { "未识别" }}\n")
        append("**风险分数**: ${task.risk_score}\n")
        append("**生成时间**: ${formatDateTime(task.created_at)}\n\n")
        if (task.summary.isNotBlank()) append("## 摘要\n${task.summary}\n\n")
        if (task.report.isNotBlank()) append("## 报告\n${task.report}\n\n")
        if (task.payload.text.isNotBlank()) append("## 原始文本\n${task.payload.text}\n")
    }
}

private fun saveTextToDownloads(
    context: android.content.Context,
    filename: String,
    mimeType: String,
    content: String,
    onMessage: (String, Boolean) -> Unit,
) {
    runCatching {
        val resolver = context.contentResolver
        val values = ContentValues().apply {
            put(MediaStore.Downloads.DISPLAY_NAME, filename)
            put(MediaStore.Downloads.MIME_TYPE, mimeType)
            put(MediaStore.Downloads.RELATIVE_PATH, Environment.DIRECTORY_DOWNLOADS)
        }
        val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, values) ?: error("无法创建导出文件")
        resolver.openOutputStream(uri)?.bufferedWriter(Charsets.UTF_8)?.use { writer -> writer.write(content) }
        onMessage("已保存到下载目录", false)
    }.onFailure {
        onMessage(it.message ?: "导出失败", true)
    }
}
