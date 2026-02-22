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

### 成功响应（200）

```json
{
  "task": {
    "task_id": "TASK-7B1A038E7452",
    "user_id": "5",
    "title": "TED学术演讲视频风险核查",
    "status": "completed",
    "created_at": "2026-02-20T23:57:30+08:00",
    "updated_at": "2026-02-20T23:57:30+08:00",
    "payload": {
      "text": "",
      "videos": ["<video_base64>"],
      "audios": [],
      "images": [],
      "video_insights": [
        "【整体视觉感受（主观特征）】\n视频为高质量的学术演讲录屏..."
      ],
      "audio_insights": [],
      "image_insights": []
    },
    "report": "1. 综合摘要\n该视频为Beau Lotto在TED平台进行的学术演讲录屏..."
  }
}
```

### 说明

- `taskId` 统一使用：`TASK-...`
- `payload` 中包含输入的多模态数据及各模态的初步分析洞察（`*_insights`）。

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
