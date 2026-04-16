package com.example.myapplication

import android.content.Context
import android.content.Intent
import android.graphics.Color as AndroidColor
import android.net.Uri
import android.view.ViewGroup
import android.webkit.WebResourceRequest
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.isSpecified
import androidx.compose.ui.viewinterop.AndroidView
import kotlin.math.roundToInt
import org.json.JSONObject
import org.commonmark.Extension
import org.commonmark.ext.autolink.AutolinkExtension
import org.commonmark.ext.gfm.strikethrough.StrikethroughExtension
import org.commonmark.ext.gfm.tables.TablesExtension
import org.commonmark.parser.Parser
import org.commonmark.renderer.html.HtmlRenderer

private val markdownExtensions: List<Extension> = listOf(
    AutolinkExtension.create(),
    StrikethroughExtension.create(),
    TablesExtension.create(),
)

private val markdownParser: Parser by lazy {
    Parser.builder()
        .extensions(markdownExtensions)
        .build()
}

private val markdownHtmlRenderer: HtmlRenderer by lazy {
    HtmlRenderer.builder()
        .extensions(markdownExtensions)
        .escapeHtml(true)
        .sanitizeUrls(true)
        .softbreak("<br>")
        .build()
}

@Composable
fun MarkdownText(
    markdown: String,
    color: Color,
    modifier: Modifier = Modifier,
    style: androidx.compose.ui.text.TextStyle = androidx.compose.ui.text.TextStyle.Default,
    linkColor: Color = Color(0xFF059669),
) {
    val shellHtml = remember(color, style, linkColor) {
        markdownShellHtml(
            textColor = color,
            textSize = style.fontSize,
            linkColor = linkColor,
        )
    }
    val bodyHtml = remember(markdown) { markdownBodyHtml(markdown) }

    AndroidView(
        modifier = modifier,
        factory = { context ->
            MarkdownWebView(context).apply {
                setBackgroundColor(AndroidColor.TRANSPARENT)
                isVerticalScrollBarEnabled = false
                isHorizontalScrollBarEnabled = false
                overScrollMode = WebView.OVER_SCROLL_NEVER
                settings.apply {
                    javaScriptEnabled = true
                    domStorageEnabled = false
                    loadsImagesAutomatically = true
                    defaultTextEncodingName = "utf-8"
                    textZoom = 100
                    builtInZoomControls = false
                    displayZoomControls = false
                    setSupportZoom(false)
                }
                webViewClient = object : WebViewClient() {
                    override fun onPageFinished(view: WebView, url: String?) {
                        (view as? MarkdownWebView)?.onShellReady()
                    }

                    override fun shouldOverrideUrlLoading(
                        view: WebView,
                        request: WebResourceRequest,
                    ): Boolean {
                        return openExternalLink(view.context, request.url)
                    }

                    override fun shouldOverrideUrlLoading(view: WebView, url: String): Boolean {
                        return openExternalLink(view.context, Uri.parse(url))
                    }
                }
            }
        },
        update = { webView ->
            if (webView.currentShellHtml != shellHtml) {
                webView.loadShell(shellHtml)
            }
            webView.setBodyHtml(bodyHtml)
        },
    )
}

private fun markdownBodyHtml(markdown: String): String {
    return markdownHtmlRenderer.render(markdownParser.parse(markdown.ifBlank { "..." }))
}

private fun markdownShellHtml(
    textColor: Color,
    textSize: TextUnit,
    linkColor: Color,
): String {
    val bodyColor = htmlColor(textColor)
    val bodyFontSize = if (textSize.isSpecified) "${textSize.value}px" else "14px"
    val link = htmlColor(linkColor)

    return """
        <!DOCTYPE html>
        <html lang="zh-CN">
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
          <style>
            html, body {
              margin: 0;
              padding: 0;
              background: transparent;
            }
            body {
              color: $bodyColor;
              font-size: $bodyFontSize;
              line-height: 1.6;
              overflow-wrap: anywhere;
              word-break: break-all;
            }
            .markdown-body {
              color: $bodyColor;
            }
            .markdown-body h1 { font-size: 17px; font-weight: 700; margin: 14px 0 10px; color: #1e293b; }
            .markdown-body h2 { font-size: 15px; font-weight: 700; margin: 12px 0 8px; color: #1e293b; }
            .markdown-body h3 { font-size: 14px; font-weight: 600; margin: 10px 0 6px; color: #334155; }
            .markdown-body h4, .markdown-body h5, .markdown-body h6 { font-size: 14px; font-weight: 600; margin: 10px 0 6px; color: #334155; }
            .markdown-body p { margin: 7px 0; line-height: 1.6; }
            .markdown-body ul, .markdown-body ol { margin: 7px 0; padding-left: 18px; }
            .markdown-body li { margin: 4px 0; }
            .markdown-body code { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; font-family: monospace; font-size: 13px; color: #059669; }
            .markdown-body pre { background: #1e293b; padding: 12px; border-radius: 8px; overflow-x: auto; margin: 12px 0; }
            .markdown-body pre code { background: transparent; color: #e2e8f0; padding: 0; }
            .markdown-body blockquote { border-left: 3px solid #059669; padding-left: 12px; margin: 12px 0; color: #64748b; font-style: italic; }
            .markdown-body a { color: $link; text-decoration: underline; }
            .markdown-body table { width: 100%; border-collapse: collapse; margin: 12px 0; }
            .markdown-body th, .markdown-body td { border: 1px solid #e2e8f0; padding: 8px 12px; text-align: left; }
            .markdown-body th { background: #f8fafc; font-weight: 600; }
            .markdown-body hr { border: none; border-top: 1px solid #e2e8f0; margin: 16px 0; }
            .markdown-body strong { font-weight: 700; color: #1e293b; }
            .markdown-body em { font-style: italic; }
            .markdown-body img { max-width: 100%; height: auto; border-radius: 8px; }
            .markdown-body p, .markdown-body div, .markdown-body span, .markdown-body li { word-break: break-all; overflow-wrap: anywhere; }
          </style>
          <script>
            window.__setContent = function(html) {
              var container = document.getElementById('markdown-body');
              if (!container) return;
              container.innerHTML = html;
            };
          </script>
        </head>
        <body>
          <div id="markdown-body" class="markdown-body"></div>
        </body>
        </html>
    """.trimIndent()
}

private fun htmlColor(color: Color): String {
    return "#%06X".format(0xFFFFFF and color.toArgb())
}

private fun openExternalLink(context: Context, uri: Uri): Boolean {
    if (uri.scheme.isNullOrBlank()) return true
    return runCatching {
        context.startActivity(
            Intent(Intent.ACTION_VIEW, uri).apply {
                addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
            },
        )
        true
    }.getOrDefault(true)
}

private class MarkdownWebView(
    context: Context,
) : WebView(context) {
    var currentShellHtml: String? = null
        private set
    private var pendingBodyHtml: String = ""
    private var pageReady: Boolean = false

    init {
        layoutParams = ViewGroup.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT,
            ViewGroup.LayoutParams.WRAP_CONTENT,
        )
    }

    fun loadShell(shellHtml: String) {
        currentShellHtml = shellHtml
        pageReady = false
        loadDataWithBaseURL(null, shellHtml, "text/html", "utf-8", null)
    }

    fun setBodyHtml(bodyHtml: String) {
        pendingBodyHtml = bodyHtml
        if (pageReady) {
            applyBodyHtml()
        }
    }

    fun onShellReady() {
        pageReady = true
        applyBodyHtml()
    }

    private fun applyBodyHtml() {
        evaluateJavascript(
            "window.__setContent(${JSONObject.quote(pendingBodyHtml)});",
            null,
        )
        post {
            requestLayout()
            postDelayed({ requestLayout() }, 32L)
        }
    }

    override fun onMeasure(widthMeasureSpec: Int, heightMeasureSpec: Int) {
        super.onMeasure(widthMeasureSpec, heightMeasureSpec)
        val width = MeasureSpec.getSize(widthMeasureSpec)
        val resolvedHeight = if (contentHeight > 0) {
            (contentHeight * scale).roundToInt().coerceAtLeast(1)
        } else {
            measuredHeight.coerceAtLeast(1)
        }
        setMeasuredDimension(width, resolvedHeight)
    }
}
