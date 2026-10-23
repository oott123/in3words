package wpfmt

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Attachments 是附件 ID 到 _wp_attachment_metadata 原文（PHP serialize() 格式）的映射。
// 附件不存在或没有元数据时不在映射中；nil 表示没有任何附件信息。
type Attachments map[uint64]string

var (
	wpImageClassRegex = regexp.MustCompile(`(?i)wp-image-([0-9]+)`)
	imgSrcAttrRegex   = regexp.MustCompile(`src="([^"]+)"`)
	imgStyleWidth     = regexp.MustCompile(`style="width:` + reSpace + `*([0-9]+)px;"`)
	imgEditHashRegex  = regexp.MustCompile(`-e[0-9]{13}`)
	imgCloseRegex     = regexp.MustCompile(`<img ([^>]+?)[/ ]*>`)
)

// AttachmentIDs 按 wp_filter_content_tags() 的规则找出正文中需要查询元数据的附件 ID（按出现顺序去重）。
// 调用方可据此批量查询 _wp_attachment_metadata 并构造 Attachments。
func AttachmentIDs(content string) []uint64 {
	var ids []uint64
	seen := map[uint64]bool{}
	for _, m := range contentTagsRegex.FindAllStringSubmatch(content, -1) {
		if m[1] != tagNameImg {
			continue
		}
		if id := imgAttachmentID(m[0]); id > 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

// imgAttachmentID 返回 img 标签中 wp-image-{id} 类名对应的附件 ID，没有时返回 0。
func imgAttachmentID(tag string) uint64 {
	m := wpImageClassRegex.FindStringSubmatch(tag)
	if m == nil {
		return 0
	}
	id, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// attachmentMetadata 对应 wp_get_attachment_metadata()：元数据不存在、无法解析或为空数组时返回 nil。
func (st *renderState) attachmentMetadata(id uint64) *phpArray {
	if meta, ok := st.metaCache[id]; ok {
		return meta
	}
	var meta *phpArray
	if raw, ok := st.att[id]; ok {
		meta = parseAttachmentMetadata(raw)
	}
	if st.metaCache == nil {
		st.metaCache = map[uint64]*phpArray{}
	}
	st.metaCache[id] = meta
	return meta
}

// parseAttachmentMetadata 解析 _wp_attachment_metadata；不是非空数组时返回 nil，sizes 不是数组时视为空数组。
func parseAttachmentMetadata(raw string) *phpArray {
	v, err := phpUnserialize(raw)
	if err != nil {
		return nil
	}
	arr, isArr := v.(*phpArray)
	if !isArr || arr.len() == 0 {
		return nil
	}
	if sizes, has := arr.get("sizes"); has {
		if _, sizesArr := sizes.(*phpArray); !sizesArr {
			arr.set("sizes", newPHPArray())
		}
	}
	return arr
}

func metaString(meta *phpArray, key string) (string, bool) {
	v, ok := meta.get(key)
	if !ok || v == nil {
		return "", false
	}
	return phpToString(v), true
}

func metaSizes(meta *phpArray) *phpArray {
	v, _ := meta.get("sizes")
	arr, _ := v.(*phpArray)
	return arr
}

// imgSrc 返回 src 属性中 ? 之前的部分。
func imgSrc(image string) string {
	m := imgSrcAttrRegex.FindStringSubmatch(image)
	if m == nil {
		return ""
	}
	src, _, _ := strings.Cut(m[1], "?")
	return src
}

// wpBasename 对应 wp_basename()。
func wpBasename(path string) string {
	path = strings.TrimRight(strings.ReplaceAll(path, `\`, "/"), "/")
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

// imgTagAddWidthAndHeightAttr 移植自 wp_img_tag_add_width_and_height_attr()。
func (st *renderState) imgTagAddWidthAndHeightAttr(image string, attachmentID uint64) string {
	src := imgSrc(image)
	if src == "" {
		return image
	}
	w, h, ok := imageSrcGetDimensions(src, st.attachmentMetadata(attachmentID))
	if !ok || w == 0 || h == 0 {
		return image
	}
	// 宽度由行内样式指定时（如行内图片），按比例计算尺寸属性。
	if m := imgStyleWidth.FindStringSubmatch(image); m != nil {
		if styleWidth, _ := strconv.Atoi(m[1]); styleWidth != 0 {
			h = int(math.Round(float64(h) * float64(styleWidth) / float64(w)))
			w = styleWidth
		}
	}
	return strings.ReplaceAll(image, "<img", "<img "+imageHWString(w, h))
}

// imageHWString 对应 trim(image_hwstring())。
func imageHWString(w, h int) string {
	var parts []string
	if w != 0 {
		parts = append(parts, fmt.Sprintf(`width="%d"`, w))
	}
	if h != 0 {
		parts = append(parts, fmt.Sprintf(`height="%d"`, h))
	}
	return strings.Join(parts, " ")
}

// imageSrcGetDimensions 移植自 wp_image_src_get_dimensions()。
func imageSrcGetDimensions(src string, meta *phpArray) (int, int, bool) {
	if meta == nil {
		return 0, 0, false
	}
	if file, ok := metaString(meta, "file"); ok && strings.Contains(src, wpBasename(file)) {
		w, _ := meta.get("width")
		h, _ := meta.get("height")
		return phpToInt(w), phpToInt(h), true
	}
	sizes := metaSizes(meta)
	if sizes.len() == 0 {
		return 0, 0, false
	}
	srcFilename := wpBasename(src)
	for _, key := range sizes.keys {
		size, _ := sizes.values[key].(*phpArray)
		file, _ := size.get("file")
		if s, isStr := file.(string); isStr && s == srcFilename {
			w, _ := size.get("width")
			h, _ := size.get("height")
			return phpToInt(w), phpToInt(h), true
		}
	}
	return 0, 0, false
}

// imgTagAddSrcsetAndSizesAttr 移植自 wp_img_tag_add_srcset_and_sizes_attr() / wp_image_add_srcset_and_sizes()。
func (st *renderState) imgTagAddSrcsetAndSizesAttr(image string, attachmentID uint64) string {
	meta := st.attachmentMetadata(attachmentID)
	if metaSizes(meta).len() == 0 {
		return image
	}
	src := imgSrc(image)
	if src == "" {
		return image
	}
	// 图片插入后又被编辑过时不处理。
	if file, ok := metaString(meta, "file"); ok {
		if hash := imgEditHashRegex.FindString(file); hash != "" && !strings.Contains(wpBasename(src), hash) {
			return image
		}
	}

	widthStr, hasWidth := submatch(imgWidthRegex, image)
	heightStr, hasHeight := submatch(imgHeightRegex, image)
	width, _ := strconv.Atoi(widthStr)
	height, _ := strconv.Atoi(heightStr)
	if !hasWidth || !hasHeight || width == 0 || height == 0 {
		var ok bool
		width, height, ok = imageSrcGetDimensions(src, meta)
		if !ok {
			return image
		}
	}

	srcset := st.calculateImageSrcset(width, height, src, meta)
	if srcset == "" {
		return image
	}
	attr := ` srcset="` + escAttr(srcset) + `"`
	// 已有 sizes 属性时只添加 srcset；calculateImageSrcset 已保证 width >= 1。
	if !strings.Contains(image, " sizes=") {
		attr += ` sizes="` + escAttr(fmt.Sprintf("(max-width: %[1]dpx) 100vw, %[1]dpx", width)) + `"`
	}
	return imgCloseRegex.ReplaceAllString(image, "<img ${1}"+strings.ReplaceAll(attr, "$", "$$")+" />")
}

// maxSrcsetImageWidth 是 max_srcset_image_width 过滤器的默认值。
const maxSrcsetImageWidth = 2048

type srcsetSource struct {
	key   string
	url   string
	value string
}

// srcsetCandidate 是参与 srcset 计算的一个尺寸。
type srcsetCandidate struct{ file, width, height any }

// srcsetCandidates 返回 $image_meta['sizes'] 中的有效尺寸，必要时在末尾追加原图；
// ok 为 false 表示应当直接放弃 srcset（GIF 动图且 src 就是原图）。
func srcsetCandidates(meta, sizes *phpArray, file, src string) ([]srcsetCandidate, bool) {
	candidates := make([]srcsetCandidate, 0, sizes.len()+1)
	for _, key := range sizes.keys {
		size, isArr := sizes.values[key].(*phpArray)
		if !isArr {
			// 元数据损坏的尺寸直接跳过。
			continue
		}
		f, _ := size.get("file")
		w, _ := size.get("width")
		h, _ := size.get("height")
		candidates = append(candidates, srcsetCandidate{f, w, h})
	}

	// GIF 缩略图意味着可能是动图，此时不把原图加入候选；src 就是原图时不输出 srcset。
	thumbnailIsGIF := false
	if thumb, hasThumb := sizes.get("thumbnail"); hasThumb {
		if t, isArr := thumb.(*phpArray); isArr {
			mime, hasMime := t.get("mime-type")
			thumbnailIsGIF = hasMime && phpToString(mime) == "image/gif"
		}
	}
	if thumbnailIsGIF {
		return candidates, !strings.Contains(src, file)
	}
	w, _ := meta.get("width")
	h, _ := meta.get("height")
	return append(candidates, srcsetCandidate{wpBasename(file), w, h}), true
}

// addSrcsetSource 模拟 PHP 中以宽度为键的数组写入：src 对应的尺寸放在最前面，相同宽度覆盖原位置。
func addSrcsetSource(sources []srcsetSource, source srcsetSource, isSrc bool) []srcsetSource {
	idx := -1
	for i := range sources {
		if sources[i].key == source.key {
			idx = i
			break
		}
	}
	switch {
	case isSrc:
		// src 对应的图片必须排在第一位（iOS8 的 bug，见 #35030）。
		if idx >= 0 {
			sources = append(sources[:idx], sources[idx+1:]...)
		}
		return append([]srcsetSource{source}, sources...)
	case idx >= 0:
		sources[idx] = source
		return sources
	default:
		return append(sources, source)
	}
}

// calculateImageSrcset 移植自 wp_calculate_image_srcset()，没有可用的 srcset 时返回空字符串。
func (st *renderState) calculateImageSrcset(width, height int, src string, meta *phpArray) string {
	sizes := metaSizes(meta)
	file, hasFile := metaString(meta, "file")
	if sizes.len() == 0 || !hasFile || len(file) < 4 || width < 1 {
		return ""
	}
	candidates, ok := srcsetCandidates(meta, sizes, file, src)
	if !ok {
		return ""
	}

	dirname := attachmentRelativePath(file)
	if dirname != "" {
		dirname = strings.TrimRight(dirname, `/\`) + "/"
	}
	baseURL := strings.TrimRight(st.opts.UploadsBaseURL, `/\`) + "/" + dirname
	editHash := imgEditHashRegex.FindString(wpBasename(src))

	var sources []srcsetSource
	srcMatched := false
	for _, c := range candidates {
		fileName := phpToString(c.file)
		isSrc := false
		if !srcMatched && strings.Contains(src, dirname+fileName) {
			srcMatched, isSrc = true, true
		}
		// 跳过之前编辑版本留下的尺寸。
		if editHash != "" && strings.Index(fileName, editHash) <= 0 {
			continue
		}
		if phpToFloat(c.width) > maxSrcsetImageWidth && !isSrc {
			continue
		}
		if !imageMatchesRatio(float64(width), float64(height), phpToFloat(c.width), phpToFloat(c.height)) {
			continue
		}
		sources = addSrcsetSource(sources, srcsetSource{
			key:   phpArrayKey(c.width),
			url:   baseURL + fileName,
			value: phpToString(c.width),
		}, isSrc)
	}

	// 至少两个候选才输出 srcset。
	if !srcMatched || len(sources) < 2 {
		return ""
	}
	parts := make([]string, len(sources))
	for i, s := range sources {
		parts[i] = strings.ReplaceAll(s.url, " ", "%20") + " " + s.value + "w"
	}
	return strings.Join(parts, ", ")
}

// attachmentRelativePath 对应 _wp_get_attachment_relative_path()。
func attachmentRelativePath(file string) string {
	dirname := phpDirname(file)
	if dirname == "." {
		return ""
	}
	if i := strings.Index(dirname, "wp-content/uploads"); i >= 0 {
		dirname = strings.TrimLeft(dirname[i+len("wp-content/uploads"):], "/")
	}
	return dirname
}

// phpDirname 对应 PHP 的 dirname()（Unix 语义）。
func phpDirname(path string) string {
	trimmed := strings.TrimRight(path, "/")
	if trimmed == "" {
		if path == "" {
			return "."
		}
		return "/"
	}
	i := strings.LastIndexByte(trimmed, '/')
	if i < 0 {
		return "."
	}
	dir := strings.TrimRight(trimmed[:i], "/")
	if dir == "" {
		return "/"
	}
	return dir
}

// imageMatchesRatio 移植自 wp_image_matches_ratio()。
func imageMatchesRatio(sourceWidth, sourceHeight, targetWidth, targetHeight float64) bool {
	var constrainedW, constrainedH, expectedW, expectedH float64
	if sourceWidth > targetWidth {
		constrainedW, constrainedH = constrainDimensions(sourceWidth, sourceHeight, targetWidth)
		expectedW, expectedH = targetWidth, targetHeight
	} else {
		constrainedW, constrainedH = constrainDimensions(targetWidth, targetHeight, sourceWidth)
		expectedW, expectedH = sourceWidth, sourceHeight
	}
	return math.Abs(constrainedW-expectedW) <= 1 && math.Abs(constrainedH-expectedH) <= 1
}

// constrainDimensions 移植自 wp_constrain_dimensions($w, $h, $maxWidth, 0)。
func constrainDimensions(currentWidth, currentHeight, maxWidth float64) (float64, float64) {
	if maxWidth == 0 {
		return currentWidth, currentHeight
	}
	widthRatio, heightRatio := 1.0, 1.0
	didWidth := false
	if maxWidth > 0 && currentWidth > 0 && currentWidth > maxWidth {
		widthRatio = maxWidth / currentWidth
		didWidth = true
	}
	smaller := math.Min(widthRatio, heightRatio)
	larger := math.Max(widthRatio, heightRatio)
	ratio := larger
	if math.Round(currentWidth*larger) > maxWidth || math.Round(currentHeight*larger) > 0 {
		ratio = smaller
	}
	w := math.Max(1, math.Round(currentWidth*ratio))
	h := math.Max(1, math.Round(currentHeight*ratio))
	if didWidth && w == maxWidth-1 {
		w = maxWidth
	}
	return w, h
}
