<?php
/**
 * 生成 wpfmt 包的金标准测试数据。
 *
 * 在 WordPress Playground 中运行（需要 Node 24）：
 *
 *     cd server/internal/wpfmt/testdata
 *     mise exec node@24 -- npx -y @wp-playground/cli@latest php \
 *         --mount=$PWD:/work -- /work/gen.php
 *
 * 脚本会把每条语料写成一篇文章，再通过 rest_do_request() 读取
 * content.rendered / title.rendered，结果写入 golden.json。
 */

require '/wordpress/wp-load.php';

$content_corpus = array(
	'empty'                => '',
	'whitespace'           => "  \n\n  ",
	'single_line'          => 'Hello world',
	'chinese_paragraphs'   => "这是第一段，包含中文标点。\n\n这是第二段。\n第二段的第二行。",
	'crlf'                 => "第一段\r\n\r\n第二段\r\n第二行",
	'multi_br'             => "line one<br><br>line two<br/>\nline three<br />\nline four",
	'triple_newlines'      => "a\n\n\n\nb\n\n\n",
	'quotes_basic'         => "\"Hello,\" she said. It's John's book. 'single quoted' text.",
	'quotes_numbers'       => "The '90s were great. He was 6'2\" tall. Quoted \"42\" and '0.42' here. Year '99.",
	'quotes_cockney'       => "'tis the season 'cause we rock 'n' roll 'til 'em drop.",
	'quotes_nested'        => "He said \"she said 'hello' to me\" yesterday.",
	'quotes_primes'        => "It's 5' long and 3\" wide. 'Twas 12'. \"Quote 9\" ends\" here.",
	'dashes'               => "a -- b --- c - d xn--fsqu00a.xn--3lr804guic\nfoo--bar\n-- start",
	'ellipsis_tm'          => "Wait... really (tm) ``quoted'' text",
	'times'                => "A 9x9 grid, 1920x1080 screen, 0x1234 hex, 0.5x2 scale, 中9x9中.",
	'ampersands'           => "Tom & Jerry &amp; friends &copy; &#169; &#x1F600; &notanentityatall; AT&T",
	'pre_code'             => "Before \"quote\"\n\n<pre>\nif (a -- b) { \"x\" }\n\nkeep...\n</pre>\n\nAfter 'quote'\n\n<code>\"no\" -- change</code> and \"yes\"",
	'pre_attrs'            => "<pre class=\"brush: php\">\n\$a = \"b\";\n\n\$c = 'd';\n</pre>\n\nText",
	'pre_code_lang'        => "<pre><code class=\"language-js\">const a = \"b\"; // it's\n\nconsole.log(a)</code></pre>",
	'script_style'         => "<script>\nvar a = \"b\" -- 1;\n\nvar c = 'd';\n</script>\n\n<style>\na::before { content: \"x\"; }\n\n</style>\n\nAfter \"x\"",
	'kbd_tt'               => "Press <kbd>\"Ctrl\"--C</kbd> and <tt>'x'</tt> then \"y\".",
	'more_tag'             => "Intro paragraph.\n\n<!--more-->\n\nRest of the post.",
	'more_inline'          => "Intro paragraph.<!--more-->Rest of the post.",
	'html_comment'         => "Text <!-- a \"comment\" -- here --> after.\n\n<!-- standalone -->\n\nNext.",
	'block_mix'            => "<div class=\"box\">Inside div\nwith newline</div>\nText after div\n\n<h2>Heading \"two\"</h2>\nSome text\n<hr>\nAfter hr",
	'div_unclosed_p'       => "<div>Para in div</div>\n\n<address>Addr</address>",
	'lists'                => "<ul>\n<li>One \"1\"</li>\n<li>Two</li>\n</ul>\n\n<ol><li>A</li><li>B</li></ol>",
	'blockquote'           => "<blockquote>Quoted \"text\"\n\nSecond para</blockquote>\n\nAfter",
	'blockquote_attr'      => "<blockquote class=\"x\" cite='y'>Line one\nLine two</blockquote>",
	'table'                => "<table>\n<tr>\n<td>Cell \"1\"</td>\n<td>Cell 2</td>\n</tr>\n</table>",
	'inline_tags'          => "Some <strong>bold \"text\"</strong> and <em>it's</em> <a href=\"http://example.com/?a=1&b=2\">link's</a>.",
	'attr_newline'         => "<a href=\"http://example.com\"\ntitle=\"x\">link</a> text\n\nmore",
	'option_select'        => "<select>\n<option>1</option>\n<option>2</option>\n</select>",
	'object_embed'         => "<object width=\"1\">\n<param name=\"a\" value=\"b\">\n<embed src=\"x.swf\">\n</object>",
	'figure_figcaption'    => "<figure>\n<img src=\"http://example.com/a.jpg\" alt=\"a\">\n<figcaption>Cap \"x\"</figcaption>\n</figure>",
	'svg_math'             => "<svg>\n<path d=\"M0 0\"/>\n</svg>\n\ntext",
	'img_plain'            => "<img src=\"http://example.com/a.jpg\" alt=\"x\">",
	'img_dims'             => "<img src=\"http://example.com/b.jpg\" width=\"300\" height=\"200\" alt=\"y\">",
	'img_dims_class'       => "<img class=\"alignnone size-full wp-image-12345\" src=\"http://example.com/c.jpg\" width=\"300\" height=\"200\" alt=\"\" />",
	'img_loading_given'    => "<img src=\"http://example.com/d.jpg\" width=\"300\" height=\"200\" loading=\"lazy\" decoding=\"sync\">",
	'img_percent'          => "<img src=\"http://example.com/e.jpg\" width=\"100%\" height=\"200\">",
	'img_no_src'           => "<img data-src=\"x.jpg\" width=\"10\" height=\"10\">",
	'img_dup'              => "<img src=\"http://example.com/f.jpg\" width=\"1\" height=\"1\"> and <img src=\"http://example.com/f.jpg\" width=\"1\" height=\"1\">",
	'img_fetch_low'        => "<img src=\"http://example.com/g.jpg\" width=\"300\" height=\"200\" fetchpriority=\"low\">",
	'iframe'               => "<iframe src=\"https://example.com/embed\" width=\"560\" height=\"315\"></iframe>\n\n<iframe src=\"https://example.com/x\"></iframe>",
	'caption_old'          => "[caption id=\"attachment_101\" align=\"alignleft\" width=\"300\" caption=\"An old \"caption\"\"]<img src=\"http://example.com/h.jpg\" width=\"300\" height=\"200\" />[/caption]",
	'caption_new'          => "[caption id=\"attachment_102\" align=\"aligncenter\" width=\"640\"]<a href=\"http://example.com/full.jpg\"><img class=\"size-large wp-image-102\" src=\"http://example.com/i.jpg\" alt=\"\" width=\"640\" height=\"480\" /></a> The caption's text[/caption]\n\nAfter caption.",
	'caption_no_width'     => "[caption id=\"attachment_103\"]<img src=\"http://example.com/j.jpg\" /> no width[/caption]",
	'caption_in_para'      => "Text before\n[caption id=\"attachment_104\" width=\"100\"]<img src=\"http://example.com/k.jpg\" /> cap[/caption]\nText after",
	'caption_class_wp'     => "[wp_caption id=\"attachment_105\" align=\"alignright\" width=\"200\" class=\"extra\" caption_id=\"my_cap\"]<img src=\"http://example.com/l.jpg\" /> cap[/wp_caption]",
	'caption_no_id'        => "[caption width=\"50\" caption=\"plain\"]<img src=\"http://example.com/m.jpg\" />[/caption]",
	'shortcode_unknown'    => "[foo bar=\"baz\"]content \"quoted\"[/foo]\n\n[unknown]",
	'shortcode_escaped'    => "[[caption]] and [[foo]] and [[caption width=\"1\"]x[/caption]]",
	'shortcode_standalone' => "Para one.\n\n[caption id=\"attachment_106\" width=\"10\"]<img src=\"http://example.com/n.jpg\" /> x[/caption]\n\nPara two.",
	'brackets_in_html'     => "<a href=\"http://example.com/[x]\" title=\"[caption]\">link</a> &#91;not&#93; [b]",
	'embed_shortcode'      => "[embed]http://localhost/no-oembed-here[/embed]",
	'smilies'              => "Happy :) sad :( wink ;) grin :-D cool 8-) green :mrgreen: :???: x:)y\n\n<code>:)</code> and <pre>:-)</pre> :!:",
	'smilies_edge'         => ":) at start and end :)\n:D\n(:) not",
	'wordpress_spelling'   => "I love Wordpress. (Wordpress) \"Wordpress\" 'Wordpress' >Wordpress Wordpress.",
	'autoembed_amp_para'   => "<blockquote>\n\t<p>\n\t\thttp://www.example.invalid/func/req?lc={语言，中文为zh}&amp;msg={utf-8编码的请求文字}</p>\n</blockquote>",
	'autoembed_amp_line'   => "Text\nhttp://www.example.invalid/a?x=1&amp;y=2\nMore &amp; more",
	'autoembed_amp_pre'    => "<pre>\nhttp://www.example.invalid/a?x=1&amp;msg=2\nfoo &amp; bar &#39;q&#39; it's\n</pre>\n\n<pre class=\"brush: php\">\$a = 'b'; // &amp;msg</pre>",
	'autoembed_tag_newline' => "<a\nhref=\"http://www.example.invalid/x?a=1&amp;b=2\">\nhttp://www.example.invalid/y?a=1&amp;b=2\n</a>",
	'url_alone'            => "Look:\n\nhttp://localhost/some/path\n\nEnd.",
	'gutenberg_paragraph'  => "<!-- wp:paragraph -->\n<p>Block \"paragraph\" it's -- here</p>\n<!-- /wp:paragraph -->\n\n<!-- wp:paragraph {\"align\":\"center\"} -->\n<p class=\"has-text-align-center\">Centered</p>\n<!-- /wp:paragraph -->",
	'gutenberg_heading'    => "<!-- wp:heading -->\n<h2>Heading</h2>\n<!-- /wp:heading -->\n\n<!-- wp:heading {\"level\":3} -->\n<h3 class=\"wp-block-heading\" id=\"x\">Third</h3>\n<!-- /wp:heading -->",
	'gutenberg_list'       => "<!-- wp:list -->\n<ul><!-- wp:list-item -->\n<li>One</li>\n<!-- /wp:list-item -->\n\n<!-- wp:list-item -->\n<li>Two</li>\n<!-- /wp:list-item --></ul>\n<!-- /wp:list -->",
	'gutenberg_code'       => "<!-- wp:code -->\n<pre class=\"wp-block-code\"><code>const a = \"b\";\n\nreturn a -- 1;</code></pre>\n<!-- /wp:code -->",
	'gutenberg_quote'      => "<!-- wp:quote -->\n<blockquote class=\"wp-block-quote\"><!-- wp:paragraph -->\n<p>Quote text</p>\n<!-- /wp:paragraph --><cite>Someone</cite></blockquote>\n<!-- /wp:quote -->",
	'gutenberg_image'      => "<!-- wp:image {\"id\":55,\"sizeSlug\":\"large\"} -->\n<figure class=\"wp-block-image size-large\"><img src=\"http://example.com/o.jpg\" alt=\"\" class=\"wp-image-55\"/><figcaption class=\"wp-element-caption\">Img caption</figcaption></figure>\n<!-- /wp:image -->",
	'gutenberg_group'      => "<!-- wp:group -->\n<div class=\"wp-block-group\"><!-- wp:paragraph -->\n<p>In group</p>\n<!-- /wp:paragraph --></div>\n<!-- /wp:group -->",
	'gutenberg_separator'  => "<!-- wp:separator -->\n<hr class=\"wp-block-separator has-alpha-channel-opacity\"/>\n<!-- /wp:separator -->",
	'gutenberg_html'       => "<!-- wp:html -->\n<div>Raw \"html\"\n\nhere</div>\n<!-- /wp:html -->",
	'gutenberg_shortcode'  => "<!-- wp:shortcode -->\n[caption id=\"attachment_107\" width=\"10\"]<img src=\"http://example.com/p.jpg\" /> c[/caption]\n<!-- /wp:shortcode -->",
	'gutenberg_freeform'   => "Some classic text\n\n<!-- wp:paragraph -->\n<p>Block</p>\n<!-- /wp:paragraph -->\n\nTrailing classic \"text\"",
	'gutenberg_void'       => "<!-- wp:spacer {\"height\":\"20px\"} /-->\n<!-- wp:paragraph --><p>x</p><!-- /wp:paragraph -->",
	'gutenberg_unknown'    => "<!-- wp:myplugin/thing {\"a\":1} -->\n<div class=\"thing\">Plugin block</div>\n<!-- /wp:myplugin/thing -->",
	'gutenberg_preformat'  => "<!-- wp:preformatted -->\n<pre class=\"wp-block-preformatted\">Pre \"formatted\"\n\ntext</pre>\n<!-- /wp:preformatted -->",
	'gutenberg_table'      => "<!-- wp:table -->\n<figure class=\"wp-block-table\"><table><tbody><tr><td>a</td><td>b</td></tr></tbody></table></figure>\n<!-- /wp:table -->",
	'gutenberg_columns'    => "<!-- wp:columns -->\n<div class=\"wp-block-columns\"><!-- wp:column -->\n<div class=\"wp-block-column\"><!-- wp:paragraph -->\n<p>Col 1</p>\n<!-- /wp:paragraph --></div>\n<!-- /wp:column --></div>\n<!-- /wp:columns -->",
	'long_chinese'         => "## 不是 Markdown\n\n“中文引号”和「直角引号」，以及 \"英文引号\"。\n\n第一行\n第二行\n\n<p>已有 p 标签</p>\n\n结尾。",
	'nbsp_spaces'          => "a&nbsp;-&nbsp;b and \"x\"&nbsp;y",
	'stray_lt'             => "a < b and c > d, x <3 y",
	'p_already'            => "<p>One</p>\n<p>Two</p>",
	'br_before_block'      => "text<br />\n<div>x</div>",
	'audio_video_source'   => "<video controls>\n<source src=\"a.mp4\">\n<track src=\"a.vtt\">\n</video>",
	'details_summary'      => "<details>\n<summary>Sum</summary>\nBody\n</details>",
	'textarea'             => "<textarea>\n:)\n\nline\n</textarea>",
);

$title_corpus = array(
	'plain'       => 'Hello World',
	'quotes'      => '"Quoted" title\'s here -- yes...',
	'chinese'     => '中文标题：“测试”',
	'amp'         => 'Tom & Jerry &amp; Co &#169; AT&T',
	'wordpress'   => 'Why Wordpress is great',
	'html'        => 'Title with <em>"em"</em> tag',
	'spaces'      => '  padded title  ',
	'times'       => '1920x1080 wallpapers',
	'numbers'     => "The '90s and 6'2\"",
);

$sanitize_corpus = array(
	'中文标签',
	'Hello World',
	'  Mixed CASE Slug  ',
	'already-a-slug',
	'%e4%b8%ad%e6%96%87',
	'%E4%B8%AD',
	'100% sure',
	'Café Crème brûlée',
	'Ångström Ærø Øre straße',
	'a.b.c',
	'a/b/c',
	'tag_with_underscore',
	'emoji 🎉 party',
	'quote "x" and \'y\'',
	'em—dash en–dash',
	'<b>bold</b> tag',
	'&amp; entity &nbsp; space',
	'中文 English 混合',
	'---dashes---',
	'C++ & C#',
	'ÀÉÎÕÜ àéîõü',
	'x × y',
);

$comment_corpus = array(
	'plain'          => 'Just a comment.',
	'bare_url'       => 'See http://example.com/path?a=1&b=2 for details.',
	'https_url'      => "Visit https://example.org/foo_(bar) now.\nAnd (https://example.org/x) too.",
	'url_trailing'   => 'Check http://example.com/a.jpg, http://example.com/b. and http://example.com/c!',
	'www_url'        => 'Go to www.example.com/page. Or ftp.example.com.',
	'email'          => 'Mail me at foo.bar+baz@example.co.uk please.',
	'internal_url'   => 'Internal: ' . home_url( '/hello/' ) . ' link.',
	'existing_link'  => 'An <a href="http://example.com">existing http://example.com</a> link.',
	'link_in_code'   => '<code>http://example.com/code</code> and <pre>www.example.com</pre>',
	'unclosed'       => '<strong>bold <em>both</strong> text <b>open',
	'stray_close'    => 'text</div> more</em>',
	'self_closing'   => 'line<br>break<br/>and<img src="x.jpg">',
	'nested_same'    => '<p>one<p>two',
	'love'           => 'I <3 you and 1<2',
	'comment_tag'    => 'a < !-- b --> c',
	'multi_para'     => "第一段评论。\n\n第二段评论，有 \"引号\" 和 -- 破折号。\n第三行",
	'chinese_url'    => '链接 http://example.com/中文路径 后面',
	'chinese_url2'   => '看这里：http://example.com/abc 。',
	'smilies'        => 'Nice :) and :D',
	'wordpress'      => 'I use Wordpress daily.',
	'custom_element' => '<my-element foo="bar">x</my-element> <x-y>z',
	'scheme_unknown' => 'odd foo://bar.baz thing',
	'quote_in_url'   => "it's http://example.com/it's here",
	'long_para'      => "Line one\nLine two\n\n<blockquote>Quoted\nreply</blockquote>\nafter",
);

// 渲染过含区块的内容之后再渲染的语料：wpautop 会被挪到优先级 10 的队尾（AutopAfterUnautop）。
$after_blocks_corpus = array(
	'caption_standalone'  => "Para one.\n\n[caption id=\"attachment_201\" width=\"10\"]<img src=\"http://example.com/q.jpg\" /> x[/caption]\n\nPara two.",
	'caption_only'        => "[caption id=\"attachment_202\" align=\"alignleft\" width=\"20\"]<img src=\"http://example.com/r.jpg\" /> y[/caption]",
	'caption_in_para'     => "Before\n[caption id=\"attachment_203\" width=\"30\"]<img src=\"http://example.com/s.jpg\" /> z[/caption]\nAfter",
	'unknown_standalone'  => "Text.\n\n[foo bar=\"baz\"]\n\n[unknown]content[/unknown]",
	'escaped_standalone'  => "[[caption]]\n\nNext",
	'embed_standalone'    => "Look\n\n[embed]http://localhost/no-oembed-there[/embed]\n\nEnd",
	'plain_paragraphs'    => "第一段 \"quote\"\n\n第二段\n第二行",
	'blocks_again'        => "<!-- wp:paragraph -->\n<p>Block</p>\n<!-- /wp:paragraph -->",
);
$after_blocks_trigger = "<!-- wp:paragraph -->\n<p>Trigger</p>\n<!-- /wp:paragraph -->";

kses_remove_filters();
wp_set_current_user( 1 );

$items = array();
foreach ( $content_corpus as $name => $content ) {
	$items[ $name ] = wp_insert_post(
		wp_slash(
			array(
				'post_title'   => 'content ' . $name,
				'post_content' => $content,
				'post_status'  => 'publish',
			)
		),
		true
	);
}
$after_blocks = array();
foreach ( $after_blocks_corpus as $name => $content ) {
	$after_blocks[ $name ] = wp_insert_post(
		wp_slash(
			array(
				'post_title'   => 'after blocks ' . $name,
				'post_content' => $content,
				'post_status'  => 'publish',
			)
		),
		true
	);
}
$after_blocks_trigger_id = wp_insert_post(
	wp_slash(
		array(
			'post_title'   => 'after blocks trigger',
			'post_content' => $after_blocks_trigger,
			'post_status'  => 'publish',
		)
	),
	true
);
$titles = array();
foreach ( $title_corpus as $name => $title ) {
	$titles[ $name ] = wp_insert_post(
		wp_slash(
			array(
				'post_title'   => $title,
				'post_content' => 'x',
				'post_status'  => 'publish',
			)
		),
		true
	);
}

$comments_post = wp_insert_post(
	array(
		'post_title'   => 'comments holder',
		'post_content' => 'x',
		'post_status'  => 'publish',
	)
);
$comments = array();
foreach ( $comment_corpus as $name => $text ) {
	$comments[ $name ] = wp_insert_comment(
		array(
			'comment_post_ID'      => $comments_post,
			'comment_author'       => 'tester',
			'comment_author_email' => 'tester@example.com',
			'comment_content'      => $text,
			'comment_approved'     => 1,
		)
	);
}

wp_set_current_user( 0 );

/**
 * 按 default-filters.php 的顺序恢复 the_content 优先级 10 的回调。
 *
 * do_blocks() 处理含区块的内容后会通过 _restore_wpautop_hook() 把 wpautop 挂到
 * 优先级 10 的队尾，这里撤销该状态，避免用例之间互相影响；同时重置请求级的静态标记。
 */
function i3w_reset_content_filters() {
	$filters = array( 'wptexturize', 'wpautop', 'shortcode_unautop', 'prepend_attachment', 'wp_replace_insecure_home_url' );
	foreach ( $filters as $filter ) {
		remove_filter( 'the_content', $filter, 10 );
	}
	remove_filter( 'the_content', '_restore_wpautop_hook', 11 );
	foreach ( $filters as $filter ) {
		add_filter( 'the_content', $filter );
	}
	// 每个 REST 请求都是新的 PHP 进程，fetchpriority="high" 的标记应当是初始值。
	wp_high_priority_element_flag( true );
}

function i3w_rest_post( $id, $reset = true ) {
	if ( $reset ) {
		i3w_reset_content_filters();
	}
	$response = rest_do_request( new WP_REST_Request( 'GET', '/wp/v2/posts/' . $id ) );
	return $response->get_data();
}

$out = array(
	'wp_version'    => get_bloginfo( 'version' ),
	'home'          => home_url(),
	'siteurl'       => site_url(),
	'html5_caption' => current_theme_supports( 'html5', 'caption' ),
	'use_smilies'   => (bool) get_option( 'use_smilies' ),
	'content'       => array(),
	'content_after_blocks' => array(),
	'title'         => array(),
	'sanitize'      => array(),
	'comment'       => array(),
);

// ---------------------------------------------------------------------------
// 附件：只写入元数据，不需要真实文件。
// ---------------------------------------------------------------------------
wp_set_current_user( 1 );
$uploads_baseurl = wp_get_upload_dir()['baseurl'];
function i3w_attachment( $file, $mime, $meta ) {
	$id = wp_insert_attachment(
		array(
			'post_mime_type' => $mime,
			'post_title'     => basename( $file ),
			'post_status'    => 'inherit',
		),
		$file
	);
	wp_update_attachment_metadata( $id, $meta );
	return $id;
}
function i3w_size( $file, $w, $h, $mime = 'image/jpeg' ) {
	return array( 'file' => $file, 'width' => $w, 'height' => $h, 'mime-type' => $mime );
}
$att_a = i3w_attachment(
	'2024/05/router-id.jpg',
	'image/jpeg',
	array(
		'width'  => 2560,
		'height' => 1020,
		'file'   => '2024/05/router-id.jpg',
		'sizes'  => array(
			'thumbnail'    => i3w_size( 'router-id-150x150.jpg', 150, 150 ),
			'medium'       => i3w_size( 'router-id-300x120.jpg', 300, 120 ),
			'medium_large' => i3w_size( 'router-id-768x306.jpg', 768, 306 ),
			'large'        => i3w_size( 'router-id-1024x408.jpg', 1024, 408 ),
			'1536x1536'    => i3w_size( 'router-id-1536x612.jpg', 1536, 612 ),
			'2048x2048'    => i3w_size( 'router-id-2048x816.jpg', 2048, 816 ),
			'odd'          => i3w_size( 'router-id-1237x493.jpg', 1237, 493 ),
		),
	)
);
$att_gif = i3w_attachment(
	'2024/06/anim.gif',
	'image/gif',
	array(
		'width'  => 600,
		'height' => 400,
		'file'   => '2024/06/anim.gif',
		'sizes'  => array(
			'thumbnail' => i3w_size( 'anim-150x150.gif', 150, 150, 'image/gif' ),
			'medium'    => i3w_size( 'anim-300x200.gif', 300, 200, 'image/gif' ),
		),
	)
);
$att_edit = i3w_attachment(
	'2024/07/photo-e1600000000000.jpg',
	'image/jpeg',
	array(
		'width'  => 800,
		'height' => 600,
		'file'   => '2024/07/photo-e1600000000000.jpg',
		'sizes'  => array(
			'medium' => i3w_size( 'photo-e1600000000000-300x225.jpg', 300, 225 ),
			'old'    => i3w_size( 'photo-400x300.jpg', 400, 300 ),
		),
	)
);
$att_old = i3w_attachment(
	'wp-content/uploads/2009/01/legacy.jpg',
	'image/jpeg',
	array(
		'width'  => 500,
		'height' => 250,
		'file'   => 'wp-content/uploads/2009/01/legacy.jpg',
		'sizes'  => array(
			'medium' => i3w_size( 'legacy-300x150.jpg', 300, 150 ),
		),
	)
);
$att_ids = array( $att_a, $att_gif, $att_edit, $att_old );
wp_set_current_user( 0 );

$up = $uploads_baseurl;
$attachment_corpus = array(
	'att_large'           => "<img src=\"$up/2024/05/router-id-1024x408.jpg\" alt=\"\" width=\"640\" height=\"255\" class=\"alignnone size-large wp-image-$att_a\" />",
	'att_no_dims'         => "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/router-id-300x120.jpg\" alt=\"\">",
	'att_full'            => "<img class=\"size-full wp-image-$att_a\" src=\"$up/2024/05/router-id.jpg\" alt=\"\">",
	'att_style_width'     => "<img class=\"wp-image-$att_a\" style=\"width: 200px;\" src=\"$up/2024/05/router-id-768x306.jpg\" alt=\"\">",
	'att_mismatch'        => "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/other.jpg\" width=\"300\" height=\"120\" alt=\"\">",
	'att_mismatch_nodims' => "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/other.jpg\" alt=\"\">",
	'att_has_srcset'      => "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/router-id-300x120.jpg\" width=\"300\" height=\"120\" srcset=\"a.jpg 1x\" alt=\"\">",
	'att_has_sizes'       => "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/router-id-300x120.jpg\" width=\"300\" height=\"120\" sizes=\"100px\" alt=\"\">",
	'att_missing'         => "<img class=\"wp-image-999999\" src=\"$up/2024/05/router-id-300x120.jpg\" width=\"300\" height=\"120\" alt=\"\">",
	'att_huge_src'        => "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/router-id.jpg\" width=\"2560\" height=\"1020\" alt=\"\">",
	'att_ratio_odd'       => "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/router-id-1237x493.jpg\" width=\"1237\" height=\"493\" alt=\"\">",
	'att_query'           => "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/router-id-768x306.jpg?ver=1\" alt=\"\">",
	'att_eager'           => "<img loading=\"eager\" class=\"wp-image-$att_a\" src=\"$up/2024/05/router-id-768x306.jpg\" width=\"768\" height=\"306\" alt=\"\">",
	'att_gif_full'        => "<img class=\"wp-image-$att_gif\" src=\"$up/2024/06/anim.gif\" width=\"600\" height=\"400\" alt=\"\">",
	'att_gif_medium'      => "<img class=\"wp-image-$att_gif\" src=\"$up/2024/06/anim-300x200.gif\" width=\"300\" height=\"200\" alt=\"\">",
	'att_edited_ok'       => "<img class=\"wp-image-$att_edit\" src=\"$up/2024/07/photo-e1600000000000-300x225.jpg\" width=\"300\" height=\"225\" alt=\"\">",
	'att_edited_old'      => "<img class=\"wp-image-$att_edit\" src=\"$up/2024/07/photo-400x300.jpg\" width=\"400\" height=\"300\" alt=\"\">",
	'att_legacy_path'     => "<img class=\"wp-image-$att_old\" src=\"$up/2009/01/legacy-300x150.jpg\" width=\"300\" height=\"150\" alt=\"\">",
	'att_caption'         => "[caption id=\"attachment_$att_a\" align=\"alignnone\" width=\"300\"]<img class=\"size-medium wp-image-$att_a\" src=\"$up/2024/05/router-id-300x120.jpg\" alt=\"\" width=\"300\" height=\"120\" /> Router[/caption]",
	'att_block'           => "<!-- wp:image {\"id\":$att_a,\"sizeSlug\":\"large\"} -->\n<figure class=\"wp-block-image size-large\"><img src=\"$up/2024/05/router-id-1024x408.jpg\" alt=\"\" class=\"wp-image-$att_a\"/></figure>\n<!-- /wp:image -->",
	'att_upper_class'     => "<img class=\"WP-IMAGE-$att_a\" src=\"$up/2024/05/router-id-300x120.jpg\" alt=\"\">",
);

// ---------------------------------------------------------------------------
// RSS2 feed：必须在任何 REST 渲染之前执行，确保 WordPress 内部的静态状态
// （媒体计数、fetchpriority 标记、wp_unique_prefixed_id 计数、wpautop 顺序）是全新的。
// ---------------------------------------------------------------------------
$feed_corpus = array(
	'images_first'   => array( 'Tom & Jerry <em>"loves"</em> it\'s', "<p>Intro</p>\n<img src=\"http://example.com/f0.jpg\" width=\"10\" height=\"10\">\n<img src=\"http://example.com/f1.jpg\" width=\"600\" height=\"400\" alt=\"\">\n<img src=\"http://example.com/f2.jpg\" width=\"600\" height=\"400\">\n<img src=\"http://example.com/f3.jpg\" width=\"10\" height=\"10\">" ),
	'images_more'    => array( 'More images', "<iframe src=\"https://example.com/v\" width=\"560\" height=\"315\"></iframe>\n<img src=\"http://example.com/f4.jpg\" width=\"10\" height=\"10\">\n<img src=\"http://example.com/f5.jpg\" width=\"600\" height=\"400\">\n<img src=\"http://example.com/f6.jpg\">\n<img src=\"http://example.com/f7.jpg\" width=\"600\" height=\"400\" loading=\"eager\">" ),
	'emoji'          => array( 'Emoji 😀 title', "Smile :) and 😀 and 👍🏽 and ©️ text ❤️\n\n<code>😀</code>\n\n<img src=\"http://example.com/e.png\" alt=\"😀\">" ),
	'more'           => array( 'More', "Teaser text\n\n<!--more-->\n\nAfter more" ),
	'more_custom'    => array( 'More custom', 'Teaser<!--more Read on-->Rest' ),
	'nextpage'       => array( 'Next page', "Page one\n<!--nextpage-->\nPage two\n\n<!--nextpage-->Page three" ),
	'noteaser'       => array( 'No teaser', "Teaser\n<!--more-->\n<!--noteaser-->\nBody" ),
	'caption_before' => array( 'Caption before', "Before\n\n[caption id=\"attachment_301\" width=\"100\"]<img src=\"http://example.com/c1.jpg\" width=\"100\" height=\"50\" /> cap one[/caption]\n\nAfter" ),
	'blocks'         => array( 'Blocks', "<!-- wp:paragraph -->\n<p>Block :)</p>\n<!-- /wp:paragraph -->\n\n<!-- wp:image -->\n<figure class=\"wp-block-image\"><img src=\"http://example.com/b.jpg\" width=\"600\" height=\"400\" alt=\"\"/></figure>\n<!-- /wp:image -->" ),
	'caption_after'  => array( 'Caption after', "[caption id=\"attachment_301\" width=\"100\"]<img src=\"http://example.com/c2.jpg\" /> cap two[/caption]\n\nTail" ),
	'block_more'     => array( 'Block more', "<!-- wp:paragraph -->\n<p>Lead</p>\n<!-- /wp:paragraph -->\n\n<!-- wp:more -->\n<!--more-->\n<!-- /wp:more -->\n\n<!-- wp:paragraph -->\n<p>Rest</p>\n<!-- /wp:paragraph -->" ),
	'cdata'          => array( 'CDATA ]]> & more', "Code: <code>a]]>b</code> end ]]>" ),
	'attachments'    => array( 'Attachments', "<img class=\"wp-image-$att_a\" src=\"$up/2024/05/router-id-300x120.jpg\" alt=\"\">\n<img src=\"$up/2024/05/router-id-1024x408.jpg\" alt=\"\" width=\"640\" height=\"255\" class=\"alignnone size-large wp-image-$att_a\" />" ),
	'embedded'       => array( 'Embedded', "<iframe class=\"wp-embedded-content\" style=\"position: absolute;\" src=\"https://example.com/embed/\" width=\"600\" height=\"400\"></iframe>" ),
);

wp_set_current_user( 1 );
update_option( 'blogdescription', 'A & B <b>bold</b> "quoted" it\'s © x|y' );
$feed_cat = wp_insert_term( 'Feed & Test', 'category', array( 'slug' => 'feed-test' ) );
$feed_ids = array();
$feed_i   = 0;
foreach ( $feed_corpus as $name => $item ) {
	$id = wp_insert_post(
		wp_slash(
			array(
				'post_title'    => $item[0],
				'post_content'  => $item[1],
				'post_status'   => 'publish',
				'post_date'     => gmdate( 'Y-m-d H:i:s', strtotime( '2020-01-01' ) + ( count( $feed_corpus ) - $feed_i ) * 86400 ),
				'post_category' => array( $feed_cat['term_id'] ),
				'tags_input'    => array( 'C&D "tag"', "it's <b>x</b>" ),
			)
		),
		true
	);
	$feed_ids[ $id ] = $name;
	++$feed_i;
}
wp_set_current_user( 0 );
update_option( 'posts_per_rss', 100 );

i3w_reset_content_filters();
ob_start();
wp( array( 'feed' => 'rss2', 'category_name' => 'feed-test' ) );
load_template( ABSPATH . WPINC . '/feed-rss2.php', false );
$feed_xml = ob_get_clean();

$feed_guids = array();
foreach ( $feed_ids as $id => $name ) {
	$feed_guids[ esc_url( get_the_guid( $id ) ) ] = $id;
}

$out['feed'] = array(
	'description_input' => get_option( 'blogdescription' ),
	'description'       => preg_match( '#<channel>.*?<description>(.*?)</description>#s', $feed_xml, $m ) ? $m[1] : null,
	'items'             => array(),
);
preg_match_all( '#<item>(.*?)</item>#s', $feed_xml, $feed_items );
foreach ( $feed_items[1] as $item_xml ) {
	preg_match( '#<guid[^>]*>(.*?)</guid>#', $item_xml, $m );
	$id   = $feed_guids[ $m[1] ];
	$post = get_post( $id );
	preg_match( '#<title>(.*?)</title>#s', $item_xml, $title );
	preg_match( '#<content:encoded><!\[CDATA\[(.*?)\]\]></content:encoded>#s', $item_xml, $encoded );
	preg_match_all( '#<category><!\[CDATA\[(.*?)\]\]></category>#s', $item_xml, $cats );
	$term_names = array();
	foreach ( array_merge( get_the_category( $id ), (array) get_the_tags( $id ) ) as $term ) {
		$term_names[] = $term->name;
	}
	$out['feed']['items'][] = array(
		'name'                => $feed_ids[ $id ],
		'id'                  => $id,
		'input'               => $post->post_content,
		'expected'            => $encoded[1],
		'title_input'         => $post->post_title,
		'title'               => $title[1],
		'categories_input'    => $term_names,
		'categories'          => $cats[1],
	);
}

// 附件元数据原文（serialize 格式），供 Go 端构造 Attachments。
$out['uploads_baseurl'] = $uploads_baseurl;
$out['attachments_meta'] = array();
foreach ( $att_ids as $att_id ) {
	$out['attachments_meta'][ (string) $att_id ] = $wpdb->get_var(
		$wpdb->prepare( "SELECT meta_value FROM $wpdb->postmeta WHERE post_id = %d AND meta_key = '_wp_attachment_metadata'", $att_id )
	);
}

// 无状态的 feed 辅助函数。
$feed_strings = array(
	'Tom & Jerry &amp; Co &copy; &#169; &nbsp; AT&T',
	'<b>bold</b> "quoted" it\'s -- dash...',
	'中文 & “引号” <em>强调</em>',
	'a|b &lt;tag&gt; &quot;q&quot; &#039;s&#039; &apos;',
	'Wordpress & 1920x1080',
	'&eacute;t&eacute; &hearts; &unknown;',
	'"Wordpress" (Wordpress) >Wordpress Wordpress',
	'',
);
$out['feed_funcs'] = array();
foreach ( $feed_strings as $x ) {
	$pre = static function () use ( $x ) {
		return $x;
	};
	add_filter( 'pre_option_blogdescription', $pre );
	$bloginfo = apply_filters( 'bloginfo_rss', get_bloginfo_rss( 'description' ), 'description' );
	remove_filter( 'pre_option_blogdescription', $pre );
	$out['feed_funcs'][] = array(
		'input'     => $x,
		'ent2ncr'   => ent2ncr( $x ),
		'title_rss' => apply_filters( 'the_title_rss', apply_filters( 'the_title', $x, 0 ) ),
		'bloginfo'  => $bloginfo,
		'document_title' => apply_filters( 'document_title', $x ),
		'term_name' => html_entity_decode( sanitize_term_field( 'name', $x, 0, 'category', 'rss' ), ENT_COMPAT, 'UTF-8' ),
	);
}
$out['esc_url'] = array();
foreach ( array( 'http://example.com/a b?x=1&y=2', "http://example.com/it's", 'example.com/path', 'javascript:alert(1)', '/relative/[x]', 'http://example.com/中文', 'mailto:a@b.c', 'http://example.com/a%0d%0ab' ) as $x ) {
	$out['esc_url'][] = array(
		'input'    => $x,
		'expected' => esc_url( $x ),
	);
}

$out['attachments'] = array();
wp_set_current_user( 1 );
$attachment_posts = array();
foreach ( $attachment_corpus as $name => $content ) {
	$attachment_posts[ $name ] = wp_insert_post(
		wp_slash(
			array(
				'post_title'   => 'attachment ' . $name,
				'post_content' => $content,
				'post_status'  => 'publish',
			)
		)
	);
}
wp_set_current_user( 0 );
foreach ( $attachment_posts as $name => $id ) {
	$data                  = i3w_rest_post( $id );
	$out['attachments'][] = array(
		'name'     => $name,
		'input'    => get_post( $id )->post_content,
		'expected' => $data['content']['rendered'],
	);
}

foreach ( $items as $name => $id ) {
	if ( is_wp_error( $id ) ) {
		fwrite( STDERR, "insert $name failed: " . $id->get_error_message() . "\n" );
		continue;
	}
	$data              = i3w_rest_post( $id );
	$out['content'][] = array(
		'name'     => $name,
		'input'    => get_post( $id )->post_content,
		'expected' => $data['content']['rendered'],
	);
}

// 先渲染一篇含区块的文章，之后不再恢复过滤器顺序。
i3w_rest_post( $after_blocks_trigger_id );
foreach ( $after_blocks as $name => $id ) {
	$data                          = i3w_rest_post( $id, false );
	$out['content_after_blocks'][] = array(
		'name'     => $name,
		'input'    => get_post( $id )->post_content,
		'expected' => $data['content']['rendered'],
	);
}

foreach ( $titles as $name => $id ) {
	$data            = i3w_rest_post( $id );
	$out['title'][] = array(
		'name'     => $name,
		'input'    => get_post( $id )->post_title,
		'expected' => $data['title']['rendered'],
	);
}

foreach ( $comments as $name => $id ) {
	$response          = rest_do_request( new WP_REST_Request( 'GET', '/wp/v2/comments/' . $id ) );
	$data              = $response->get_data();
	$out['comment'][] = array(
		'name'     => $name,
		'input'    => get_comment( $id )->comment_content,
		'expected' => $data['content']['rendered'] ?? null,
	);
}

foreach ( $sanitize_corpus as $input ) {
	$out['sanitize'][] = array(
		'input'    => $input,
		'expected' => sanitize_title( $input ),
	);
}

file_put_contents( '/work/golden.json', json_encode( $out, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES ) . "\n" );
echo 'ok ' . count( $out['content'] ) . ' content, ' . count( $out['content_after_blocks'] ) . ' after blocks, ' . count( $out['title'] ) . ' title, ' . count( $out['sanitize'] ) . ' sanitize, ' . count( $out['comment'] ) . ' comment, ' . count( $out['feed']['items'] ) . " feed items\n";
