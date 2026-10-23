package wpfmt

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	contentTagsRegex  = regexp.MustCompile(`<(img|iframe)` + reSpace + `[^>]+>`)
	imgWidthRegex     = regexp.MustCompile(` width=["']([0-9]+)["']`)
	imgHeightRegex    = regexp.MustCompile(` height=["']([0-9]+)["']`)
	imgLoadingRegex   = regexp.MustCompile(` loading=["']([A-Za-z]+)["']`)
	imgFetchPrioRegex = regexp.MustCompile(` fetchpriority=["']([A-Za-z]+)["']`)
	imgDecodingRegex  = regexp.MustCompile(` decoding=["']([A-Za-z]+)["']`)
)

// wpFilterContentTags 移植自 wp_filter_content_tags($content, 'the_content')。
// renderState.inMainLoop 区分 REST 请求与 feed 主循环两种场景。
//
// 带 wp-image-{id} 类名的图片会根据 renderState.att 中的附件元数据补充 width/height/srcset/sizes。
func (st *renderState) wpFilterContentTags(content string) string {
	matches := contentTagsRegex.FindAllStringSubmatch(content, -1)
	if matches == nil {
		return content
	}

	done := map[string]bool{}
	for _, m := range matches {
		tag, name := m[0], m[1]
		if done[tag] {
			continue
		}
		done[tag] = true

		var filtered string
		if name == tagNameImg {
			filtered = st.filterImageTag(tag)
		} else {
			filtered = tag
			if !strings.Contains(tag, " loading=") {
				filtered = st.iframeTagAddLoadingAttr(tag)
			}
		}
		if filtered != tag {
			content = strings.ReplaceAll(content, tag, filtered)
		}
	}
	return content
}

// filterImageTag 依次执行 wp_filter_content_tags() 中针对单个 img 标签的处理。
func (st *renderState) filterImageTag(image string) string {
	if id := imgAttachmentID(image); id > 0 {
		if !strings.Contains(image, " width=") && !strings.Contains(image, " height=") {
			image = st.imgTagAddWidthAndHeightAttr(image, id)
		}
		if !strings.Contains(image, " srcset=") {
			image = st.imgTagAddSrcsetAndSizesAttr(image, id)
		}
	}
	return imgTagAddAutoSizes(st.imgTagAddLoadingOptimizationAttrs(image))
}

func submatch(re *regexp.Regexp, s string) (string, bool) {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// imgTagAddLoadingOptimizationAttrs 移植自 wp_img_tag_add_loading_optimization_attrs()。
func (st *renderState) imgTagAddLoadingOptimizationAttrs(image string) string {
	widthStr, hasWidth := submatch(imgWidthRegex, image)
	heightStr, hasHeight := submatch(imgHeightRegex, image)
	loadingVal, _ := submatch(imgLoadingRegex, image)
	fetchPrioVal, _ := submatch(imgFetchPrioRegex, image)
	decodingVal, hasDecoding := submatch(imgDecodingRegex, image)

	width, _ := strconv.Atoi(widthStr)
	height, _ := strconv.Atoi(heightStr)
	opt := st.loadingOptimizationAttributes(
		true,
		hasWidth && hasHeight,
		width*height,
		loadingVal,
		fetchPrioVal,
		decodingVal,
		hasDecoding,
	)

	if !strings.Contains(image, ` src="`) {
		return image
	}

	if phpEmpty(decodingVal) && opt.decoding != "" {
		image = strings.ReplaceAll(image, "<img", `<img decoding="`+escAttr(opt.decoding)+`"`)
	}

	if !strings.Contains(image, ` width="`) || !strings.Contains(image, ` height="`) {
		return image
	}

	if phpEmpty(loadingVal) && opt.loading != "" {
		image = strings.ReplaceAll(image, "<img", `<img loading="`+escAttr(opt.loading)+`"`)
	}

	if phpEmpty(fetchPrioVal) && opt.fetchPriority != "" {
		image = strings.ReplaceAll(image, "<img", `<img fetchpriority="`+escAttr(opt.fetchPriority)+`"`)
	}
	return image
}

const (
	loadingLazy  = "lazy"
	priorityHigh = "high"
	priorityLow  = "low"
	priorityAuto = "auto"
	tagNameImg   = "img"
)

type loadingAttrs struct {
	decoding, loading, fetchPriority string
}

// loadingOptimizationAttributes 移植自 wp_get_loading_optimization_attributes()。
// REST 请求不在主循环内、也没有触发 get_header；feed 请求在主循环内（in_the_loop && is_main_query），
// 前 wp_omit_loading_attr_threshold() 个媒体元素视为在首屏内，不加 loading="lazy"。
func (st *renderState) loadingOptimizationAttributes(
	isImg, hasDims bool, area int, loading, fetchPriority, decoding string, hasDecoding bool,
) loadingAttrs {
	var attrs loadingAttrs
	if isImg {
		attrs.decoding = "async"
		if hasDecoding {
			attrs.decoding = decoding
		}
	}
	if !hasDims {
		return attrs
	}

	// nil 表示未知，与 PHP 中的 null 对应。
	var maybeInViewport *bool
	setViewport := func(v bool) { maybeInViewport = &v }
	if loading != "" {
		setViewport(loading != loadingLazy)
	}

	isLowFetchPriority := fetchPriority == priorityLow
	increaseCount := false
	switch fetchPriority {
	case priorityHigh:
		if maybeInViewport != nil && !*maybeInViewport {
			attrs.fetchPriority = priorityHigh
		} else {
			setViewport(true)
		}
	case priorityLow:
		setViewport(false)
		attrs.fetchPriority = priorityLow
	case priorityAuto:
		attrs.fetchPriority = priorityAuto
	}

	if maybeInViewport == nil && st.inMainLoop {
		increaseCount = true
		setViewport(st.contentMediaCount < omitLoadingAttrThreshold)
	}

	switch {
	case maybeInViewport != nil && *maybeInViewport:
		if isImg {
			st.maybeAddFetchPriorityHigh(&attrs, fetchPriority, area)
		}
	case !isLowFetchPriority:
		attrs.loading = loadingLazy
	}

	if fetchPriority != priorityAuto && increaseCount {
		st.contentMediaCount++
	}
	return attrs
}

// omitLoadingAttrThreshold 是 wp_omit_loading_attr_threshold() 的默认值。
const omitLoadingAttrThreshold = 3

// maybeAddFetchPriorityHigh 移植自 wp_maybe_add_fetchpriority_high_attr()。
// wp_high_priority_element_flag() 的状态保存在 renderState 中：TheContent 每次调用独立，
// FeedRenderer 在整个 feed 请求内共享。
//
//nolint:mnd // 正则分组编号与阈值沿用 WordPress 原实现
func (st *renderState) maybeAddFetchPriorityHigh(attrs *loadingAttrs, existing string, area int) {
	if existing != "" && existing != priorityAuto {
		if existing == priorityHigh {
			attrs.fetchPriority = priorityHigh
			st.highPriorityDone = true
		}
		return
	}
	if attrs.loading == loadingLazy || st.highPriorityDone {
		return
	}
	if area >= 50000 {
		if existing != priorityAuto {
			attrs.fetchPriority = priorityHigh
		}
		st.highPriorityDone = true
	}
}

// iframeTagAddLoadingAttr 移植自 wp_iframe_tag_add_loading_attr()。
func (st *renderState) iframeTagAddLoadingAttr(iframe string) string {
	hasDims := strings.Contains(iframe, ` width="`) && strings.Contains(iframe, ` height="`)
	opt := st.loadingOptimizationAttributes(false, hasDims, 0, "", "", "", false)
	if !strings.Contains(iframe, ` src="`) || !hasDims || opt.loading == "" {
		return iframe
	}
	return strings.ReplaceAll(iframe, "<iframe", `<iframe loading="`+escAttr(opt.loading)+`"`)
}

// imgTagAddAutoSizes 移植自 wp_img_tag_add_auto_sizes()。
func imgTagAddAutoSizes(image string) string {
	tp := newTagProcessor(image)
	if !tp.nextTag(tagNameImg) {
		return image
	}
	loading, ok := tp.getAttribute("loading")
	if !ok || strings.ToLower(strings.Trim(loading, " \t\f\r\n")) != loadingLazy {
		return image
	}
	if width, hasWidth := tp.getAttribute("width"); !hasWidth || width == "" {
		return image
	}
	sizes, ok := tp.getAttribute("sizes")
	// wp_sizes_attribute_includes_valid_auto()
	first, _, _ := strings.Cut(sizes, ",")
	if !ok || strings.ToLower(strings.Trim(first, " \t\f\r\n")) == "auto" {
		return image
	}
	tp.setAttribute("sizes", "auto, "+sizes)
	return tp.updatedHTML()
}
