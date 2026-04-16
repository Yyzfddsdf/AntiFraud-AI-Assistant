package com.example.myapplication

import android.content.Context
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class AppConfig(
    @SerialName("api_origin")
    val apiOrigin: String = "http://10.0.2.2:8081",
    @SerialName("default_screen")
    val defaultScreen: String = "dashboard",
)

object AppConfigLoader {
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        explicitNulls = false
    }

    fun load(context: Context): AppConfig {
        return runCatching {
            context.assets.open("config/app_config.json").bufferedReader().use { reader ->
                json.decodeFromString<AppConfig>(reader.readText())
            }
        }.getOrElse { AppConfig() }
    }
}

fun String.toAppScreen(): AppScreen = when (trim().lowercase()) {
    "history" -> AppScreen.History
    "risk_trend", "risktrend" -> AppScreen.RiskTrend
    "alerts" -> AppScreen.Alerts
    "submit" -> AppScreen.Submit
    "simulation_quiz", "simulationquiz" -> AppScreen.SimulationQuiz
    "chat" -> AppScreen.Chat
    "family" -> AppScreen.Family
    "family_manage", "familymanage" -> AppScreen.FamilyManage
    "profile" -> AppScreen.Profile
    "profile_privacy", "profileprivacy" -> AppScreen.ProfilePrivacy
    else -> AppScreen.Dashboard
}
