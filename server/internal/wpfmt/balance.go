package wpfmt

import (
	"regexp"
	"slices"
	"strings"
)

//nolint:gochecknoglobals // 预编译的正则与移植自 WordPress 的只读常量表
var (
	balanceLoveFix = regexp.MustCompile(`<([0-9])`)
	balanceTag     = regexp.MustCompile(
		`<(/?)((?:[a-z](?:[a-z0-9._]*)-(?:[a-z0-9._-]+)+)|(?:[A-Za-z0-9_:]+))` +
			`(?:` + reSpace + `*(/?)|(` + reSpace + `+)([^>]*))>`,
	)

	balanceSingleTags = []string{
		"area", "base", "basefont", "br", "col", "command", "embed", "frame", "hr", "img",
		"input", "isindex", "link", "meta", "param", "source", "track", "wbr",
	}
	balanceNestableTags = []string{
		"article",
		"aside",
		"blockquote",
		"details",
		"div",
		"figure",
		"object",
		"q",
		"section",
		"span",
	}
)

// forceBalanceTags 移植自 force_balance_tags()。
//
//nolint:funlen,gocognit,mnd,nestif // 逐行移植 WordPress 实现，保持原有结构便于对照；正则分组编号与阈值沿用 WordPress 原实现
func forceBalanceTags(text string) string {
	var tagstack []string
	var tagqueue string
	var newtext strings.Builder

	// 兼容有人真的想输入 "< !--" 的情况。
	text = strings.ReplaceAll(text, "< !--", "<    !--")
	// 兼容 "<3" 之类 < 后面跟数字的情况。
	text = balanceLoveFix.ReplaceAllString(text, "&lt;$1")

	for {
		loc := balanceTag.FindStringSubmatchIndex(text)
		if loc == nil {
			break
		}
		group := func(i int) (string, bool) {
			if loc[2*i] < 0 {
				return "", false
			}
			return text[loc[2*i]:loc[2*i+1]], true
		}
		fullMatch, _ := group(0)
		leadingSlash, _ := group(1)
		tagName, _ := group(2)
		tag := strings.ToLower(tagName)
		isSingleTag := containsString(balanceSingleTags, tag)
		preAttributeWS, _ := group(4)
		attributes, ok := group(5)
		if !ok {
			attributes, _ = group(3)
		}
		attributes = phpTrim(attributes)
		hasSelfCloser := strings.HasSuffix(attributes, "/")

		newtext.WriteString(tagqueue)
		i := loc[0]
		tagqueue = ""

		if !phpEmpty(leadingSlash) {
			// 结束标签。
			switch {
			case len(tagstack) == 0:
				tag = ""
			case tagstack[len(tagstack)-1] == tag:
				tag = "</" + tag + ">"
				tagstack = tagstack[:len(tagstack)-1]
			default:
				// 结束标签不在栈顶，往下查找。
				if j := lastIndexOf(tagstack, tag); j >= 0 {
					var closing strings.Builder
					for _, t := range slices.Backward(tagstack[j:]) {
						closing.WriteString("</" + t + ">")
					}
					tagqueue += closing.String()
					tagstack = tagstack[:j]
				}
				tag = ""
			}
		} else {
			// 开始标签。
			switch {
			case hasSelfCloser:
				// 不是已知单标签却自闭合时，补上结束标签。
				if !isSingleTag {
					attributes = phpTrim(attributes[:len(attributes)-1]) + "></" + tag
				}
			case isSingleTag:
				// 已知单标签但没有自闭合，补上 /。
				preAttributeWS = " "
				attributes += "/"
			default:
				// 不允许嵌套自身的标签遇到同名栈顶时，先关闭栈顶。
				if len(tagstack) > 0 && !containsString(balanceNestableTags, tag) && tagstack[len(tagstack)-1] == tag {
					tagqueue = "</" + tagstack[len(tagstack)-1] + ">"
					tagstack = tagstack[:len(tagstack)-1]
				}
				tagstack = append(tagstack, tag)
			}

			if hasSelfCloser && isSingleTag {
				// 偏好 <br /> 而不是 <br/>。
				preAttributeWS = " "
			}

			tag = "<" + tag + preAttributeWS + attributes + ">"
			if !phpEmpty(tagqueue) {
				tagqueue += tag
				tag = ""
			}
		}
		newtext.WriteString(text[:i])
		newtext.WriteString(tag)
		text = text[i+len(fullMatch):]
	}

	newtext.WriteString(tagqueue)
	newtext.WriteString(text)

	// array_pop 遇到假值（如 "0"）时停止。
	for len(tagstack) > 0 {
		x := tagstack[len(tagstack)-1]
		tagstack = tagstack[:len(tagstack)-1]
		if phpEmpty(x) {
			break
		}
		newtext.WriteString("</" + x + ">")
	}

	out := newtext.String()
	out = strings.ReplaceAll(out, "< !--", "<!--")
	return strings.ReplaceAll(out, "<    !--", "< !--")
}

// lastIndexOf 返回 s 在 list 中最后一次出现的位置，不存在时返回 -1。
func lastIndexOf(list []string, s string) int {
	for i, v := range slices.Backward(list) {
		if v == s {
			return i
		}
	}
	return -1
}
