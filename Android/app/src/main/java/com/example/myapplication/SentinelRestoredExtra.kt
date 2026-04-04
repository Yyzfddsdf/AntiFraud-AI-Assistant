package com.example.myapplication

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Add
import androidx.compose.material.icons.outlined.Check
import androidx.compose.material.icons.outlined.DeleteOutline
import androidx.compose.material.icons.outlined.Image
import androidx.compose.material.icons.outlined.KeyboardArrowRight
import androidx.compose.material.icons.outlined.Place
import androidx.compose.material.icons.outlined.Shield
import androidx.compose.material3.AssistChip
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

enum class GuardFeatureIconKind {
    QuickAnalyze,
    Accessibility,
}

private data class FeatureIconPalette(
    val backgroundBrush: Brush,
    val borderColor: Color,
    val shieldTint: Color,
    val checkTint: Color,
    val accentColor: Color,
)

@Composable
private fun ShieldCheckGlyph(
    modifier: Modifier = Modifier,
    shieldTint: Color,
    checkTint: Color,
) {
    Box(modifier = modifier, contentAlignment = Alignment.Center) {
        Icon(
            Icons.Outlined.Shield,
            contentDescription = null,
            tint = shieldTint,
            modifier = Modifier.size(22.dp),
        )
        Icon(
            Icons.Outlined.Check,
            contentDescription = null,
            tint = checkTint,
            modifier = Modifier.size(12.dp),
        )
    }
}

@Composable
fun BrandShieldIconTile(modifier: Modifier = Modifier) {
    Box(
        modifier = modifier
            .background(
                Brush.linearGradient(listOf(Color(0xFF34D399), Color(0xFF14B8A6))),
                RoundedCornerShape(14.dp),
            )
            .border(1.dp, Color.White.copy(alpha = 0.22f), RoundedCornerShape(14.dp)),
        contentAlignment = Alignment.Center,
    ) {
        Box(
            modifier = Modifier
                .size(30.dp)
                .background(Color.White.copy(alpha = 0.18f), CircleShape),
            contentAlignment = Alignment.Center,
        ) {
            LucideSvgIcon(
                iconName = "shield-check",
                contentDescription = null,
                tint = Color.White,
                size = 20.dp,
            )
        }
    }
}

@Composable
fun GuardFeatureIconTile(
    kind: GuardFeatureIconKind,
    modifier: Modifier = Modifier,
) {
    val palette = when (kind) {
        GuardFeatureIconKind.QuickAnalyze -> FeatureIconPalette(
            backgroundBrush = Brush.linearGradient(listOf(Color(0xFFE0F2FE), Color(0xFFDCFCE7))),
            borderColor = Color(0xFFBAE6FD),
            shieldTint = Color(0xFF0F766E),
            checkTint = Color(0xFF10B981),
            accentColor = Color(0xFF22C55E),
        )
        GuardFeatureIconKind.Accessibility -> FeatureIconPalette(
            backgroundBrush = Brush.linearGradient(listOf(Color(0xFFFFF7ED), Color(0xFFECFDF5))),
            borderColor = Color(0xFFFCD34D),
            shieldTint = Color(0xFF92400E),
            checkTint = Color(0xFFF59E0B),
            accentColor = Color(0xFFF59E0B),
        )
    }

    Box(
        modifier = modifier
            .size(46.dp)
            .background(palette.backgroundBrush, RoundedCornerShape(16.dp))
            .border(1.dp, palette.borderColor, RoundedCornerShape(16.dp)),
        contentAlignment = Alignment.Center,
    ) {
        if (kind == GuardFeatureIconKind.Accessibility) {
            Box(
                modifier = Modifier
                    .size(34.dp)
                    .border(1.5.dp, palette.accentColor.copy(alpha = 0.28f), CircleShape),
            )
        }
        Box(
            modifier = Modifier
                .size(34.dp)
                .background(Color.White.copy(alpha = 0.88f), CircleShape),
            contentAlignment = Alignment.Center,
        ) {
            ShieldCheckGlyph(
                modifier = Modifier.size(24.dp),
                shieldTint = palette.shieldTint,
                checkTint = palette.checkTint,
            )
        }
        Box(
            modifier = Modifier
                .align(Alignment.TopEnd)
                .offset(x = (-5).dp, y = 5.dp)
                .size(if (kind == GuardFeatureIconKind.QuickAnalyze) 10.dp else 8.dp)
                .background(palette.accentColor, CircleShape),
        )
        if (kind == GuardFeatureIconKind.QuickAnalyze) {
            Box(
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .offset(x = 6.dp, y = (-6).dp)
                    .size(8.dp)
                    .background(Color.White.copy(alpha = 0.92f), CircleShape),
            )
        }
    }
}

@Composable
fun StatusChip(status: String) {
    val (bg, fg, text) = when (status) {
        "pending" -> Triple(Color(0xFFFEF3C7), Color(0xFFD97706), "等待中")
        "processing" -> Triple(Color(0xFFDBEAFE), Color(0xFF2563EB), "分析中")
        "completed" -> Triple(Color(0xFFD1FAE5), Color(0xFF059669), "已完成")
        else -> Triple(Color(0xFFFEE2E2), Color(0xFFDC2626), "失败")
    }
    Box(modifier = Modifier.background(bg, RoundedCornerShape(12.dp)).padding(horizontal = 10.dp, vertical = 5.dp)) {
        Text(text, fontSize = 10.sp, fontWeight = FontWeight.Black, color = fg)
    }
}

@Composable
fun RiskSummaryFeatureCard(
    label: String,
    value: Int,
    deltaText: String,
    deltaColor: Color,
    accent: Color,
    icon: ImageVector,
    modifier: Modifier = Modifier,
) {
    Card(
        modifier = modifier.height(152.dp),
        shape = RoundedCornerShape(24.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.88f)),
        border = BorderStroke(1.dp, Color(0x75E2E8F0)),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(18.dp),
            verticalArrangement = Arrangement.SpaceBetween,
        ) {
            Box(
                modifier = Modifier
                    .size(42.dp)
                    .background(
                        Brush.verticalGradient(
                            if (accent == MobileRose) {
                                listOf(Color(0xFFFFF1F2), Color(0xFFFFE4E6))
                            } else {
                                listOf(Color(0xFFECFDF5), Color(0xFFDCFCE7))
                            },
                        ),
                        RoundedCornerShape(16.dp),
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Icon(icon, contentDescription = null, tint = accent, modifier = Modifier.size(18.dp))
            }
            Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Text(label, fontSize = 12.sp, fontWeight = FontWeight.Bold, color = Color(0xFF94A3B8))
                Row(verticalAlignment = Alignment.Bottom, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                    Text(value.toString(), fontSize = 27.sp, lineHeight = 27.sp, fontWeight = FontWeight.Black, color = MobileText)
                    Text(
                        deltaText,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Bold,
                        color = if (deltaText == "--") Color(0xFF94A3B8) else deltaColor,
                        modifier = Modifier.padding(bottom = 2.dp),
                    )
                }
            }
        }
    }
}

@Composable
fun RegionStatsCard(stats: CurrentRegionCaseStatsResponse) {
    var selectedWindow by remember { mutableStateOf("week") }
    val summary = stats.summary
    val windowCards = listOf(
        Triple("day", "今日", summary?.today_count ?: 0),
        Triple("week", "近7天", summary?.last_7d_count ?: 0),
        Triple("month", "近30天", summary?.last_30d_count ?: 0),
    )
    val activeNote = when (selectedWindow) {
        "day" -> "今日新增态势，用于观察短时抬升。"
        "month" -> "近30天累计样本适合判断长期变化。"
        else -> "近7天波动更能反映当前风险热度。"
    }
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(24.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.88f)),
        border = BorderStroke(1.dp, Color(0x75E2E8F0)),
    ) {
        Column(
            modifier = Modifier.padding(18.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Row(modifier = Modifier.weight(1f), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(9.dp)) {
                    Box(
                        modifier = Modifier
                            .width(5.dp)
                            .height(24.dp)
                            .background(Brush.verticalGradient(listOf(Color(0xFF4F46E5), Color(0xFF7367FF))), RoundedCornerShape(999.dp)),
                    )
                    Text("所在区态势", fontSize = 17.sp, fontWeight = FontWeight.Black, color = MobileText)
                }
                Box(
                    modifier = Modifier
                        .background(Color(0xFFF1F5F9), RoundedCornerShape(999.dp))
                        .padding(horizontal = 10.dp, vertical = 6.dp),
                ) {
                    Text(
                        "总计 ${summary?.total_count ?: 0} 起",
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Black,
                        color = Color(0xFF94A3B8),
                    )
                }
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(
                        Brush.verticalGradient(listOf(Color(0xFFF8FAFC), Color.White)),
                        RoundedCornerShape(20.dp),
                    )
                    .border(1.dp, Color(0x6BE2E8F0), RoundedCornerShape(20.dp))
                    .padding(horizontal = 14.dp, vertical = 16.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Box(
                    modifier = Modifier
                        .size(34.dp)
                        .background(Color(0x1A6366F1), CircleShape),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Outlined.Place, contentDescription = null, tint = Color(0xFF4F46E5), modifier = Modifier.size(17.dp))
                }
                Text(
                    regionLocationText(stats),
                    fontSize = 15.sp,
                    fontWeight = FontWeight.ExtraBold,
                    color = MobileText,
                    modifier = Modifier.weight(1f),
                )
            }

            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), modifier = Modifier.fillMaxWidth()) {
                windowCards.forEach { (key, label, value) ->
                    RegionMetricMini(
                        label = label,
                        value = value,
                        selected = selectedWindow == key,
                        modifier = Modifier.weight(1f),
                    ) {
                        selectedWindow = key
                    }
                }
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Box(
                    modifier = Modifier
                        .size(7.dp)
                        .background(Color(0xFF4F46E5), CircleShape),
                )
                Text(activeNote, fontSize = 12.sp, fontWeight = FontWeight.Bold, color = Color(0xFF64748B))
            }

            if (stats.top_scam_types.isNotEmpty()) {
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text("高发骗局 TOP 5", fontSize = 12.sp, fontWeight = FontWeight.Black, color = Color(0xFF94A3B8))
                    stats.top_scam_types.take(5).forEachIndexed { index, item ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(
                                    Brush.verticalGradient(listOf(Color.White.copy(alpha = 0.98f), Color(0xFFF8FAFC))),
                                    RoundedCornerShape(18.dp),
                                )
                                .border(1.dp, Color(0x61E2E8F0), RoundedCornerShape(18.dp))
                                .padding(horizontal = 16.dp, vertical = 14.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            Row(
                                modifier = Modifier.weight(1f),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(12.dp),
                            ) {
                                Box(
                                    modifier = Modifier
                                        .size(24.dp)
                                        .background(Color(0xFFEEF2FF), CircleShape),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    Text("${index + 1}", fontSize = 11.sp, fontWeight = FontWeight.Black, color = Color(0xFF64748B))
                                }
                                Text(
                                    item.scam_type,
                                    fontSize = 15.sp,
                                    fontWeight = FontWeight.ExtraBold,
                                    color = MobileText,
                                    modifier = Modifier.weight(1f),
                                    maxLines = 1,
                                    overflow = TextOverflow.Ellipsis,
                                )
                            }
                            Text(item.count.toString(), fontSize = 18.sp, fontWeight = FontWeight.Black, color = Color(0xFF4F46E5))
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun RegionMetricMini(
    label: String,
    value: Int,
    selected: Boolean,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    val borderColor by animateColorAsState(if (selected) Color(0x804F46E5) else Color(0x61E2E8F0), label = "regionChipBorder")
    val valueColor by animateColorAsState(if (selected) Color.White else MobileText, label = "regionChipValue")
    val labelColor by animateColorAsState(if (selected) Color.White else Color(0xFF94A3B8), label = "regionChipLabel")
    Column(
        modifier = modifier
            .background(
                if (selected) Brush.verticalGradient(listOf(Color(0xFF5E4EF7), Color(0xFF4F46E5))) else Brush.verticalGradient(listOf(Color.White, Color(0xFFF8FAFC))),
                RoundedCornerShape(18.dp),
            )
            .border(1.dp, borderColor, RoundedCornerShape(18.dp))
            .clickable(onClick = onClick)
            .padding(vertical = 14.dp, horizontal = 12.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(label, fontSize = 11.sp, color = labelColor, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(8.dp))
        Text("$value", fontSize = 29.sp, lineHeight = 29.sp, color = valueColor, fontWeight = FontWeight.Black)
    }
}

@Composable
fun TrendRowCard(point: RiskTrendPoint) {
    val dotColor = trendDotColor(point)
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(24.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.88f)),
        border = BorderStroke(1.dp, Color(0x75E2E8F0)),
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Box(
                            modifier = Modifier
                                .size(9.dp)
                                .background(dotColor, CircleShape),
                        )
                        Text(formatTrendBucketLabel(point.time_bucket), color = MobileText, fontWeight = FontWeight.Black, fontSize = 17.sp)
                    }
                    Text(buildRecentTrendHeadline(point), color = Color(0xFF64748B), fontWeight = FontWeight.Medium, fontSize = 12.sp)
                }
                Box(
                    modifier = Modifier
                        .background(Color(0xFFF1F5F9), RoundedCornerShape(999.dp))
                        .padding(horizontal = 10.dp, vertical = 6.dp),
                ) {
                    Text("总计 ${point.total} 笔", fontSize = 11.sp, fontWeight = FontWeight.Black, color = Color(0xFF94A3B8))
                }
            }
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(1.dp)
                    .background(Color(0xFFE2E8F0)),
            )
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                TrendMetricBox("高危", point.high, Color(0xFFE11D48), Color(0xEBFFF1F2), Modifier.weight(1f))
                TrendMetricBox("中危", point.medium, Color(0xFFD97706), Color(0xF5FFFBEB), Modifier.weight(1f))
                TrendMetricBox("安全", point.low, Color(0xFF059669), Color(0xF5ECFDF5), Modifier.weight(1f))
            }
        }
    }
}

@Composable
private fun TrendMetricBox(
    label: String,
    value: Int,
    tint: Color,
    background: Color,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .background(background, RoundedCornerShape(16.dp))
            .padding(horizontal = 10.dp, vertical = 11.dp),
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Text(label, fontSize = 10.sp, fontWeight = FontWeight.Black, color = tint)
        Text(value.toString(), fontSize = 18.sp, fontWeight = FontWeight.Black, color = tint)
    }
}

@Composable
fun RiskTrendEmptyCard() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(24.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.88f)),
        border = BorderStroke(1.dp, Color(0x75E2E8F0)),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 44.dp, horizontal = 16.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Box(
                modifier = Modifier
                    .size(56.dp)
                    .background(Color(0xFFF1F5F9), CircleShape),
                contentAlignment = Alignment.Center,
            ) {
                Icon(Icons.Outlined.Add, contentDescription = null, tint = Color(0xFFCBD5E1), modifier = Modifier.size(28.dp))
            }
            Text("近期暂无检测数据记录", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = Color(0xFF94A3B8))
        }
    }
}

@Composable
fun AlertListCard(item: AlertEvent, onOpen: () -> Unit) {
    val accent = when (normalizeRiskLevel(item.risk_level)) {
        "高" -> Color(0xFFEF4444)
        "低" -> MobileGreen
        else -> Color(0xFFF59E0B)
    }
    Card(modifier = Modifier.fillMaxWidth().clickable(onClick = onOpen), shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Box(modifier = Modifier.fillMaxWidth()) {
            Box(
                modifier = Modifier
                    .align(Alignment.TopStart)
                    .width(6.dp)
                    .fillMaxHeight()
                    .background(accent, RoundedCornerShape(topStart = 24.dp, bottomStart = 24.dp)),
            )
            Column(modifier = Modifier.padding(start = 18.dp, end = 16.dp, top = 16.dp, bottom = 16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    RiskChip(item.risk_level)
                    Spacer(Modifier.weight(1f))
                    Text(formatDateTime(item.created_at.ifBlank { item.sent_at }), fontSize = 11.sp, color = Color(0xFF94A3B8))
                }
                Text(item.title.ifBlank { "风险预警" }, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 15.sp)
                Text(item.case_summary.ifBlank { "请及时核查相关风险事件。" }, fontSize = 12.sp, color = MobileSubtle, maxLines = 2, overflow = TextOverflow.Ellipsis)
            }
        }
    }
}

@Composable
fun UploadActionTile(label: String, count: Int, icon: ImageVector, modifier: Modifier, onClick: () -> Unit) {
    Card(
        modifier = modifier.clickable(onClick = onClick),
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
    ) {
        Column(
            modifier = Modifier
                .border(2.dp, Color(0xFFCACACA), RoundedCornerShape(20.dp))
                .padding(vertical = 16.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            Box(modifier = Modifier.size(36.dp).background(Color(0xFFF8FAFC), CircleShape), contentAlignment = Alignment.Center) {
                Icon(icon, contentDescription = null, tint = MobileSubtle, modifier = Modifier.size(18.dp))
            }
            Text(label, fontWeight = FontWeight.Bold, fontSize = 12.sp, color = MobileText)
            if (count > 0) Text("$count", fontSize = 11.sp, color = MobileGreen, fontWeight = FontWeight.Black)
        }
    }
}

@Composable
fun DifficultyChip(label: String, value: String, selected: String, modifier: Modifier = Modifier, onSelected: (String) -> Unit) {
    Button(
        onClick = { onSelected(value) },
        modifier = modifier.height(44.dp),
        shape = RoundedCornerShape(14.dp),
        colors = ButtonDefaults.buttonColors(
            containerColor = if (value == selected) Color.White else Color(0xFFF8FAFC),
            contentColor = when {
                value != selected -> Color(0xFF64748B)
                value == "hard" -> Color(0xFFBE123C)
                value == "medium" -> Color(0xFFB45309)
                else -> Color(0xFF047857)
            },
        ),
    ) {
        Text(label, fontWeight = FontWeight.Bold)
    }
}

@Composable
fun SimulationPackCard(pack: SimulationPack, onStart: () -> Unit) {
    Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                AssistChip(onClick = {}, enabled = false, label = { Text(pack.case_type.ifBlank { "演练场景" }) })
                AssistChip(onClick = {}, enabled = false, label = { Text(difficultyLabel(pack.difficulty)) })
            }
            Text(pack.title.ifBlank { "未命名题包" }, fontWeight = FontWeight.ExtraBold, color = MobileText)
            Text(pack.intro.ifBlank { "进入演练后将逐步完成实景答题。" }, fontSize = 12.sp, color = MobileSubtle)
            Button(onClick = onStart, modifier = Modifier.fillMaxWidth().height(46.dp), shape = RoundedCornerShape(16.dp), colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFECFDF5), contentColor = MobileGreen)) {
                Text("开始挑战", fontWeight = FontWeight.Bold)
            }
        }
    }
}

@Composable
fun SimulationSessionCard(session: SimulationSessionItem, deleting: Boolean, onResume: () -> Unit, onDelete: () -> Unit) {
    val scoreColor = when {
        session.score >= 80 -> MobileGreen
        session.score >= 60 -> Color(0xFFF59E0B)
        else -> MobileRose
    }
    Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(modifier = Modifier.size(52.dp).background(scoreColor.copy(alpha = 0.12f), CircleShape), contentAlignment = Alignment.Center) {
                Text("${session.score}", color = scoreColor, fontWeight = FontWeight.ExtraBold)
            }
            Column(modifier = Modifier.weight(1f).padding(start = 12.dp)) {
                Text(session.title.ifBlank { "未命名演练" }, fontWeight = FontWeight.ExtraBold, color = MobileText)
                Text("${session.level.ifBlank { "未评级" }} · ${if (session.status == "completed") "已完成" else "未完成"}", fontSize = 12.sp, color = MobileSubtle)
            }
            if (session.status == "completed") {
                IconButton(onClick = onDelete, enabled = !deleting) { Icon(Icons.Outlined.DeleteOutline, contentDescription = "删除") }
            } else {
                IconButton(onClick = onResume) { Icon(Icons.Outlined.KeyboardArrowRight, contentDescription = "继续") }
            }
        }
    }
}

@Composable
fun IntroExamCard(pack: SimulationPack) {
    Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Column(modifier = Modifier.padding(18.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            AssistChip(onClick = {}, enabled = false, label = { Text("任务目标") })
            Text(pack.title.ifBlank { "模拟演练" }, fontWeight = FontWeight.ExtraBold, fontSize = 22.sp, color = MobileText)
            Text(pack.intro.ifBlank { "请按照提示完成场景问答。" }, color = MobileSubtle, lineHeight = 20.sp)
        }
    }
}

@Composable
fun ScenarioCard(step: SimulationStep) {
    Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Column(modifier = Modifier.padding(18.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("场景提示 · ${step.step_type}", fontWeight = FontWeight.Bold, color = Color(0xFF94A3B8), fontSize = 11.sp)
            Text(step.narrative, color = MobileSubtle, lineHeight = 20.sp)
            Text(step.question, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 20.sp)
        }
    }
}

@Composable
fun AnswerOptionCard(option: SimulationOption, enabled: Boolean, selected: Boolean = false, onClick: () -> Unit) {
    val backgroundColor by animateColorAsState(if (selected) Color(0xFFF0FDF4) else Color.White, label = "answerOptionBg")
    val borderColor by animateColorAsState(if (selected) Color(0xFF86EFAC) else MobileBorder, label = "answerOptionBorder")
    val badgeBackground by animateColorAsState(if (selected) MobileGreen else Color(0xFFF8FAFC), label = "answerOptionBadgeBg")
    val badgeTextColor by animateColorAsState(if (selected) Color.White else MobileSubtle, label = "answerOptionBadgeText")
    val textColor by animateColorAsState(if (selected) Color(0xFF0F172A) else MobileText, label = "answerOptionText")
    val cardScale by animateFloatAsState(if (selected) 0.985f else 1f, label = "answerOptionScale")
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .scale(cardScale)
            .clickable(enabled = enabled, onClick = onClick),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = backgroundColor),
    ) {
        Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(modifier = Modifier.size(40.dp).background(badgeBackground, CircleShape), contentAlignment = Alignment.Center) {
                Text(option.key, fontWeight = FontWeight.ExtraBold, color = badgeTextColor)
            }
            Text(option.text, modifier = Modifier.weight(1f).padding(start = 14.dp), color = textColor, fontWeight = FontWeight.Bold)
        }
    }
}

@Composable
fun SimulationResultCard(result: SimulationResult, onBack: () -> Unit) {
    val adviceItems = result.advice.orEmpty()
    Column(verticalArrangement = Arrangement.spacedBy(16.dp), horizontalAlignment = Alignment.CenterHorizontally, modifier = Modifier.fillMaxWidth()) {
        Box(modifier = Modifier.size(96.dp).background(Color(0xFFD1FAE5), CircleShape), contentAlignment = Alignment.Center) {
            Icon(Icons.Outlined.Shield, contentDescription = null, tint = MobileGreen, modifier = Modifier.size(42.dp))
        }
        Text("完成挑战", fontWeight = FontWeight.ExtraBold, fontSize = 28.sp, color = MobileText)
        Text("评级：${result.level} · 得分：${result.total_score}", color = MobileGreen, fontWeight = FontWeight.Bold)
        Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                Text("防诈建议", fontWeight = FontWeight.ExtraBold, color = MobileText)
                if (adviceItems.isEmpty()) {
                    Text("当前暂无补充建议。", color = MobileSubtle)
                } else {
                    adviceItems.forEachIndexed { index, advice -> Text("${index + 1}. $advice", color = MobileSubtle) }
                }
            }
        }
        Button(onClick = onBack, modifier = Modifier.fillMaxWidth().height(54.dp), shape = RoundedCornerShape(18.dp), colors = ButtonDefaults.buttonColors(containerColor = MobileText)) {
            Text("返回题包列表", fontWeight = FontWeight.Bold)
        }
    }
}

@Composable
fun ChatWelcomeCard() {
    Column(modifier = Modifier.fillMaxWidth().padding(vertical = 40.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Box(modifier = Modifier.size(72.dp).background(Brush.linearGradient(listOf(MobileGreen, MobileTeal)), CircleShape), contentAlignment = Alignment.Center) {
            Icon(Icons.Outlined.Shield, contentDescription = null, tint = Color.White, modifier = Modifier.size(32.dp))
        }
        Text("我是您的反诈助手", fontWeight = FontWeight.ExtraBold, fontSize = 20.sp, color = MobileText)
        Text("可以帮您识别诈骗信息、分析风险案例或提供安全建议。", modifier = Modifier.width(260.dp), textAlign = TextAlign.Center, color = MobileSubtle)
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun MobileChatBubble(message: DisplayChatMessage) {
    when (message.type) {
        "tool" -> {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Center) {
                Text(
                    text = "* ${message.content}",
                    color = Color(0xFF94A3B8),
                    fontSize = 11.sp,
                )
            }
        }
        "error" -> {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Start) {
                Box(
                    modifier = Modifier
                        .widthIn(max = 320.dp)
                        .background(Color(0xFFFEF2F2), RoundedCornerShape(16.dp))
                        .border(1.dp, Color(0xFFFECACA), RoundedCornerShape(16.dp))
                        .padding(horizontal = 16.dp, vertical = 12.dp),
                ) {
                    Text(message.content, color = Color(0xFFB91C1C), fontSize = 14.sp)
                }
            }
        }
        "user" -> {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                Column(
                    modifier = Modifier.widthIn(max = 320.dp),
                    horizontalAlignment = Alignment.End,
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    if (message.content.isNotBlank()) {
                        Text(
                            message.content,
                            color = Color(0xFF334155),
                            fontSize = 15.sp,
                            lineHeight = 26.sp,
                            textAlign = TextAlign.End,
                        )
                    }
                    if (message.images.isNotEmpty()) {
                        FlowRow(
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            message.images.forEach { image ->
                                Box(
                                    modifier = Modifier
                                        .size(width = 132.dp, height = 112.dp)
                                        .clip(RoundedCornerShape(16.dp))
                                        .border(1.dp, MobileBorder, RoundedCornerShape(16.dp)),
                                ) {
                                    Base64Thumbnail(image, 112.dp, Modifier.fillMaxWidth().height(112.dp))
                                }
                            }
                        }
                    }
                }
            }
        }
        else -> {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Start) {
                Column(
                    modifier = Modifier.widthIn(max = 320.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    if (message.content.isNotBlank()) {
                        MarkdownText(message.content, MobileText)
                    }
                    if (message.images.isNotEmpty()) {
                        FlowRow(
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            message.images.forEach { image ->
                                Box(
                                    modifier = Modifier
                                        .size(width = 132.dp, height = 112.dp)
                                        .clip(RoundedCornerShape(16.dp))
                                        .border(1.dp, MobileBorder, RoundedCornerShape(16.dp)),
                                ) {
                                    Base64Thumbnail(image, 112.dp, Modifier.fillMaxWidth().height(112.dp))
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
fun FamilyInvitationCard(invitation: FamilyInvitation, accepting: Boolean, onAccept: () -> Unit) {
    val inviterText = invitation.inviter_name.ifBlank {
        invitation.inviter_email.ifBlank { invitation.inviter_phone.ifBlank { "未知" } }
    }
    val pending = invitation.status == "pending"
    Card(
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = BorderStroke(1.dp, Color(0x14CBD5E1)),
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Row(verticalAlignment = Alignment.Top, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Text(invitation.family_name.ifBlank { "家庭邀请" }, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 14.sp)
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        LucideSvgIcon(iconName = "users", contentDescription = null, tint = Color(0xFF64748B), size = 12.dp)
                        Text(inviterText, fontSize = 11.sp, color = MobileSubtle)
                    }
                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        FamilyTinyTag("角色: ${invitation.role.ifBlank { "member" }}")
                        FamilyTinyTag("关系: ${invitation.relation.ifBlank { "未填写" }}")
                    }
                }
                Box(
                    modifier = Modifier
                        .background(
                            if (pending) Color(0xFF0F172A) else Color(0x80E5E7EB),
                            RoundedCornerShape(8.dp),
                        )
                        .padding(horizontal = 8.dp, vertical = 4.dp),
                ) {
                    Text(
                        if (pending) "待处理" else invitation.status.ifBlank { "已处理" },
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Bold,
                        color = if (pending) Color.White else Color(0xFF64748B),
                    )
                }
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFFF8FAFC), RoundedCornerShape(12.dp))
                    .border(1.dp, Color(0x14CBD5E1), RoundedCornerShape(12.dp))
                    .padding(horizontal = 12.dp, vertical = 12.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text("邀请码", fontSize = 10.sp, color = Color(0xFF94A3B8), fontWeight = FontWeight.Black)
                Text(invitation.invite_code, fontSize = 13.sp, fontWeight = FontWeight.ExtraBold, color = Color(0xFF334155))
            }

            Button(
                onClick = onAccept,
                modifier = Modifier.fillMaxWidth().height(40.dp),
                shape = RoundedCornerShape(12.dp),
                enabled = pending && !accepting,
                colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF0F172A), contentColor = Color.White),
            ) {
                Text(if (accepting) "加入中..." else "接受邀请", fontWeight = FontWeight.Bold, fontSize = 13.sp)
            }
        }
    }
}

@Composable
fun FamilySummaryHero(state: MainUiState, onManage: () -> Unit) {
    val family = state.familyOverview.family ?: return
    val members = state.familyOverview.members
    val unreadCount = state.familyOverview.unread_notification_count.takeIf { it > 0 }
        ?: state.familyNotifications.count { it.read_at.isBlank() }
    Card(
        shape = RoundedCornerShape(24.dp),
        colors = CardDefaults.cardColors(containerColor = Color(0xFFF0FCFA)),
        border = BorderStroke(1.dp, Color(0xFFD9F5F0)),
    ) {
        Box {
            Box(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .size(96.dp)
                    .background(Color.White.copy(alpha = 0.4f), RoundedCornerShape(bottomStart = 96.dp)),
            )
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(20.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                Row(verticalAlignment = Alignment.Top, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                        Text(family.name, fontWeight = FontWeight.ExtraBold, fontSize = 20.sp, color = MobileText)
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            FamilyCountTag("成员 ${members.size}")
                            if (unreadCount > 0) {
                                Box(
                                    modifier = Modifier
                                        .background(Color(0xFFFFF1F2), RoundedCornerShape(8.dp))
                                        .border(1.dp, Color(0x1AF43F5E), RoundedCornerShape(8.dp))
                                        .padding(horizontal = 8.dp, vertical = 4.dp),
                                ) {
                                    Text("未读 $unreadCount", color = Color(0xFFEF4444), fontSize = 10.sp, fontWeight = FontWeight.Black)
                                }
                            }
                        }
                    }
                    Box(
                        modifier = Modifier
                            .size(40.dp)
                            .background(Color.White, CircleShape)
                            .border(1.dp, Color(0xFFD9F5F0), CircleShape),
                        contentAlignment = Alignment.Center,
                    ) {
                        LucideSvgIcon(iconName = "users", contentDescription = null, tint = Color(0xFF14B8A6), size = 16.dp)
                    }
                }

                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White, RoundedCornerShape(16.dp))
                        .border(1.dp, Color(0xFFD9F5F0), RoundedCornerShape(16.dp))
                        .padding(horizontal = 14.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Text("专属邀请码", color = Color(0xFF94A3B8), fontSize = 10.sp, fontWeight = FontWeight.Black)
                        Text(family.invite_code.ifBlank { "暂无邀请码" }, color = Color(0xFF334155), fontWeight = FontWeight.ExtraBold, fontSize = 14.sp)
                    }
                    Button(
                        onClick = onManage,
                        shape = RoundedCornerShape(10.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFE8FAF6), contentColor = Color(0xFF0F766E)),
                        contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 12.dp, vertical = 0.dp),
                        modifier = Modifier.height(32.dp),
                    ) {
                        Text("管理", fontWeight = FontWeight.Bold, fontSize = 11.sp)
                    }
                }

                Row {
                    members.take(6).forEachIndexed { index, member ->
                        Box(
                            modifier = Modifier
                                .offset(x = ((-6) * index).dp)
                                .size(32.dp)
                                .background(Color.White, CircleShape)
                                .border(2.dp, Color(0xFFF0FCFA), CircleShape),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(member.username.take(1).ifBlank { "U" }, fontSize = 11.sp, fontWeight = FontWeight.ExtraBold, color = Color(0xFF475569))
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun FamilyMemberCard(member: FamilyMember) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, RoundedCornerShape(16.dp))
            .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(16.dp))
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Box(
            modifier = Modifier
                .size(36.dp)
                .background(Color(0xFFF0FCFA), CircleShape)
                .border(1.dp, Color(0xFFE0F8F5), CircleShape),
            contentAlignment = Alignment.Center,
        ) {
            Text(member.username.take(1).ifBlank { "U" }, fontWeight = FontWeight.ExtraBold, color = Color(0xFF14B8A6), fontSize = 12.sp)
        }
        Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                Text(member.username.ifBlank { "未命名成员" }, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 13.sp)
                FamilyTinyRole(member.role.ifBlank { "member" })
            }
            Text(
                "${member.relation.ifBlank { "未设置关系" }} | ${member.email.ifBlank { member.phone.ifBlank { "无联系方式" } }}",
                fontSize = 10.sp,
                color = Color(0xFF94A3B8),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
fun GuardianRelationCard(link: GuardianLink) {
    Card(
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = BorderStroke(1.dp, Color(0x14CBD5E1)),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Text(link.guardian_name.ifBlank { "守护人" }, fontWeight = FontWeight.ExtraBold, color = Color(0xFF334155), fontSize = 13.sp)
                    LucideSvgIcon(iconName = "arrow-left", contentDescription = null, tint = Color(0xFFCBD5E1), size = 14.dp, modifier = Modifier.rotate(180f))
                    Text(link.member_name.ifBlank { "被守护人" }, fontWeight = FontWeight.ExtraBold, color = Color(0xFF334155), fontSize = 13.sp)
                }
                Text(
                    "${link.guardian_email.ifBlank { "-" }} / ${link.member_email.ifBlank { "-" }}",
                    fontSize = 10.sp,
                    color = Color(0xFF94A3B8),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
            Box(
                modifier = Modifier
                    .background(Color(0xFFF8FAFC), RoundedCornerShape(8.dp))
                    .border(1.dp, Color(0x14CBD5E1), RoundedCornerShape(8.dp))
                    .padding(horizontal = 8.dp, vertical = 6.dp),
            ) {
                Text("守护中", fontSize = 9.sp, fontWeight = FontWeight.Bold, color = Color(0xFF64748B))
            }
        }
    }
}

@Composable
fun FamilyNotificationCard(note: FamilyNotification) {
    Card(
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = BorderStroke(1.dp, Color(0x14CBD5E1)),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 14.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Box(
                modifier = Modifier
                    .size(36.dp)
                    .background(Color(0xFFF8FAFC), CircleShape)
                    .border(1.dp, Color(0x12CBD5E1), CircleShape),
                contentAlignment = Alignment.Center,
            ) {
                LucideSvgIcon(iconName = "alert-octagon", contentDescription = null, tint = Color(0xFF64748B), size = 16.dp)
            }
            Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        note.title.ifBlank { "事件预警" },
                        fontSize = 13.sp,
                        fontWeight = FontWeight.ExtraBold,
                        color = MobileText,
                        modifier = Modifier.weight(1f),
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Text(formatDateTime(note.event_at), fontSize = 10.sp, color = Color(0xFF94A3B8))
                }
                Text(
                    note.case_summary.ifBlank { note.summary.ifBlank { "无最新动态" } },
                    fontSize = 11.sp,
                    color = Color(0xFF64748B),
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    lineHeight = 16.sp,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                    FamilyTinyTag(note.target_name.ifBlank { "家庭成员" })
                    if (note.scam_type.isNotBlank()) {
                        FamilyTinyTag(note.scam_type)
                    }
                    Box(
                        modifier = Modifier
                            .background(Color(0xFFFFF1F2), RoundedCornerShape(6.dp))
                            .border(1.dp, Color(0x12FB7185), RoundedCornerShape(6.dp))
                            .padding(horizontal = 6.dp, vertical = 3.dp),
                    ) {
                        Text(note.risk_level.ifBlank { "高危" }, fontSize = 9.sp, color = Color(0xFFEF4444), fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}

@Composable
fun FamilyManageMemberCard(member: FamilyMember, deleting: Boolean, canDelete: Boolean, onDelete: () -> Unit) {
    Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Row(modifier = Modifier.padding(14.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(modifier = Modifier.size(40.dp).background(Color(0xFFEEF2FF), CircleShape), contentAlignment = Alignment.Center) {
                Text(member.username.take(1).ifBlank { "M" }, fontWeight = FontWeight.ExtraBold, color = Color(0xFF4338CA))
            }
            Column(modifier = Modifier.weight(1f).padding(start = 12.dp)) {
                Text(member.username, fontWeight = FontWeight.ExtraBold, color = MobileText)
                Text("${member.relation.ifBlank { member.role }} · ${member.email.ifBlank { member.phone.ifBlank { "-" } }}", fontSize = 12.sp, color = MobileSubtle)
            }
            if (canDelete) {
                Button(onClick = onDelete, enabled = !deleting, shape = RoundedCornerShape(12.dp), colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFFEE2E2), contentColor = MobileRose)) {
                    Text(if (deleting) "移除中" else "移除")
                }
            } else {
                AssistChip(onClick = {}, enabled = false, label = { Text(member.role) })
            }
        }
    }
}

@Composable
fun FamilyManageInvitationCard(invitation: FamilyInvitation) {
    Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(invitation.invitee_email.ifBlank { invitation.invitee_phone.ifBlank { "未指定对象" } }, fontWeight = FontWeight.ExtraBold, color = MobileText)
            Text("角色：${invitation.role} / 关系：${invitation.relation.ifBlank { "未填写" }}", fontSize = 12.sp, color = MobileSubtle)
            Text("邀请码：${invitation.invite_code}", fontSize = 12.sp, color = MobileGreen, fontWeight = FontWeight.Bold)
            Text("状态：${invitation.status} / 截止：${formatDateTime(invitation.expires_at)}", fontSize = 11.sp, color = Color(0xFF94A3B8))
        }
    }
}

@Composable
fun NavigationCell(title: String, subtitle: String, icon: ImageVector, tint: Color, onClick: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth().clickable(onClick = onClick),
        shape = RoundedCornerShape(28.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = BorderStroke(1.dp, Color(0x12CBD5E1)),
    ) {
        Row(modifier = Modifier.padding(20.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .background(Brush.linearGradient(listOf(tint.copy(alpha = 0.10f), tint.copy(alpha = 0.18f))), RoundedCornerShape(18.dp)),
                contentAlignment = Alignment.Center,
            ) {
                Icon(icon, contentDescription = null, tint = tint)
            }
            Column(modifier = Modifier.weight(1f).padding(start = 14.dp)) {
                Text(title, fontWeight = FontWeight.ExtraBold, color = MobileText, fontSize = 16.sp)
                Text(subtitle, fontSize = 13.sp, color = MobileSubtle, lineHeight = 18.sp)
            }
            Box(
                modifier = Modifier
                    .size(32.dp)
                    .background(Color(0xFFF8FAFC), CircleShape),
                contentAlignment = Alignment.Center,
            ) {
                Icon(Icons.Outlined.KeyboardArrowRight, contentDescription = null, tint = Color(0xFF94A3B8), modifier = Modifier.size(16.dp))
            }
        }
    }
}

@Composable
private fun FamilyTinyTag(text: String) {
    Box(
        modifier = Modifier
            .background(Color(0xFFF8FAFC), RoundedCornerShape(6.dp))
            .border(1.dp, Color(0x12CBD5E1), RoundedCornerShape(6.dp))
            .padding(horizontal = 6.dp, vertical = 3.dp),
    ) {
        Text(text, fontSize = 9.sp, color = Color(0xFF64748B), fontWeight = FontWeight.Bold)
    }
}

@Composable
private fun FamilyTinyRole(text: String) {
    Box(
        modifier = Modifier
            .background(Color(0xFFF0FCFA), RoundedCornerShape(6.dp))
            .border(1.dp, Color(0xFFE0F8F5), RoundedCornerShape(6.dp))
            .padding(horizontal = 6.dp, vertical = 3.dp),
    ) {
        Text(text, fontSize = 9.sp, color = Color(0xFF14B8A6), fontWeight = FontWeight.Bold)
    }
}

@Composable
private fun FamilyCountTag(text: String) {
    Box(
        modifier = Modifier
            .background(Color.White.copy(alpha = 0.92f), RoundedCornerShape(8.dp))
            .border(1.dp, Color(0xFFD9F5F0), RoundedCornerShape(8.dp))
            .padding(horizontal = 8.dp, vertical = 4.dp),
    ) {
        Text(text, fontSize = 10.sp, fontWeight = FontWeight.Bold, color = Color(0xFF64748B))
    }
}

@Composable
fun InfoRow(label: String, value: String) {
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(label, fontSize = 11.sp, color = Color(0xFF94A3B8), fontWeight = FontWeight.Bold)
        Text(value, color = MobileText, fontWeight = FontWeight.Bold)
    }
}

fun difficultyLabel(value: String): String = when (value) {
    "hard" -> "困难"
    "medium" -> "中等"
    else -> "简单"
}

@Composable
fun SelectionDropdownField(
    label: String,
    valueLabel: String,
    placeholder: String,
    options: List<DropdownOption>,
    selectedValue: String = "",
    hint: String = "",
    accentColor: Color = MobileGreen,
    enabled: Boolean = true,
    modifier: Modifier = Modifier,
    onSelect: (String) -> Unit,
) {
    var expanded by remember { mutableStateOf(false) }
    val arrowRotation by animateFloatAsState(targetValue = if (expanded) 90f else 0f, label = "dropdownArrow")

    Column(modifier = modifier, verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(label, color = MobileSubtle, fontSize = 12.sp, fontWeight = FontWeight.Bold)
        Box {
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable(enabled = enabled) { expanded = !expanded },
                shape = RoundedCornerShape(18.dp),
                colors = CardDefaults.cardColors(
                    containerColor = if (enabled) {
                        if (expanded) Color.White else Color(0xFFF8FAFC)
                    } else {
                        Color(0xFFF8FAFC)
                    },
                ),
                border = BorderStroke(
                    1.dp,
                    if (expanded) accentColor.copy(alpha = 0.35f) else MobileBorder,
                ),
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 14.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Text(
                            text = if (valueLabel.isNotBlank()) valueLabel else placeholder,
                            textAlign = TextAlign.Start,
                            color = if (valueLabel.isNotBlank()) MobileText else MobileSubtle,
                            fontWeight = FontWeight.ExtraBold,
                            fontSize = 14.sp,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                        )
                        if (hint.isNotBlank()) {
                            Text(
                                text = hint,
                                color = if (expanded) accentColor.copy(alpha = 0.85f) else MobileSubtle.copy(alpha = 0.8f),
                                fontSize = 11.sp,
                                maxLines = 2,
                                overflow = TextOverflow.Ellipsis,
                            )
                        }
                    }
                    Icon(
                        Icons.Outlined.KeyboardArrowRight,
                        contentDescription = null,
                        tint = if (expanded) accentColor else MobileSubtle,
                        modifier = Modifier.rotate(arrowRotation),
                    )
                }
            }
            DropdownMenu(
                expanded = expanded,
                onDismissRequest = { expanded = false },
                modifier = Modifier
                    .widthIn(min = 260.dp, max = 340.dp)
                    .heightIn(max = 320.dp),
            ) {
                options.forEach { option ->
                    val selected = option.value == selectedValue
                    DropdownMenuItem(
                        text = {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(
                                        if (selected) accentColor.copy(alpha = 0.12f) else Color.Transparent,
                                        RoundedCornerShape(14.dp),
                                    )
                                    .padding(horizontal = 12.dp, vertical = 10.dp),
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                    Text(
                                        option.label,
                                        fontWeight = FontWeight.Bold,
                                        color = if (selected) accentColor else MobileText,
                                        maxLines = 1,
                                        overflow = TextOverflow.Ellipsis,
                                    )
                                    if (option.hint.isNotBlank()) {
                                        Text(
                                            option.hint,
                                            fontSize = 11.sp,
                                            color = if (selected) accentColor.copy(alpha = 0.8f) else MobileSubtle,
                                            maxLines = 2,
                                            overflow = TextOverflow.Ellipsis,
                                        )
                                    }
                                }
                                if (selected) {
                                    Icon(Icons.Outlined.Check, contentDescription = null, tint = accentColor)
                                }
                            }
                        },
                        onClick = {
                            expanded = false
                            onSelect(option.value)
                        },
                    )
                }
            }
        }
    }
}
