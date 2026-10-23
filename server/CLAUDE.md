# server

以 WordPress REST API 格式提供只读接口的 Go 服务，直接读取 WordPress 的 MySQL 数据库。只实现前端（`app/data/*.ts`）用到的接口和字段，以及全文 RSS（`/feed/`，description 也输出全文，受密码保护的文章不输出）。

## 约定

- 路由 huma + chi，入口 cobra，配置 viper（环境变量前缀 `I3W_`），日志 logrus（请求相关日志用 `logging.FromContext(ctx)` 带上 request_id）。
- 数据库 sqlx + sqlc，代码中不写 SQL；查询放在 `internal/wpdb/queries`，改完执行 `go generate ./internal/wpdb`。表名必须写成带反引号的 `` `wp_xxx` ``，运行时替换为实际表前缀。
- 不做迁移：表结构归 WordPress 所有，`internal/wpdb/schema.sql` 只供 sqlc 推导类型。
- 错误处理：跨包返回的错误用 `fmt.Errorf("动词短语: %w", err)` 包装，不写 failed to；只在最外层输出。接口错误使用 `api.Error`（WP_Error 格式）。
- lint 使用 maratori 的 golangci-lint 配置（`.golangci.yml`），工具版本见仓库根目录 `mise.toml`。
- 测试断言用 testify 与 go-cmp。

## 与 WordPress 对齐

- `internal/wpfmt` 移植 `the_content`、`the_title`、`comment_text`、feed 相关过滤器链与 `sanitize_title`，正文图片的 srcset/sizes 依赖附件元数据（`_wp_attachment_metadata`）；`internal/api` 对齐各 REST 控制器的参数、排序、权限与错误码。
- 错误信息按 `WPLANG` 选项本地化，译文摘自官方语言包（`internal/api/errors.go`）。
- 金标准由 WordPress Playground 运行真实 WP 生成，命令见各 `testdata/gen.php` 文件头；端到端测试用 go-mysql-server 加载 `testdata/tables.json`。
