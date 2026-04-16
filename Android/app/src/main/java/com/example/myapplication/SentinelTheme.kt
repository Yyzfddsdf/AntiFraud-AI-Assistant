package com.example.myapplication

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.colorResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

private val PlusJakartaSans = FontFamily(
    Font(R.font.plus_jakarta_sans_400, FontWeight.Normal),
    Font(R.font.plus_jakarta_sans_500, FontWeight.Medium),
    Font(R.font.plus_jakarta_sans_600, FontWeight.SemiBold),
    Font(R.font.plus_jakarta_sans_700, FontWeight.Bold),
    Font(R.font.plus_jakarta_sans_800, FontWeight.ExtraBold),
)

private val SentinelTypography = androidx.compose.material3.Typography(
    headlineLarge = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.ExtraBold, fontSize = 30.sp),
    headlineMedium = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.Bold, fontSize = 24.sp),
    titleLarge = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.Bold, fontSize = 20.sp),
    titleMedium = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.SemiBold, fontSize = 16.sp),
    bodyLarge = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.Normal, fontSize = 15.sp),
    bodyMedium = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.Normal, fontSize = 14.sp),
    bodySmall = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.Normal, fontSize = 12.sp),
    labelLarge = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.SemiBold, fontSize = 14.sp),
    labelMedium = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.Medium, fontSize = 12.sp),
    labelSmall = TextStyle(fontFamily = PlusJakartaSans, fontWeight = FontWeight.Bold, fontSize = 10.sp),
)

@Composable
fun SentinelTheme(content: @Composable () -> Unit) {
    val isDark = isSystemInDarkTheme()
    val lightColors = lightColorScheme(
        primary = colorResource(R.color.sentinel_primary),
        onPrimary = Color.White,
        secondary = colorResource(R.color.sentinel_teal),
        surface = Color.White,
        surfaceVariant = colorResource(R.color.sentinel_surface_alt),
        background = colorResource(R.color.sentinel_surface),
        onBackground = colorResource(R.color.sentinel_text),
        onSurface = colorResource(R.color.sentinel_text),
        outline = colorResource(R.color.sentinel_border),
        error = colorResource(R.color.sentinel_danger),
    )
    val darkColors = darkColorScheme(
        primary = colorResource(R.color.sentinel_primary),
        secondary = colorResource(R.color.sentinel_teal),
        background = Color(0xFF0F172A),
        surface = Color(0xFF111827),
        surfaceVariant = Color(0xFF1E293B),
        onBackground = Color.White,
        onSurface = Color.White,
    )

    MaterialTheme(
        colorScheme = if (isDark) darkColors else lightColors,
        typography = SentinelTypography,
        content = content,
    )
}
