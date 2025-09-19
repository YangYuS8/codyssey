# 数据库迁移（Go 后端 / goose）

后端使用 [goose](https://github.com/pressly/goose) 管理数据库结构演进（运行时通过 `pgx` stdlib 驱动执行）。迁移文件位于 `backend/migrations/` 目录，命名格式：

```
<version>_<description>.sql
# 例如
0001_create_problems.sql
0002_create_users.sql
```

## 1. 文件结构示例

```sql
-- +goose Up
CREATE TABLE example (...);
-- +goose Down
DROP TABLE example;
```

支持多条语句；`-- +goose Up` 与 `-- +goose Down` 是分隔块。

## 2. 新建迁移

使用 goose CLI（建议全局安装一次）：

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
cd backend
# 创建 SQL 迁移（自动生成当前时间戳版本号，可改为序号方式）
goose create add_something sql
```

由于项目采用递增数字版本（0001, 0002...），可以手动复制已有文件并修改序号与内容，确保不冲突。

## 3. 应用迁移（开发环境）

后端启动时会自动执行：
```
goose Up backend/migrations
```
代码中在 `server.Start()` 调用了 `goose.Up`。如果你想单独运行：

```bash
cd backend
export $(grep -v '^#' ../.env | xargs)  # 或自行设置数据库变量
# 直接使用连接串
goose -dir ./migrations pgx "$DB_DSN" up
```

`$DB_DSN` 形如：
```
postgres://user:pass@localhost:5432/codyssey?sslmode=disable
```

## 4. 回滚迁移

```bash
# 回滚一个版本
goose -dir ./migrations pgx "$DB_DSN" down
# 回滚到指定版本
goose -dir ./migrations pgx "$DB_DSN" up 0002
```

## 5. 常见问题

| 问题 | 可能原因 | 解决 |
|------|----------|------|
| 启动时报 `relation "problems" does not exist` | 迁移未执行或失败 | 查看日志 `migrations applied`，检查数据库连接串 |
| goose CLI 找不到 | 未执行 `go install` | 再次运行安装命令，并确认 `$GOBIN` 在 `PATH` 中 |
| 序号跳跃或重复 | 手动命名冲突 | 保持递增；不要修改已发布版本号 |

## 6. 下一步扩展建议

- （已采用）使用 pgx stdlib：`import _ "github.com/jackc/pgx/v5/stdlib"` + `sql.Open("pgx", dsn)`。
- 增加 `users`, `submissions`, `contests`, `judge_runs` 等表的分阶段迁移。
- 为关键表添加索引与约束（唯一、外键）。

### 已实现里程碑补充

- 0001_create_problems.sql: 初始 `problems` 表。
- 0002_create_users.sql: 初始 `users` 表，包含 `username`、`created_at`。
- 0003_add_roles_to_users.sql: 引入 `roles TEXT[]` 列（使用 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` 简化实现，去除 DO 块避免某些环境 dollar-quote 解析问题）。
	- 设计原因：RBAC 权限系统需要在用户实体上直接存储多角色集合，便于后续鉴权一次加载。
	- 未额外创建用户名索引：`username UNIQUE` 已隐式生成唯一索引。
	- 如未来要抽离角色为独立实体，可新增 `roles` / `user_roles` 关系并迁移数据。
 - 0004_add_password_hash_to_users.sql: 添加 `password_hash` 列（可为 NULL 兼容旧数据；新注册必须写入）。
	 - 密码使用 bcrypt(DefaultCost) 存储；若后续需要更强策略（argon2id / pepper），新增迁移与逻辑即可。
	 - Refresh/Access Token 均使用对称 HS256 JWT，后续可引入旋转/黑名单（存储 refresh jti）。
 - 0005_create_submissions.sql: 初始 `submissions` 表，包含基础字段（status 默认 pending）。
 - 0006_add_submission_metrics_and_logs.sql: 为 `submissions` 增加执行指标列：`runtime_ms INT DEFAULT 0`、`memory_kb INT DEFAULT 0`、`error_message TEXT NULL`、`version INT DEFAULT 1`，并新增 `submission_status_logs` 表记录状态流转（from_status -> to_status, created_at）。
 - 0007_create_judge_runs.sql: 新增 `judge_runs` 表，用于记录判题执行实例（多次重判/不同判题版本）。
	 - 关键字段：
		 - `submission_id`: 关联提交；允许一个 submission 拥有多个 run（重判或升级判题程序）
		 - `status`: 生命周期 `queued` -> `running` -> (`succeeded`|`failed`|`canceled`)
		 - 指标：`runtime_ms`, `memory_kb`, `exit_code`, `error_message`
		 - 时间戳：`started_at`, `finished_at`（仅在 running / 终态更新）
		 - `judge_version`: 记录判题沙箱/镜像或配置版本，支持回溯复现
	 - 索引：`submission_id`（列表/聚合查询），`status`（调度器过滤 queued），后续可考虑 `(submission_id, created_at DESC)` 复合索引用于最近 run 查询。
	 - 设计原因：与 `submissions` 拆分：submission 反映用户一次提交的“意图与最终结果”，judge_runs 细化底层判题过程（可多次尝试 / 重跑），为后续调度、追踪、性能统计、回放提供结构化数据。

## 7. 约定

- 迁移一旦合并主分支，不要“修改旧迁移”，只能追加新文件。
- Down 语句保持对称，但允许在复杂场景下标注 `-- irreversible` 注释说明不可逆。

---
如有需要可在 `docs/development.md` 中加入指向本文件的链接。
