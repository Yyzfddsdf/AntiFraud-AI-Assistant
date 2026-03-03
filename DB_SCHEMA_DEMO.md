# 数据库结构演示（两个数据库）

本文档把当前项目中使用的两个 SQLite 数据库文件、表结构和字段约束写清楚，便于联调和排查。

## 1. 数据库文件总览

| 数据库 | 默认路径 | 环境变量覆盖 | 主要用途 |
|---|---|---|---|
| 业务主库 | `DB/auth_system.db` | `DB_PATH` | 用户认证数据 + 多模态任务状态/历史 |
| 历史案件向量库 | `DB/historical_case_library.db` | `HISTORICAL_CASE_DB_PATH` | 历史案件结构化数据 + embedding 向量 |

说明：

- `auth_system.db` 由 `database/setup.go` 初始化连接。
- `historical_case_library.db` 由 `multi_agent/case_library/store.go` 独立初始化连接。
- 两个库连接分离，不共享 `gorm.DB`。

## 2. 业务主库 `auth_system.db`

### 2.1 `users`（登录用户表）

来源：

- 启动时执行 `DB.AutoMigrate(&models.User{})` 自动建表。

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `id` | `uint` / `integer` | 主键 | 用户主键 |
| `created_at` | `time.Time` / `datetime` | 无 | 创建时间（来自 `gorm.Model`） |
| `updated_at` | `time.Time` / `datetime` | 无 | 更新时间（来自 `gorm.Model`） |
| `deleted_at` | `gorm.DeletedAt` / `datetime` | 索引 | 软删除时间（来自 `gorm.Model`） |
| `username` | `string` / `text` | `unique`, `not null` | 用户名 |
| `email` | `string` / `text` | `unique`, `not null` | 邮箱 |
| `age` | `*int` / `integer` | 可空 | 年龄 |
| `password` | `string` / `text` | `not null` | 密码哈希 |
| `role` | `string` / `text` | 默认值 `'user'` | 角色（`user`/`admin`） |

### 2.2 `pending_tasks`（任务进行中表）

来源：

- 首次触发多模态状态存储时执行 `AutoMigrate(&pendingTaskEntity{}, &historyCaseEntity{})`。

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `task_id` | `string` / `varchar(64)` | 主键 | 任务 ID |
| `user_id` | `string` / `text` | 索引, `not null` | 用户 ID |
| `title` | `string` / `varchar(255)` | `not null` | 任务标题 |
| `status` | `string` / `varchar(32)` | 索引, `not null` | 任务状态：`pending/processing` 等 |
| `payload_text` | `string` / `text` | 无 | 文本输入 |
| `payload_videos` | `string` / `text` | 无 | 视频输入数组（JSON 字符串） |
| `payload_audios` | `string` / `text` | 无 | 音频输入数组（JSON 字符串） |
| `payload_images` | `string` / `text` | 无 | 图片输入数组（JSON 字符串） |
| `payload_video_insights` | `string` / `text` | 无 | 视频洞察数组（JSON 字符串） |
| `payload_audio_insights` | `string` / `text` | 无 | 音频洞察数组（JSON 字符串） |
| `payload_image_insights` | `string` / `text` | 无 | 图片洞察数组（JSON 字符串） |
| `report` | `string` / `text` | 无 | 过程中可能暂存的报告文本 |
| `error` | `string` / `text` | 无 | 错误信息 |
| `history_ref` | `string` / `varchar(64)` | 无 | 归档引用 ID |
| `created_at` | `time.Time` / `datetime` | 索引, `not null` | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 索引, `not null` | 更新时间 |

数据格式说明：

- `payload_videos/payload_audios/payload_images` 是 JSON 数组字符串。
- 数组元素通常是 Base64 Data URL 或可访问媒体 URL（取决于前端提交内容）。

### 2.3 `history_cases`（任务历史归档表）

来源：

- 与 `pending_tasks` 同一次 `AutoMigrate` 创建。
- 完成/失败时从 `pending_tasks` 迁移并落历史。

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `record_id` | `string` / `varchar(64)` | 主键 | 历史记录 ID（通常等于任务 ID） |
| `user_id` | `string` / `text` | 索引, `not null` | 用户 ID |
| `title` | `string` / `varchar(255)` | `not null` | 历史标题 |
| `case_summary` | `string` / `text` | 无 | 案件摘要 |
| `status` | `string` / `varchar(32)` | 索引, `not null` | `completed/failed` |
| `risk_level` | `string` / `varchar(32)` | 索引 | 风险等级（高/中/低） |
| `payload_text` | `string` / `text` | 无 | 原始文本 |
| `payload_videos` | `string` / `text` | 无 | 原始视频数组（JSON 字符串） |
| `payload_audios` | `string` / `text` | 无 | 原始音频数组（JSON 字符串） |
| `payload_images` | `string` / `text` | 无 | 原始图片数组（JSON 字符串） |
| `payload_video_insights` | `string` / `text` | 无 | 视频洞察数组（JSON 字符串） |
| `payload_audio_insights` | `string` / `text` | 无 | 音频洞察数组（JSON 字符串） |
| `payload_image_insights` | `string` / `text` | 无 | 图片洞察数组（JSON 字符串） |
| `report` | `string` / `text` | 无 | 最终报告 |
| `created_at` | `time.Time` / `datetime` | 索引, `not null` | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 索引, `not null` | 更新时间 |

## 3. 历史案件向量库 `historical_case_library.db`

### 3.1 `historical_case_library`（历史案件语义检索表）

来源：

- 服务启动阶段由 `database.InitHistoricalCaseDB()` 主动初始化连接，并执行 `AutoMigrate(&model.HistoricalCaseEntity{})`。

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `id` | `uint` / `integer` | 主键 | 自增主键 |
| `case_id` | `string` / `varchar(32)` | `uniqueIndex`, `not null` | 业务案件 ID（如 `HCASE-XXXX`） |
| `created_by` | `string` / `varchar(64)` | 索引, `not null` | 创建人（用户 ID） |
| `title` | `string` / `text` | `not null` | 案件标题 |
| `target_group` | `string` / `varchar(32)` | 索引, `not null` | 目标人群（固定枚举） |
| `risk_level` | `string` / `varchar(16)` | 索引, `not null`, 默认值 `中` | 风险等级（固定枚举：高/中/低） |
| `scam_type` | `string` / `varchar(64)` | 索引, `not null`, 默认值 `其他诈骗类` | 诈骗类型（固定 15 类） |
| `case_description` | `string` / `text` | `not null` | 案件描述 |
| `typical_scripts` | `string` / `text` | `not null` | 典型话术数组（JSON 字符串） |
| `keywords` | `string` / `text` | `not null` | 关键词数组（JSON 字符串） |
| `violated_law` | `string` / `text` | `not null` | 违反法律说明 |
| `suggestion` | `string` / `text` | `not null` | 处置建议 |
| `embedding_vector` | `string` / `text` | `not null` | 向量数组（JSON 字符串，`[]float64`） |
| `embedding_model` | `string` / `varchar(128)` | `not null` | 向量模型名 |
| `embedding_dimension` | `int` / `integer` | `not null` | 向量维度 |
| `created_at` | `time.Time` / `datetime` | 索引 | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 无 | 更新时间 |

数据格式说明：

- `typical_scripts`、`keywords`、`embedding_vector` 都是 JSON 字符串列。
- `embedding_vector` 示例（截断）：`[0.0123,-0.0456,...]`。

## 4. 快速核对 SQL（演示）

### 4.1 查看 `auth_system.db` 所有表

```bash
sqlite3 DB/auth_system.db ".tables"
```

### 4.2 查看 `auth_system.db` 表结构

```bash
sqlite3 DB/auth_system.db ".schema users"
sqlite3 DB/auth_system.db ".schema pending_tasks"
sqlite3 DB/auth_system.db ".schema history_cases"
```

### 4.3 查看 `historical_case_library.db` 所有表和结构

```bash
sqlite3 DB/historical_case_library.db ".tables"
sqlite3 DB/historical_case_library.db ".schema historical_case_library"
```

### 4.4 逐字段查看（PRAGMA）

```bash
sqlite3 DB/auth_system.db "PRAGMA table_info(users);"
sqlite3 DB/auth_system.db "PRAGMA table_info(pending_tasks);"
sqlite3 DB/auth_system.db "PRAGMA table_info(history_cases);"
sqlite3 DB/historical_case_library.db "PRAGMA table_info(historical_case_library);"
```

---

## 历史案件库字段总览（更新）

表名：`historical_case_library`

1. `id`：主键（自增）
2. `case_id`：案件 ID（唯一）
3. `created_by`：创建人用户 ID
4. `title`：案件标题（必填）
5. `target_group`：目标人群（必填，固定枚举）
6. `risk_level`：风险等级（必填，固定枚举）
7. `scam_type`：诈骗类型（必填，固定 15 类）
8. `case_description`：案件描述（必填，12~400 字符）
9. `typical_scripts`：典型话术列表（可选，JSON 字符串）
10. `keywords`：关键词列表（可选，JSON 字符串）
11. `violated_law`：违反法律（可选）
12. `suggestion`：建议（可选）
13. `embedding_vector`：向量（JSON 字符串）
14. `embedding_model`：向量模型名
15. `embedding_dimension`：向量维度
16. `created_at`：创建时间
17. `updated_at`：更新时间

### `scam_type` 固定 15 类

1. 冒充客服类
2. 冒充公检法类
3. 刷单返利类
4. 虚假投资理财类
5. 虚假网络贷款类
6. 虚假征信类
7. 冒充领导熟人类
8. 婚恋交友类
9. 博彩赌博类
10. 虚假购物服务类
11. 机票退改签类
12. 兼职招聘类
13. 网络游戏交易类
14. 直播打赏类
15. 其他诈骗类

### 校验规则（重点）

- `title`、`target_group`、`risk_level`、`scam_type`、`case_description` 必填。
- `scam_type` 不在上述 15 类内：直接校验失败，不写入数据库。
- 不会自动把非法值改成默认值。
