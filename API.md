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
  "email": "test_user@example.com"
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
    "email": "test_user@example.com"
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
  "email": "test_user@example.com"
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

## 7) 查询当前用户任务状态（需鉴权）

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
    },
    {
      "task_id": "TASK-123456",
      "user_id": "1",
      "title": "可疑转账引导",
      "status": "completed",
      "created_at": "2026-02-20T10:20:00+08:00",
      "updated_at": "2026-02-20T10:25:00+08:00"
    }
  ]
}
```

> 说明：任务与历史统一使用 `TASK-...` 标识。

---

## 8) 查询当前用户历史案件明细（需鉴权）

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
      "created_at": "2026-02-20T10:35:00+08:00",
      "report": "1. 综合摘要\n...",
      "payload": {
        "text": "...",
        "videos": ["<video_base64_1>"],
        "audios": ["<audio_base64_1>"],
        "images": ["<image_base64_1>"]
      }
    }
  ]
}
```

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

## 11) 测试页面

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
10. `DELETE /api/user`

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
