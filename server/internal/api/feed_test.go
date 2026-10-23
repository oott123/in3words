package api_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type rssFeed struct {
	Channel struct {
		Title         string    `xml:"title"`
		AtomLink      atomLink  `xml:"http://www.w3.org/2005/Atom link"`
		Link          string    `xml:"link"`
		Description   string    `xml:"description"`
		LastBuildDate string    `xml:"lastBuildDate"`
		Language      string    `xml:"language"`
		Items         []rssItem `xml:"item"`
	} `xml:"channel"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type rssItem struct {
	Title      string   `xml:"title"`
	Link       string   `xml:"link"`
	Comments   string   `xml:"comments"`
	Creator    string   `xml:"http://purl.org/dc/elements/1.1/ creator"`
	PubDate    string   `xml:"pubDate"`
	Categories []string `xml:"category"`
	GUID       string   `xml:"guid"`
	// description 不比对：本服务输出全文，WP 输出摘要。
	Content    string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	CommentRSS string `xml:"http://wellformedweb.org/CommentAPI/ commentRss"`
	Slash      string `xml:"http://purl.org/rss/1.0/modules/slash/ comments"`
}

func TestFeedGolden(t *testing.T) {
	t.Parallel()

	handler, db := newTestServer(t)

	var variants []struct {
		Options map[string]string `json:"options"`
		Body    string            `json:"body"`
	}
	readJSON(t, "testdata/feeds.json", &variants)

	for _, v := range variants {
		setOptions(t.Context(), t, db, v.Options)

		var want rssFeed
		require.NoError(t, xml.Unmarshal([]byte(v.Body), &want))
		// 受密码保护的文章不进入本服务的 feed。
		items := want.Channel.Items[:0]
		for _, item := range want.Channel.Items {
			if !strings.HasPrefix(item.Title, "Protected: ") {
				items = append(items, item)
			}
		}
		want.Channel.Items = items

		canonical, other := "/feed", "/feed/"
		if strings.HasSuffix(v.Options["permalink_structure"], "/") {
			canonical, other = other, canonical
		}

		// 非规范地址按 redirect_canonical() 跳转，并保留查询参数。
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, other+"?x=1", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMovedPermanently, rec.Code)
		require.Equal(t, canonical+"?x=1", rec.Header().Get("Location"))

		req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, canonical, nil)
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "application/rss+xml; charset=UTF-8", rec.Header().Get("Content-Type"))

		var got rssFeed
		require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &got))
		require.NotEmpty(t, got.Channel.Items)
		require.Len(t, got.Channel.Items, len(want.Channel.Items))
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("feed mismatch for %v (-wp +ours):\n%s", v.Options, diff)
		}

		// 条件请求：ETag 匹配时返回 304。
		req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, canonical, nil)
		req.Header.Set("If-None-Match", rec.Header().Get("ETag"))
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotModified, rec.Code)
	}
}

func setOptions(ctx context.Context, t *testing.T, db *sqlx.DB, options map[string]string) {
	t.Helper()
	for name, value := range options {
		_, err := db.ExecContext(ctx, "DELETE FROM `"+testTablePrefix+"options` WHERE option_name = ?", name)
		require.NoError(t, err)
		_, err = db.ExecContext(
			ctx,
			"INSERT INTO `"+testTablePrefix+"options` (option_name, option_value, autoload) VALUES (?, ?, 'on')",
			name, value,
		)
		require.NoError(t, err)
	}
}

func TestLocalizedErrors(t *testing.T) {
	t.Parallel()

	handler, db := newTestServer(t)
	setOptions(t.Context(), t, db, map[string]string{"WPLANG": "zh_CN"})

	cases := map[string]string{
		"/wp-json/wp/v2/posts?page=0":    `{"code":"rest_invalid_param","message":"无效参数：page","data":{"status":400,"params":{"page":"page 必须大于或等于1"}}}`,
		"/wp-json/wp/v2/posts/99999":     `{"code":"rest_post_invalid_id","message":"文章 ID 无效。","data":{"status":404}}`,
		"/wp-json/wp/v2/no-such-route":   `{"code":"rest_no_route","message":"未找到匹配 URL 和请求方式的路由。","data":{"status":404}}`,
		"/wp-json/wp/v2/posts?page=9999": `{"code":"rest_post_invalid_page_number","message":"请求的页码大于总页数。","data":{"status":400}}`,
	}
	for path, want := range cases {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var got, expected any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.NoError(t, json.Unmarshal([]byte(want), &expected))
		require.Equal(t, expected, got, path)
	}
}
