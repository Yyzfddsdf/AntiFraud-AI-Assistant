package com.example.myapplication

import android.app.Application
import android.Manifest
import android.content.pm.PackageManager
import android.location.Location
import android.location.LocationManager
import android.content.Context
import android.net.Uri
import android.util.Base64
import android.webkit.MimeTypeMap
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import java.io.IOException
import java.time.Duration
import java.time.Instant
import java.time.OffsetDateTime
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.Locale
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonObject
import okhttp3.WebSocket
import androidx.core.content.ContextCompat

data class AuthFormState(
    val account: String = "",
    val username: String = "",
    val email: String = "",
    val phone: String = "",
    val password: String = "",
    val captchaCode: String = "",
    val smsCode: String = "",
)

private data class RegionResolvePreparation(
    val request: ResolveRegionRequest,
    val usedOverseasFallback: Boolean = false,
    val detectedCountry: String = "",
)

data class FamilyInviteFormState(
    val inviteeEmail: String = "",
    val inviteePhone: String = "",
    val role: String = "member",
    val relation: String = "",
)

data class FamilyGuardianFormState(
    val guardianUserId: String = "",
    val memberUserId: String = "",
)

data class AnalyzeFormState(
    val text: String = "",
    val videos: List<String> = emptyList(),
    val audios: List<String> = emptyList(),
    val images: List<String> = emptyList(),
)

data class ProfileFormState(
    val age: Int = 28,
    val occupation: String = "",
    val provinceCode: String = "",
    val provinceName: String = "",
    val cityCode: String = "",
    val cityName: String = "",
    val districtCode: String = "",
    val districtName: String = "",
    val locationSource: String = "manual",
)

data class SimulationFormState(
    val caseType: String = "",
    val targetPersona: String = "",
    val difficulty: String = "easy",
    val locale: String = "zh-CN",
)

enum class SimulationViewMode {
    Overview,
    Exam,
}

data class AlertMessageItem(
    val event: AlertEvent,
    val read: Boolean = false,
)

data class MainUiState(
    val token: String = "",
    val isAuthenticated: Boolean = false,
    val user: UserProfile = UserProfile(),
    val authMode: AuthMode = AuthMode.Login,
    val loginMethod: LoginMethod = LoginMethod.Password,
    val screen: AppScreen = AppScreen.Dashboard,
    val loading: Boolean = false,
    val analyzing: Boolean = false,
    val chatLoading: Boolean = false,
    val isChatting: Boolean = false,
    val authForm: AuthFormState = AuthFormState(),
    val captchaImage: String = "",
    val captchaId: String = "",
    val smsCooldownSeconds: Int = 0,
    val ageEditorVisible: Boolean = false,
    val profileSaving: Boolean = false,
    val locationResolving: Boolean = false,
    val ageInput: String = "28",
    val occupationInput: String = "",
    val recentTagsInput: String = "",
    val occupationOptions: List<String> = emptyList(),
    val profileForm: ProfileFormState = ProfileFormState(),
    val provinceOptions: List<RegionOption> = emptyList(),
    val cityOptions: List<RegionOption> = emptyList(),
    val districtOptions: List<RegionOption> = emptyList(),
    val analyzeForm: AnalyzeFormState = AnalyzeFormState(),
    val tasks: List<TaskSummary> = emptyList(),
    val history: List<HistoryRecord> = emptyList(),
    val riskInterval: String = "day",
    val riskData: RiskOverviewResponse? = null,
    val regionCaseStats: CurrentRegionCaseStatsResponse? = null,
    val selectedTask: TaskDetail? = null,
    val alertItems: List<AlertMessageItem> = emptyList(),
    val activeAlert: AlertEvent? = null,
    val alertConnectionState: RealtimeState = RealtimeState.Disconnected,
    val familyOverview: FamilyOverviewResponse = FamilyOverviewResponse(),
    val familyReceivedInvitations: List<FamilyInvitation> = emptyList(),
    val familyNotifications: List<FamilyNotification> = emptyList(),
    val activeFamilyNotification: FamilyNotification? = null,
    val familyNotificationState: RealtimeState = RealtimeState.Disconnected,
    val familyCreateName: String = "",
    val familyInviteForm: FamilyInviteFormState = FamilyInviteFormState(),
    val familyAcceptCode: String = "",
    val familyGuardianForm: FamilyGuardianFormState = FamilyGuardianFormState(),
    val deletingHistoryIds: Set<String> = emptySet(),
    val deletingFamilyMemberIds: Set<Int> = emptySet(),
    val deletingGuardianLinkIds: Set<Int> = emptySet(),
    val acceptingInvitationIds: Set<Int> = emptySet(),
    val markingNotificationIds: Set<Int> = emptySet(),
    val chatMessages: List<DisplayChatMessage> = listOf(
        DisplayChatMessage(type = "ai", content = "你好，我是你的反诈智能助手。我可以帮你分析风险、总结近期安全情况，或解答防诈问题。"),
    ),
    val chatInput: String = "",
    val chatImages: List<String> = emptyList(),
    val chatHistoryLoaded: Boolean = false,
    val simulationGenerating: Boolean = false,
    val simulationSubmitting: Boolean = false,
    val simulationForm: SimulationFormState = SimulationFormState(),
    val simulationPackList: List<SimulationPack> = emptyList(),
    val simulationSessionList: List<SimulationSessionItem> = emptyList(),
    val simulationPack: SimulationPack? = null,
    val simulationPackId: String = "",
    val simulationCurrentStep: SimulationStep? = null,
    val simulationCurrentScore: Int = 60,
    val simulationStatus: String = "idle",
    val simulationViewMode: SimulationViewMode = SimulationViewMode.Overview,
    val simulationAnswers: List<SimulationSessionAnswer> = emptyList(),
    val simulationResult: SimulationResult? = null,
    val deletingSimulationSessionIds: Set<String> = emptySet(),
    val quickAnalyzeBubbleEnabled: Boolean = false,
    val accessibilityAutoAnalyzeEnabled: Boolean = false,
    val accessibilityAutoAnalyzePermissionGranted: Boolean = false,
)

class SentinelViewModel(
    application: Application,
) : AndroidViewModel(application) {
    companion object {
        private const val FALLBACK_PROVINCE = "浙江省"
        private const val FALLBACK_CITY = "杭州市"
        private const val FALLBACK_DISTRICT = "钱塘区"
    }

    private val appConfig = AppConfigLoader.load(application)
    private val repository = SentinelRepository(appConfig.apiOrigin)
    private val preferences = application.getSharedPreferences("sentinel", Context.MODE_PRIVATE)
    private val deduplicator = AlertDeduplicator(application)
    private val initialQuickAnalyzeBubbleEnabled = preferences
        .getBoolean(QuickAnalyzeOverlayService.PREF_ENABLED, false)
        .let { enabled ->
            if (enabled && !QuickAnalyzeOverlayService.isRunning()) {
                preferences.edit().putBoolean(QuickAnalyzeOverlayService.PREF_ENABLED, false).apply()
                false
            } else {
                enabled
            }
        }
    private val initialAccessibilityAutoAnalyzeEnabled =
        preferences.getBoolean(AccessibilityAutoAnalyzeService.PREF_ENABLED, false)
    private val initialAccessibilityAutoAnalyzePermissionGranted =
        AccessibilityAutoAnalyzeService.isEnabledInSystem(application)

    var uiState by mutableStateOf(
        MainUiState(
            token = preferences.getString("token", "").orEmpty(),
            screen = appConfig.defaultScreen.toAppScreen(),
            quickAnalyzeBubbleEnabled = initialQuickAnalyzeBubbleEnabled,
            accessibilityAutoAnalyzeEnabled = initialAccessibilityAutoAnalyzeEnabled,
            accessibilityAutoAnalyzePermissionGranted = initialAccessibilityAutoAnalyzePermissionGranted,
        ),
    )
        private set

    var latestMessage by mutableStateOf<UiMessage?>(null)
        private set

    private var pollingJob: Job? = null
    private var smsCooldownJob: Job? = null
    private var alertSocket: WebSocket? = null
    private var familySocket: WebSocket? = null
    private var chatStream: AutoCloseable? = null
    private var alertReconnectJob: Job? = null
    private var familyReconnectJob: Job? = null
    private var simulationGenerationPollingJob: Job? = null
    private var alertReconnectAttempts = 0
    private var familyReconnectAttempts = 0

    init {
        fetchCaptcha()
        if (uiState.token.isNotBlank()) {
            restoreSession()
        }
    }

    override fun onCleared() {
        super.onCleared()
        stopRealtime()
        simulationGenerationPollingJob?.cancel()
        chatStream?.close()
        repository.close()
    }

    fun updateAuthMode(mode: AuthMode) {
        uiState = uiState.copy(authMode = mode)
        if (requiresGraphCaptcha()) {
            fetchCaptcha()
        }
    }

    fun updateLoginMethod(method: LoginMethod) {
        uiState = uiState.copy(loginMethod = method)
        if (requiresGraphCaptcha()) {
            fetchCaptcha()
        }
    }

    fun updateAuthAccount(value: String) = mutateAuthForm { copy(account = value) }
    fun updateAuthUsername(value: String) = mutateAuthForm { copy(username = value) }
    fun updateAuthEmail(value: String) = mutateAuthForm { copy(email = value) }
    fun updateAuthPhone(value: String) = mutateAuthForm { copy(phone = value) }
    fun updateAuthPassword(value: String) = mutateAuthForm { copy(password = value) }
    fun updateAuthCaptchaCode(value: String) = mutateAuthForm { copy(captchaCode = value) }
    fun updateAuthSmsCode(value: String) = mutateAuthForm { copy(smsCode = value) }
    fun updateAgeInput(value: String) {
        uiState = uiState.copy(
            ageInput = value,
            profileForm = uiState.profileForm.copy(age = value.toIntOrNull()?.coerceIn(1, 150) ?: uiState.profileForm.age),
        )
    }
    fun updateOccupationInput(value: String) {
        uiState = uiState.copy(
            occupationInput = value,
            profileForm = uiState.profileForm.copy(occupation = value),
        )
    }
    fun updateRecentTagsInput(value: String) { uiState = uiState.copy(recentTagsInput = value) }
    fun updateProfileAge(value: Int) {
        val normalized = value.coerceIn(1, 150)
        uiState = uiState.copy(
            ageInput = normalized.toString(),
            profileForm = uiState.profileForm.copy(age = normalized),
        )
    }
    fun updateProfileOccupation(value: String) {
        uiState = uiState.copy(
            occupationInput = value,
            profileForm = uiState.profileForm.copy(occupation = value),
        )
    }
    fun updateSimulationCaseType(value: String) {
        uiState = uiState.copy(simulationForm = uiState.simulationForm.copy(caseType = value))
    }
    fun updateSimulationTargetPersona(value: String) {
        uiState = uiState.copy(simulationForm = uiState.simulationForm.copy(targetPersona = value))
    }
    fun updateSimulationDifficulty(value: String) {
        uiState = uiState.copy(simulationForm = uiState.simulationForm.copy(difficulty = value))
    }
    fun updateAnalyzeText(value: String) { uiState = uiState.copy(analyzeForm = uiState.analyzeForm.copy(text = value)) }
    fun updateRiskInterval(value: String) { uiState = uiState.copy(riskInterval = value) }
    fun updateFamilyCreateName(value: String) { uiState = uiState.copy(familyCreateName = value) }
    fun updateFamilyInviteEmail(value: String) { uiState = uiState.copy(familyInviteForm = uiState.familyInviteForm.copy(inviteeEmail = value)) }
    fun updateFamilyInvitePhone(value: String) { uiState = uiState.copy(familyInviteForm = uiState.familyInviteForm.copy(inviteePhone = value)) }
    fun updateFamilyInviteRelation(value: String) { uiState = uiState.copy(familyInviteForm = uiState.familyInviteForm.copy(relation = value)) }
    fun updateFamilyInviteRole(value: String) { uiState = uiState.copy(familyInviteForm = uiState.familyInviteForm.copy(role = value)) }
    fun updateFamilyAcceptCode(value: String) { uiState = uiState.copy(familyAcceptCode = value) }
    fun updateGuardianUser(value: String) { uiState = uiState.copy(familyGuardianForm = uiState.familyGuardianForm.copy(guardianUserId = value)) }
    fun updateProtectedUser(value: String) { uiState = uiState.copy(familyGuardianForm = uiState.familyGuardianForm.copy(memberUserId = value)) }
    fun updateChatInput(value: String) { uiState = uiState.copy(chatInput = value) }
    fun toggleAgeEditor() { uiState = uiState.copy(ageEditorVisible = !uiState.ageEditorVisible) }
    fun cancelAgeEditor() {
        uiState = uiState.copy(ageEditorVisible = false, profileForm = uiState.profileForm.copy())
        syncUserProfile(uiState.user, uiState.occupationOptions, editorVisible = false)
    }
    fun closeTaskDetail() { uiState = uiState.copy(selectedTask = null) }
    fun openSimulationOverview() {
        uiState = uiState.copy(screen = AppScreen.SimulationQuiz, simulationViewMode = SimulationViewMode.Overview)
    }
    fun closeSimulationExamView() {
        uiState = uiState.copy(simulationViewMode = SimulationViewMode.Overview)
    }
    fun setQuickAnalyzeBubbleEnabled(enabled: Boolean) {
        preferences.edit().putBoolean(QuickAnalyzeOverlayService.PREF_ENABLED, enabled).apply()
        uiState = uiState.copy(quickAnalyzeBubbleEnabled = enabled)
    }
    fun setAccessibilityAutoAnalyzeEnabled(enabled: Boolean) {
        preferences.edit().putBoolean(AccessibilityAutoAnalyzeService.PREF_ENABLED, enabled).apply()
        uiState = uiState.copy(accessibilityAutoAnalyzeEnabled = enabled)
    }
    fun setAccessibilityAutoAnalyzePermissionGranted(granted: Boolean) {
        uiState = uiState.copy(accessibilityAutoAnalyzePermissionGranted = granted)
    }

    fun dismissAlert() {
        uiState.activeAlert?.record_id?.let(::markAlertRead)
        uiState = uiState.copy(activeAlert = null)
    }

    fun dismissFamilyAlert() {
        uiState.activeFamilyNotification?.id?.let(::markFamilyNotificationRead)
        uiState = uiState.copy(activeFamilyNotification = null)
    }

    fun openScreen(screen: AppScreen) {
        uiState = when (screen) {
            AppScreen.Profile -> uiState.copy(screen = screen, ageEditorVisible = false)
            AppScreen.SimulationQuiz -> uiState.copy(screen = screen, simulationViewMode = SimulationViewMode.Overview)
            else -> uiState.copy(screen = screen)
        }
    }

    fun fetchCaptcha() {
        viewModelScope.launch {
            runRequest(silent = true) {
                val response = repository.fetchCaptcha()
                uiState = uiState.copy(captchaId = response.captchaId, captchaImage = response.captchaImage)
            }
        }
    }

    fun sendSmsCode() {
        val phone = uiState.authForm.phone.trim()
        if (phone.isBlank()) {
            showMessage("请输入手机号")
            return
        }
        if (uiState.smsCooldownSeconds > 0) return
        viewModelScope.launch {
            runRequest {
                val response = repository.sendSmsCode(phone)
                showMessage(response.message.ifBlank { "短信验证码已发送，请使用 000000" })
                startSmsCooldown()
            }
        }
    }

    fun submitAuth() {
        uiState = uiState.copy(loading = true)
        viewModelScope.launch {
            try {
                when (uiState.authMode) {
                    AuthMode.Register -> {
                        repository.register(
                            AuthRegisterRequest(
                                username = uiState.authForm.username.trim(),
                                email = uiState.authForm.email.trim(),
                                phone = uiState.authForm.phone.trim(),
                                password = uiState.authForm.password,
                                captchaId = uiState.captchaId,
                                captchaCode = uiState.authForm.captchaCode.trim(),
                                smsCode = uiState.authForm.smsCode.trim(),
                            ),
                        )
                        uiState = uiState.copy(
                            loading = false,
                            authMode = AuthMode.Login,
                            loginMethod = LoginMethod.Password,
                            authForm = uiState.authForm.copy(
                                account = uiState.authForm.email.trim(),
                                password = "",
                                captchaCode = "",
                                smsCode = "",
                            ),
                        )
                        showMessage("注册成功，请登录")
                        fetchCaptcha()
                    }

                    AuthMode.Login -> {
                        val response = when (uiState.loginMethod) {
                            LoginMethod.Password -> repository.loginWithPassword(
                                AuthPasswordLoginRequest(
                                    account = uiState.authForm.account.trim(),
                                    password = uiState.authForm.password,
                                    captchaId = uiState.captchaId,
                                    captchaCode = uiState.authForm.captchaCode.trim(),
                                ),
                            )

                            LoginMethod.Sms -> repository.loginWithSms(
                                AuthSmsLoginRequest(
                                    phone = uiState.authForm.phone.trim(),
                                    smsCode = uiState.authForm.smsCode.trim(),
                                ),
                            )
                        }
                        startSession(response.token, response.user)
                        showMessage("登录成功")
                    }
                }
            } catch (exception: ApiException) {
                uiState = uiState.copy(loading = false)
                showMessage(exception.message)
                if (requiresGraphCaptcha()) {
                    fetchCaptcha()
                }
            }
        }
    }

    fun logout() {
        preferences.edit().remove("token").apply()
        preferences.edit().putBoolean(BackgroundAlertService.PREF_HAS_FAMILY, false).apply()
        preferences.edit().putBoolean(QuickAnalyzeOverlayService.PREF_ENABLED, false).apply()
        preferences.edit().putBoolean(AccessibilityAutoAnalyzeService.PREF_ENABLED, false).apply()
        BackgroundAlertService.stop(getApplication())
        QuickAnalyzeOverlayService.stop(getApplication())
        simulationGenerationPollingJob?.cancel()
        stopRealtime()
        chatStream?.close()
        uiState = MainUiState()
        fetchCaptcha()
    }

    fun refreshUser() {
        restoreSession()
    }

    fun updateAge() {
        updateUserProfile()
    }

    fun updateUserProfile() {
        val age = uiState.profileForm.age
        if (age !in 1..150) {
            showMessage("年龄请输入 1 到 150 之间的数字")
            return
        }
        withToken { token ->
            uiState = uiState.copy(profileSaving = true)
            viewModelScope.launch {
                try {
                    val response = repository.updateUserProfile(
                        token,
                        UpdateUserProfileRequest(
                            age = age,
                            occupation = uiState.profileForm.occupation.trim(),
                            province_code = uiState.profileForm.provinceCode.trim(),
                            province_name = uiState.profileForm.provinceName.trim(),
                            city_code = uiState.profileForm.cityCode.trim(),
                            city_name = uiState.profileForm.cityName.trim(),
                            district_code = uiState.profileForm.districtCode.trim(),
                            district_name = uiState.profileForm.districtName.trim(),
                            location_source = uiState.profileForm.locationSource.trim().ifBlank { "manual" },
                        ),
                    )
                    syncUserProfile(response.user, uiState.occupationOptions, editorVisible = false)
                    fetchCurrentRegionCaseStatsIfConfigured(silent = true)
                    showMessage(response.message.ifBlank { "用户画像更新成功" })
                } catch (exception: ApiException) {
                    showMessage(exception.message)
                } finally {
                    uiState = uiState.copy(profileSaving = false)
                }
            }
        }
    }

    fun deleteAccount() {
        withToken { token ->
            viewModelScope.launch {
                runRequest {
                    repository.deleteUser(token)
                    showMessage("账户已删除")
                    logout()
                }
            }
        }
    }

    fun addAnalyzeAssets(kind: AnalyzeAssetKind, uris: List<Uri>) {
        if (uris.isEmpty()) return
        viewModelScope.launch {
            val encoded = uris.mapNotNull { uriToDataUrl(it) }
            val next = when (kind) {
                AnalyzeAssetKind.Images -> uiState.analyzeForm.copy(images = uiState.analyzeForm.images + encoded)
                AnalyzeAssetKind.Audios -> uiState.analyzeForm.copy(audios = uiState.analyzeForm.audios + encoded)
                AnalyzeAssetKind.Videos -> uiState.analyzeForm.copy(videos = uiState.analyzeForm.videos + encoded)
            }
            uiState = uiState.copy(analyzeForm = next)
            showMessage("已添加 ${encoded.size} 个文件")
        }
    }

    fun submitAnalysis() {
        val form = uiState.analyzeForm
        if (form.text.isBlank() && form.images.isEmpty() && form.audios.isEmpty() && form.videos.isEmpty()) {
            showMessage("请至少提供一种输入")
            return
        }
        withToken { token ->
            uiState = uiState.copy(analyzing = true)
            viewModelScope.launch {
                try {
                    repository.analyze(
                        token = token,
                        request = AnalyzeRequest(
                            text = form.text,
                            images = form.images,
                            audios = form.audios,
                            videos = form.videos,
                        ),
                    )
                    uiState = uiState.copy(
                        analyzing = false,
                        analyzeForm = AnalyzeFormState(),
                        screen = AppScreen.Dashboard,
                    )
                    showMessage("任务已提交")
                    fetchTasks(silent = true)
                } catch (exception: ApiException) {
                    uiState = uiState.copy(analyzing = false)
                    showMessage(exception.message)
                }
            }
        }
    }

    fun fetchTasks(silent: Boolean) {
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val response = repository.fetchTasks(token)
                    uiState = uiState.copy(tasks = response.tasks.sortedByDescending { parseInstant(it.created_at) })
                }
            }
        }
    }

    fun fetchHistory(silent: Boolean) {
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val response = repository.fetchHistory(token)
                    uiState = uiState.copy(history = response.history.sortedByDescending { parseInstant(it.created_at) })
                }
            }
        }
    }

    fun fetchRiskOverview(silent: Boolean) {
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val response = repository.fetchRiskOverview(token, uiState.riskInterval)
                    uiState = uiState.copy(riskData = response)
                }
            }
        }
    }

    fun fetchCurrentRegionCaseStatsIfConfigured(silent: Boolean) {
        withToken { token ->
            if (uiState.profileForm.provinceCode.isBlank() && uiState.user.province_code.isBlank()) {
                uiState = uiState.copy(regionCaseStats = null)
                return@withToken
            }
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val response = repository.fetchCurrentRegionCaseStats(token)
                    uiState = uiState.copy(regionCaseStats = response)
                }
            }
        }
    }

    fun fetchProvinceOptions(silent: Boolean = true) {
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val response = repository.fetchProvinceOptions(token)
                    uiState = uiState.copy(provinceOptions = response.provinces)
                }
            }
        }
    }

    fun fetchCityOptions(provinceCode: String, silent: Boolean = true) {
        val normalized = provinceCode.trim()
        if (normalized.isBlank()) {
            uiState = uiState.copy(cityOptions = emptyList(), districtOptions = emptyList())
            return
        }
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val response = repository.fetchCityOptions(token, normalized)
                    uiState = uiState.copy(cityOptions = response.cities)
                }
            }
        }
    }

    fun fetchDistrictOptions(cityCode: String, silent: Boolean = true) {
        val normalized = cityCode.trim()
        if (normalized.isBlank()) {
            uiState = uiState.copy(districtOptions = emptyList())
            return
        }
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val response = repository.fetchDistrictOptions(token, normalized)
                    uiState = uiState.copy(districtOptions = response.districts)
                }
            }
        }
    }

    fun selectProvinceValue(code: String) {
        val normalized = code.trim()
        val selected = uiState.provinceOptions.firstOrNull { it.code == normalized }
        uiState = uiState.copy(
            profileForm = uiState.profileForm.copy(
                provinceCode = normalized,
                provinceName = selected?.name.orEmpty(),
                cityCode = "",
                cityName = "",
                districtCode = "",
                districtName = "",
                locationSource = "manual",
            ),
            cityOptions = emptyList(),
            districtOptions = emptyList(),
        )
        fetchCityOptions(normalized)
    }

    fun selectCityValue(code: String) {
        val normalized = code.trim()
        val selected = uiState.cityOptions.firstOrNull { it.code == normalized }
        uiState = uiState.copy(
            profileForm = uiState.profileForm.copy(
                cityCode = normalized,
                cityName = selected?.name.orEmpty(),
                districtCode = "",
                districtName = "",
                locationSource = "manual",
            ),
            districtOptions = emptyList(),
        )
        fetchDistrictOptions(normalized)
    }

    fun selectDistrictValue(code: String) {
        val normalized = code.trim()
        val selected = uiState.districtOptions.firstOrNull { it.code == normalized }
        uiState = uiState.copy(
            profileForm = uiState.profileForm.copy(
                districtCode = normalized,
                districtName = selected?.name.orEmpty(),
                locationSource = "manual",
            ),
        )
    }

    fun requestCurrentRegion() {
        withToken { token ->
            if (!hasAnyLocationPermission()) {
                showMessage("请先授予定位权限", isError = true)
                return@withToken
            }
            uiState = uiState.copy(locationResolving = true)
            viewModelScope.launch {
                try {
                    val location = withContext(Dispatchers.IO) { getBestLastKnownLocation() }
                        ?: throw ApiException(0, "未获取到可用定位，请稍后重试")
                    val resolvePreparation = buildResolveRegionRequest(location)
                    val request = resolvePreparation.request
                    val response = repository.resolveRegion(token, request)
                    val region = response.region ?: throw ApiException(0, "当前位置未匹配到标准行政区")
                    applyResolvedRegion(region, source = "auto")
                    fetchCurrentRegionCaseStatsIfConfigured(silent = true)
                    if (resolvePreparation.usedOverseasFallback) {
                        showMessage(
                            "检测到当前位置不在国内（${resolvePreparation.detectedCountry.ifBlank { "海外/未知地区" }}），已默认切换到杭州钱塘区",
                        )
                    } else {
                        showMessage("已自动识别当前位置")
                    }
                } catch (exception: ApiException) {
                    showMessage(exception.message, isError = true)
                } catch (exception: Exception) {
                    showMessage("当前位置解析失败", isError = true)
                } finally {
                    uiState = uiState.copy(locationResolving = false)
                }
            }
        }
    }

    fun fetchSimulationPacks(silent: Boolean = true) {
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val packs = loadSimulationPacks(token)
                    uiState = uiState.copy(simulationPackList = packs)
                }
            }
        }
    }

    fun fetchSimulationSessions(silent: Boolean = true) {
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val sessions = loadSimulationSessions(token)
                    uiState = uiState.copy(simulationSessionList = sessions)
                }
            }
        }
    }

    fun deleteSimulationSession(sessionId: String) {
        if (sessionId.isBlank()) return
        withToken { token ->
            uiState = uiState.copy(deletingSimulationSessionIds = uiState.deletingSimulationSessionIds + sessionId)
            viewModelScope.launch {
                try {
                    val response = repository.deleteSimulationSession(token, sessionId)
                    val sessions = loadSimulationSessions(token)
                    val packs = loadSimulationPacks(token)
                    uiState = uiState.copy(
                        deletingSimulationSessionIds = uiState.deletingSimulationSessionIds - sessionId,
                        simulationSessionList = sessions,
                        simulationPackList = packs,
                    )
                    showMessage(response.message.ifBlank { "报告删除成功" })
                } catch (exception: ApiException) {
                    uiState = uiState.copy(deletingSimulationSessionIds = uiState.deletingSimulationSessionIds - sessionId)
                    showMessage(exception.message, isError = true)
                }
            }
        }
    }

    fun generateSimulationPack() {
        withToken { token ->
            if (uiState.simulationGenerating) return@withToken
            uiState = uiState.copy(simulationGenerating = true, screen = AppScreen.SimulationQuiz)
            simulationGenerationPollingJob?.cancel()
            viewModelScope.launch {
                try {
                    val response = repository.generateSimulationPack(
                        token,
                        SimulationGeneratePackRequest(
                            case_type = uiState.simulationForm.caseType.trim(),
                            target_persona = uiState.simulationForm.targetPersona.trim(),
                            difficulty = uiState.simulationForm.difficulty,
                            locale = uiState.simulationForm.locale,
                        ),
                    )
                    showMessage(response.message.ifBlank { "题目生成任务已提交，请稍后查看题目列表" })
                    simulationGenerationPollingJob = launch {
                        repeat(30) { _ ->
                            delay(15_000)
                            val packs = runCatching { loadSimulationPacks(token) }.getOrElse { emptyList() }
                            if (packs.isNotEmpty()) {
                                val firstPack = packs.first()
                                uiState = uiState.copy(
                                    simulationPackList = packs,
                                    simulationPack = firstPack,
                                    simulationPackId = firstPack.pack_id,
                                    simulationStatus = "pack_ready",
                                    simulationCurrentScore = 60,
                                    simulationAnswers = emptyList(),
                                    simulationResult = null,
                                )
                                showMessage("模拟题包生成完成")
                                return@launch
                            }
                        }
                        showMessage("题目仍在生成，请稍后刷新题目列表")
                    }
                    val sessions = loadSimulationSessions(token)
                    uiState = uiState.copy(simulationSessionList = sessions)
                } catch (exception: ApiException) {
                    showMessage(exception.message, isError = true)
                } finally {
                    uiState = uiState.copy(simulationGenerating = false)
                }
            }
        }
    }

    fun startSimulationSession(packIdOverride: String = "") {
        withToken { token ->
            val packId = packIdOverride.takeIf { it.isNotBlank() }
                ?: uiState.simulationPackId.takeIf { it.isNotBlank() }
                ?: uiState.simulationPack?.pack_id
                ?: ""
            if (packId.isBlank()) {
                showMessage("请先生成题包", isError = true)
                return@withToken
            }
            uiState = uiState.copy(simulationSubmitting = true, screen = AppScreen.SimulationQuiz)
            viewModelScope.launch {
                try {
                    val response = runCatching {
                        repository.answerSimulationSession(token, SimulationAnswerRequest(pack_id = packId))
                    }.getOrElse {
                        val ongoing = repository.fetchOngoingSimulation(token, packId)
                        if (ongoing.status == "in_progress") {
                            uiState = uiState.copy(
                                simulationPackId = ongoing.pack_id.takeIf { it.isNotBlank() } ?: packId,
                                simulationPack = ongoing.pack ?: uiState.simulationPack,
                                simulationCurrentStep = ongoing.next_step,
                                simulationCurrentScore = ongoing.current_score.takeIf { score -> score > 0 } ?: uiState.simulationCurrentScore,
                                simulationStatus = "in_progress",
                                simulationViewMode = SimulationViewMode.Exam,
                            )
                            val sessions = loadSimulationSessions(token)
                            uiState = uiState.copy(simulationSessionList = sessions)
                            return@launch
                        }
                        throw it
                    }
                    val pack = response.pack ?: uiState.simulationPack
                    uiState = uiState.copy(
                        simulationPackId = packId,
                        simulationStatus = response.status.ifBlank { "in_progress" },
                        simulationCurrentScore = response.current_score.takeIf { it > 0 } ?: 60,
                        simulationPack = pack,
                        simulationCurrentStep = response.next_step ?: pack?.steps?.firstOrNull(),
                        simulationViewMode = SimulationViewMode.Exam,
                        simulationAnswers = emptyList(),
                        simulationResult = response.result,
                    )
                    val sessions = loadSimulationSessions(token)
                    uiState = uiState.copy(simulationSessionList = sessions)
                    showMessage(response.message.ifBlank { "答题已开始" })
                } catch (exception: ApiException) {
                    showMessage(exception.message, isError = true)
                } finally {
                    uiState = uiState.copy(simulationSubmitting = false)
                }
            }
        }
    }

    fun submitSimulationAnswer(optionKey: String) {
        withToken { token ->
            val packId = uiState.simulationPackId
            val currentStep = uiState.simulationCurrentStep
            if (packId.isBlank() || currentStep == null || uiState.simulationStatus != "in_progress") return@withToken
            uiState = uiState.copy(simulationSubmitting = true)
            viewModelScope.launch {
                try {
                    val response = repository.answerSimulationSession(
                        token,
                        SimulationAnswerRequest(
                            pack_id = packId,
                            step_id = currentStep.step_id,
                            option_key = optionKey,
                        ),
                    )
                    val sessions = loadSimulationSessions(token)
                    val currentSession = sessions.firstOrNull { it.pack_id == packId }
                    uiState = uiState.copy(
                        simulationSessionList = sessions,
                        simulationStatus = response.status.ifBlank { uiState.simulationStatus },
                        simulationCurrentScore = response.current_score.takeIf { it > 0 } ?: uiState.simulationCurrentScore,
                        simulationCurrentStep = response.next_step,
                        simulationResult = currentSession?.result ?: response.result ?: uiState.simulationResult,
                        simulationAnswers = currentSession?.answers ?: uiState.simulationAnswers,
                        simulationPack = currentSession?.pack ?: response.pack ?: uiState.simulationPack,
                        simulationViewMode = SimulationViewMode.Exam,
                    )
                    if (uiState.simulationStatus == "completed") {
                        val packs = loadSimulationPacks(token)
                        uiState = uiState.copy(simulationPackList = packs)
                        showMessage("模拟答题完成")
                    }
                } catch (exception: ApiException) {
                    showMessage(exception.message, isError = true)
                } finally {
                    uiState = uiState.copy(simulationSubmitting = false)
                }
            }
        }
    }

    fun resumeOngoingSimulationSession(packIdOverride: String = "") {
        withToken { token ->
            val packId = packIdOverride.takeIf { it.isNotBlank() }
                ?: uiState.simulationPackId.takeIf { it.isNotBlank() }
                ?: uiState.simulationPack?.pack_id
                ?: ""
            if (packId.isBlank()) return@withToken
            viewModelScope.launch {
                runRequest(silent = true) {
                    val response = repository.fetchOngoingSimulation(token, packId)
                    val sessions = loadSimulationSessions(token)
                    val resolvedPackId = response.pack_id.takeIf { it.isNotBlank() } ?: packId
                    val currentSession = sessions.firstOrNull { it.pack_id == resolvedPackId }
                    uiState = uiState.copy(
                        simulationPackId = resolvedPackId,
                        simulationPack = currentSession?.pack ?: response.pack ?: uiState.simulationPack,
                        simulationStatus = response.status.ifBlank { "in_progress" },
                        simulationCurrentScore = response.current_score.takeIf { it > 0 } ?: uiState.simulationCurrentScore,
                        simulationCurrentStep = response.next_step,
                        simulationResult = currentSession?.result ?: response.result ?: uiState.simulationResult,
                        simulationAnswers = currentSession?.answers ?: uiState.simulationAnswers,
                        simulationSessionList = sessions,
                        simulationViewMode = SimulationViewMode.Exam,
                    )
                }
            }
        }
    }

    fun openTaskDetail(taskId: String) {
        withToken { token ->
            viewModelScope.launch {
                runRequest {
                    val response = repository.fetchTaskDetail(token, taskId)
                    val task = response.task
                    val resolvedRiskLevel = task.risk_level.ifBlank {
                        extractTaskRiskLevel(task.risk_summary)
                            .ifBlank { extractRiskLevelFromReport(task.report) }
                            .ifBlank { extractTaskRiskScore(task.risk_summary)?.let(::inferRiskLevelFromScore).orEmpty() }
                            .ifBlank { inferRiskLevelFromScore(task.risk_score) }
                    }
                    uiState = uiState.copy(
                        selectedTask = if (resolvedRiskLevel.isBlank()) task else task.copy(risk_level = resolvedRiskLevel),
                    )
                }
            }
        }
    }

    fun deleteHistoryRecord(recordId: String) {
        withToken { token ->
            uiState = uiState.copy(deletingHistoryIds = uiState.deletingHistoryIds + recordId)
            viewModelScope.launch {
                try {
                    repository.deleteHistoryRecord(token, recordId)
                    uiState = uiState.copy(
                        deletingHistoryIds = uiState.deletingHistoryIds - recordId,
                        history = uiState.history.filterNot { it.record_id == recordId },
                        selectedTask = uiState.selectedTask?.takeUnless { it.task_id == recordId },
                    )
                    showMessage("历史案件已删除")
                } catch (exception: ApiException) {
                    uiState = uiState.copy(deletingHistoryIds = uiState.deletingHistoryIds - recordId)
                    showMessage(exception.message)
                }
            }
        }
    }

    fun fetchFamilyOverview(silent: Boolean) {
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val hadFamily = hasFamilyGroup()
                    val overview = repository.fetchFamilyOverview(token)
                    val hasFamily = overview.family != null
                    uiState = uiState.copy(
                        familyOverview = overview,
                        familyNotifications = if (hasFamily) uiState.familyNotifications else emptyList(),
                        activeFamilyNotification = if (hasFamily) uiState.activeFamilyNotification else null,
                    )
                    if (hadFamily != hasFamily) {
                        preferences.edit().putBoolean(BackgroundAlertService.PREF_HAS_FAMILY, hasFamily).apply()
                        BackgroundAlertService.refresh(getApplication())
                    }
                    if (hasFamily) connectFamilySocket() else disconnectFamilySocket()
                }
            }
        }
    }

    fun fetchReceivedInvitations(silent: Boolean) {
        withToken { token ->
            viewModelScope.launch {
                runRequest(silent = silent) {
                    val response = repository.fetchReceivedFamilyInvitations(token)
                    uiState = uiState.copy(familyReceivedInvitations = response.invitations)
                }
            }
        }
    }

    fun createFamily() {
        val name = uiState.familyCreateName.trim()
        if (name.isBlank()) {
            showMessage("请输入家庭名称")
            return
        }
        withToken { token ->
            viewModelScope.launch {
                runRequest {
                    val overview = repository.createFamily(token, name)
                    uiState = uiState.copy(
                        familyOverview = overview,
                        familyCreateName = "",
                        familyReceivedInvitations = emptyList(),
                        screen = AppScreen.Family,
                    )
                    connectFamilySocket()
                    preferences.edit().putBoolean(BackgroundAlertService.PREF_HAS_FAMILY, true).apply()
                    BackgroundAlertService.refresh(getApplication())
                    showMessage("家庭创建成功")
                }
            }
        }
    }

    fun createFamilyInvitation() {
        withToken { token ->
            viewModelScope.launch {
                runRequest {
                    val response = repository.createFamilyInvitation(
                        token = token,
                        request = CreateFamilyInvitationRequest(
                            invitee_email = uiState.familyInviteForm.inviteeEmail.trim(),
                            invitee_phone = uiState.familyInviteForm.inviteePhone.trim(),
                            role = uiState.familyInviteForm.role,
                            relation = uiState.familyInviteForm.relation.trim(),
                        ),
                    )
                    if (response.invitation != null) {
                        uiState = uiState.copy(familyInviteForm = FamilyInviteFormState())
                        fetchFamilyOverview(silent = true)
                        showMessage("家庭邀请已创建")
                    }
                }
            }
        }
    }

    fun acceptFamilyInvitation(inviteCode: String = uiState.familyAcceptCode.trim(), invitationId: Int = 0) {
        if (inviteCode.isBlank()) {
            showMessage("请输入家庭邀请码")
            return
        }
        withToken { token ->
            if (invitationId > 0) {
                uiState = uiState.copy(acceptingInvitationIds = uiState.acceptingInvitationIds + invitationId)
            }
            viewModelScope.launch {
                try {
                    val overview = repository.acceptFamilyInvitation(token, inviteCode)
                    uiState = uiState.copy(
                        familyOverview = overview,
                        familyAcceptCode = "",
                        familyReceivedInvitations = emptyList(),
                        acceptingInvitationIds = uiState.acceptingInvitationIds - invitationId,
                        screen = AppScreen.Family,
                    )
                    connectFamilySocket()
                    preferences.edit().putBoolean(BackgroundAlertService.PREF_HAS_FAMILY, true).apply()
                    BackgroundAlertService.refresh(getApplication())
                    showMessage("已加入家庭")
                } catch (exception: ApiException) {
                    uiState = uiState.copy(acceptingInvitationIds = uiState.acceptingInvitationIds - invitationId)
                    showMessage(exception.message)
                }
            }
        }
    }

    fun createGuardianLink() {
        val guardianId = uiState.familyGuardianForm.guardianUserId.toIntOrNull()
        val memberId = uiState.familyGuardianForm.memberUserId.toIntOrNull()
        if (guardianId == null || memberId == null) {
            showMessage("请选择守护人和被守护人")
            return
        }
        withToken { token ->
            viewModelScope.launch {
                runRequest {
                    val response = repository.createGuardianLink(
                        token,
                        CreateGuardianLinkRequest(guardian_user_id = guardianId, member_user_id = memberId),
                    )
                    if (response.guardian_link != null) {
                        uiState = uiState.copy(familyGuardianForm = FamilyGuardianFormState())
                        fetchFamilyOverview(silent = true)
                        showMessage("守护关系已保存")
                    }
                }
            }
        }
    }

    fun deleteFamilyMember(memberId: Int) {
        withToken { token ->
            uiState = uiState.copy(deletingFamilyMemberIds = uiState.deletingFamilyMemberIds + memberId)
            viewModelScope.launch {
                try {
                    repository.deleteFamilyMember(token, memberId)
                    uiState = uiState.copy(deletingFamilyMemberIds = uiState.deletingFamilyMemberIds - memberId)
                    fetchFamilyOverview(silent = true)
                    showMessage("成员已移除")
                } catch (exception: ApiException) {
                    uiState = uiState.copy(deletingFamilyMemberIds = uiState.deletingFamilyMemberIds - memberId)
                    showMessage(exception.message)
                }
            }
        }
    }

    fun deleteGuardianLink(linkId: Int) {
        withToken { token ->
            uiState = uiState.copy(deletingGuardianLinkIds = uiState.deletingGuardianLinkIds + linkId)
            viewModelScope.launch {
                try {
                    repository.deleteGuardianLink(token, linkId)
                    uiState = uiState.copy(deletingGuardianLinkIds = uiState.deletingGuardianLinkIds - linkId)
                    fetchFamilyOverview(silent = true)
                    showMessage("守护关系已删除")
                } catch (exception: ApiException) {
                    uiState = uiState.copy(deletingGuardianLinkIds = uiState.deletingGuardianLinkIds - linkId)
                    showMessage(exception.message)
                }
            }
        }
    }

    fun markFamilyNotificationRead(notificationId: Int) {
        withToken { token ->
            if (uiState.markingNotificationIds.contains(notificationId)) return@withToken
            uiState = uiState.copy(markingNotificationIds = uiState.markingNotificationIds + notificationId)
            viewModelScope.launch {
                try {
                    repository.markFamilyNotificationRead(token, notificationId)
                } catch (_: ApiException) {
                    // Ignore server ack failure for local state.
                } finally {
                    val now = Instant.now().toString()
                    uiState = uiState.copy(
                        markingNotificationIds = uiState.markingNotificationIds - notificationId,
                        familyNotifications = uiState.familyNotifications.map {
                            if (it.id == notificationId) it.copy(read_at = now) else it
                        },
                    )
                }
            }
        }
    }

    data class UiMessage(
        val text: String,
        val isError: Boolean,
        val id: Long,
        val channel: String,
    )
    
    fun addChatImages(uris: List<Uri>) {
        if (uris.isEmpty()) return
        viewModelScope.launch {
            val encoded = uris.mapNotNull { uriToDataUrl(it) }
            uiState = uiState.copy(chatImages = uiState.chatImages + encoded)
            showMessage("已添加 ${encoded.size} 张图片")
        }
    }

    fun removeChatImage(index: Int) {
        uiState = uiState.copy(chatImages = uiState.chatImages.filterIndexed { current, _ -> current != index })
    }

    fun fetchChatHistory() {
        withToken { token ->
            if (uiState.chatLoading) return@withToken
            uiState = uiState.copy(chatLoading = true)
            viewModelScope.launch {
                try {
                    val context = repository.fetchChatContext(token)
                    val history = mutableListOf(
                        DisplayChatMessage(type = "ai", content = "你好，我是你的反诈智能助手。我可以帮你分析风险、总结近期安全情况，或解答防诈问题。"),
                    )
                    context.messages.forEach { message ->
                        when (message.role) {
                            "assistant" -> {
                                message.tool_calls.forEach { toolCall ->
                                    history += DisplayChatMessage(type = "tool", content = "姝ｅ湪璋冪敤宸ュ叿: ${toolCall.name.ifBlank { "tool" }}...")
                                }
                                if (message.content.isNotBlank()) {
                                    history += DisplayChatMessage(type = "ai", content = message.content)
                                }
                            }
                            "tool" -> history += DisplayChatMessage(type = "tool", content = "宸ュ叿璋冪敤瀹屾垚")
                            "user" -> history += DisplayChatMessage(
                                type = "user",
                                content = message.content,
                                images = message.image_urls,
                            )
                        }
                    }
                    uiState = uiState.copy(chatLoading = false, chatHistoryLoaded = true, chatMessages = history)
                } catch (exception: ApiException) {
                    uiState = uiState.copy(chatLoading = false)
                    showMessage(exception.message)
                }
            }
        }
    }

    fun sendChatMessage() {
        withToken { token ->
            val text = uiState.chatInput.trim()
            val images = uiState.chatImages
            if (text.isBlank() && images.isEmpty()) return@withToken

            uiState = uiState.copy(
                chatMessages = uiState.chatMessages +
                    DisplayChatMessage(type = "user", content = text, images = images) +
                    DisplayChatMessage(type = "ai", content = ""),
                chatInput = "",
                chatImages = emptyList(),
                isChatting = true,
                chatHistoryLoaded = true,
            )

            chatStream?.close()
            chatStream = repository.streamChat(
                token = token,
                request = ChatMessageRequest(message = text, images = images),
                onEvent = { envelope -> viewModelScope.launch { handleChatStreamEvent(envelope) } },
                onFailure = { error ->
                    viewModelScope.launch {
                        uiState = uiState.copy(
                            isChatting = false,
                            chatMessages = uiState.chatMessages + DisplayChatMessage(
                                type = "error",
                                content = error.message ?: "聊天服务暂时不可用",
                            ),
                        )
                    }
                },
                onClosed = { viewModelScope.launch { uiState = uiState.copy(isChatting = false) } },
            )
        }
    }

    fun clearChatHistory() {
        withToken { token ->
            viewModelScope.launch {
                runRequest {
                    repository.refreshChatContext(token)
                    uiState = uiState.copy(
                        chatHistoryLoaded = true,
                        chatMessages = listOf(DisplayChatMessage(type = "ai", content = "对话历史已清空。")),
                    )
                    showMessage("对话历史已重置")
                }
            }
        }
    }

    fun openAlertTask(recordId: String) {
        markAlertRead(recordId)
        uiState = uiState.copy(activeAlert = null, screen = AppScreen.History)
        fetchHistory(silent = true)
        openTaskDetail(recordId)
    }

    fun openFamilyNotificationCenter() {
        uiState.activeFamilyNotification?.id?.let(::markFamilyNotificationRead)
        uiState = uiState.copy(activeFamilyNotification = null, screen = AppScreen.Family)
        fetchFamilyOverview(silent = true)
    }

    private fun restoreSession() {
        withToken { token ->
            viewModelScope.launch {
                try {
                    val user = repository.fetchUser(token)
                    val occupations = repository.fetchOccupationOptions(token).occupations
                    syncUserProfile(
                        user = user,
                        occupations = occupations,
                        editorVisible = false,
                        token = token,
                        authenticated = true,
                    )
                    hydrateAuthenticatedContent(token)
                    BackgroundAlertService.start(getApplication())
                    startRealtime()
                    fetchTasks(silent = true)
                    fetchHistory(silent = true)
                    fetchRiskOverview(silent = true)
                    fetchFamilyOverview(silent = true)
                } catch (_: ApiException) {
                    logout()
                }
            }
        }
    }

    private fun startSession(token: String, user: UserProfile) {
        preferences.edit().putString("token", token).apply()
        viewModelScope.launch {
            val occupations = runCatching { repository.fetchOccupationOptions(token).occupations }
                .getOrElse { emptyList() }
            syncUserProfile(
                user = user,
                occupations = occupations,
                editorVisible = false,
                token = token,
                authenticated = true,
                loading = false,
                screen = AppScreen.Dashboard,
            )
            hydrateAuthenticatedContent(token)
            BackgroundAlertService.start(getApplication())
            startRealtime()
            fetchTasks(silent = true)
            fetchHistory(silent = true)
            fetchRiskOverview(silent = true)
            fetchFamilyOverview(silent = true)
        }
    }

    private fun syncUserProfile(
        user: UserProfile,
        occupations: List<String> = uiState.occupationOptions,
        editorVisible: Boolean = uiState.ageEditorVisible,
        token: String = uiState.token,
        authenticated: Boolean = uiState.isAuthenticated,
        loading: Boolean = uiState.loading,
        screen: AppScreen = uiState.screen,
    ) {
        uiState = uiState.copy(
            token = token,
            isAuthenticated = authenticated,
            user = user,
            ageInput = (user.age ?: 28).toString(),
            occupationInput = user.occupation,
            recentTagsInput = user.recent_tags.joinToString("\n"),
            occupationOptions = occupations,
            profileForm = ProfileFormState(
                age = user.age ?: 28,
                occupation = user.occupation,
                provinceCode = user.province_code,
                provinceName = user.province_name,
                cityCode = user.city_code,
                cityName = user.city_name,
                districtCode = user.district_code,
                districtName = user.district_name,
                locationSource = user.location_source.ifBlank { "manual" },
            ),
            ageEditorVisible = editorVisible,
            loading = loading,
            screen = screen,
        )
    }

    private suspend fun hydrateAuthenticatedContent(token: String) {
        val provinces = runCatching { repository.fetchProvinceOptions(token).provinces }.getOrElse { emptyList() }
        val cities = if (uiState.profileForm.provinceCode.isNotBlank()) {
            runCatching { repository.fetchCityOptions(token, uiState.profileForm.provinceCode).cities }.getOrElse { emptyList() }
        } else {
            emptyList()
        }
        val districts = if (uiState.profileForm.cityCode.isNotBlank()) {
            runCatching { repository.fetchDistrictOptions(token, uiState.profileForm.cityCode).districts }.getOrElse { emptyList() }
        } else {
            emptyList()
        }
        val packs = runCatching { loadSimulationPacks(token) }.getOrElse { emptyList() }
        val sessions = runCatching { loadSimulationSessions(token) }.getOrElse { emptyList() }
        val firstPack = packs.firstOrNull()
        uiState = uiState.copy(
            provinceOptions = provinces,
            cityOptions = cities,
            districtOptions = districts,
            simulationPackList = packs,
            simulationSessionList = sessions,
            simulationPack = uiState.simulationPack ?: firstPack,
            simulationPackId = uiState.simulationPackId.takeIf { it.isNotBlank() } ?: firstPack?.pack_id.orEmpty(),
        )
        sessions.firstOrNull { it.status == "in_progress" && it.pack_id.isNotBlank() }?.let { session ->
            runCatching { repository.fetchOngoingSimulation(token, session.pack_id) }.getOrNull()?.let { ongoing ->
                uiState = uiState.copy(
                    simulationPackId = ongoing.pack_id.takeIf { it.isNotBlank() } ?: session.pack_id,
                    simulationPack = session.pack ?: ongoing.pack ?: uiState.simulationPack,
                    simulationStatus = ongoing.status.ifBlank { "in_progress" },
                    simulationCurrentScore = ongoing.current_score.takeIf { it > 0 } ?: uiState.simulationCurrentScore,
                    simulationCurrentStep = ongoing.next_step,
                    simulationResult = session.result ?: ongoing.result,
                    simulationAnswers = session.answers,
                    simulationViewMode = SimulationViewMode.Overview,
                )
            }
        }
        if (uiState.profileForm.provinceCode.isNotBlank()) {
            runCatching { repository.fetchCurrentRegionCaseStats(token) }.getOrNull()?.let { stats ->
                uiState = uiState.copy(regionCaseStats = stats)
            }
        }
    }

    private fun startRealtime() {
        connectAlertSocket()
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            while (true) {
                delay(5_000)
                if (!uiState.isAuthenticated || uiState.token.isBlank()) continue
                fetchTasks(silent = true)
                fetchRiskOverview(silent = true)
                if (uiState.screen == AppScreen.Dashboard) {
                    fetchHistory(silent = true)
                }
                if (uiState.screen == AppScreen.History || uiState.screen == AppScreen.Alerts) {
                    fetchHistory(silent = true)
                }
                if (uiState.screen == AppScreen.Family || uiState.screen == AppScreen.FamilyManage) {
                    fetchFamilyOverview(silent = true)
                    if (!hasFamilyGroup()) fetchReceivedInvitations(silent = true)
                }
            }
        }
    }

    private fun stopRealtime() {
        pollingJob?.cancel()
        pollingJob = null
        alertReconnectJob?.cancel()
        familyReconnectJob?.cancel()
        disconnectAlertSocket()
        disconnectFamilySocket()
    }

    private fun connectAlertSocket() {
        if (!uiState.isAuthenticated || uiState.token.isBlank() || alertSocket != null) return
        alertSocket = repository.connectAlertSocket(
            token = uiState.token,
            onStateChange = { state ->
                if (state == RealtimeState.Connected) alertReconnectAttempts = 0
                uiState = uiState.copy(alertConnectionState = state)
                if (state == RealtimeState.Reconnecting && uiState.isAuthenticated) scheduleAlertReconnect()
            },
            onMessage = { event ->
                if (event.type != "risk_alert" && event.type != "high_risk_alert") return@connectAlertSocket
                if (uiState.alertItems.any { it.event.record_id == event.record_id }) return@connectAlertSocket
                uiState = uiState.copy(
                    alertItems = listOf(AlertMessageItem(event = event)) + uiState.alertItems,
                    activeAlert = event,
                )
                if (AppVisibilityTracker.isForeground) {
                    val riskLevel = normalizePersonalAlertRiskLabel(event.risk_level)
                    showMessage("${riskLevel}风险预警: ${event.title}", isError = riskLevel == "高")
                }
                fetchHistory(silent = true)
            },
        )
    }

    private fun disconnectAlertSocket() {
        alertSocket?.close(1000, "logout")
        alertSocket = null
        uiState = uiState.copy(alertConnectionState = RealtimeState.Disconnected)
    }

    private fun connectFamilySocket() {
        if (!uiState.isAuthenticated || uiState.token.isBlank() || !hasFamilyGroup() || familySocket != null) return
        familySocket = repository.connectFamilyNotificationSocket(
            token = uiState.token,
            onStateChange = { state ->
                if (state == RealtimeState.Connected) familyReconnectAttempts = 0
                uiState = uiState.copy(familyNotificationState = state)
                if (state == RealtimeState.Reconnecting && uiState.isAuthenticated && hasFamilyGroup()) scheduleFamilyReconnect()
            },
            onMessage = { event ->
                if (event.type != "family_high_risk_alert") return@connectFamilyNotificationSocket
                if (uiState.familyNotifications.any { it.id == event.notification_id }) return@connectFamilyNotificationSocket
                val notification = FamilyNotification(
                    id = event.notification_id,
                    family_id = event.family_id,
                    target_user_id = event.target_user_id,
                    target_name = event.target_name,
                    event_type = event.event_type,
                    record_id = event.record_id,
                    title = event.title,
                    case_summary = event.case_summary,
                    summary = event.summary,
                    scam_type = event.scam_type,
                    risk_level = event.risk_level,
                    event_at = event.event_at,
                    read_at = event.read_at,
                )
                uiState = uiState.copy(
                    familyNotifications = listOf(notification) + uiState.familyNotifications,
                    activeFamilyNotification = notification,
                )
                if (AppVisibilityTracker.isForeground) {
                    showMessage(notification.summary.ifBlank { "收到家庭高风险通知" }, isError = true)
                }
            },
        )
    }

    private fun disconnectFamilySocket() {
        familySocket?.close(1000, "family closed")
        familySocket = null
        uiState = uiState.copy(familyNotificationState = RealtimeState.Disconnected)
    }

    private fun scheduleAlertReconnect() {
        if (alertReconnectJob?.isActive == true) return
        val delayMs = (1000L * (1 shl alertReconnectAttempts.coerceAtMost(5))).coerceAtMost(30_000L)
        alertReconnectAttempts += 1
        alertReconnectJob = viewModelScope.launch {
            delay(delayMs)
            alertSocket = null
            connectAlertSocket()
        }
    }

    private fun scheduleFamilyReconnect() {
        if (familyReconnectJob?.isActive == true) return
        val delayMs = (1000L * (1 shl familyReconnectAttempts.coerceAtMost(5))).coerceAtMost(30_000L)
        familyReconnectAttempts += 1
        familyReconnectJob = viewModelScope.launch {
            delay(delayMs)
            familySocket = null
            connectFamilySocket()
        }
    }

    private fun handleChatStreamEvent(envelope: ChatStreamEnvelope) {
        when (envelope.type) {
            "content" -> {
                val index = uiState.chatMessages.indexOfLast { it.type == "ai" }
                if (index >= 0) {
                    val current = uiState.chatMessages[index]
                    uiState = uiState.copy(
                        chatMessages = uiState.chatMessages.toMutableList().apply {
                            set(index, current.copy(content = current.content + envelope.content))
                        },
                    )
                }
            }
            "tool_call" -> uiState = uiState.copy(
                chatMessages = uiState.chatMessages + DisplayChatMessage(
                    type = "tool",
                    content = "姝ｅ湪璋冪敤宸ュ叿: ${envelope.tool.ifBlank { "tool" }}...",
                ),
            )
            "tool_result" -> uiState = uiState.copy(
                chatMessages = uiState.chatMessages + DisplayChatMessage(
                    type = "tool",
                    content = "宸ュ叿 ${envelope.tool.ifBlank { "tool" }} 璋冪敤瀹屾垚",
                ),
            )
            "done" -> uiState = uiState.copy(isChatting = false)
        }
    }

    private fun startSmsCooldown() {
        smsCooldownJob?.cancel()
        uiState = uiState.copy(smsCooldownSeconds = 60)
        smsCooldownJob = viewModelScope.launch {
            while (uiState.smsCooldownSeconds > 0) {
                delay(1_000)
                uiState = uiState.copy(smsCooldownSeconds = (uiState.smsCooldownSeconds - 1).coerceAtLeast(0))
            }
        }
    }

    private suspend fun loadSimulationPacks(token: String): List<SimulationPack> {
        return repository.fetchSimulationPacks(token).packs
    }

    private suspend fun loadSimulationSessions(token: String): List<SimulationSessionItem> {
        return repository.fetchSimulationSessions(token).sessions
    }

    private suspend fun applyResolvedRegion(region: ResolvedRegion, source: String = "manual") {
        val token = uiState.token
        val cities = if (token.isNotBlank() && region.province_code.isNotBlank()) {
            runCatching { repository.fetchCityOptions(token, region.province_code).cities }.getOrElse { emptyList() }
        } else {
            emptyList()
        }
        val districts = if (token.isNotBlank() && region.city_code.isNotBlank()) {
            runCatching { repository.fetchDistrictOptions(token, region.city_code).districts }.getOrElse { emptyList() }
        } else {
            emptyList()
        }
        uiState = uiState.copy(
            profileForm = uiState.profileForm.copy(
                provinceCode = region.province_code,
                provinceName = region.province_name,
                cityCode = region.city_code,
                cityName = region.city_name,
                districtCode = region.district_code,
                districtName = region.district_name,
                locationSource = source.ifBlank { region.location_source.ifBlank { "manual" } },
            ),
            cityOptions = cities,
            districtOptions = districts,
        )
    }

    private fun hasAnyLocationPermission(): Boolean {
        val context = getApplication<Application>()
        return ContextCompat.checkSelfPermission(context, Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED ||
            ContextCompat.checkSelfPermission(context, Manifest.permission.ACCESS_COARSE_LOCATION) == PackageManager.PERMISSION_GRANTED
    }

    private fun getBestLastKnownLocation(): Location? {
        val context = getApplication<Application>()
        val locationManager = context.getSystemService(Context.LOCATION_SERVICE) as? LocationManager ?: return null
        val providers = runCatching { locationManager.getProviders(true) }.getOrDefault(emptyList())
        val locations = providers
            .mapNotNull { provider -> runCatching { locationManager.getLastKnownLocation(provider) }.getOrNull() }
        return locations
            .sortedWith(compareByDescending<Location> { it.time }.thenBy { it.accuracy })
            .firstOrNull()
    }

    private suspend fun buildResolveRegionRequest(location: Location): RegionResolvePreparation {
        val geoData = repository.fetchReverseGeocode(location.latitude, location.longitude)
        if (!isChinaCountryName(geoData.countryName)) {
            return RegionResolvePreparation(
                request = ResolveRegionRequest(
                    province_name = FALLBACK_PROVINCE,
                    city_name = FALLBACK_CITY,
                    district_name = FALLBACK_DISTRICT,
                    district_candidates = listOf(FALLBACK_DISTRICT),
                ),
                usedOverseasFallback = true,
                detectedCountry = geoData.countryName.trim(),
            )
        }
        val administrative = geoData.localityInfo.administrative
        val districtCandidates = administrative
            .map { it.name.trim() }
            .filter { name ->
                name.isNotBlank() &&
                    name != geoData.countryName &&
                    name != geoData.principalSubdivision
            }
            .reversed()
            .distinct()
        val districtName = listOf(
            districtCandidates.firstOrNull().orEmpty(),
            geoData.locality.trim(),
            geoData.city.trim(),
        ).firstOrNull { it.isNotBlank() }.orEmpty()
        val provinceName = geoData.principalSubdivision.trim()
        val cityName = geoData.city.trim()
        if (provinceName.isBlank() && cityName.isBlank() && districtName.isBlank()) {
            throw ApiException(0, "当前位置解析失败，请稍后重试或手动选择地区")
        }
        return RegionResolvePreparation(
            request = ResolveRegionRequest(
                province_name = provinceName,
                city_name = cityName,
                district_name = districtName,
                district_candidates = districtCandidates,
            ),
        )
    }

    private fun isChinaCountryName(countryName: String): Boolean {
        val normalized = countryName.trim().lowercase(Locale.ROOT)
        return normalized.contains("中国") ||
            normalized.contains("china") ||
            normalized.contains("中华人民共和国")
    }

    private fun mutateAuthForm(block: AuthFormState.() -> AuthFormState) {
        uiState = uiState.copy(authForm = uiState.authForm.block())
    }

    private fun requiresGraphCaptcha(): Boolean {
        return uiState.authMode == AuthMode.Register || uiState.loginMethod == LoginMethod.Password
    }

    private fun hasFamilyGroup(): Boolean = uiState.familyOverview.family != null

    private fun withToken(block: (String) -> Unit) {
        val token = uiState.token
        if (token.isBlank()) return
        block(token)
    }

    private suspend fun uriToDataUrl(uri: Uri): String? {
        return try {
            val resolver = getApplication<Application>().contentResolver
            val mimeType = resolver.getType(uri)
                ?: MimeTypeMap.getSingleton().getMimeTypeFromExtension(MimeTypeMap.getFileExtensionFromUrl(uri.toString()))
                ?: "application/octet-stream"
            val bytes = resolver.openInputStream(uri)?.use { it.readBytes() } ?: return null
            val encoded = Base64.encodeToString(bytes, Base64.NO_WRAP)
            "data:$mimeType;base64,$encoded"
        } catch (_: IOException) {
            null
        }
    }

    private suspend fun runRequest(
        silent: Boolean = false,
        block: suspend () -> Unit,
    ) {
        try {
            block()
        } catch (exception: ApiException) {
            if (exception.statusCode == 401 && uiState.isAuthenticated) {
                logout()
            } else if (!silent) {
                showMessage(exception.message)
            }
        }
    }

    private fun markAlertRead(recordId: String) {
        uiState = uiState.copy(
            alertItems = uiState.alertItems.map {
                if (it.event.record_id == recordId) it.copy(read = true) else it
            },
        )
    }

    private fun showMessage(message: String, isError: Boolean = false) {
        if (message.isBlank()) return
        latestMessage = UiMessage(
            text = message,
            isError = isError,
            id = System.nanoTime(),
            channel = if (isError) "error" else "info",
        )
    }

    fun emitUiMessage(message: String, isError: Boolean = false) {
        showMessage(message, isError)
    }
}

internal fun parseInstant(value: String): Instant {
    return runCatching { Instant.parse(value) }
        .recoverCatching { OffsetDateTime.parse(value).toInstant() }
        .getOrElse { Instant.EPOCH }
}

internal fun formatDateTime(value: String): String {
    return DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm")
        .withZone(ZoneId.systemDefault())
        .format(parseInstant(value))
}

internal fun recentWithinLastHour(value: String): Boolean {
    val instant = parseInstant(value)
    return Duration.between(instant, Instant.now()).toMinutes() in 0..60
}

internal fun normalizePersonalAlertRiskLabel(value: String): String = when {
    value.contains("高") -> "高"
    value.contains("中") -> "中"
    value.contains("低") -> "低"
    else -> "中"
}

internal fun extractTaskRiskLevel(riskSummary: String): String {
    if (riskSummary.isBlank()) return ""
    val normalized = normalizeRiskSummaryPayload(riskSummary)
    return Regex(""""(?:risk_level|level)"\s*:\s*"([^"]+)"""")
        .find(normalized)
        ?.groupValues
        ?.getOrNull(1)
        .orEmpty()
}

internal fun extractTaskRiskScore(riskSummary: String): Int? {
    if (riskSummary.isBlank()) return null
    val normalized = normalizeRiskSummaryPayload(riskSummary)
    return Regex(""""score"\s*:\s*(\d+)""")
        .find(normalized)
        ?.groupValues
        ?.getOrNull(1)
        ?.toIntOrNull()
}

internal fun extractRiskLevelFromReport(report: String): String {
    if (report.isBlank()) return ""
    return Regex("""风险等级\s*[:：]\s*([高中低])""")
        .find(report)
        ?.groupValues
        ?.getOrNull(1)
        .orEmpty()
}

internal fun inferRiskLevelFromScore(score: Int): String = when {
    score >= 80 -> "高"
    score >= 40 -> "中"
    score > 0 -> "低"
    else -> ""
}

private fun normalizeRiskSummaryPayload(raw: String): String {
    return raw.trim()
        .removeSurrounding("\"")
        .replace("\\\"", "\"")
        .replace("\\n", "\n")
        .replace("\\r", "\r")
        .replace("\\t", "\t")
}
