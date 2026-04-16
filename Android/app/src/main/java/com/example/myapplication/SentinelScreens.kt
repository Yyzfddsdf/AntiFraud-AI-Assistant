package com.example.myapplication

import android.content.ContentValues
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Canvas as AndroidCanvas
import android.os.Environment
import android.provider.MediaStore
import com.caverock.androidsvg.SVG
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Add
import androidx.compose.material.icons.outlined.ArrowBack
import androidx.compose.material.icons.outlined.ChatBubbleOutline
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material.icons.outlined.DeleteOutline
import androidx.compose.material.icons.outlined.Email
import androidx.compose.material.icons.outlined.Image
import androidx.compose.material.icons.outlined.Layers
import androidx.compose.material.icons.outlined.Mic
import androidx.compose.material.icons.outlined.NotificationsNone
import androidx.compose.material.icons.outlined.PersonOutline
import androidx.compose.material.icons.outlined.Phone
import androidx.compose.material.icons.outlined.Send
import androidx.compose.material.icons.outlined.Security
import androidx.compose.material.icons.outlined.Shield
import androidx.compose.material.icons.outlined.ShowChart
import androidx.compose.material.icons.outlined.WarningAmber
import androidx.compose.material.icons.outlined.Description
import androidx.compose.material.icons.outlined.Groups
import androidx.compose.material.icons.outlined.Videocam
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Switch
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex
import java.net.URLDecoder
import java.nio.charset.StandardCharsets
import java.text.SimpleDateFormat
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.Date
import java.util.Locale
import java.util.Base64
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject

@Composable
fun AuthScreen(
    state: MainUiState,
    viewModel: SentinelViewModel,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color(0xFFF8FAFC)),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(320.dp)
                .background(
                    Brush.verticalGradient(
                        listOf(Color(0x3334D399), Color(0x1014B8A6), Color.Transparent),
                    ),
                ),
        )
        Box(
            modifier = Modifier
                .size(320.dp)
                .align(Alignment.TopEnd)
                .offset(x = 110.dp, y = (-130).dp)
                .background(Color(0x140EA5E9), CircleShape),
        )
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 24.dp, vertical = 24.dp),
            verticalArrangement = Arrangement.Center,
        ) {
            Spacer(Modifier.height(18.dp))
            Column(
                modifier = Modifier.fillMaxWidth(),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Box(
                    modifier = Modifier
                        .size(96.dp)
                        .background(Color(0x22059669), RoundedCornerShape(20.dp))
                        .padding(10.dp),
                ) {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .offset(x = 4.dp, y = 4.dp)
                            .background(Color(0x22059669), RoundedCornerShape(28.dp)),
                    )
                    BrandShieldIconTile(modifier = Modifier.fillMaxSize())
                }
                Spacer(Modifier.height(24.dp))
                Text("反诈卫士", style = MaterialTheme.typography.headlineLarge, color = Color(0xFF0F172A), fontWeight = FontWeight.ExtraBold)
                Spacer(Modifier.height(6.dp))
                Text("守护您和家人的财产安全", color = Color(0xFF64748B), fontSize = 15.sp, fontWeight = FontWeight.Medium)
            }
            Spacer(Modifier.height(28.dp))
            Card(
                shape = RoundedCornerShape(32.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.84f)),
                modifier = Modifier.fillMaxWidth(),
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.White.copy(alpha = 0.84f))
                        .padding(horizontal = 20.dp, vertical = 22.dp),
                    verticalArrangement = Arrangement.spacedBy(14.dp),
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 2.dp),
                        horizontalArrangement = Arrangement.Center,
                    ) {
                        AuthModeTab(
                            text = "登录",
                            selected = state.authMode == AuthMode.Login,
                            onClick = { viewModel.updateAuthMode(AuthMode.Login) },
                        )
                        Spacer(Modifier.width(28.dp))
                        AuthModeTab(
                            text = "注册",
                            selected = state.authMode == AuthMode.Register,
                            onClick = { viewModel.updateAuthMode(AuthMode.Register) },
                        )
                    }
                    HorizontalDivider(color = Color(0xFFF1F5F9), thickness = 1.dp)

                    if (state.authMode == AuthMode.Login) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(Color(0xFFF1F5F9), RoundedCornerShape(18.dp))
                                .padding(4.dp),
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            LoginMethodPill(
                                text = "密码登录",
                                selected = state.loginMethod == LoginMethod.Password,
                                onClick = { viewModel.updateLoginMethod(LoginMethod.Password) },
                                modifier = Modifier.weight(1f),
                            )
                            LoginMethodPill(
                                text = "验证码登录",
                                selected = state.loginMethod == LoginMethod.Sms,
                                onClick = { viewModel.updateLoginMethod(LoginMethod.Sms) },
                                modifier = Modifier.weight(1f),
                            )
                        }
                    }

                    if (state.authMode == AuthMode.Register) {
                        AuthUnderlineField(
                            label = "用户名",
                            value = state.authForm.username,
                            onValueChange = viewModel::updateAuthUsername,
                            placeholder = "设置用户名",
                        )
                        AuthUnderlineField(
                            label = "邮箱",
                            value = state.authForm.email,
                            onValueChange = viewModel::updateAuthEmail,
                            placeholder = "邮箱地址",
                            keyboardType = KeyboardType.Email,
                        )
                    }

                    if (state.authMode == AuthMode.Login && state.loginMethod == LoginMethod.Password) {
                        AuthUnderlineField(
                            label = "账号",
                            value = state.authForm.account,
                            onValueChange = viewModel::updateAuthAccount,
                            placeholder = "邮箱或手机号",
                        )
                    }

                    if (state.authMode == AuthMode.Register || state.loginMethod == LoginMethod.Sms) {
                        AuthUnderlineField(
                            label = "手机号",
                            value = state.authForm.phone,
                            onValueChange = viewModel::updateAuthPhone,
                            placeholder = "11位手机号",
                            keyboardType = KeyboardType.Phone,
                        )
                    }

                    if (state.authMode == AuthMode.Register || state.loginMethod == LoginMethod.Password) {
                        AuthUnderlineField(
                            label = "密码",
                            value = state.authForm.password,
                            onValueChange = viewModel::updateAuthPassword,
                            placeholder = "输入密码",
                        )
                    }

                    if (state.authMode == AuthMode.Register || state.loginMethod == LoginMethod.Sms) {
                        AuthInlineActionField(
                            label = "短信验证码",
                            value = state.authForm.smsCode,
                            onValueChange = viewModel::updateAuthSmsCode,
                            placeholder = "短信验证码",
                            action = {
                                Button(
                                    onClick = viewModel::sendSmsCode,
                                    enabled = state.smsCooldownSeconds == 0,
                                    shape = RoundedCornerShape(18.dp),
                                    colors = ButtonDefaults.buttonColors(
                                        containerColor = Color(0xFFECFDF5),
                                        contentColor = Color(0xFF047857),
                                    ),
                                    contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 18.dp, vertical = 16.dp),
                                ) {
                                    Text(if (state.smsCooldownSeconds > 0) "${state.smsCooldownSeconds}s" else "发送", fontSize = 13.sp, fontWeight = FontWeight.Black)
                                }
                            },
                        )
                    }

                    if (state.authMode == AuthMode.Register || state.loginMethod == LoginMethod.Password) {
                        AuthInlineActionField(
                            label = "图形验证码",
                            value = state.authForm.captchaCode,
                            onValueChange = viewModel::updateAuthCaptchaCode,
                            placeholder = "图形验证码",
                            action = {
                                Box(
                                    modifier = Modifier
                                        .width(112.dp)
                                        .height(56.dp)
                                        .clip(RoundedCornerShape(18.dp))
                                        .background(Color(0xFFF1F5F9)),
                                ) {
                                    Base64Thumbnail(
                                        dataUrl = state.captchaImage,
                                        size = 112.dp,
                                        modifier = Modifier.fillMaxSize(),
                                        bitmapWidth = 112.dp,
                                        bitmapHeight = 56.dp,
                                        framed = false,
                                        contentScale = ContentScale.FillBounds,
                                        onClick = viewModel::fetchCaptcha,
                                    )
                                }
                            },
                        )
                    }

                    Button(
                        onClick = viewModel::submitAuth,
                        enabled = !state.loading,
                        shape = RoundedCornerShape(20.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = Color.Transparent),
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(56.dp)
                            .background(
                                Brush.linearGradient(listOf(Color(0xFF10B981), Color(0xFF14B8A6))),
                                RoundedCornerShape(20.dp),
                            ),
                    ) {
                        if (state.loading) {
                            CircularProgressIndicator(modifier = Modifier.size(18.dp), strokeWidth = 2.dp, color = Color.White)
                        } else {
                            Text(
                                when {
                                    state.authMode == AuthMode.Register -> "创建账户"
                                    state.loginMethod == LoginMethod.Sms -> "短信登录"
                                    else -> "立即登录"
                                },
                                color = Color.White,
                                fontWeight = FontWeight.Black,
                                fontSize = 16.sp,
                            )
                        }
                    }

                    Text(
                        "登录即代表同意服务协议与隐私政策",
                        color = Color(0xFF94A3B8),
                        fontSize = 11.sp,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.fillMaxWidth(),
                    )
                }
            }
        }
    }
}

@Composable
private fun AuthModeTab(
    text: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    Column(
        modifier = Modifier.clickable(
            interactionSource = remember { MutableInteractionSource() },
            indication = null,
            onClick = onClick,
        ),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = text,
            color = if (selected) Color(0xFF059669) else Color(0xFF9CA3AF),
            fontSize = 17.sp,
            fontWeight = FontWeight.ExtraBold,
        )
        Spacer(Modifier.height(8.dp))
        Box(
            modifier = Modifier
                .height(4.dp)
                .width(24.dp)
                .clip(RoundedCornerShape(topStart = 999.dp, topEnd = 999.dp))
                .background(if (selected) Color(0xFF10B981) else Color.Transparent),
        )
    }
}

@Composable
private fun LoginMethodPill(
    text: String,
    selected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Button(
        onClick = onClick,
        modifier = modifier,
        shape = RoundedCornerShape(14.dp),
        colors = ButtonDefaults.buttonColors(
            containerColor = if (selected) Color.White else Color.Transparent,
            contentColor = if (selected) Color(0xFF047857) else Color(0xFF64748B),
        ),
        elevation = null,
        contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 18.dp, vertical = 12.dp),
    ) {
        Text(text, fontSize = 13.sp, fontWeight = FontWeight.Bold)
    }
}

@Composable
private fun AuthUnderlineField(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    keyboardType: KeyboardType = KeyboardType.Text,
    modifier: Modifier = Modifier,
) {
    Column(modifier = modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(label, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFF94A3B8))
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(18.dp))
                .background(Color(0xFFF8FAFC))
                .border(1.dp, Color(0xFFE2E8F0), RoundedCornerShape(18.dp))
                .padding(horizontal = 18.dp, vertical = 16.dp),
        ) {
            BasicTextField(
                value = value,
                onValueChange = onValueChange,
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                textStyle = MaterialTheme.typography.bodyLarge.copy(color = Color(0xFF0F172A), fontSize = 15.sp),
                cursorBrush = SolidColor(Color(0xFF059669)),
                keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
                decorationBox = { innerTextField ->
                    if (value.isEmpty()) {
                        Text(placeholder, color = Color(0xFF9CA3AF), fontSize = 15.sp)
                    }
                    innerTextField()
                },
            )
        }
    }
}

@Composable
private fun AuthInlineActionField(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    action: @Composable () -> Unit,
) {
    Column(Modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(label, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFF94A3B8))
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .weight(1f)
                    .clip(RoundedCornerShape(18.dp))
                    .background(Color(0xFFF8FAFC))
                    .border(1.dp, Color(0xFFE2E8F0), RoundedCornerShape(18.dp))
                    .padding(horizontal = 18.dp, vertical = 16.dp),
            ) {
                BasicTextField(
                    value = value,
                    onValueChange = onValueChange,
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                    textStyle = MaterialTheme.typography.bodyLarge.copy(color = Color(0xFF0F172A), fontSize = 15.sp),
                    cursorBrush = SolidColor(Color(0xFF059669)),
                    decorationBox = { innerTextField ->
                        if (value.isEmpty()) {
                            Text(placeholder, color = Color(0xFF9CA3AF), fontSize = 15.sp)
                        }
                        innerTextField()
                    },
                )
            }
            action()
        }
    }
}

@Composable
fun DashboardScreen(
    state: MainUiState,
    padding: PaddingValues,
    onScreenChange: (AppScreen) -> Unit,
    onOpenTask: (String) -> Unit,
) {
    val banners = listOf(R.drawable.banner_background, R.drawable.banner_alt)
    var bannerIndex by remember { mutableIntStateOf(0) }
    val todayDetectionCount = remember(state.history) { todayHistoryCount(state.history) }
    val todayHighRiskCount = remember(state.history) { todayHighRiskCount(state.history) }

    LaunchedEffect(Unit) {
        while (true) {
            delay(5_000)
            bannerIndex = (bannerIndex + 1) % banners.size
        }
    }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(
            start = 0.dp,
            end = 0.dp,
            top = padding.calculateTopPadding() + 8.dp,
            bottom = padding.calculateBottomPadding() + 24.dp,
        ),
        verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        item {
            Box(modifier = Modifier.padding(horizontal = 16.dp)) {
                BannerCarousel(
                    banners = banners,
                    currentIndex = bannerIndex,
                    onSelect = { bannerIndex = it },
                )
            }
        }
        item { DashboardQuickActions(onScreenChange) }
        item {
            Row(
                modifier = Modifier.padding(horizontal = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                StatCard("今日检测", todayDetectionCount.toString(), Color(0xFF059669), Modifier.weight(1f))
                StatCard("高风险", todayHighRiskCount.toString(), Color(0xFFF97316), Modifier.weight(1f))
            }
        }
        item {
            RiskInsightCard(
                data = state.riskData,
                onOpenDetail = { onScreenChange(AppScreen.RiskTrend) },
            )
        }
        item {
            SectionHeader(title = "最近任务", action = "查看全部", onAction = { onScreenChange(AppScreen.History) }, modifier = Modifier.padding(horizontal = 16.dp))
        }
        if (state.tasks.isEmpty()) {
            item { EmptyState("暂无任务", modifier = Modifier.padding(horizontal = 16.dp)) }
        } else {
            items(state.tasks.take(3), key = { it.task_id }) { task ->
                TaskCard(task = task, onClick = { onOpenTask(task.task_id) }, modifier = Modifier.padding(horizontal = 16.dp))
            }
        }
    }
}

@Composable
fun HistoryScreen(
    state: MainUiState,
    padding: PaddingValues,
    onBack: () -> Unit,
    onOpenTask: (String) -> Unit,
    onDelete: (String) -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(top = padding.calculateTopPadding())
            .padding(horizontal = 12.dp, vertical = 12.dp),
    ) {
        ScreenBackHeader(title = "历史档案", subtitle = "${state.history.size} 条", onBack = onBack)
        Card(
            shape = RoundedCornerShape(18.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White),
            modifier = Modifier
                .padding(top = 12.dp)
                .padding(horizontal = 4.dp),
        ) {
            LazyColumn(
                contentPadding = PaddingValues(vertical = 4.dp),
                verticalArrangement = Arrangement.spacedBy(0.dp),
            ) {
                if (state.history.isEmpty()) {
                    item { EmptyState("暂无历史记录", modifier = Modifier.padding(12.dp)) }
                } else {
                    items(state.history, key = { it.record_id }) { item ->
                        HistoryCard(
                            item = item,
                            deleting = state.deletingHistoryIds.contains(item.record_id),
                            onOpen = { onOpenTask(item.record_id) },
                            onDelete = { onDelete(item.record_id) },
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun RiskTrendScreen(
    state: MainUiState,
    padding: PaddingValues,
    onBack: () -> Unit,
    onIntervalChange: (String) -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(top = padding.calculateTopPadding())
            .verticalScroll(rememberScrollState())
            .padding(
                start = 12.dp,
                end = 12.dp,
                top = 12.dp,
                bottom = padding.calculateBottomPadding() + 28.dp,
            ),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 4.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Card(
                shape = CircleShape,
                colors = CardDefaults.cardColors(containerColor = Color.White),
                modifier = Modifier.size(40.dp),
                onClick = onBack,
            ) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Icon(Icons.Outlined.ArrowBack, contentDescription = "返回", tint = Color(0xFF475569), modifier = Modifier.size(18.dp))
                }
            }
            Text("风险详情", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.ExtraBold, color = Color(0xFF0F172A))
        }
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp), modifier = Modifier.padding(horizontal = 4.dp)) {
            StatCard("总检测", (state.riskData?.stats?.total ?: 0).toString(), Color(0xFF059669), Modifier.weight(1f))
            StatCard("高风险", (state.riskData?.stats?.high ?: 0).toString(), Color(0xFFF97316), Modifier.weight(1f))
        }
        RiskInsightCard(data = state.riskData, onOpenDetail = {})
        Row(modifier = Modifier.padding(horizontal = 4.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            Text("最近窗口", fontWeight = FontWeight.Bold, fontSize = 16.sp, color = Color(0xFF0F172A))
            Text("近7天", color = Color(0xFF94A3B8), fontSize = 12.sp)
        }
        state.riskData?.trend?.sortedByDescending { it.time_bucket }?.take(6)?.forEach { point ->
            Card(shape = RoundedCornerShape(18.dp), modifier = Modifier.padding(horizontal = 4.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                            Box(Modifier.width(4.dp).height(18.dp).clip(RoundedCornerShape(99.dp)).background(Color(0xFF0F172A)))
                            Text(point.time_bucket, fontWeight = FontWeight.Bold, color = Color(0xFF0F172A))
                        }
                        Box(
                            modifier = Modifier
                                .clip(RoundedCornerShape(8.dp))
                                .background(Color(0xFFF8FAFC))
                                .padding(horizontal = 8.dp, vertical = 4.dp),
                        ) {
                            Text("总计 ${point.total}", color = Color(0xFF94A3B8), fontSize = 12.sp, fontWeight = FontWeight.Bold)
                        }
                    }
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        MiniRiskCount("高风险", point.high, Color(0xFFF87171), Color(0xFFDC2626), Modifier.weight(1f))
                        MiniRiskCount("中风险", point.medium, Color(0xFFFBBF24), Color(0xFFD97706), Modifier.weight(1f))
                        MiniRiskCount("低风险", point.low, Color(0xFF34D399), Color(0xFF059669), Modifier.weight(1f))
                    }
                }
            }
        }
    }
}

@Composable
fun AlertsScreen(
    state: MainUiState,
    padding: PaddingValues,
    onOpenTask: (String) -> Unit,
) {
    val alerts = remember(state.alertItems, state.history) { recentAlertCases(state) }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(
            start = 12.dp,
            end = 12.dp,
            top = padding.calculateTopPadding() + 10.dp,
            bottom = padding.calculateBottomPadding() + 24.dp,
        ),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (alerts.isEmpty()) {
            item {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 56.dp, bottom = 24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    Box(
                        modifier = Modifier
                            .size(56.dp)
                            .clip(CircleShape)
                            .background(Color(0xFFF3F4F6)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(Icons.Outlined.NotificationsNone, contentDescription = null, tint = Color(0xFFD1D5DB), modifier = Modifier.size(28.dp))
                    }
                    Text("暂无风险预警", color = Color(0xFF9CA3AF), fontSize = 14.sp)
                }
            }
        } else {
            items(alerts, key = { it.record_id }) { alert ->
                val theme = personalAlertTheme(alert.risk_level)
                Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color.White), onClick = { onOpenTask(alert.record_id) }) {
                    Box(Modifier.fillMaxWidth()) {
                        if (state.alertItems.any { it.event.record_id == alert.record_id && !it.read }) {
                            Box(
                                modifier = Modifier
                                    .width(4.dp)
                                    .fillMaxSize()
                                    .background(theme.accent),
                            )
                        }
                        Column(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 16.dp, vertical = 16.dp),
                            verticalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                                Box(
                                    modifier = Modifier
                                        .clip(RoundedCornerShape(8.dp))
                                        .background(theme.soft)
                                        .padding(horizontal = 8.dp, vertical = 3.dp),
                                ) {
                                    Text("${normalizePersonalAlertRiskLabel(alert.risk_level)} Risk", color = theme.accent, fontSize = 10.sp, fontWeight = FontWeight.Bold)
                                }
                                Text(formatDateTime(alert.created_at.ifBlank { alert.sent_at }), color = Color(0xFF9CA3AF), fontSize = 12.sp)
                            }
                            Text(alert.title.ifBlank { "风险预警" }, fontWeight = FontWeight.Bold, fontSize = 16.sp, color = Color(0xFF0F172A), maxLines = 1, overflow = TextOverflow.Ellipsis)
                            Text(alert.case_summary.ifBlank { "${normalizePersonalAlertRiskLabel(alert.risk_level)}风险事件已触发预警，请及时核查。" }, color = Color(0xFF6B7280), fontSize = 14.sp, maxLines = 2, overflow = TextOverflow.Ellipsis)
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun SubmitScreen(
    state: MainUiState,
    padding: PaddingValues,
    viewModel: SentinelViewModel,
    onPickImages: () -> Unit,
    onPickAudios: () -> Unit,
    onPickVideos: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(top = padding.calculateTopPadding())
            .verticalScroll(rememberScrollState())
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Text("新建分析任务", style = MaterialTheme.typography.headlineMedium, modifier = Modifier.padding(horizontal = 4.dp))
        Card(shape = RoundedCornerShape(20.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
            Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
                AppTextField(
                    value = state.analyzeForm.text,
                    onValueChange = viewModel::updateAnalyzeText,
                    label = "描述案情或粘贴文本",
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(140.dp),
                )
                Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    UploadTile("图片", state.analyzeForm.images.size, Icons.Outlined.Image, Modifier.weight(1f), onPickImages)
                    UploadTile("音频", state.analyzeForm.audios.size, Icons.Outlined.Mic, Modifier.weight(1f), onPickAudios)
                    UploadTile("视频", state.analyzeForm.videos.size, Icons.Outlined.Videocam, Modifier.weight(1f), onPickVideos)
                }
                Button(
                    onClick = viewModel::submitAnalysis,
                    enabled = !state.analyzing,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(50.dp),
                    shape = RoundedCornerShape(18.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF059669)),
                ) {
                    if (state.analyzing) {
                        CircularProgressIndicator(modifier = Modifier.size(18.dp), strokeWidth = 2.dp, color = Color.White)
                    } else {
                        Text("开始分析")
                    }
                }
            }
        }
    }
}

@Composable
fun ChatScreen(
    state: MainUiState,
    padding: PaddingValues,
    onBack: () -> Unit,
    viewModel: SentinelViewModel,
    onPickImages: () -> Unit,
) {
    val listState = rememberLazyListState()

    LaunchedEffect(state.chatMessages.size) {
        if (state.chatMessages.isNotEmpty()) {
            listState.animateScrollToItem(state.chatMessages.lastIndex)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(top = padding.calculateTopPadding())
            .background(Color.White),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 10.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                IconButton(onClick = onBack) { Icon(Icons.Outlined.ArrowBack, contentDescription = "返回", tint = Color(0xFF475569)) }
                Column {
                    Text("Sentinel AI", fontWeight = FontWeight.Bold, fontSize = 14.sp, color = Color(0xFF1F2937))
                    Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                        Box(Modifier.size(6.dp).clip(CircleShape).background(Color(0xFF22C55E)))
                        Text("Online", color = Color(0xFF22C55E), fontSize = 10.sp, fontWeight = FontWeight.Medium)
                    }
                }
            }
            IconButton(onClick = viewModel::clearChatHistory) {
                Icon(Icons.Outlined.DeleteOutline, contentDescription = "清空对话", tint = Color(0xFF94A3B8))
            }
        }
        if (state.chatMessages.size <= 1 && !state.chatLoading) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 56.dp, bottom = 8.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Box(
                    modifier = Modifier
                        .size(64.dp)
                        .clip(CircleShape)
                        .background(Brush.linearGradient(listOf(Color(0xFF34D399), Color(0xFF14B8A6)))),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Outlined.Shield, contentDescription = null, tint = Color.White, modifier = Modifier.size(28.dp))
                }
                Text("我是您的反诈助手", fontWeight = FontWeight.Bold, color = Color(0xFF1E293B), fontSize = 18.sp)
                Text(
                    "可帮您识别诈骗信息、分析风险案例或提供安全建议。直接发送文字或图片即可。",
                    color = Color(0xFF64748B),
                    fontSize = 14.sp,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.width(260.dp),
                )
            }
        }
        LazyColumn(
            modifier = Modifier.weight(1f),
            state = listState,
            contentPadding = PaddingValues(horizontal = 12.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            items(state.chatMessages, key = { it.id }) { message ->
                ChatBubble(message)
            }
        }
        if (state.chatImages.isNotEmpty()) {
            LazyRow(
                contentPadding = PaddingValues(horizontal = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(state.chatImages.size) { index ->
                    Box {
                        Base64Thumbnail(dataUrl = state.chatImages[index], size = 72.dp)
                        IconButton(
                            onClick = { viewModel.removeChatImage(index) },
                            modifier = Modifier.align(Alignment.TopEnd),
                        ) {
                            Icon(Icons.Outlined.Close, contentDescription = "移除")
                        }
                    }
                }
            }
            Spacer(Modifier.height(8.dp))
        }
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp, vertical = 10.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(RoundedCornerShape(16.dp))
                    .background(Color(0xFFF9FAFB))
                    .border(1.dp, Color(0xFFE5E7EB), RoundedCornerShape(16.dp))
                    .padding(horizontal = 8.dp, vertical = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalAlignment = Alignment.Bottom,
            ) {
                Box(
                    modifier = Modifier
                        .size(36.dp)
                        .clip(CircleShape)
                        .clickable(enabled = !state.isChatting, onClick = onPickImages),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Outlined.Add, contentDescription = "上传图片", tint = Color(0xFF94A3B8))
                }
                BasicTextField(
                    value = state.chatInput,
                    onValueChange = viewModel::updateChatInput,
                    modifier = Modifier
                        .weight(1f)
                        .padding(vertical = 8.dp),
                    textStyle = MaterialTheme.typography.bodyMedium.copy(
                        color = Color(0xFF1E293B),
                        fontSize = 14.sp,
                        lineHeight = 20.sp,
                    ),
                    cursorBrush = SolidColor(Color(0xFF059669)),
                    decorationBox = { innerTextField ->
                        Box {
                            if (state.chatInput.isBlank()) {
                                Text("发送消息...", color = Color(0xFF94A3B8), fontSize = 14.sp)
                            }
                            innerTextField()
                        }
                    },
                )
                Box(
                    modifier = Modifier
                        .size(40.dp)
                        .clip(CircleShape)
                        .background(
                            if ((!state.isChatting && (state.chatInput.isNotBlank() || state.chatImages.isNotEmpty()))) {
                                Color(0xFF059669)
                            } else {
                                Color(0xFFCBD5E1)
                            },
                        )
                        .clickable(
                            enabled = !state.isChatting && (state.chatInput.isNotBlank() || state.chatImages.isNotEmpty()),
                            onClick = viewModel::sendChatMessage,
                        ),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Outlined.Send, contentDescription = "发送", tint = Color.White, modifier = Modifier.size(18.dp))
                }
            }
            Text(
                "Sentinel AI 可能会产生错误信息，请核实重要信息。",
                color = Color(0xFF94A3B8),
                fontSize = 10.sp,
                textAlign = TextAlign.Center,
            )
        }
    }
}

@Composable
fun FamilyScreen(
    state: MainUiState,
    padding: PaddingValues,
    onOpenManage: () -> Unit,
    viewModel: SentinelViewModel,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(top = padding.calculateTopPadding())
            .verticalScroll(rememberScrollState())
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        if (state.familyOverview.family == null) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 16.dp, bottom = 4.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .clip(CircleShape)
                        .background(Color(0xFFF3F4F6)),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(Icons.Outlined.Groups, contentDescription = null, tint = Color(0xFF9CA3AF), modifier = Modifier.size(28.dp))
                }
                Text("加入或创建家庭", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
            }
            Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
                Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text("创建新家庭", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    AppTextField(value = state.familyCreateName, onValueChange = viewModel::updateFamilyCreateName, label = "家庭名称")
                    Button(onClick = viewModel::createFamily, modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(18.dp), colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF059669))) { Text("创建") }
                }
            }
            Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
                Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text("邀请码加入", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Text("输入家人发来的邀请码，快速加入已有家庭。", color = Color(0xFF64748B), fontSize = 12.sp)
                    AppTextField(value = state.familyAcceptCode, onValueChange = viewModel::updateFamilyAcceptCode, label = "家庭邀请码")
                    Button(onClick = { viewModel.acceptFamilyInvitation() }, modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(18.dp), colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF059669))) { Text("加入家庭") }
                }
            }
            if (state.familyReceivedInvitations.isNotEmpty()) {
                SectionHeader("收到的邀请")
                state.familyReceivedInvitations.forEach { invitation ->
                    Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF0FDF4))) {
                        Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                                Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                                    Text(invitation.family_name.ifBlank { "家庭邀请" }, fontWeight = FontWeight.Bold, color = Color(0xFF1E293B))
                                    Text("邀请人：${invitation.inviter_name.ifBlank { invitation.inviter_email.ifBlank { invitation.inviter_phone.ifBlank { "未知" } } }}", color = Color(0xFF64748B), fontSize = 12.sp)
                                    Text("角色：${invitation.role} / 关系：${invitation.relation.ifBlank { "未填写" }}", color = Color(0xFF64748B), fontSize = 12.sp)
                                }
                                Box(
                                    modifier = Modifier
                                        .clip(RoundedCornerShape(999.dp))
                                        .background(Color.White)
                                        .border(1.dp, Color(0xFFA7F3D0), RoundedCornerShape(999.dp))
                                        .padding(horizontal = 10.dp, vertical = 5.dp),
                                ) {
                                    Text(if (invitation.status == "pending") "待处理" else invitation.status, color = Color(0xFF047857), fontSize = 11.sp, fontWeight = FontWeight.Bold)
                                }
                            }
                            Card(shape = RoundedCornerShape(14.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
                                Column(Modifier.padding(horizontal = 12.dp, vertical = 10.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                                    Text("邀请码", color = Color(0xFF94A3B8), fontSize = 11.sp, fontWeight = FontWeight.Bold)
                                    Text(invitation.invite_code, color = Color(0xFF1E293B), fontWeight = FontWeight.SemiBold)
                                }
                            }
                            Text("有效期至 ${formatDateTime(invitation.expires_at)}", color = Color(0xFF94A3B8), fontSize = 12.sp)
                            Button(
                                onClick = { viewModel.acceptFamilyInvitation(invitation.invite_code, invitation.id) },
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(18.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF059669)),
                            ) {
                                Text(if (state.acceptingInvitationIds.contains(invitation.id)) "加入中..." else "接受邀请")
                            }
                        }
                    }
                }
            }
        } else {
            FamilySummaryCard(state = state, onOpenManage = onOpenManage)
            SectionHeader("家庭成员", modifier = Modifier.padding(horizontal = 4.dp))
            if (state.familyOverview.members.isEmpty()) {
                EmptyState("暂无家庭成员", modifier = Modifier.padding(horizontal = 4.dp))
            } else {
                state.familyOverview.members.forEach { member ->
                    MemberCard(member, modifier = Modifier.padding(horizontal = 4.dp))
                }
            }
            SectionHeader("守护关系", modifier = Modifier.padding(horizontal = 4.dp))
            if (state.familyOverview.guardian_links.isEmpty()) {
                EmptyState("当前还没有守护关系", modifier = Modifier.padding(horizontal = 4.dp))
            } else {
                state.familyOverview.guardian_links.forEach { link ->
                    GuardianLinkCard(link, modifier = Modifier.padding(horizontal = 4.dp))
                }
            }
            SectionHeader("最新动态", modifier = Modifier.padding(horizontal = 4.dp))
            if (state.familyNotifications.isEmpty()) {
                EmptyState("无家庭动态", modifier = Modifier.padding(horizontal = 4.dp))
            } else {
                state.familyNotifications.forEach { notification ->
                    FamilyNotificationCard(notification = notification, onRead = { viewModel.markFamilyNotificationRead(notification.id) }, modifier = Modifier.padding(horizontal = 4.dp))
                }
            }
        }
    }
}

@Composable
fun FamilyManageScreen(
    state: MainUiState,
    padding: PaddingValues,
    onBack: () -> Unit,
    viewModel: SentinelViewModel,
) {
    val guardianOptions = state.familyOverview.members.filter { it.role == "owner" || it.role == "guardian" }
    val protectedOptions = state.familyOverview.members.filter { it.role != "owner" }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(top = padding.calculateTopPadding())
            .verticalScroll(rememberScrollState())
            .padding(horizontal = 12.dp, vertical = 8.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        ScreenBackHeader(title = "家庭管理", onBack = onBack)
        Card(shape = RoundedCornerShape(20.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
            Column(Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("邀请成员", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                AppTextField(value = state.familyInviteForm.inviteeEmail, onValueChange = viewModel::updateFamilyInviteEmail, label = "邮箱", keyboardType = KeyboardType.Email)
                AppTextField(value = state.familyInviteForm.inviteePhone, onValueChange = viewModel::updateFamilyInvitePhone, label = "手机号", keyboardType = KeyboardType.Phone)
                ChoiceRow(
                    title = "成员角色",
                    options = listOf("member" to "普通成员", "guardian" to "守护人"),
                    selected = state.familyInviteForm.role,
                    onSelected = viewModel::updateFamilyInviteRole,
                )
                AppTextField(value = state.familyInviteForm.relation, onValueChange = viewModel::updateFamilyInviteRelation, label = "关系", placeholder = "如：父亲")
                Button(onClick = viewModel::createFamilyInvitation, modifier = Modifier.fillMaxWidth().height(40.dp), shape = RoundedCornerShape(18.dp), colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF059669))) { Text("发送邀请") }
            }
        }
        Card(shape = RoundedCornerShape(20.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
            Column(Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("配置守护", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                ChoiceRow(
                    title = "守护人",
                    options = guardianOptions.map { it.user_id.toString() to it.username },
                    selected = state.familyGuardianForm.guardianUserId,
                    onSelected = viewModel::updateGuardianUser,
                )
                ChoiceRow(
                    title = "被守护人",
                    options = protectedOptions.map { it.user_id.toString() to it.username },
                    selected = state.familyGuardianForm.memberUserId,
                    onSelected = viewModel::updateProtectedUser,
                )
                Button(onClick = viewModel::createGuardianLink, modifier = Modifier.fillMaxWidth().height(40.dp), shape = RoundedCornerShape(18.dp), colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF059669))) { Text("保存关系") }
            }
        }
        SectionHeader("家庭成员", modifier = Modifier.padding(horizontal = 4.dp))
        state.familyOverview.members.forEach { member ->
            Card(shape = RoundedCornerShape(18.dp), modifier = Modifier.padding(horizontal = 4.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 14.dp, vertical = 12.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Column(verticalArrangement = Arrangement.spacedBy(4.dp), modifier = Modifier.weight(1f)) {
                        Text(member.username, fontWeight = FontWeight.Bold, color = Color(0xFF1E293B))
                        Text(member.email.ifBlank { member.phone.ifBlank { "-" } }, color = Color(0xFF64748B), fontSize = 12.sp)
                        Text(member.relation.ifBlank { member.role }, color = Color(0xFF94A3B8), fontSize = 12.sp)
                    }
                    if (state.familyOverview.current_member?.role == "owner" && member.role != "owner") {
                        TextButton(onClick = { viewModel.deleteFamilyMember(member.member_id) }) {
                            Text("移除", color = MaterialTheme.colorScheme.error)
                        }
                    }
                }
            }
        }
        SectionHeader("邀请记录", modifier = Modifier.padding(horizontal = 4.dp))
        if (state.familyOverview.invitations.isEmpty()) {
            EmptyState("暂无邀请记录", modifier = Modifier.padding(horizontal = 4.dp))
        } else {
            state.familyOverview.invitations.forEach { invitation ->
                Card(shape = RoundedCornerShape(18.dp), modifier = Modifier.padding(horizontal = 4.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
                    Column(Modifier.padding(horizontal = 14.dp, vertical = 12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Text(invitation.invitee_email.ifBlank { invitation.invitee_phone.ifBlank { "未指定目标" } }, fontWeight = FontWeight.Bold, color = Color(0xFF1E293B))
                        Text("角色：${invitation.role} / 关系：${invitation.relation.ifBlank { "未填写" }}", color = Color(0xFF64748B), fontSize = 12.sp)
                        Text("邀请码：${invitation.invite_code}", color = Color(0xFF059669), fontWeight = FontWeight.SemiBold, fontSize = 12.sp)
                        Text("状态：${invitation.status} / 截止：${formatDateTime(invitation.expires_at)}", color = Color(0xFF94A3B8), fontSize = 12.sp)
                    }
                }
            }
        }
    }
}

@Composable
fun ProfileScreen(
    state: MainUiState,
    padding: PaddingValues,
    onOpenPrivacy: () -> Unit,
    onQuickAnalyzeBubbleChange: (Boolean) -> Unit,
    onAccessibilityAutoAnalyzeChange: (Boolean) -> Unit,
    onLogout: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(top = padding.calculateTopPadding())
            .verticalScroll(rememberScrollState())
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Card(
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White),
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 4.dp),
        ) {
            Row(
                horizontalArrangement = Arrangement.spacedBy(16.dp),
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
            ) {
                AvatarCircle(label = state.user.username.take(1).ifBlank { "U" }, size = 64.dp)
                Column {
                    Text(state.user.username.ifBlank { "未设置用户名" }, style = MaterialTheme.typography.headlineMedium)
                    Box(
                        modifier = Modifier
                            .padding(top = 6.dp)
                            .clip(RoundedCornerShape(8.dp))
                            .background(Color(0xFFF3F4F6))
                            .padding(horizontal = 8.dp, vertical = 3.dp),
                    ) {
                        Text(state.user.role.ifBlank { "user" }.uppercase(), color = Color(0xFF6B7280), fontSize = 11.sp, fontWeight = FontWeight.Medium)
                    }
                }
            }
        }
        Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color.White), onClick = onOpenPrivacy, modifier = Modifier.padding(horizontal = 4.dp)) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Column {
                    Text("隐私资料", fontWeight = FontWeight.Bold)
                    Text("查看手机号、邮箱和年龄", color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                Icon(Icons.Outlined.PersonOutline, contentDescription = null)
            }
        }
        Card(
            shape = RoundedCornerShape(18.dp),
            colors = CardDefaults.cardColors(
                containerColor = if (state.quickAnalyzeBubbleEnabled) Color(0xFFF0FDF4) else Color.White,
            ),
            modifier = Modifier.padding(horizontal = 4.dp),
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 14.dp),
                horizontalArrangement = Arrangement.spacedBy(14.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Box(
                    modifier = Modifier
                        .size(42.dp)
                        .clip(RoundedCornerShape(14.dp))
                        .background(
                            if (state.quickAnalyzeBubbleEnabled) Color(0xFFD1FAE5) else Color(0xFFF1F5F9),
                        ),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        Icons.Outlined.Layers,
                        contentDescription = null,
                        tint = if (state.quickAnalyzeBubbleEnabled) Color(0xFF059669) else Color(0xFF475569),
                    )
                }
                Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Text("悬浮球快捷分析", fontWeight = FontWeight.Bold)
                    Text(
                        if (state.quickAnalyzeBubbleEnabled) {
                            "已开启，点击屏幕气泡会自动截屏并调用快速分析。"
                        } else {
                            "开启后会在屏幕常驻一个气泡，点击即可截屏并快速识别风险。"
                        },
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        fontSize = 13.sp,
                    )
                    Text(
                        "首次开启会申请悬浮窗和截屏权限。",
                        color = Color(0xFF64748B),
                        fontSize = 12.sp,
                    )
                }
                Switch(
                    checked = state.quickAnalyzeBubbleEnabled,
                    onCheckedChange = onQuickAnalyzeBubbleChange,
                )
            }
        }
        Card(
            shape = RoundedCornerShape(18.dp),
            colors = CardDefaults.cardColors(
                containerColor = if (state.accessibilityAutoAnalyzeEnabled) Color(0xFFFFFBEB) else Color.White,
            ),
            modifier = Modifier.padding(horizontal = 4.dp),
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 14.dp),
                horizontalArrangement = Arrangement.spacedBy(14.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Box(
                    modifier = Modifier
                        .size(42.dp)
                        .clip(RoundedCornerShape(14.dp))
                        .background(
                            if (state.accessibilityAutoAnalyzeEnabled) Color(0xFFFFEDD5) else Color(0xFFF1F5F9),
                        ),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        Icons.Outlined.Security,
                        contentDescription = null,
                        tint = if (state.accessibilityAutoAnalyzeEnabled) Color(0xFFD97706) else Color(0xFF475569),
                    )
                }
                Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Text("无障碍自动守护", fontWeight = FontWeight.Bold)
                    Text(
                        when {
                            state.accessibilityAutoAnalyzeEnabled && state.accessibilityAutoAnalyzePermissionGranted ->
                                "已开启，命中转账、验证码、私下收款、下载引导等敏感场景时会自动快速分析。"

                            state.accessibilityAutoAnalyzeEnabled ->
                                "已准备开启，请在系统无障碍设置里启用“反诈屏幕守护”。"

                            else ->
                                "开启后会实时判断屏幕元素，发现敏感场景时自动调用快速分析。"
                        },
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        fontSize = 13.sp,
                    )
                    Text(
                        "仅 Android 11 及以上支持自动截图分析。",
                        color = Color(0xFF64748B),
                        fontSize = 12.sp,
                    )
                }
                Switch(
                    checked = state.accessibilityAutoAnalyzeEnabled,
                    onCheckedChange = onAccessibilityAutoAnalyzeChange,
                )
            }
        }
        OutlinedButton(
            onClick = onLogout,
            modifier = Modifier.fillMaxWidth().padding(horizontal = 4.dp),
            shape = RoundedCornerShape(14.dp),
            colors = ButtonDefaults.outlinedButtonColors(
                containerColor = Color(0xFFFEF2F2),
                contentColor = MaterialTheme.colorScheme.error,
            ),
            border = null,
        ) {
            Text("退出登录")
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun ProfilePrivacyScreen(
    state: MainUiState,
    padding: PaddingValues,
    onBack: () -> Unit,
    viewModel: SentinelViewModel,
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(top = padding.calculateTopPadding())
            .verticalScroll(rememberScrollState())
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        ScreenBackHeader(title = "隐私资料", onBack = onBack)
        InfoCard(icon = Icons.Outlined.Phone, title = "手机号", value = state.user.phone.ifBlank { "未设置" }, modifier = Modifier.padding(horizontal = 4.dp))
        InfoCard(icon = Icons.Outlined.Email, title = "邮箱", value = state.user.email.ifBlank { "未设置" }, modifier = Modifier.padding(horizontal = 4.dp))
        Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color.White), modifier = Modifier.padding(horizontal = 4.dp)) {
            Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                    Column {
                        Text("画像资料", fontWeight = FontWeight.Bold)
                        Text("年龄 ${(state.user.age ?: state.ageInput.toIntOrNull() ?: 28)}", color = MaterialTheme.colorScheme.onSurfaceVariant)
                        Text("职业 ${state.user.occupation.ifBlank { "未设置" }}", color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                    TextButton(onClick = viewModel::toggleAgeEditor) {
                        Text(if (state.ageEditorVisible) "收起" else "编辑")
                    }
                }
                if (state.user.recent_tags.isNotEmpty()) {
                    FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        state.user.recent_tags.forEach { tag ->
                            FilterChip(
                                selected = false,
                                onClick = {},
                                enabled = false,
                                label = { Text(tag) },
                            )
                        }
                    }
                } else {
                    Text("近期标签未设置", color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 13.sp)
                }
                AnimatedVisibility(visible = state.ageEditorVisible) {
                    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        AppTextField(
                            value = state.ageInput,
                            onValueChange = viewModel::updateAgeInput,
                            label = "年龄",
                            keyboardType = KeyboardType.Number,
                            modifier = Modifier.fillMaxWidth(),
                        )
                        if (state.occupationOptions.isNotEmpty()) {
                            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                Text("职业", fontWeight = FontWeight.Bold)
                                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                    state.occupationOptions.forEach { occupation ->
                                        FilterChip(
                                            selected = state.occupationInput == occupation,
                                            onClick = { viewModel.updateOccupationInput(occupation) },
                                            label = { Text(occupation) },
                                        )
                                    }
                                }
                            }
                        } else {
                            AppTextField(
                                value = state.occupationInput,
                                onValueChange = viewModel::updateOccupationInput,
                                label = "职业",
                                modifier = Modifier.fillMaxWidth(),
                            )
                        }
                        Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
                            Column(Modifier.fillMaxWidth().padding(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                Text("近期标签", fontWeight = FontWeight.Bold)
                                if (state.user.recent_tags.isNotEmpty()) {
                                    FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                        state.user.recent_tags.forEach { tag ->
                                            FilterChip(
                                                selected = false,
                                                onClick = {},
                                                enabled = false,
                                                label = { Text(tag) },
                                            )
                                        }
                                    }
                                } else {
                                    Text("近期标签未设置", color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 13.sp)
                                }
                            }
                        }
                        Button(onClick = viewModel::updateAge, modifier = Modifier.fillMaxWidth()) { Text("保存资料") }
                    }
                }
            }
        }
        OutlinedButton(
            onClick = viewModel::deleteAccount,
            modifier = Modifier.fillMaxWidth().padding(horizontal = 4.dp),
            shape = RoundedCornerShape(14.dp),
            colors = ButtonDefaults.outlinedButtonColors(
                containerColor = Color(0xFFFEF2F2),
                contentColor = MaterialTheme.colorScheme.error,
            ),
            border = null,
        ) {
            Text("删除账户")
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun TaskDetailSheet(
    task: TaskDetail,
    onDismiss: () -> Unit,
    onMessage: (String, Boolean) -> Unit,
) {
    val context = LocalContext.current
    var exportMenuExpanded by remember { mutableStateOf<Boolean>(false) }
    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 20.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                Column {
                    Text("任务详情", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold)
                    Text("ID: ${task.task_id}", color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 12.sp)
                }
                IconButton(onClick = onDismiss) {
                    Icon(Icons.Outlined.Close, contentDescription = "关闭", tint = Color(0xFF94A3B8))
                }
            }
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                StatusBadge(task.status)
                Text(formatDateTime(task.created_at), color = Color(0xFF64748B), fontSize = 12.sp)
            }
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End,
            ) {
                Box {
                    Box(
                        modifier = Modifier
                            .size(44.dp)
                            .clip(CircleShape)
                            .clickable { exportMenuExpanded = !exportMenuExpanded },
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            Icons.Outlined.Add,
                            contentDescription = "导出",
                            tint = if (exportMenuExpanded) Color(0xFF059669) else Color(0xFF0F172A),
                            modifier = Modifier.size(24.dp),
                        )
                    }
                    DropdownMenu(
                        expanded = exportMenuExpanded,
                        onDismissRequest = { exportMenuExpanded = false },
                        modifier = Modifier.width(176.dp),
                    ) {
                        DropdownMenuItem(
                            text = {
                                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                    Text("Markdown", color = Color(0xFF0F172A), fontWeight = FontWeight.SemiBold)
                                    Text("导出结构化文本报告", color = Color(0xFF94A3B8), fontSize = 11.sp)
                                }
                            },
                            onClick = {
                                saveTextToDownloads(
                                    context = context,
                                    filename = buildExportFilename(task, "md"),
                                    mimeType = "text/markdown",
                                    content = buildTaskMarkdown(task),
                                    onMessage = onMessage,
                                )
                                exportMenuExpanded = false
                            },
                        )
                        DropdownMenuItem(
                            text = {
                                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                    Text("JSON", color = Color(0xFF0F172A), fontWeight = FontWeight.SemiBold)
                                    Text("导出完整结构化数据", color = Color(0xFF94A3B8), fontSize = 11.sp)
                                }
                            },
                            onClick = {
                                val json = Json {
                                    prettyPrint = true
                                    ignoreUnknownKeys = true
                                }
                                saveTextToDownloads(
                                    context = context,
                                    filename = buildExportFilename(task, "json"),
                                    mimeType = "application/json",
                                    content = json.encodeToString(task),
                                    onMessage = onMessage,
                                )
                                exportMenuExpanded = false
                            },
                        )
                    }
                }
            }
            if (task.summary.isNotBlank()) {
                Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
                    Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Text("案件摘要", fontWeight = FontWeight.Bold, color = Color(0xFF334155), fontSize = 14.sp)
                        Text(task.summary, color = Color(0xFF475569), fontSize = 14.sp)
                    }
                }
            }
            if (task.risk_score > 0 || task.risk_summary.isNotBlank()) {
                val parsedRiskSummary = parseRiskSummary(task.risk_summary)
                Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF0FDF4))) {
                    Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                            Text("风险评分", fontWeight = FontWeight.Bold, color = Color(0xFF166534), fontSize = 14.sp)
                            Text(task.risk_score.toString(), color = Color(0xFF059669), fontWeight = FontWeight.ExtraBold, fontSize = 16.sp)
                        }
                        if (parsedRiskSummary != null) {
                            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                RiskScoreLine("社工话术", parsedRiskSummary.dimensions["social_engineering"] ?: 0)
                                RiskScoreLine("诱导动作", parsedRiskSummary.dimensions["requested_actions"] ?: 0)
                                RiskScoreLine("证据强度", parsedRiskSummary.dimensions["evidence_strength"] ?: 0)
                                RiskScoreLine("受害暴露", parsedRiskSummary.dimensions["loss_exposure"] ?: 0)
                            }
                            if (parsedRiskSummary.hitRules.isNotEmpty()) {
                                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                    parsedRiskSummary.hitRules.forEach { item ->
                                        Box(
                                            modifier = Modifier
                                                .clip(RoundedCornerShape(999.dp))
                                                .background(Color.White)
                                                .border(1.dp, Color(0xFFD1FAE5), RoundedCornerShape(999.dp))
                                                .padding(horizontal = 10.dp, vertical = 6.dp),
                                        ) {
                                            Text(item, color = Color(0xFF065F46), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                                        }
                                    }
                                }
                            }
                        } else {
                            Text(task.risk_summary, color = Color(0xFF475569), fontSize = 13.sp)
                        }
                    }
                }
            }
            if (task.report.isNotBlank()) {
                Card(shape = RoundedCornerShape(20.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
                    Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Text("综合分析报告", fontWeight = FontWeight.Bold, color = Color(0xFF334155))
                        val sections = parseReport(task.report)
                        if (sections.isEmpty()) {
                            Text(task.report, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        } else {
                            sections.forEach { section ->
                                Text(section.title, fontWeight = FontWeight.SemiBold, color = Color(0xFF0F172A))
                                Text(section.content, color = MaterialTheme.colorScheme.onSurfaceVariant)
                            }
                        }
                    }
                }
                val attackSteps = extractAttackSteps(task.report)
                if (attackSteps.isNotEmpty()) {
                    Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFFFF1F2))) {
                        Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            Text("诈骗链路时间线", fontWeight = FontWeight.Bold, color = Color(0xFFBE123C), fontSize = 14.sp)
                            attackSteps.forEachIndexed { index, step ->
                                Text("${index + 1}. $step", color = Color(0xFF475569), fontSize = 14.sp)
                            }
                        }
                    }
                }
                val keywords = extractKeywordSentences(task.report)
                if (keywords.isNotEmpty()) {
                    Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFFEF4FF))) {
                        Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            Text("诈骗关键词句", fontWeight = FontWeight.Bold, color = Color(0xFFC026D3), fontSize = 14.sp)
                            FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                keywords.forEach { keyword ->
                                    Box(
                                        modifier = Modifier
                                            .clip(RoundedCornerShape(999.dp))
                                            .background(Color.White)
                                            .border(1.dp, Color(0xFFF5D0FE), RoundedCornerShape(999.dp))
                                            .padding(horizontal = 10.dp, vertical = 6.dp),
                                    ) {
                                        Text(keyword, color = Color(0xFFC026D3), fontSize = 12.sp, fontWeight = FontWeight.Bold)
                                    }
                                }
                            }
                        }
                    }
                }
            }
            if (task.risk_level.isNotBlank()) {
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    Text("风险等级", color = Color(0xFF475569), fontWeight = FontWeight.Bold)
                    RiskBadge(task.risk_level)
                }
            }
            Card(shape = RoundedCornerShape(20.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
                Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text("输入概览", fontWeight = FontWeight.Bold)
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                        StatMiniCard("文本", if (task.payload.text.isNotBlank()) "已提供" else "无", Modifier.weight(1f))
                        StatMiniCard("图片", task.payload.images.size.toString(), Modifier.weight(1f))
                    }
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                        StatMiniCard("音频", task.payload.audios.size.toString(), Modifier.weight(1f))
                        StatMiniCard("视频", task.payload.videos.size.toString(), Modifier.weight(1f))
                    }
                    Card(
                        shape = RoundedCornerShape(14.dp),
                        colors = CardDefaults.cardColors(containerColor = Color.White),
                    ) {
                        Text(
                            "手机端主详情页只展示分析结果与输入概览，原始多模态材料不在这里展开。",
                            color = Color(0xFF94A3B8),
                            fontSize = 12.sp,
                            modifier = Modifier.padding(horizontal = 14.dp, vertical = 12.dp),
                        )
                    }
                }
            }
            Spacer(Modifier.height(24.dp))
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AlertSheet(
    event: AlertEvent,
    onDismiss: () -> Unit,
    onOpenCase: () -> Unit,
) {
    val theme = personalAlertTheme(event.risk_level)
    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(Modifier.fillMaxWidth().padding(20.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("风险预警", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold)
            RiskBadge(event.risk_level)
            Text(event.title.ifBlank { "风险预警" }, fontWeight = FontWeight.Bold, color = Color(0xFF0F172A))
            Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = theme.surface)) {
                Text(
                    event.case_summary.ifBlank { "${normalizePersonalAlertRiskLabel(event.risk_level)}风险事件已触发预警，请及时核查。" },
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.padding(16.dp),
                )
            }
            Row(horizontalArrangement = Arrangement.spacedBy(18.dp)) {
                Text("诈骗类型：${event.scam_type.ifBlank { "未知类型" }}", color = Color(0xFF64748B), fontSize = 12.sp)
                Text("案件 ID：${event.record_id}", color = Color(0xFF64748B), fontSize = 12.sp)
            }
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), modifier = Modifier.fillMaxWidth()) {
                OutlinedButton(onClick = onDismiss, modifier = Modifier.weight(1f), shape = RoundedCornerShape(14.dp)) { Text("稍后处理") }
                Button(onClick = onOpenCase, modifier = Modifier.weight(1f), shape = RoundedCornerShape(14.dp), colors = ButtonDefaults.buttonColors(containerColor = theme.action)) { Text("查看案件") }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FamilyAlertSheet(
    notification: FamilyNotification,
    onDismiss: () -> Unit,
    onOpenCenter: () -> Unit,
) {
    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(Modifier.fillMaxWidth().padding(20.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("家庭联防通知", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold)
            RiskBadge(notification.risk_level.ifBlank { "高" })
            Text(notification.title.ifBlank { "高风险案件预警" }, fontWeight = FontWeight.Bold, color = Color(0xFF0F172A))
            Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFFFF1F2))) {
                Text(
                    notification.case_summary.ifBlank { notification.summary.ifBlank { "家庭成员触发高风险案件，请及时核查。" } },
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.padding(16.dp),
                )
            }
            Row(horizontalArrangement = Arrangement.spacedBy(18.dp)) {
                Text("家庭成员：${notification.target_name}", color = Color(0xFF64748B), fontSize = 12.sp)
                Text("诈骗类型：${notification.scam_type.ifBlank { "待分析" }}", color = Color(0xFF64748B), fontSize = 12.sp)
            }
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), modifier = Modifier.fillMaxWidth()) {
                OutlinedButton(onClick = onDismiss, modifier = Modifier.weight(1f), shape = RoundedCornerShape(14.dp)) { Text("稍后处理") }
                Button(onClick = onOpenCenter, modifier = Modifier.weight(1f), shape = RoundedCornerShape(14.dp), colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFDC2626))) { Text("进入家庭中心") }
            }
        }
    }
}

@Composable
private fun AppTextField(
    value: String,
    onValueChange: (String) -> Unit,
    label: String,
    modifier: Modifier = Modifier,
    placeholder: String = "",
    keyboardType: KeyboardType = KeyboardType.Text,
    imeAction: ImeAction = ImeAction.Default,
) {
    OutlinedTextField(
        value = value,
        onValueChange = onValueChange,
        modifier = modifier,
        label = { Text(label) },
        placeholder = if (placeholder.isBlank()) null else ({ Text(placeholder) }),
        keyboardOptions = KeyboardOptions(keyboardType = keyboardType, imeAction = imeAction),
    )
}

@Composable
private fun DashboardQuickActions(onScreenChange: (AppScreen) -> Unit) {
    Card(
        shape = RoundedCornerShape(16.dp),
        modifier = Modifier
            .padding(horizontal = 16.dp)
            .offset(y = (-2).dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 14.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            QuickAction("AI助手", Icons.Outlined.ChatBubbleOutline, Color(0xFF10B981)) { onScreenChange(AppScreen.Chat) }
            QuickAction("历史档案", Icons.Outlined.Description, Color(0xFF14B8A6)) { onScreenChange(AppScreen.History) }
            QuickAction("风险趋势", Icons.Outlined.ShowChart, Color(0xFF06B6D4)) { onScreenChange(AppScreen.RiskTrend) }
            QuickAction("家庭守护", Icons.Outlined.Groups, Color(0xFF84CC16)) { onScreenChange(AppScreen.Family) }
        }
    }
}

@Composable
private fun QuickAction(
    label: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    color: Color,
    onClick: () -> Unit,
) {
    val gradient = when (label) {
        "AI助手" -> Brush.verticalGradient(listOf(Color(0xFF34D399), Color(0xFF10B981)))
        "历史档案" -> Brush.verticalGradient(listOf(Color(0xFF2DD4BF), Color(0xFF14B8A6)))
        "风险趋势" -> Brush.verticalGradient(listOf(Color(0xFF22D3EE), Color(0xFF06B6D4)))
        else -> Brush.verticalGradient(listOf(Color(0xFFA3E635), Color(0xFF84CC16)))
    }
    Column(horizontalAlignment = Alignment.CenterHorizontally, modifier = Modifier.clickable(onClick = onClick)) {
        Box(
            modifier = Modifier
                .size(52.dp)
                .shadow(12.dp, RoundedCornerShape(18.dp), clip = false)
                .clip(RoundedCornerShape(18.dp))
                .background(gradient),
            contentAlignment = Alignment.Center,
        ) {
            Icon(icon, contentDescription = label, tint = Color.White)
        }
        Spacer(Modifier.height(8.dp))
        Text(label, fontSize = 12.sp, textAlign = TextAlign.Center, fontWeight = FontWeight.Bold, color = Color(0xFF334155))
    }
}

@Composable
private fun StatCard(
    title: String,
    value: String,
    accent: Color,
    modifier: Modifier = Modifier,
) {
    val iconColor = when (title) {
        "今日检测", "总检测" -> Color(0xFF059669)
        else -> Color(0xFFF97316)
    }
    val icon = when (title) {
        "今日检测", "总检测" -> Icons.Outlined.Security
        else -> Icons.Outlined.WarningAmber
    }
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(96.dp)
                .background(Brush.verticalGradient(listOf(accent.copy(alpha = 0.08f), Color.White))),
        ) {
            Box(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .size(72.dp)
                    .offset(y = (-18).dp)
                    .clip(RoundedCornerShape(bottomStart = 72.dp))
                    .background(accent.copy(alpha = 0.12f)),
            )
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(16.dp),
                verticalArrangement = Arrangement.Center,
            ) {
                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Text(title, color = accent, fontWeight = FontWeight.Bold, fontSize = 12.sp)
                    Text(value, style = MaterialTheme.typography.headlineLarge, color = Color(0xFF0F172A))
                }
            }
            Icon(
                icon,
                contentDescription = null,
                tint = iconColor,
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(top = 16.dp, end = 16.dp)
                    .size(22.dp),
            )
        }
    }
}

@Composable
private fun StatMiniCard(
    title: String,
    value: String,
    modifier: Modifier = Modifier,
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(18.dp),
        colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC)),
    ) {
        Column(Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(title, color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 12.sp)
            Text(value, fontWeight = FontWeight.Bold)
        }
    }
}

@Composable
private fun SectionHeader(
    title: String,
    action: String? = null,
    onAction: (() -> Unit)? = null,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Box(
                Modifier
                    .width(4.dp)
                    .height(18.dp)
                    .clip(RoundedCornerShape(99.dp))
                    .background(Color(0xFF059669)),
            )
            Text(title, style = MaterialTheme.typography.titleMedium, color = Color(0xFF0F172A), fontWeight = FontWeight.Bold)
        }
        if (action != null && onAction != null) {
            TextButton(onClick = onAction) { Text(action, color = Color(0xFF6B7280), fontSize = 12.sp) }
        }
    }
}

@Composable
private fun MiniRiskCount(
    title: String,
    value: Int,
    backgroundAccent: Color,
    textAccent: Color,
    modifier: Modifier = Modifier,
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = backgroundAccent.copy(alpha = 0.12f)),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 10.dp, horizontal = 8.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Text(title, color = backgroundAccent, fontSize = 10.sp, fontWeight = FontWeight.Bold)
            Text(value.toString(), color = textAccent, fontSize = 20.sp, fontWeight = FontWeight.ExtraBold)
        }
    }
}

@Composable
private fun EmptyState(text: String, modifier: Modifier = Modifier) {
    Card(modifier = modifier, shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(24.dp),
            contentAlignment = Alignment.Center,
        ) {
            Text(text, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
private fun TaskCard(task: TaskSummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Card(modifier = modifier, shape = RoundedCornerShape(20.dp), onClick = onClick) {
        Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                StatusBadge(task.status)
                Text(formatDateTime(task.created_at), color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 12.sp)
            }
            Text(task.title.ifBlank { "未命名任务" }, fontWeight = FontWeight.Bold, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(task.summary.ifBlank { "等待分析..." }, color = MaterialTheme.colorScheme.onSurfaceVariant, maxLines = 1, overflow = TextOverflow.Ellipsis)
        }
    }
}

@Composable
private fun HistoryCard(
    item: HistoryRecord,
    deleting: Boolean,
    onOpen: () -> Unit,
    onDelete: () -> Unit,
) {
    Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 14.dp, vertical = 12.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.Top,
        ) {
            Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    RiskBadge(item.risk_level)
                    Text(item.record_id.take(6), color = Color(0xFF94A3B8), fontSize = 10.sp)
                }
                Text(
                    item.title.ifBlank { "无标题" },
                    fontWeight = FontWeight.ExtraBold,
                    fontSize = 13.sp,
                    color = Color(0xFF0F172A),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Text(formatDateTime(item.created_at), color = Color(0xFF94A3B8), fontSize = 10.sp, maxLines = 1, overflow = TextOverflow.Ellipsis)
                    Box(Modifier.size(3.dp).clip(CircleShape).background(Color(0xFFCBD5E1)))
                    Text(item.scam_type.ifBlank { "未识别类型" }, color = Color(0xFF94A3B8), fontSize = 10.sp, maxLines = 1, overflow = TextOverflow.Ellipsis)
                }
            }
            Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                IconButton(
                    onClick = onOpen,
                    modifier = Modifier
                        .size(32.dp)
                        .clip(CircleShape)
                        .background(Color(0xFFF1F5F9)),
                ) {
                    Icon(Icons.Outlined.ArrowBack, contentDescription = "打开", modifier = Modifier.size(16.dp), tint = Color(0xFF64748B))
                }
                IconButton(
                    onClick = onDelete,
                    enabled = !deleting,
                    modifier = Modifier
                        .size(32.dp)
                        .clip(CircleShape)
                        .background(Color(0xFFFEF2F2)),
                ) {
                    if (deleting) {
                        CircularProgressIndicator(modifier = Modifier.size(16.dp), strokeWidth = 2.dp, color = Color(0xFFDC2626))
                    } else {
                        Icon(Icons.Outlined.DeleteOutline, contentDescription = "删除", modifier = Modifier.size(16.dp), tint = Color(0xFFEF4444))
                    }
                }
            }
        }
    }
}

@Composable
private fun FamilySummaryCard(
    state: MainUiState,
    onOpenManage: () -> Unit,
) {
    val family = state.familyOverview.family ?: return

    Card(
        shape = RoundedCornerShape(28.dp),
        modifier = Modifier.padding(horizontal = 4.dp),
        colors = CardDefaults.cardColors(containerColor = Color.Transparent),
        elevation = CardDefaults.cardElevation(defaultElevation = 8.dp),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Brush.linearGradient(listOf(Color(0xFF10B981), Color(0xFF14B8A6), Color(0xFF06B6D4))))
                .padding(18.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.Top) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(family.name, color = Color.White, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold, maxLines = 1, overflow = TextOverflow.Ellipsis)
                    Text("成员 ${family.member_count} 人 · 未读通知 ${state.familyNotifications.count { it.read_at.isBlank() }} 条", color = Color.White.copy(alpha = 0.78f), fontSize = 12.sp)
                }
                OutlinedButton(
                    onClick = onOpenManage,
                    shape = RoundedCornerShape(999.dp),
                    colors = ButtonDefaults.outlinedButtonColors(
                        containerColor = Color.White.copy(alpha = 0.16f),
                        contentColor = Color.White,
                    ),
                ) {
                    Text("管理/邀请", fontSize = 12.sp)
                }
            }
            Card(
                shape = RoundedCornerShape(18.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White.copy(alpha = 0.12f)),
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 12.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.Top,
                ) {
                    Column(verticalArrangement = Arrangement.spacedBy(4.dp), modifier = Modifier.weight(1f)) {
                        Text("家庭邀请码", color = Color.White.copy(alpha = 0.75f), fontSize = 11.sp, fontWeight = FontWeight.Bold)
                        Text(family.invite_code.ifBlank { "暂无邀请码" }, color = Color.White, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                    }
                    Text("加入家庭时使用", color = Color.White.copy(alpha = 0.7f), fontSize = 11.sp)
                }
            }
            Row(horizontalArrangement = Arrangement.spacedBy((-6).dp)) {
                state.familyOverview.members.take(6).forEach { member ->
                    Box(
                        modifier = Modifier
                            .size(34.dp)
                            .clip(CircleShape)
                            .background(Color.White.copy(alpha = 0.9f))
                            .border(2.dp, Color.White.copy(alpha = 0.75f), CircleShape),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(member.username.take(1).ifBlank { "U" }, color = Color(0xFF0F766E), fontSize = 11.sp, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}

@Composable
private fun MemberCard(
    member: FamilyMember,
    modifier: Modifier = Modifier,
) {
    Card(
        shape = RoundedCornerShape(18.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        modifier = modifier,
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .size(38.dp)
                    .clip(RoundedCornerShape(14.dp))
                    .background(Color(0xFFECFDF5)),
                contentAlignment = Alignment.Center,
            ) {
                Text(member.username.take(1).ifBlank { "U" }, color = Color(0xFF047857), fontWeight = FontWeight.Bold, fontSize = 12.sp)
            }
            Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    Text(member.username, fontWeight = FontWeight.SemiBold, color = Color(0xFF1E293B))
                    Box(
                        modifier = Modifier
                            .clip(RoundedCornerShape(999.dp))
                            .background(Color(0xFFECFDF5))
                            .border(1.dp, Color(0xFFD1FAE5), RoundedCornerShape(999.dp))
                            .padding(horizontal = 8.dp, vertical = 2.dp),
                    ) {
                        Text(member.role, color = Color(0xFF047857), fontSize = 11.sp, fontWeight = FontWeight.Bold)
                    }
                }
                Text(member.relation.ifBlank { "未设置关系" }, color = Color(0xFF64748B), fontSize = 12.sp)
                Text(member.email.ifBlank { member.phone.ifBlank { "-" } }, color = Color(0xFF94A3B8), fontSize = 12.sp)
            }
        }
    }
}

@Composable
private fun GuardianLinkCard(link: GuardianLink, modifier: Modifier = Modifier) {
    Card(shape = RoundedCornerShape(18.dp), colors = CardDefaults.cardColors(containerColor = Color.White), modifier = modifier) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.Top,
        ) {
            Column(verticalArrangement = Arrangement.spacedBy(4.dp), modifier = Modifier.weight(1f)) {
                Text("${link.guardian_name} -> ${link.member_name}", fontWeight = FontWeight.SemiBold, color = Color(0xFF1E293B))
                Text("${link.guardian_email} / ${link.member_email}", color = Color(0xFF64748B), fontSize = 12.sp)
            }
            Box(
                modifier = Modifier
                    .clip(RoundedCornerShape(999.dp))
                    .background(Color(0xFFECFEFF))
                    .border(1.dp, Color(0xFFCFFAFE), RoundedCornerShape(999.dp))
                    .padding(horizontal = 8.dp, vertical = 3.dp),
            ) {
                Text("守护中", color = Color(0xFF0E7490), fontSize = 11.sp, fontWeight = FontWeight.Bold)
            }
        }
    }
}

@Composable
private fun FamilyNotificationCard(
    notification: FamilyNotification,
    onRead: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Card(shape = RoundedCornerShape(18.dp), onClick = onRead, modifier = modifier, colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(14.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.Top,
        ) {
            Box(
                modifier = Modifier
                    .size(32.dp)
                    .clip(CircleShape)
                    .background(Color(0xFFFEF2F2)),
                contentAlignment = Alignment.Center,
            ) {
                Icon(Icons.Outlined.WarningAmber, contentDescription = null, tint = Color(0xFFDC2626), modifier = Modifier.size(18.dp))
            }
            Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .clip(RoundedCornerShape(8.dp))
                            .background(Color(0xFFFEE2E2))
                            .padding(horizontal = 8.dp, vertical = 3.dp),
                    ) {
                        Text("${notification.risk_level.ifBlank { "高" }}风险", color = Color(0xFFDC2626), fontSize = 10.sp, fontWeight = FontWeight.Bold)
                    }
                    Text(formatDateTime(notification.event_at), color = Color(0xFF9CA3AF), fontSize = 10.sp)
                }
                Text(notification.title.ifBlank { "高风险案件预警" }, fontWeight = FontWeight.Bold, color = Color(0xFF1E293B))
                Text(notification.case_summary.ifBlank { notification.summary }, color = Color(0xFF64748B), maxLines = 2, overflow = TextOverflow.Ellipsis, fontSize = 12.sp)
                Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text("成员：${notification.target_name}", color = Color(0xFF94A3B8), fontSize = 10.sp)
                    if (notification.scam_type.isNotBlank()) {
                        Text("类型：${notification.scam_type}", color = Color(0xFF94A3B8), fontSize = 10.sp)
                    }
                }
            }
        }
    }
}

@Composable
private fun StatusBanner(
    title: String,
    value: String,
) {
    Card(shape = RoundedCornerShape(18.dp)) {
        Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(title, color = MaterialTheme.colorScheme.onSurfaceVariant)
            Text(value, fontWeight = FontWeight.Bold)
        }
    }
}

@Composable
private fun ScreenBackHeader(
    title: String,
    subtitle: String? = null,
    onBack: () -> Unit,
    trailing: @Composable (() -> Unit)? = null,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
            IconButton(onClick = onBack) { Icon(Icons.Outlined.ArrowBack, contentDescription = "返回") }
            Column {
                Text(title, style = MaterialTheme.typography.headlineMedium)
                if (subtitle != null) {
                    Text(subtitle, color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 12.sp)
                }
            }
        }
        trailing?.invoke()
    }
}

@Composable
private fun InfoCard(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    title: String,
    value: String,
    modifier: Modifier = Modifier,
) {
    Card(shape = RoundedCornerShape(20.dp), modifier = modifier, colors = CardDefaults.cardColors(containerColor = Color.White)) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(icon, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
            Column {
                Text(title, color = MaterialTheme.colorScheme.onSurfaceVariant)
                Text(value, fontWeight = FontWeight.Bold)
            }
        }
    }
}

@Composable
private fun AvatarCircle(
    label: String,
    size: Dp,
) {
    Box(
        modifier = Modifier
            .size(size)
            .clip(CircleShape)
            .background(Color(0xFFE2E8F0)),
        contentAlignment = Alignment.Center,
    ) {
        Text(label.uppercase(), fontWeight = FontWeight.Bold)
    }
}

@Composable
private fun UploadTile(
    title: String,
    count: Int,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    Box(
        modifier = modifier
            .aspectRatio(1f)
            .clip(RoundedCornerShape(20.dp))
            .border(1.dp, Color(0xFFD1D5DB), RoundedCornerShape(20.dp))
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Icon(icon, contentDescription = title, tint = MaterialTheme.colorScheme.onSurfaceVariant)
            Text(title, color = MaterialTheme.colorScheme.onSurfaceVariant)
            if (count > 0) {
                Text(count.toString(), color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.Bold)
            }
        }
    }
}

@Composable
private fun ChoiceRow(
    title: String,
    options: List<Pair<String, String>>,
    selected: String,
    onSelected: (String) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(title, fontWeight = FontWeight.Bold)
        Row(
            modifier = Modifier.horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            options.forEach { option ->
                FilterChip(
                    selected = selected == option.first,
                    onClick = { onSelected(option.first) },
                    label = { Text(option.second) },
                )
            }
        }
    }
}

@Composable
private fun ChatBubble(message: DisplayChatMessage) {
    val isUser = message.type == "user"
    val messageText = message.content.ifBlank { if (message.type == "ai") "..." else "" }
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = if (isUser) Arrangement.End else Arrangement.Start,
    ) {
        if (message.type == "ai") {
            Column(
                modifier = Modifier
                    .width(280.dp)
                    .padding(horizontal = 2.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                if (message.images.isNotEmpty()) {
                    Row(
                        modifier = Modifier.horizontalScroll(rememberScrollState()),
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        message.images.forEach { image ->
                            Base64Thumbnail(dataUrl = image, size = 72.dp)
                        }
                    }
                }
                MarkdownText(
                    markdown = messageText,
                    color = MaterialTheme.colorScheme.onSurface,
                    style = MaterialTheme.typography.bodyMedium.copy(lineHeight = 22.sp),
                )
            }
        } else {
            Card(
                shape = RoundedCornerShape(20.dp),
                colors = CardDefaults.cardColors(
                    containerColor = when (message.type) {
                        "user" -> Color(0xFF0F172A)
                        "tool" -> Color(0xFFF8FAFC)
                        "error" -> Color(0xFFFEE2E2)
                        else -> Color.White
                    },
                ),
            ) {
                Column(
                    modifier = Modifier
                        .width(280.dp)
                        .padding(14.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    if (message.images.isNotEmpty()) {
                        Row(
                            modifier = Modifier.horizontalScroll(rememberScrollState()),
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            message.images.forEach { image ->
                                Base64Thumbnail(dataUrl = image, size = 72.dp)
                            }
                        }
                    }
                    Text(
                        text = messageText,
                        color = when (message.type) {
                            "user" -> Color.White
                            "error" -> Color(0xFFB91C1C)
                            else -> MaterialTheme.colorScheme.onSurface
                        },
                        style = MaterialTheme.typography.bodyMedium.copy(lineHeight = 22.sp),
                    )
                }
            }
        }
    }
}

@Composable
private fun RiskInsightCard(
    data: RiskOverviewResponse?,
    onOpenDetail: () -> Unit,
) {
    Card(shape = RoundedCornerShape(16.dp), modifier = Modifier.padding(horizontal = 16.dp)) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color(0xFF0F172A))
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text("AI 风险洞察", color = Color(0xFF93C5FD), fontWeight = FontWeight.Bold)
            Text(
                text = data?.analysis?.summary?.ifBlank { "近期暂无风险分析摘要" } ?: "近期暂无风险分析摘要",
                color = Color.White,
                style = MaterialTheme.typography.titleMedium,
            )
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                StatMiniCard("整体趋势", formatRiskDescriptor(data?.analysis?.overall_trend), Modifier.weight(1f))
                StatMiniCard("高危信号", formatRiskDescriptor(data?.analysis?.high_risk_trend, high = true), Modifier.weight(1f))
            }
            TextButton(onClick = onOpenDetail) {
                Text("查看详情", color = Color.White)
            }
        }
    }
}

@Composable
private fun RiskBadge(level: String) {
    val label = normalizeRiskLabel(level)
    val background = when (label) {
        "高" -> Color(0xFFFEE2E2)
        "低" -> Color(0xFFDCFCE7)
        else -> Color(0xFFFEF3C7)
    }
    val foreground = when (label) {
        "高" -> Color(0xFFB91C1C)
        "低" -> Color(0xFF15803D)
        else -> Color(0xFFB45309)
    }
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(999.dp))
            .background(background)
            .padding(horizontal = 10.dp, vertical = 4.dp),
    ) {
        Text(label, color = foreground, fontWeight = FontWeight.Bold, fontSize = 12.sp)
    }
}

private data class PersonalAlertTheme(
    val accent: Color,
    val soft: Color,
    val surface: Color,
    val action: Color,
)

private fun personalAlertTheme(level: String): PersonalAlertTheme {
    return when (normalizePersonalAlertRiskLabel(level)) {
        "高" -> PersonalAlertTheme(
            accent = Color(0xFFDC2626),
            soft = Color(0xFFFEE2E2),
            surface = Color(0xFFFEF2F2),
            action = Color(0xFFDC2626),
        )
        else -> PersonalAlertTheme(
            accent = Color(0xFFD97706),
            soft = Color(0xFFFEF3C7),
            surface = Color(0xFFFFFBEB),
            action = Color(0xFFD97706),
        )
    }
}

private fun isPersonalAlertLevel(level: String): Boolean {
    val normalized = normalizePersonalAlertRiskLabel(level)
    return normalized == "高" || normalized == "中"
}

@Composable
private fun StatusBadge(status: String) {
    val color = when (status) {
        "pending" -> Color(0xFFF59E0B)
        "processing" -> Color(0xFF3B82F6)
        "completed" -> Color(0xFF10B981)
        else -> Color(0xFFEF4444)
    }
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(999.dp))
            .background(color.copy(alpha = 0.14f))
            .padding(horizontal = 10.dp, vertical = 4.dp),
    ) {
        Text(statusLabel(status), color = color, fontWeight = FontWeight.Bold, fontSize = 12.sp)
    }
}

@Composable
private fun SectionCard(
    title: String,
    body: String,
) {
    Card(shape = RoundedCornerShape(20.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
        Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Text(title, fontWeight = FontWeight.Bold)
            Text(body, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
private fun DonutChart(
    stats: RiskStats,
    modifier: Modifier = Modifier,
) {
    val slices = listOf(
        stats.high.toFloat() to Color(0xFFEF4444),
        stats.medium.toFloat() to Color(0xFFF59E0B),
        stats.low.toFloat() to Color(0xFF10B981),
    )
    val total = slices.sumOf { it.first.toDouble() }.toFloat().coerceAtLeast(1f)

    Box(modifier = modifier, contentAlignment = Alignment.Center) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            var start = -90f
            slices.forEach { (value, color) ->
                val sweep = value / total * 360f
                drawArc(
                    color = color,
                    startAngle = start,
                    sweepAngle = sweep,
                    useCenter = false,
                    topLeft = Offset(40f, 40f),
                    size = Size(size.minDimension - 80f, size.minDimension - 80f),
                    style = Stroke(width = 42f, cap = StrokeCap.Round),
                )
                start += sweep
            }
        }
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(stats.total.toString(), style = MaterialTheme.typography.headlineLarge)
            Text("总检测", color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
private fun TrendChart(
    points: List<RiskTrendPoint>,
    modifier: Modifier = Modifier,
) {
    val sorted = points.sortedBy { it.time_bucket }
    val maxValue = sorted.flatMap { listOf(it.high, it.medium, it.low) }.maxOrNull()?.coerceAtLeast(1) ?: 1

    Card(modifier = modifier, shape = RoundedCornerShape(20.dp), colors = CardDefaults.cardColors(containerColor = Color(0xFFF8FAFC))) {
        if (sorted.isEmpty()) {
            EmptyState("暂无趋势数据")
        } else {
            Canvas(modifier = Modifier.fillMaxSize().padding(16.dp)) {
                val spacingX = if (sorted.size == 1) size.width else size.width / (sorted.size - 1)

                fun yOf(value: Int): Float = size.height - (value / maxValue.toFloat()) * (size.height - 20f)

                listOf(
                    sorted.mapIndexed { index, item -> Offset(index * spacingX, yOf(item.high)) } to Color(0xFFEF4444),
                    sorted.mapIndexed { index, item -> Offset(index * spacingX, yOf(item.medium)) } to Color(0xFFF59E0B),
                    sorted.mapIndexed { index, item -> Offset(index * spacingX, yOf(item.low)) } to Color(0xFF10B981),
                ).forEach { (line, color) ->
                    val path = Path()
                    line.forEachIndexed { index, offset ->
                        if (index == 0) path.moveTo(offset.x, offset.y) else path.lineTo(offset.x, offset.y)
                    }
                    drawPath(path = path, color = color, style = Stroke(width = 5f, cap = StrokeCap.Round))
                }
            }
        }
    }
}

@Composable
private fun Base64Thumbnail(
    dataUrl: String,
    size: Dp,
    modifier: Modifier = Modifier,
    bitmapWidth: Dp = size,
    bitmapHeight: Dp = size,
    framed: Boolean = true,
    contentScale: ContentScale = ContentScale.Crop,
    onClick: (() -> Unit)? = null,
) {
    val bitmapWidthPx = with(LocalDensity.current) { bitmapWidth.roundToPx() }
    val bitmapHeightPx = with(LocalDensity.current) { bitmapHeight.roundToPx() }
    val bitmap = remember(dataUrl, bitmapWidthPx, bitmapHeightPx) {
        decodeDataUrlBitmap(dataUrl, bitmapWidthPx, bitmapHeightPx)
    }
    val containerModifier = if (modifier == Modifier) Modifier.size(size) else modifier
    Box(
        modifier = containerModifier
            .then(
                if (framed) {
                    Modifier
                        .clip(RoundedCornerShape(16.dp))
                        .background(Color(0xFFF1F5F9))
                } else {
                    Modifier
                }
            )
            .clickable(enabled = onClick != null) { onClick?.invoke() },
        contentAlignment = Alignment.Center,
    ) {
        if (bitmap != null) {
            Image(
                bitmap = bitmap.asImageBitmap(),
                contentDescription = null,
                modifier = Modifier.fillMaxSize(),
                contentScale = contentScale,
            )
        } else {
            Icon(Icons.Outlined.Image, contentDescription = null, tint = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
private fun BannerCarousel(
    banners: List<Int>,
    currentIndex: Int,
    onSelect: (Int) -> Unit,
) {
    val pagerState = rememberPagerState(initialPage = currentIndex, pageCount = { banners.size })
    val scope = rememberCoroutineScope()

    LaunchedEffect(currentIndex) {
        if (pagerState.currentPage != currentIndex) {
            pagerState.animateScrollToPage(currentIndex)
        }
    }

    Card(shape = RoundedCornerShape(16.dp)) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(160.dp),
        ) {
            HorizontalPager(
                state = pagerState,
                modifier = Modifier.fillMaxSize(),
            ) { page ->
                val banner = banners[page]
                    Image(
                        painter = painterResource(banner),
                        contentDescription = null,
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(160.dp),
                        contentScale = ContentScale.Crop,
                    )
            }

            Row(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 10.dp),
                horizontalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                banners.indices.forEach { index ->
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .clip(CircleShape)
                            .background(if (index == currentIndex) Color.White else Color.White.copy(alpha = 0.5f))
                            .clickable {
                                onSelect(index)
                                scope.launch { pagerState.animateScrollToPage(index) }
                            },
                    )
                }
            }
        }
    }
}

private fun recentAlertCases(state: MainUiState): List<AlertEvent> {
    val fromSocket = state.alertItems.map { it.event }
    val fromHistory = state.history
        .filter { isPersonalAlertLevel(it.risk_level) && recentWithinLastHour(it.created_at) }
        .map {
            AlertEvent(
                record_id = it.record_id,
                title = it.title,
                case_summary = it.case_summary,
                scam_type = it.scam_type,
                risk_level = it.risk_level,
                created_at = it.created_at,
            )
        }

    return (fromSocket + fromHistory)
        .distinctBy { it.record_id }
        .sortedByDescending { parseInstant(it.created_at.ifBlank { it.sent_at }) }
}

private fun realtimeStateLabel(state: RealtimeState): String = when (state) {
    RealtimeState.Connected -> "通道已连接"
    RealtimeState.Connecting -> "通道连接中"
    RealtimeState.Reconnecting -> "通道重连中"
    RealtimeState.Disconnected -> "通道未连接"
}

private fun formatRiskDescriptor(value: String?, high: Boolean = false): String = when (value.orEmpty()) {
    "上升" -> if (high) "高危暴露增强" else "风险热度抬升"
    "下降" -> if (high) "高危暴露收敛" else "风险热度回落"
    "平稳" -> if (high) "高危信号平稳" else "风险热度平稳"
    else -> if (high) "高危信号待观察" else "风险热度待观察"
}

private fun parseReport(text: String): List<ReportSection> {
    val sections = mutableListOf<ReportSection>()
    var currentTitle = ""
    val builder = StringBuilder()
    var currentId = 0

    text.lines().forEach { line ->
        val match = Regex("""^(\d+)\.\s+(.+)$""").matchEntire(line.trim())
        if (match != null) {
            if (currentTitle.isNotBlank()) {
                sections += ReportSection(currentId, currentTitle, builder.toString().trim())
            }
            currentId = match.groupValues[1].toIntOrNull() ?: currentId + 1
            currentTitle = match.groupValues[2]
            builder.clear()
        } else {
            builder.appendLine(line)
        }
    }

    if (currentTitle.isNotBlank()) {
        sections += ReportSection(currentId, currentTitle, builder.toString().trim())
    }

    return sections
}

private fun extractAttackSteps(text: String): List<String> {
    return parseReport(text)
        .firstOrNull { it.title.contains("链路") }
        ?.content
        ?.lines()
        ?.map { it.trim().trimStart('-', '*', '•') }
        ?.filter { it.isNotBlank() }
        .orEmpty()
}

private fun extractKeywordSentences(text: String): List<String> {
    return parseReport(text)
        .firstOrNull { it.title.contains("关键词") }
        ?.content
        ?.lines()
        ?.map { it.trim().trimStart('-', '*', '•') }
        ?.filter { it.isNotBlank() }
        .orEmpty()
}

private fun statusLabel(status: String): String = when (status) {
    "pending" -> "等待中"
    "processing" -> "分析中"
    "completed" -> "已完成"
    "failed" -> "失败"
    else -> status
}

private fun normalizeRiskLabel(value: String): String = when {
    value.contains("高") -> "高"
    value.contains("低") -> "低"
    else -> "中"
}

private fun brandBrush(): Brush = Brush.linearGradient(listOf(Color(0xFF10B981), Color(0xFF0F766E)))

private fun buildExportFilename(task: TaskDetail, extension: String): String {
    val date = SimpleDateFormat("yyyy-MM-dd", Locale.getDefault()).format(Date())
    return "scam-report-${task.task_id}-$date.$extension"
}

private fun buildTaskMarkdown(task: TaskDetail): String {
    val builder = StringBuilder()
    builder.append("# 诈骗风险分析报告\n\n")
    builder.append("**任务ID**: ${task.task_id}\n")
    builder.append("**标题**: ${task.title.ifBlank { "未命名任务" }}\n")
    builder.append("**诈骗类型**: ${task.scam_type.ifBlank { "未识别" }}\n")
    builder.append("**风险分数**: ${task.risk_score}\n")
    builder.append("**生成时间**: ${formatDateTime(task.created_at)}\n")
    builder.append("**状态**: ${statusLabel(task.status)}\n\n")
    if (task.risk_summary.isNotBlank()) {
        builder.append("## 风险结构化摘要\n")
        builder.append(task.risk_summary)
        builder.append("\n\n")
    }
    if (task.report.isNotBlank()) {
        builder.append("## 综合分析报告\n")
        builder.append(task.report)
        builder.append("\n\n")
    }
    if (task.payload.video_insights.isNotEmpty()) {
        builder.append("## 视频分析洞察\n")
        task.payload.video_insights.forEachIndexed { index, insight ->
            builder.append("### 视频 #${index + 1}\n")
            builder.append(insight)
            builder.append("\n\n")
        }
    }
    if (task.payload.audio_insights.isNotEmpty()) {
        builder.append("## 音频分析洞察\n")
        task.payload.audio_insights.forEachIndexed { index, insight ->
            builder.append("### 音频 #${index + 1}\n")
            builder.append(insight)
            builder.append("\n\n")
        }
    }
    if (task.payload.image_insights.isNotEmpty()) {
        builder.append("## 图片分析洞察\n")
        task.payload.image_insights.forEachIndexed { index, insight ->
            builder.append("### 图片 #${index + 1}\n")
            builder.append(insight)
            builder.append("\n\n")
        }
    }
    if (task.payload.text.isNotBlank()) {
        builder.append("## 原始文本证据\n")
        builder.append(task.payload.text)
        builder.append("\n")
    }
    return builder.toString()
}

private data class ParsedRiskSummary(
    val dimensions: Map<String, Int>,
    val hitRules: List<String>,
)

@Composable
private fun RiskScoreLine(label: String, value: Int) {
    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, color = Color(0xFF475569), fontSize = 13.sp)
        Text(value.toString(), color = Color(0xFF0F172A), fontSize = 13.sp, fontWeight = FontWeight.Bold)
    }
}

private fun parseRiskSummary(raw: String): ParsedRiskSummary? {
    if (raw.isBlank()) return null
    return runCatching {
        val root = Json.parseToJsonElement(raw).jsonObject
        val dimensions = root["dimensions"]?.jsonObject?.mapValues { (_, value) ->
            value.toString().trim().trim('"').toIntOrNull() ?: 0
        } ?: emptyMap()
        val hitRules = root["hit_rules"]?.jsonArray?.mapNotNull { element ->
            element.toString().trim().trim('"').takeIf { it.isNotBlank() }
        } ?: emptyList()
        ParsedRiskSummary(dimensions = dimensions, hitRules = hitRules)
    }.getOrNull()
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
        val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, values)
            ?: error("无法创建导出文件")
        resolver.openOutputStream(uri)?.bufferedWriter(Charsets.UTF_8)?.use { writer ->
            writer.write(content)
        } ?: error("无法写入导出文件")
        onMessage("已保存到下载目录", false)
    }.onFailure {
        onMessage(it.message ?: "导出失败", true)
    }
}

private fun todayHistoryCount(history: List<HistoryRecord>): Int {
    val formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd").withZone(ZoneId.systemDefault())
    val today = formatter.format(Instant.now())
    return history.count { formatter.format(parseInstant(it.created_at)) == today }
}

private fun todayHighRiskCount(history: List<HistoryRecord>): Int {
    val formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd").withZone(ZoneId.systemDefault())
    val today = formatter.format(Instant.now())
    return history.count {
        formatter.format(parseInstant(it.created_at)) == today && normalizeRiskLabel(it.risk_level) == "高"
    }
}

private fun decodeDataUrlBitmap(dataUrl: String, widthPx: Int, heightPx: Int) = runCatching {
    when {
        dataUrl.startsWith("data:image/svg+xml", ignoreCase = true) -> decodeSvgBitmap(dataUrl, widthPx, heightPx)
        else -> decodeRasterBitmap(dataUrl)
    }
}.getOrNull()

private fun decodeRasterBitmap(dataUrl: String): Bitmap? {
    val payload = dataUrl.substringAfter("base64,", "")
    if (payload.isBlank()) return null
    val bytes = Base64.getDecoder().decode(payload)
    return BitmapFactory.decodeByteArray(bytes, 0, bytes.size)
}

private fun decodeSvgBitmap(dataUrl: String, widthPx: Int, heightPx: Int): Bitmap? {
    val header = dataUrl.substringBefore(',', "")
    val payload = dataUrl.substringAfter(',', "")
    if (payload.isBlank()) return null

    val svgText = if (header.contains(";base64", ignoreCase = true)) {
        String(Base64.getDecoder().decode(payload), StandardCharsets.UTF_8)
    } else {
        URLDecoder.decode(payload, StandardCharsets.UTF_8.name())
    }

    val svg = SVG.getFromString(svgText)
    val safeWidth = widthPx.coerceAtLeast(1)
    val safeHeight = heightPx.coerceAtLeast(1)
    val bitmap = Bitmap.createBitmap(safeWidth, safeHeight, Bitmap.Config.ARGB_8888)
    val canvas = AndroidCanvas(bitmap)
    svg.setDocumentWidth("100%")
    svg.setDocumentHeight("100%")
    svg.renderToCanvas(canvas)
    return bitmap
}
