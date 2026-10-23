package wpfmt

import (
	"strings"

	"github.com/dlclark/regexp2"
)

// wptexturize() 使用的替换字符。这里使用 en_US 的默认翻译；
// 其他语言包可能会翻译这些字符（例如引号），本移植不支持。
const (
	texOpeningQuote       = "&#8220;"
	texClosingQuote       = "&#8221;"
	texApos               = "&#8217;"
	texPrime              = "&#8242;"
	texDoublePrime        = "&#8243;"
	texOpeningSingleQuote = "&#8216;"
	texClosingSingleQuote = "&#8217;"
	texEnDash             = "&#8211;"
	texEmDash             = "&#8212;"

	texOpenQFlag  = "<!--oq-->"
	texOpenSQFlag = "<!--osq-->"
	texAposFlag   = "<!--apos-->"
	texPrimeFlag  = "<!--wp-prime-or-quote-->"
)

// texStatic 是 $static_characters 与 $static_replacements（含 cockney 列表）。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var texStatic = []string{
	"...", "&#8230;",
	"``", texOpeningQuote,
	"''", texClosingQuote,
	" (tm)", " &#8482;",
	"'tain't", "&#8217;tain&#8217;t",
	"'twere", "&#8217;twere",
	"'twas", "&#8217;twas",
	"'tis", "&#8217;tis",
	"'twill", "&#8217;twill",
	"'til", "&#8217;til",
	"'bout", "&#8217;bout",
	"'nuff", "&#8217;nuff",
	"'round", "&#8217;round",
	"'cause", "&#8217;cause",
	"'em", "&#8217;em",
}

type reReplacement struct {
	re   *regexp2.Regexp
	repl string
}

// texTail 是若干模式中"后面紧跟结尾、标点或空白"的公共部分。
const texTail = `(?=\Z|[.,:;!?)}\-\]]|&gt;|` + reSpaces + `)`

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	texDynamicApos = []reReplacement{
		{mustRE(`'([0-9][0-9])'`+texTail, 0), texAposFlag + "$1" + texClosingSingleQuote},
		{mustRE(`'([0-9][0-9])"`+texTail, 0), texAposFlag + "$1" + texClosingQuote},
		{mustRE(`'(?=[0-9][0-9](?:\Z|(?![%0-9]|[.,][0-9])))`, 0), texAposFlag},
		{mustRE(`(?<=\A|`+reSpaces+`)'([0-9][.,0-9]*)'`, 0), texOpenSQFlag + "$1" + texClosingSingleQuote},
		{mustRE(`(?<=\A|[(\[{"\-]|&lt;|`+reSpaces+`)'`, 0), texOpenSQFlag},
		{mustRE(`(?<!`+reSpaces+`)'(?!\Z|[.,:;!?"'(){}\[\]\-]|&[lg]t;|`+reSpaces+`)`, 0), texAposFlag},
	}
	texDynamicQuote = []reReplacement{
		{mustRE(`(?<=\A|`+reSpaces+`)"([0-9][.,0-9]*)"`, 0), texOpenQFlag + "$1" + texClosingQuote},
		{mustRE(`(?<=\A|[(\[{\-]|&lt;|`+reSpaces+`)"(?!`+reSpaces+`)`, 0), texOpenQFlag},
	}
	texDynamicDash = []reReplacement{
		{mustRE(`---`, 0), texEmDash},
		{mustRE(`(?<=^|`+reSpaces+`)--(?=$|`+reSpaces+`)`, 0), texEmDash},
		{mustRE(`(?<!xn)--`, 0), texEnDash},
		{mustRE(`(?<=^|`+reSpaces+`)-(?=$|`+reSpaces+`)`, 0), texEnDash},
	}

	// 9x9 (times)，但不处理 0x9999。PHP 原式使用了 \b 与条件分组，这里改写为等价的环视。
	texTimesCheck = mustRE(`(?<=[0-9])x[0-9]`, 0)
	texTimes      = mustRE(
		`(?<!`+reWord+`)([0-9](?:(?<=0)[0-9.,]+|(?<!0)[0-9.,]*))x([0-9][0-9.,]*)`+
			`(?:(?<=`+reWord+`)(?!`+reWord+`)|(?<!`+reWord+`)(?=`+reWord+`))`,
		0,
	)

	// 把不像实体的 & 替换成 &#038;。
	texAmp = mustRE(`&(?!#(?:[0-9]+|x[a-f0-9]+);|[a-z1-4]{1,8};)`, regexp2.IgnoreCase)

	// 查找内容中出现的短代码名。
	shortcodeNameFinder = mustRE(`\[/?([^<>&/\[\]\x00-\x20=]+)`, 0)

	// _get_wptexturize_split_regex() 中的 HTML 部分。
	texHTMLRegex = `<(?:(?=!--)!(?>(?:-(?!->)(?>[^\-]*))*)(?:-->)?|(?!!--)[^>]*>?)`
	texSplitHTML = mustRE(`(`+texHTMLRegex+`)`, 0)

	texNoTags = []string{"pre", "code", "kbd", "style", "script", "tt"}
	// no_texturize_shortcodes 默认只有 code，而 code 不是已注册的短代码。
	texNoShortcodes = []string{"code"}

	texSplitCache compileCache[*texShortcodeRegexes]
)

type texShortcodeRegexes struct {
	split *regexp2.Regexp
	full  *regexp2.Regexp
}

// texShortcodeRegexFor 返回包含给定短代码名的分割正则与完整匹配正则。
func texShortcodeRegexFor(tagnames []string) *texShortcodeRegexes {
	return texSplitCache.get(strings.Join(tagnames, "|"), func() *texShortcodeRegexes {
		quoted := make([]string, len(tagnames))
		for i, t := range tagnames {
			quoted[i] = regexp2.Escape(t)
		}
		// _get_wptexturize_shortcode_regex()
		sc := `\[[/\[]?(?:` + strings.Join(quoted, "|") + `)(?=[\t\n\v\f\r \]/])` +
			`(?>(?:[^\[\]<>]+|<[^\[\]>]*>)*)\]\]?`
		return &texShortcodeRegexes{
			split: mustRE(`(`+texHTMLRegex+`|`+sc+`)`, 0),
			full:  mustRE(`^(?:`+sc+`)$`, 0),
		}
	})
}

// foundShortcodeNames 对应 array_intersect(array_keys($shortcode_tags), $matches[1])。
func foundShortcodeNames(text string, registered []string) []string {
	if !strings.Contains(text, "[") {
		return nil
	}
	seen := map[string]bool{}
	for _, m := range findAllMatches(shortcodeNameFinder, text, -1) {
		seen[m.group(1)] = true
	}
	var names []string
	for _, name := range registered {
		if seen[name] {
			names = append(names, name)
		}
	}
	return names
}

// wptexturize 移植自 wptexturize()。
func wptexturize(text string) string {
	if text == "" || text == "0" {
		return text
	}

	tagnames := foundShortcodeNames(text, registeredShortcodes)
	var scRegexes *texShortcodeRegexes
	split := texSplitHTML
	if len(tagnames) > 0 {
		scRegexes = texShortcodeRegexFor(tagnames)
		split = scRegexes.split
	}

	var tagsStack, shortcodesStack []string
	parts := splitDelim(split, text, true)
	for i, curl := range parts {
		switch {
		case curl[0] == '<':
			if strings.HasPrefix(curl, "<!--") {
				continue
			}
			curl = replaceAll(texAmp, curl, "&#038;")
			parts[i] = curl
			tagsStack = texPushPopElement(curl, tagsStack, texNoTags)
		case phpTrim(curl) == "":
			continue
		case curl[0] == '[' && scRegexes != nil && isMatch(scRegexes.full, curl):
			if !strings.HasPrefix(curl, "[[") && !strings.HasSuffix(curl, "]]") {
				shortcodesStack = texPushPopElement(curl, shortcodesStack, texNoShortcodes)
			}
		case len(shortcodesStack) == 0 && len(tagsStack) == 0:
			parts[i] = texturizeText(curl)
		}
	}
	return strings.Join(parts, "")
}

// texturizeText 对不在禁用元素中的纯文本片段做替换。
func texturizeText(curl string) string {
	curl = strReplacePairs(curl, texStatic...)

	if strings.Contains(curl, "'") {
		for _, r := range texDynamicApos {
			curl = replaceAll(r.re, curl, r.repl)
		}
		curl = wptexturizePrimes(curl, "'", texPrime, texOpenSQFlag, texClosingSingleQuote)
		curl = strings.ReplaceAll(curl, texAposFlag, texApos)
		curl = strings.ReplaceAll(curl, texOpenSQFlag, texOpeningSingleQuote)
	}
	if strings.Contains(curl, `"`) {
		for _, r := range texDynamicQuote {
			curl = replaceAll(r.re, curl, r.repl)
		}
		curl = wptexturizePrimes(curl, `"`, texDoublePrime, texOpenQFlag, texClosingQuote)
		curl = strings.ReplaceAll(curl, texOpenQFlag, texOpeningQuote)
	}
	if strings.Contains(curl, "-") {
		for _, r := range texDynamicDash {
			curl = replaceAll(r.re, curl, r.repl)
		}
	}

	if isMatch(texTimesCheck, curl) {
		curl = replaceAll(texTimes, curl, "$1&#215;$2")
	}

	return replaceAll(texAmp, curl, "&#038;")
}

// texPrimeRegexes 是 wptexturize_primes() 针对某个引号字符使用的正则。
type texPrimeRegexes struct {
	quote, prime, flagAfterDigit, flagNoDigit *regexp2.Regexp
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var texPrimes = map[string]*texPrimeRegexes{
	"'": newTexPrimeRegexes("'"),
	`"`: newTexPrimeRegexes(`"`),
}

func newTexPrimeRegexes(needle string) *texPrimeRegexes {
	return &texPrimeRegexes{
		quote:          mustRE(needle+texTail, 0),
		prime:          mustRE(`(?<=[0-9])`+needle, 0),
		flagAfterDigit: mustRE(`(?<=[0-9])`+texPrimeFlag, 0),
		flagNoDigit:    mustRE(`(?<![0-9])`+texPrimeFlag, 0),
	}
}

// wptexturizePrimes 移植自 wptexturize_primes()。
func wptexturizePrimes(haystack, needle, prime, openQuote, closeQuote string) string {
	res := texPrimes[needle]
	sentences := strings.Split(haystack, openQuote)
	for key, sentence := range sentences {
		switch {
		case !strings.Contains(sentence, needle):
			continue
		case key != 0 && !strings.Contains(sentence, closeQuote):
			var count int
			sentence, count = replaceCount(res.quote, sentence, texPrimeFlag)
			switch {
			case count > 1:
				// 这个句子看起来有多个闭合引号，尝试 Vulcan 逻辑。
				var count2 int
				sentence, count2 = replaceCount(res.flagNoDigit, sentence, closeQuote)
				if count2 == 0 {
					// 优先认为最右侧的"引号+句号"是引文结束。
					pos := strings.LastIndex(sentence, texPrimeFlag+".")
					if pos < 0 {
						pos = strings.LastIndex(sentence, texPrimeFlag)
					}
					sentence = sentence[:pos] + closeQuote + sentence[pos+len(texPrimeFlag):]
				}
				sentence = replaceAll(res.prime, sentence, prime)
				sentence = replaceAll(res.flagAfterDigit, sentence, prime)
				sentence = strings.ReplaceAll(sentence, texPrimeFlag, closeQuote)
			case count == 1:
				sentence = strings.ReplaceAll(sentence, texPrimeFlag, closeQuote)
				sentence = replaceAll(res.prime, sentence, prime)
			default:
				sentence = replaceAll(res.prime, sentence, prime)
			}
		default:
			sentence = replaceAll(res.prime, sentence, prime)
			sentence = replaceAll(res.quote, sentence, closeQuote)
		}
		if needle == `"` && strings.Contains(sentence, `"`) {
			sentence = strings.ReplaceAll(sentence, `"`, closeQuote)
		}
		sentences[key] = sentence
	}
	return strings.Join(sentences, openQuote)
}

// texPushPopElement 移植自 _wptexturize_pushpop_element()。
func texPushPopElement(text string, stack, disabled []string) []string {
	var opening bool
	var nameOffset int
	switch {
	case len(text) > 1 && text[1] != '/':
		opening, nameOffset = true, 1
	case len(stack) == 0:
		return stack
	default:
		opening, nameOffset = false, 2
	}

	// 解析出标签名：没有空格时 $space 为 -1，即去掉最后一个字符（> 或 ]）。
	space := strings.IndexByte(text, ' ')
	if space < 0 {
		space = -1
	} else {
		space -= nameOffset
	}
	tag := phpSubstr(text, nameOffset, space)

	for _, d := range disabled {
		if d != tag {
			continue
		}
		if opening {
			return append(stack, tag)
		}
		if len(stack) > 0 && stack[len(stack)-1] == tag {
			return stack[:len(stack)-1]
		}
		break
	}
	return stack
}
