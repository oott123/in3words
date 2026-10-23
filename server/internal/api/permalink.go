package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/oott123/in3words/server/internal/wpdb"
)

// linker 生成与 get_permalink() 一致的文章链接。
type linker struct {
	home      string
	structure string
	// categories 仅在固定链接包含 %category% 时加载，用于拼接父分类路径。
	categories      map[uint64]wpdb.ListTaxonomyTermsRow
	defaultCategory uint64
}

func (s *Server) newLinker(ctx context.Context, st *site) (*linker, error) {
	l := &linker{home: st.option("home", ""), structure: st.option("permalink_structure", "")}
	if !strings.Contains(l.structure, "%category%") {
		return l, nil
	}
	rows, err := s.q.ListTaxonomyTerms(ctx, taxonomyCategory)
	if err != nil {
		return nil, fmt.Errorf("list categories for permalink: %w", err)
	}
	l.categories = make(map[uint64]wpdb.ListTaxonomyTermsRow, len(rows))
	for _, row := range rows {
		l.categories[row.TermID] = row
	}
	l.defaultCategory, _ = strconv.ParseUint(st.option("default_category", "1"), 10, 64)
	return l, nil
}

// homeURL 对应 home_url( $path )。
func (l *linker) homeURL(path string) string {
	if path == "" {
		return l.home
	}
	return l.home + "/" + strings.TrimLeft(path, "/")
}

// trailingSlash 对应 user_trailingslashit()：是否带尾部斜杠取决于固定链接结构本身。
func (l *linker) trailingSlash(s string) string {
	s = strings.TrimRight(s, `/\`)
	if strings.HasSuffix(l.structure, "/") {
		return s + "/"
	}
	return s
}

// postLink 对应文章类型为 post 时的 get_permalink()。categoryIDs 是文章所属分类。
func (l *linker) postLink(row wpdb.WpPost, categoryIDs []uint64, authorNicename string) string {
	if l.structure == "" {
		return l.homeURL("?p=" + strconv.FormatUint(row.ID, 10))
	}
	d := row.PostDate
	link := strings.NewReplacer(
		"%year%", fmt.Sprintf("%04d", d.Year()),
		"%monthnum%", fmt.Sprintf("%02d", int(d.Month())),
		"%day%", fmt.Sprintf("%02d", d.Day()),
		"%hour%", fmt.Sprintf("%02d", d.Hour()),
		"%minute%", fmt.Sprintf("%02d", d.Minute()),
		"%second%", fmt.Sprintf("%02d", d.Second()),
		"%postname%", row.PostName,
		"%post_id%", strconv.FormatUint(row.ID, 10),
		"%category%", l.categoryPath(categoryIDs),
		"%author%", authorNicename,
		"%pagename%", row.PostName,
	).Replace(l.structure)
	return l.trailingSlash(l.homeURL(link))
}

// categoryPath 取 term_id 最小的分类，并带上 get_category_parents() 生成的父分类路径。
func (l *linker) categoryPath(categoryIDs []uint64) string {
	if l.categories == nil {
		return ""
	}
	id := l.defaultCategory
	if len(categoryIDs) > 0 {
		id = categoryIDs[0]
		for _, c := range categoryIDs[1:] {
			id = min(id, c)
		}
	}
	cat, ok := l.categories[id]
	if !ok {
		return ""
	}
	path := cat.Slug
	seen := map[uint64]bool{id: true}
	for parent := cat.Parent; parent != 0 && !seen[parent]; {
		seen[parent] = true
		p, found := l.categories[parent]
		if !found {
			break
		}
		path = p.Slug + "/" + path
		parent = p.Parent
	}
	return path
}

// commentsFeedLink 对应 get_post_comments_feed_link( $post_id, 'rss2' )。
func (l *linker) commentsFeedLink(postID uint64, permalink string) string {
	if l.structure == "" {
		return l.homeURL("/") + "?feed=rss2&p=" + strconv.FormatUint(postID, 10)
	}
	return l.trailingSlash(strings.TrimRight(permalink, `/\`) + "/feed")
}
