package com.example.myapplication

import android.content.Context
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class AccessibilityAutoAnalyzeConfig(
    @SerialName("debug_toast_enabled")
    val debugToastEnabled: Boolean = true,
    @SerialName("debug_toast_min_interval_ms")
    val debugToastMinIntervalMs: Long = 1_500L,
    @SerialName("ocr_fallback_enabled")
    val ocrFallbackEnabled: Boolean = true,
    @SerialName("ocr_attempt_cooldown_ms")
    val ocrAttemptCooldownMs: Long = 15_000L,
    @SerialName("min_sensitive_score")
    val minSensitiveScore: Int = 3,
    @SerialName("global_cooldown_ms")
    val globalCooldownMs: Long = 45_000L,
    @SerialName("same_screen_cooldown_ms")
    val sameScreenCooldownMs: Long = 120_000L,
    @SerialName("max_captured_texts")
    val maxCapturedTexts: Int = 60,
    @SerialName("max_text_length")
    val maxTextLength: Int = 48,
    @SerialName("multi_group_bonus_threshold")
    val multiGroupBonusThreshold: Int = 3,
    @SerialName("multi_group_bonus")
    val multiGroupBonus: Int = 1,
    @SerialName("link_indicator_bonus")
    val linkIndicatorBonus: Int = 1,
    @SerialName("verification_code_bonus")
    val verificationCodeBonus: Int = 1,
    @SerialName("ignored_package_prefixes")
    val ignoredPackagePrefixes: List<String> = listOf(
        "com.android.systemui",
        "com.android.settings",
        "com.google.android.permissioncontroller",
    ),
    @SerialName("ignored_package_keywords")
    val ignoredPackageKeywords: List<String> = listOf("permissioncontroller"),
    @SerialName("link_indicators")
    val linkIndicators: List<String> = listOf(
        "http",
        "www.",
        ".com",
        "点击链接",
        "下载",
        "安装",
    ),
    @SerialName("keyword_groups")
    val keywordGroups: List<AccessibilityKeywordGroup> = listOf(
        AccessibilityKeywordGroup(
            id = "support",
            label = "客服身份",
            weight = 1,
            keywords = listOf("客服", "官方", "专员", "工作人员"),
        ),
        AccessibilityKeywordGroup(
            id = "payment",
            label = "转账收款",
            weight = 2,
            keywords = listOf("转账", "汇款", "付款", "收款", "充值", "打款", "提现", "银行卡", "账户", "扫码"),
        ),
        AccessibilityKeywordGroup(
            id = "verification",
            label = "验证码",
            weight = 2,
            keywords = listOf("验证码", "校验码", "短信码", "动态码", "口令"),
        ),
        AccessibilityKeywordGroup(
            id = "install",
            label = "下载链接",
            weight = 2,
            keywords = listOf("下载", "安装", "链接", "点击链接", "网址", "浏览器"),
        ),
        AccessibilityKeywordGroup(
            id = "finance",
            label = "资金诱导",
            weight = 2,
            keywords = listOf("退款", "会员", "刷单", "贷款", "投资", "理财", "征信", "解冻", "认证", "保证金", "返利", "做任务"),
        ),
        AccessibilityKeywordGroup(
            id = "contact",
            label = "私下联系",
            weight = 1,
            keywords = listOf("私聊", "私下", "加微", "加V", "微信", "QQ"),
        ),
    ),
    @SerialName("package_overrides")
    val packageOverrides: List<AccessibilityPackageOverride> = emptyList(),
)

@Serializable
data class AccessibilityKeywordGroup(
    val id: String,
    val label: String,
    val weight: Int = 1,
    val keywords: List<String> = emptyList(),
)

@Serializable
data class AccessibilityPackageOverride(
    @SerialName("package_names")
    val packageNames: List<String> = emptyList(),
    @SerialName("ocr_fallback_enabled")
    val ocrFallbackEnabled: Boolean? = null,
    @SerialName("ocr_attempt_cooldown_ms")
    val ocrAttemptCooldownMs: Long? = null,
    @SerialName("min_sensitive_score")
    val minSensitiveScore: Int? = null,
    @SerialName("global_cooldown_ms")
    val globalCooldownMs: Long? = null,
    @SerialName("same_screen_cooldown_ms")
    val sameScreenCooldownMs: Long? = null,
    @SerialName("max_captured_texts")
    val maxCapturedTexts: Int? = null,
    @SerialName("max_text_length")
    val maxTextLength: Int? = null,
    @SerialName("multi_group_bonus_threshold")
    val multiGroupBonusThreshold: Int? = null,
    @SerialName("multi_group_bonus")
    val multiGroupBonus: Int? = null,
    @SerialName("link_indicator_bonus")
    val linkIndicatorBonus: Int? = null,
    @SerialName("verification_code_bonus")
    val verificationCodeBonus: Int? = null,
    @SerialName("link_indicators")
    val linkIndicators: List<String>? = null,
    @SerialName("keyword_groups")
    val keywordGroups: List<AccessibilityKeywordGroup>? = null,
)

data class ResolvedAccessibilityAutoAnalyzeConfig(
    val debugToastEnabled: Boolean,
    val debugToastMinIntervalMs: Long,
    val ocrFallbackEnabled: Boolean,
    val ocrAttemptCooldownMs: Long,
    val minSensitiveScore: Int,
    val globalCooldownMs: Long,
    val sameScreenCooldownMs: Long,
    val maxCapturedTexts: Int,
    val maxTextLength: Int,
    val multiGroupBonusThreshold: Int,
    val multiGroupBonus: Int,
    val linkIndicatorBonus: Int,
    val verificationCodeBonus: Int,
    val ignoredPackagePrefixes: List<String>,
    val ignoredPackageKeywords: List<String>,
    val linkIndicators: List<String>,
    val keywordGroups: List<AccessibilityKeywordGroup>,
)

fun AccessibilityAutoAnalyzeConfig.resolveForPackage(packageName: String): ResolvedAccessibilityAutoAnalyzeConfig {
    val override = packageOverrides.firstOrNull { rule -> rule.packageNames.any { it == packageName } }
    return ResolvedAccessibilityAutoAnalyzeConfig(
        debugToastEnabled = debugToastEnabled,
        debugToastMinIntervalMs = debugToastMinIntervalMs,
        ocrFallbackEnabled = override?.ocrFallbackEnabled ?: ocrFallbackEnabled,
        ocrAttemptCooldownMs = override?.ocrAttemptCooldownMs ?: ocrAttemptCooldownMs,
        minSensitiveScore = override?.minSensitiveScore ?: minSensitiveScore,
        globalCooldownMs = override?.globalCooldownMs ?: globalCooldownMs,
        sameScreenCooldownMs = override?.sameScreenCooldownMs ?: sameScreenCooldownMs,
        maxCapturedTexts = override?.maxCapturedTexts ?: maxCapturedTexts,
        maxTextLength = override?.maxTextLength ?: maxTextLength,
        multiGroupBonusThreshold = override?.multiGroupBonusThreshold ?: multiGroupBonusThreshold,
        multiGroupBonus = override?.multiGroupBonus ?: multiGroupBonus,
        linkIndicatorBonus = override?.linkIndicatorBonus ?: linkIndicatorBonus,
        verificationCodeBonus = override?.verificationCodeBonus ?: verificationCodeBonus,
        ignoredPackagePrefixes = ignoredPackagePrefixes,
        ignoredPackageKeywords = ignoredPackageKeywords,
        linkIndicators = override?.linkIndicators ?: linkIndicators,
        keywordGroups = override?.keywordGroups ?: keywordGroups,
    )
}

object AccessibilityAutoAnalyzeConfigLoader {
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        explicitNulls = false
    }

    fun load(context: Context): AccessibilityAutoAnalyzeConfig {
        return runCatching {
            context.assets.open("config/accessibility_auto_analyze_config.json").bufferedReader().use { reader ->
                json.decodeFromString<AccessibilityAutoAnalyzeConfig>(reader.readText())
            }
        }.getOrElse { AccessibilityAutoAnalyzeConfig() }
    }
}
