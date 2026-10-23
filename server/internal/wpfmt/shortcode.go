package wpfmt

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
)

// registeredShortcodes 是 WordPress 默认注册的短代码（顺序与注册顺序一致）。
// 它们都会参与 wptexturize / shortcode_unautop / do_shortcode 的正则构造，
// 但只有 caption、wp_caption、embed 真正实现了渲染，其余的原样保留。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var registeredShortcodes = []string{"wp_caption", "caption", "gallery", "playlist", "audio", "video", embedTag}

const embedTag = "embed"

func registeredShortcodesPattern() string {
	quoted := make([]string, len(registeredShortcodes))
	for i, t := range registeredShortcodes {
		quoted[i] = regexp2.Escape(t)
	}
	return strings.Join(quoted, "|")
}

// shortcodeAtts 对应 shortcode_parse_atts() 的返回值。
type shortcodeAtts struct {
	named      map[string]string
	positional []string
}

func (a *shortcodeAtts) get(name string) (string, bool) {
	if a == nil {
		return "", false
	}
	v, ok := a.named[name]
	return v, ok
}

// shortcodeHandler 渲染一个短代码；handled 为 false 时保留短代码原文。
type shortcodeHandler func(st *renderState, atts *shortcodeAtts, content string) (out string, handled bool)

// shortcodeLookup 相当于某一时刻的 $shortcode_tags，按名称返回处理函数。
type shortcodeLookup func(tag string) shortcodeHandler

var shortcodeRegexCache compileCache[*regexp2.Regexp] //nolint:gochecknoglobals // 正则缓存

// shortcodeRegex 对应 get_shortcode_regex($tagnames)。
func shortcodeRegex(tagnames []string) *regexp2.Regexp {
	return shortcodeRegexCache.get(strings.Join(tagnames, "|"), func() *regexp2.Regexp {
		quoted := make([]string, len(tagnames))
		for i, t := range tagnames {
			quoted[i] = regexp2.Escape(t)
		}
		return mustRE(
			`\[(\[?)(`+strings.Join(quoted, "|")+`)(?![A-Za-z0-9_-])`+
				`([^\]/]*(?:/(?!\])[^\]/]*)*?)`+
				`(?:(/)\]|\](?:((?>[^\[]*)(?>(?:\[(?!/\2\])(?>[^\[]*))*))\[/\2\])?)`+
				`(\]?)`,
			0,
		)
	})
}

// doShortcode 移植自 do_shortcode()，handlers 相当于当时的 $shortcode_tags。
func (st *renderState) doShortcode(
	content string,
	ignoreHTML bool,
	handlers shortcodeLookup,
	order []string,
) string {
	if !strings.Contains(content, "[") {
		return content
	}
	tagnames := foundShortcodeNames(content, order)
	if len(tagnames) == 0 {
		return content
	}

	content = st.doShortcodesInHTMLTags(content, ignoreHTML, tagnames, handlers)
	content, _ = replaceFunc(shortcodeRegex(tagnames), content, func(m *reMatch) string {
		return st.doShortcodeTag(m, handlers)
	})
	// unescape_invalid_shortcodes()
	return strReplacePairs(content, "&#91;", "[", "&#93;", "]")
}

// doShortcodeTag 移植自 do_shortcode_tag()。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func (st *renderState) doShortcodeTag(m *reMatch, handlers shortcodeLookup) string {
	// [[foo]] 用于转义短代码。
	if m.group(1) == "[" && m.group(6) == "]" {
		whole := m.group(0)
		return whole[1 : len(whole)-1]
	}
	handler := handlers(m.group(2))
	if handler == nil {
		return m.group(0)
	}
	out, handled := handler(st, shortcodeParseAtts(m.group(3)), m.group(5))
	if !handled {
		return m.group(0)
	}
	return m.group(1) + out + m.group(6)
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	ksesAttrParse  = mustRE(`^(<`+reSpace+`*)(/`+reSpace+`*)?([a-zA-Z0-9]+`+reSpace+`*)([^>]*)(>?)$`, 0)
	ksesXHTMLSlash = mustRE(reSpace+`*/`+reSpace+`*$`, 0)
	ksesHairRegex  = `(?:[_a-zA-Z][-_a-zA-Z0-9:.]*|\[\[?[^\[\]]+\]\]?)` +
		`(?:` + reSpace + `*=` + reSpace + `*(?:"[^"]*"|'[^']*'|[^\t\n\v\f\r "']+(?:` + reSpace + `|$))|(?:` + reSpace + `|$))` +
		reSpace + `*`
	ksesHairValidate     = mustRE(`^(?:`+ksesHairRegex+`)+$`, 0)
	ksesHairExtract      = mustRE(ksesHairRegex, 0)
	shortcodeInElemStart = mustRE(`^<`+reSpace+`*\[\[?[^\[\]]+\]`, 0)
)

// ksesAttrParse 移植自 wp_kses_attr_parse()：返回 [front, attr..., back]，无法解析时返回 nil。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func wpKsesAttrParse(element string) []string {
	m := findMatch(ksesAttrParse, element)
	if m == nil || m.group(2) != "" {
		return nil
	}
	attr := m.group(4)
	xhtmlSlash := ""
	if sm := findMatch(ksesXHTMLSlash, attr); sm != nil {
		xhtmlSlash = sm.group(0)
		attr = attr[:len(attr)-len(xhtmlSlash)]
	}

	// wp_kses_hair_parse()
	var attrs []string
	if attr != "" {
		if !isMatch(ksesHairValidate, attr) {
			return nil
		}
		for _, am := range findAllMatches(ksesHairExtract, attr, -1) {
			attrs = append(attrs, am.group(0))
		}
	}

	result := make([]string, 0, len(attrs)+2)
	result = append(result, m.group(1)+m.group(2)+m.group(3))
	result = append(result, attrs...)
	return append(result, xhtmlSlash+m.group(5))
}

// doShortcodesInHTMLTags 移植自 do_shortcodes_in_html_tags()。
// 与 WP 的差异：属性中的短代码展开后不经过 wp_kses_one_attr() 过滤。
//
//nolint:gocognit // 逐行移植 WordPress 实现，保持原有结构便于对照
func (st *renderState) doShortcodesInHTMLTags(
	content string,
	ignoreHTML bool,
	tagnames []string,
	handlers shortcodeLookup,
) string {
	content = strReplacePairs(content, "&#91;", "&#091;", "&#93;", "&#093;")
	encode := func(s string) string { return strReplacePairs(s, "[", "&#91;", "]", "&#93;") }
	pattern := shortcodeRegex(tagnames)
	expand := func(s string) (string, int) {
		return replaceFunc(pattern, s, func(m *reMatch) string { return st.doShortcodeTag(m, handlers) })
	}

	parts := wpHTMLSplit(content)
	for i, element := range parts {
		if element == "" || element[0] != '<' {
			continue
		}
		noOpen := !strings.Contains(element, "[")
		noClose := !strings.Contains(element, "]")
		if noOpen || noClose {
			if noOpen != noClose {
				parts[i] = encode(element)
			}
			continue
		}
		if ignoreHTML || strings.HasPrefix(element, "<!--") || strings.HasPrefix(element, "<![CDATA[") {
			parts[i] = encode(element)
			continue
		}

		attributes := wpKsesAttrParse(element)
		if attributes == nil {
			// 某些插件会写出 [name] <[email]> 这样的内容。
			if isMatch(shortcodeInElemStart, element) {
				element, _ = expand(element)
			}
			parts[i] = encode(element)
			continue
		}

		front, back := attributes[0], attributes[len(attributes)-1]
		attrs := attributes[1 : len(attributes)-1]
		for j, attr := range attrs {
			open := strings.IndexByte(attr, '[')
			if open < 0 || !strings.Contains(attr, "]") {
				continue
			}
			double := strings.IndexByte(attr, '"')
			single := strings.IndexByte(attr, '\'')
			if (single < 0 || open < single) && (double < 0 || open < double) {
				// 短代码出现在属性名位置，或者属性值没有引号。
				attrs[j], _ = expand(attr)
			} else if newAttr, count := expand(attr); count > 0 && phpTrim(newAttr) != "" {
				attrs[j] = newAttr
			}
		}
		parts[i] = encode(front + strings.Join(attrs, "") + back)
	}
	return strings.Join(parts, "")
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	shortcodeAttsRegex = mustRE(
		`([A-Za-z0-9_-]+)`+reSpace+`*=`+reSpace+`*"([^"]*)"(?:`+reSpace+`|$)`+
			`|([A-Za-z0-9_-]+)`+reSpace+`*=`+reSpace+`*'([^']*)'(?:`+reSpace+`|$)`+
			`|([A-Za-z0-9_-]+)`+reSpace+`*=`+reSpace+`*([^\t\n\v\f\r '"]+)(?:`+reSpace+`|$)`+
			`|"([^"]*)"(?:`+reSpace+`|$)`+
			`|'([^']*)'(?:`+reSpace+`|$)`+
			`|([^\t\n\v\f\r ]+)(?:`+reSpace+`|$)`,
		0,
	)
	shortcodeAttsSpaces   = regexp.MustCompile(`[\x{00a0}\x{200b}]+`)
	shortcodeAttsBalanced = mustRE(`^(?>[^<]*)(?>(?:<(?>[^>]*)>(?>[^<]*))*)$`, 0)
)

// shortcodeParseAtts 移植自 shortcode_parse_atts()。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func shortcodeParseAtts(text string) *shortcodeAtts {
	atts := &shortcodeAtts{named: map[string]string{}}
	text = shortcodeAttsSpaces.ReplaceAllString(text, " ")
	var order []string
	for _, m := range findAllMatches(shortcodeAttsRegex, text, -1) {
		switch {
		case !phpEmpty(m.group(1)):
			atts.set(strings.ToLower(m.group(1)), stripcslashes(m.group(2)), &order)
		case !phpEmpty(m.group(3)):
			atts.set(strings.ToLower(m.group(3)), stripcslashes(m.group(4)), &order)
		case !phpEmpty(m.group(5)):
			atts.set(strings.ToLower(m.group(5)), stripcslashes(m.group(6)), &order)
		case m.isset(7) && m.group(7) != "":
			atts.positional = append(atts.positional, stripcslashes(m.group(7)))
		case m.isset(8) && m.group(8) != "":
			atts.positional = append(atts.positional, stripcslashes(m.group(8)))
		case m.isset(9):
			atts.positional = append(atts.positional, stripcslashes(m.group(9)))
		}
	}

	// 拒绝未闭合的 HTML 元素。
	reject := func(v string) string {
		if strings.Contains(v, "<") && !isMatch(shortcodeAttsBalanced, v) {
			return ""
		}
		return v
	}
	for _, k := range order {
		atts.named[k] = reject(atts.named[k])
	}
	for i, v := range atts.positional {
		atts.positional[i] = reject(v)
	}
	return atts
}

func (a *shortcodeAtts) set(name, value string, order *[]string) {
	if _, ok := a.named[name]; !ok {
		*order = append(*order, name)
	}
	a.named[name] = value
}

// stripcslashes 移植自 PHP 的 stripcslashes()。
//
//nolint:gocognit,mnd // 逐行移植 WordPress 实现，保持原有结构便于对照；正则分组编号与阈值沿用 WordPress 原实现
func stripcslashes(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' || i+1 >= len(s) {
			b.WriteByte(c)
			continue
		}
		i++
		switch c = s[i]; c {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case 'a':
			b.WriteByte('\a')
		case 'v':
			b.WriteByte('\v')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'x':
			if i+1 < len(s) && isHexDigit(s[i+1]) {
				v, n := 0, 0
				for n < 2 && i+1 < len(s) && isHexDigit(s[i+1]) {
					i++
					v = v*16 + hexValue(s[i])
					n++
				}
				b.WriteByte(byte(v))
			} else {
				b.WriteByte('x')
			}
		default:
			if c >= '0' && c <= '7' {
				v, n := int(c-'0'), 1
				for n < 3 && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '7' {
					i++
					v = v*8 + int(s[i]-'0')
					n++
				}
				b.WriteByte(byte(v))
			} else {
				b.WriteByte(c)
			}
		}
	}
	return b.String()
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// hexValue 返回十六进制数字字符对应的值。
func hexValue(c byte) int {
	const letterBase = 10 // a / A 对应的值
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + letterBase
	default:
		return int(c-'A') + letterBase
	}
}

// contentShortcodes 返回 the_content 第 11 优先级 do_shortcode 时注册的短代码。
func contentShortcodes(tag string) shortcodeHandler {
	switch tag {
	case "wp_caption", "caption":
		return imgCaptionShortcode
	case embedTag:
		// embed 在 do_shortcode 阶段注册为 __return_false，真正的处理在 run_shortcode 阶段。
		return func(*renderState, *shortcodeAtts, string) (string, bool) { return "", true }
	default:
		// gallery / playlist / audio / video 需要读取附件数据，未实现，保留原文。
		return nil
	}
}

// embedShortcodes 返回 WP_Embed::run_shortcode() 期间注册的短代码。
func embedShortcodes(tag string) shortcodeHandler {
	if tag == embedTag {
		return embedShortcode
	}
	return nil
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var embedShortcodeOrder = []string{embedTag}

// embedShortcode 是 WP_Embed::shortcode() 的简化版本。
// 与 WP 的差异：不查询 oEmbed 缓存（postmeta / oembed_cache 文章）也不发起网络请求，
// 也不处理内置的 YouTube / 音视频 embed handler，统一按"未知 URL"输出链接。
func embedShortcode(_ *renderState, atts *shortcodeAtts, url string) (string, bool) {
	if url == "" {
		url, _ = atts.get("src")
	}
	if url == "" {
		return "", true
	}
	url = strings.ReplaceAll(url, "&amp;", "&")
	return `<a href="` + escURL(url) + `">` + escHTML(url) + `</a>`, true
}

// autoembedLineMark 是 autoembed() 中标签内换行的占位符。
const autoembedLineMark = "<!-- wp-line-break -->"

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	autoembedCheck   = regexp.MustCompile(`(?i)(?:^|[\t\n\v\f\r >])https?://`)
	autoembedOwnLine = mustRE(
		`^(`+reSpace+`*)(https?://[^\t\n\v\f\r <>"]+)(`+reSpace+`*)$`,
		regexp2.IgnoreCase|regexp2.Multiline,
	)
	autoembedOwnPara = mustRE(
		`(<p(?: [^>]*)?>`+reSpace+`*)(https?://[^\t\n\v\f\r <>"]+)(`+reSpace+`*</p>)`,
		regexp2.IgnoreCase,
	)
)

// autoembed 移植自 WP_Embed::autoembed()：处理独占一行或独占一段的 URL。
//
// 与 WP 的差异：WP 会先尝试内置的 embed handler（YouTube embed 地址、音视频文件）、
// postmeta 中的 oEmbed 缓存，最后发起 oEmbed 请求；本移植全部按"未知 URL"处理，
// 即 WP_Embed::shortcode() 失败分支的结果——URL 中的 &amp; 被还原为 &，其余不变。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func autoembed(content string) string {
	content = wpReplaceInHTMLTags(content, "\n", autoembedLineMark)
	if autoembedCheck.MatchString(content) {
		callback := func(m *reMatch) string {
			return m.group(1) + strings.ReplaceAll(m.group(2), "&amp;", "&") + m.group(3)
		}
		content, _ = replaceFunc(autoembedOwnLine, content, callback)
		content, _ = replaceFunc(autoembedOwnPara, content, callback)
	}
	return strings.ReplaceAll(content, autoembedLineMark, "\n")
}

var (
	captionNewStyle = regexp.MustCompile(
		`(?is)((?:<a [^>]+>` + reSpace + `*)?<img [^>]+>(?:` + reSpace + `*</a>)?)(.*)`,
	)
	captionTrailingN = regexp.MustCompile(`-1$`)
)

// imgCaptionShortcode 移植自 img_caption_shortcode()。
// 与 WP 的差异：caption 属性中含 HTML 时不经过 wp_kses_post() 过滤；
// wp_unique_prefixed_id() 的计数器只在单次 TheContent 调用内有效。
func imgCaptionShortcode(st *renderState, atts *shortcodeAtts, content string) (string, bool) {
	caption, hasCaption := atts.get("caption")
	if !hasCaption {
		if m := captionNewStyle.FindStringSubmatch(content); m != nil {
			content = m[1]
			caption = phpTrim(m[2])
		}
	}

	attr := func(name, def string) string {
		if v, ok := atts.get(name); ok {
			return v
		}
		return def
	}
	id := attr("id", "")
	captionID := attr("caption_id", "")
	align := attr("align", "alignnone")
	extraClass := attr("class", "")
	width := phpIntval(attr("width", ""))

	if width < 1 || phpEmpty(caption) {
		return content, true
	}

	var idAttr, captionIDAttr, describedBy, uniqueID, captionIDValue string
	if !phpEmpty(id) {
		id = sanitizeHTMLClass(id)
		uniqueID = captionTrailingN.ReplaceAllString(st.uniquePrefixedID(id+"-"), "")
		idAttr = `id="` + escAttr(uniqueID) + `" `
	}
	if !phpEmpty(captionID) {
		captionID = sanitizeHTMLClass(captionID)
		captionIDValue = captionTrailingN.ReplaceAllString(st.uniquePrefixedID(captionID+"-"), "")
	} else if !phpEmpty(uniqueID) {
		captionIDValue = "caption-" + strings.ReplaceAll(uniqueID, "_", "-")
	}
	if !phpEmpty(captionIDValue) {
		captionIDAttr = `id="` + escAttr(captionIDValue) + `" `
		describedBy = `aria-describedby="` + escAttr(captionIDValue) + `" `
	}

	class := phpTrim("wp-caption " + align + " " + extraClass)

	captionWidth := width
	if !st.opts.HTML5Caption {
		// HTML5 的 caption 从来不额外加 10px。
		captionWidth += 10
	}
	style := ""
	if captionWidth != 0 {
		style = fmt.Sprintf(`style="width: %dpx" `, captionWidth)
	}

	inner := st.doShortcode(content, false, contentShortcodes, registeredShortcodes)
	if st.opts.HTML5Caption {
		return fmt.Sprintf(
			`<figure %s%s%sclass="%s">%s<figcaption %sclass="wp-caption-text">%s</figcaption></figure>`,
			idAttr, describedBy, style, escAttr(class), inner, captionIDAttr, caption,
		), true
	}
	return fmt.Sprintf(
		`<div %s%sclass="%s">%s<p %sclass="wp-caption-text">%s</p></div>`,
		idAttr, style, escAttr(class), strings.ReplaceAll(inner, "<img ", "<img "+describedBy), captionIDAttr, caption,
	), true
}

// uniquePrefixedID 对应 wp_unique_prefixed_id()。
func (st *renderState) uniquePrefixedID(prefix string) string {
	if st.uniqueIDs == nil {
		st.uniqueIDs = map[string]int{}
	}
	st.uniqueIDs[prefix]++
	return fmt.Sprintf("%s%d", prefix, st.uniqueIDs[prefix])
}
