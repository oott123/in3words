package wpfmt

import (
	"html"
	"slices"
	"sort"
	"strings"
)

// tagProcessor 是 WP_HTML_Tag_Processor 的极简移植，只覆盖本包用到的
// next_tag / get_attribute / set_attribute / add_class / get_updated_html。
type tagProcessor struct {
	html string
	pos  int // 下一次扫描的起始位置

	cur      *tagToken
	classAdd []string
	edits    []textEdit
}

type tagToken struct {
	name    string // 小写标签名
	closer  bool
	nameEnd int // 标签名结束的位置
	attrs   []tagAttr
}

type tagAttr struct {
	name       string // 小写属性名
	start, end int    // 整个属性（含值）在 HTML 中的范围
	value      string // 原始值（未解码）
	hasValue   bool
}

type textEdit struct {
	start, end int
	text       string
	seq        int
}

func newTagProcessor(s string) *tagProcessor {
	return &tagProcessor{html: s}
}

// rawTextElements 中的元素内容不会被当作标签解析。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var rawTextElements = map[string]bool{
	"script": true, "style": true, "textarea": true, "title": true,
	"xmp": true, "iframe": true, "noembed": true, "noframes": true,
}

func isASCIIAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isHTMLSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\f' || c == '\r'
}

// nextToken 移动到下一个标签（开始或结束标签），没有更多标签时返回 false。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func (p *tagProcessor) nextToken() bool {
	p.flushClassUpdates()
	s := p.html
	for p.pos < len(s) {
		lt := strings.IndexByte(s[p.pos:], '<')
		if lt < 0 {
			p.pos = len(s)
			break
		}
		at := p.pos + lt
		if at+1 >= len(s) {
			p.pos = len(s)
			break
		}
		c := s[at+1]
		switch {
		case strings.HasPrefix(s[at:], "<!--"):
			end := strings.Index(s[at+4:], "-->")
			if end < 0 {
				p.pos = len(s)
			} else {
				p.pos = at + 4 + end + 3
			}
		case c == '!' || c == '?':
			p.skipPast(at, '>')
		case c == '/' && at+2 < len(s) && isASCIIAlpha(s[at+2]):
			p.cur = p.parseTag(at+2, true)
			return true
		case isASCIIAlpha(c):
			p.cur = p.parseTag(at+1, false)
			if rawTextElements[p.cur.name] {
				p.skipRawText(p.cur.name)
			}
			return true
		default:
			p.pos = at + 1
		}
	}
	p.cur = nil
	return false
}

func (p *tagProcessor) skipPast(from int, c byte) {
	end := strings.IndexByte(p.html[from:], c)
	if end < 0 {
		p.pos = len(p.html)
		return
	}
	p.pos = from + end + 1
}

func (p *tagProcessor) skipRawText(name string) {
	lower := strings.ToLower(p.html[p.pos:])
	end := strings.Index(lower, "</"+name)
	if end < 0 {
		return
	}
	p.pos += end
}

// parseTag 从标签名开始解析一个标签，结束后 p.pos 指向 > 之后。
//
//nolint:gocognit,nestif // 逐行移植 WordPress 实现，保持原有结构便于对照
func (p *tagProcessor) parseTag(nameStart int, closer bool) *tagToken {
	s := p.html
	i := nameStart
	for i < len(s) && !isHTMLSpace(s[i]) && s[i] != '/' && s[i] != '>' {
		i++
	}
	tok := &tagToken{name: strings.ToLower(s[nameStart:i]), closer: closer, nameEnd: i}
	for i < len(s) {
		for i < len(s) && (isHTMLSpace(s[i]) || s[i] == '/') {
			i++
		}
		if i >= len(s) || s[i] == '>' {
			break
		}
		attrStart := i
		i++ // 属性名的第一个字符可以是 =
		for i < len(s) && !isHTMLSpace(s[i]) && s[i] != '/' && s[i] != '>' && s[i] != '=' {
			i++
		}
		attr := tagAttr{name: strings.ToLower(s[attrStart:i]), start: attrStart}
		j := i
		for j < len(s) && isHTMLSpace(s[j]) {
			j++
		}
		if j < len(s) && s[j] == '=' {
			j++
			for j < len(s) && isHTMLSpace(s[j]) {
				j++
			}
			attr.hasValue = true
			if j < len(s) && (s[j] == '"' || s[j] == '\'') {
				q := s[j]
				end := strings.IndexByte(s[j+1:], q)
				if end < 0 {
					attr.value = s[j+1:]
					j = len(s)
				} else {
					attr.value = s[j+1 : j+1+end]
					j = j + 1 + end + 1
				}
			} else {
				k := j
				for k < len(s) && !isHTMLSpace(s[k]) && s[k] != '>' {
					k++
				}
				attr.value = s[j:k]
				j = k
			}
			i = j
		}
		attr.end = i
		tok.attrs = append(tok.attrs, attr)
	}
	if i < len(s) {
		i++
	}
	p.pos = i
	return tok
}

// nextTag 移动到下一个名为 name（为空时为任意）的开始标签。
func (p *tagProcessor) nextTag(name string) bool {
	for p.nextToken() {
		if !p.cur.closer && (name == "" || strings.EqualFold(p.cur.name, name)) {
			return true
		}
	}
	return false
}

func (p *tagProcessor) attr(name string) *tagAttr {
	if p.cur == nil || p.cur.closer {
		return nil
	}
	name = strings.ToLower(name)
	for i := range p.cur.attrs {
		if p.cur.attrs[i].name == name {
			return &p.cur.attrs[i]
		}
	}
	return nil
}

// getAttribute 返回解码后的属性值；布尔属性返回空字符串与 true。
func (p *tagProcessor) getAttribute(name string) (string, bool) {
	a := p.attr(name)
	if a == nil {
		return "", false
	}
	return html.UnescapeString(a.value), true
}

// uriAttributes 对应 wp_kses_uri_attributes()。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var uriAttributes = map[string]bool{
	"action": true, "archive": true, "background": true, "cite": true, "classid": true,
	"codebase": true, "data": true, "formaction": true, "href": true, "icon": true,
	"longdesc": true, "manifest": true, "poster": true, "profile": true, "src": true,
	"usemap": true, "xmlns": true,
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var attrValueEscaper = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;", `"`, "&quot;", "'", "&apos;")

// setAttribute 设置当前标签的属性。
func (p *tagProcessor) setAttribute(name, value string) {
	if p.cur == nil || p.cur.closer {
		return
	}
	var escaped string
	if uriAttributes[strings.ToLower(name)] {
		escaped = escURL(value)
		if escaped == "" && value != "" {
			return
		}
	} else {
		escaped = attrValueEscaper.Replace(value)
	}
	updated := name + `="` + escaped + `"`
	if a := p.attr(name); a != nil {
		p.addEdit(a.start, a.end, updated)
	} else {
		p.addEdit(p.cur.nameEnd, p.cur.nameEnd, " "+updated)
	}
	if strings.EqualFold(name, "class") {
		p.classAdd = nil
	}
}

func (p *tagProcessor) addEdit(start, end int, text string) {
	// 同一属性的多次修改只保留最后一次。
	for i := range p.edits {
		if p.edits[i].start == start && p.edits[i].end == end && start != end {
			p.edits[i].text = text
			return
		}
	}
	p.edits = append(p.edits, textEdit{start: start, end: end, text: text, seq: len(p.edits)})
}

// addClass 给当前标签添加一个类名。
func (p *tagProcessor) addClass(name string) {
	if p.cur == nil || p.cur.closer {
		return
	}
	p.classAdd = append(p.classAdd, name)
}

// flushClassUpdates 对应 class_name_updates_to_attributes_updates()。
//
//nolint:gocognit // 逐行移植 WordPress 实现，保持原有结构便于对照
func (p *tagProcessor) flushClassUpdates() {
	if len(p.classAdd) == 0 || p.cur == nil {
		return
	}
	adds := p.classAdd
	p.classAdd = nil

	existing, _ := p.getAttribute("class")
	var b strings.Builder
	var seen []string
	at := 0
	for at < len(existing) {
		wsAt := at
		for at < len(existing) && strings.IndexByte(" \t\f\r\n", existing[at]) >= 0 {
			at++
		}
		nameStart := at
		for at < len(existing) && strings.IndexByte(" \t\f\r\n", existing[at]) < 0 {
			at++
		}
		if at == nameStart {
			break
		}
		name := existing[nameStart:at]
		if containsString(seen, name) {
			continue
		}
		seen = append(seen, name)
		if b.Len() > 0 {
			b.WriteString(existing[wsAt:nameStart])
		}
		b.WriteString(name)
	}
	modified := false
	for _, name := range adds {
		if containsString(seen, name) {
			continue
		}
		seen = append(seen, name)
		modified = true
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(name)
	}
	if modified {
		p.setAttribute("class", b.String())
	}
}

func containsString(list []string, s string) bool {
	return slices.Contains(list, s)
}

// updatedHTML 返回应用了所有修改后的 HTML。
func (p *tagProcessor) updatedHTML() string {
	p.flushClassUpdates()
	if len(p.edits) == 0 {
		return p.html
	}
	edits := append([]textEdit(nil), p.edits...)
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start < edits[j].start
		}
		return edits[i].seq < edits[j].seq
	})
	var b strings.Builder
	prev := 0
	for _, e := range edits {
		if e.start < prev {
			continue
		}
		b.WriteString(p.html[prev:e.start])
		b.WriteString(e.text)
		prev = e.end
	}
	b.WriteString(p.html[prev:])
	return b.String()
}

// hasClass 判断当前标签是否带有指定类名。
func (p *tagProcessor) hasClass(name string) bool {
	class, ok := p.getAttribute("class")
	if !ok {
		return false
	}
	return slices.Contains(strings.Fields(class), name)
}

// removeAttribute 删除当前标签的属性（只删除属性本身，保留前面的空白）。
func (p *tagProcessor) removeAttribute(name string) {
	if a := p.attr(name); a != nil {
		p.addEdit(a.start, a.end, "")
	}
}
