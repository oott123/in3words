package wpfmt

import (
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
)

// clickableURLChar 对应 [\w\x80-\xff#%\~/@\[\]*(+=&$-]：PHP 按字节匹配，所有非 ASCII 字节都算 URL 字符。
const clickableURLChar = `(?:[A-Za-z0-9_#%~/@\[\]*(+=&$\-]|[^\x00-\x7F])`

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	clickableSplit     = mustRE(`(<[^<>]+>)`, 0)
	clickableCodeOpen  = regexp.MustCompile(`(?i)^<(?:code|pre|script|style)[\t\n\v\f\r >]`)
	clickableTagScheme = regexp.MustCompile(`^<` + reSpace + `*[A-Za-z0-9_]{1,20}://`)
	clickableURL       = mustRE(
		`([\t\n\v\f\r (<.,;:!?])`+
			`((?>[A-Za-z0-9_]{1,20})://(?=[^\t\n\v\f\r ]{1,2000}[\t\n\v\f\r ])`+
			`(?>`+clickableURLChar+`*)(?:['.,;:!?)](?>`+clickableURLChar+`+))*)`+
			`(\)?)`+
			`(\.[A-Za-z0-9_]{2,6})?`,
		0,
	)
	clickableWebFTP = mustRE(
		`([\t\n\v\f\r >])((www|ftp)\.(?:[A-Za-z0-9_#$%&~/.\-;:=,?@\[\]+]|[^\x00-\x7F])+)`,
		regexp2.IgnoreCase|regexp2.Singleline,
	)
	clickableEmail   = regexp.MustCompile(`(?i)([\t\n\v\f\r >])([.0-9a-z_+-]+)@(([0-9a-z-]+\.)+[0-9a-z]{2,})`)
	clickableNestedA = regexp.MustCompile(`(?i)(<a([ \r\n\t]+[^>]+?>|>))<a [^>]+?>([^>]+?)</a></a>`)
)

// makeClickable 移植自 make_clickable()（在 comment_text 过滤器中运行，因此链接会带上 ugc）。
//
//nolint:gocognit,mnd // 逐行移植 WordPress 实现，保持原有结构便于对照；正则分组编号与阈值沿用 WordPress 原实现
func (st *renderState) makeClickable(text string) string {
	var r strings.Builder
	nestedCodePre := 0
	for _, piece := range splitDelim(clickableSplit, text, false) {
		lower := strings.ToLower(piece)
		if clickableCodeOpen.MatchString(piece) {
			nestedCodePre++
		} else if nestedCodePre > 0 && (lower == "</code>" || lower == "</pre>" || lower == "</script>" || lower == "</style>") {
			nestedCodePre--
		}

		if nestedCodePre > 0 || phpEmpty(piece) || (piece[0] == '<' && !clickableTagScheme.MatchString(piece)) {
			r.WriteString(piece)
			continue
		}

		if len(piece) > 10000 {
			// 很长的字符串可能触发昂贵的边界情况，先按空白拆开。
			for _, chunk := range splitStrByWhitespace(piece, 2100) {
				if len(chunk) > 2101 {
					r.WriteString(chunk)
				} else {
					r.WriteString(st.makeClickable(chunk))
				}
			}
			continue
		}

		ret := " " + piece + " "
		ret, _ = replaceFunc(clickableURL, ret, st.makeURLClickable)
		ret, _ = replaceFunc(clickableWebFTP, ret, st.makeWebFTPClickable)
		ret = clickableEmail.ReplaceAllStringFunc(ret, func(s string) string {
			m := clickableEmail.FindStringSubmatch(s)
			email := m[2] + "@" + m[3]
			return m[1] + `<a href="mailto:` + email + `">` + email + `</a>`
		})
		r.WriteString(ret[1 : len(ret)-1])
	}
	return clickableNestedA.ReplaceAllString(r.String(), "${1}${3}</a>")
}

// makeURLClickable 移植自 _make_url_clickable_cb()。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func (st *renderState) makeURLClickable(m *reMatch) string {
	url := m.group(2)
	suffix := m.group(3)
	if m.group(3) == ")" && strings.Index(url, "(") > 0 {
		url += m.group(3)
		suffix = ""
	}
	if m.isset(4) && !phpEmpty(m.group(4)) {
		url += m.group(4)
	}
	// 只有括号成对时才把括号算进 URL。
	for strings.Count(url, "(") < strings.Count(url, ")") {
		pos := strings.LastIndex(url, ")")
		suffix = url[pos:] + suffix
		url = url[:pos]
	}

	url = escURL(url)
	if phpEmpty(url) {
		return m.group(0)
	}
	return m.group(1) + `<a href="` + url + `"` + st.clickableRelAttr(url) + `>` + url + `</a>` + suffix
}

// makeWebFTPClickable 移植自 _make_web_ftp_clickable_cb()。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func (st *renderState) makeWebFTPClickable(m *reMatch) string {
	ret := ""
	dest := "http://" + m.group(2)
	if last := dest[len(dest)-1]; strings.IndexByte(".,;:)", last) >= 0 {
		ret = string(last)
		dest = dest[:len(dest)-1]
	}
	dest = escURL(dest)
	if phpEmpty(dest) {
		return m.group(0)
	}
	return m.group(1) + `<a href="` + dest + `"` + st.clickableRelAttr(dest) + `>` + dest + `</a>` + ret
}

// clickableRelAttr 移植自 _make_clickable_rel_attr()（comment_text 上下文）。
func (st *renderState) clickableRelAttr(url string) string {
	var parts []string
	scheme := strings.ToLower(urlScheme(url))
	if (scheme == "http" || scheme == "https") && !st.isInternalLink(url) {
		parts = append(parts, "nofollow")
	}
	parts = append(parts, "ugc")
	return ` rel="` + escAttr(strings.Join(parts, " ")) + `"`
}

// urlScheme 返回 parse_url($url, PHP_URL_SCHEME) 的近似结果。
func urlScheme(url string) string {
	i := strings.IndexByte(url, ':')
	if i <= 0 {
		return ""
	}
	for j := range i {
		c := url[j]
		if !isASCIIAlpha(c) && (j == 0 || ((c < '0' || c > '9') && c != '+' && c != '-' && c != '.')) {
			return ""
		}
	}
	return url[:i]
}

// urlHost 返回 parse_url($url, PHP_URL_HOST) 的近似结果。
func urlHost(url string) string {
	scheme := urlScheme(url)
	rest := url[len(scheme):]
	if scheme != "" {
		rest = rest[1:]
	}
	if !strings.HasPrefix(rest, "//") {
		return ""
	}
	rest = rest[2:]
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		rest = rest[:i]
	}
	if i := strings.LastIndexByte(rest, '@'); i >= 0 {
		rest = rest[i+1:]
	}
	if strings.HasPrefix(rest, "[") {
		if i := strings.IndexByte(rest, ']'); i >= 0 {
			return rest[:i+1]
		}
	}
	if i := strings.LastIndexByte(rest, ':'); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// isInternalLink 对应 wp_is_internal_link()。
func (st *renderState) isInternalLink(link string) bool {
	link = strings.ToLower(link)
	if !allowedProtocols[urlScheme(link)] {
		return false
	}
	home := strings.ToLower(urlHost(st.opts.HomeURL))
	return urlHost(link) == home
}

// splitStrByWhitespace 移植自 _split_str_by_whitespace()。
func splitStrByWhitespace(text string, goal int) []string {
	var chunks []string
	nullspace := strings.Map(func(r rune) rune {
		if strings.ContainsRune("\r\n\t\v\f ", r) {
			return 0
		}
		return r
	}, text)
	for goal < len(nullspace) {
		pos := strings.LastIndexByte(nullspace[:goal+1], 0)
		if pos < 0 {
			next := strings.IndexByte(nullspace[goal+1:], 0)
			if next < 0 {
				break
			}
			pos = goal + 1 + next
		}
		chunks = append(chunks, text[:pos+1])
		text = text[pos+1:]
		nullspace = nullspace[pos+1:]
	}
	if !phpEmpty(text) {
		chunks = append(chunks, text)
	}
	return chunks
}
