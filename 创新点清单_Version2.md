# AntiFraud AI Assistant 创新点清单

## 🟢 简单

### 1. 用户风险趋势
- **实现方式**：遍历 `GetCaseHistory()`，按时间排序聚合每条记录的风险等级，新增 `GET /api/user/risk-trend` 接口返回趋势数组和趋势方向（上升/下降/平稳）

### 2. 案件时序热力图
- **实现方式**：对 `historical_case_library` 表按 `created_at` 月份 + `scam_type` 做 SQL 聚合查询，新增管理员接口 `GET /api/scam/case-library/stats/heatmap` 返回各月各类型案件数量

### 3. 防诈知识库问答
- **实现方式**：在 Chat 工具列表中加入 `search_similar_cases` 工具，更新系统提示词，使模型在用户询问诈骗知识时主动调用案件库检索并以知识问答形式回答

---

## 🟡 中等

### 4. 诈骗链路还原
- **实现方式**：在 `submit_final_report` 的 JSON schema 中新增 `attack_steps: []string` 字段，更新主智能体提示词要求输出骗局步骤，前端渲染成时间线展示

### 5. 关键词风险标注
- **实现方式**：在 `submit_final_report` schema 中新增 `risk_tokens: [{text, risk_level}]` 字段，主智能体分析完成后对原文中的高危词汇进行标注并返回，前端高亮渲染

### 6. 多轮对话反诈引导
- **实现方式**：新增一套引导模式系统提示词，当 Chat 检测到高危关键词时切换为主动追问模式，依次引导用户回答"对方如何联系你"、"是否要求下载 APP"、"是否提到安全账户"等关键问题

### 7. 用户行为异常检测 + WebSocket 实时警报
- **实现方式**：新增独立的 `GET /api/alert/ws` WebSocket 接口，连接存在时启动 goroutine 每 30 秒轮询一次 `history_cases` 表，触发条件（如1小时内高风险记录 ≥ 2 次）时主动推送警报，断开连接后 goroutine 自动退出，不改动任何现有分析接口

---

## 🔴 较复杂

### 8. 反诈知识图谱
- **实现方式**：新建关联表存储诈骗类型之间的关系（关联类型、常见话术、目标人群、升级路径），管理员可通过接口维护图谱数据，`search_similar_cases` 检索结果中附带"该类型常见升级方向"，让智能体在分析时能预判骗局下一步走向