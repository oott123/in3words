package wpfmt

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
)

// PHP 的 PCRE 在没有 /u 修饰符时按字节匹配，\s、\d、\w、\b 只认 ASCII 字符；
// regexp2 按 rune 匹配，并且这些字符类是 Unicode 语义（例如会把全角空格当成 \s、
// 把汉字当成 \w）。为了与 WordPress 的行为一致，所有模式都显式写出 ASCII 字符类，
// 不使用 \s、\d、\w、\b。
const (
	// reSpace 对应 PHP 的 \s。
	reSpace = `[\t\n\v\f\r ]`
	// reWord 对应 PHP 的 \w。
	reWord = `[A-Za-z0-9_]`
	// reSpaces 对应 wp_spaces_regexp() 的默认值。
	reSpaces = `(?:[\r\n\t ]|\u00A0|&nbsp;)`
)

// mustRE 编译一个 regexp2 正则，用于包级变量初始化。
func mustRE(pattern string, opts regexp2.RegexOptions) *regexp2.Regexp {
	return regexp2.MustCompile(pattern, opts)
}

// reMatch 是一次匹配的结果，所有位置都是输入字符串的字节偏移。
type reMatch struct {
	src    string
	groups [][2]int // 未参与匹配的分组为 {-1, -1}
	last   int      // 最后一个参与匹配的分组编号，用于模拟 PHP 省略末尾未匹配分组的行为
}

func (m *reMatch) start() int { return m.groups[0][0] }
func (m *reMatch) end() int   { return m.groups[0][1] }

// group 返回第 i 个分组的内容；未匹配时返回空字符串（与 PHP 一致）。
func (m *reMatch) group(i int) string {
	if i >= len(m.groups) || m.groups[i][0] < 0 {
		return ""
	}
	return m.src[m.groups[i][0]:m.groups[i][1]]
}

// matched 判断第 i 个分组是否参与了匹配。
func (m *reMatch) matched(i int) bool {
	return i < len(m.groups) && m.groups[i][0] >= 0
}

// isset 模拟 PHP 中 isset($matches[$i])：PHP 会省略末尾未参与匹配的分组，
// 但中间未参与匹配的分组会以空字符串出现。
func (m *reMatch) isset(i int) bool {
	return i <= m.last
}

// runeOffsets 返回每个 rune 下标对应的字节偏移，末尾额外附加 len(s)。
// 对非法 UTF-8 字节，range 与 []rune 转换都会把单个字节当成一个 rune，二者一致。
func runeOffsets(s string) []int {
	offsets := make([]int, 0, utf8.RuneCountInString(s)+1)
	for i := range s {
		offsets = append(offsets, i)
	}
	return append(offsets, len(s))
}

// findAllMatches 返回 re 在 s 中所有不重叠的匹配；limit < 0 表示不限制数量。
func findAllMatches(re *regexp2.Regexp, s string, limit int) []*reMatch {
	if limit == 0 {
		return nil
	}
	runes := []rune(s)
	var offsets []int
	var result []*reMatch
	// regexp2 只会在超时时返回错误，这里没有设置超时，因此忽略错误。
	m, _ := re.FindRunesMatch(runes)
	for m != nil && (limit < 0 || len(result) < limit) {
		if offsets == nil {
			offsets = runeOffsets(s)
		}
		groups := m.Groups()
		rm := &reMatch{src: s, groups: make([][2]int, len(groups))}
		for i := range groups {
			g := &groups[i]
			if len(g.Captures) == 0 {
				rm.groups[i] = [2]int{-1, -1}
				continue
			}
			rm.groups[i] = [2]int{offsets[g.Index], offsets[g.Index+g.Length]}
			rm.last = i
		}
		result = append(result, rm)
		m, _ = re.FindNextMatch(m)
	}
	return result
}

// findMatch 返回第一个匹配，没有匹配时返回 nil。
func findMatch(re *regexp2.Regexp, s string) *reMatch {
	ms := findAllMatches(re, s, 1)
	if len(ms) == 0 {
		return nil
	}
	return ms[0]
}

// isMatch 对应 1 === preg_match(...)。
func isMatch(re *regexp2.Regexp, s string) bool {
	ok, _ := re.MatchString(s)
	return ok
}

// replaceFunc 对应 preg_replace_callback，返回替换结果与替换次数。
func replaceFunc(re *regexp2.Regexp, s string, fn func(m *reMatch) string) (string, int) {
	ms := findAllMatches(re, s, -1)
	if len(ms) == 0 {
		return s, 0
	}
	var b strings.Builder
	b.Grow(len(s))
	prev := 0
	for _, m := range ms {
		b.WriteString(s[prev:m.start()])
		b.WriteString(fn(m))
		prev = m.end()
	}
	b.WriteString(s[prev:])
	return b.String(), len(ms)
}

// replaceAll 对应 preg_replace，模板中的 $n / ${n} / \n 表示分组引用。
func replaceAll(re *regexp2.Regexp, s, template string) string {
	out, _ := replaceCount(re, s, template)
	return out
}

// replaceCount 与 replaceAll 相同，但额外返回替换次数。
func replaceCount(re *regexp2.Regexp, s, template string) (string, int) {
	return replaceFunc(re, s, func(m *reMatch) string { return expandTemplate(template, m) })
}

// expandTemplate 按 PHP preg_replace 的规则展开替换模板。
//
//nolint:gocognit,mnd // 逐行移植 WordPress 实现，保持原有结构便于对照；正则分组编号与阈值沿用 WordPress 原实现
func expandTemplate(template string, m *reMatch) string {
	if !strings.ContainsAny(template, `$\`) {
		return template
	}
	var b strings.Builder
	for i := 0; i < len(template); i++ {
		c := template[i]
		if (c == '$' || c == '\\') && i+1 < len(template) {
			j := i + 1
			braced := c == '$' && template[j] == '{'
			if braced {
				j++
			}
			n, k := 0, j
			for k < len(template) && k-j < 2 && template[k] >= '0' && template[k] <= '9' {
				n = n*10 + int(template[k]-'0')
				k++
			}
			if k > j && (!braced || (k < len(template) && template[k] == '}')) {
				if braced {
					k++
				}
				b.WriteString(m.group(n))
				i = k - 1
				continue
			}
		}
		b.WriteByte(c)
	}
	return b.String()
}

// splitDelim 对应 preg_split($re, $s, -1, PREG_SPLIT_DELIM_CAPTURE [| PREG_SPLIT_NO_EMPTY])，
// 仅支持模式中只有一个捕获分组的情况。
func splitDelim(re *regexp2.Regexp, s string, noEmpty bool) []string {
	ms := findAllMatches(re, s, -1)
	parts := make([]string, 0, len(ms)*2+1)
	add := func(p string) {
		if noEmpty && p == "" {
			return
		}
		parts = append(parts, p)
	}
	prev := 0
	for _, m := range ms {
		add(s[prev:m.start()])
		if m.matched(1) {
			add(m.group(1))
		}
		prev = m.end()
	}
	add(s[prev:])
	return parts
}

// phpTrim 对应 PHP trim() 的默认字符集。
func phpTrim(s string) string {
	return strings.Trim(s, " \t\n\r\x00\x0B")
}

// phpEmpty 对应 PHP 中字符串的 empty() 判断。
func phpEmpty(s string) bool {
	return s == "" || s == "0"
}

// strReplacePairs 对应 str_replace(array, array, $s)：按顺序逐对替换。
func strReplacePairs(s string, pairs ...string) string {
	for i := 0; i+1 < len(pairs); i += 2 {
		if strings.Contains(s, pairs[i]) {
			s = strings.ReplaceAll(s, pairs[i], pairs[i+1])
		}
	}
	return s
}

// phpSubstr 对应 PHP 的 substr($s, $offset, $length)（offset 非负）。
func phpSubstr(s string, offset, length int) string {
	if offset >= len(s) {
		return ""
	}
	s = s[offset:]
	switch {
	case length >= 0:
		return s[:min(length, len(s))]
	case -length >= len(s):
		return ""
	default:
		return s[:len(s)+length]
	}
}

// compileCache 按 key 缓存根据内容动态构造的正则。
type compileCache[T any] struct {
	mu sync.Mutex
	m  map[string]T
}

// get 返回 key 对应的值，不存在时调用 build 构造并缓存。
func (c *compileCache[T]) get(key string, build func() T) T {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.m[key]; ok {
		return v
	}
	if c.m == nil {
		c.m = map[string]T{}
	}
	v := build()
	c.m[key] = v
	return v
}
