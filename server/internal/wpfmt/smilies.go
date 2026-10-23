package wpfmt

import (
	"regexp"
	"sort"
	"strings"

	"github.com/dlclark/regexp2"
)

// smilies 是 smilies_init() 中默认的 $wpsmiliestrans。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var smilies = map[string]string{
	":mrgreen:": "mrgreen.png",
	":neutral:": "\xf0\x9f\x98\x90",
	":twisted:": "\xf0\x9f\x98\x88",
	":arrow:":   "\xe2\x9e\xa1",
	":shock:":   "\xf0\x9f\x98\xaf",
	":smile:":   "\xf0\x9f\x99\x82",
	":???:":     "\xf0\x9f\x98\x95",
	":cool:":    "\xf0\x9f\x98\x8e",
	":evil:":    "\xf0\x9f\x91\xbf",
	":grin:":    "\xf0\x9f\x98\x80",
	":idea:":    "\xf0\x9f\x92\xa1",
	":oops:":    "\xf0\x9f\x98\xb3",
	":razz:":    "\xf0\x9f\x98\x9b",
	":roll:":    "\xf0\x9f\x99\x84",
	":wink:":    "\xf0\x9f\x98\x89",
	":cry:":     "\xf0\x9f\x98\xa5",
	":eek:":     "\xf0\x9f\x98\xae",
	":lol:":     "\xf0\x9f\x98\x86",
	":mad:":     "\xf0\x9f\x98\xa1",
	":sad:":     "\xf0\x9f\x99\x81",
	"8-)":       "\xf0\x9f\x98\x8e",
	"8-O":       "\xf0\x9f\x98\xaf",
	":-(":       "\xf0\x9f\x99\x81",
	":-)":       "\xf0\x9f\x99\x82",
	":-?":       "\xf0\x9f\x98\x95",
	":-D":       "\xf0\x9f\x98\x80",
	":-P":       "\xf0\x9f\x98\x9b",
	":-o":       "\xf0\x9f\x98\xae",
	":-x":       "\xf0\x9f\x98\xa1",
	":-|":       "\xf0\x9f\x98\x90",
	";-)":       "\xf0\x9f\x98\x89",
	"8O":        "\xf0\x9f\x98\xaf",
	":(":        "\xf0\x9f\x99\x81",
	":)":        "\xf0\x9f\x99\x82",
	":?":        "\xf0\x9f\x98\x95",
	":D":        "\xf0\x9f\x98\x80",
	":P":        "\xf0\x9f\x98\x9b",
	":o":        "\xf0\x9f\x98\xae",
	":x":        "\xf0\x9f\x98\xa1",
	":|":        "\xf0\x9f\x98\x90",
	";)":        "\xf0\x9f\x98\x89",
	":!:":       "\xe2\x9d\x97",
	":?:":       "\xe2\x9d\x93",
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	smiliesSearch     = buildSmiliesSearch()
	smiliesTagSplit   = mustRE(`(<[^>]*>)`, 0)
	smiliesIgnoreTag  = regexp.MustCompile(`^<(code|pre|style|script|textarea)[^>]*>`)
	smileyImageExtReg = regexp.MustCompile(`\.([^.]+)$`)
)

// buildSmiliesSearch 按 smilies_init() 的方式构造 $wp_smiliessearch。
func buildSmiliesSearch() *regexp2.Regexp {
	keys := make([]string, 0, len(smilies))
	for k := range smilies {
		keys = append(keys, k)
	}
	// krsort()：按键名逆序排列，保证优先匹配更长的表情。
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	var b strings.Builder
	b.WriteString(`(?<=` + reSpaces + `|^)`)
	subchar := ""
	for _, smiley := range keys {
		first, rest := smiley[:1], smiley[1:]
		if first != subchar {
			if subchar != "" {
				b.WriteString(`)(?=` + reSpaces + `|$)`)
				b.WriteString(`|(?<=` + reSpaces + `|^)`)
			}
			subchar = first
			b.WriteString(regexp2.Escape(first) + `(?:`)
		} else {
			b.WriteString(`|`)
		}
		b.WriteString(regexp2.Escape(rest))
	}
	b.WriteString(`)(?=` + reSpaces + `|$)`)
	return mustRE(b.String(), regexp2.Multiline)
}

// convertSmilies 移植自 convert_smilies()。
func (st *renderState) convertSmilies(text string) string {
	if !st.opts.UseSmilies {
		return text
	}
	parts := splitDelim(smiliesTagSplit, text, false)
	ignore := ""
	var b strings.Builder
	for _, content := range parts {
		if ignore == "" {
			if m := smiliesIgnoreTag.FindStringSubmatch(content); m != nil {
				ignore = m[1]
			}
		}
		if ignore == "" && content != "" && content[0] != '<' {
			content, _ = replaceFunc(smiliesSearch, content, st.translateSmiley)
		}
		if ignore != "" && content == "</"+ignore+">" {
			ignore = ""
		}
		b.WriteString(content)
	}
	return b.String()
}

// translateSmiley 移植自 translate_smiley()。
func (st *renderState) translateSmiley(m *reMatch) string {
	smiley := phpTrim(m.group(0))
	img := smilies[smiley]
	ext := ""
	if em := smileyImageExtReg.FindStringSubmatch(img); em != nil {
		ext = strings.ToLower(em[1])
	}
	switch ext {
	case "jpg", "jpeg", "jpe", "gif", "png", "webp", "avif":
	default:
		// 不是图片的表情直接输出 emoji 字符。
		return img
	}
	src := strings.TrimRight(st.opts.SiteURL, "/") + "/wp-includes/images/smilies/" + img
	return `<img src="` + escURL(
		src,
	) + `" alt="` + escAttr(
		smiley,
	) + `" class="wp-smiley" style="height: 1em; max-height: 1em;" />`
}
