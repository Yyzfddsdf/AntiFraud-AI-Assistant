package com.example.myapplication

import kotlinx.serialization.Serializable

enum class AuthMode {
    Login,
    Register,
}

enum class LoginMethod {
    Password,
    Sms,
}

enum class AppScreen {
    Dashboard,
    History,
    RiskTrend,
    Alerts,
    Submit,
    SimulationQuiz,
    Chat,
    Family,
    FamilyManage,
    Profile,
    ProfilePrivacy,
}

enum class AnalyzeAssetKind {
    Images,
    Audios,
    Videos,
}

enum class RealtimeState {
    Disconnected,
    Connecting,
    Connected,
    Reconnecting,
}

@Serializable
data class ErrorResponse(
    val error: String? = null,
    val message: String? = null,
)

@Serializable
data class CaptchaResponse(
    val captchaId: String = "",
    val captchaImage: String = "",
    val expiresIn: Int = 0,
)

@Serializable
data class SmsCodeRequest(
    val phone: String,
)

@Serializable
data class AuthRegisterRequest(
    val username: String,
    val email: String,
    val phone: String,
    val password: String,
    val captchaId: String,
    val captchaCode: String,
    val smsCode: String,
)

@Serializable
data class AuthPasswordLoginRequest(
    val account: String,
    val password: String,
    val captchaId: String,
    val captchaCode: String,
)

@Serializable
data class AuthSmsLoginRequest(
    val phone: String,
    val smsCode: String,
)

@Serializable
data class UserProfile(
    val id: Int? = null,
    val username: String = "",
    val email: String = "",
    val phone: String = "",
    val role: String = "user",
    val age: Int? = null,
    val occupation: String = "",
    val recent_tags: List<String> = emptyList(),
    val province_code: String = "",
    val province_name: String = "",
    val city_code: String = "",
    val city_name: String = "",
    val district_code: String = "",
    val district_name: String = "",
    val location_source: String = "",
)

@Serializable
data class AuthLoginResponse(
    val message: String = "",
    val token: String = "",
    val user: UserProfile = UserProfile(),
)

@Serializable
data class MessageResponse(
    val message: String = "",
)

@Serializable
data class AnalyzeRequest(
    val text: String = "",
    val videos: List<String> = emptyList(),
    val audios: List<String> = emptyList(),
    val images: List<String> = emptyList(),
)

@Serializable
data class AnalyzeResponse(
    val task_id: String = "",
    val status: String = "",
    val message: String = "",
)

@Serializable
data class QuickAnalyzeRequest(
    val image: String = "",
)

@Serializable
data class QuickAnalyzeResponse(
    val risk_level: String = "",
    val reason: String = "",
)

@Serializable
data class TaskSummary(
    val task_id: String = "",
    val user_id: String = "",
    val title: String = "",
    val status: String = "",
    val summary: String = "",
    val created_at: String = "",
    val updated_at: String = "",
)

@Serializable
data class TaskListResponse(
    val user_id: String = "",
    val tasks: List<TaskSummary> = emptyList(),
)

@Serializable
data class HistoryRecord(
    val record_id: String = "",
    val title: String = "",
    val case_summary: String = "",
    val scam_type: String = "",
    val risk_level: String = "",
    val created_at: String = "",
)

@Serializable
data class HistoryListResponse(
    val user_id: String = "",
    val history: List<HistoryRecord> = emptyList(),
)

@Serializable
data class RiskStats(
    val high: Int = 0,
    val medium: Int = 0,
    val low: Int = 0,
    val total: Int = 0,
)

@Serializable
data class RiskAnalysis(
    val current_bucket: String = "",
    val previous_bucket: String = "",
    val overall_trend: String = "",
    val high_risk_trend: String = "",
    val summary: String = "",
)

@Serializable
data class RiskTrendPoint(
    val time_bucket: String = "",
    val high: Int = 0,
    val medium: Int = 0,
    val low: Int = 0,
    val total: Int = 0,
)

@Serializable
data class RiskOverviewResponse(
    val stats: RiskStats = RiskStats(),
    val analysis: RiskAnalysis = RiskAnalysis(),
    val trend: List<RiskTrendPoint> = emptyList(),
)

@Serializable
data class TaskPayload(
    val text: String = "",
    val videos: List<String> = emptyList(),
    val audios: List<String> = emptyList(),
    val images: List<String> = emptyList(),
    val video_insights: List<String> = emptyList(),
    val audio_insights: List<String> = emptyList(),
    val image_insights: List<String> = emptyList(),
)

@Serializable
data class TaskDetail(
    val task_id: String = "",
    val user_id: String = "",
    val title: String = "",
    val status: String = "",
    val scam_type: String = "",
    val risk_level: String = "",
    val risk_score: Int = 0,
    val risk_summary: String = "",
    val summary: String = "",
    val created_at: String = "",
    val updated_at: String = "",
    val payload: TaskPayload = TaskPayload(),
    val report: String = "",
)

@Serializable
data class TaskDetailResponse(
    val task: TaskDetail = TaskDetail(),
)

@Serializable
data class UpdateAgeRequest(
    val age: Int,
)

@Serializable
data class UpdateUserProfileRequest(
    val age: Int,
    val occupation: String = "",
    val province_code: String = "",
    val province_name: String = "",
    val city_code: String = "",
    val city_name: String = "",
    val district_code: String = "",
    val district_name: String = "",
    val location_source: String = "",
)

@Serializable
data class UpdateUserProfileResponse(
    val message: String = "",
    val user: UserProfile = UserProfile(),
)

@Serializable
data class OccupationOptionsResponse(
    val occupations: List<String> = emptyList(),
    val count: Int = 0,
)

@Serializable
data class RegionOption(
    val code: String = "",
    val name: String = "",
)

@Serializable
data class ProvinceListResponse(
    val provinces: List<RegionOption> = emptyList(),
)

@Serializable
data class CityListResponse(
    val cities: List<RegionOption> = emptyList(),
)

@Serializable
data class DistrictListResponse(
    val districts: List<RegionOption> = emptyList(),
)

@Serializable
data class ResolveRegionRequest(
    val province_name: String = "",
    val city_name: String = "",
    val district_name: String = "",
    val district_candidates: List<String> = emptyList(),
)

@Serializable
data class ResolvedRegion(
    val province_code: String = "",
    val province_name: String = "",
    val city_code: String = "",
    val city_name: String = "",
    val district_code: String = "",
    val district_name: String = "",
    val location_source: String = "",
)

@Serializable
data class ResolveRegionResponse(
    val region: ResolvedRegion? = null,
)

@Serializable
data class BigDataCloudAdministrativeItem(
    val name: String = "",
)

@Serializable
data class BigDataCloudLocalityInfo(
    val administrative: List<BigDataCloudAdministrativeItem> = emptyList(),
)

@Serializable
data class BigDataCloudReverseGeocodeResponse(
    val countryName: String = "",
    val principalSubdivision: String = "",
    val city: String = "",
    val locality: String = "",
    val localityInfo: BigDataCloudLocalityInfo = BigDataCloudLocalityInfo(),
    val description: String = "",
)

@Serializable
data class RegionStatsLocation(
    val granularity: String = "",
    val granularity_label: String = "",
    val province_code: String = "",
    val province_name: String = "",
    val city_code: String = "",
    val city_name: String = "",
    val district_code: String = "",
    val district_name: String = "",
)

@Serializable
data class RegionCaseStatsSummary(
    val total_count: Int = 0,
    val today_count: Int = 0,
    val last_7d_count: Int = 0,
    val last_30d_count: Int = 0,
    val high_count: Int = 0,
)

@Serializable
data class RegionScamTypeCount(
    val scam_type: String = "",
    val count: Int = 0,
)

@Serializable
data class CurrentRegionCaseStatsResponse(
    val region: RegionStatsLocation? = null,
    val summary: RegionCaseStatsSummary? = null,
    val top_scam_types: List<RegionScamTypeCount> = emptyList(),
)

@Serializable
data class SimulationGeneratePackRequest(
    val case_type: String = "",
    val target_persona: String = "",
    val difficulty: String = "easy",
    val locale: String = "zh-CN",
)

@Serializable
data class SimulationGeneratePackResponse(
    val message: String = "",
)

@Serializable
data class SimulationOption(
    val key: String = "",
    val text: String = "",
)

@Serializable
data class SimulationStep(
    val step_id: String = "",
    val step_type: String = "",
    val narrative: String = "",
    val question: String = "",
    val options: List<SimulationOption> = emptyList(),
)

@Serializable
data class SimulationPack(
    val pack_id: String = "",
    val title: String = "",
    val intro: String = "",
    val case_type: String = "",
    val target_persona: String = "",
    val difficulty: String = "easy",
    val locale: String = "zh-CN",
    val steps: List<SimulationStep> = emptyList(),
)

@Serializable
data class SimulationPackListResponse(
    val packs: List<SimulationPack> = emptyList(),
)

@Serializable
data class SimulationSessionAnswer(
    val step_id: String = "",
    val option_key: String = "",
    val score_delta: Int = 0,
    val is_correct: Boolean? = null,
)

@Serializable
data class SimulationResult(
    val level: String = "",
    val total_score: Int = 0,
    val weaknesses: List<String>? = null,
    val strengths: List<String>? = null,
    val advice: List<String>? = null,
)

@Serializable
data class SimulationSessionItem(
    val pack_id: String = "",
    val title: String = "",
    val score: Int = 0,
    val level: String = "",
    val status: String = "",
    val answers: List<SimulationSessionAnswer> = emptyList(),
    val result: SimulationResult? = null,
    val pack: SimulationPack? = null,
)

@Serializable
data class SimulationSessionListResponse(
    val sessions: List<SimulationSessionItem> = emptyList(),
)

@Serializable
data class SimulationAnswerRequest(
    val pack_id: String,
    val step_id: String? = null,
    val option_key: String? = null,
)

@Serializable
data class SimulationSessionResponse(
    val status: String = "",
    val current_score: Int = 0,
    val pack: SimulationPack? = null,
    val next_step: SimulationStep? = null,
    val result: SimulationResult? = null,
    val message: String = "",
)

@Serializable
data class SimulationOngoingResponse(
    val pack_id: String = "",
    val status: String = "",
    val current_score: Int = 0,
    val pack: SimulationPack? = null,
    val next_step: SimulationStep? = null,
    val result: SimulationResult? = null,
)

@Serializable
data class FamilyInfo(
    val id: Int = 0,
    val name: String = "",
    val owner_user_id: Int = 0,
    val owner_name: String = "",
    val owner_email: String = "",
    val owner_phone: String = "",
    val invite_code: String = "",
    val status: String = "",
    val member_count: Int = 0,
    val guardian_count: Int = 0,
)

@Serializable
data class FamilyMember(
    val member_id: Int = 0,
    val family_id: Int = 0,
    val user_id: Int = 0,
    val username: String = "",
    val email: String = "",
    val phone: String = "",
    val role: String = "",
    val relation: String = "",
    val status: String = "",
    val created_at: String = "",
)

@Serializable
data class FamilyInvitation(
    val id: Int = 0,
    val family_id: Int = 0,
    val family_name: String = "",
    val inviter_user_id: Int = 0,
    val inviter_name: String = "",
    val inviter_email: String = "",
    val inviter_phone: String = "",
    val invitee_email: String = "",
    val invitee_phone: String = "",
    val role: String = "",
    val relation: String = "",
    val invite_code: String = "",
    val status: String = "",
    val expires_at: String = "",
)

@Serializable
data class FamilyInvitationListResponse(
    val invitations: List<FamilyInvitation> = emptyList(),
)

@Serializable
data class FamilyInvitationCreateResponse(
    val invitation: FamilyInvitation? = null,
)

@Serializable
data class GuardianLink(
    val id: Int = 0,
    val guardian_user_id: Int = 0,
    val guardian_name: String = "",
    val guardian_email: String = "",
    val member_user_id: Int = 0,
    val member_name: String = "",
    val member_email: String = "",
    val created_at: String = "",
)

@Serializable
data class FamilyOverviewResponse(
    val family: FamilyInfo? = null,
    val current_member: FamilyMember? = null,
    val members: List<FamilyMember> = emptyList(),
    val invitations: List<FamilyInvitation> = emptyList(),
    val guardian_links: List<GuardianLink> = emptyList(),
    val unread_notification_count: Int = 0,
)

@Serializable
data class GuardianLinkCreateResponse(
    val guardian_link: GuardianLink? = null,
)

@Serializable
data class CreateFamilyRequest(
    val name: String,
)

@Serializable
data class CreateFamilyInvitationRequest(
    val invitee_email: String = "",
    val invitee_phone: String = "",
    val role: String = "",
    val relation: String = "",
)

@Serializable
data class AcceptFamilyInvitationRequest(
    val invite_code: String,
)

@Serializable
data class CreateGuardianLinkRequest(
    val guardian_user_id: Int,
    val member_user_id: Int,
)

@Serializable
data class FamilyNotification(
    val id: Int = 0,
    val family_id: Int = 0,
    val target_user_id: Int = 0,
    val target_name: String = "",
    val event_type: String = "",
    val record_id: String = "",
    val title: String = "",
    val case_summary: String = "",
    val summary: String = "",
    val scam_type: String = "",
    val risk_level: String = "",
    val event_at: String = "",
    val read_at: String = "",
)

@Serializable
data class AlertEvent(
    val type: String = "",
    val user_id: String = "",
    val record_id: String = "",
    val title: String = "",
    val case_summary: String = "",
    val scam_type: String = "",
    val risk_level: String = "",
    val created_at: String = "",
    val sent_at: String = "",
)

@Serializable
data class FamilyNotificationEvent(
    val type: String = "",
    val notification_id: Int = 0,
    val family_id: Int = 0,
    val target_user_id: Int = 0,
    val target_name: String = "",
    val event_type: String = "",
    val record_id: String = "",
    val title: String = "",
    val case_summary: String = "",
    val summary: String = "",
    val scam_type: String = "",
    val risk_level: String = "",
    val event_at: String = "",
    val read_at: String = "",
)

@Serializable
data class ChatMessageRequest(
    val message: String = "",
    val images: List<String> = emptyList(),
)

@Serializable
data class ChatToolCall(
    val id: String = "",
    val name: String = "",
    val arguments: String = "",
)

@Serializable
data class ChatContextMessage(
    val role: String = "",
    val content: String = "",
    val image_urls: List<String> = emptyList(),
    val tool_calls: List<ChatToolCall> = emptyList(),
    val tool_call_id: String = "",
)

@Serializable
data class ChatContextResponse(
    val user_id: String = "",
    val has_context: Boolean = false,
    val ttl_seconds: Int = 0,
    val messages: List<ChatContextMessage> = emptyList(),
)

@Serializable
data class ChatStreamEnvelope(
    val type: String = "",
    val content: String = "",
    val tool: String = "",
    val id: String = "",
    val reason: String = "",
)

data class DisplayChatMessage(
    val id: Long = System.nanoTime(),
    val type: String = "ai",
    val content: String = "",
    val images: List<String> = emptyList(),
)

data class DropdownOption(
    val value: String,
    val label: String,
    val hint: String,
)

data class ReportSection(
    val id: Int,
    val title: String,
    val content: String,
)

class ApiException(
    val statusCode: Int,
    override val message: String,
) : Exception(message)
