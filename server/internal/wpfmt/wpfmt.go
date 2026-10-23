// Package wpfmt 移植了 WordPress 7.1.3 中生成 REST API content.rendered、
// title.rendered 等字段所需的 PHP 格式化逻辑（wptexturize、wpautop、短代码、区块等）。
//
// 移植以 WordPress 默认（en_US、未安装插件）的过滤器链为准，与 WP 的已知差异写在
// 各函数的注释中。
package wpfmt

import "strings"

// Options 是格式化时需要的站点设置。
type Options struct {
	HomeURL                string // "home" option，给 wp_replace_insecure_home_url 与 wp_is_internal_link 用
	SiteURL                string // "siteurl" option，用于拼出 :mrgreen: 等图片表情的地址
	HTTPSMigrationRequired bool   // "https_migration_required" option
	UseSmilies             bool   // "use_smilies" option
	// EmojiBaseURL 与 EmojiExt 对应 emoji_url / emoji_ext 过滤器的默认值。
	// WordPress 7.1.3 的 convert_smilies() 直接输出 emoji 字符，只有 feed / 邮件中的
	// wp_staticize_emoji() 才会用到它们，TheContent 与 CommentText 不使用。
	EmojiBaseURL string
	EmojiExt     string
	HTML5Caption bool // current_theme_supports('html5','caption')
	// UploadsBaseURL 是 wp_get_upload_dir()['baseurl']，用于拼出 srcset 中各尺寸图片的地址。
	UploadsBaseURL string
	// AutopAfterUnautop 模拟同一请求中渲染过含区块的内容后，_restore_wpautop_hook 把 wpautop 挂到优先级 10 队尾的状态。
	AutopAfterUnautop bool
}

// DefaultOptions 返回 WordPress 全新安装时的默认设置。
func DefaultOptions() Options {
	return Options{
		UseSmilies:   true,
		EmojiBaseURL: "https://s.w.org/images/core/emoji/17.0.2/72x72/",
		EmojiExt:     ".png",
		HTML5Caption: true,
	}
}

// renderState 保存渲染过程中 WordPress 用静态变量保存的状态。
type renderState struct {
	opts              Options
	uniqueIDs         map[string]int // wp_unique_prefixed_id()
	highPriorityDone  bool           // wp_high_priority_element_flag() 已被置为 false
	inMainLoop        bool           // in_the_loop() && is_main_query()
	contentMediaCount int            // wp_increase_content_media_count()
	att               Attachments
	metaCache         map[uint64]*phpArray
}

// TheContent 依次执行 WP 默认的 the_content 过滤器：
// WP_Embed::run_shortcode(8)、WP_Embed::autoembed(8)、do_blocks(9)、wptexturize(10)、
// wpautop(10，内容含有区块时跳过)、shortcode_unautop(10)、
// wp_replace_insecure_home_url(10)、capital_P_dangit(11)、do_shortcode(11)、
// wp_filter_content_tags(12)、convert_smilies(20)。
// opts.AutopAfterUnautop 为 true 时，wpautop 改为在 wp_replace_insecure_home_url 之后执行，
// 因此独立成段的短代码外层会保留 <p>。
//
// 与 WP 的差异：无法读取 oEmbed 缓存，独占一行的 YouTube 等链接不会变成嵌入代码（见 autoembed）。
//
// att 提供正文中 wp-image-{id} 图片的附件元数据（见 AttachmentIDs），为 nil 时不补充 srcset 等属性。
func TheContent(content string, opts Options, att Attachments) string {
	st := &renderState{opts: opts, att: att}
	content, _ = st.theContent(content)
	return content
}

// theContent 执行 the_content 过滤器链，同时返回 do_blocks 是否遇到了区块
// （即之后的调用是否进入 AutopAfterUnautop 状态）。
func (st *renderState) theContent(content string) (string, bool) {
	opts := st.opts
	content = st.doShortcode(content, true, embedShortcodes, embedShortcodeOrder)
	content = autoembed(content)

	hadBlocks := HasBlocks(content)
	runAutop := !hadBlocks
	content = st.doBlocks(content)
	content = wptexturize(content)
	if runAutop && !opts.AutopAfterUnautop {
		content = wpautop(content)
	}
	content = shortcodeUnautop(content)
	content = st.replaceInsecureHomeURL(content)
	if runAutop && opts.AutopAfterUnautop {
		content = wpautop(content)
	}
	content = capitalPDangit(content)
	content = st.doShortcode(content, false, contentShortcodes, registeredShortcodes)
	content = st.wpFilterContentTags(content)
	return st.convertSmilies(content), hadBlocks
}

// TheTitle 对应 the_title 过滤器：wptexturize、convert_chars、trim、capital_P_dangit（标题版本）。
// 受密码保护或私密文章的"密码保护："等前缀由调用方处理。
func TheTitle(title string) string {
	title = wptexturize(title)
	title = convertChars(title)
	title = phpTrim(title)
	return capitalPDangitTitle(title)
}

// CommentText 对应 REST 评论 content.rendered 使用的 comment_text 过滤器链（非 admin 环境）：
// make_clickable(9)、wptexturize(10)、convert_chars(10)、convert_smilies(20)、
// force_balance_tags(25)、wpautop(30)、capital_P_dangit(31)。
func CommentText(content string, opts Options) string {
	st := &renderState{opts: opts}
	content = st.makeClickable(content)
	content = wptexturize(content)
	content = convertChars(content)
	content = st.convertSmilies(content)
	content = forceBalanceTags(content)
	content = wpautop(content)
	return capitalPDangit(content)
}

// replaceInsecureHomeURL 移植自 wp_replace_insecure_home_url()。
func (st *renderState) replaceInsecureHomeURL(content string) string {
	home := strings.TrimRight(st.opts.HomeURL, "/")
	site := st.opts.SiteURL
	if site == "" {
		site = home
	}
	if !st.opts.HTTPSMigrationRequired ||
		!strings.HasPrefix(strings.ToLower(home), "https://") ||
		!strings.HasPrefix(strings.ToLower(site), "https://") ||
		urlHost(home) != urlHost(site) {
		return content
	}
	httpsURL := "https://" + home[len("https://"):]
	httpURL := "http://" + home[len("https://"):]
	return strReplacePairs(content,
		httpURL, httpsURL,
		strings.ReplaceAll(httpURL, "/", `\/`), strings.ReplaceAll(httpsURL, "/", `\/`),
	)
}
