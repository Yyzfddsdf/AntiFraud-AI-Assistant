# Agent 多模态风控服务

本项目是一个基于 Go 的多模态风控分析服务，包含两部分能力：

- 登录与鉴权系统（验证码、注册、登录、JWT、限流）
- 多模态诈骗分析任务系统（文本/视频/音频/图像，异步队列执行，任务与历史可查询）

服务默认端口：`8081`

---

## 1. 技术栈

- Go 1.25
- Web 框架：Gin
- ORM/数据库：GORM + SQLite
- 鉴权：JWT
- 多模态与主分析：`go-openai` 兼容接口调用

依赖见 `go.mod`。

---

## 2. 快速启动

### 2.1 安装依赖

```bash
go mod tidy
```

### 2.2 启动服务

```bash
go run .
```

服务启动后：

- API 基址：`http://localhost:8081/api`
- 测试页面：`http://localhost:8081/test-login`

---

## 3. 环境变量

- `PORT`：服务端口，默认 `8081`
- `JWT_SECRET`：JWT 密钥（生产必须设置）
- `DB_PATH`：登录库路径，默认 `DB/auth_system.db`

---

## 4. 配置文件

主分析配置默认从 `config/config.json` 读取，字段包括：

- `api_key`
- `base_url`
- `image_model`
- `audio_model`
- `main_model`

路径解析支持相对路径兜底（见 `config/config.go`）。

---

## 5. 目录结构（核心）

- `main.go`：服务入口、路由注册
- `login_system/`：验证码/注册/登录/JWT/中间件
- `multi_agent/`：多模态分析、主智能体、工具调用、状态存储
  - `httpapi/`：多模态任务相关 HTTP 接口
  - `queue/`：异步任务队列 worker
  - `state/`：任务与历史状态持久化（`DB/multi_agent_state.json`）
  - `tool/`：主智能体工具定义与实现
- `config/`：模型配置
- `DB/`：本地 SQLite 与状态文件
- `API.md`：接口文档（详细）

---

## 6. 主要 API

鉴权相关：

- `GET /api/auth/captcha`
- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/user`
- `DELETE /api/user`

多模态相关（需鉴权）：

- `PUT /api/scam/multimodal/user/age`
- `POST /api/scam/multimodal/analyze`
- `GET /api/scam/multimodal/tasks`
- `GET /api/scam/multimodal/tasks/:taskId`
- `GET /api/scam/multimodal/history`

完整请求/响应示例见 `API.md`。

---

## 7. 数据落盘说明

- 登录用户信息：SQLite（默认 `DB/auth_system.db`）
- 多模态任务状态：JSON（`DB/multi_agent_state.json`）
  - `pending`：排队/处理中任务
  - `history`：历史案件

任务 payload 除原始模态数据外，还会保存函数级解读内容：

- `video_insights`
- `audio_insights`
- `image_insights`

这些解读来自子函数分析结果，不依赖主模型生成。

---

## 8. 主智能体工具机制

当前采用“最小必要参数 + 服务端上下文补全”策略：

- 模型只传业务必要参数（例如标题、摘要、风险等级）
- `user_id`、`task_id`、原始多模态输入、函数解读等由后端上下文绑定

这样可以减少模型侧参数复杂度，避免越权/伪造内部字段。

---

## 9. 开发建议

- 提交前执行：

```bash
go test ./...
```

- 若需调试接口，优先使用 `/test-login` 页面走完整流程。

---

## 10. 备注

本仓库当前同时包含登录系统与多模态分析能力，后续如需扩展（例如向量化检索、外部向量库）建议在 `multi_agent/` 下按模块增量演进。