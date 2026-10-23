package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/oott123/in3words/server/internal/api"
	"github.com/oott123/in3words/server/internal/config"
	"github.com/oott123/in3words/server/internal/wpdb"
	"github.com/oott123/in3words/server/internal/wpfmt"
)

// 测试使用非默认的表前缀，顺带验证表前缀替换。
const testTablePrefix = "i3w_"

// skipTotalCheck 返回是否跳过分页头比对，用于与 WP 有意保持不同的请求。
func skipTotalCheck(path string) bool {
	// WP 先分页再逐条过滤无权阅读的评论，导致总数包含不可见评论；本服务在 SQL 中过滤。
	return path == "/wp/v2/comments"
}

type goldenCase struct {
	Path       string `json:"path"`
	Status     int    `json:"status"`
	Total      any    `json:"total"` // WP 的分页头有时是整数，有时是字符串
	TotalPages any    `json:"total_pages"`
	Body       any    `json:"body"`
}

func TestGolden(t *testing.T) {
	t.Parallel()

	handler, _ := newTestServer(t)

	var cases []goldenCase
	readJSON(t, "testdata/golden.json", &cases)

	for _, c := range cases {
		t.Run(c.Path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/wp-json"+c.Path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, c.Status, rec.Code, "status, body: %s", rec.Body.String())

			if !skipTotalCheck(c.Path) {
				requireHeader(t, c.Total, rec.Header().Get("X-Wp-Total"), "X-WP-Total")
				requireHeader(t, c.TotalPages, rec.Header().Get("X-Wp-Totalpages"), "X-WP-TotalPages")
			}

			var got any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
			want := normalize(c.Body)
			if diff := cmp.Diff(project(want, got), got); diff != "" {
				t.Errorf("body mismatch (-wp +ours):\n%s", diff)
			}
		})
	}
}

// captionIDRe 匹配 wp_unique_prefixed_id() 生成的后缀。该计数器在同一 PHP 进程内递增，
// 金标准是在一个进程里连续请求得到的，后缀没有意义。
var captionIDRe = regexp.MustCompile(`((?:caption-)?attachment[_-]\d+)-\d+"`)

// normalize 去掉金标准中依赖 PHP 进程状态的部分。
func normalize(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, xv := range x {
			out[k] = normalize(xv)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, xv := range x {
			out[i] = normalize(xv)
		}
		return out
	case string:
		return captionIDRe.ReplaceAllString(x, `$1"`)
	}
	return v
}

func requireHeader(t *testing.T, want any, got, name string) {
	t.Helper()
	if want == nil {
		require.Empty(t, got, name)
		return
	}
	require.Equal(t, fmt.Sprint(want), got, name)
}

// project 只保留 WP 响应中本服务也输出了的字段，用于比对子集。
// 本服务输出了但 WP 没有的字段会以 "<missing>" 出现在差异中。
func project(want, got any) any {
	switch g := got.(type) {
	case map[string]any:
		w, ok := want.(map[string]any)
		if !ok {
			return want
		}
		out := make(map[string]any, len(g))
		for k, gv := range g {
			wv, found := w[k]
			if !found {
				out[k] = "<missing>"
				continue
			}
			out[k] = project(wv, gv)
		}
		return out
	case []any:
		w, ok := want.([]any)
		if !ok {
			return want
		}
		out := make([]any, len(w))
		for i, wv := range w {
			if i < len(g) {
				out[i] = project(wv, g[i])
			} else {
				out[i] = wv
			}
		}
		return out
	}
	return want
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, v))
}

func newTestServer(t *testing.T) (http.Handler, *sqlx.DB) {
	t.Helper()

	addr := startMySQL(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	db, err := wpdb.Open(ctx, config.DB{
		Host:    addr,
		User:    "root",
		Name:    "wordpress",
		Charset: "utf8mb4",
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	createTables(ctx, t, db)
	loadTables(ctx, t, db)

	fmtOptions := wpfmt.DefaultOptions()
	fmtOptions.HTML5Caption = true
	handler, err := api.New(wpdb.NewQueries(db, testTablePrefix), api.Options{Format: fmtOptions, CacheSize: 128})
	require.NoError(t, err)
	return handler, db
}

// startMySQL 启动一个内存中的 MySQL 兼容服务（go-mysql-server），返回监听地址。
func startMySQL(t *testing.T) string {
	t.Helper()

	pro := memory.NewDBProvider(memory.NewDatabase("wordpress"))
	engine := sqle.NewDefault(pro)

	lc := net.ListenConfig{}
	l, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())

	srv, err := server.NewServer(
		server.Config{Protocol: "tcp", Address: addr},
		engine,
		sql.NewContext,
		memory.NewSessionBuilder(pro),
		nil,
	)
	require.NoError(t, err)
	go func() { _ = srv.Start() }()
	t.Cleanup(func() { _ = srv.Close() })
	return addr
}

// createTables 按 schema.sql 建表，使用与 WP 默认一致的不区分大小写的排序规则。
func createTables(ctx context.Context, t *testing.T, db *sqlx.DB) {
	t.Helper()

	data, err := os.ReadFile("../wpdb/schema.sql")
	require.NoError(t, err)
	schema := strings.ReplaceAll(string(data), "`wp_", "`"+testTablePrefix)

	for stmt := range strings.SplitSeq(schema, ";") {
		lines := []string{}
		for line := range strings.SplitSeq(stmt, "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "--") {
				lines = append(lines, line)
			}
		}
		stmt = strings.TrimSpace(strings.Join(lines, "\n"))
		if stmt == "" {
			continue
		}
		_, err = db.ExecContext(ctx, stmt+" DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
		require.NoError(t, err, stmt)
	}
}

func loadTables(ctx context.Context, t *testing.T, db *sqlx.DB) {
	t.Helper()

	var tables map[string][]map[string]any
	readJSON(t, "testdata/tables.json", &tables)

	for name, rows := range tables {
		for _, row := range rows {
			cols := make([]string, 0, len(row))
			for col := range row {
				cols = append(cols, col)
			}
			sort.Strings(cols)

			args := make([]any, 0, len(cols))
			quoted := make([]string, 0, len(cols))
			for _, col := range cols {
				quoted = append(quoted, "`"+col+"`")
				args = append(args, row[col])
			}
			query := fmt.Sprintf(
				"INSERT INTO `%s%s` (%s) VALUES (%s)",
				testTablePrefix, name,
				strings.Join(quoted, ", "),
				strings.TrimSuffix(strings.Repeat("?, ", len(cols)), ", "),
			)
			_, err := db.ExecContext(ctx, query, args...)
			require.NoError(t, err, query)
		}
	}
}
