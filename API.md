# API 文档（登录系统）

## 基础信息

- 服务默认地址：`http://localhost:8081`
- 统一前缀：`/api`
- 数据格式：`application/json`
- 鉴权方式：`Authorization: Bearer <token>`

---

## 1) 获取验证码

- **Method**: `GET`
- **Path**: `/api/auth/captcha`
- **Header**（可选）:
  - `Accept: application/json`
- **说明**:
  - 返回花色 SVG 验证码（Data URL）
  - 验证码有效期 3 分钟
  - 验证码一次性使用（校验后即失效）

### 响应示例

```json
{
  "captchaId": "b6f2f9d5f0c64a0d8a53f8a1",
  "captchaImage": "data:image/svg+xml;utf8,<svg ...>",
  "expiresIn": 180
}
```

---

## 2) 用户注册

- **Method**: `POST`
- **Path**: `/api/auth/register`
- **Header**:
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "username": "test_user",
  "email": "test_user@example.com",
  "password": "Test@1234",
  "captchaId": "b6f2f9d5f0c64a0d8a53f8a1",
  "captchaCode": "AB7K9"
}
```

### 校验规则

- `username` 必填
- `email` 必填，且必须是邮箱格式
- `password` 必填，且需满足：
  - 至少一个大写字母
  - 至少一个小写字母
  - 至少一个符号
- `captchaId` / `captchaCode` 必填且必须匹配
- 用户年龄在注册时默认写入 `28`（无需在请求体传入 `age`）。

### 成功响应（201）

```json
{
  "id": 1,
  "username": "test_user",
  "email": "test_user@example.com",
  "age": 28,
  "role": "user"
}
```

### 常见失败响应

- `400` 请求参数错误 / 密码不满足复杂度 / 验证码错误或过期
- `409` 邮箱或用户名已存在

---

## 3) 用户登录

- **Method**: `POST`
- **Path**: `/api/auth/login`
- **Header**:
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "email": "test_user@example.com",
  "password": "Test@1234",
  "captchaId": "b6f2f9d5f0c64a0d8a53f8a1",
  "captchaCode": "AB7K9"
}
```

### 成功响应（200）

```json
{
  "message": "登录成功",
  "token": "<JWT_TOKEN>",
  "user": {
    "id": 1,
    "username": "test_user",
    "email": "test_user@example.com",
    "role": "user",
    "age": 28
  }
}
```

### 常见失败响应
- `401` 邮箱或密码不正确

---

## 4) 获取当前用户（需鉴权）
- **Method**: `GET`
- **Path**: `/api/user`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 成功响应（200）

```json
{
  "id": 1,
  "username": "test_user",
  "email": "test_user@example.com",
  "role": "user",
  "age": 28
}
```

### 常见失败响应

- `401` 未提供 Token / Token 无效或过期 / 用户不存在或已删除 / 用户信息不匹配

---

## 5) 删除当前用户（需鉴权）

- **Method**: `DELETE`
- **Path**: `/api/user`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 成功响应（200）

```json
{
  "message": "用户已删除"
}
```

### 常见失败响应

- `401` 用户未认证
- `500` 删除用户失败

---

## 6) 提交多模态诈骗分析任务（需鉴权）

- **Method**: `POST`
- **Path**: `/api/scam/multimodal/analyze`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "text": "我接到自称客服的电话，说我开通了会员需要转账取消",
  "videos": ["<video_base64_1>", "<video_base64_2>"],
  "audios": ["<audio_base64_1>"],
  "images": ["<image_base64_1>"]
}
```

### 说明

- `text/videos/audios/images` 至少提供一种输入。
- `videos/audios/images` 数组元素为对应文件的 Base64 字符串。

### 成功响应（202）

```json
{
  "task_id": "TASK-7FA12BC09D11",
  "status": "pending",
  "message": "任务已入队，后台处理中，请通过查询接口获取状态与结果"
}
```

### 常见失败响应

- `400` 请求参数错误 / 未提供任何可分析输入
- `503` 队列繁忙，任务入队失败

---

## 7) 查询当前用户进行中任务（需鉴权）

- **Method**: `GET`
- **Path**: `/api/scam/multimodal/tasks`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 成功响应（200）

```json
{
  "user_id": "1",
  "tasks": [
    {
      "task_id": "TASK-7FA12BC09D11",
      "user_id": "1",
      "title": "疑似冒充客服退款",
      "status": "processing",
      "summary": "",
      "created_at": "2026-02-20T10:30:00+08:00",
      "updated_at": "2026-02-20T10:30:08+08:00"
    }
  ]
}
```

> 说明：此接口仅返回状态为 `pending` 或 `processing` 的任务。已完成的任务请在历史记录中查询。

---

## 8) 查询当前用户历史案件列表（需鉴权）

- **Method**: `GET`
- **Path**: `/api/scam/multimodal/history`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 成功响应（200）

```json
{
  "user_id": "1",
  "history": [
    {
      "record_id": "TASK-123456",
      "title": "疑似冒充客服退款",
      "case_summary": "对方要求转入安全账户",
      "scam_type": "冒充客服类",
      "risk_level": "高",
      "created_at": "2026-02-20T10:35:00+08:00"
    }
  ]
}
```

> 说明：
> 1. 此接口仅返回历史案件的元数据（轻量级），不包含详细报告和原始文件。
> 2. 如需查看详情，请使用 `GET /api/scam/multimodal/tasks/:taskId` 接口。

---

## 8.1) 查询当前用户风险总览（需鉴权）

- **Method**: `GET`
- **Path**: `/api/scam/multimodal/history/overview`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### Query 参数

- `interval`（可选）：时间聚合粒度，支持 `day`、`week`、`month`，默认 `day`。

### 成功响应（200）

```json
{
  "stats": {
    "high": 3,
    "medium": 5,
    "low": 2,
    "total": 10
  },
  "trend": [
    {
      "time_bucket": "2026-03-01",
      "high": 1,
      "medium": 2,
      "low": 0,
      "total": 3
    },
    {
      "time_bucket": "2026-03-02",
      "high": 2,
      "medium": 1,
      "low": 1,
      "total": 4
    }
  ]
}
```

### 说明

- 该接口直接基于 `GetCaseHistory` 聚合，返回“风险变化趋势 + 高中低数量统计”两类总览信息。
- `time_bucket` 格式：
  - `day`：`YYYY-MM-DD`
  - `week`：`YYYY-Www`（ISO 周，例如 `2026-W10`）
  - `month`：`YYYY-MM`

### 常见失败响应

- `400` `interval` 非法（仅支持 `day/week/month`）
- `401` 未认证

---

## 8.2) 删除当前用户历史案件（需鉴权）

- **Method**: `DELETE`
- **Path**: `/api/scam/multimodal/history/:recordId`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 参数说明

- `recordId`：历史案件 ID，可从 `GET /api/scam/multimodal/history` 返回的 `record_id` 获取。
- 仅允许删除当前登录用户自己的历史案件。

### 成功响应（200）

```json
{
  "user_id": "1",
  "record_id": "TASK-123456",
  "message": "历史案件删除成功"
}
```

### 常见失败响应

- `400` `recordId` 为空
- `401` 未认证
- `404` 历史案件不存在或不属于当前用户
- `500` 删除失败

---

## 9) 更新当前用户年龄（需鉴权）

- **Method**: `PUT`
- **Path**: `/api/scam/multimodal/user/age`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "age": 28
}
```

### 说明

- `age` 取值范围：`1 ~ 150`
- 写入后会持久化到多模态状态 DB 的用户基础信息中。

### 成功响应（200）

```json
{
  "user_id": "1",
  "age": 28,
  "message": "年龄更新成功"
}
```

### 常见失败响应

- `400` 请求参数错误 / `age` 超出范围
- `401` 未认证
- `500` 年龄写入失败

---

## 10) 查询指定任务详情（需鉴权）

- **Method**: `GET`
- **Path**: `/api/scam/multimodal/tasks/:taskId`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 成功响应（200）- 单文件示例（单模态 1 条数据）

```json
{
  "task": {
    "task_id": "TASK-7B1A038E7452",
    "user_id": "5",
    "title": "TED学术演讲视频风险核查",
    "status": "completed",
    "scam_type": "冒充客服类",
    "summary": "该内容以公开演讲为主，未发现直接诈骗指令。",
    "created_at": "2026-02-20T23:57:30+08:00",
    "updated_at": "2026-02-20T23:57:45+08:00",
    "payload": {
      "text": "",
      "videos": ["<video_base64_1>"],
      "audios": [],
      "images": [],
      "video_insights": [
        "【整体视觉感受】\n画面为公开视频演讲场景，未见明显诱导操作。\n\n【关键信息提取】\n未出现“转账”“验证码”“下载指定 App”等指令。\n\n【可疑点清单】\n- 未发现明显可疑信号"
      ],
      "audio_insights": [],
      "image_insights": []
    },
    "report": "1. 综合摘要\n该内容以公开演讲为主，未发现直接诈骗指令。\n\n2. 多模态关键发现\n- 文本: 未提供文本输入\n- 图像: 未提供图像输入\n- 视频: 画面与语义一致，未见明显诈骗套路\n- 音频: 未提供音频输入\n\n3. 风险信号\n- 未发现明确风险信号\n\n4. 风险等级与理由\n- 风险等级: 低\n- 理由: 未出现诱导转账、索要敏感信息等关键风险特征\n\n5. 建议的下一步动作\n- 保留原始素材与上下文供后续复核\n\n6. 诈骗链路还原\n- 证据不足，暂无法还原完整诈骗链路"
  }
}
```

### 成功响应（200）- 多文件示例（同一模态多条数据）

```json
{
  "task": {
    "task_id": "TASK-9C4D2F71A0B3",
    "user_id": "5",
    "title": "批量视频线索核查",
    "status": "completed",
    "scam_type": "冒充客服类",
    "summary": "3 条视频中有 1-2 条出现明显风险信号，需人工复核。",
    "created_at": "2026-02-21T10:00:00+08:00",
    "updated_at": "2026-02-21T10:00:24+08:00",
    "payload": {
      "text": "",
      "videos": ["<video_base64_1>", "<video_base64_2>", "<video_base64_3>"],
      "audios": [],
      "images": [],
      "video_insights": [
        "【整体视觉感受】\n画面为交易聊天演示，存在催促动作。\n\n【关键信息提取】\n出现“立即转账”“私下联系”等指令型文案。\n\n【可疑点清单】\n1. 存在私下转账引导。\n2. 存在规避平台担保交易提示。",
        "【整体视觉感受】\n画面含账户切换和收款二维码展示。\n\n【关键信息提取】\n出现新的收款主体，与上下文不一致。\n\n【可疑点清单】\n1. 收款主体与既有信息不一致。",
        "Error: video 3: no content returned"
      ],
      "audio_insights": [],
      "image_insights": []
    },
    "report": "1. 综合摘要\n3 条视频中有 1-2 条出现明显风险信号，需人工复核。\n\n2. 多模态关键发现\n- 文本: 未提供文本输入\n- 图像: 未提供图像输入\n- 视频: 存在私下转账引导与收款主体不一致线索\n- 音频: 未提供音频输入\n\n3. 风险信号\n- 私下转账引导\n- 收款主体异常切换\n\n4. 风险等级与理由\n- 风险等级: 中\n- 理由: 存在关键风险特征，但仍需补充上下文证据\n\n5. 建议的下一步动作\n- 对风险片段进行人工复核\n- 交叉核验聊天记录、付款凭证与账户信息\n\n6. 诈骗链路还原\n- 短视频导流接触受害者并展示高收益截图\n- 私聊阶段引导受害者绕开平台担保进行转账\n- 通过更换收款主体继续索款，随后拉黑断联"
  }
}
```

### 成功响应（200）- 处理中示例（报告尚未生成）

```json
{
  "task": {
    "task_id": "TASK-1F20AB3C9D8E",
    "user_id": "5",
    "title": "多模态线索核查",
    "status": "processing",
    "summary": "",
    "created_at": "2026-02-21T11:00:00+08:00",
    "updated_at": "2026-02-21T11:00:05+08:00",
    "payload": {
      "text": "请判断是否诈骗",
      "videos": ["<video_base64_1>"],
      "audios": [],
      "images": [],
      "video_insights": [],
      "audio_insights": [],
      "image_insights": []
    },
    "report": ""
  }
}
```

### `video_insights` 真实格式（单条元素）

`video_insights` 数组中的每个元素是字符串，而不是 JSON 对象。单条元素通常是 3 段文本：

```text
【整体视觉感受】
{visual_impression}

【关键信息提取】
{key_content}

【可疑点清单】
1. {point_1}
2. {point_2}
...
```

兼容性说明：

- 不同实现的标题可能略有差异（如 `【整体视觉感受】` 或 `【整体视觉感受（主观特征）】`）。
- 可疑点为空时，第三段可能返回 `- 未发现明显可疑信号` 或 `- 未发现明显视觉异常`。
- 单条分析失败时，对应元素会是 `Error: ...` 文本。

前端建议解析流程：

1. 先判断是否以 `Error:` 开头；若是，按失败文本处理。
2. 按标题 `【...】` 或双换行分段。
3. 第 1 段映射 `visual_impression`，第 2 段映射 `key_content`。
4. 第 3 段按 `^\d+\.` 或 `^-` 提取为 `suspicious_points[]`。

### `report` 报告详细格式（固定模板）

`report` 为纯文本，由 `submit_final_report` 的结构化字段格式化生成，字段来源如下：

```json
{
  "summary": "string",
  "text_finding": "string",
  "image_finding": "string",
  "video_finding": "string",
  "audio_finding": "string",
  "risk_signals": ["string"],
  "risk_level": "低|中|高",
  "risk_reason": "string",
  "next_actions": ["string"],
  "attack_steps": ["string"],
  "scam_keyword_sentences": ["string"]
}
```

说明：

- `attack_steps`、`scam_keyword_sentences` 为可选字段（有内容时传数组；无内容可不传）。
- 若传入这两个字段，必须满足数组约束。

`attack_steps` 约束（严格执行）：

- 若提供，必须是字符串数组（`[]string`）。
- 每个元素仅允许一个步骤，按时间顺序排列。
- 禁止把整条链路写成单个元素（错误示例：`"发布收费通知（班级收书费）→指定线上交费方式→要求每人支付50元"`）。
- 正确示例：`["发布收费通知（班级收书费）", "指定线上交费方式", "要求每人支付50元"]`。

前端渲染兜底（兼容历史/异常数据）：

- 若单个步骤文本内仍出现箭头串联（`->` / `→` / `=>`），前端时间线渲染会按箭头自动拆分为多个步骤节点。
- 若步骤文本为“证据不足，暂无法还原完整诈骗链路”，前端不会渲染时间线节点。

`scam_keyword_sentences` 约束（严格执行）：

- 若提供，必须是字符串数组（`[]string`）。
- 每个元素仅允许一个关键词或关键句。
- 禁止把多个关键词句拼接成单个元素。
- 若返回“未提取到明确诈骗关键词句”，前端不会渲染关键词标签。

字段缺省时 API 返回（重点）：

- 当 `attack_steps` 未提供或为空数组时，`report` 第 6 节固定返回：
  `- 证据不足，暂无法还原完整诈骗链路`
- 当 `scam_keyword_sentences` 未提供或为空数组时，`report` 第 7 节固定返回：
  `- 未提取到明确诈骗关键词句`
- 以上两种兜底文案会出现在 `GET /api/scam/multimodal/tasks/:taskId` 的 `task.report` 文本中（任务完成后）。

渲染后的 `report` 通常为：

```text
1. 综合摘要
{summary}

2. 多模态关键发现
- 文本: {text_finding}
- 图像: {image_finding}
- 视频: {video_finding}
- 音频: {audio_finding}

3. 风险信号
- {risk_signal_1}
- {risk_signal_2}
...

4. 风险等级与理由
- 风险等级: 低 | 中 | 高
- 理由: {risk_reason}

5. 建议的下一步动作
- {next_action_1}
- {next_action_2}
...

6. 诈骗链路还原
- {attack_step_1}
- {attack_step_2}
...

7. 诈骗关键词句
- {scam_keyword_sentence_1}
- {scam_keyword_sentence_2}
...
```

### 字段与返回规则

- `taskId` 统一使用 `TASK-...`。
- `payload` 包含原始输入（`text/videos/audios/images`）与各模态分析结果（`*_insights`）。
- `summary` 为任务摘要字段：
  - 历史归档任务（`completed`/`failed`）返回归档摘要；
  - 进行中任务（`pending`/`processing`）固定返回空字符串 `""`。
- `scam_type` 为可选字段：
  - 历史归档任务（`completed`/`failed`）若已识别诈骗类型则返回；
  - 进行中任务（`pending`/`processing`）通常不返回该字段。
- `*_insights` 始终是字符串数组：
  - 单文件时，长度通常为 `1`。
  - 多文件时，长度通常与输入文件数量一致，按输入顺序对应。
  - 某条失败时，该元素为 `Error: ...`。
  - 未提供某模态时，返回空数组 `[]`。
- `report` 仅在任务完成后返回完整文本；`pending/processing` 可能为空字符串。
- `error`、`history_ref` 属于可选扩展字段，可能返回也可能省略（不同实现略有差异）。

### 常见失败响应

- `400` taskId 为空
- `404` 任务不存在

---

## 11) 聊天对话（需鉴权）

- **Method**: `POST`
- **Path**: `/api/chat`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`
  - `Accept: text/event-stream`

### 请求体

```json
{
  "message": "你好，帮我总结一下我最近的风险情况"
}
```

### 说明

- 纯聊天接口，不调用任何 tool。
- 服务端将每个用户当前会话上下文存入 Redis 缓存，不做持久化落库。
- 响应为 SSE 流式输出：`content` 分片返回，最后返回 `done`。
- Redis 上下文键：`chat:context:<user_id>`。
- 上下文缓存过期时间：5 分钟；每次新请求会刷新 TTL 为 5 分钟。
- 若缓存过期，下一次请求将视为新对话。
- 聊天配置读取：`config/config.json` 的 `chat` 节点（`prompt`、`model`、`api_key`、`base_url`）。
- Redis 配置读取：`config/config.json` 的 `redis` 节点（`addr`、`password`、`db`）。

### 响应事件

- `content`：AI 回复内容片段
  ```json
  {"type": "content", "content": "你好，"}
  ```
- `tool_call`：模型触发工具调用通知
  ```json
  {"type": "tool_call", "tool": "chat_query_user_info", "id": "call_xxx", "arguments": "{}"}
  ```
- `tool_result`：工具执行结果
  ```json
  {"type": "tool_result", "tool": "chat_query_user_info", "id": "call_xxx", "result": {"user": {...}}}
  ```
- `done`：当前轮结束
  ```json
  {"type": "done", "reason": "stop"}
  ```

### SSE 返回格式示例（原始）

```text
event: event
data:{"type":"content","content":"你好，"}

event: event
data:{"type":"content","content":"我可以帮你..."}

event: event
data:{"type":"done","reason":"stop"}
```

> 前端应按 SSE 事件逐条解析并拼接 `content`。注意 `data:` 后可能没有空格，建议使用 `slice(5).trim()` 或类似方式解析。

### 常见失败响应

- `400` 请求参数错误或 message 为空
- `401` 未认证
- `500` 配置加载失败 / Redis 上下文加载失败 / 调用模型失败 / Redis 上下文写入失败

---

## 12) 获取当前对话上下文（需鉴权）

- **Method**: `GET`
- **Path**: `/api/chat/context`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 返回当前用户 Redis 中缓存的会话上下文与剩余有效期。
- `messages` 中会保留完整对话轨迹字段：
  - 普通消息：`role` + `content`
  - 工具调用消息（assistant）：额外包含 `tool_calls`（`id/name/arguments`）
  - 工具结果消息（tool）：额外包含 `tool_call_id`
- 可用于前端判断是否为新对话：
  - `has_context = false` 或 `messages` 为空，表示当前没有上下文（新对话）。
  - `has_context = true` 且 `ttl_seconds > 0`，表示存在正在进行中的会话。

### 成功响应（200）

```json
{
  "user_id": "1",
  "has_context": true,
  "ttl_seconds": 286,
  "messages": [
    {
      "role": "user",
      "content": "帮我看下我的账号风险情况"
    },
    {
      "role": "assistant",
      "content": "",
      "tool_calls": [
        {
          "id": "chatcmpl-tool-9948fb773791ad7c",
          "name": "chat_query_user_info",
          "arguments": "{}"
        }
      ]
    },
    {
      "role": "tool",
      "tool_call_id": "chatcmpl-tool-9948fb773791ad7c",
      "content": "{\"user\":{\"account_status\":\"active\",\"age\":28,\"completed_case_count\":1,\"historical_risk\":\"低\",\"pending_task_count\":0,\"risk_case_count\":{\"中\":0,\"低\":1,\"高\":0},\"user_id\":\"1\",\"user_name\":\"用户1\"}}"
    },
    {
      "role": "assistant",
      "content": "您好，用户1！您的账户状态正常，历史风险等级为低。"
    }
  ]
}
```

### 常见失败响应

- `401` 未认证
- `500` 配置加载失败 / Redis 查询失败

---

## 13) 刷新对话上下文（需鉴权）

- **Method**: `POST`
- **Path**: `/api/chat/refresh`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 立即清除当前用户 Redis 对话上下文，不等待 TTL 过期。
- 刷新后下一次 `POST /api/chat` 将视为新对话。

### 成功响应（200）

```json
{
  "user_id": "1",
  "message": "对话上下文已刷新"
}
```

### 常见失败响应

- `401` 未认证
- `500` 配置加载失败 / Redis 清理失败

---

## 13.1) 实时高风险告警 WebSocket（需鉴权）

- **Method**: `GET`
- **Path**: `/api/alert/ws`
- **协议**: WebSocket

### 触发逻辑（服务端）

- 连接建立后，服务端按配置轮询当前用户 `history_cases`：
  - `config/config.json -> alert_ws.poll_interval_seconds`
- 当发现“`risk_level = 高` 且 `created_at` 在告警窗口内”的记录时，主动推送告警消息：
  - `config/config.json -> alert_ws.recent_window_minutes`
- 同一连接内，同一 `record_id` 只推送一次。
- 连接断开后，后台轮询 goroutine 自动退出。

默认值（配置缺失或非法时自动回退）：

- `poll_interval_seconds = 30`
- `recent_window_minutes = 60`

### 鉴权方式

- 非浏览器客户端：推荐使用 `Authorization: Bearer <JWT_TOKEN>`。
- 浏览器原生 WebSocket：使用 Query 参数 `token` 传 JWT，例如：
  - `ws://localhost:8081/api/alert/ws?token=<JWT_TOKEN>`

### 消息示例（服务端 -> 客户端）

```json
{
  "type": "high_risk_alert",
  "user_id": "1",
  "record_id": "TASK-123456",
  "title": "疑似冒充客服退款",
  "case_summary": "发现转账引导与敏感信息索取",
  "scam_type": "冒充客服类",
  "risk_level": "高",
  "created_at": "2026-03-05T12:01:00Z",
  "sent_at": "2026-03-05T12:01:30Z"
}
```

### 浏览器接入示例

```js
const jwt = localStorage.getItem('token');
const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
const ws = new WebSocket(`${protocol}://${location.host}/api/alert/ws?token=${encodeURIComponent(jwt)}`);

ws.onmessage = (event) => {
  const payload = JSON.parse(event.data);
  if (payload.type === 'high_risk_alert') {
    console.log('收到高风险告警', payload);
  }
};
```

### 常见失败响应

- 握手阶段返回 `401`：Token 缺失、无效或过期。
- 网络断开：客户端需自行重连（建议指数退避）。

---

## 14) 账户升级（需鉴权）

- **Method**: `POST`
- **Path**: `/api/upgrade`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "invite_code": "Secret_Admin_Invite_Code_2026"
}
```

### 说明

- `invite_code`：管理员邀请码，硬编码或通过环境变量 `INVITE_CODE_ADMIN` 配置。
- 成功后用户角色将变更为 `admin`。

### 成功响应（200）

```json
{
  "message": "账户已升级为管理员",
  "user": {
    "id": 1,
    "username": "test_user",
    "email": "test_user@example.com",
    "role": "admin",
    "age": 28
  }
}
```

### 常见失败响应

- `400` 请求参数错误
- `401` 未认证
- `403` 无效的邀请码
- `500` 升级失败

---

## 15) 获取用户列表（仅管理员）

### 15.1 获取所有用户

- **Method**: `GET`
- **Path**: `/api/users`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

**请求示例**：
```http
GET /api/users
```

**成功响应（200）**:

```json
{
  "users": [
    {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "age": 28
    },
    {
      "id": 2,
      "username": "test_user",
      "email": "test@example.com",
      "role": "user",
      "age": 25
    }
  ],
  "count": 2
}
```

### 15.2 搜索特定用户

- **Method**: `GET`
- **Path**: `/api/users`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`
- **Query参数**:
  - `query`: 搜索关键词（模糊匹配用户名或邮箱）

**请求示例**：
```http
GET /api/users?query=admin
```

**成功响应（200）**:

```json
{
  "users": [
    {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "age": 28
    }
  ],
  "count": 1
}
```

### 常见失败响应

- `401` 用户未认证
- `403` 权限不足（非管理员）

---

## 16) 测试页面

- **Method**: `GET`
- **Path**: `/test-login`
- **说明**: 浏览器可视化测试页面，已接入验证码、注册、登录、鉴权查询、删除。

---

## 推荐调用顺序

1. `GET /api/auth/captcha`
2. `POST /api/auth/register`
3. `GET /api/auth/captcha`（登录前建议刷新）
4. `POST /api/auth/login`
5. `GET /api/alert/ws`（WebSocket，建议登录后立即建立）
6. `GET /api/user`
7. `POST /api/scam/multimodal/analyze`
8. `GET /api/scam/multimodal/tasks`
9. `GET /api/scam/multimodal/history`
10. `GET /api/scam/multimodal/history/overview`
11. `GET /api/scam/multimodal/tasks/:taskId`
12. `POST /api/chat`
13. `GET /api/chat/context`
14. `POST /api/chat/refresh`
15. `DELETE /api/user`

---

## cURL 示例

### 获取验证码

```bash
curl -X GET "http://localhost:8081/api/auth/captcha"
```

### 注册

```bash
curl -X POST "http://localhost:8081/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username":"test_user",
    "email":"test_user@example.com",
    "password":"Test@1234",
    "captchaId":"<captchaId>",
    "captchaCode":"<captchaCode>"
  }'
```

### 登录

```bash
curl -X POST "http://localhost:8081/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email":"test_user@example.com",
    "password":"Test@1234",
    "captchaId":"<captchaId>",
    "captchaCode":"<captchaCode>"
  }'
```

### 获取当前用户

```bash
curl -X GET "http://localhost:8081/api/user" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 获取所有用户列表（仅管理员）

```bash
curl -X GET "http://localhost:8081/api/users" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 搜索特定用户（仅管理员）

```bash
curl -X GET "http://localhost:8081/api/users?query=admin" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 删除当前用户

```bash
curl -X DELETE "http://localhost:8081/api/user" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 多模态诈骗智能助手分析

```bash
curl -X POST "http://localhost:8081/api/scam/multimodal/analyze" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "text":"我接到自称客服的电话，说我开通了会员需要转账取消",
    "videos":[],
    "audios":[],
    "images":[]
  }'
```

### 查询当前用户任务状态

```bash
curl -X GET "http://localhost:8081/api/scam/multimodal/tasks" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 查询指定任务详情

```bash
curl -X GET "http://localhost:8081/api/scam/multimodal/tasks/<TASK_ID>" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 查询当前用户历史案件明细

```bash
curl -X GET "http://localhost:8081/api/scam/multimodal/history" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 查询当前用户风险总览

```bash
curl -X GET "http://localhost:8081/api/scam/multimodal/history/overview?interval=day" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 聊天对话（SSE）

```bash
curl -N -X POST "http://localhost:8081/api/chat" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "message": "你好，帮我总结一下我最近的风险情况"
  }'
```

### 刷新对话上下文

```bash
curl -X POST "http://localhost:8081/api/chat/refresh" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 获取当前对话上下文

```bash
curl -X GET "http://localhost:8081/api/chat/context" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

## 17) 上传历史案件并自动向量化入库（仅管理员）

- **Method**: `POST`
- **Path**: `/api/scam/case-library/cases`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "title": "冒充客服退款引导转账",
  "target_group": "老人",
  "risk_level": "高",
  "scam_type": "冒充客服类",
  "case_description": "诈骗方冒充平台客服，以“会员自动续费”名义要求受害者将资金转入所谓安全账户。",
  "typical_scripts": [
    "您不开通取消会每月自动扣费。",
    "为了资金安全，请先把钱转到监管账户。"
  ],
  "keywords": [
    "客服退款",
    "安全账户",
    "自动续费"
  ],
  "violated_law": "涉嫌违反《中华人民共和国刑法》第二百六十六条（诈骗罪）。",
  "suggestion": "立即停止转账，保存聊天和转账凭证，并第一时间报警。"
}
```

### 字段说明

- `title`: 历史案件标题，必填。
- `target_group`: 目标人群，必填，固定枚举值：`老人`、`青年`、`中年`、`未成年`、`学生`、`其他`。
- `risk_level`: 风险等级，必填，固定枚举值：`高`、`中`、`低`。
- `scam_type`: 诈骗类型，必填，固定 15 类：`冒充客服类`、`冒充公检法类`、`刷单返利类`、`虚假投资理财类`、`虚假网络贷款类`、`虚假征信类`、`冒充领导熟人类`、`婚恋交友类`、`博彩赌博类`、`虚假购物服务类`、`机票退改签类`、`兼职招聘类`、`网络游戏交易类`、`直播打赏类`、`其他诈骗类`。
- `case_description`: 案件描述，必填；长度需在 12 到 400 个字符之间，且不能是明显随机字符串（例如长连续字母数字串）。
- `typical_scripts`: 典型话术列表，可选；传入时建议每条为非空字符串。若为空数组则按“未提供”处理。
- `keywords`: 关键词列表，可选；传入时建议为语义关键词。若为空数组则按“未提供”处理。
- `violated_law`: 违反法律说明，可选。空字符串会按“未提供”处理。
- `suggestion`: 处置建议，可选。空字符串会按“未提供”处理。

### 入库与向量化说明

- 仅管理员可调用此接口。
- 服务端在接收请求后，会自动拼接上述字段并调用 embedding 模型生成向量。
- 向量与结构化字段会一起写入独立 SQLite 数据库文件：
  - 默认路径：`DB/historical_case_library.db`
  - 可通过环境变量覆盖：`HISTORICAL_CASE_DB_PATH`
- 该库与现有 `auth_system.db`、多模态任务状态库分离，不共享连接。

### 成功响应（201）

```json
{
  "message": "historical case stored",
  "case": {
    "case_id": "HCASE-5F3C91AA12DE",
    "created_by": "1",
    "title": "冒充客服退款引导转账",
    "target_group": "老人",
    "risk_level": "高",
    "scam_type": "冒充客服类",
    "case_description": "诈骗方冒充平台客服，以“会员自动续费”名义要求受害者将资金转入所谓安全账户。",
    "typical_scripts": [
      "您不开通取消会每月自动扣费。",
      "为了资金安全，请先把钱转到监管账户。"
    ],
    "keywords": [
      "客服退款",
      "安全账户",
      "自动续费"
    ],
    "violated_law": "涉嫌违反《中华人民共和国刑法》第二百六十六条（诈骗罪）。",
    "suggestion": "立即停止转账，保存聊天和转账凭证，并第一时间报警。",
    "embedding_model": "baai/bge-m3",
    "embedding_dimension": 1024,
    "created_at": "2026-03-02T20:40:31+08:00"
  }
}
```

### 常见失败响应

- `400` 必填字段缺失（`title`/`target_group`/`risk_level`/`scam_type`/`case_description`） / 字段格式错误 / `target_group`、`risk_level` 或 `scam_type` 非固定枚举值 / `case_description` 过短、过长（超过 400 字符）或疑似随机字符串。
- `401` 未认证。
- `403` 权限不足（非管理员）。
- `500` embedding 调用失败 / 独立数据库写入失败。

### cURL 示例

```bash
curl -X POST "http://localhost:8081/api/scam/case-library/cases" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "冒充客服退款引导转账",
    "target_group": "老人",
    "risk_level": "高",
    "scam_type": "冒充客服类",
    "case_description": "诈骗方冒充平台客服，以“会员自动续费”名义要求受害者将资金转入所谓安全账户。",
    "typical_scripts": [
      "您不开通取消会每月自动扣费。",
      "为了资金安全，请先把钱转到监管账户。"
    ],
    "keywords": ["客服退款", "安全账户", "自动续费"],
    "violated_law": "涉嫌违反《中华人民共和国刑法》第二百六十六条（诈骗罪）。",
    "suggestion": "立即停止转账，保存聊天和转账凭证，并第一时间报警。"
  }'
```

---

## 18) 历史案件预览列表（仅管理员）

- **Method**: `GET`
- **Path**: `/api/scam/case-library/cases`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 仅管理员可调用此接口。
- 返回所有历史案件的预览信息。
- 按 `created_at desc` 倒序返回。
- 预览仅包含：`title`、`target_group`、`risk_level`、`scam_type`，并附带 `case_id` 方便请求详情。

### 成功响应（200）

```json
{
  "total": 2,
  "cases": [
    {
      "case_id": "HCASE-5F3C91AA12DE",
      "title": "冒充客服退款引导转账",
      "target_group": "老人",
      "risk_level": "高",
      "scam_type": "冒充客服类"
    },
    {
      "case_id": "HCASE-A1B2C3D4E5F6",
      "title": "虚假投资平台拉群荐股",
      "target_group": "青年",
      "risk_level": "中",
      "scam_type": "虚假投资理财类"
    }
  ]
}
```

### 常见失败响应

- `401` 未认证。
- `403` 权限不足（非管理员）。
- `500` 预览查询失败。

### cURL 示例

```bash
curl -X GET "http://localhost:8081/api/scam/case-library/cases" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

## 18.1) 历史案件库统计总览（仅管理员）

- **Method**: `GET`
- **Path**: `/api/scam/case-library/cases/overview`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### Query 参数

- `interval`（可选）：时间聚合粒度，支持 `day`、`week`、`month`，默认 `day`。

### 说明

- 仅管理员可调用此接口。
- 该接口基于 `ListHistoricalCasePreviews` 进行聚合，适用于后台总览看板。
- 返回三类信息：
  - 时间趋势统计（`trend`）：按 `interval` 聚合每个时间桶的案件数量；
  - 按诈骗类型统计（`by_scam_type`）；
  - 按目标人群统计（`by_target_group`）。

### 成功响应（200）

```json
{
  "interval": "week",
  "total": 12,
  "by_scam_type": [
    {"name": "冒充客服类", "count": 5},
    {"name": "刷单返利类", "count": 4},
    {"name": "未知", "count": 3}
  ],
  "by_target_group": [
    {"name": "老人", "count": 7},
    {"name": "青年", "count": 5}
  ],
  "trend": [
    {"time_bucket": "2026-W08", "count": 3},
    {"time_bucket": "2026-W09", "count": 6},
    {"time_bucket": "2026-W10", "count": 3}
  ]
}
```

### 时间桶格式

- `day`：`YYYY-MM-DD`
- `week`：`YYYY-Www`（ISO 周，例如 `2026-W10`）
- `month`：`YYYY-MM`

### 常见失败响应

- `400` `interval` 非法（仅支持 `day/week/month`）
- `401` 未认证
- `403` 权限不足（非管理员）
- `500` 统计查询失败

### cURL 示例

```bash
curl -X GET "http://localhost:8081/api/scam/case-library/cases/overview?interval=week" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

## 18.2) 获取可选诈骗类型列表（仅管理员）

- **Method**: `GET`
- **Path**: `/api/scam/case-library/options/scam-types`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 返回当前后端配置中可用的诈骗类型列表。
- 前端录入历史案件时可直接使用该返回渲染下拉选项，避免硬编码。

### 成功响应（200）

```json
{
  "total": 15,
  "options": [
    "冒充客服类",
    "冒充公检法类",
    "刷单返利类"
  ]
}
```

### 常见失败响应

- `401` 未认证。
- `403` 权限不足（非管理员）。

---

## 18.3) 获取可选目标人群列表（仅管理员）

- **Method**: `GET`
- **Path**: `/api/scam/case-library/options/target-groups`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 返回当前后端配置中可用的目标人群列表。
- 前端录入历史案件时可直接使用该返回渲染下拉选项，避免硬编码。

### 成功响应（200）

```json
{
  "total": 6,
  "options": [
    "老人",
    "青年",
    "中年"
  ]
}
```

### 常见失败响应

- `401` 未认证。
- `403` 权限不足（非管理员）。

---

## 19) 历史案件详情（仅管理员）

- **Method**: `GET`
- **Path**: `/api/scam/case-library/cases/:caseId`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 仅管理员可调用此接口。
- 按 `case_id` 返回单条历史案件完整内容。
- 返回字段包含结构化字段和 `embedding_vector`。

### 成功响应（200）

```json
{
  "case": {
    "case_id": "HCASE-5F3C91AA12DE",
    "created_by": "1",
    "title": "冒充客服退款引导转账",
    "target_group": "老人",
    "risk_level": "高",
    "scam_type": "冒充客服类",
    "case_description": "诈骗方冒充平台客服，以“会员自动续费”名义要求受害者将资金转入所谓安全账户。",
    "typical_scripts": [
      "您不开通取消会每月自动扣费。",
      "为了资金安全，请先把钱转到监管账户。"
    ],
    "keywords": [
      "客服退款",
      "安全账户",
      "自动续费"
    ],
    "violated_law": "涉嫌违反《中华人民共和国刑法》第二百六十六条（诈骗罪）。",
    "suggestion": "立即停止转账，保存聊天和转账凭证，并第一时间报警。",
    "embedding_vector": [0.0123, -0.0456, 0.0034],
    "embedding_model": "baai/bge-m3",
    "embedding_dimension": 1024,
    "created_at": "2026-03-02T20:40:31+08:00",
    "updated_at": "2026-03-02T20:40:31+08:00"
  }
}
```

### 常见失败响应

- `400` `caseId` 为空。
- `401` 未认证。
- `403` 权限不足（非管理员）。
- `404` 指定 `caseId` 不存在。
- `500` 详情查询失败。

### cURL 示例

```bash
curl -X GET "http://localhost:8081/api/scam/case-library/cases/HCASE-5F3C91AA12DE" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

## 20) 删除历史案件（仅管理员）

- **Method**: `DELETE`
- **Path**: `/api/scam/case-library/cases/:caseId`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 仅管理员可调用此接口。
- 按 `case_id` 删除指定历史案件。

### 成功响应（200）

```json
{
  "case_id": "HCASE-5F3C91AA12DE",
  "message": "historical case deleted"
}
```

### 常见失败响应

- `400` `caseId` 为空。
- `401` 未认证。
- `403` 权限不足（非管理员）。
- `404` 指定 `caseId` 不存在。
- `500` 删除失败。

### cURL 示例

```bash
curl -X DELETE "http://localhost:8081/api/scam/case-library/cases/HCASE-5F3C91AA12DE" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```
