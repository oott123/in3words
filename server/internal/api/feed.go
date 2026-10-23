package api

import (
	"context"
	"crypto/md5" //nolint:gosec // ETag 与 WP 一致使用 md5，不涉及安全
	"encoding/hex"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/oott123/in3words/server/internal/wpdb"
	"github.com/oott123/in3words/server/internal/wpfmt"
)

const (
	rssDateLayout       = "Mon, 02 Jan 2006 15:04:05 +0000"
	httpDateLayout      = "Mon, 02 Jan 2006 15:04:05 GMT"
	defaultPostsPerRSS  = 10
	defaultCommentsDays = 14
	hoursPerDay         = 24
	feedContentType     = "application/rss+xml; charset=UTF-8"
)

// serveFeed 输出与 feed-rss2.php 一致的 RSS2，description 与 content:encoded 都使用全文。
func (s *Server) serveFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st, err := s.loadSite(ctx)
	if err != nil {
		s.writeFeedError(w, r, err)
		return
	}
	canonical := canonicalFeedPath(st.option("permalink_structure", ""))
	if canonical != "" && r.URL.Path != canonical {
		// 对应 redirect_canonical()；使用相对地址，经反向代理访问时不会跳出本服务。
		target := canonical
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		//nolint:gosec // 目标固定以 /feed 开头，只附带原查询串，不会跳到其他主机
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		return
	}
	feed, err := s.buildFeed(ctx, st)
	if err != nil {
		s.writeFeedError(w, r, err)
		return
	}

	if feed.lastBuild != nil {
		// 与 WP::send_headers() 一致的 Last-Modified、ETag 与条件请求处理。
		lastModified := feed.lastBuild.Format(httpDateLayout)
		sum := md5.Sum([]byte(lastModified)) //nolint:gosec // 同上
		etag := `"` + hex.EncodeToString(sum[:]) + `"`
		w.Header().Set("Last-Modified", lastModified)
		w.Header().Set("ETag", etag)
		if notModified(r, *feed.lastBuild, etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.Header().Set("Content-Type", feedContentType)
	_, _ = w.Write([]byte(feed.render(selfLink(r, st.option("home", "")), st)))
}

func (s *Server) writeFeedError(w http.ResponseWriter, r *http.Request, err error) {
	e := internalError(r.Context(), err)
	e.localize(s.locale(r.Context()))
	http.Error(w, e.Message, e.Data.Status)
}

// canonicalFeedPath 返回 feed 的规范路径：尾部斜杠跟随固定链接结构，朴素链接时保持请求路径不变。
func canonicalFeedPath(structure string) string {
	switch {
	case structure == "":
		return ""
	case strings.HasSuffix(structure, "/"):
		return "/feed/"
	}
	return "/feed"
}

// notModified 对应 WP::send_headers() 中的条件请求判断。
func notModified(r *http.Request, lastModified time.Time, etag string) bool {
	clientETag := r.Header.Get("If-None-Match")
	clientModified := strings.TrimSpace(r.Header.Get("If-Modified-Since"))
	var clientTime time.Time
	if clientModified != "" {
		clientTime, _ = http.ParseTime(clientModified)
	}
	newer := !clientTime.Before(lastModified)
	if clientModified != "" && clientETag != "" {
		return newer && clientETag == etag
	}
	return (clientModified != "" && newer) || clientETag == etag
}

// selfLink 对应 get_self_link()：主机取自 home，路径取自当前请求。
func selfLink(r *http.Request, home string) string {
	scheme, host, _ := strings.Cut(home, "://")
	host, _, _ = strings.Cut(host, "/")
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + host + r.URL.RequestURI()
}

type feedItem struct {
	title        string
	link         string
	commentsLink string
	commentsFeed string
	comments     int64
	showComments bool
	creator      string
	pubDate      string
	categories   []string
	guid         string
	content      string
}

type feed struct {
	items     []feedItem
	lastBuild *time.Time
}

func (s *Server) buildFeed(ctx context.Context, st *site) (*feed, error) {
	limit, err := strconv.ParseInt(st.option("posts_per_rss", strconv.Itoa(defaultPostsPerRSS)), 10, 32)
	if err != nil || limit < 1 {
		limit = defaultPostsPerRSS
	}
	rows, err := s.q.ListFeedPosts(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("list feed posts: %w", err)
	}
	termsByPost, err := s.loadPostTerms(ctx, rows)
	if err != nil {
		return nil, err
	}
	authors, err := s.feedAuthors(ctx, rows)
	if err != nil {
		return nil, err
	}
	l, err := s.newLinker(ctx, st)
	if err != nil {
		return nil, err
	}
	att, err := s.loadAttachments(ctx, rows)
	if err != nil {
		return nil, err
	}

	renderer := wpfmt.NewFeedRenderer(st.fmt)
	f := &feed{items: make([]feedItem, 0, len(rows))}
	for _, row := range rows {
		if !mysqlZero(row.PostModifiedGmt) && (f.lastBuild == nil || row.PostModifiedGmt.After(*f.lastBuild)) {
			t := row.PostModifiedGmt
			f.lastBuild = &t
		}
		item := s.feedItem(st, l, row, termsByPost[row.ID], authors[row.PostAuthor])
		item.content = renderer.Content(row.PostContent, row.ID, att)
		f.items = append(f.items, item)
	}
	return f, nil
}

func (s *Server) feedAuthors(ctx context.Context, rows []wpdb.WpPost) (map[uint64]wpdb.ListUsersByIDsRow, error) {
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.PostAuthor)
	}
	result := map[uint64]wpdb.ListUsersByIDsRow{}
	if len(ids) == 0 {
		return result, nil
	}
	users, err := s.q.ListUsersByIDs(ctx, wpdb.ListUsersByIDsParams{PostTypes: publicPostTypes(), Ids: ids})
	if err != nil {
		return nil, fmt.Errorf("list feed authors: %w", err)
	}
	for _, u := range users {
		result[u.ID] = u
	}
	return result, nil
}

func (s *Server) feedItem(
	st *site,
	l *linker,
	row wpdb.WpPost,
	terms []wpdb.ListObjectTermsRow,
	author wpdb.ListUsersByIDsRow,
) feedItem {
	// get_the_category_rss() 先列分类再列标签，名称去重。
	var categoryIDs []uint64
	var categories, tags []string
	for _, t := range terms {
		switch t.Taxonomy {
		case taxonomyCategory:
			categoryIDs = append(categoryIDs, t.TermID)
			categories = append(categories, wpfmt.TermNameRSS(t.Name))
		case taxonomyTag:
			tags = append(tags, wpfmt.TermNameRSS(t.Name))
		}
	}
	names := make([]string, 0, len(categories)+len(tags))
	for _, name := range append(categories, tags...) {
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}

	permalink := l.postLink(row, categoryIDs, author.UserNicename)
	hash := "#respond"
	if row.CommentCount > 0 {
		hash = "#comments"
	}
	return feedItem{
		title:        wpfmt.TitleRSS(row.PostTitle),
		link:         wpfmt.EscURL(permalink),
		commentsLink: wpfmt.EscURL(permalink + hash),
		commentsFeed: wpfmt.EscURL(l.commentsFeedLink(row.ID, permalink)),
		comments:     row.CommentCount,
		showComments: row.CommentCount > 0 || st.commentsOpen(row),
		creator:      wpfmt.Ent2ncr(author.DisplayName),
		pubDate:      row.PostDateGmt.Format(rssDateLayout),
		categories:   names,
		guid:         wpfmt.EscURL(row.Guid),
	}
}

// commentsOpen 对应 comments_open()，包括 _close_comments_for_old_post() 对旧文章自动关闭评论的处理。
func (s *site) commentsOpen(row wpdb.WpPost) bool {
	if row.CommentStatus != "open" {
		return false
	}
	if !truthy(s.option("close_comments_for_old_posts", "")) || row.PostType != postTypePost {
		return true
	}
	days, err := strconv.Atoi(s.option("close_comments_days_old", strconv.Itoa(defaultCommentsDays)))
	if err != nil || days == 0 || mysqlZero(row.PostDateGmt) {
		return true
	}
	return time.Since(row.PostDateGmt) <= time.Duration(days)*hoursPerDay*time.Hour
}

// htmlLang 对应 get_bloginfo( 'language' )。
func (s *site) htmlLang() string {
	locale := s.option("WPLANG", "")
	if locale == "" {
		locale = "en_US"
	}
	// 语言包翻译了 html_lang_attribute 的语言。
	switch locale {
	case "zh_CN":
		return "zh-Hans"
	case "zh_TW":
		return "zh-Hant"
	}
	return strings.ReplaceAll(locale, "_", "-")
}

func (f *feed) render(self string, st *site) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"
	xmlns:content="http://purl.org/rss/1.0/modules/content/"
	xmlns:wfw="http://wellformedweb.org/CommentAPI/"
	xmlns:dc="http://purl.org/dc/elements/1.1/"
	xmlns:atom="http://www.w3.org/2005/Atom"
	xmlns:sy="http://purl.org/rss/1.0/modules/syndication/"
	xmlns:slash="http://purl.org/rss/1.0/modules/slash/"
	>

<channel>
`)
	lastBuild := ""
	if f.lastBuild != nil {
		lastBuild = f.lastBuild.Format(rssDateLayout)
	}
	fmt.Fprintf(&b, "\t<title>%s</title>\n", wpfmt.DocumentTitle(st.option("blogname", "")))
	fmt.Fprintf(&b, "\t<atom:link href=\"%s\" rel=\"self\" type=\"application/rss+xml\" />\n", wpfmt.EscURL(self))
	fmt.Fprintf(&b, "\t<link>%s</link>\n", wpfmt.BloginfoRSS(st.option("home", "")))
	fmt.Fprintf(&b, "\t<description>%s</description>\n", wpfmt.BloginfoRSS(st.option("blogdescription", "")))
	fmt.Fprintf(&b, "\t<lastBuildDate>%s</lastBuildDate>\n", lastBuild)
	fmt.Fprintf(&b, "\t<language>%s</language>\n", wpfmt.BloginfoRSS(st.htmlLang()))
	b.WriteString("\t<sy:updatePeriod>hourly</sy:updatePeriod>\n\t<sy:updateFrequency>1</sy:updateFrequency>\n")
	for _, item := range f.items {
		item.write(&b)
	}
	b.WriteString("</channel>\n</rss>\n")
	return b.String()
}

func (item *feedItem) write(b *strings.Builder) {
	b.WriteString("\t<item>\n")
	fmt.Fprintf(b, "\t\t<title>%s</title>\n", item.title)
	fmt.Fprintf(b, "\t\t<link>%s</link>\n", item.link)
	if item.showComments {
		fmt.Fprintf(b, "\t\t<comments>%s</comments>\n", item.commentsLink)
	}
	fmt.Fprintf(b, "\t\t<dc:creator><![CDATA[%s]]></dc:creator>\n", item.creator)
	fmt.Fprintf(b, "\t\t<pubDate>%s</pubDate>\n", item.pubDate)
	for _, name := range item.categories {
		fmt.Fprintf(b, "\t\t<category><![CDATA[%s]]></category>\n", name)
	}
	fmt.Fprintf(b, "\t\t<guid isPermaLink=\"false\">%s</guid>\n", item.guid)
	fmt.Fprintf(b, "\t\t<description><![CDATA[%s]]></description>\n", item.content)
	fmt.Fprintf(b, "\t\t<content:encoded><![CDATA[%s]]></content:encoded>\n", item.content)
	if item.showComments {
		fmt.Fprintf(b, "\t\t<wfw:commentRss>%s</wfw:commentRss>\n", item.commentsFeed)
		fmt.Fprintf(b, "\t\t<slash:comments>%d</slash:comments>\n", item.comments)
	}
	b.WriteString("\t</item>\n")
}
