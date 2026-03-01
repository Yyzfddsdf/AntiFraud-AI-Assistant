# AntiFraud AI Assistant

一个基于 Go 的反诈智能助手服务，包含两条核心能力：

1. 登录与鉴权系统（验证码、注册、登录、JWT、限流、管理员能力）
2. 多模态反诈分析系统（文本/图像/视频/音频，异步任务执行，历史归档查询）

默认服务端口：`8081`

---

## 1. 技术栈

- Go 1.25
- Web 框架：Gin
- ORM / 数据库：GORM + SQLite
- 鉴权：JWT
- 会话上下文：Redis（聊天上下文）
- 模型调用：OpenAI 兼容接口（已统一到自定义 `llm` 客户端）

---

## 2. 快速启动

### 2.1 安装依赖

```bash
go mod tidy
```

### 2.2 运行服务

```bash
go run .
```

启动后可访问：

- API 基地址：`http://localhost:8081/api`
- 登录测试页：`http://localhost:8081/test-login`

---

## 3. 环境变量

- `PORT`：服务端口，默认 `8081`
- `JWT_SECRET`：JWT 密钥（生产环境必须设置）
- `DB_PATH`：登录数据库路径，默认 `DB/auth_system.db`

---

## 4. 配置文件

主配置文件：`config/config.json`

当前为统一配置结构（按智能体拆分）：

- `agents.main.{model, api_key, base_url, max_tokens, top_p, temperature}`
- `agents.image.{model, api_key, base_url, max_tokens, top_p, temperature}`
- `agents.video.{model, api_key, base_url, max_tokens, top_p, temperature}`
- `agents.audio.{model, api_key, base_url, max_tokens, top_p, temperature}`
- `prompts.{main, image, video, audio}`
- `retry.{max_retries, retry_delay_ms}`

配置加载逻辑见 `config/config.go`，包含：

- 路径兜底解析（相对路径 + 项目根目录）
- 字段标准化（trim）
- 完整性校验（模型参数、提示词、重试参数）

---

## 5. 项目结构（核心模块）

- `main.go`
  - 服务启动入口
  - 路由注册（auth/chat/multimodal）
  - 全局 CORS 与限流中间件

- `login_system/`
  - 认证与用户管理
  - `controllers/`：注册/登录/用户查询/升级等接口
  - `middleware/`：JWT 鉴权、管理员校验、限流
  - `database/`：SQLite 初始化与迁移

- `chat_system/`
  - 聊天能力（工具调用 + Redis 上下文）
  - `httpapi/`：聊天接口与上下文接口
  - `service/`：工具调用闭环 + 流式回复 + 上下文持久化
  - `tool/`：聊天工具（用户信息、历史案件）

- `multi_agent/`
  - 多模态分析主流程
  - `main_agent.go`：主智能体编排与工具循环
  - `image.go / video.go / ali_asr.go`：子智能体分析
  - `queue/`：异步入队与后台处理
  - `state/`：任务状态与历史归档持久化（两张表）
  - `tool/`：主智能体工具定义与处理逻辑
  - `httpapi/`：多模态任务接口

- `llm/`
  - OpenAI 兼容自定义客户端与协议结构

---

## 6. 关键业务流程

### 6.1 多模态任务流程

1. `POST /api/scam/multimodal/analyze` 提交任务
2. 任务写入 `pending_tasks`，后台 goroutine 启动处理
3. 并行执行图像/视频/音频子智能体分析
4. 主智能体聚合结果并走工具调用流程
5. 提交 `submit_final_report` 后，调用 `write_user_history_case`
6. 任务从进行中迁移到 `history_cases`

### 6.2 聊天流程

1. 加载 Redis 会话上下文
2. 必要时触发工具调用（用户画像/历史）
3. 流式返回回答内容（SSE）
4. 将本轮消息写回 Redis，并刷新 TTL

---

## 7. 主要 API

### 7.1 鉴权相关

- `GET /api/auth/captcha`
- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/user`
- `DELETE /api/user`
- `POST /api/upgrade`
- `GET /api/users`（管理员）

### 7.2 对话相关（需鉴权）

- `POST /api/chat`
- `GET /api/chat/context`
- `POST /api/chat/refresh`

### 7.3 多模态相关（需鉴权）

- `PUT /api/scam/multimodal/user/age`
- `POST /api/scam/multimodal/analyze`
- `GET /api/scam/multimodal/tasks`
- `GET /api/scam/multimodal/tasks/:taskId`
- `GET /api/scam/multimodal/history`
- `DELETE /api/scam/multimodal/history/:recordId`

更完整的请求/响应样例见 `API.md`。

---

## 8. 持久化说明

- 登录与用户信息：`DB/auth_system.db`（SQLite）
- 多模态任务状态：
  - `pending_tasks`：进行中任务
  - `history_cases`：历史案件归档
- 聊天上下文：Redis（按用户维度，带 TTL）

---

## 9. 开发建议

- 提交前执行：

```bash
go test ./...
```

- 本地联调建议：
  1. 先走 `test-login` 页面获取 token
  2. 再调用多模态任务接口并轮询任务状态
  3. 最后查看历史归档与聊天工具返回是否一致

---

## 10. 备注

当前仓库处于持续重构阶段（统一配置、统一工具调用、统一客户端协议）。  
如需继续扩展新模型或新子智能体，建议优先沿用现有的：

- `CommonAgent / SubAgentBase` 继承结构
- `config/config.json` 按模型独立配置
- `tool` 注册中心 + handler 映射
