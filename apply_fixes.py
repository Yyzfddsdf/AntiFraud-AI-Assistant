import re

file_path = r"d:\Work\AntiFraud-AI-Assistant\competition_docs\专利\使用文档\文档\反诈卫士（Sensible AI）_用户操作说明书_V1.0.md"

with open(file_path, "r", encoding="utf-8") as f:
    content = f.read()

# Remove the 8th item from 1.2
content = content.replace("7. 用户参与 AI 反诈模拟训练。\n8. 系统功能体验与操作演示场景。", "7. 用户参与 AI 反诈模拟训练。")

# Remove all remaining "通常" from the whole document
content = content.replace("通常会", "会")
content = content.replace("通常在", "在")
content = content.replace("通常可", "可")
content = content.replace("通常通过", "通过")
content = content.replace("通常情况", "情况")
content = content.replace("通常于", "于")
content = content.replace("通常是", "是")
content = content.replace("通常", "") # catch-all for any remaining

with open(file_path, "w", encoding="utf-8") as f:
    f.write(content)

print("Remaining issues fixed!")
