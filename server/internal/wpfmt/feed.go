package wpfmt

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// FeedRenderer 按 WP 主循环（in_the_loop && is_main_query、is_feed）的语义依次渲染一次 feed 请求中的各篇文章，保存跨文章的状态。
//
// 跨文章的状态包括：渲染过含区块的内容后 wpautop 被挪到优先级 10 队尾（见 Options.AutopAfterUnautop）、
// 主循环中已出现的媒体元素数量（前 3 个不加 loading="lazy"）、fetchpriority="high" 只给一张图片、
// [caption] 的 id 去重计数。
//
// feed 模板在正文之前还会调用 the_excerpt_rss()。wp_trim_excerpt() 在调用 the_content 前会
// strip_shortcodes()，并临时移除 do_blocks 与 wp_filter_content_tags，因此摘要不会改变上述任何状态，
// 这里无需模拟。
type FeedRenderer struct {
	st        *renderState
	sawBlocks bool
}

// NewFeedRenderer 创建一次 feed 请求使用的渲染器。
func NewFeedRenderer(opts Options) *FeedRenderer {
	return &FeedRenderer{st: &renderState{opts: opts, inMainLoop: true}}
}

// Content 对应 get_the_content_feed('rss2')：$more=1、只取第 1 页（generate_postdata 的 <!--nextpage--> 处理）、
// <!--more--> 替换为 <span id="more-ID"></span>、<!--noteaser-->，然后 the_content、]]> 替换，
// 再执行 the_content_feed 上的默认过滤器（wp_staticize_emoji、_oembed_filter_feed_content）。
// 受密码保护的文章由调用方处理（WP 会输出密码表单）。att 的含义同 TheContent。
func (r *FeedRenderer) Content(rawContent string, postID uint64, att Attachments) string {
	content := feedPostContent(rawContent, postID)
	r.st.att = att

	base := r.st.opts.AutopAfterUnautop
	if r.sawBlocks {
		r.st.opts.AutopAfterUnautop = true
	}
	content, hadBlocks := r.st.theContent(content)
	r.st.opts.AutopAfterUnautop = base
	if hadBlocks {
		r.sawBlocks = true
	}

	content = strings.ReplaceAll(content, "]]>", "]]&gt;")
	content = staticizeEmoji(content, r.st.opts.EmojiBaseURL, r.st.opts.EmojiExt)
	return oembedFilterFeedContent(content)
}

var (
	moreTagRegex   = regexp.MustCompile(`<!--more(.*?)?-->`)
	moreBlockRegex = regexp.MustCompile(`<!-- /?wp:more(.*?) -->`)
)

// feedPostContent 对应 feed 主循环中 generate_postdata() + get_the_content() 的结果（$more = 1，第 1 页）。
func feedPostContent(raw string, postID uint64) string {
	// generate_postdata()：按 <!--nextpage--> 分页。
	pages := []string{raw}
	if strings.Contains(raw, "<!--nextpage-->") {
		content := strReplacePairs(raw,
			"\n<!--nextpage-->\n", "<!--nextpage-->",
			"\n<!--nextpage-->", "<!--nextpage-->",
			"<!--nextpage-->\n", "<!--nextpage-->",
			"<!-- wp:nextpage -->", "",
			"<!-- /wp:nextpage -->", "",
		)
		content = strings.TrimPrefix(content, "<!--nextpage-->")
		pages = strings.Split(content, "<!--nextpage-->")
	}

	// get_the_content()
	content := pages[0]
	parts := []string{content}
	hasTeaser := false
	if m := moreTagRegex.FindString(content); m != "" {
		if hasBlock("more", content) {
			// 去掉 core/more 区块的分隔注释，否则拆分后会残留。
			content = moreBlockRegex.ReplaceAllString(content, "")
		}
		parts = strings.SplitN(content, m, 2) //nolint:mnd // explode($matches[0], $content, 2)
		hasTeaser = true
	}

	// feed 中 $page 恒为 1，因此只要原文含有 <!--noteaser--> 就去掉摘要部分。
	teaser := parts[0]
	if hasTeaser && strings.Contains(raw, "<!--noteaser-->") {
		teaser = ""
	}
	if len(parts) > 1 {
		return teaser + `<span id="more-` + strconv.FormatUint(postID, 10) + `"></span>` + parts[1]
	}
	return teaser
}

// hasBlock 对应 has_block($name, $content)，name 为不带命名空间的核心区块名。
func hasBlock(name, content string) bool {
	return HasBlocks(content) &&
		(strings.Contains(content, "<!-- wp:core/"+name+" ") || strings.Contains(content, "<!-- wp:"+name+" "))
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	emojiPartialMap  = buildEmojiPartialMap()
	staticizeSplit   = mustRE(`(<.*?>)`, 0)
	staticizeIgnore  = regexp.MustCompile(`^<(code|pre|style|script|textarea)>`)
	hexEntityPattern = regexp.MustCompile(`&#x([0-9a-fA-F]+);`)
)

// buildEmojiPartialMap 把 _wp_emoji_list('partials') 转为字符到实体的映射。
func buildEmojiPartialMap() map[rune]string {
	m := make(map[rune]string, len(emojiPartials))
	for _, e := range emojiPartials {
		if r, ok := decodeHexEntity(e); ok {
			m[r] = e
		}
	}
	return m
}

func decodeHexEntity(entity string) (rune, bool) {
	sub := hexEntityPattern.FindStringSubmatch(entity)
	if sub == nil {
		return 0, false
	}
	v, err := strconv.ParseInt(sub[1], 16, 32)
	if err != nil {
		return 0, false
	}
	return rune(v), true
}

// decodeHexEntities 对应 html_entity_decode()，只处理 emoji 列表中出现的十六进制实体。
func decodeHexEntities(s string) string {
	return hexEntityPattern.ReplaceAllStringFunc(s, func(e string) string {
		if r, ok := decodeHexEntity(e); ok {
			return string(r)
		}
		return e
	})
}

// wpEncodeEmoji 移植自 wp_encode_emoji()：把 emoji 字符转为十六进制实体。
func wpEncodeEmoji(content string) string {
	var b strings.Builder
	changed := false
	for _, r := range content {
		if e, ok := emojiPartialMap[r]; ok {
			b.WriteString(e)
			changed = true
			continue
		}
		b.WriteRune(r)
	}
	if !changed {
		return content
	}
	return b.String()
}

func isASCII(s string) bool {
	for i := range len(s) {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// staticizeEmoji 移植自 wp_staticize_emoji()：把 emoji 替换为图片。
func staticizeEmoji(text, baseURL, ext string) string {
	text, possible := emojiCandidates(text)
	if len(possible) == 0 {
		return text
	}

	var b strings.Builder
	ignore := ""
	for _, content := range splitDelim(staticizeSplit, text, false) {
		if ignore == "" {
			if m := staticizeIgnore.FindStringSubmatch(content); m != nil {
				ignore = m[1]
			}
		}
		if ignore == "" && content != "" && content[0] != '<' && strings.Contains(content, "&#x") {
			content = replaceEmojiEntities(content, possible, baseURL, ext)
		}
		if ignore != "" && content == "</"+ignore+">" {
			ignore = ""
		}
		b.WriteString(content)
	}
	// 最后去掉残留的 U+FE0F。
	return strings.ReplaceAll(b.String(), "&#xfe0f;", "")
}

// emojiCandidates 把文本中的 emoji 编码为实体，并返回其中可能需要替换的 emoji 实体（按 WP 列表顺序）。
func emojiCandidates(text string) (string, []string) {
	if !strings.Contains(text, "&#x") {
		if isASCII(text) {
			return text, nil
		}
		encoded := wpEncodeEmoji(text)
		if encoded == text {
			return text, nil
		}
		text = encoded
	}
	var possible []string
	for _, e := range emojiEntities {
		if strings.Contains(text, e) {
			possible = append(possible, e)
		}
	}
	return text, possible
}

// replaceEmojiEntities 把文本片段中的 emoji 实体替换为图片。
func replaceEmojiEntities(content string, possible []string, baseURL, ext string) string {
	for _, e := range possible {
		if !strings.Contains(content, e) {
			continue
		}
		file := strings.ReplaceAll(e, ";&#x", "-")
		file = strReplacePairs(file, "&#x", "", ";", "")
		img := fmt.Sprintf(
			`<img src="%s" alt="%s" class="wp-smiley" style="height: 1em; max-height: 1em;" />`,
			baseURL+file+ext, decodeHexEntities(e),
		)
		content = strings.ReplaceAll(content, e, img)
	}
	return content
}

// oembedFilterFeedContent 移植自 _oembed_filter_feed_content()：去掉嵌入 iframe 的 style 属性。
func oembedFilterFeedContent(content string) string {
	if !strings.Contains(content, "<iframe") {
		return content
	}
	p := newTagProcessor(content)
	for p.nextTag("iframe") {
		if p.hasClass("wp-embedded-content") {
			p.removeAttribute("style")
		}
	}
	return p.updatedHTML()
}

// Ent2ncr 对应 ent2ncr()：把 HTML 命名实体转换为数字字符引用。
func Ent2ncr(s string) string {
	if !strings.ContainsAny(s, "&|") {
		return s
	}
	return strReplacePairs(s, ent2ncrPairs...)
}

// TitleRSS 对应非密码保护文章的 get_the_title() 加上 the_title_rss 过滤器
// （ent2ncr(8)、strip_tags(10)、esc_html(10)）。
func TitleRSS(rawTitle string) string {
	title := TheTitle(rawTitle)
	title = Ent2ncr(title)
	title = stripTags(title)
	return escHTML(title)
}

// BloginfoRSS 对应 get_bloginfo_rss()（strip_tags、convert_chars）加上 bloginfo_rss 过滤器（ent2ncr）。
// s 为 get_bloginfo() 返回的原始值（即 option 值）。
func BloginfoRSS(s string) string {
	return Ent2ncr(convertChars(stripTags(s)))
}

// TermNameRSS 对应 get_the_category_rss() 中写进 CDATA 的分类/标签名：
// sanitize_term_field('name', …, 'rss')（term_name_rss → convert_chars）后 html_entity_decode(ENT_COMPAT)。
func TermNameRSS(name string) string {
	return htmlEntityDecodeCompat(convertChars(name))
}

// DocumentTitle 对应 document_title 过滤器链：wptexturize、convert_chars、esc_html、
// capital_P_dangit(11，与 the_content 相同的分支)。feed 频道标题 wp_title_rss() 使用它。
func DocumentTitle(s string) string {
	return capitalPDangit(escHTML(convertChars(wptexturize(s))))
}

// EscURL 对应 esc_url()。
func EscURL(url string) string {
	return escURL(url)
}

var htmlEntityRegex = regexp.MustCompile(`&(?:#[0-9]+|#[xX][0-9a-fA-F]+|[A-Za-z][A-Za-z0-9]*);`)

// htmlEntityDecodeCompat 对应 html_entity_decode($s, ENT_COMPAT, 'UTF-8')：单引号实体保持不变。
// 与 PHP 的差异：命名实体使用 HTML5 实体表（PHP 在该参数下使用 HTML 4.01 实体表）。
func htmlEntityDecodeCompat(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	return htmlEntityRegex.ReplaceAllStringFunc(s, func(e string) string {
		decoded := html.UnescapeString(e)
		if decoded == "'" || decoded == e {
			return e
		}
		if strings.HasPrefix(e, "&#") {
			if r, _ := utf8.DecodeRuneInString(decoded); r == utf8.RuneError || !validUnicode(int(r)) {
				return e
			}
		}
		return decoded
	})
}
