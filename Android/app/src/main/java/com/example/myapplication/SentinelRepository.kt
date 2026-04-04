package com.example.myapplication

import java.io.IOException
import java.net.ConnectException
import java.net.SocketTimeoutException
import java.net.UnknownHostException
import java.time.Instant
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlinx.serialization.Serializable
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okhttp3.sse.EventSource
import okhttp3.sse.EventSourceListener
import okhttp3.sse.EventSources

private const val HEARTBEAT_TYPE_PING = "ping"
private const val HEARTBEAT_TYPE_PONG = "pong"

@Serializable
private data class RealtimeHeartbeatEnvelope(
    val type: String = "",
    val sent_at: String = "",
    val received_at: String = "",
)

class SentinelRepository(
    private val origin: String,
) {
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        explicitNulls = false
        coerceInputValues = true
    }

    private val httpClient = OkHttpClient.Builder()
        .connectTimeout(20, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .readTimeout(60, TimeUnit.SECONDS)
        .build()

    private val eventSourceFactory = EventSources.createFactory(httpClient)
    private val jsonMediaType = "application/json; charset=utf-8".toMediaType()
    private val apiBase = "${origin.trimEnd('/')}/api"

    suspend fun fetchCaptcha(): CaptchaResponse = get("/auth/captcha", authToken = null)

    suspend fun sendSmsCode(phone: String): MessageResponse = post(
        path = "/auth/sms-code",
        authToken = null,
        body = json.encodeToString(SmsCodeRequest(phone)),
    )

    suspend fun register(request: AuthRegisterRequest): UserProfile = post(
        path = "/auth/register",
        authToken = null,
        body = json.encodeToString(request),
    )

    suspend fun loginWithPassword(request: AuthPasswordLoginRequest): AuthLoginResponse = post(
        path = "/auth/login",
        authToken = null,
        body = json.encodeToString(request),
    )

    suspend fun loginWithSms(request: AuthSmsLoginRequest): AuthLoginResponse = post(
        path = "/auth/login",
        authToken = null,
        body = json.encodeToString(request),
    )

    suspend fun fetchUser(token: String): UserProfile = get("/user", token)

    suspend fun fetchOccupationOptions(token: String): OccupationOptionsResponse =
        get("/user/profile/options/occupations", token)

    suspend fun fetchProvinceOptions(token: String): ProvinceListResponse =
        get("/regions/provinces", token)

    suspend fun fetchCityOptions(token: String, provinceCode: String): CityListResponse =
        get("/regions/cities?province_code=${provinceCode.trim()}", token)

    suspend fun fetchDistrictOptions(token: String, cityCode: String): DistrictListResponse =
        get("/regions/districts?city_code=${cityCode.trim()}", token)

    suspend fun resolveRegion(
        token: String,
        request: ResolveRegionRequest,
    ): ResolveRegionResponse = post(
        path = "/regions/resolve",
        authToken = token,
        body = json.encodeToString(request),
    )

    suspend fun fetchCurrentRegionCaseStats(token: String): CurrentRegionCaseStatsResponse =
        get("/regions/cases/stats/current", token)

    suspend fun fetchReverseGeocode(
        latitude: Double,
        longitude: Double,
    ): BigDataCloudReverseGeocodeResponse = getAbsolute(
        "https://api.bigdatacloud.net/data/reverse-geocode-client" +
            "?latitude=$latitude&longitude=$longitude&localityLanguage=zh",
    )

    suspend fun deleteUser(token: String): MessageResponse = delete("/user", token)

    suspend fun updateAge(token: String, age: Int): MessageResponse = put(
        path = "/scam/multimodal/user/age",
        authToken = token,
        body = json.encodeToString(UpdateAgeRequest(age)),
    )

    suspend fun updateUserProfile(
        token: String,
        request: UpdateUserProfileRequest,
    ): UpdateUserProfileResponse = put(
        path = "/user/profile",
        authToken = token,
        body = json.encodeToString(request),
    )

    suspend fun analyze(token: String, request: AnalyzeRequest): AnalyzeResponse = post(
        path = "/scam/multimodal/analyze",
        authToken = token,
        body = json.encodeToString(request),
    )

    suspend fun quickAnalyzeImage(token: String, image: String): QuickAnalyzeResponse = post(
        path = "/scam/image/quick-analyze",
        authToken = token,
        body = json.encodeToString(QuickAnalyzeRequest(image = image)),
    )

    suspend fun fetchTasks(token: String): TaskListResponse = get("/scam/multimodal/tasks", token)

    suspend fun fetchHistory(token: String): HistoryListResponse = get("/scam/multimodal/history", token)

    suspend fun fetchTaskDetail(token: String, taskId: String): TaskDetailResponse =
        get("/scam/multimodal/tasks/${taskId.trim()}", token)

    suspend fun deleteHistoryRecord(token: String, recordId: String): MessageResponse =
        delete("/scam/multimodal/history/${recordId.trim()}", token)

    suspend fun fetchRiskOverview(token: String, interval: String): RiskOverviewResponse =
        get("/scam/multimodal/history/overview?interval=$interval", token)

    suspend fun fetchSimulationPacks(
        token: String,
        limit: Int = 50,
    ): SimulationPackListResponse = get("/scam/simulation/packs?limit=$limit", token)

    suspend fun fetchSimulationSessions(
        token: String,
        limit: Int = 50,
    ): SimulationSessionListResponse = get("/scam/simulation/sessions?limit=$limit", token)

    suspend fun deleteSimulationSession(token: String, sessionId: String): MessageResponse =
        delete("/scam/simulation/sessions/$sessionId", token)

    suspend fun generateSimulationPack(
        token: String,
        request: SimulationGeneratePackRequest,
    ): SimulationGeneratePackResponse = post(
        path = "/scam/simulation/packs/generate",
        authToken = token,
        body = json.encodeToString(request),
    )

    suspend fun answerSimulationSession(
        token: String,
        request: SimulationAnswerRequest,
    ): SimulationSessionResponse = post(
        path = "/scam/simulation/sessions/answer",
        authToken = token,
        body = json.encodeToString(request),
    )

    suspend fun fetchOngoingSimulation(token: String, packId: String): SimulationOngoingResponse =
        get("/scam/simulation/packs/$packId/ongoing", token)

    suspend fun fetchChatContext(token: String): ChatContextResponse = get("/chat/context", token)

    suspend fun refreshChatContext(token: String): MessageResponse = post(
        path = "/chat/refresh",
        authToken = token,
        body = null,
    )

    suspend fun fetchFamilyOverview(token: String): FamilyOverviewResponse = get("/families/me", token)

    suspend fun fetchReceivedFamilyInvitations(token: String): FamilyInvitationListResponse =
        get("/families/invitations/received", token)

    suspend fun createFamily(token: String, name: String): FamilyOverviewResponse = post(
        path = "/families",
        authToken = token,
        body = json.encodeToString(CreateFamilyRequest(name)),
    )

    suspend fun createFamilyInvitation(
        token: String,
        request: CreateFamilyInvitationRequest,
    ): FamilyInvitationCreateResponse = post(
        path = "/families/invitations",
        authToken = token,
        body = json.encodeToString(request),
    )

    suspend fun acceptFamilyInvitation(
        token: String,
        inviteCode: String,
    ): FamilyOverviewResponse = post(
        path = "/families/invitations/accept",
        authToken = token,
        body = json.encodeToString(AcceptFamilyInvitationRequest(inviteCode)),
    )

    suspend fun createGuardianLink(
        token: String,
        request: CreateGuardianLinkRequest,
    ): GuardianLinkCreateResponse = post(
        path = "/families/guardian-links",
        authToken = token,
        body = json.encodeToString(request),
    )

    suspend fun deleteFamilyMember(token: String, memberId: Int): MessageResponse =
        delete("/families/members/$memberId", token)

    suspend fun deleteGuardianLink(token: String, linkId: Int): MessageResponse =
        delete("/families/guardian-links/$linkId", token)

    suspend fun markFamilyNotificationRead(token: String, notificationId: Int): MessageResponse =
        post(
            path = "/families/notifications/$notificationId/read",
            authToken = token,
            body = null,
        )

    fun streamChat(
        token: String,
        request: ChatMessageRequest,
        onEvent: (ChatStreamEnvelope) -> Unit,
        onFailure: (Throwable) -> Unit,
        onClosed: () -> Unit,
    ): AutoCloseable {
        val body = json.encodeToString(request).toRequestBody(jsonMediaType)
        val httpRequest = Request.Builder()
            .url("$apiBase/chat")
            .addHeader("Authorization", "Bearer $token")
            .addHeader("Accept", "text/event-stream")
            .post(body)
            .build()

        val eventSource = eventSourceFactory.newEventSource(
            httpRequest,
            object : EventSourceListener() {
                override fun onEvent(
                    eventSource: EventSource,
                    id: String?,
                    type: String?,
                    data: String,
                ) {
                    runCatching {
                        json.decodeFromString<ChatStreamEnvelope>(data)
                    }.onSuccess(onEvent).onFailure(onFailure)
                }

                override fun onFailure(
                    eventSource: EventSource,
                    t: Throwable?,
                    response: Response?,
                ) {
                    onFailure(t ?: IOException("chat stream failed"))
                }

                override fun onClosed(eventSource: EventSource) {
                    onClosed()
                }
            },
        )

        return AutoCloseable { eventSource.cancel() }
    }

    fun connectAlertSocket(
        token: String,
        onStateChange: (RealtimeState) -> Unit,
        onMessage: (AlertEvent) -> Unit,
    ): WebSocket {
        val request = Request.Builder()
            .url(websocketUrl("/alert/ws?token=${token.trim()}"))
            .build()

        onStateChange(RealtimeState.Connecting)
        return httpClient.newWebSocket(
            request,
            object : WebSocketListener() {
                override fun onOpen(webSocket: WebSocket, response: Response) {
                    onStateChange(RealtimeState.Connected)
                }

                override fun onMessage(webSocket: WebSocket, text: String) {
                    if (handleRealtimeHeartbeatMessage(webSocket, text)) return
                    runCatching { json.decodeFromString<AlertEvent>(text) }
                        .onSuccess(onMessage)
                }

                override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                    onStateChange(RealtimeState.Reconnecting)
                }

                override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                    onStateChange(RealtimeState.Reconnecting)
                }
            },
        )
    }

    fun connectFamilyNotificationSocket(
        token: String,
        onStateChange: (RealtimeState) -> Unit,
        onMessage: (FamilyNotificationEvent) -> Unit,
    ): WebSocket {
        val request = Request.Builder()
            .url(websocketUrl("/families/notifications/ws?token=${token.trim()}"))
            .build()

        onStateChange(RealtimeState.Connecting)
        return httpClient.newWebSocket(
            request,
            object : WebSocketListener() {
                override fun onOpen(webSocket: WebSocket, response: Response) {
                    onStateChange(RealtimeState.Connected)
                }

                override fun onMessage(webSocket: WebSocket, text: String) {
                    if (handleRealtimeHeartbeatMessage(webSocket, text)) return
                    runCatching { json.decodeFromString<FamilyNotificationEvent>(text) }
                        .onSuccess(onMessage)
                }

                override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                    onStateChange(RealtimeState.Reconnecting)
                }

                override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                    onStateChange(RealtimeState.Reconnecting)
                }
            },
        )
    }

    fun close() {
        httpClient.dispatcher.executorService.shutdown()
        httpClient.connectionPool.evictAll()
    }

    private fun websocketUrl(path: String): String {
        val normalizedOrigin = origin.trimEnd('/')
        val scheme = if (normalizedOrigin.startsWith("https://")) "wss://" else "ws://"
        val host = normalizedOrigin.removePrefix("https://").removePrefix("http://")
        return "$scheme$host/api${if (path.startsWith("/")) path else "/$path"}"
    }

    private fun handleRealtimeHeartbeatMessage(webSocket: WebSocket, text: String): Boolean {
        val envelope = runCatching { json.decodeFromString<RealtimeHeartbeatEnvelope>(text) }
            .getOrNull()
            ?: return false

        return when (envelope.type.trim()) {
            HEARTBEAT_TYPE_PING -> {
                runCatching {
                    webSocket.send(
                        json.encodeToString(
                            RealtimeHeartbeatEnvelope(
                                type = HEARTBEAT_TYPE_PONG,
                                sent_at = envelope.sent_at.trim(),
                                received_at = Instant.now().toString(),
                            ),
                        ),
                    )
                }
                true
            }

            HEARTBEAT_TYPE_PONG -> true
            else -> false
        }
    }

    private suspend inline fun <reified T> get(path: String, authToken: String?): T = withContext(Dispatchers.IO) {
        executeJson(
            request = Request.Builder()
                .url("$apiBase$path")
                .applyAuth(authToken)
                .addHeader("Accept", "application/json")
                .get()
                .build(),
        )
    }

    private suspend inline fun <reified T> getAbsolute(url: String): T = withContext(Dispatchers.IO) {
        executeJson(
            request = Request.Builder()
                .url(url)
                .addHeader("Accept", "application/json")
                .get()
                .build(),
        )
    }

    private suspend inline fun <reified T> post(
        path: String,
        authToken: String?,
        body: String?,
    ): T = withContext(Dispatchers.IO) {
        executeJson(
            request = Request.Builder()
                .url("$apiBase$path")
                .applyAuth(authToken)
                .addHeader("Accept", "application/json")
                .post((body ?: "").toRequestBody(jsonMediaType))
                .build(),
        )
    }

    private suspend inline fun <reified T> put(
        path: String,
        authToken: String?,
        body: String,
    ): T = withContext(Dispatchers.IO) {
        executeJson(
            request = Request.Builder()
                .url("$apiBase$path")
                .applyAuth(authToken)
                .addHeader("Accept", "application/json")
                .put(body.toRequestBody(jsonMediaType))
                .build(),
        )
    }

    private suspend inline fun <reified T> delete(path: String, authToken: String?): T = withContext(Dispatchers.IO) {
        executeJson(
            request = Request.Builder()
                .url("$apiBase$path")
                .applyAuth(authToken)
                .addHeader("Accept", "application/json")
                .delete()
                .build(),
        )
    }

    private inline fun Request.Builder.applyAuth(token: String?): Request.Builder = apply {
        if (!token.isNullOrBlank()) {
            addHeader("Authorization", "Bearer $token")
        }
    }

    private inline fun <reified T> executeJson(request: Request): T {
        try {
            httpClient.newCall(request).execute().use { response ->
                val body = response.body?.string().orEmpty()
                if (!response.isSuccessful) {
                    val errorMessage = runCatching {
                        json.decodeFromString<ErrorResponse>(body).error
                            ?: json.decodeFromString<ErrorResponse>(body).message
                    }.getOrNull().orEmpty()
                    throw ApiException(
                        statusCode = response.code,
                        message = errorMessage.ifBlank { "Request failed: ${response.code}" },
                    )
                }
                return json.decodeFromString(body)
            }
        } catch (exception: ApiException) {
            throw exception
        } catch (exception: UnknownHostException) {
            throw ApiException(0, "无法连接服务器，请检查地址配置")
        } catch (exception: ConnectException) {
            throw ApiException(0, "服务器未启动或无法访问")
        } catch (exception: SocketTimeoutException) {
            throw ApiException(0, "请求超时，请稍后重试")
        } catch (exception: IOException) {
            throw ApiException(0, exception.message ?: "网络请求失败")
        } catch (exception: Exception) {
            throw ApiException(0, exception.message ?: "请求处理失败")
        }
    }
}
