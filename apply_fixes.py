import sys
import re

file_path = r"d:\Work\AntiFraud-AI-Assistant\competition_docs\专利\使用文档\文档\反诈卫士（Sensible AI）_用户操作说明书_V1.0.md"

with open(file_path, "r", encoding="utf-8") as f:
    content = f.read()

# 1) 首页和导航
content = content.replace("根据当前实现和演示前端，首页通常会包含以下区域：", "首页主要包含以下区域：")
content = content.replace("根据当前实现和演示前端，首页通常会包含以下区域", "首页主要包含以下区域")
content = content.replace("当前用户侧底部导航一般包含以下入口：", "底部导航包含以下入口：")
content = content.replace("当前用户侧底部导航一般包含以下入口", "底部导航包含以下入口")

# 3) 头部和 1.2
content = content.replace("适用对象：普通用户、家庭守护用户、比赛评审人员、演示讲解人员", "适用对象：普通用户、家庭守护用户、系统体验人员")
content = content.replace("适用终端：Android 端、移动端演示前端、与用户侧相关的系统功能界面", "适用终端：Android 端及与用户侧相关的系统功能界面")
content = content.replace("8. 比赛演示人员依据用户视角展示产品能力。", "8. 系统功能体验与操作演示场景。")

# 4) 统一替换 “通常 / 一般 / 可能”
content = content.replace("首次启动时，通常会看到欢迎页或登录页", "首次启动时，会看到欢迎页或登录页")
content = content.replace("登录页通常包含以下区域", "登录页包含以下区域")
content = content.replace("注册成功后，通常会出现以下一种或多种反馈", "注册成功后，会出现以下一种或多种反馈")
content = content.replace("历史档案区域通常显示", "历史档案区域显示")
content = content.replace("结果页一般由以下部分组成", "结果页由以下部分组成")
content = content.replace("系统可能主动提醒", "系统会主动提醒")
content = content.replace("地区风险页一般可能展示", "地区风险页展示")
content = content.replace("进入方式通常包括", "进入方式包括")
content = content.replace("通常在以下情况下触发", "在以下情况下触发")

content = content.replace("通常会发生", "会发生")
content = content.replace("通常会分为", "会分为")
content = content.replace("一般包含", "包含")
content = content.replace("一般会显示", "会显示")
content = content.replace("可能有以下", "包含以下")
content = content.replace("进入方式通常如下", "进入方式如下")
content = content.replace("系统可能提示", "系统会提示")

# 6) Shortening sections
content = re.sub(
    r"### 2.2 软件定位.*?### 2.3",
    "### 2.2 软件定位\n\n本软件是一套面向普通公众的智能反诈系统，提供核心防护与辅助判断能力：\n\n1. 帮助识别可疑内容。\n2. 在确认前提供辅助判断。\n3. 保存历史记录用于长期复盘。\n4. 高风险场景下主动提醒。\n5. 在需要时通知家庭守护人。\n6. 通过模拟训练提升防骗意识。\n\n### 2.3",
    content, flags=re.DOTALL
)

content = re.sub(
    r"### 2.5 使用价值.*?---",
    "### 2.5 使用价值\n\n本软件的核心价值为：\n\n1. 降低紧张场景下的判断压力，将复杂风险转化为易读结论。\n2. 将个人防护扩展为长期的历史记忆与家庭协同防护机制。\n3. 实现从事后补救到事中提醒与事前训练的防范前移。\n\n---",
    content, flags=re.DOTALL
)

content = re.sub(
    r"### 7.1 为什么建议先走一遍完整流程.*?### 7.2",
    "### 7.1 为什么建议先走一遍完整流程\n\n为确保在真正的高风险紧急场景中能迅速操作，建议您首次使用时在不着急的状态下预先完成一次完整流程体验和必要权限配置。\n\n### 7.2",
    content, flags=re.DOTALL
)

content = re.sub(
    r"### 8.1 分析功能的总体作用.*?### 8.2",
    "### 8.1 分析功能的总体作用\n\n分析功能是系统的核心能力，旨在将可疑内容交给系统智能识别，直接输出普通用户易于理解的判断结论。\n\n### 8.2",
    content, flags=re.DOTALL
)

content = re.sub(
    r"### 14.1 结果页的重要性.*?### 14.2",
    "### 14.1 结果页的作用\n\n结果页用于直观解答“怎么看”的问题，展示风险等级、原因和操作指引，是阅读和理解风险的核心页面。\n\n### 14.2",
    content, flags=re.DOTALL
)

content = re.sub(
    r"### 17.1 风险趋势分析的作用.*?### 17.2",
    "### 17.1 风险趋势分析的作用\n\n帮助用户直观了解近期接触到的风险信息变化趋势，判断风险频率是在增多还是减少。\n\n### 17.2",
    content, flags=re.DOTALL
)

content = re.sub(
    r"### 18.1 地区风险信息的作用.*?### 18.2",
    "### 18.1 地区风险信息的作用\n\n地区风险信息帮助用户快速感知本地近期高发骗局，建立对本地风险态势的共识与防范意识。\n\n### 18.2",
    content, flags=re.DOTALL
)

content = re.sub(
    r"### 23.2 后台守护的主要价值.*?### 23.3",
    "### 23.2 后台守护的主要价值\n\n能够在用户意外访问高危网页或遭遇风险操作时，提供即时的拦截与警示。\n\n### 23.3",
    content, flags=re.DOTALL
)


# 5) Renumber chapters 33 to 40 down by 1
def renumber_chapter(match):
    prefix = match.group(1)
    ch_num = int(match.group(2))
    suffix = match.group(3)
    if ch_num >= 33:
        return f"{prefix}{ch_num - 1}{suffix}"
    return match.group(0)

# This targets `## 33.` and `### 33.1.` and `#### 33.1.1.`
content = re.sub(r"^(#{2,4}\s+)(\d+)(\.)", renumber_chapter, content, flags=re.MULTILINE)

# Some formatting cleanup explicitly for old Chapter 39 which had bullet points with numbers like 1., 2..
# They should not be affected by `^(#{2,4}\s+)` since they lack `#`.

with open(file_path, "w", encoding="utf-8") as f:
    f.write(content)

print("Replacement done.")
