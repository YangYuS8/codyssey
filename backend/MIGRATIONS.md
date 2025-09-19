# 数据库迁移（Go 后端 / goose）

后端使用 [goose](https://github.com/pressly/goose) 管理数据库结构演进。迁移文件位于 `backend/migrations/` 目录，命名格式：

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
goose -dir ./migrations postgres "$DB_DSN" up
```

`$DB_DSN` 形如：
```
postgres://user:pass@localhost:5432/codyssey?sslmode=disable
```

## 4. 回滚迁移

```bash
# 回滚一个版本
goose -dir ./migrations postgres "$DB_DSN" down
# 回滚到指定版本
goose -dir ./migrations postgres "$DB_DSN" up 0002
```

## 5. 常见问题

| 问题 | 可能原因 | 解决 |
|------|----------|------|
| 启动时报 `relation "problems" does not exist` | 迁移未执行或失败 | 查看日志 `migrations applied`，检查数据库连接串 |
| goose CLI 找不到 | 未执行 `go install` | 再次运行安装命令，并确认 `$GOBIN` 在 `PATH` 中 |
| 序号跳跃或重复 | 手动命名冲突 | 保持递增；不要修改已发布版本号 |

## 6. 下一步扩展建议

- 将 `runMigrations` 改为使用 pgx stdlib：`import _ "github.com/jackc/pgx/v5/stdlib"`，然后 `sql.Open("pgx", dsn)`。
- 增加 `users`, `submissions`, `contests`, `judge_runs` 等表的分阶段迁移。
- 为关键表添加索引与约束（唯一、外键）。

## 7. 约定

- 迁移一旦合并主分支，不要“修改旧迁移”，只能追加新文件。
- Down 语句保持对称，但允许在复杂场景下标注 `-- irreversible` 注释说明不可逆。

---
如有需要可在 `docs/development.md` 中加入指向本文件的链接。
