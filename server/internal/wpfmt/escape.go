package wpfmt

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

var (
	normalizeNamedEntity = regexp.MustCompile(`&amp;([A-Za-z]{2,8}[0-9]{0,2});`)
	normalizeDecEntity   = regexp.MustCompile(`&amp;#(0*[0-9]{1,7});`)
	normalizeHexEntity   = regexp.MustCompile(`&amp;#[Xx](0*[0-9A-Fa-f]{1,6});`)
	specialCharsCheck    = regexp.MustCompile(`[&<>"']`)
)

// wpKsesNormalizeEntities 移植自 wp_kses_normalize_entities($content, 'html')。
func wpKsesNormalizeEntities(content string) string {
	if !strings.Contains(content, "&") {
		return content
	}
	content = strings.ReplaceAll(content, "&", "&amp;")
	content = normalizeNamedEntity.ReplaceAllStringFunc(content, func(s string) string {
		name := normalizeNamedEntity.FindStringSubmatch(s)[1]
		if allowedEntityNames[name] {
			return "&" + name + ";"
		}
		return s
	})
	content = normalizeDecEntity.ReplaceAllStringFunc(content, func(s string) string {
		digits := normalizeDecEntity.FindStringSubmatch(s)[1]
		n, err := strconv.Atoi(digits)
		if err != nil || !validUnicode(n) {
			return "&amp;#" + digits + ";"
		}
		trimmed := strings.TrimLeft(digits, "0")
		for len(trimmed) < 3 {
			trimmed = "0" + trimmed
		}
		return "&#" + trimmed + ";"
	})
	return normalizeHexEntity.ReplaceAllStringFunc(content, func(s string) string {
		digits := normalizeHexEntity.FindStringSubmatch(s)[1]
		n, err := strconv.ParseInt(digits, 16, 32)
		if err != nil || !validUnicode(int(n)) {
			return "&amp;#x" + digits + ";"
		}
		return "&#x" + strings.TrimLeft(digits, "0") + ";"
	})
}

// validUnicode 对应 valid_unicode()。
func validUnicode(i int) bool {
	return i == 0x9 || i == 0xa || i == 0xd ||
		(i >= 0x20 && i <= 0xd7ff) ||
		(i >= 0xe000 && i <= 0xfffd) ||
		(i >= 0x10000 && i <= 0x10ffff)
}

// wpSpecialChars 对应 _wp_specialchars($text, ENT_QUOTES, false, false)，即 esc_attr / esc_html 的核心。
func wpSpecialChars(text string) string {
	if text == "" || !specialCharsCheck.MatchString(text) {
		return text
	}
	text = wpKsesNormalizeEntities(text)
	// htmlspecialchars(..., ENT_QUOTES, double_encode = false) 只认 HTML 4.01 实体表，
	// 白名单中唯一不在该表里的 &apos; 会被再次编码。
	text = strings.ReplaceAll(text, "&apos;", "&amp;apos;")
	return strings.NewReplacer(`<`, "&lt;", `>`, "&gt;", `"`, "&quot;", `'`, "&#039;").Replace(text)
}

// escAttr 对应 esc_attr()。
func escAttr(text string) string { return wpSpecialChars(text) }

// escHTML 对应 esc_html()。
func escHTML(text string) string { return wpSpecialChars(text) }

var (
	escURLDisallowed = regexp.MustCompile(`[^a-zA-Z0-9\-~+_.?#=!&;,/:%@$|*'()\[\]\x{80}-\x{10FFFF}]`)
	escURLNoScheme   = regexp.MustCompile(`(?i)^[a-z0-9-]+?\.php`)
	escURLFront      = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.\-]*://[^/?#]*`)
	escURLProtocol   = regexp.MustCompile(`(?i):|&#0*58;|&#x0*3a;`)
)

// allowedProtocols 对应 wp_allowed_protocols() 的默认值。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var allowedProtocols = map[string]bool{
	"http": true, "https": true, "ftp": true, "ftps": true, "mailto": true, "news": true,
	"irc": true, "irc6": true, "ircs": true, "gopher": true, "nntp": true, "feed": true,
	"telnet": true, "mms": true, "rtsp": true, "sms": true, "svn": true, "tel": true,
	"fax": true, "xmpp": true, "webcal": true, "urn": true,
}

// escURL 是 esc_url($url) 的移植（display 上下文）。
// 与 WP 的差异：方括号处理与 wp_kses_bad_protocol() 做了简化。
//
//nolint:gocognit,nestif // 逐行移植 WordPress 实现，保持原有结构便于对照
func escURL(url string) string {
	if url == "" {
		return url
	}
	url = strings.ReplaceAll(strings.TrimLeft(url, " \t\n\r\x00\x0B"), " ", "%20")
	url = escURLDisallowed.ReplaceAllString(url, "")
	if url == "" {
		return url
	}

	if !strings.HasPrefix(strings.ToLower(url), "mailto:") {
		// _deep_replace()
		for {
			replaced := strReplacePairs(url, "%0d", "", "%0a", "", "%0D", "", "%0A", "")
			if replaced == url {
				break
			}
			url = replaced
		}
	}

	url = strings.ReplaceAll(url, ";//", "://")
	if !strings.Contains(url, ":") && !strings.ContainsAny(url[:1], "/#?") && !escURLNoScheme.MatchString(url) {
		url = "http://" + url
	}

	url = wpKsesNormalizeEntities(url)
	url = strings.ReplaceAll(url, "&amp;", "&#038;")
	url = strings.ReplaceAll(url, "'", "&#039;")

	if strings.ContainsAny(url, "[]") {
		front := escURLFront.FindString(url)
		end := url[len(front):]
		url = front + strReplacePairs(end, "[", "%5B", "]", "%5D")
	}

	if url[0] != '/' {
		// wp_kses_bad_protocol()：协议不在白名单中时返回空字符串。
		if loc := escURLProtocol.FindStringIndex(url); loc != nil {
			scheme := url[:loc[0]]
			if !strings.Contains(scheme, "/?") {
				scheme = strings.ToLower(strings.Map(func(r rune) rune {
					if r < 0x20 || r == 0x7f || r == ' ' {
						return -1
					}
					return r
				}, scheme))
				if !allowedProtocols[scheme] {
					return ""
				}
			}
		}
	}
	return url
}

var (
	sanitizeHTMLClassOctets = regexp.MustCompile(`%[a-fA-F0-9][a-fA-F0-9]`)
	sanitizeHTMLClassChars  = regexp.MustCompile(`[^A-Za-z0-9_-]`)
)

// sanitizeHTMLClass 对应 sanitize_html_class()。
func sanitizeHTMLClass(class string) string {
	class = sanitizeHTMLClassOctets.ReplaceAllString(class, "")
	return sanitizeHTMLClassChars.ReplaceAllString(class, "")
}

var phpNumericPrefix = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?`)

// phpIntval 对应 PHP 中 (int) $string。
func phpIntval(s string) int {
	s = strings.TrimLeft(s, " \t\n\r\v\f")
	num := phpNumericPrefix.FindString(s)
	if num == "" {
		return 0
	}
	f, err := strconv.ParseFloat(num, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	return int(f)
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var convertCharsRegex = mustRE(`&([^#])(?![a-z1-4]{1,8};)`, regexp2.IgnoreCase)

// convertChars 对应 convert_chars()。
func convertChars(content string) string {
	if !strings.Contains(content, "&") {
		return content
	}
	return replaceAll(convertCharsRegex, content, "&#038;$1")
}

// capitalPDangit 对应 capital_P_dangit() 在 the_content / comment_text 中的版本。
func capitalPDangit(text string) string {
	return strReplacePairs(text,
		" Wordpress", " WordPress",
		"&#8216;Wordpress", "&#8216;WordPress",
		texOpeningQuote+"Wordpress", texOpeningQuote+"WordPress",
		">Wordpress", ">WordPress",
		"(Wordpress", "(WordPress",
	)
}

// capitalPDangitTitle 对应 capital_P_dangit() 在 the_title 中的版本。
func capitalPDangitTitle(text string) string {
	return strings.ReplaceAll(text, "Wordpress", "WordPress")
}
