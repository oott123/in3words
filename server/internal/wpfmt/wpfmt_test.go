package wpfmt_test

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oott123/in3words/server/internal/wpfmt"
)

// golden 是 testdata/gen.php 在 WordPress Playground 中生成的期望输出。
type golden struct {
	WPVersion    string       `json:"wp_version"`
	Home         string       `json:"home"`
	SiteURL      string       `json:"siteurl"`
	HTML5Caption bool         `json:"html5_caption"`
	UseSmilies   bool         `json:"use_smilies"`
	Content      []goldenCase `json:"content"`
	AfterBlocks  []goldenCase `json:"content_after_blocks"`
	Title        []goldenCase `json:"title"`
	Comment      []goldenCase `json:"comment"`
	Sanitize     []goldenCase `json:"sanitize"`
	Feed         goldenFeed   `json:"feed"`
	FeedFuncs    []struct {
		Input    string `json:"input"`
		Ent2ncr  string `json:"ent2ncr"`
		TitleRSS string `json:"title_rss"`
		Bloginfo string `json:"bloginfo"`
		DocTitle string `json:"document_title"`
		TermName string `json:"term_name"`
	} `json:"feed_funcs"`
	EscURL          []goldenCase      `json:"esc_url"`
	UploadsBaseURL  string            `json:"uploads_baseurl"`
	AttachmentsMeta map[string]string `json:"attachments_meta"`
	Attachments     []goldenCase      `json:"attachments"`
}

// attachments 把金标准中的附件元数据转换为 wpfmt.Attachments。
func (g *golden) attachments(t *testing.T) wpfmt.Attachments {
	t.Helper()
	att := wpfmt.Attachments{}
	for id, meta := range g.AttachmentsMeta {
		n, err := strconv.ParseUint(id, 10, 64)
		require.NoError(t, err)
		att[n] = meta
	}
	return att
}

// goldenFeed 是真实 feed-rss2.php 渲染结果中与正文、标题、分类相关的部分。
type goldenFeed struct {
	DescriptionInput string `json:"description_input"`
	Description      string `json:"description"`
	Items            []struct {
		Name            string   `json:"name"`
		ID              uint64   `json:"id"`
		Input           string   `json:"input"`
		Expected        string   `json:"expected"`
		TitleInput      string   `json:"title_input"`
		Title           string   `json:"title"`
		CategoriesInput []string `json:"categories_input"`
		Categories      []string `json:"categories"`
	} `json:"items"`
}

type goldenCase struct {
	Name     string `json:"name"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

func loadGolden(t *testing.T) (*golden, wpfmt.Options) {
	t.Helper()
	data, err := os.ReadFile("testdata/golden.json")
	require.NoError(t, err)
	var g golden
	require.NoError(t, json.Unmarshal(data, &g))

	opts := wpfmt.DefaultOptions()
	opts.HomeURL = g.Home
	opts.SiteURL = g.SiteURL
	opts.HTML5Caption = g.HTML5Caption
	opts.UseSmilies = g.UseSmilies
	opts.UploadsBaseURL = g.UploadsBaseURL
	return &g, opts
}

// wpContainerClass 匹配依赖主题 theme.json 的 wp-container-* 类名，本包有意不生成它。
var wpContainerClass = regexp.MustCompile(` wp-container-[a-z0-9-]+`)

func TestTheContentGolden(t *testing.T) {
	t.Parallel()
	g, opts := loadGolden(t)
	for _, c := range g.Content {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			expected := wpContainerClass.ReplaceAllString(c.Expected, "")
			assert.Equal(t, expected, wpfmt.TheContent(c.Input, opts, nil), "input: %q", c.Input)
		})
	}
}

// TestTheContentAfterBlocksGolden 覆盖同一请求中渲染过含区块内容之后的过滤器顺序。
func TestTheContentAfterBlocksGolden(t *testing.T) {
	t.Parallel()
	g, opts := loadGolden(t)
	opts.AutopAfterUnautop = true
	require.NotEmpty(t, g.AfterBlocks)
	for _, c := range g.AfterBlocks {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.Expected, wpfmt.TheContent(c.Input, opts, nil), "input: %q", c.Input)
		})
	}
}

func TestHasBlocks(t *testing.T) {
	t.Parallel()
	assert.True(t, wpfmt.HasBlocks("<!-- wp:paragraph --><p>x</p><!-- /wp:paragraph -->"))
	assert.False(t, wpfmt.HasBlocks("<p>classic</p><!-- more -->"))
}

// TestFeedRendererGolden 按 feed 中的顺序渲染，验证跨文章状态（媒体计数、fetchpriority、
// wpautop 顺序、caption id 去重）。
func TestFeedRendererGolden(t *testing.T) {
	t.Parallel()
	g, opts := loadGolden(t)
	require.NotEmpty(t, g.Feed.Items)
	assert.Equal(t, g.Feed.Description, wpfmt.BloginfoRSS(g.Feed.DescriptionInput))

	att := g.attachments(t)
	r := wpfmt.NewFeedRenderer(opts)
	for _, item := range g.Feed.Items {
		assert.Equal(t, item.Expected, r.Content(item.Input, item.ID, att), "content of %s: %q", item.Name, item.Input)
		assert.Equal(t, item.Title, wpfmt.TitleRSS(item.TitleInput), "title of %s", item.Name)

		var categories []string
		for _, name := range item.CategoriesInput {
			if c := wpfmt.TermNameRSS(name); !slices.Contains(categories, c) {
				categories = append(categories, c)
			}
		}
		assert.Equal(t, item.Categories, categories, "categories of %s", item.Name)
	}
}

// TestAttachmentsGolden 覆盖 wp-image-{id} 图片的 width/height、srcset、sizes 补充。
func TestAttachmentsGolden(t *testing.T) {
	t.Parallel()
	g, opts := loadGolden(t)
	require.NotEmpty(t, g.Attachments)
	att := g.attachments(t)
	for _, c := range g.Attachments {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.Expected, wpfmt.TheContent(c.Input, opts, att), "input: %q", c.Input)
		})
	}
}

func TestAttachmentIDs(t *testing.T) {
	t.Parallel()
	content := `<img class="wp-image-3" src="a.jpg"><img class="WP-IMAGE-5 wp-image-7" src="b.jpg">` +
		`<p class="wp-image-9">x</p><img src="c.jpg" class="wp-image-3"><iframe class="wp-image-11" src="d"></iframe>`
	assert.Equal(t, []uint64{3, 5}, wpfmt.AttachmentIDs(content))
	assert.Empty(t, wpfmt.AttachmentIDs("<p>no images</p>"))
}

// TestCorruptAttachmentMetadata 确认损坏或非数组的元数据按"附件不存在"处理。
func TestCorruptAttachmentMetadata(t *testing.T) {
	t.Parallel()
	opts := wpfmt.DefaultOptions()
	opts.UploadsBaseURL = "https://example.com/wp-content/uploads"
	input := `<img class="wp-image-1" src="https://example.com/wp-content/uploads/a-300x200.jpg" alt="">`
	expected := wpfmt.TheContent(input, opts, nil)
	for _, meta := range []string{"garbage", `s:3:"abc";`, "a:0:{}", `a:1:{s:5:"sizes";s:1:"x";}`, `a:1:{s:4:"file";O:8:"stdClass":0:{}}`} {
		assert.Equal(t, expected, wpfmt.TheContent(input, opts, wpfmt.Attachments{1: meta}), "meta: %q", meta)
	}
}

func TestFeedFuncsGolden(t *testing.T) {
	t.Parallel()
	g, _ := loadGolden(t)
	require.NotEmpty(t, g.FeedFuncs)
	for _, c := range g.FeedFuncs {
		assert.Equal(t, c.Ent2ncr, wpfmt.Ent2ncr(c.Input), "Ent2ncr(%q)", c.Input)
		assert.Equal(t, c.TitleRSS, wpfmt.TitleRSS(c.Input), "TitleRSS(%q)", c.Input)
		assert.Equal(t, c.Bloginfo, wpfmt.BloginfoRSS(c.Input), "BloginfoRSS(%q)", c.Input)
		assert.Equal(t, c.TermName, wpfmt.TermNameRSS(c.Input), "TermNameRSS(%q)", c.Input)
		assert.Equal(t, c.DocTitle, wpfmt.DocumentTitle(c.Input), "DocumentTitle(%q)", c.Input)
	}
	for _, c := range g.EscURL {
		assert.Equal(t, c.Expected, wpfmt.EscURL(c.Input), "EscURL(%q)", c.Input)
	}
}

func TestTheTitleGolden(t *testing.T) {
	t.Parallel()
	g, _ := loadGolden(t)
	for _, c := range g.Title {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.Expected, wpfmt.TheTitle(c.Input), "input: %q", c.Input)
		})
	}
}

func TestCommentTextGolden(t *testing.T) {
	t.Parallel()
	g, opts := loadGolden(t)
	for _, c := range g.Comment {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.Expected, wpfmt.CommentText(c.Input, opts), "input: %q", c.Input)
		})
	}
}

func TestSanitizeTitleGolden(t *testing.T) {
	t.Parallel()
	g, _ := loadGolden(t)
	for _, c := range g.Sanitize {
		t.Run(c.Input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.Expected, wpfmt.SanitizeTitle(c.Input))
		})
	}
}

func TestReplaceInsecureHomeURL(t *testing.T) {
	t.Parallel()
	opts := wpfmt.DefaultOptions()
	opts.HomeURL = "https://example.com"
	opts.SiteURL = "https://example.com"
	opts.HTTPSMigrationRequired = true
	input := `<a href="http://example.com/x">x</a> http:\/\/example.com\/y http://other.com/z`
	expected := "<p><a href=\"https://example.com/x\">x</a> https:\\/\\/example.com\\/y http://other.com/z</p>\n"
	assert.Equal(t, expected, wpfmt.TheContent(input, opts, nil))

	opts.HTTPSMigrationRequired = false
	assert.Equal(
		t,
		"<p><a href=\"http://example.com/x\">x</a></p>\n",
		wpfmt.TheContent(`<a href="http://example.com/x">x</a>`, opts, nil),
	)
}
