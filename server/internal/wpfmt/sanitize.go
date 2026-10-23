package wpfmt

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// SanitizeTitle 对应 sanitize_title($title)（默认 'save' 上下文，含 remove_accents），
// 用于把 REST 请求中的 slug 参数规范化成数据库中保存的形式（wp_parse_slug_list）。
// 只实现 en_US 等非特殊区域的 remove_accents 规则。
func SanitizeTitle(title string) string {
	return sanitizeTitleWithDashes(removeAccents(title))
}

// removeAccents 移植自 remove_accents()（UTF-8 分支）。
func removeAccents(text string) string {
	if !hasHighByte(text) {
		return text
	}
	if !utf8.ValidString(text) {
		// WP 对非 UTF-8 字符串按 ISO-8859-1 处理；数据库内容都是 UTF-8，这里不做处理。
		return text
	}
	text = norm.NFC.String(text)
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if repl, ok := removeAccentsTable[r]; ok {
			b.WriteString(repl)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func hasHighByte(s string) bool {
	for i := range len(s) {
		if s[i] >= 0x80 {
			return true
		}
	}
	return false
}

var (
	sanitizeOctets       = regexp.MustCompile(`%([a-fA-F0-9][a-fA-F0-9])`)
	sanitizeOctetsMarked = regexp.MustCompile(`---([a-fA-F0-9][a-fA-F0-9])---`)
	sanitizeEntities     = regexp.MustCompile(`&.+?;`)
	sanitizeDisallowed   = regexp.MustCompile(`[^%a-z0-9 _-]`)
	sanitizeSpaces       = regexp.MustCompile(reSpace + `+`)
	sanitizeDashes       = regexp.MustCompile(`-+`)
)

// sanitizeSaveStrip 是 'save' 上下文中直接删除的编码序列。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var sanitizeSaveStrip = []string{
	"%c2%ad", "%c2%a1", "%c2%bf", "%c2%ab", "%c2%bb", "%e2%80%b9", "%e2%80%ba",
	"%e2%80%98", "%e2%80%99", "%e2%80%9c", "%e2%80%9d", "%e2%80%9a", "%e2%80%9b",
	"%e2%80%9e", "%e2%80%9f", "%e2%80%a2", "%c2%a9", "%c2%ae", "%c2%b0", "%e2%80%a6",
	"%e2%84%a2", "%c2%b4", "%cb%8a", "%cc%81", "%cd%81", "%cc%80", "%cc%84", "%cc%8c",
	"%e2%80%8b", "%e2%80%8c", "%e2%80%8d", "%e2%80%8e", "%e2%80%8f", "%e2%80%aa",
	"%e2%80%ab", "%e2%80%ac", "%e2%80%ad", "%e2%80%ae", "%ef%bb%bf", "%ef%bf%bc",
}

// sanitizeSaveDash 是 'save' 上下文中转换为连字符的编码序列。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var sanitizeSaveDash = []string{
	"%e2%80%80", "%e2%80%81", "%e2%80%82", "%e2%80%83", "%e2%80%84", "%e2%80%85",
	"%e2%80%86", "%e2%80%87", "%e2%80%88", "%e2%80%89", "%e2%80%8a", "%e2%80%a8",
	"%e2%80%a9", "%e2%80%af",
}

// sanitizeTitleWithDashes 移植自 sanitize_title_with_dashes($title, ”, 'save')。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func sanitizeTitleWithDashes(title string) string {
	title = stripTags(title)
	// 保留已经编码的字节，删除不属于编码的 %。
	title = sanitizeOctets.ReplaceAllString(title, "---$1---")
	title = strings.ReplaceAll(title, "%", "")
	title = sanitizeOctetsMarked.ReplaceAllString(title, "%$1")

	if utf8.ValidString(title) {
		title = utf8URIEncode(strings.ToLower(title), 200)
	}
	title = asciiLower(title)

	title = strReplacePairs(title, "%c2%a0", "-", "%e2%80%91", "-", "%e2%80%93", "-", "%e2%80%94", "-")
	for _, e := range []string{"&nbsp;", "&#8209;", "&#160;", "&ndash;", "&#8211;", "&mdash;", "&#8212;"} {
		title = strings.ReplaceAll(title, e, "-")
	}
	title = strings.ReplaceAll(title, "/", "-")
	for _, s := range sanitizeSaveStrip {
		title = strings.ReplaceAll(title, s, "")
	}
	for _, s := range sanitizeSaveDash {
		title = strings.ReplaceAll(title, s, "-")
	}
	title = strings.ReplaceAll(title, "%c3%97", "x")

	title = sanitizeEntities.ReplaceAllString(title, "")
	title = strings.ReplaceAll(title, ".", "-")
	title = sanitizeDisallowed.ReplaceAllString(title, "")
	title = sanitizeSpaces.ReplaceAllString(title, "-")
	title = sanitizeDashes.ReplaceAllString(title, "-")
	return strings.Trim(title, "-")
}

func asciiLower(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, s)
}

// utf8URIEncode 移植自 utf8_uri_encode($utf8_string, $length)。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func utf8URIEncode(s string, length int) string {
	var b strings.Builder
	unicodeLength := 0
	for _, r := range s {
		if r < 0x80 {
			if length > 0 && unicodeLength+1 > length {
				break
			}
			b.WriteRune(r)
			unicodeLength++
			continue
		}
		var buf [utf8.UTFMax]byte
		n := utf8.EncodeRune(buf[:], r)
		if length > 0 && unicodeLength+n*3 > length {
			break
		}
		for _, c := range buf[:n] {
			// 非 ASCII 字符的每个字节都 >= 0x80，dechex() 结果固定为两位。
			b.WriteString("%" + hexByte(c))
		}
		unicodeLength += n * 3
	}
	return b.String()
}

func hexByte(c byte) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[c>>4], digits[c&0x0f]})
}

// stripTags 是 PHP strip_tags() 的简化移植：删除 HTML 标签、注释与 PHP 标签。
//
//nolint:gocognit,mnd // 逐行移植 WordPress 实现，保持原有结构便于对照；正则分组编号与阈值沿用 WordPress 原实现
func stripTags(s string) string {
	if !strings.Contains(s, "<") {
		return s
	}
	var b strings.Builder
	depth := 0
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case depth == 0 && c == '<':
			if i+1 < len(s) && isHTMLSpace(s[i+1]) {
				b.WriteByte(c)
				continue
			}
			if strings.HasPrefix(s[i:], "<!--") {
				end := strings.Index(s[i+4:], "-->")
				if end < 0 {
					return b.String()
				}
				i += 4 + end + 2
				continue
			}
			depth = 1
		case depth > 0 && quote != 0:
			if c == quote {
				quote = 0
			}
		case depth > 0 && (c == '"' || c == '\''):
			quote = c
		case depth > 0 && c == '<':
			depth++
		case depth > 0 && c == '>':
			depth--
		case depth == 0:
			b.WriteByte(c)
		}
	}
	return b.String()
}
