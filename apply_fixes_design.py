import re

file_path = r"d:\Work\AntiFraud-AI-Assistant\competition_docs\专利\设计文档\文档\反诈卫士（Sensible AI）_程序设计说明书_V1.0.md"

with open(file_path, "r", encoding="utf-8") as f:
    content = f.read()

# 1)
content = content.replace("终端交互层设计，包括管理员端、移动端以及依据比赛方案整理的 Android 主动守护设计。", "终端交互层设计，包括管理员端、移动端及 Android 端主动守护能力设计。")

# 2)
content = content.replace("移动端采用独立工程，并依据比赛方案对应 Android 原生端设计", "移动端采用独立工程，并对应 Android 原生端设计，用于承接用户侧交互流程")

# 3)
content = content.replace("Android 原生端设计方案。依据比赛文档，Android 原生端主要负责", "Android 原生端设计方案。Android 原生端主要负责后台守护、悬浮球、无障碍监听、OCR 兜底、通知提醒和家庭联动入口等高频触发能力。")

# 4) Delete GitHub reference
content = re.sub(r"系统源码统一托管于 GitHub 仓库 https?://github\.com/[^\n]+\n?", "", content)

# 5) 13.2
content = re.sub(
    r"(### 13\.2 服务启动与部署模型.*?)(### 13\.3)",
    r"### 13.2 服务启动与部署模型\n\n管理员端前端与用户侧移动前端分别作为独立工程部署，通过开发代理或网关与后端 API 通信。Android 端以安装包形式交付，并通过后端接口完成联调。\n\n\2",
    content, flags=re.DOTALL
)

# 6) 15.1
content = re.sub(
    r"(### 15\.1 关键实现依据说明\n\n本文档的主要实现依据来自以下内容：\n\n).*?(?=\n\n本文依据)",
    r"\1本文档的主要实现依据来自现有后端源码结构、接口文档、数据库文档、终端设计方案以及前端工程模块划分。",
    content, flags=re.DOTALL
)

content = re.sub(
    r"本文依据比赛文档中已明确的终端设计方案进行整理[^\n]*",
    "其中，Android 原生端后台守护、悬浮球、无障碍与 OCR 快路径部分，因当前仓库未直接包含完整 Android 源码，本文依据现有终端设计方案与系统能力进行整理，并作为整体程序设计说明的一部分描述。",
    content
)

# 7) 1.2
content = re.sub(
    r"对于移动端高频风险场景，系统进一步设计了快路径识别机制.*?能力闭环。",
    "",
    content, flags=re.DOTALL
)
content = re.sub(r"\n{3,}", "\n\n", content)

# 8) 4.3
def compress_4_3(m):
    return """### 4.3 Android 后台守护加权判分流程

后台守护涉及无障碍页面文本获取与 OCR 兜底：

1. **快路径识别**：通过关键字提取工具对待检测文本进行提取，计算基准风险得分。
2. **场景因子修正**：根据页面类型（如转账页面）、应用特征等，进行权重修正。
3. **决策与拦截**：短时间内达到高风险阈值触发静默期并记录；未达阈值则缓存。

该流程确保了极高的识别速度，并大幅降低了高频调用的大模型计算开销，实现低功耗设计。

"""

content = re.sub(
    r"### 4\.3 Android 后台守护加权判分流程.*?### 4\.4",
    compress_4_3(None) + r"### 4.4",
    content, flags=re.DOTALL
)

# 9) 5.2
def compress_5_2(m):
    return """### 5.2 六边形架构模型设计

本项目后端由于采用了六边形架构，整体模块切分：

1. **领域规则和业务概念优先**：分析服务、家庭联动规则作为系统核心不受外部资源影响。
2. **边缘接口作为适配器**：HTTP、数据库、缓存、模型服务都作为适配器存在。

这样隔离之后，核心逻辑完全不绑定第三方应用，从而实现了设计层面的高内聚低耦合。

"""

content = re.sub(
    r"### 5\.2 六边形架构模型说明.*?### 5\.3",
    compress_5_2(None) + r"### 5.3",
    content, flags=re.DOTALL
)
content = re.sub(
    r"### 5\.2 六边形架构说明.*?### 5\.3",
    compress_5_2(None) + r"### 5.3",
    content, flags=re.DOTALL
)

# 10) 14
content = content.replace("未发现关键逻辑缺陷。", "测试结果表明，各模块输入输出符合设计预期，核心业务流程运行稳定。")
content = content.replace("具备极高的可信度。", "整体识别结果具有较高稳定性和可用性。")
content = content.replace("业务无感知的“自愈”率达 100%。", "相关异常场景下系统能够完成自动恢复，测试中恢复成功率为 100%。")
content = content.replace("业务无感知的自愈率达 100%。", "相关异常场景下系统能够完成自动恢复，测试中恢复成功率为 100%。")
content = content.replace("可作为软件著作权登记及生产部署的可靠鉴别材料。", "测试结果表明，系统已基本达到设计目标，可支持后续软件著作权登记材料整理、系统归档与部署参考。")
content = content.replace("作为软件著作权登记及生产部署的可靠鉴别材料", "测试结果表明，系统已基本达到设计目标，可支持后续软件著作权登记材料整理、系统归档与部署参考")

# 11) Replace arbitrary remaining terms
content = content.replace("极高的可信度", "较高的稳定性和可用性")
content = content.replace("未发现关键逻辑缺陷", "核心业务流程运行稳定")
content = content.replace("可靠鉴别材料", "系统归档与部署参考材料")
content = content.replace("依据比赛方案", "依据项目方案")
content = content.replace("依据比赛文档", "依据项目文档")
content = content.replace("比赛项目", "研发项目")
content = content.replace("比赛文档", "项目文档")
content = content.replace("比赛方案", "产品设计方案")
content = content.replace("比赛评审", "系统评审")
content = content.replace("比赛演示", "成果演示")
content = content.replace("比赛", "项目")

with open(file_path, "w", encoding="utf-8") as f:
    f.write(content)
