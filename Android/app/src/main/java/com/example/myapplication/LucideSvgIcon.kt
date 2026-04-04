package com.example.myapplication

import android.graphics.Bitmap
import android.graphics.Canvas as AndroidCanvas
import androidx.compose.foundation.Image
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.caverock.androidsvg.SVG
import java.util.Locale
import kotlin.math.roundToInt

@Composable
fun LucideSvgIcon(
    iconName: String,
    contentDescription: String?,
    tint: Color,
    size: Dp = 24.dp,
    modifier: Modifier = Modifier,
) {
    val context = LocalContext.current
    val bitmap = remember(iconName, tint) {
        runCatching {
            val svgText = context.assets.open("lucide/$iconName.svg").bufferedReader(Charsets.UTF_8).use { it.readText() }
            val tintedSvg = svgText.replace("currentColor", colorToCssHex(tint))
            val svg = SVG.getFromString(tintedSvg)
            val outputSize = 96
            val bitmap = Bitmap.createBitmap(outputSize, outputSize, Bitmap.Config.ARGB_8888)
            val canvas = AndroidCanvas(bitmap)
            svg.setDocumentWidth("100%")
            svg.setDocumentHeight("100%")
            svg.renderToCanvas(canvas)
            bitmap.asImageBitmap()
        }.getOrNull()
    }

    if (bitmap != null) {
        Image(
            bitmap = bitmap,
            contentDescription = contentDescription,
            modifier = modifier.size(size),
            contentScale = ContentScale.Fit,
        )
    }
}

private fun colorToCssHex(color: Color): String {
    val red = (color.red * 255).roundToInt().coerceIn(0, 255)
    val green = (color.green * 255).roundToInt().coerceIn(0, 255)
    val blue = (color.blue * 255).roundToInt().coerceIn(0, 255)
    val alpha = color.alpha.coerceIn(0f, 1f)

    return if (alpha >= 0.999f) {
        String.format("#%02X%02X%02X", red, green, blue)
    } else {
        String.format(Locale.US, "rgba(%d,%d,%d,%.3f)", red, green, blue, alpha)
    }
}
