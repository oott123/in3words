package wpfmt

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

// autopAllBlocks 是 wpautop() 中的 $allblocks。
const autopAllBlocks = `(?:table|thead|tfoot|caption|col|colgroup|tbody|tr|td|th|div|dl|dd|dt|ul|ol|li|pre|form|map|area|blockquote|address|style|p|h[1-6]|hr|fieldset|legend|section|article|aside|hgroup|header|footer|nav|figure|figcaption|details|menu|summary)`

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	autopMultiBR        = regexp.MustCompile(`<br` + reSpace + `*/?>` + reSpace + `*<br` + reSpace + `*/?>`)
	autopBlockOpen      = regexp.MustCompile(`(<` + autopAllBlocks + `[\t\n\v\f\r />])`)
	autopBlockClose     = regexp.MustCompile(`(</` + autopAllBlocks + `>)`)
	autopHR             = regexp.MustCompile(`(<hr` + reSpace + `*?/?>)`)
	autopOptionBefore   = regexp.MustCompile(reSpace + `*<option`)
	autopOptionAfter    = regexp.MustCompile(`</option>` + reSpace + `*`)
	autopObjectOpen     = regexp.MustCompile(`(<object[^>]*>)` + reSpace + `*`)
	autopObjectClose    = regexp.MustCompile(reSpace + `*</object>`)
	autopParamEmbed     = regexp.MustCompile(reSpace + `*(</?(?:param|embed)[^>]*>)` + reSpace + `*`)
	autopMediaOpen      = regexp.MustCompile(`([<\[](?:audio|video)[^>\]]*[>\]])` + reSpace + `*`)
	autopMediaClose     = regexp.MustCompile(reSpace + `*([<\[]/(?:audio|video)[>\]])`)
	autopSourceTrack    = regexp.MustCompile(reSpace + `*(<(?:source|track)[^>]*>)` + reSpace + `*`)
	autopFigcaptionOpen = regexp.MustCompile(reSpace + `*(<figcaption[^>]*>)`)
	autopFigcaptionEnd  = regexp.MustCompile(`</figcaption>` + reSpace + `*`)
	autopManyNewlines   = regexp.MustCompile(`\n\n+`)
	autopParagraphSplit = regexp.MustCompile(`\n` + reSpace + `*\n`)
	autopEmptyP         = regexp.MustCompile(`<p>` + reSpace + `*</p>`)
	autopUnclosedP      = regexp.MustCompile(`<p>([^<]+)</(div|address|form)>`)
	autopWrappedBlock   = regexp.MustCompile(`<p>` + reSpace + `*(</?` + autopAllBlocks + `[^>]*>)` + reSpace + `*</p>`)
	autopWrappedLI      = regexp.MustCompile(`<p>(<li.+?)</p>`)
	autopBlockquote     = regexp.MustCompile(`(?i)<p><blockquote((?:[^>"']|"[^"]*"|'[^']*')*)>`)
	autopPBeforeBlock   = regexp.MustCompile(`<p>` + reSpace + `*(</?` + autopAllBlocks + `[^>]*>)`)
	autopPAfterBlock    = regexp.MustCompile(`(</?` + autopAllBlocks + `[^>]*>)` + reSpace + `*</p>`)
	autopBRAfterBlock   = regexp.MustCompile(`(</?` + autopAllBlocks + `[^>]*>)` + reSpace + `*<br />`)
	autopBRBeforeBlock  = regexp.MustCompile(`<br />(` + reSpace + `*</?(?:p|li|div|dl|dd|dt|th|pre|td|ul|ol)[^>]*>)`)
	autopTrailingNL     = regexp.MustCompile(`\n</p>(\n?)\z`)

	autopPreserveNL = mustRE(`<(script|style|svg|math).*?</\1>`, regexp2.Singleline)
	autopNewlineBR  = mustRE(`(?<!<br />)`+reSpace+`*\n`, 0)

	// get_html_split_regex().
	htmlSplitRegex = mustRE(
		`(<(?:(?=!--|!\[CDATA\[)`+
			`(?:(?=!-)!(?>(?:-(?!->)(?>[^\-]*))*)(?:-->)?`+
			`|(?!!-)!\[CDATA\[(?>[^\]]*)(?>(?:\](?!\]>)(?>[^\]]*))*)(?:\]\]>)?)`+
			`|(?!!--|!\[CDATA\[)[^>]*>?))`,
		0,
	)
)

// wpautop 移植自 wpautop($text, true)。
//
//nolint:funlen // 逐行移植 WordPress 实现，保持原有结构便于对照
func wpautop(text string) string {
	if phpTrim(text) == "" {
		return ""
	}

	text += "\n"

	// pre 标签不应被处理：先替换为占位符，最后再还原。
	var preTags []string
	if strings.Contains(text, "<pre") {
		parts := strings.Split(text, "</pre>")
		last := parts[len(parts)-1]
		var b strings.Builder
		for _, part := range parts[:len(parts)-1] {
			start := strings.Index(part, "<pre")
			if start < 0 {
				// HTML 格式有误。
				b.WriteString(part)
				continue
			}
			name := "<pre wp-pre-tag-" + strconv.Itoa(len(preTags)) + "></pre>"
			preTags = append(preTags, name, part[start:]+"</pre>")
			b.WriteString(part[:start])
			b.WriteString(name)
		}
		b.WriteString(last)
		text = b.String()
	}

	text = autopMultiBR.ReplaceAllString(text, "\n\n")
	text = autopBlockOpen.ReplaceAllString(text, "\n\n$1")
	text = autopBlockClose.ReplaceAllString(text, "$1\n\n")
	text = autopHR.ReplaceAllString(text, "$1\n\n")
	text = strReplacePairs(text, "\r\n", "\n", "\r", "\n")

	// 标签内部的换行替换为占位符。
	text = wpReplaceInHTMLTags(text, "\n", " <!-- wpnl --> ")

	if strings.Contains(text, "<option") {
		text = autopOptionBefore.ReplaceAllString(text, "<option")
		text = autopOptionAfter.ReplaceAllString(text, "</option>")
	}

	if strings.Contains(text, "</object>") {
		text = autopObjectOpen.ReplaceAllString(text, "$1")
		text = autopObjectClose.ReplaceAllString(text, "</object>")
		text = autopParamEmbed.ReplaceAllString(text, "$1")
	}

	if strings.Contains(text, "<source") || strings.Contains(text, "<track") {
		text = autopMediaOpen.ReplaceAllString(text, "$1")
		text = autopMediaClose.ReplaceAllString(text, "$1")
		text = autopSourceTrack.ReplaceAllString(text, "$1")
	}

	if strings.Contains(text, "<figcaption") {
		text = autopFigcaptionOpen.ReplaceAllString(text, "$1")
		text = autopFigcaptionEnd.ReplaceAllString(text, "</figcaption>")
	}

	text = autopManyNewlines.ReplaceAllString(text, "\n\n")

	var b strings.Builder
	for _, paragraph := range autopParagraphSplit.Split(text, -1) {
		if paragraph == "" {
			continue
		}
		b.WriteString("<p>")
		b.WriteString(strings.Trim(paragraph, "\n"))
		b.WriteString("</p>\n")
	}
	text = b.String()

	text = autopEmptyP.ReplaceAllString(text, "")
	text = autopUnclosedP.ReplaceAllString(text, "<p>$1</p></$2>")
	text = autopWrappedBlock.ReplaceAllString(text, "$1")
	text = autopWrappedLI.ReplaceAllString(text, "$1")
	text = autopBlockquote.ReplaceAllString(text, "<blockquote$1><p>")
	text = strings.ReplaceAll(text, "</blockquote></p>", "</p></blockquote>")
	text = autopPBeforeBlock.ReplaceAllString(text, "$1")
	text = autopPAfterBlock.ReplaceAllString(text, "$1")

	// 插入换行标签。
	text, _ = replaceFunc(autopPreserveNL, text, func(m *reMatch) string {
		return strings.ReplaceAll(m.group(0), "\n", "<WPPreserveNewline />")
	})
	text = strReplacePairs(text, "<br>", "<br />", "<br/>", "<br />")
	text = replaceAll(autopNewlineBR, text, "<br />\n")
	text = strings.ReplaceAll(text, "<WPPreserveNewline />", "\n")

	text = autopBRAfterBlock.ReplaceAllString(text, "$1")
	text = autopBRBeforeBlock.ReplaceAllString(text, "$1")
	text = autopTrailingNL.ReplaceAllString(text, "</p>$1")

	if len(preTags) > 0 {
		text = strReplacePairs(text, preTags...)
	}

	if strings.Contains(text, "<!-- wpnl -->") {
		text = strReplacePairs(text, " <!-- wpnl --> ", "\n", "<!-- wpnl -->", "\n")
	}

	return text
}

// wpHTMLSplit 对应 wp_html_split()：奇数下标是 HTML 元素或注释，偶数下标是文本。
func wpHTMLSplit(input string) []string {
	return splitDelim(htmlSplitRegex, input, false)
}

// wpReplaceInHTMLTags 对应 wp_replace_in_html_tags() 只有一组替换时的情况。
func wpReplaceInHTMLTags(haystack, needle, replace string) string {
	parts := wpHTMLSplit(haystack)
	changed := false
	for i := 1; i < len(parts); i += 2 {
		if strings.Contains(parts[i], needle) {
			parts[i] = strings.ReplaceAll(parts[i], needle, replace)
			changed = true
		}
	}
	if !changed {
		return haystack
	}
	return strings.Join(parts, "")
}

// shortcodeUnautopRegex 对应 shortcode_unautop() 使用的正则（基于全部已注册短代码）。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var shortcodeUnautopRegex = mustRE(
	`<p>(?>(?:`+reSpaces+`)*)`+
		`(\[(`+registeredShortcodesPattern()+`)(?![A-Za-z0-9_-])`+
		`[^\]/]*(?:/(?!\])[^\]/]*)*?`+
		`(?:/\]|\](?:(?>[^\[]*)(?>(?:\[(?!/\2\])(?>[^\[]*))*)\[/\2\])?))`+
		`(?>(?:`+reSpaces+`)*)</p>`,
	0,
)

// shortcodeUnautop 移植自 shortcode_unautop()。
func shortcodeUnautop(text string) string {
	return replaceAll(shortcodeUnautopRegex, text, "$1")
}
