package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/oott123/in3words/server/internal/wpdb"
	"github.com/oott123/in3words/server/internal/wpfmt"
)

const (
	postTypePost = "post"
	postTypePage = "page"
)

// Rendered 对应 WP 中 {"rendered": "..."} 形式的字段。
type Rendered struct {
	Rendered string `json:"rendered"`
}

// PostContent 是文章正文。
type PostContent struct {
	Rendered  string `json:"rendered"`
	Protected bool   `json:"protected"`
}

// Post 是 /wp/v2/posts 与 /wp/v2/pages 的响应条目。
type Post struct {
	ID            uint64        `json:"id"`
	Date          *string       `json:"date"`
	DateGMT       *string       `json:"date_gmt"`
	Modified      *string       `json:"modified"`
	ModifiedGMT   *string       `json:"modified_gmt"`
	Slug          string        `json:"slug"`
	Status        string        `json:"status"`
	Type          string        `json:"type"`
	Title         Rendered      `json:"title"`
	Content       PostContent   `json:"content"`
	Author        uint64        `json:"author"`
	Parent        *uint64       `json:"parent,omitempty"`
	MenuOrder     *int32        `json:"menu_order,omitempty"`
	CommentStatus string        `json:"comment_status"`
	PingStatus    string        `json:"ping_status"`
	Sticky        *bool         `json:"sticky,omitempty"`
	Categories    *[]uint64     `json:"categories,omitempty"`
	Tags          *[]uint64     `json:"tags,omitempty"`
	Embedded      *PostEmbedded `json:"_embedded,omitempty"`
}

// PostEmbedded 是 _embed 时附带的关联资源。
type PostEmbedded struct {
	// Author 的元素是 EmbedUser 或 *Error。
	Author []any         `json:"author"`
	Terms  [][]EmbedTerm `json:"wp:term,omitempty"`
}

// EmbedParam 识别 _embed 参数是否出现（WP 只看是否存在，值可以为空）。
// 与 Pagination 同时嵌入时 Resolve 方法有歧义、不会被提升，必须是导出类型 huma 才会解析。
type EmbedParam struct {
	Embed bool
}

func (e *EmbedParam) Resolve(ctx huma.Context) []error {
	u := ctx.URL()
	e.Embed = u.Query().Has("_embed")
	return nil
}

// Pagination 是列表接口共用的分页参数，必须以导出类型嵌入，huma 才会解析其中的字段。
// 范围校验放在 Resolve 中，以便输出与 WP 相同的错误信息。
type Pagination struct {
	Page    int `query:"page"     default:"1"  doc:"页码，从 1 开始"`
	PerPage int `query:"per_page" default:"10" doc:"每页条数，1 到 100"`
}

const maxPerPage = 100

func (p *Pagination) Resolve(_ huma.Context) []error {
	var errs []error
	if p.Page < 1 {
		errs = append(errs, newParamError("page", "%s must be greater than or equal to %d", "page", 1))
	}
	if p.PerPage < 1 || p.PerPage > maxPerPage {
		errs = append(errs, newParamError(
			"per_page", "%s must be between %d (inclusive) and %d (inclusive)", "per_page", 1, maxPerPage,
		))
	}
	return errs
}

func (p *Pagination) offset() int32 {
	return int32((p.Page - 1) * p.PerPage) //nolint:gosec // page 与 per_page 已由参数校验限定范围
}

func (p *Pagination) totalPages(total int64) int64 {
	return (total + int64(p.PerPage) - 1) / int64(p.PerPage)
}

type listPostsInput struct {
	EmbedParam
	Pagination

	Search     string   `query:"search"     doc:"搜索关键词"`
	After      string   `query:"after"      doc:"只返回此时间之后发布的文章（ISO 8601）"`
	Before     string   `query:"before"     doc:"只返回此时间之前发布的文章（ISO 8601）"`
	Author     []uint64 `query:"author"     doc:"作者 ID"`
	Categories []uint64 `query:"categories" doc:"分类 ID"`
	Tags       []uint64 `query:"tags"       doc:"标签 ID"`
	Slug       []string `query:"slug"       doc:"别名"`
}

type listPagesInput struct {
	EmbedParam
	Pagination

	Search string   `query:"search" doc:"搜索关键词"`
	Author []uint64 `query:"author" doc:"作者 ID"`
	Slug   []string `query:"slug"   doc:"别名"`
}

type listPostsOutput struct {
	Total      int64 `header:"X-WP-Total"`
	TotalPages int64 `header:"X-WP-TotalPages"`
	Body       []Post
}

type getPostInput struct {
	EmbedParam

	ID string `path:"id"`
}

type getPostOutput struct {
	Body Post
}

func (s *Server) listPosts(ctx context.Context, in *listPostsInput) (*listPostsOutput, error) {
	st, err := s.loadSite(ctx)
	if err != nil {
		return nil, internalError(ctx, err)
	}
	f := postFilter{
		postType:   postTypePost,
		search:     in.Search,
		authors:    in.Author,
		categories: in.Categories,
		tags:       in.Tags,
		slugs:      in.Slug,
	}
	if in.After != "" {
		t, ok := st.parseRESTDate(in.After)
		if !ok {
			return nil, errInvalidParam("after", "Invalid date.")
		}
		f.after = &t
	}
	if in.Before != "" {
		t, ok := st.parseRESTDate(in.Before)
		if !ok {
			return nil, errInvalidParam("before", "Invalid date.")
		}
		f.before = &t
	}
	return s.queryPosts(ctx, st, f, &in.Pagination, in.Embed)
}

func (s *Server) listPages(ctx context.Context, in *listPagesInput) (*listPostsOutput, error) {
	st, err := s.loadSite(ctx)
	if err != nil {
		return nil, internalError(ctx, err)
	}
	f := postFilter{
		postType: postTypePage,
		search:   in.Search,
		authors:  in.Author,
		slugs:    in.Slug,
	}
	return s.queryPosts(ctx, st, f, &in.Pagination, in.Embed)
}

type postFilter struct {
	postType   string
	search     string
	authors    []uint64
	categories []uint64
	tags       []uint64
	slugs      []string
	after      *time.Time
	before     *time.Time
}

func (s *Server) queryPosts(
	ctx context.Context,
	st *site,
	f postFilter,
	page *Pagination,
	embed bool,
) (*listPostsOutput, error) {
	slugs := sanitizeSlugs(f.slugs)
	terms := parseSearch(f.search)

	args := wpdb.ListPostsParams{
		PostType:         f.postType,
		FilterAuthors:    flag(len(f.authors) > 0),
		Authors:          f.authors,
		FilterSlugs:      flag(len(f.slugs) > 0),
		Slugs:            slugs,
		FilterCategories: flag(len(f.categories) > 0),
		Categories:       f.categories,
		FilterTags:       flag(len(f.tags) > 0),
		Tags:             f.tags,
		Searching:        flag(len(terms) > 0),
		Limit:            int32(page.PerPage), //nolint:gosec // per_page 已由参数校验限定范围
		Offset:           page.offset(),
	}
	if f.after != nil {
		args.After = sql.NullTime{Time: st.localWallClock(*f.after), Valid: true}
	}
	if f.before != nil {
		args.Before = sql.NullTime{Time: st.localWallClock(*f.before), Valid: true}
	}
	applySearch(&args, terms)

	total, err := s.q.CountPosts(ctx, countPostsParams(args))
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("count posts: %w", err))
	}
	totalPages := page.totalPages(total)
	if int64(page.Page) > totalPages && total > 0 {
		return nil, newError(
			http.StatusBadRequest,
			"rest_post_invalid_page_number",
			"The page number requested is larger than the number of pages available.",
		)
	}

	out := &listPostsOutput{Total: total, TotalPages: totalPages, Body: []Post{}}
	if total == 0 {
		return out, nil
	}

	rows, err := s.q.ListPosts(ctx, args)
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("list posts: %w", err))
	}
	posts, err := s.preparePosts(ctx, st, rows, embed)
	if err != nil {
		return nil, internalError(ctx, err)
	}
	out.Body = posts
	return out, nil
}

// countPostsParams 从列表参数中取出筛选条件；两条查询的 WHERE 一致，所以字段一一对应。
func countPostsParams(p wpdb.ListPostsParams) wpdb.CountPostsParams {
	return wpdb.CountPostsParams{
		PostType:         p.PostType,
		FilterAuthors:    p.FilterAuthors,
		Authors:          p.Authors,
		FilterSlugs:      p.FilterSlugs,
		Slugs:            p.Slugs,
		FilterCategories: p.FilterCategories,
		Categories:       p.Categories,
		FilterTags:       p.FilterTags,
		Tags:             p.Tags,
		After:            p.After,
		Before:           p.Before,
		Searching:        p.Searching,
		S1:               p.S1,
		S2:               p.S2,
		S3:               p.S3,
		S4:               p.S4,
		S5:               p.S5,
		S6:               p.S6,
		S7:               p.S7,
		S8:               p.S8,
		S9:               p.S9,
		X1:               p.X1,
		X2:               p.X2,
		X3:               p.X3,
		X4:               p.X4,
		X5:               p.X5,
		X6:               p.X6,
		X7:               p.X7,
		X8:               p.X8,
		X9:               p.X9,
	}
}

func (s *Server) getPost(ctx context.Context, in *getPostInput) (*getPostOutput, error) {
	invalid := newError(http.StatusNotFound, "rest_post_invalid_id", "Invalid post ID.")
	id, err := strconv.ParseUint(in.ID, 10, 64)
	if err != nil {
		return nil, errNoRoute()
	}
	if id == 0 {
		return nil, invalid
	}

	row, err := s.q.GetPost(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, invalid
	}
	if err != nil {
		return nil, internalError(ctx, fmt.Errorf("get post %d: %w", id, err))
	}
	if row.PostType != postTypePost {
		return nil, invalid
	}
	if row.PostStatus != "publish" {
		return nil, newError(http.StatusUnauthorized, "rest_forbidden", "Sorry, you are not allowed to do that.")
	}

	st, err := s.loadSite(ctx)
	if err != nil {
		return nil, internalError(ctx, err)
	}
	posts, err := s.preparePosts(ctx, st, []wpdb.WpPost{row}, in.Embed)
	if err != nil {
		return nil, internalError(ctx, err)
	}
	return &getPostOutput{Body: posts[0]}, nil
}

// preparePosts 对应 WP_REST_Posts_Controller::prepare_item_for_response()，批量查询关联的分类与作者。
func (s *Server) preparePosts(ctx context.Context, st *site, rows []wpdb.WpPost, embed bool) ([]Post, error) {
	authorIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		authorIDs = append(authorIDs, row.PostAuthor)
	}
	termsByPost, err := s.loadPostTerms(ctx, rows)
	if err != nil {
		return nil, err
	}
	att, err := s.loadAttachments(ctx, rows)
	if err != nil {
		return nil, err
	}

	var authors map[uint64]any
	if embed {
		authors, err = s.embedUsers(ctx, st, authorIDs)
		if err != nil {
			return nil, err
		}
	}

	sticky := st.stickyPosts()
	posts := make([]Post, 0, len(rows))
	fmtOpts := st.fmt
	for _, row := range rows {
		p := s.preparePost(st, row, fmtOpts, att)
		// WP 在同一请求中渲染过含区块的正文后，wpautop 会排到 shortcode_unautop 之后。
		if row.PostPassword == "" && wpfmt.HasBlocks(row.PostContent) {
			fmtOpts.AutopAfterUnautop = true
		}
		if row.PostType == postTypePost {
			isSticky := sticky[row.ID]
			p.Sticky = &isSticky
			attachTerms(&p, termsByPost[row.ID], embed)
		}
		// WP 只在 post_author 非空时添加 author 链接，没有作者也没有分类时不输出 _embedded。
		if embed && row.PostAuthor != 0 {
			if p.Embedded == nil {
				p.Embedded = &PostEmbedded{}
			}
			p.Embedded.Author = []any{authors[row.PostAuthor]}
		}
		posts = append(posts, p)
	}
	return posts, nil
}

// loadPostTerms 批量查询文章（不含页面）的分类与标签，按文章 ID 分组。
func (s *Server) loadPostTerms(ctx context.Context, rows []wpdb.WpPost) (map[uint64][]wpdb.ListObjectTermsRow, error) {
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		if row.PostType == postTypePost {
			ids = append(ids, row.ID)
		}
	}
	result := map[uint64][]wpdb.ListObjectTermsRow{}
	if len(ids) == 0 {
		return result, nil
	}
	termRows, err := s.q.ListObjectTerms(ctx, wpdb.ListObjectTermsParams{
		ObjectIds:  ids,
		Taxonomies: []string{taxonomyCategory, taxonomyTag},
	})
	if err != nil {
		return nil, fmt.Errorf("list post terms: %w", err)
	}
	for _, t := range termRows {
		result[t.ObjectID] = append(result[t.ObjectID], t)
	}
	return result, nil
}

// loadAttachments 批量读取正文中图片引用的附件元数据，受密码保护的文章不渲染正文，因此跳过。
func (s *Server) loadAttachments(ctx context.Context, rows []wpdb.WpPost) (wpfmt.Attachments, error) {
	var ids []uint64
	for _, row := range rows {
		if row.PostPassword == "" {
			ids = append(ids, wpfmt.AttachmentIDs(row.PostContent)...)
		}
	}
	att := wpfmt.Attachments{}
	if len(ids) == 0 {
		return att, nil
	}
	metas, err := s.q.ListAttachmentMetadata(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list attachment metadata: %w", err)
	}
	for _, m := range metas {
		if _, exists := att[m.PostID]; !exists {
			att[m.PostID] = m.MetaValue
		}
	}
	return att, nil
}

// attachTerms 填充文章的 categories、tags 字段，embed 时同时填充 _embedded["wp:term"]。
// terms 已按名称排序，与 get_the_terms() 的顺序一致。
func attachTerms(p *Post, terms []wpdb.ListObjectTermsRow, embed bool) {
	categories, tags := []uint64{}, []uint64{}
	embedCategories, embedTags := []EmbedTerm{}, []EmbedTerm{}
	for _, t := range terms {
		et := embedTerm(t.TermID, t.Name, t.Slug, t.Taxonomy)
		switch t.Taxonomy {
		case taxonomyCategory:
			categories = append(categories, t.TermID)
			embedCategories = append(embedCategories, et)
		case taxonomyTag:
			tags = append(tags, t.TermID)
			embedTags = append(embedTags, et)
		}
	}
	p.Categories = &categories
	p.Tags = &tags
	if embed {
		p.Embedded = &PostEmbedded{Terms: [][]EmbedTerm{embedCategories, embedTags}}
	}
}

func (s *Server) preparePost(st *site, row wpdb.WpPost, fmtOpts wpfmt.Options, att wpfmt.Attachments) Post {
	dateGMT := row.PostDateGmt
	if mysqlZero(dateGMT) {
		dateGMT = st.gmtFromLocal(row.PostDate)
	}
	modifiedGMT := row.PostModifiedGmt
	if mysqlZero(modifiedGMT) {
		offset := time.Duration(st.gmtOffset() * float64(time.Hour))
		modifiedGMT = row.PostModified.Add(-offset)
	}

	content := ""
	// 对应 post_password_required()：访客无法提供密码，受保护文章的正文一律为空。
	if row.PostPassword == "" {
		content = s.render(renderContent, row.ID, row.PostContent, fmtOpts, att)
	}

	p := Post{
		ID:            row.ID,
		Date:          prepareDate(row.PostDateGmt, row.PostDate),
		DateGMT:       prepareDate(dateGMT, dateGMT),
		Modified:      prepareDate(row.PostModifiedGmt, row.PostModified),
		ModifiedGMT:   prepareDate(modifiedGMT, modifiedGMT),
		Slug:          row.PostName,
		Status:        row.PostStatus,
		Type:          row.PostType,
		Title:         Rendered{Rendered: s.render(renderTitle, row.ID, row.PostTitle, wpfmt.Options{}, nil)},
		Content:       PostContent{Rendered: content, Protected: row.PostPassword != ""},
		Author:        row.PostAuthor,
		CommentStatus: row.CommentStatus,
		PingStatus:    row.PingStatus,
	}
	if row.PostType == postTypePage {
		parent, menuOrder := row.PostParent, row.MenuOrder
		p.Parent = &parent
		p.MenuOrder = &menuOrder
	}
	return p
}

func flag(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func sanitizeSlugs(slugs []string) []string {
	result := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		result = append(result, wpfmt.SanitizeTitle(slug))
	}
	return result
}
