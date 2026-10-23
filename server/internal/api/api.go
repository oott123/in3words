// Package api 以 WordPress REST API 的格式提供只读接口，数据直接读取 WordPress 的 MySQL 数据库。
// 只实现了 in3words 前端用到的接口与字段。
package api

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/oott123/in3words/server/internal/logging"
	"github.com/oott123/in3words/server/internal/wpdb"
	"github.com/oott123/in3words/server/internal/wpfmt"
)

// Options 是创建 Server 所需的配置。
type Options struct {
	// Format 中与数据库无关的部分（主题特性、表情图片地址等）；
	// HomeURL、UseSmilies、HTTPSMigrationRequired 会在每次请求时用数据库中的选项覆盖。
	Format wpfmt.Options
	// CacheSize 是渲染结果缓存的条目数，0 表示不缓存。
	CacheSize int
}

// Server 实现各个接口。
type Server struct {
	q          *wpdb.Queries
	fmtOptions wpfmt.Options
	cache      *lru.Cache[renderKey, string]
	lang       localeCache
}

// New 创建 HTTP handler。
func New(q *wpdb.Queries, opts Options) (http.Handler, error) {
	s := &Server{q: q, fmtOptions: opts.Format}
	if opts.CacheSize > 0 {
		cache, err := lru.New[renderKey, string](opts.CacheSize)
		if err != nil {
			return nil, fmt.Errorf("create render cache: %w", err)
		}
		s.cache = cache
	}

	hookHumaErrors()

	router := chi.NewMux()
	router.Use(logging.Middleware)
	router.Get("/feed", s.serveFeed)
	router.Get("/feed/", s.serveFeed)
	router.NotFound(s.writeNoRoute)
	router.MethodNotAllowed(s.writeNoRoute)

	cfg := huma.DefaultConfig("in3words WordPress API", "1.0.0")
	// 不在响应里注入 $schema 字段和 Link 头，保持与 WP 的响应一致。
	cfg.CreateHooks = nil
	cfg.DocsPath = ""
	cfg.Transformers = append(cfg.Transformers, s.localizeError, filterFields)
	api := humachi.New(router, cfg)

	s.register(api)
	return router, nil
}

func (s *Server) register(api huma.API) {
	huma.Get(api, "/wp-json", s.getIndex, hidden)
	huma.Get(api, "/wp-json/", s.getIndex, func(o *huma.Operation) {
		o.OperationID = "get-index"
		o.Summary = "站点信息"
	})
	huma.Get(api, "/wp-json/wp/v2/posts", s.listPosts, func(o *huma.Operation) {
		o.Summary = "文章列表"
	})
	huma.Get(api, "/wp-json/wp/v2/posts/{id}", s.getPost, func(o *huma.Operation) {
		o.Summary = "单篇文章"
	})
	huma.Get(api, "/wp-json/wp/v2/pages", s.listPages, func(o *huma.Operation) {
		o.Summary = "页面列表"
	})
	huma.Get(api, "/wp-json/wp/v2/categories", s.listCategories, func(o *huma.Operation) {
		o.Summary = "分类列表"
	})
	huma.Get(api, "/wp-json/wp/v2/tags", s.listTags, func(o *huma.Operation) {
		o.Summary = "标签列表"
	})
	huma.Get(api, "/wp-json/wp/v2/users", s.listUsers, func(o *huma.Operation) {
		o.Summary = "用户列表"
	})
	huma.Get(api, "/wp-json/wp/v2/comments", s.listComments, func(o *huma.Operation) {
		o.Summary = "评论列表"
	})
}

func hidden(o *huma.Operation) {
	o.Hidden = true
	o.OperationID += "-no-slash"
}

func (s *Server) writeNoRoute(w http.ResponseWriter, r *http.Request) {
	e := errNoRoute()
	e.localize(s.locale(r.Context()))
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(e)
}

func (s *Server) localizeError(ctx huma.Context, _ string, v any) (any, error) {
	if e, ok := v.(*Error); ok {
		e.localize(s.locale(ctx.Context()))
	}
	return v, nil
}

// renderKind 区分不同的渲染管线，避免缓存键冲突。
type renderKind uint8

const (
	renderContent renderKind = iota + 1
	renderTitle
	renderComment
)

// renderKey 以原文的哈希而不是修改时间作为键，这样直接改库也不会读到旧结果。
type renderKey struct {
	kind renderKind
	id   uint64
	hash uint64
	opts wpfmt.Options
}

// render 渲染并缓存结果。att 只在正文渲染时使用，缓存键包含正文引用到的附件元数据。
func (s *Server) render(kind renderKind, id uint64, raw string, opts wpfmt.Options, att wpfmt.Attachments) string {
	if raw == "" {
		return render(kind, raw, opts, att)
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(raw))
	if kind == renderContent {
		for _, attID := range wpfmt.AttachmentIDs(raw) {
			_, _ = fmt.Fprintf(h, "\x00%d\x00%s", attID, att[attID])
		}
	}
	key := renderKey{kind: kind, id: id, hash: h.Sum64(), opts: opts}

	if s.cache != nil {
		if v, ok := s.cache.Get(key); ok {
			return v
		}
	}
	v := render(kind, raw, opts, att)
	if s.cache != nil {
		s.cache.Add(key, v)
	}
	return v
}

func render(kind renderKind, raw string, opts wpfmt.Options, att wpfmt.Attachments) string {
	switch kind {
	case renderContent:
		return wpfmt.TheContent(raw, opts, att)
	case renderTitle:
		return wpfmt.TheTitle(raw)
	case renderComment:
		return wpfmt.CommentText(raw, opts)
	}
	return raw
}
