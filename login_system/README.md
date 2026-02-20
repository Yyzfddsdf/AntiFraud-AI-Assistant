# Login System

一个独立的用户登录系统，复用 MyGo 的核心安全逻辑，但不包含任何 AI 聊天和 token 计数逻辑。

## 功能

- 用户注册：`POST /api/auth/register`
- 用户登录：`POST /api/auth/login`
- 获取当前用户：`GET /api/user`（需鉴权）
- 删除当前用户：`DELETE /api/user`（需鉴权）

## 安全逻辑

- 密码哈希（bcrypt）
- JWT 鉴权（24 小时过期）
- 频率限制（每个 IP 每秒最多 5 次）

## 运行

1. 在项目根目录运行：

```bash
go mod tidy
```

2. 启动服务：

```bash
go run .
```

默认端口 `8081`，可通过 `PORT` 环境变量覆盖。

## 环境变量

- `JWT_SECRET`：JWT 密钥（生产环境务必设置）
- `DB_PATH`：SQLite 数据库文件路径（默认固定为项目根目录下 `DB/auth_system.db`，不受启动目录影响）
- `PORT`：服务端口（默认 `8081`）

## 常量集中管理

- 安全与时效相关常量统一在 `login_system/settings/security.go` 管理。
- 包括：JWT 过期时长、验证码长度/有效期/清理周期、限流窗口与阈值。

## 受保护接口请求头示例

- `Authorization: Bearer <token>`
