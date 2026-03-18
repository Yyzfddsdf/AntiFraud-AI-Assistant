# API 文档（登录系统）

## 基础信息

- 服务默认地址：`http://localhost:8081`
- 统一前缀：`/api`
- 数据格式：`application/json`
- 鉴权方式：`Authorization: Bearer <token>`
- 活跃会话策略：单用户最多保留 `2` 个最近活跃 token，活跃 TTL 为 `5` 分钟；超出后会按队列语义挤掉最旧 token

## 全局 401 约定（前端必须统一处理）

- **适用范围**：除登录/注册/验证码等免鉴权接口外，所有需要 JWT 的接口都适用本约定。
- **核心含义**：只要受保护接口返回 `401`，就表示“当前登录态已不可继续使用”，前端应视为需要重新登录。
- **常见原因**：
  - 未提供授权 Token
  - Token 无效或已过期
  - 当前登录已在其他设备被挤下线
  - 用户不存在、已删除，或 Token 中的用户信息与数据库不匹配
- **前端统一处理建议**：
  - 清理本地 `token` 与当前用户态缓存
  - 停止继续使用当前登录态重试受保护接口
  - 给出统一提示后跳转登录页
- **例外说明**：`/api/auth/login` 返回 `401` 时，表示“邮箱或密码不正确”，属于登录失败，不应按“登录态失效”处理。
- **当前后端可能返回的典型错误文案**：
  - `未提供授权 Token`
  - `无效或过期的 Token`
  - `当前登录已在其他设备被挤下线，请重新登录`
  - `用户不存在或已被删除`
  - `用户信息不匹配，Token可能已失效`

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

## 1.1) 发送短信验证码

- **Method**: `POST`
- **Path**: `/api/auth/sms-code`
- **Header**:
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "phone": "13800138000"
}
```

### 说明

- 当前短信发送与校验仍为 `TODO` 占位实现
- 当前演示环境下，短信验证码固定为 `000000`
- 该接口同时供“注册”和“短信登录”使用

### 成功响应（200）

```json
{
  "message": "短信验证码已发送，当前演示环境请使用 000000"
}
```

### 常见失败响应

- `400` 请求参数错误 / 手机号格式错误

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
  "phone": "13800138000",
  "password": "Test@1234",
  "captchaId": "b6f2f9d5f0c64a0d8a53f8a1",
  "captchaCode": "AB7K9",
  "smsCode": "000000"
}
```

### 校验规则

- `username` 必填
- `email` 必填，且必须是邮箱格式
- `phone` 必填，且必须是 11 位大陆手机号
- `password` 必填，且需满足：
  - 至少一个大写字母
  - 至少一个小写字母
  - 至少一个符号
- `captchaId` / `captchaCode` 必填且必须匹配
- `smsCode` 必填，当前演示环境固定校验 `000000`
- 用户年龄在注册时默认写入 `28`（无需在请求体传入 `age`）。

### 成功响应（201）

```json
{
  "id": 1,
  "username": "test_user",
  "email": "test_user@example.com",
  "phone": "13800138000",
  "age": 28,
  "role": "user"
}
```

### 常见失败响应

- `400` 请求参数错误 / 手机号格式错误 / 密码不满足复杂度 / 图形验证码错误或过期 / 短信验证码错误
- `409` 邮箱、手机号或用户名已存在

---

## 3) 用户登录

- **Method**: `POST`
- **Path**: `/api/auth/login`
- **Header**:
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

密码登录：

```json
{
  "account": "test_user@example.com",
  "password": "Test@1234",
  "captchaId": "b6f2f9d5f0c64a0d8a53f8a1",
  "captchaCode": "AB7K9"
}
```

或：

```json
{
  "account": "13800138000",
  "password": "Test@1234",
  "captchaId": "b6f2f9d5f0c64a0d8a53f8a1",
  "captchaCode": "AB7K9"
}
```

短信登录：

```json
{
  "phone": "13800138000",
  "smsCode": "000000"
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
    "phone": "13800138000",
    "role": "user",
    "age": 28
  }
}
```

### 说明

- 密码登录支持“邮箱或手机号 + 密码 + 图形验证码”
- 短信登录支持“手机号 + 短信验证码”
- 当前演示环境下，短信验证码固定为 `000000`

### 常见失败响应
- `400` 请求参数不完整 / 手机号格式错误 / 图形验证码错误或过期
- `401` 账号或密码不正确 / 手机号或短信验证码不正确

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
  "phone": "13800138000",
  "role": "user",
  "age": 28,
  "occupation": "企业职员",
  "recent_tags": [
    "近期频繁网购",
    "正在找工作"
  ]
}
```

### 常见失败响应

- `401` 未提供 Token / Token 无效或过期 / 用户不存在或已删除 / 用户信息不匹配

---

## 4.1) 获取职业枚举选项（需鉴权）

- **Method**: `GET`
- **Path**: `/api/user/profile/options/occupations`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 成功响应（200）

```json
{
  "occupations": [
    "学生",
    "企业职员",
    "个体经营",
    "自由职业",
    "教师",
    "医护人员",
    "公务员/事业单位",
    "家庭主妇/主夫",
    "退休人员",
    "待业",
    "其他"
  ],
  "count": 11
}
```

### 说明

- 枚举值来自 `config/occupations.json`。
- 前端修改职业时应从该接口返回值中选择。

---

## 4.2) 更新当前用户画像（需鉴权）

- **Method**: `PUT`
- **Path**: `/api/user/profile`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "age": 28,
  "occupation": "企业职员"
}
```

### 说明

- `age` 取值范围：`1 ~ 150`
- `occupation` 允许为空；若非空，必须命中 `config/occupations.json`
- `recent_tags` 为只读用户画像字段，不允许通过该接口修改

### 成功响应（200）

```json
{
  "message": "用户画像更新成功",
  "user": {
    "id": 1,
    "username": "test_user",
    "email": "test_user@example.com",
    "phone": "13800138000",
    "role": "user",
    "age": 28,
    "occupation": "企业职员",
    "recent_tags": [
      "近期频繁网购",
      "正在找工作"
    ]
  }
}
```

### 常见失败响应

- `400` 请求参数错误 / 年龄越界 / 职业不在枚举内
- `401` 未认证

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

## 6.1) 单图快速风险识别（需鉴权）

- **Method**: `POST`
- **Path**: `/api/scam/image/quick-analyze`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "image": "<image_base64_or_data_url>"
}
```

### 说明

- 该接口面向“单张图片快速判别”场景，直接同步返回结果，不会创建异步任务。
- 使用配置文件中的 `agents.image_quick` 模型与 `prompts.image_quick` 提示词。
- `image` 支持纯 Base64 字符串，也支持 `data:image/...;base64,...` 形式的 Data URL。
- 模型会被约束通过标准化工具输出，只返回两项：
  - `risk_level`：`高 / 中 / 低`
  - `reason`：简洁、客观、可追踪的判断理由
- 该接口不会写入任务队列、历史记录或案件库，适合作为前置快速筛查。

### 成功响应（200）

```json
{
  "risk_level": "高",
  "reason": "图片中出现仿冒客服页面、收款信息和明显转账引导，存在较强诈骗风险。"
}
```

### 常见失败响应

- `400` 请求参数错误 / `image` 为空
- `401` 未认证
- `502` 上游模型调用失败 / 模型未按约定返回标准化结果

### cURL 示例

```bash
curl -X POST "http://localhost:8081/api/scam/image/quick-analyze" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "<image_base64_or_data_url>"
  }'
```

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
  "analysis": {
    "current_bucket": "2026-03-01 ~ 2026-03-07",
    "previous_bucket": "2026-02-22 ~ 2026-02-28",
    "overall_trend": "上升",
    "high_risk_trend": "上升",
    "summary": "基于最近7天与上一窗口的对比，高风险案件上升（1→2），整体风险上升（3→4）。"
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

- 该接口直接基于 `GetCaseHistory` 聚合，返回“风险变化趋势 + 高中低数量统计 + 中文趋势分析”三类总览信息。
- **聚合与填充策略**：
  - 后端仅返回**存在历史案件**的时间桶。若某时间段内无案件，对应的 `time_bucket` 项将不会出现在 `trend` 数组中（即“稀疏数据”）。
  - 后端**不进行自动补零**。若前端图表（如折线图）需要展示连续的时间轴，需由前端根据 `interval` 自行计算完整的时间序列并进行补零填充。
- `analysis` 为轻量趋势判断，当前基于最近两个“活跃窗口”做比较：
  - `day`：最近 `7` 天 vs 上一个 `7` 天
  - `week`：最近 `2` 周 vs 上一个 `2` 周
  - `month`：最近 `1` 个月 vs 上一个 `1` 个月
  - `overall_trend`：比较两个窗口的 `total`
  - `high_risk_trend`：比较两个窗口的 `high`
  - 取值示例：`上升` / `下降` / `平稳` / `暂无足够数据` / `暂无数据`
- 若最近窗口内没有任何案件，则直接返回：`近期无案件`，不再继续做趋势升降判断。

### 近期无案件响应示例

```json
{
  "stats": {
    "high": 3,
    "medium": 5,
    "low": 2,
    "total": 10
  },
  "analysis": {
    "current_bucket": "2026-03-03 ~ 2026-03-09",
    "previous_bucket": "2026-02-24 ~ 2026-03-02",
    "overall_trend": "近期无案件",
    "high_risk_trend": "近期无案件",
    "summary": "最近7天内暂无新增案件，暂不进行风险趋势判断。"
  },
  "trend": [
    {
      "time_bucket": "2026-02-10",
      "high": 1,
      "medium": 0,
      "low": 0,
      "total": 1
    }
  ]
}
```
- `time_bucket` 格式：
  - `day`：`YYYY-MM-DD`
  - `week`：`YYYY-Www`（ISO 周，例如 `2026-W10`）
  - `month`：`YYYY-MM`
- `current_bucket` / `previous_bucket` 格式：
  - 不是单个时间桶，而是“分析窗口标签”
  - 由窗口起止两个桶拼接而成：`<start_bucket> ~ <end_bucket>`
  - 其中每个桶本身仍沿用 `time_bucket` 的格式规则：
    - `day`：如 `2026-03-03 ~ 2026-03-09`
    - `week`：如 `2026-W07 ~ 2026-W10`
    - `month`：如 `2026-01 ~ 2026-03`

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
    "risk_score": 12,
    "risk_summary": "{\"score\":12,\"risk_level\":\"低\",\"dimensions\":{\"social_engineering\":0,\"requested_actions\":0,\"evidence_strength\":12,\"loss_exposure\":0},\"victim_action_stage\":\"未操作\",\"similar_case_strength\":\"无\",\"multimodal_evidence\":\"中\",\"hit_rules\":[\"出现仿冒官方视觉证据\"],\"key_evidence\":[]}",
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
    "risk_score": 58,
    "risk_summary": "{\"score\":58,\"risk_level\":\"中\",\"dimensions\":{\"social_engineering\":8,\"requested_actions\":36,\"evidence_strength\":14,\"loss_exposure\":0},\"victim_action_stage\":\"未操作\",\"similar_case_strength\":\"中\",\"multimodal_evidence\":\"中\",\"hit_rules\":[\"紧迫催促\",\"要求转账/充值\",\"要求点击链接/安装应用\"],\"key_evidence\":[\"立即转账\",\"私下联系\"]}",
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
    "risk_score": 0,
    "risk_summary": "",
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
  "scam_type": "string",
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
- 诈骗类型: {scam_type}
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
  "message": "请帮我看看这些图片是否可疑",
  "images": [
    "data:image/png;base64,iVBORw0KGgoAAA...",
    "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ..."
  ]
}
```

### 说明

- `message` 与 `images` 至少要提供一个。
- `images` 为可选数组，支持多张图片，元素内容为 Base64 Data URL。
- 服务端内部使用 Responses API。
- 系统提示词会映射为 `developer` 角色。
- 用户文本与图片会映射为 `message.content[]` 中的 `input_text` 与 `input_image`。
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
  {"type": "tool_call", "tool": "chat_query_user_info", "id": "call_xxx"}
  ```
- `tool_result`：工具执行结果
  ```json
  {"type": "tool_result", "tool": "chat_query_user_info", "id": "call_xxx"}
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
  - 用户图片消息：额外包含 `image_urls`
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
      "content": "帮我看下这些图片",
      "image_urls": [
        "data:image/png;base64,iVBORw0KGgoAAA...",
        "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ..."
      ]
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
      "content": "{\"user\":{\"user_name\":\"用户1\",\"age\":28,\"occupation\":\"企业职员\",\"recent_tags\":[\"近期频繁网购\"],\"total_case_count\":1,\"historical_score\":18,\"high_risk_case_ratio\":0,\"mid_risk_case_ratio\":0,\"low_risk_case_ratio\":1,\"risk_trend_analysis\":{\"interval\":\"day\",\"current_bucket\":\"2026-03-10~2026-03-16\",\"previous_bucket\":\"2026-03-03~2026-03-09\",\"overall_trend\":\"持平\",\"high_risk_trend\":\"持平\",\"summary\":\"基于最近7天与上一窗口的对比，高风险案件持平（0→0），整体风险持平（1→1）。\"}}}"
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

## 13.1) 实时风险预警 WebSocket（需鉴权，中高风险）

- **Method**: `GET`
- **Path**: `/api/alert/ws`
- **协议**: WebSocket

### 触发逻辑（服务端）

- 连接建立后，服务端按配置轮询当前用户 `history_cases`：
  - `config/config.json -> alert_ws.poll_interval_seconds`
- 当发现“`risk_level = 中/高` 且 `created_at` 在告警窗口内”的记录时，主动推送风险预警消息：
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
  "type": "risk_alert",
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
  if (payload.type === 'risk_alert') {
    console.log('收到风险预警', payload);
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
      "phone": "13800138000",
      "role": "admin",
      "age": 28
    },
    {
      "id": 2,
      "username": "test_user",
      "email": "test@example.com",
      "phone": "13900139000",
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
  - `query`: 搜索关键词（模糊匹配用户名、邮箱或手机号）

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
      "phone": "13800138000",
      "role": "admin",
      "age": 28
    }
  ],
  "count": 1
}
```

### 说明

- `/api/users` 返回的单个用户结构当前包含：`id`、`username`、`email`、`phone`、`role`、`age`
- 若某个用户尚未绑定手机号，`phone` 字段可能为空或省略

### 常见失败响应

- `401` 用户未认证
- `403` 权限不足（非管理员）

---

## 15.3) 家庭系统（需鉴权）

### 15.3.1 创建家庭

- **Method**: `POST`
- **Path**: `/api/families`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`

```json
{
  "name": "张三家庭"
}
```

说明：

- 当前家庭系统 MVP 阶段默认一个用户只加入一个家庭
- 创建成功后，当前用户会自动成为 `owner`

### 15.3.2 获取我的家庭总览

- **Method**: `GET`
- **Path**: `/api/families/me`

成功响应（已加入家庭）：

```json
{
  "family": {
    "id": 1,
    "name": "张三家庭",
    "owner_user_id": 1,
    "owner_name": "zhangsan",
    "owner_email": "zhangsan@example.com",
    "owner_phone": "13800138000",
    "invite_code": "FAM-8A1BC9D0E1F2",
    "status": "active",
    "member_count": 3,
    "guardian_count": 2
  },
  "current_member": {
    "member_id": 1,
    "family_id": 1,
    "user_id": 1,
    "username": "zhangsan",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "role": "owner",
    "relation": "家庭创建者",
    "status": "active",
    "created_at": "2026-03-11T09:30:00+08:00"
  },
  "members": [],
  "invitations": [],
  "guardian_links": [],
  "unread_notification_count": 2
}
```

成功响应（未加入家庭）：

```json
{
  "family": null,
  "current_member": null,
  "members": [],
  "invitations": [],
  "guardian_links": [],
  "unread_notification_count": 0
}
```

### 15.3.3 创建家庭邀请

- **Method**: `POST`
- **Path**: `/api/families/invitations`

```json
{
  "invitee_email": "parent@example.com",
  "invitee_phone": "13900139000",
  "role": "member",
  "relation": "父亲"
}
```

说明：

- 仅家庭 `owner` 可创建邀请
- `invitee_email` 与 `invitee_phone` 至少填写一项
- `role` 当前支持：`guardian`、`member`

### 15.3.4 查询家庭邀请列表

- **Method**: `GET`
- **Path**: `/api/families/invitations`

说明：

- 仅已加入家庭的成员可查看当前家庭的邀请记录

### 15.3.5 查询我收到的家庭邀请

- **Method**: `GET`
- **Path**: `/api/families/invitations/received`

成功响应示例：

```json
{
  "invitations": [
    {
      "id": 12,
      "family_id": 1,
      "family_name": "张三家庭",
      "inviter_user_id": 3,
      "inviter_name": "zhangsan",
      "inviter_email": "zhangsan@example.com",
      "inviter_phone": "13800138000",
      "invitee_email": "parent@example.com",
      "invitee_phone": "13900139000",
      "role": "member",
      "relation": "父亲",
      "invite_code": "INV-2B73D5E1A0C4",
      "status": "pending",
      "expires_at": "2026-03-19T10:00:00+08:00"
    }
  ]
}
```

说明：

- 返回当前登录账号按邮箱/手机号匹配到的邀请
- 可用于被邀请人侧展示“收到的家庭邀请”列表
- `status` 可能为 `pending`、`revoked`
- 过期邀请会在服务端自动清理，默认不会继续出现在列表中

### 15.3.6 接受家庭邀请

- **Method**: `POST`
- **Path**: `/api/families/invitations/accept`

```json
{
  "invite_code": "INV-2B73D5E1A0C4"
}
```

说明：

- 当前登录用户的邮箱/手机号必须与邀请目标匹配
- 邀请接受后会写入 `family_members`
- 邀请接受后会物理删除所有匹配当前用户邮箱/手机号的家庭邀请记录，不保留已接受记录
- 邀请已过期时，服务端会删除该邀请记录并返回“邀请已过期”

### 15.3.6 查询家庭成员

- **Method**: `GET`
- **Path**: `/api/families/members`

### 15.3.7 更新家庭成员角色/关系

- **Method**: `PATCH`
- **Path**: `/api/families/members/:memberId`

```json
{
  "role": "guardian",
  "relation": "女儿"
}
```

说明：

- 仅家庭 `owner` 可更新
- 家庭创建者不可降级

### 15.3.8 移除家庭成员

- **Method**: `DELETE`
- **Path**: `/api/families/members/:memberId`

### 15.3.9 创建守护关系

- **Method**: `POST`
- **Path**: `/api/families/guardian-links`

```json
{
  "guardian_user_id": 2,
  "member_user_id": 3
}
```

说明：

- 仅家庭 `owner` 可配置
- 守护人角色必须为 `owner` 或 `guardian`

### 15.3.10 查询守护关系

- **Method**: `GET`
- **Path**: `/api/families/guardian-links`

### 15.3.11 删除守护关系

- **Method**: `DELETE`
- **Path**: `/api/families/guardian-links/:linkId`

### 15.3.12 家庭通知 WebSocket

- **Method**: `GET`
- **Path**: `/api/families/notifications/ws`

说明：

- 当前家庭通知来源于“历史归档事件回调”
- 仅当被守护成员归档为高风险案件时，系统才会为对应守护人创建通知
- 连接建立后会按最近窗口轮询 `family_notifications`
- 只推送“当前用户可见 + 未读 + 最近窗口内”的家庭通知
- 家庭通知窗口与轮询频率来自 `config/config.json -> family_alert_ws`

推送示例：

```json
{
  "type": "family_high_risk_alert",
  "notification_id": 1,
  "family_id": 1,
  "target_user_id": 3,
  "target_name": "parent_user",
  "event_type": "high_risk_case",
  "record_id": "TASK-7FA12BC09D11",
  "title": "疑似冒充客服退款",
  "scam_type": "冒充客服类",
  "summary": "家庭成员 parent_user 触发高风险案件，请及时核查。",
  "risk_level": "高",
  "event_at": "2026-03-11T10:15:00+08:00"
}
```

### 15.3.13 标记家庭通知已读

- **Method**: `POST`
- **Path**: `/api/families/notifications/:notificationId/read`

### 常见失败响应

- `400` 请求参数错误 / 邀请码无效 / 邀请目标不匹配 / 无效守护关系配置
- `401` 用户未认证
- `403` 无权操作当前家庭
- `404` 当前用户未加入家庭 / 家庭成员不存在 / 守护关系不存在
- `409` 当前用户已加入家庭 / 邀请已处理

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
7. `GET /api/user/profile/options/occupations`
8. `PUT /api/user/profile`
9. `POST /api/scam/multimodal/analyze`
10. `GET /api/scam/multimodal/tasks`
11. `GET /api/scam/multimodal/history`
12. `GET /api/scam/multimodal/history/overview`
13. `GET /api/scam/multimodal/tasks/:taskId`
14. `POST /api/chat`
15. `GET /api/chat/context`
16. `POST /api/chat/refresh`
16. `DELETE /api/user`

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
    "message": "请分析这些图片",
    "images": [
      "data:image/png;base64,iVBORw0KGgoAAA...",
      "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQ..."
    ]
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
- **聚合与填充策略**：
  - 后端仅返回**存在案件记录**的时间桶。若某时间段内无案件，对应的 `time_bucket` 项将不会出现在 `trend` 数组中（即“稀疏数据”）。
  - 后端**不进行自动补零**。若前端图表（如折线图）需要展示连续的时间轴，需由前端根据 `interval` 自行计算完整的时间序列并进行补零填充。
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

## 18.2) 历史案件库图谱分析（仅管理员）

- **Method**: `GET`
- **Path**: `/api/scam/case-library/cases/graph`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### Query 参数

- `focus_type`（可选）：仅分析指定诈骗类型的画像与局部图谱。
- `focus_group`（可选）：仅返回指定目标人群的 `target_group_top_scam_types` 统计项。
- `top_k`（可选）：每个诈骗类型返回多少个高频目标人群 / 高频关键词 / 相似类型，默认 `5`，最大 `10`。

### 说明

- 仅管理员可调用此接口。
- V1 为**只读派生分析**，不改动数据库结构。
- 数据来源完全来自 `historical_case_library` 现有字段：`scam_type`、`target_group`、`risk_level`、`keywords`、`embedding_vector` 等。
- 返回三部分：
  - `profiles`：每个诈骗类型的画像摘要；
  - `graph`：节点与边组成的轻量图谱。
  - `target_group_top_scam_types`：每类目标人群下，按案件数量排序后的诈骗类型 TopK，`score` 为该类型在该人群总案件中的占比。
- `focus_type` 的行为：
  - 不传或传空字符串：返回**全库所有诈骗类型**的画像与图谱；
  - 传入具体诈骗类型：当前 V1 会收缩为**该诈骗类型的局部画像**，`profiles` 只保留该类型，`graph` 也围绕该类型展开。
- `focus_group` 的行为：
  - 不传或传空字符串：返回所有目标人群的 `target_group_top_scam_types`；
  - 传入具体目标人群：仅保留该人群的 `target_group_top_scam_types` 统计项，不影响 `profiles` 与 `graph`。
- `top_k` 的行为：
  - 对每个诈骗类型，分别最多保留 `top_k` 个高频目标人群、`top_k` 个高频关键词、`top_k` 个相似诈骗类型；
  - 不是“案件 × 关键词”逐条连线，而是**先按诈骗类型聚合**，再从聚合结果里取 TopK。
- `graph` 的阅读方式：
  - `profiles` 是**给人直接看的画像摘要**；
  - `graph` 是**给前端做图或后续交互可视化准备的底层结构**，更偏机器可读。
- 当前类型相似度分数由三部分加权得到：
  - 向量中心余弦相似度 `0.6`
  - 关键词集合重合度 `0.25`
  - 目标人群集合重合度 `0.15`

### `graph` 字段详解

#### `nodes`

- `id`：节点唯一标识，格式为 `<node_type>:<label>`，例如 `scam_type:冒充客服类`
- `node_type`：节点类别，当前 V1 支持：
  - `scam_type`
  - `target_group`
  - `keyword`
- `label`：给人看的节点名称
- `weight`：节点权重，含义取决于节点类型：
  - `scam_type`：该诈骗类型在知识库中的案例数
  - `target_group` / `keyword`：该节点在当前图谱结果中的聚合计数

#### `edges`

- `source`：起点节点 ID
- `target`：终点节点 ID
- `relation`：关系类型，当前 V1 支持：
  - `targets`：诈骗类型 → 目标人群
  - `keyword`：诈骗类型 → 高频关键词
  - `similar`：诈骗类型 → 相似诈骗类型
- `score`：关系强度，含义按关系类型区分：
  - `targets`：该目标人群在该诈骗类型中的出现占比
  - `keyword`：该关键词在该诈骗类型中的出现占比
  - `similar`：诈骗类型之间的综合相似度分数

#### 一个最小理解例子

- `{"source": "scam_type:冒充客服类", "target": "target_group:老人", "relation": "targets", "score": 0.6}`
  - 表示：`冒充客服类` 的案例中，约 60% 聚合到 `老人` 这一目标人群标签。
- `{"source": "scam_type:冒充客服类", "target": "keyword:退款", "relation": "keyword", "score": 0.8}`
  - 表示：`退款` 是 `冒充客服类` 的高频关键词，出现占比约 80%。
- `{"source": "scam_type:冒充客服类", "target": "scam_type:虚假征信类", "relation": "similar", "score": 0.8123}`
  - 表示：这两个诈骗类型在知识特征上相似度较高。

### 成功响应（200）

```json
{
  "summary": {
    "focus_type": "",
    "focus_group": "",
    "top_k": 3,
    "total_cases": 12,
    "scam_type_count": 4,
    "target_group_count": 5,
    "keyword_count": 11
  },
  "profiles": [
    {
      "scam_type": "冒充客服类",
      "case_count": 5,
      "risk_distribution": {
        "high": 3,
        "medium": 1,
        "low": 1,
        "total": 5
      },
      "top_target_groups": [
        {"name": "老人", "count": 3},
        {"name": "中青年", "count": 2}
      ],
      "top_keywords": [
        {"name": "退款", "count": 4},
        {"name": "客服", "count": 4},
        {"name": "征信", "count": 2}
      ],
      "similar_types": [
        {"scam_type": "虚假征信类", "score": 0.8123},
        {"scam_type": "冒充公检法类", "score": 0.4211}
      ]
    }
  ],
  "graph": {
    "nodes": [
      {"id": "scam_type:冒充客服类", "node_type": "scam_type", "label": "冒充客服类", "weight": 5},
      {"id": "target_group:老人", "node_type": "target_group", "label": "老人", "weight": 3},
      {"id": "keyword:退款", "node_type": "keyword", "label": "退款", "weight": 4}
    ],
    "edges": [
      {"source": "scam_type:冒充客服类", "target": "target_group:老人", "relation": "targets", "score": 0.6},
      {"source": "scam_type:冒充客服类", "target": "keyword:退款", "relation": "keyword", "score": 0.8},
      {"source": "scam_type:冒充客服类", "target": "scam_type:虚假征信类", "relation": "similar", "score": 0.8123}
    ]
  },
  "target_group_top_scam_types": [
    {
      "target_group": "老人",
      "total_cases": 4,
      "top_scam_types": [
        {"scam_type": "冒充客服类", "score": 0.5},
        {"scam_type": "虚假征信类", "score": 0.5}
      ]
    }
  ]
}
```

### 常见失败响应

- `400` `top_k` 非整数。
- `401` 未认证。
- `403` 权限不足（非管理员）。
- `500` 图谱分析失败。

### cURL 示例

```bash
curl -X GET "http://localhost:8081/api/scam/case-library/cases/graph?focus_type=冒充客服类&focus_group=老人&top_k=3" \
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
  "message": "历史案件删除成功"
}
```

### 常见失败响应

- `400` `caseId` 为空。
- `401` 未认证。
- `403` 权限不足（非管理员）。
- `404` 指定 `caseId` 不存在。
- `500` 删除失败。

---

## 21) 待审核案件列表（仅管理员）

- **Method**: `GET`
- **Path**: `/api/scam/review/cases`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 仅管理员可调用此接口。
- 返回当前待审核案件预览列表。
- 案件来源：用户通过多模态分析后，智能体自动提交的典型案例（不再直接入库，而是先进入待审核队列）。
- 返回结果会携带 `violated_law`，便于管理员在列表页快速查看是否存在明确法律依据。

### 成功响应（200）

```json
{
  "total": 2,
  "cases": [
    {
      "record_id": "PREV-5F3C91AA12DE",
      "title": "冒充客服退款引导转账",
      "target_group": "老人",
      "risk_level": "高",
      "scam_type": "冒充客服类",
      "violated_law": "涉嫌违反《中华人民共和国刑法》第二百六十六条（诈骗罪）。",
      "created_at": "2026-03-14T10:30:00Z"
    }
  ]
}
```

### 常见失败响应

- `401` 未认证。
- `403` 权限不足（非管理员）。
- `500` 查询失败。

---

## 22) 待审核案件详情（仅管理员）

- **Method**: `GET`
- **Path**: `/api/scam/review/cases/:recordId`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 仅管理员可调用此接口。
- 返回指定 `recordId` 的待审核案件完整详情。

### 成功响应（200）

```json
{
  "case": {
    "record_id": "PREV-5F3C91AA12DE",
    "user_id": "42",
    "title": "冒充客服退款引导转账",
    "target_group": "老人",
    "risk_level": "高",
    "scam_type": "冒充客服类",
    "case_description": "诈骗方冒充平台客服，以"会员自动续费"名义要求受害者将资金转入所谓安全账户。",
    "typical_scripts": ["您不开通取消会每月自动扣费。"],
    "keywords": ["客服退款", "安全账户"],
    "violated_law": "涉嫌违反《中华人民共和国刑法》第二百六十六条（诈骗罪）。",
    "suggestion": "立即停止转账，保存聊天和转账凭证，并第一时间报警。",
    "created_at": "2026-03-14T10:30:00Z",
    "updated_at": "2026-03-14T10:30:00Z"
  }
}
```

### 常见失败响应

- `400` `recordId` 为空。
- `401` 未认证。
- `403` 权限不足（非管理员）。
- `404` 指定 `recordId` 不存在。
- `500` 查询失败。

---

## 23) 审核通过待审核案件（仅管理员）

- **Method**: `POST`
- **Path**: `/api/scam/review/cases/:recordId/approve`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Accept: application/json`

### 说明

- 仅管理员可调用此接口。
- 审核通过后，系统会自动调用 `CreateHistoricalCase` 完成 embedding 生成并写入 `historical_case_library` 知识库。
- 对应待审核记录会从 `pending_review_cases` 物理删除，不再保留已通过副本。

### 成功响应（200）

```json
{
  "message": "审核通过，案件已入库知识库",
  "case_id": "HCASE-5F3C91AA12DE"
}
```

### 常见失败响应

- `400` `recordId` 为空。
- `401` 未认证。
- `403` 权限不足（非管理员）。
- `500` 审核入库失败（可能原因：待审核记录不存在或已处理、embedding 生成失败、待审核记录删除失败等）。

### cURL 示例

```bash
curl -X POST "http://localhost:8081/api/scam/review/cases/PREV-5F3C91AA12DE/approve" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

## 24) 后台启动案件采集（仅管理员）

- **Method**: `POST`
- **Path**: `/api/scam/case-collection/search`
- **Header**:
  - `Authorization: Bearer <JWT_TOKEN>`
  - `Content-Type: application/json`
  - `Accept: application/json`

### 请求体

```json
{
  "query": "冒充客服退款诈骗",
  "case_count": 5
}
```

### 说明

- 仅管理员可调用此接口。
- 接口只负责启动后台 goroutine，不会等待案件采集执行完成。
- 后台流程会驱动案件采集智能体持续调用 `search_web` 和 `upload_historical_case_to_vector_db`，逐条把结果写入待审核案件库。
- 该接口不会创建任务记录，也不会把请求写入数据库。
- 当前没有配套的“任务状态查询接口”；启动后可直接到“待审核案件列表”查看新增结果。

### 参数说明

- `query`：采集主题或检索方向，不能为空。
- `case_count`：希望后台尝试生成的待审核案件数量，取值范围 `1-20`。

### 成功响应（202）

```json
{
  "message": "案件采集任务已在后台启动"
}
```

### 常见失败响应

- `400` 请求参数错误，或 `query` 为空，或 `case_count` 超出 `1-20`。
- `401` 未认证。
- `403` 权限不足（非管理员）。
- `500` 后台入队失败。

### cURL 示例

```bash
curl -X POST "http://localhost:8081/api/scam/case-collection/search" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "冒充客服退款诈骗",
    "case_count": 5
  }'
```
