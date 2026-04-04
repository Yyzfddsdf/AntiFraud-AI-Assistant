package com.example.myapplication

data class QuickAnalyzePresentation(
    val title: String,
    val body: String,
    val riskLevel: String,
)

fun QuickAnalyzePresentation.shouldNotifyForAutoTrigger(): Boolean {
    return riskLevel == "中" || riskLevel == "高"
}

fun buildQuickAnalyzePresentation(
    result: QuickAnalyzeResponse,
    contextPrefix: String = "",
): QuickAnalyzePresentation {
    val riskLevel = when {
        result.risk_level.contains("高") -> "高"
        result.risk_level.contains("中") -> "中"
        else -> "低"
    }
    val title = if (riskLevel == "低") {
        "当前风险等级较低，安全"
    } else {
        "当前疑似存在高风险因素，请进入软件进行深度分析"
    }
    val fallbackReason = if (riskLevel == "低") {
        "当前画面未识别到明显的诱导、仿冒或转账风险线索。"
    } else {
        "当前画面存在可疑诱导、仿冒或风险暗示，请继续在应用内做深度分析。"
    }
    val reason = result.reason.ifBlank { fallbackReason }
    val body = if (contextPrefix.isBlank()) {
        reason
    } else {
        "$contextPrefix$reason"
    }
    return QuickAnalyzePresentation(
        title = title,
        body = body,
        riskLevel = riskLevel,
    )
}
