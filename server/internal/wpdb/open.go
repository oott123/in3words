package wpdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/oott123/in3words/server/internal/config"
)

//go:generate sqlc generate -f ../../sqlc.yaml

// queryTablePrefix 是 queries/*.sql 中书写表名时使用的前缀，表名必须用反引号包裹，
// 这样 prefixDB 才能安全地替换成实际的表前缀。
const queryTablePrefix = "`wp_"

const connMaxLifetime = 5 * time.Minute

// Open 按 wp-config.php 风格的配置连接 MySQL。
// 日期统一按 UTC 解析，post_date 等本地时间字段会原样作为“墙上时间”读出。
func Open(ctx context.Context, cfg config.DB) (*sqlx.DB, error) {
	mc := mysql.NewConfig()
	mc.User = cfg.User
	mc.Passwd = cfg.Password
	mc.DBName = cfg.Name
	mc.Net, mc.Addr = parseHost(cfg.Host)
	mc.ParseTime = true
	mc.Loc = time.UTC
	mc.Params = map[string]string{"charset": cfg.Charset}

	db, err := sqlx.Open("mysql", mc.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql %s: %w", mc.Addr, err)
	}
	db.SetConnMaxLifetime(connMaxLifetime)
	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql %s: %w", mc.Addr, err)
	}
	return db, nil
}

const (
	netTCP      = "tcp"
	netUnix     = "unix"
	defaultPort = ":3306"
)

// parseHost 解析 DB_HOST，返回网络类型与地址。
// 支持 "host"、"host:port"、"[::1]:port" 和 "host:/path/to.sock"。
func parseHost(host string) (string, string) {
	if i := strings.Index(host, ":/"); i >= 0 {
		return netUnix, host[i+1:]
	}
	if strings.HasPrefix(host, "/") {
		return netUnix, host
	}
	if strings.HasPrefix(host, "[") {
		if strings.Contains(host, "]:") {
			return netTCP, host
		}
		return netTCP, host + defaultPort
	}
	if strings.Contains(host, ":") {
		return netTCP, host
	}
	return netTCP, host + defaultPort
}

// NewQueries 返回使用指定表前缀的查询对象。
func NewQueries(db *sqlx.DB, tablePrefix string) *Queries {
	return New(&prefixDB{db: db, prefix: "`" + tablePrefix})
}

// prefixDB 在执行前把查询中的默认表前缀替换为实际前缀，其余行为与底层连接一致。
type prefixDB struct {
	db     *sqlx.DB
	prefix string
}

func (p *prefixDB) rewrite(query string) string {
	if p.prefix == queryTablePrefix {
		return query
	}
	return strings.ReplaceAll(query, queryTablePrefix, p.prefix)
}

func (p *prefixDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	//nolint:wrapcheck // 透明的 DBTX 适配器，由调用方包装错误
	return p.db.ExecContext(ctx, p.rewrite(query), args...)
}

func (p *prefixDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//nolint:wrapcheck // 透明的 DBTX 适配器，由调用方包装错误
	return p.db.PrepareContext(ctx, p.rewrite(query))
}

func (p *prefixDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	//nolint:wrapcheck // 透明的 DBTX 适配器，由调用方包装错误
	return p.db.QueryContext(ctx, p.rewrite(query), args...)
}

func (p *prefixDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return p.db.QueryRowContext(ctx, p.rewrite(query), args...)
}
