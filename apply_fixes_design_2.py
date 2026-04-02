import re

file_path = r"d:\Work\AntiFraud-AI-Assistant\competition_docs\专利\设计文档\文档\反诈卫士（Sensible AI）_程序设计说明书_V1.0.md"

with open(file_path, "r", encoding="utf-8") as f:
    content = f.read()

# 1) 13.1 Deployment specific lines replacement
chunk_13_1 = """其中，前端与 Android 端的部署形态说明如下：

1. 管理员端前端位于 `frontend/desktop-vue`，用于案件审核、知识库管理、统计分析、图谱分析和地图态势展示。  
2. 用户侧移动前端位于 `frontend/mobile-vue`，用于分析提交、风险历史、趋势查看、聊天咨询、家庭中心和模拟训练等功能。  
3. Android 原生端当前按安装包形式交付，部署时只需提供 APK 文件供安装，不要求在软著材料中展开完整 Android 工程构建过程。  
4. Android APK 主要承担悬浮触发、后台提醒、无障碍守护、OCR 快路径识别以及与后端接口联动的终端能力。  
5. 系统源码统一托管于 GitHub 仓库 `https://github.com/Yyzfddsdf/AntiFraud-AI-Assistant`，部署时可先从该仓库获取后端与前端代码，再按环境要求分别启动。  """

chunk_13_1_new = """其中，前端与 Android 端的部署形态说明如下：

1. 管理员端前端、移动端前端、Android APK 分别部署并通过后端接口联调。  
2. 管理员端前端用于案件审核、知识库管理、统计分析、图谱分析和地图态势展示。  
3. 移动端前端用于分析提交、风险历史、趋势查看、聊天咨询、家庭中心和模拟训练等功能。  
4. Android APK 主要承担悬浮触发、后台提醒、无障碍守护、OCR 快路径识别以及与后端接口联动的终端能力。  """
content = content.replace(chunk_13_1, chunk_13_1_new)

# 13.2 Deployment specifics replacement
chunk_13_2 = """前端与 APK 的启动方式可进一步说明为：

1. 管理员端前端在 `frontend/desktop-vue` 目录执行 `npm install` 和 `npm run dev`，默认开发地址为 `http://localhost:5173`。  
2. 用户侧移动前端在 `frontend/mobile-vue` 目录执行 `npm install` 和 `npm run dev`，默认开发地址为 `http://localhost:5174`。  
3. 两个前端工程均通过同源代理或开发代理访问后端 `/api`。  
4. Android 端不参与本仓库内构建流程说明，部署时只需说明“安装 APK 后，通过后端接口地址完成联调”即可。  
5. 若从 GitHub 仓库获取代码，可直接使用仓库中的后端目录与前端目录进行启动；Android 端则按单独 APK 安装包形式分发。"""

chunk_13_2_new = """前端与 APK 的运行部署可进一步说明为：

管理员端前端、移动端前端、Android APK 分别部署并通过后端接口联调。前端应用可通过网关或反向代理连接后端服务；Android 端在安装完成后，直接配置后端服务器地址即可接入使用。"""
content = content.replace(chunk_13_2, chunk_13_2_new)

# 2) 14.1 Replaced text
content = content.replace(
    "结论：各模块交互逻辑严密，输入输出符合设计预期，测试结果表明，各模块输入输出符合设计预期，核心业务流程运行稳定。",
    "结论：各模块交互逻辑清晰，输入输出符合设计预期，核心业务流程运行稳定。"
)

# 3) 15.1 Replaced text
content = content.replace(
    "其中，Android 原生端后台守护、悬浮球、无障碍与 OCR 快路径部分，因当前仓库未直接包含完整 Android 源码，其中，Android 原生端后台守护、悬浮球、无障碍与 OCR 快路径部分，因当前仓库未直接包含完整 Android 源码，本文依据现有终端设计方案与系统能力进行整理，并作为整体程序设计说明的一部分描述。",
    "其中，Android 原生端后台守护、悬浮球、无障碍与 OCR 快路径部分，因当前仓库未直接包含完整 Android 源码，本文依据现有终端设计方案与系统能力进行整理，并作为整体程序设计说明的一部分描述。"
)

# 4) ER diagram numbering from 8.6 to 8.8
content = content.replace(
    "### 8.6 ER 图占位\n\n![图 9 数据库 ER 图](../svg/db_er_diagram.svg)",
    "### 8.8 ER 图占位\n\n![图 9 数据库 ER 图](../svg/db_er_diagram.svg)"
)

# 5) Section 2.2 and 2.4 Replacements
content = content.replace("移动端采用独立移动前端工程，并依据项目方案对应 Android 原生端设计", "移动端采用独立移动前端工程，并对应 Android 原生端设计")
content = content.replace("依据项目文档补充的 Android 原生端设计负责无障碍守护、悬浮球与后台提醒", "Android 原生端设计负责无障碍守护、悬浮球与后台提醒")

# 6) 4.3 Appending to the conclusion
content = content.replace(
    "该流程确保了极高的识别速度，并大幅降低了高频调用的大模型计算开销，实现低功耗设计。",
    "该流程确保了较快的识别速度，并大幅降低了高频调用的大模型计算开销。该流程以本地文本提取、加权评分和阈值判定为核心，优先满足低时延、低功耗和可解释性要求。"
)

with open(file_path, "w", encoding="utf-8") as f:
    f.write(content)
