package com.example.myapplication

import android.graphics.BitmapFactory
import android.util.Base64
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.ArrowBack
import androidx.compose.material.icons.outlined.Description
import androidx.compose.material.icons.outlined.DeleteOutline
import androidx.compose.material.icons.outlined.KeyboardArrowRight
import androidx.compose.material.icons.outlined.Schedule
import androidx.compose.material.icons.outlined.Shield
import androidx.compose.material3.AssistChip
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.draw.scale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import java.time.LocalDate
import kotlin.math.roundToInt

@Composable
fun MobileBannerCard(
    banners: List<Int>,
    currentIndex: Int,
    onSelect: (Int) -> Unit,
) {
    Card(
        shape = RoundedCornerShape(28.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
    ) {
        Box(modifier = Modifier.fillMaxWidth().height(170.dp)) {
            androidx.compose.foundation.Image(
                painter = androidx.compose.ui.res.painterResource(banners[currentIndex]),
                contentDescription = null,
                modifier = Modifier.fillMaxWidth().height(170.dp),
                contentScale = ContentScale.Crop,
            )
            Row(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 12.dp),
                horizontalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                banners.indices.forEach { index ->
                    Box(
                        modifier = Modifier
                            .size(width = if (index == currentIndex) 16.dp else 8.dp, height = 8.dp)
                            .background(if (index == currentIndex) Color.White else Color.White.copy(alpha = 0.45f), RoundedCornerShape(999.dp))
                            .clickable { onSelect(index) },
                    )
                }
            }
        }
    }
}

@Composable
fun DashboardAction(title: String, iconName: String, tint: Color, modifier: Modifier = Modifier, onClick: () -> Unit) {
    Column(
        modifier = modifier.clickable(onClick = onClick).padding(vertical = 4.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Box(
            modifier = Modifier
                .size(48.dp)
                .background(tint.copy(alpha = 0.1f), RoundedCornerShape(16.dp))
                .border(1.dp, tint.copy(alpha = 0.08f), RoundedCornerShape(16.dp)),
            contentAlignment = Alignment.Center,
        ) {
            LucideSvgIcon(
                iconName = iconName,
                contentDescription = null,
                tint = tint,
                size = 20.dp,
            )
        }
        Text(title, fontSize = 11.sp, fontWeight = FontWeight.Medium, color = Color(0xFF475569), maxLines = 1)
    }
}

@Composable
fun MetricCard(label: String, value: String, accent: Color, iconName: String, modifier: Modifier = Modifier) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(24.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Brush.verticalGradient(listOf(Color.White, accent.copy(alpha = 0.03f))))
                .border(1.dp, MobileBorder.copy(alpha = 0.5f), RoundedCornerShape(24.dp))
                .padding(horizontal = 16.dp, vertical = 16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Box(
                modifier = Modifier
                    .size(28.dp)
                    .background(accent.copy(alpha = 0.12f), RoundedCornerShape(10.dp)),
                contentAlignment = Alignment.Center,
            ) {
                LucideSvgIcon(
                    iconName = iconName,
                    contentDescription = null,
                    tint = accent,
                    size = 14.dp,
                )
            }
            Text(value, fontSize = 22.sp, fontWeight = FontWeight.ExtraBold, color = MobileText)
            Text(label, fontSize = 11.sp, fontWeight = FontWeight.Medium, color = Color(0xFF94A3B8))
        }
    }
}

@Composable
fun AiInsightCard(
    headline: String,
    body: String,
    overallLabel: String,
    overallColor: Color,
    currentRiskLabel: String,
    currentRiskColor: Color,
) {
    Card(shape = RoundedCornerShape(28.dp), colors = CardDefaults.cardColors(containerColor = Color.Transparent)) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Brush.linearGradient(listOf(Color(0xFF0F172A), Color(0xFF334155))), RoundedCornerShape(28.dp))
                .padding(horizontal = 20.dp, vertical = 20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Box(
                modifier = Modifier
                    .background(Color(0x4D6366F1), RoundedCornerShape(8.dp))
                    .border(1.dp, Color(0x666366F1), RoundedCornerShape(8.dp))
                    .padding(horizontal = 10.dp, vertical = 4.dp),
            ) {
                Text("AI Insight", color = Color(0xFFBFDBFE), fontWeight = FontWeight.Black, fontSize = 9.sp)
            }
            Text(headline, color = Color.White, fontWeight = FontWeight.ExtraBold, fontSize = 16.sp, lineHeight = 22.sp)
            Text(body, color = Color(0xFF94A3B8), fontSize = 11.sp, lineHeight = 18.sp)
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(1.dp)
                    .background(Color.White.copy(alpha = 0.06f)),
            )
            Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                InsightMiniCard("整体走势", overallLabel, overallColor, Modifier.weight(1f))
                InsightMiniCard("当前风险", currentRiskLabel, currentRiskColor, Modifier.weight(1f))
            }
        }
    }
}

@Composable
private fun InsightMiniCard(label: String, value: String, valueColor: Color, modifier: Modifier = Modifier) {
    Card(modifier = modifier, shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.05f))) {
        Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Text(label, fontSize = 9.sp, color = Color(0xFF64748B), fontWeight = FontWeight.Black)
            Text(value, color = valueColor, fontWeight = FontWeight.ExtraBold, fontSize = 12.sp)
        }
    }
}

@Composable
fun MobileTaskRow(task: TaskSummary, onOpen: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onOpen),
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.Top,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    StatusChip(task.status)
                    Box(
                        modifier = Modifier
                            .background(Color(0xFFF8FAFC), RoundedCornerShape(8.dp))
                            .padding(horizontal = 6.dp, vertical = 4.dp),
                    ) {
                        Text(task.task_id.take(8), fontSize = 10.sp, color = Color(0xFF94A3B8))
                    }
                }
                Text(
                    task.title.ifBlank { "多模态风险检测任务" },
                    fontWeight = FontWeight.ExtraBold,
                    color = MobileText,
                    fontSize = 14.sp,
                    lineHeight = 20.sp,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        formatDateTime(task.created_at).let { if (it.length > 5) it.substring(5) else it },
                        color = Color(0xFF64748B),
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Medium,
                    )
                    Box(modifier = Modifier.size(4.dp).background(Color(0xFFCBD5E1), CircleShape))
                    Text(
                        task.summary.ifBlank { "等待分析详情" },
                        fontSize = 11.sp,
                        color = Color(0xFF64748B),
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
            }
            Box(
                modifier = Modifier
                    .size(32.dp)
                    .background(Color(0xFFF8FAFC), CircleShape),
                contentAlignment = Alignment.Center,
            ) {
                Icon(Icons.Outlined.KeyboardArrowRight, contentDescription = null, tint = Color(0xFF94A3B8), modifier = Modifier.size(18.dp))
            }
        }
    }
}

@Composable
fun MobileBackHeader(title: String, subtitle: String = "", onBack: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(32.dp)
                .background(Color(0xFFF8FAFC), CircleShape)
                .clickable(onClick = onBack),
            contentAlignment = Alignment.Center,
        ) {
            Icon(Icons.Outlined.ArrowBack, contentDescription = "返回", modifier = Modifier.size(18.dp), tint = MobileText)
        }
        Column(
            modifier = Modifier.weight(1f).padding(start = 12.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(title, fontWeight = FontWeight.ExtraBold, fontSize = 18.sp, color = MobileText)
            if (subtitle.isNotBlank()) Text(subtitle, fontSize = 11.sp, color = Color(0xFF94A3B8))
        }
        Spacer(modifier = Modifier.width(32.dp))
    }
}

@Composable
fun SecurityOverviewCard(totalCount: Int) {
    val pulse = rememberInfiniteTransition(label = "securityPulse")
    val pulseScale = pulse.animateFloat(
        initialValue = 0.92f,
        targetValue = 1.08f,
        animationSpec = infiniteRepeatable(animation = tween(1800), repeatMode = RepeatMode.Reverse),
        label = "securityPulseScale",
    )

    Card(shape = RoundedCornerShape(32.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp, vertical = 20.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(verticalArrangement = Arrangement.spacedBy(6.dp), modifier = Modifier.weight(1f)) {
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(modifier = Modifier.size(6.dp).background(MobileGreen, CircleShape))
                    Text("System Secured", fontSize = 10.sp, fontWeight = FontWeight.Black, color = MobileGreen)
                }
                Text("全域实时守护中", fontSize = 20.sp, fontWeight = FontWeight.ExtraBold, color = MobileText)
                Text("累计检测 $totalCount 次风险内容", fontSize = 12.sp, color = Color(0xFF94A3B8))
            }
            Box(
                modifier = Modifier.size(80.dp),
                contentAlignment = Alignment.Center,
            ) {
                Box(
                    modifier = Modifier
                        .size(80.dp)
                        .scale(pulseScale.value)
                        .background(MobileGreen.copy(alpha = 0.1f), CircleShape),
                )
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .background(Color.White, CircleShape)
                        .border(2.dp, MobileGreen.copy(alpha = 0.16f), CircleShape),
                    contentAlignment = Alignment.Center,
                ) {
                    LucideSvgIcon(
                        iconName = "check",
                        contentDescription = null,
                        tint = MobileGreen,
                        size = 28.dp,
                    )
                }
            }
        }
    }
}

@Composable
fun HistoryOverviewCard(history: List<HistoryRecord>) {
    val highRiskCount = history.count { normalizeRiskLevel(it.risk_level) == "高" }

    Card(
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.92f)),
        border = BorderStroke(1.dp, Color(0x57E2E8F0)),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(15.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.Top) {
                Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Text("风险记录档案库", fontSize = 18.sp, fontWeight = FontWeight.ExtraBold, color = MobileText)
                    Text("保留近期检测记录与风险判定，便于回看与复盘。", fontSize = 11.sp, lineHeight = 18.sp, color = Color(0xFF64748B))
                }
                Box(
                    modifier = Modifier
                        .size(36.dp)
                        .background(
                            Brush.verticalGradient(listOf(Color.White, Color(0xFFF8FAFC))),
                            RoundedCornerShape(14.dp),
                        )
                        .border(1.dp, Color(0x57E2E8F0), RoundedCornerShape(14.dp)),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Outlined.Description, contentDescription = null, tint = MobileText, modifier = Modifier.size(18.dp))
                }
            }
            Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                HistoryOverviewMetric("档案总数", history.size.toString(), Modifier.weight(1f))
                HistoryOverviewMetric("高危记录", highRiskCount.toString(), Modifier.weight(1f))
            }
        }
    }
}

@Composable
private fun HistoryOverviewMetric(label: String, value: String, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier
            .background(Brush.verticalGradient(listOf(Color.White, Color(0xFFF8FAFC))), RoundedCornerShape(16.dp))
            .border(1.dp, Color(0x4DE2E8F0), RoundedCornerShape(16.dp))
            .padding(horizontal = 12.dp, vertical = 12.dp),
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Text(label, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFF94A3B8))
        Text(value, fontSize = 24.sp, fontWeight = FontWeight.ExtraBold, color = MobileText)
    }
}

@Composable
fun RiskHeroCard(title: String, detail: String, footer: String) {
    Card(shape = RoundedCornerShape(26.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MobileBorder.copy(alpha = 0.48f), RoundedCornerShape(26.dp))
                .padding(horizontal = 18.dp, vertical = 18.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.Top) {
                Box(
                    modifier = Modifier
                        .size(38.dp)
                        .background(Brush.verticalGradient(listOf(Color(0xFFE8FFF7), Color(0xFFD8FAEE))), RoundedCornerShape(14.dp)),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Outlined.Shield, contentDescription = null, tint = MobileGreen, modifier = Modifier.size(18.dp))
                }
                Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Text(title, fontSize = 18.sp, fontWeight = FontWeight.ExtraBold, color = MobileText, lineHeight = 24.sp)
                    Text(detail, fontSize = 12.sp, lineHeight = 20.sp, color = Color(0xFF64748B))
                }
            }
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                Text("最近高发:", fontSize = 12.sp, fontWeight = FontWeight.Bold, color = Color(0xFF94A3B8))
                Text(footer, fontSize = 12.sp, fontWeight = FontWeight.ExtraBold, color = Color(0xFF334155))
            }
        }
    }
}

@Composable
fun SectionTitle(title: String, action: String = "", actionColor: Color = MobileGreen, onAction: (() -> Unit)? = null) {
    Row(modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
        Row(modifier = Modifier.weight(1f), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Box(
                modifier = Modifier
                    .width(4.dp)
                    .height(18.dp)
                    .background(MobileGreen, RoundedCornerShape(999.dp)),
            )
            Text(title, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 15.sp)
        }
        if (action.isNotBlank() && onAction != null) TextButton(onClick = onAction) { Text(action, color = actionColor, fontWeight = FontWeight.Bold) }
    }
}

@Composable
fun EmptyCard(text: String) {
    Card(shape = RoundedCornerShape(24.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MobileBorder.copy(alpha = 0.45f), RoundedCornerShape(24.dp))
                .padding(vertical = 36.dp, horizontal = 16.dp),
            contentAlignment = Alignment.Center,
        ) {
            Text(text, color = Color(0xFF94A3B8), fontWeight = FontWeight.Medium, textAlign = TextAlign.Center)
        }
    }
}

@Composable
fun HomeRecentTasksEmptyCard() {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color(0xFFF8FAFC), RoundedCornerShape(24.dp))
            .drawBehind {
                drawRoundRect(
                    color = Color(0xFFD8E0EA),
                    cornerRadius = CornerRadius(24.dp.toPx(), 24.dp.toPx()),
                    style = Stroke(
                        width = 2.dp.toPx(),
                        pathEffect = PathEffect.dashPathEffect(floatArrayOf(14.dp.toPx(), 10.dp.toPx())),
                    ),
                )
            }
            .padding(vertical = 40.dp, horizontal = 16.dp),
        contentAlignment = Alignment.Center,
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .background(Color.White, CircleShape),
                contentAlignment = Alignment.Center,
            ) {
                LucideSvgIcon(
                    iconName = "inbox",
                    contentDescription = null,
                    tint = Color(0xFFCBD5E1),
                    size = 22.dp,
                )
            }
            Text("暂无进行中的任务", color = Color(0xFF94A3B8), fontWeight = FontWeight.Medium, textAlign = TextAlign.Center)
        }
    }
}

@Composable
fun HistoryArchiveCard(item: HistoryRecord, deleting: Boolean, onOpen: () -> Unit, onDelete: () -> Unit) {
    val normalizedRisk = normalizeRiskLevel(item.risk_level)
    val (accentBrush, badgeBackground, badgeText) = when (normalizedRisk) {
        "高" -> Triple(
            Brush.verticalGradient(listOf(Color(0xEEFB7185), Color(0xA6F43F5E))),
            Color(0xF5FFF1F2),
            Color(0xFFBE123C),
        )
        "低" -> Triple(
            Brush.verticalGradient(listOf(Color(0xE634D399), Color(0x9910B981))),
            Color(0xF5ECFDF5),
            Color(0xFF047857),
        )
        else -> Triple(
            Brush.verticalGradient(listOf(Color(0xE6FBBF24), Color(0x9EF59E0B))),
            Color(0xF5FFFBEB),
            Color(0xFFB45309),
        )
    }
    val createdAtLabel = formatDateTime(item.created_at).let { formatted ->
        if (formatted.length >= 16) formatted.substring(5, 16) else formatted
    }
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onOpen),
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = BorderStroke(1.dp, Color(0x57E2E8F0)),
    ) {
        Box(modifier = Modifier.fillMaxWidth()) {
            Box(
                modifier = Modifier
                    .align(Alignment.CenterStart)
                    .width(3.dp)
                    .fillMaxHeight()
                    .padding(vertical = 14.dp)
                    .background(accentBrush, RoundedCornerShape(topEnd = 999.dp, bottomEnd = 999.dp)),
            )
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(start = 16.dp, end = 14.dp, top = 14.dp, bottom = 14.dp),
                verticalAlignment = Alignment.Top,
                horizontalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Column(
                    modifier = Modifier.weight(1f),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        if (item.risk_level.isNotBlank()) {
                            Box(
                                modifier = Modifier
                                    .background(badgeBackground, RoundedCornerShape(999.dp))
                                    .padding(horizontal = 8.dp, vertical = 4.dp),
                            ) {
                                Text(normalizedRisk, fontSize = 9.sp, fontWeight = FontWeight.Black, color = badgeText)
                            }
                        }
                        Box(
                            modifier = Modifier
                                .background(Color(0xFFF8FAFC), RoundedCornerShape(999.dp))
                                .padding(horizontal = 6.dp, vertical = 4.dp),
                        ) {
                            Text(item.record_id.take(8), fontSize = 9.sp, fontWeight = FontWeight.Black, color = Color(0xFF64748B))
                        }
                        Box(
                            modifier = Modifier
                                .background(Color(0xFFF8FAFC), RoundedCornerShape(999.dp))
                                .padding(horizontal = 6.dp, vertical = 4.dp),
                        ) {
                            Text(item.scam_type.ifBlank { "未知类型" }, fontSize = 9.sp, fontWeight = FontWeight.Black, color = Color(0xFF64748B))
                        }
                    }
                    Text(
                        item.title.ifBlank { "无标题检测记录" },
                        fontWeight = FontWeight.ExtraBold,
                        color = MobileText,
                        fontSize = 14.sp,
                        lineHeight = 20.sp,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Icon(
                            Icons.Outlined.Schedule,
                            contentDescription = null,
                            tint = Color(0xFF94A3B8),
                            modifier = Modifier.size(14.dp),
                        )
                        Text(
                            createdAtLabel,
                            fontSize = 10.sp,
                            color = Color(0xFF64748B),
                            fontWeight = FontWeight.Medium,
                        )
                    }
                }
                IconButton(
                    onClick = onDelete,
                    enabled = !deleting,
                    modifier = Modifier
                        .size(30.dp)
                        .background(Color(0xFFF8FAFC), CircleShape),
                ) {
                    Icon(
                        Icons.Outlined.DeleteOutline,
                        contentDescription = "删除",
                        tint = if (deleting) Color(0xFFCBD5E1) else Color(0xFF94A3B8),
                        modifier = Modifier.size(15.dp),
                    )
                }
            }
        }
    }
}

@Composable
fun HistoryEmptyStateCard() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.92f)),
        border = BorderStroke(1.dp, Color(0x57E2E8F0)),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .height(240.dp)
                .padding(horizontal = 16.dp, vertical = 28.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center,
        ) {
            Box(
                modifier = Modifier
                    .size(68.dp)
                    .background(Brush.verticalGradient(listOf(Color.White, Color(0xFFF8FAFC))), CircleShape)
                    .border(1.dp, Color(0x57E2E8F0), CircleShape),
                contentAlignment = Alignment.Center,
            ) {
                Icon(Icons.Outlined.Description, contentDescription = null, tint = Color(0xFFCBD5E1), modifier = Modifier.size(32.dp))
            }
            Spacer(Modifier.height(16.dp))
            Text("暂无历史档案", fontSize = 16.sp, fontWeight = FontWeight.Bold, color = MobileText)
            Text("您的检测记录将在这里安全保存", fontSize = 12.sp, color = Color(0xFF94A3B8))
        }
    }
}

@Composable
fun RiskChip(level: String) {
    val normalized = normalizeRiskLevel(level)
    val (bg, fg) = when (normalized) {
        "高" -> Color(0xFFFEE2E2) to Color(0xFFDC2626)
        "低" -> Color(0xFFD1FAE5) to Color(0xFF059669)
        else -> Color(0xFFFEF3C7) to Color(0xFFD97706)
    }
    Box(modifier = Modifier.background(bg, RoundedCornerShape(12.dp)).padding(horizontal = 10.dp, vertical = 5.dp)) {
        Text(normalized, fontSize = 10.sp, fontWeight = FontWeight.Black, color = fg)
    }
}

@Composable
fun MarkdownText(text: String, color: Color) {
    MarkdownText(
        markdown = text,
        color = color,
        style = MaterialTheme.typography.bodyMedium.copy(lineHeight = 20.sp),
    )
}

@Composable
fun Base64Thumbnail(
    dataUrl: String,
    size: androidx.compose.ui.unit.Dp,
    modifier: Modifier = Modifier,
) {
    val bitmap = remember(dataUrl) {
        val payload = dataUrl.substringAfter("base64,", "")
        if (payload.isBlank()) null else runCatching {
            BitmapFactory.decodeByteArray(Base64.decode(payload, Base64.DEFAULT), 0, Base64.decode(payload, Base64.DEFAULT).size)
        }.getOrNull()
    }
    val imageModifier = if (modifier == Modifier) {
        Modifier.size(size).background(Color.White, RoundedCornerShape(16.dp))
    } else {
        modifier.background(Color.White, RoundedCornerShape(16.dp))
    }
    if (bitmap != null) {
        androidx.compose.foundation.Image(
            bitmap = bitmap.asImageBitmap(),
            contentDescription = null,
            modifier = imageModifier,
            contentScale = ContentScale.Crop,
        )
    }
}

fun normalizeRiskLevel(value: String): String = when {
    value.contains("高") -> "高"
    value.contains("低") -> "低"
    else -> "中"
}

fun currentInsightHeadline(analysis: RiskAnalysis?): String = when {
    analysis == null -> "AI 研判：近期暂无风险趋势摘要"
    analysis.overall_trend == "上升" && analysis.high_risk_trend == "上升" -> "AI 研判：风险热度正在拉升"
    analysis.overall_trend == "下降" && analysis.high_risk_trend == "下降" -> "AI 研判：风险热度出现回落"
    analysis.high_risk_trend == "上升" -> "AI 研判：高危暴露正在增强"
    analysis.overall_trend == "上升" -> "AI 研判：整体风险有所抬头"
    analysis.overall_trend == "下降" -> "AI 研判：整体风险趋于缓和"
    analysis.overall_trend == "平稳" && analysis.high_risk_trend == "平稳" -> "AI 研判：当前波动较为平稳"
    else -> "AI 研判：风险走势仍需持续观察"
}

fun currentInsightDescriptor(value: String?, high: Boolean = false): String = when (value.orEmpty()) {
    "上升" -> if (high) "高危暴露增强" else "风险热度抬升"
    "下降" -> if (high) "高危暴露收敛" else "风险热度回落"
    "平稳" -> if (high) "高危信号平稳" else "风险热度平稳"
    else -> if (high) "高危信号待观察" else "风险热度待观察"
}

fun homeOverallTrendLabel(value: String?): String = when (value.orEmpty()) {
    "上升" -> "风险热度拉升"
    "下降" -> "风险热度回落"
    "平稳" -> "风险热度平稳"
    else -> "风险热度待观察"
}

fun homeOverallTrendColor(value: String?): Color = when (value.orEmpty()) {
    "上升" -> Color(0xFFFB7185)
    "下降" -> Color(0xFF34D399)
    "平稳" -> Color(0xFFFBBF24)
    else -> Color(0xFFCBD5E1)
}

fun homeCurrentRiskLabel(stats: RiskStats?): String {
    val total = stats?.total ?: 0
    val high = stats?.high ?: 0
    val medium = stats?.medium ?: 0
    return when {
        total <= 0 -> "待观察"
        high > 0 || high.toFloat() / total.coerceAtLeast(1) >= 0.2f -> "高风险"
        medium > 0 -> "中风险"
        else -> "低风险"
    }
}

fun homeCurrentRiskColor(label: String): Color = when (label) {
    "高风险" -> Color(0xFFFB7185)
    "中风险" -> Color(0xFFFBBF24)
    "低风险" -> Color.White
    else -> Color(0xFFCBD5E1)
}

fun formatTrendBucketLabel(label: String): String {
    return when {
        label.contains("-W") -> runCatching {
            val (yearValue, weekValue) = label.split("-W")
            val year = yearValue.toInt()
            val week = weekValue.toInt()
            val jan4 = LocalDate.of(year, 1, 4)
            val week1Start = jan4.minusDays((jan4.dayOfWeek.value - 1).toLong())
            val start = week1Start.plusWeeks((week - 1).toLong())
            val end = start.plusDays(6)
            "${year}年第${week}周 (${start.monthValue}.${start.dayOfMonth}-${end.monthValue}.${end.dayOfMonth})"
        }.getOrElse { label }
        Regex("""^\d{4}-\d{2}$""").matches(label) -> runCatching {
            val (year, month) = label.split("-")
            "${year}年${month.toInt()}月"
        }.getOrElse { label }
        else -> label
    }
}

private fun detectTrendInterval(points: List<RiskTrendPoint>): String = when {
    points.any { it.time_bucket.contains("-W") } -> "week"
    points.any { Regex("""^\d{4}-\d{2}$""").matches(it.time_bucket) } -> "month"
    else -> "day"
}

private fun fillTrendGaps(points: List<RiskTrendPoint>): List<RiskTrendPoint> {
    if (points.isEmpty()) return emptyList()
    val sorted = points.sortedBy { it.time_bucket }
    val interval = detectTrendInterval(sorted)
    val dataMap = sorted.associateBy { it.time_bucket }
    val filled = mutableListOf<RiskTrendPoint>()
    var current = sorted.first().time_bucket
    val end = sorted.last().time_bucket
    var guard = 0

    while (current <= end && guard < 500) {
        guard += 1
        filled += dataMap[current] ?: RiskTrendPoint(time_bucket = current)
        val nextBucket = when (interval) {
            "day" -> runCatching { LocalDate.parse(current).plusDays(1).toString() }.getOrNull()
            "week" -> runCatching {
                val (yearValue, weekValue) = current.split("-W")
                var year = yearValue.toInt()
                var week = weekValue.toInt() + 1
                if (week > 53) {
                    week = 1
                    year += 1
                }
                "%04d-W%02d".format(year, week)
            }.getOrNull()
            "month" -> runCatching {
                val (yearValue, monthValue) = current.split("-")
                var year = yearValue.toInt()
                var month = monthValue.toInt() + 1
                if (month > 12) {
                    month = 1
                    year += 1
                }
                "%04d-%02d".format(year, month)
            }.getOrNull()
            else -> null
        } ?: break
        current = nextBucket
    }
    return filled
}

fun recentTrendRows(points: List<RiskTrendPoint>, limit: Int): List<RiskTrendPoint> {
    val normalizedLimit = limit.coerceAtLeast(1)
    return fillTrendGaps(points).takeLast(normalizedLimit).reversed()
}

fun buildTrendDeltaText(current: Int, previous: Int): String {
    if (previous <= 0) return if (current > 0) "+100%" else "--"
    val delta = (((current - previous).toDouble() / previous.toDouble()) * 100.0).roundToInt()
    return if (delta == 0) "--" else "${if (delta > 0) "+" else ""}$delta%"
}

fun buildRecentTrendHeadline(point: RiskTrendPoint): String {
    return when {
        point.total <= 0 -> "当日暂无新增检测记录"
        point.high > 0 -> "${point.high} 笔高危信号，建议优先核验来源"
        point.medium > 0 && point.medium >= point.low -> "中风险样本占比更高，注意退款客服与陌生链接"
        else -> "以常规预警样本为主，整体波动较稳"
    }
}

fun trendDotColor(point: RiskTrendPoint): Color = when {
    point.high > 0 -> Color(0xFFF43F5E)
    point.medium > 0 -> Color(0xFFF59E0B)
    else -> Color(0xFF4F46E5)
}

data class RegionSignalSummary(
    val level: String,
    val title: String,
    val detail: String,
)

fun regionSignalSummary(stats: CurrentRegionCaseStatsResponse): RegionSignalSummary {
    val summary = stats.summary
        ?: return RegionSignalSummary(
            level = "neutral",
            title = "地区风险态势待补充",
            detail = "当前统计样本不足，建议继续观察。",
        )

    val todayCount = summary.today_count
    val last7dCount = summary.last_7d_count
    val totalCount = summary.total_count
    val highCount = summary.high_count
    val weeklyAverage = if (last7dCount > 0) last7dCount / 7.0 else 0.0
    val highRiskRatio = if (totalCount > 0) highCount.toDouble() / totalCount.toDouble() else 0.0

    return when {
        todayCount.toDouble() >= maxOf(3.0, weeklyAverage * 1.5) || highRiskRatio >= 0.35 -> RegionSignalSummary(
            level = "high",
            title = "近期风险有抬升信号",
            detail = "今日增量或高风险占比偏高，建议减少陌生转账与验证码操作。",
        )
        todayCount.toDouble() >= maxOf(1.0, weeklyAverage * 0.8) || highRiskRatio >= 0.2 -> RegionSignalSummary(
            level = "medium",
            title = "近期风险保持活跃",
            detail = "建议重点核验陌生来电、客服退款与投资荐股类话术。",
        )
        else -> RegionSignalSummary(
            level = "low",
            title = "近期风险相对平稳",
            detail = "仍需保持警惕，遇到催转账和索要验证码请先核实。",
        )
    }
}

fun regionTopScamHint(stats: CurrentRegionCaseStatsResponse): String {
    val topItem = stats.top_scam_types.firstOrNull()
        ?: return "近期未形成明显高发类型，注意通用防诈规则。"
    val scamType = topItem.scam_type.ifBlank { "当前高发类型" }
    return "最近高发：$scamType${if (topItem.count > 0) "（${topItem.count}起）" else ""}，同类话术请优先核验来源。"
}

fun regionLocationText(stats: CurrentRegionCaseStatsResponse): String {
    val region = stats.region ?: return "浙江省 / 杭州市 / 钱塘区"
    val parts = when (region.granularity) {
        "county", "district" -> listOf(region.province_name, region.city_name, region.district_name)
        else -> listOf(region.province_name, region.city_name)
    }
    return parts.filter { it.isNotBlank() }.joinToString(" / ").ifBlank { "浙江省 / 杭州市 / 钱塘区" }
}

fun examProgress(state: MainUiState): Float {
    val steps = state.simulationPack?.steps.orEmpty()
    if (steps.isEmpty()) return 0f
    return (((state.simulationAnswers.size + 1).coerceAtMost(steps.size)).toFloat() / steps.size.toFloat()).coerceIn(0f, 1f)
}

fun combinedAlerts(state: MainUiState): List<AlertEvent> {
    return state.alertItems
        .map { it.event }
        .distinctBy { it.record_id }
        .sortedByDescending { parseInstant(it.created_at.ifBlank { it.sent_at }) }
}
