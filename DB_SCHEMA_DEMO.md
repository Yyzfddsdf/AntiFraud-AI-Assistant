# 数据库结构演示（两个数据库）

本文档把当前项目中使用的两个 SQLite 数据库文件、表结构和字段约束写清楚，便于联调和排查。

## 1. 数据库文件总览

| 数据库 | 默认路径 | 环境变量覆盖 | 主要用途 |
|---|---|---|---|
| 业务主库 | `DB/auth_system.db` | `DB_PATH` | 用户认证数据 + 多模态任务状态/历史 + 用户历史向量索引 |
| 历史案件向量库 | `DB/historical_case_library.db` | `HISTORICAL_CASE_DB_PATH` | 历史案件结构化数据 + embedding 向量 |

说明：

- `auth_system.db` 与 `historical_case_library.db` 均由启动阶段统一入口 `database.InitPersistence()` 初始化。
- 主业务库 schema 通过注册式初始化器一次性建表，避免分散在功能首次调用时才懒迁移。
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
| `phone` | `*string` / `text` | `unique`, 可空 | 手机号 |
| `age` | `*int` / `integer` | 可空 | 年龄 |
| `occupation` | `string` / `text` | 可空 | 职业，枚举值来自 `config/occupations.json` |
| `recent_tags` | `string` / `text` | 可空 | 近期标签数组（JSON 字符串） |
| `password` | `string` / `text` | `not null` | 密码哈希 |
| `role` | `string` / `text` | 默认值 `'user'` | 角色（`user`/`admin`） |

说明：

- `occupation` 允许为空；若非空，必须命中 `config/occupations.json` 中的枚举值。
- `recent_tags` 由服务端编码为 JSON 字符串数组，例如 `["近期频繁网购","正在找工作"]`。

### 2.1.1 `family_groups`（家庭组表）

来源：

- 服务启动时由 `family_system.EnsureSchema(database.DB)` 自动迁移。

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `id` | `uint` / `integer` | 主键 | 家庭组主键 |
| `created_at` | `time.Time` / `datetime` | 无 | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 无 | 更新时间 |
| `deleted_at` | `gorm.DeletedAt` / `datetime` | 索引 | 软删除时间 |
| `name` | `string` / `varchar(128)` | `not null` | 家庭名称 |
| `owner_user_id` | `uint` / `integer` | 索引, `not null` | 家庭创建者用户 ID |
| `invite_code` | `string` / `varchar(32)` | `uniqueIndex`, `not null` | 家庭邀请码 |
| `status` | `string` / `varchar(32)` | 索引, `not null` | 家庭状态，当前为 `active` |

### 2.1.2 `family_members`（家庭成员表）

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `id` | `uint` / `integer` | 主键 | 成员关系主键 |
| `family_id` | `uint` / `integer` | 索引, `not null` | 所属家庭 |
| `user_id` | `uint` / `integer` | 索引, 唯一, `not null` | 成员用户 ID（MVP 阶段一人仅加入一个家庭） |
| `role` | `string` / `varchar(32)` | 索引, `not null` | 角色：`owner/guardian/member` |
| `relation` | `string` / `varchar(64)` | 可空 | 关系描述，如父亲/女儿 |
| `status` | `string` / `varchar(32)` | 索引, `not null` | 当前实现固定为 `active` |
| `created_by` | `uint` / `integer` | 索引, `not null` | 创建该成员关系的用户 ID |
| `created_at` | `time.Time` / `datetime` | 无 | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 无 | 更新时间 |
| `deleted_at` | `gorm.DeletedAt` / `datetime` | 索引 | 软删除时间 |

### 2.1.3 `family_invitations`（家庭邀请表）

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `id` | `uint` / `integer` | 主键 | 邀请记录主键 |
| `family_id` | `uint` / `integer` | 索引, `not null` | 所属家庭 |
| `inviter_user_id` | `uint` / `integer` | 索引, `not null` | 邀请人 |
| `invitee_email` | `*string` / `varchar(255)` | 索引, 可空 | 受邀邮箱 |
| `invitee_phone` | `*string` / `varchar(32)` | 索引, 可空 | 受邀手机号 |
| `role` | `string` / `varchar(32)` | 索引, `not null` | 邀请加入后的角色 |
| `relation` | `string` / `varchar(64)` | 可空 | 关系描述 |
| `invite_code` | `string` / `varchar(32)` | `uniqueIndex`, `not null` | 邀请码 |
| `status` | `string` / `varchar(32)` | 索引, `not null` | `pending/revoked` |
| `expires_at` | `time.Time` / `datetime` | 索引, `not null` | 过期时间 |
| `created_at` | `time.Time` / `datetime` | 无 | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 无 | 更新时间 |
| `deleted_at` | `gorm.DeletedAt` / `datetime` | 索引 | 软删除时间 |

说明：

- 用户接受邀请并成功加入家庭后，服务端会物理删除所有与该用户邮箱/手机号匹配的邀请记录，不保留已接受记录。
- 邀请过期后，服务端会在查询、创建新邀请、接受邀请等路径中自动清理过期记录。
- `deleted_at` 来自 `gorm.Model`；业务清理路径使用物理删除，不依赖软删除归档已接受或已过期邀请。

### 2.1.4 `family_guardian_links`（守护关系表）

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `id` | `uint` / `integer` | 主键 | 守护关系主键 |
| `family_id` | `uint` / `integer` | 索引, `not null` | 所属家庭 |
| `guardian_user_id` | `uint` / `integer` | 索引, 复合唯一, `not null` | 守护人用户 ID |
| `member_user_id` | `uint` / `integer` | 索引, 复合唯一, `not null` | 被守护成员用户 ID |
| `status` | `string` / `varchar(32)` | 索引, `not null` | 当前实现固定为 `active` |
| `created_at` | `time.Time` / `datetime` | 无 | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 无 | 更新时间 |
| `deleted_at` | `gorm.DeletedAt` / `datetime` | 索引 | 软删除时间 |

### 2.1.5 `family_notifications`（家庭通知表）

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `id` | `uint` / `integer` | 主键 | 家庭通知主键 |
| `family_id` | `uint` / `integer` | 索引, `not null` | 所属家庭 |
| `target_user_id` | `uint` / `integer` | 索引, `not null` | 触发风险事件的成员用户 ID |
| `receiver_user_id` | `uint` / `integer` | 索引, 复合唯一, `not null` | 通知接收人（守护人） |
| `event_type` | `string` / `varchar(64)` | 索引, 复合唯一, `not null` | 事件类型，当前为 `high_risk_case` |
| `record_id` | `string` / `varchar(64)` | 索引, 复合唯一, `not null` | 关联历史案件记录 ID |
| `title` | `string` / `varchar(255)` | `not null` | 案件标题 |
| `case_summary` | `string` / `text` | 可空 | 案件摘要 |
| `scam_type` | `string` / `varchar(64)` | 索引 | 诈骗类型 |
| `risk_level` | `string` / `varchar(32)` | 索引 | 风险等级 |
| `summary` | `string` / `text` | `not null` | 面向守护人的简版通知文案 |
| `event_at` | `time.Time` / `datetime` | 索引, `not null` | 风险事件发生时间 |
| `read_at` | `*time.Time` / `datetime` | 索引, 可空 | 已读时间 |
| `created_at` | `time.Time` / `datetime` | 无 | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 无 | 更新时间 |
| `deleted_at` | `gorm.DeletedAt` / `datetime` | 索引 | 软删除时间 |

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
| `payload_videos` | `string` / `text` | 无 | 视频输入数组（逗号分隔 Base64 字符串） |
| `payload_audios` | `string` / `text` | 无 | 音频输入数组（逗号分隔 Base64 字符串） |
| `payload_images` | `string` / `text` | 无 | 图片输入数组（逗号分隔 Base64 字符串） |
| `payload_video_insights` | `string` / `text` | 无 | 视频洞察数组（逗号分隔 Base64 字符串） |
| `payload_audio_insights` | `string` / `text` | 无 | 音频洞察数组（逗号分隔 Base64 字符串） |
| `payload_image_insights` | `string` / `text` | 无 | 图片洞察数组（逗号分隔 Base64 字符串） |
| `report` | `string` / `text` | 无 | 过程中可能暂存的报告文本 |
| `error` | `string` / `text` | 无 | 错误信息 |
| `history_ref` | `string` / `varchar(64)` | 无 | 归档引用 ID |
| `created_at` | `time.Time` / `datetime` | 索引, `not null` | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 索引, `not null` | 更新时间 |

数据格式说明：

- `payload_*` 列由服务端内部编码为“逗号分隔 Base64 字符串”（不是 JSON 数组字符串）。
- 空数组会存为空字符串 `""`，读取时解码为 `[]`。
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
| `scam_type` | `string` / `varchar(64)` | 索引 | 诈骗类型（可空，来源于归档/工具写入） |
| `status` | `string` / `varchar(32)` | 索引, `not null` | `completed/failed` |
| `risk_level` | `string` / `varchar(32)` | 索引 | 风险等级（高/中/低） |
| `risk_score` | `int` / `integer` | 默认值 `0` | 当前案件风险分（0-100） |
| `risk_summary` | `string` / `text` | 无 | 风险结构化摘要（JSON 字符串） |
| `payload_text` | `string` / `text` | 无 | 原始文本 |
| `payload_videos` | `string` / `text` | 无 | 原始视频数组（逗号分隔 Base64 字符串） |
| `payload_audios` | `string` / `text` | 无 | 原始音频数组（逗号分隔 Base64 字符串） |
| `payload_images` | `string` / `text` | 无 | 原始图片数组（逗号分隔 Base64 字符串） |
| `payload_video_insights` | `string` / `text` | 无 | 视频洞察数组（逗号分隔 Base64 字符串） |
| `payload_audio_insights` | `string` / `text` | 无 | 音频洞察数组（逗号分隔 Base64 字符串） |
| `payload_image_insights` | `string` / `text` | 无 | 图片洞察数组（逗号分隔 Base64 字符串） |
| `report` | `string` / `text` | 无 | 最终报告 |
| `created_at` | `time.Time` / `datetime` | 索引, `not null` | 创建时间 |
| `updated_at` | `time.Time` / `datetime` | 索引, `not null` | 更新时间 |

说明：

- `risk_score` 由主分析阶段调用 `submit_current_risk_assessment` 后由系统规则计算，不允许模型直接编造。
- `risk_summary` 为结构化 JSON 文本，保存各维度得分、命中规则与关键证据摘要，供详情页展示与后续历史分数算法使用。

### 2.4 `user_history_vectors`（用户历史语义索引表）

来源：

- 首次调用 `write_user_history_case` 或 `search_user_history` 时，由 `multi_agent/user_history_index` 执行 `AutoMigrate(&userHistoryVectorEntity{})` 自动创建。
- 该表与 `history_cases` 同属业务主库 `auth_system.db`，但职责独立：
  - `history_cases` 保存业务归档事实；
  - `user_history_vectors` 保存语义检索索引。

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `record_id` | `string` / `varchar(64)` | 复合主键 | 对应 `history_cases.record_id` |
| `user_id` | `string` / `varchar(64)` | 复合主键, 索引, `not null` | 对应 `history_cases.user_id` |
| `embedding_vector` | `string` / `text` | `not null` | embedding 向量（JSON 数组字符串） |
| `embedding_model` | `string` / `varchar(128)` | `not null` | 生成向量所用模型名 |
| `embedding_dimension` | `int` / `integer` | `not null` | 向量维度 |
| `created_at` | `time.Time` / `datetime` | 索引, `not null` | 继承归档记录创建时间 |
| `updated_at` | `time.Time` / `datetime` | 索引, `not null` | 索引最近更新时间 |

说明：

- 该表通过 `record_id + user_id` 与 `history_cases` 关联。
- 该表只承担“语义索引”职责，案件标题、摘要、风险等级等详情统一回 `history_cases` 读取。
- 向量写入失败不会回滚 `history_cases` 归档；工具层会返回 `vector_index_status=failed`，避免连环报错。
- 当前召回范围仅限“当前用户”的索引记录，不会跨用户搜索。
- 当前 embedding 输入仅使用：`title`、`case_summary`、`scam_type`。

## 3. 历史案件向量库 `historical_case_library.db`

### 3.1 `historical_case_library`（历史案件语义检索表）

来源：

- 服务启动阶段由 `database.InitHistoricalCaseDB()` 主动初始化连接，并执行 `AutoMigrate(&model.HistoricalCaseEntity{}, &model.PendingReviewEntity{})`。

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

### 3.2 `pending_review_cases`（待审核案件表）

来源：

- 与 `historical_case_library` 同一次 `AutoMigrate` 创建。
- 智能体分析完成后，通过 `upload_historical_case_to_vector_db` 工具写入此表（不再直接入库知识库）。
- 写入前会先生成 embedding，并与真实案件库做 top1 向量比对；若相似度 `>= 0.9`，则视为重复案件并拒绝写入待审核表。
- 管理员审核通过后，会先调用 `CreateHistoricalCase` 写入 `historical_case_library`，随后从本表物理删除对应待审核记录。
- 管理员待审核列表与详情页都会直接读取本表中的 `violated_law` 字段；若为空，前端按“未提供”处理。

字段：

| 字段名 | 类型（GORM/SQLite） | 约束 | 说明 |
|---|---|---|---|
| `id` | `uint` / `integer` | 主键 | 自增主键 |
| `record_id` | `string` / `varchar(32)` | `uniqueIndex`, `not null` | 业务记录 ID（如 `PREV-XXXX`） |
| `user_id` | `string` / `varchar(64)` | 索引, `not null` | 提交用户 ID |
| `title` | `string` / `text` | `not null` | 案件标题 |
| `target_group` | `string` / `varchar(32)` | 索引, `not null` | 目标人群（固定枚举） |
| `risk_level` | `string` / `varchar(16)` | 索引, `not null`, 默认值 `中` | 风险等级（高/中/低） |
| `scam_type` | `string` / `varchar(64)` | 索引, `not null`, 默认值 `其他诈骗类` | 诈骗类型 |
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

### 3.3 历史案件库枚举与校验补充

说明：

- `historical_case_library` 的完整字段定义以上方 `3.1` 表格为准。
- 本节只补充固定枚举和值校验规则，避免重复列出整张表结构。

#### `scam_type` 固定 15 类

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

#### 校验规则（重点）

- `title`、`target_group`、`risk_level`、`scam_type`、`case_description` 必填。
- `scam_type` 不在上述 15 类内：直接校验失败，不写入数据库。
- 不会自动把非法值改成默认值。

## 4. 快速核对 SQL（演示）

### 4.1 查看 `auth_system.db` 所有表

```bash
sqlite3 DB/auth_system.db ".tables"
```

### 4.2 查看 `auth_system.db` 表结构

```bash
sqlite3 DB/auth_system.db ".schema users"
sqlite3 DB/auth_system.db ".schema family_groups"
sqlite3 DB/auth_system.db ".schema family_members"
sqlite3 DB/auth_system.db ".schema family_invitations"
sqlite3 DB/auth_system.db ".schema family_guardian_links"
sqlite3 DB/auth_system.db ".schema family_notifications"
sqlite3 DB/auth_system.db ".schema pending_tasks"
sqlite3 DB/auth_system.db ".schema history_cases"
sqlite3 DB/auth_system.db ".schema user_history_vectors"
```

### 4.3 查看 `historical_case_library.db` 所有表和结构

```bash
sqlite3 DB/historical_case_library.db ".tables"
sqlite3 DB/historical_case_library.db ".schema historical_case_library"
sqlite3 DB/historical_case_library.db ".schema pending_review_cases"
```

### 4.4 逐字段查看（PRAGMA）

```bash
sqlite3 DB/auth_system.db "PRAGMA table_info(users);"
sqlite3 DB/auth_system.db "PRAGMA table_info(family_groups);"
sqlite3 DB/auth_system.db "PRAGMA table_info(family_members);"
sqlite3 DB/auth_system.db "PRAGMA table_info(family_invitations);"
sqlite3 DB/auth_system.db "PRAGMA table_info(family_guardian_links);"
sqlite3 DB/auth_system.db "PRAGMA table_info(family_notifications);"
sqlite3 DB/auth_system.db "PRAGMA table_info(pending_tasks);"
sqlite3 DB/auth_system.db "PRAGMA table_info(history_cases);"
sqlite3 DB/auth_system.db "PRAGMA table_info(user_history_vectors);"
sqlite3 DB/historical_case_library.db "PRAGMA table_info(historical_case_library);"
sqlite3 DB/historical_case_library.db "PRAGMA table_info(pending_review_cases);"
```
