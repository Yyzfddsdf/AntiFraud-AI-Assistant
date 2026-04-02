import json
import re

file_path = r'd:\Work\AntiFraud-AI-Assistant\competition_docs\专利\文档\反诈卫士（Sensible AI）_程序设计说明书_V1.0.md'

replacements = {
    '【插图占位：图 1 系统总体架构图】': '![图 1 系统总体架构图](../svg/sys_architecture.svg)',
    '【插图占位：图 2 系统功能结构图】': '![图 2 系统功能结构图](../svg/func_structure.svg)',
    '【插图占位：图 3 普通用户完整分析流程图】': '![图 3 普通用户完整分析流程图](../svg/user_flow_full.svg)',
    '【插图占位：图 4 快路径识别流程图】': '![图 4 快路径识别流程图](../svg/fast_path_flow.svg)',
    '【插图占位：图 5 后端模块架构图】': '![图 5 后端模块架构图](../svg/backend_module.svg)',
    '【插图占位：图 6 Android 端模块架构图】': '![图 6 Android 端模块架构图](../svg/android_module.svg)',
    '【插图占位：图 7 多智能体协同流程图】': '![图 7 多智能体协同流程图](../svg/multi_agent_flow.svg)',
    '【插图占位：图 8 数据库 ER 图】': '![图 8 数据库 ER 图](../svg/db_er_diagram.svg)'
}

with open(file_path, 'r', encoding='utf-8') as f:
    content = f.read()

# Replace placeholders
for old_text, new_text in replacements.items():
    content = content.replace(old_text, new_text)

# We also need to remove the lines starting with 【图注建议：
content = re.sub(r'【图注建议：.*?】\n?', '', content)

with open(file_path, 'w', encoding='utf-8') as f:
    f.write(content)

print('Replaced successfully.')
