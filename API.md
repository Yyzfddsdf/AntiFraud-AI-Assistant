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

### 成功响应（201）

```json
{
  "id": 1,
  "username": "test_user",
  "email": "test_user@example.com",
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

## 删除当前用户历史案件（需鉴权）

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
    "report": "1. 综合摘要\n该内容以公开演讲为主，未发现直接诈骗指令。\n\n2. 多模态关键发现\n- 文本: 未提供文本输入\n- 图像: 未提供图像输入\n- 视频: 画面与语义一致，未见明显诈骗套路\n- 音频: 未提供音频输入\n\n3. 风险信号\n- 未发现明确风险信号\n\n4. 风险等级与理由\n- 风险等级: 低\n- 理由: 未出现诱导转账、索要敏感信息等关键风险特征\n\n5. 建议的下一步动作\n- 保留原始素材与上下文供后续复核"
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
    "report": "1. 综合摘要\n3 条视频中有 1-2 条出现明显风险信号，需人工复核。\n\n2. 多模态关键发现\n- 文本: 未提供文本输入\n- 图像: 未提供图像输入\n- 视频: 存在私下转账引导与收款主体不一致线索\n- 音频: 未提供音频输入\n\n3. 风险信号\n- 私下转账引导\n- 收款主体异常切换\n\n4. 风险等级与理由\n- 风险等级: 中\n- 理由: 存在关键风险特征，但仍需补充上下文证据\n\n5. 建议的下一步动作\n- 对风险片段进行人工复核\n- 交叉核验聊天记录、付款凭证与账户信息"
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
  "next_actions": ["string"]
}
```

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
```

### 字段与返回规则

- `taskId` 统一使用 `TASK-...`。
- `payload` 包含原始输入（`text/videos/audios/images`）与各模态分析结果（`*_insights`）。
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
- 聊天模型配置读取：`chat_system/config/config.json`（需正确配置 `api_key`、`base_url`、`chat_model`）。
- Redis 配置同样读取 `chat_system/config/config.json`（`redis_addr`、`redis_password`、`redis_db`）。

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
5. `GET /api/user`
6. `POST /api/scam/multimodal/analyze`
7. `GET /api/scam/multimodal/tasks`
8. `GET /api/scam/multimodal/history`
9. `GET /api/scam/multimodal/tasks/:taskId`
10. `POST /api/chat`
11. `GET /api/chat/context`
12. `POST /api/chat/refresh`
13. `DELETE /api/user`

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
