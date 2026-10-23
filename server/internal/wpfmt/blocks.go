package wpfmt

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/dlclark/regexp2"
)

// parsedBlock 对应 WP_Block_Parser_Block。
type parsedBlock struct {
	name         string // 空字符串表示自由格式（经典）内容
	attrs        map[string]any
	innerBlocks  []*parsedBlock
	innerHTML    string
	innerContent []*string // nil 元素表示内部区块的位置
}

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var blockTokenRegex = mustRE(
	// 分组：1 结束标记、2 命名空间、3 名称、4 属性 JSON、5 自闭合标记。
	`<!--`+reSpace+`+(/)?wp:([a-z][a-z0-9_-]*/)?([a-z][a-z0-9_-]*)`+reSpace+`+`+
		`(\{(?>(?:[^}]+|\}+(?=\})|(?!\}`+reSpace+`+/?-->).)*)?\}`+reSpace+`+)?(/)?-->`,
	regexp2.Singleline,
)

type blockParserFrame struct {
	block            *parsedBlock
	tokenStart       int
	tokenLength      int
	prevOffset       int
	leadingHTMLStart int // -1 表示没有
}

// blockParser 移植自 WP_Block_Parser。
type blockParser struct {
	document string
	offset   int
	output   []*parsedBlock
	stack    []*blockParserFrame
	tokens   []*reMatch
	next     int
}

// parseBlocks 对应 parse_blocks()。
//
//go:fix inline
func parseBlocks(document string) []*parsedBlock {
	p := &blockParser{document: document, tokens: findAllMatches(blockTokenRegex, document, -1)}
	for p.proceed() {
	}
	return p.output
}

func freeformBlock(html string) *parsedBlock {
	return &parsedBlock{innerHTML: html, innerContent: []*string{new(html)}}
}

//nolint:funlen,gocognit,mnd // 逐行移植 WordPress 实现，保持原有结构便于对照；正则分组编号与阈值沿用 WordPress 原实现
func (p *blockParser) proceed() bool {
	if p.next >= len(p.tokens) {
		// no-more-tokens
		switch len(p.stack) {
		case 0:
			p.addFreeform()
		case 1:
			p.addBlockFromStack(-1)
		default:
			for len(p.stack) > 0 {
				p.addBlockFromStack(-1)
			}
		}
		return false
	}

	tok := p.tokens[p.next]
	p.next++
	startOffset, tokenLength := tok.start(), tok.end()-tok.start()
	leadingHTMLStart := -1
	if startOffset > p.offset {
		leadingHTMLStart = p.offset
	}

	namespace := tok.group(2)
	if !tok.matched(2) {
		namespace = "core/"
	}
	name := namespace + tok.group(3)
	isCloser := tok.matched(1)
	isVoid := tok.matched(5)
	var attrs map[string]any
	if tok.matched(4) {
		// json_decode 失败时为 null。
		if err := json.Unmarshal([]byte(strings.TrimRight(tok.group(4), " \t\n\v\f\r")), &attrs); err != nil {
			attrs = nil
		}
	} else {
		attrs = map[string]any{}
	}

	switch {
	case isVoid:
		block := &parsedBlock{name: name, attrs: attrs}
		if len(p.stack) == 0 {
			if leadingHTMLStart >= 0 {
				p.output = append(p.output, freeformBlock(p.document[leadingHTMLStart:startOffset]))
			}
			p.output = append(p.output, block)
		} else {
			p.addInnerBlock(block, startOffset, tokenLength, -1)
		}
		p.offset = startOffset + tokenLength
		return true

	case isCloser:
		switch len(p.stack) {
		case 0:
			p.addFreeform()
			return false
		case 1:
			p.addBlockFromStack(startOffset)
			p.offset = startOffset + tokenLength
			return true
		}
		top := p.stack[len(p.stack)-1]
		p.stack = p.stack[:len(p.stack)-1]
		html := p.document[top.prevOffset:startOffset]
		top.block.innerHTML += html
		top.block.innerContent = append(top.block.innerContent, new(html))
		top.prevOffset = startOffset + tokenLength
		p.addInnerBlock(top.block, top.tokenStart, top.tokenLength, startOffset+tokenLength)
		p.offset = startOffset + tokenLength
		return true

	default:
		p.stack = append(p.stack, &blockParserFrame{
			block:            &parsedBlock{name: name, attrs: attrs},
			tokenStart:       startOffset,
			tokenLength:      tokenLength,
			prevOffset:       startOffset + tokenLength,
			leadingHTMLStart: leadingHTMLStart,
		})
		p.offset = startOffset + tokenLength
		return true
	}
}

func (p *blockParser) addFreeform() {
	if p.offset >= len(p.document) {
		return
	}
	p.output = append(p.output, freeformBlock(p.document[p.offset:]))
}

func (p *blockParser) addInnerBlock(block *parsedBlock, tokenStart, tokenLength, lastOffset int) {
	parent := p.stack[len(p.stack)-1]
	parent.block.innerBlocks = append(parent.block.innerBlocks, block)
	html := p.document[parent.prevOffset:tokenStart]
	if !phpEmpty(html) {
		parent.block.innerHTML += html
		parent.block.innerContent = append(parent.block.innerContent, new(html))
	}
	parent.block.innerContent = append(parent.block.innerContent, nil)
	if lastOffset > 0 {
		parent.prevOffset = lastOffset
	} else {
		parent.prevOffset = tokenStart + tokenLength
	}
}

func (p *blockParser) addBlockFromStack(endOffset int) {
	top := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	var html string
	if endOffset >= 0 {
		html = p.document[top.prevOffset:endOffset]
	} else {
		html = p.document[top.prevOffset:]
	}
	if !phpEmpty(html) {
		top.block.innerHTML += html
		top.block.innerContent = append(top.block.innerContent, new(html))
	}
	if top.leadingHTMLStart >= 0 {
		p.output = append(p.output, freeformBlock(p.document[top.leadingHTMLStart:top.tokenStart]))
	}
	p.output = append(p.output, top.block)
}

// doBlocks 移植自 do_blocks()。
//
// 与 WP 的差异：
//   - 需要服务端渲染的动态区块（如 latest-posts、archives 等）无法访问 WordPress 运行时，
//     这里按静态区块处理，直接输出保存在文章中的 HTML（多数为空）；
//   - 只实现了 core/paragraph、core/heading、core/list 的类名补充，core/image 的空 figcaption
//     清理，core/shortcode 的 wpautop，以及布局支持（layout）添加的类名；
//   - 布局支持不会生成 wp-container-* 类名：这个类名的哈希依赖主题的 theme.json，
//     配套的 CSS 也只会输出在主题页面中，对 REST 消费方没有意义；
//   - 其他依赖区块属性的 block supports（typography、elements、duotone 等）未实现。
func (st *renderState) doBlocks(content string) string {
	blocks := parseBlocks(content)
	var b strings.Builder
	for _, block := range blocks {
		b.WriteString(st.renderBlock(block))
	}
	return b.String()
}

// HasBlocks 对应 has_blocks($content)：内容中是否含有区块分隔注释。
// 调用方可据此判断 WordPress 是否会进入 Options.AutopAfterUnautop 描述的状态。
func HasBlocks(content string) bool {
	return strings.Contains(content, "<!-- wp:")
}

// renderBlock 对应 render_block() / WP_Block::render()。
func (st *renderState) renderBlock(block *parsedBlock) string {
	var b strings.Builder
	index := 0
	for _, chunk := range block.innerContent {
		if chunk != nil {
			b.WriteString(*chunk)
			continue
		}
		if index < len(block.innerBlocks) {
			b.WriteString(st.renderBlock(block.innerBlocks[index]))
		}
		index++
	}
	content := b.String()
	if block.name == "" {
		return content
	}

	// 动态区块的 render_callback。
	switch block.name {
	case "core/heading":
		content = addClassToFirst(content, "wp-block-heading", "h1", "h2", "h3", "h4", "h5", "h6")
	case "core/list":
		content = addClassToFirst(content, "wp-block-list", "ol", "ul")
	case "core/shortcode":
		content = wpautop(content)
	case "core/image":
		content = renderCoreImage(block.attrs, content)
	}

	// render_block 过滤器。
	content = renderLayoutSupport(block, content)

	// render_block_{$name} 过滤器。
	if block.name == "core/paragraph" {
		content = addClassToFirst(content, "wp-block-paragraph", "p")
	}
	return content
}

// addClassToFirst 给第一个匹配的标签添加类名；内容为空时原样返回。
func addClassToFirst(content, class string, tags ...string) string {
	if content == "" {
		return content
	}
	p := newTagProcessor(content)
	for p.nextTag("") {
		if containsString(tags, p.cur.name) {
			p.addClass(class)
			break
		}
	}
	return p.updatedHTML()
}

// renderCoreImage 是 render_block_core_image() 的简化版本：没有带 src 的图片时输出空字符串，
// 并移除空的 figcaption。灯箱等交互特性未实现。
func renderCoreImage(attrs map[string]any, content string) string {
	if !strings.Contains(strings.ToLower(content), "<img") {
		return ""
	}
	p := newTagProcessor(content)
	if !p.nextTag("img") {
		return ""
	}
	if src, _ := p.getAttribute("src"); src == "" {
		return ""
	}
	if caption, ok := attrs["caption"].(string); ok && caption != "" {
		return content
	}
	// 移除空的 figcaption（中间允许有注释）。
	if !p.nextTag("figcaption") {
		return content
	}
	openStart := strings.LastIndex(content[:p.pos], "<figcaption")
	rest := content[p.pos:]
	for strings.HasPrefix(rest, "<!--") {
		end := strings.Index(rest, "-->")
		if end < 0 {
			return content
		}
		rest = rest[end+3:]
	}
	if openStart < 0 || !strings.HasPrefix(strings.ToLower(rest), "</figcaption>") {
		return content
	}
	closeEnd := len(content) - len(rest) + len("</figcaption>")
	return content[:openStart] + content[closeEnd:]
}

// 布局类型。
const (
	layoutDefault     = "default"
	layoutConstrained = "constrained"
	layoutFlex        = "flex"
	layoutGrid        = "grid"
)

// layoutBlocks 是支持 layout 的核心区块及其默认布局类型（取自各 block.json）。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var layoutBlocks = map[string]string{
	"core/accordion-item":      "",
	"core/accordion-panel":     "",
	"core/accordion":           "",
	"core/buttons":             layoutFlex,
	"core/column":              "",
	"core/columns":             layoutFlex,
	"core/comments-pagination": layoutFlex,
	"core/cover":               "",
	"core/details":             "",
	"core/gallery":             layoutFlex,
	"core/group":               "",
	"core/latest-posts":        "",
	"core/navigation":          layoutFlex,
	"core/post-content":        "",
	"core/post-template":       "",
	"core/query-pagination":    layoutFlex,
	"core/query":               "",
	"core/quote":               "",
	"core/social-links":        layoutFlex,
	"core/tab-list":            layoutFlex,
	"core/tab-panel":           "",
	"core/tabs":                "",
	"core/term-template":       "",
	"core/terms-query":         "",
}

// layoutClassNames 对应 wp_get_layout_definitions() 中各布局的 className。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var layoutClassNames = map[string]string{
	layoutDefault:     "is-layout-flow",
	layoutConstrained: "is-layout-constrained",
	layoutFlex:        "is-layout-flex",
	layoutGrid:        "is-layout-grid",
}

func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return !phpEmpty(x)
	case float64:
		return x != 0
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	default:
		return true
	}
}

// renderLayoutSupport 是 wp_render_layout_support_flag() 的简化版本，见 doBlocks 的说明。
func renderLayoutSupport(block *parsedBlock, content string) string {
	defaultType, supported := layoutBlocks[block.name]
	if !supported {
		return content
	}
	p := newTagProcessor(content)
	if !p.nextTag("") {
		return content
	}

	// 有内部区块时，类名加在包裹内部区块的元素上。
	if wrapper := layoutInnerWrapperClass(block); wrapper != "" {
		for {
			if class, ok := p.getAttribute("class"); ok && strings.Contains(class, wrapper) {
				break
			}
			if !p.nextTag("") {
				break
			}
		}
	}

	for _, c := range layoutClasses(block, defaultType) {
		p.addClass(c)
	}
	return p.updatedHTML()
}

// layoutClasses 计算布局支持要添加的类名（不含 wp-container-*）。
func layoutClasses(block *parsedBlock, defaultType string) []string {
	usedLayout := map[string]any{}
	if defaultType != "" {
		usedLayout["type"] = defaultType
	}
	attrLayout, _ := block.attrs["layout"].(map[string]any)
	if attrLayout != nil {
		usedLayout = attrLayout
	}
	if truthy(usedLayout["inherit"]) || truthy(usedLayout["contentSize"]) {
		usedLayout = map[string]any{"type": layoutConstrained}
	}

	var classNames []string
	if v, ok := attrLayout["orientation"].(string); ok && !phpEmpty(v) {
		classNames = append(classNames, "is-"+SanitizeTitle(v))
	}
	if v, ok := attrLayout["justifyContent"].(string); ok && !phpEmpty(v) {
		classNames = append(classNames, "is-content-justification-"+SanitizeTitle(v))
	}
	if v, ok := attrLayout["flexWrap"].(string); ok && v == "nowrap" {
		classNames = append(classNames, "is-nowrap")
	}

	layoutType := layoutDefault
	if t, ok := usedLayout["type"].(string); ok {
		layoutType = t
	}
	layoutClass := layoutClassNames[layoutType]
	if layoutClass != "" {
		classNames = append(classNames, SanitizeTitle(layoutClass))
	}

	parts := strings.Split(block.name, "/")
	fullName := strings.Join(parts, "-")
	if parts[0] == "core" {
		fullName = parts[len(parts)-1]
	}
	return append(classNames, "wp-block-"+fullName+"-"+layoutClass)
}

// layoutInnerWrapperClass 返回包裹内部区块的最内层元素的 class 属性；没有内部区块时返回空字符串。
func layoutInnerWrapperClass(block *parsedBlock) string {
	if len(block.innerContent) <= 1 || block.innerContent[0] == nil {
		return ""
	}
	var stack []string
	p := newTagProcessor(*block.innerContent[0])
	for p.nextToken() {
		switch {
		case p.cur.closer:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case !voidElements[p.cur.name]:
			class, _ := p.getAttribute("class")
			stack = append(stack, class)
		}
	}
	for _, class := range slices.Backward(stack) {
		if class != "" {
			return class
		}
	}
	return ""
}

// voidElements 对应 WP_HTML_Processor::is_void()。
//
//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var voidElements = map[string]bool{
	"area": true, "base": true, "basefont": true, "bgsound": true, "br": true, "col": true,
	"embed": true, "frame": true, "hr": true, "img": true, "input": true, "keygen": true,
	"link": true, "meta": true, "param": true, "source": true, "track": true, "wbr": true,
}
