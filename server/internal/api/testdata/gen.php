<?php
/**
 * 生成端到端测试的金标准数据。
 *
 * 在 WordPress Playground 中运行（每次运行都是全新站点）：
 *   cd server/internal/api/testdata && \
 *   mise exec node@24 -- npx -y @wp-playground/cli@latest php --mount=$PWD:/work -- /work/gen.php
 *
 * 输出：
 *   tables.json  种子数据写入后各数据表的内容，Go 测试会把它导入内存 MySQL；
 *   golden.json  对每个请求调用 rest_do_request 得到的状态码、分页头与响应体；
 *   feeds.json   不同选项组合下 /feed/ 的 RSS2 输出。
 */

require '/wordpress/wp-load.php';

global $wpdb;

update_option( 'blogname', '三言三语' );
update_option( 'blogdescription', 'best33.com | 希望和你交朋友！' );
update_option( 'timezone_string', 'Asia/Shanghai' );
update_option( 'gmt_offset', '' );
update_option( 'show_avatars', '1' );
update_option( 'avatar_default', 'mystery' );
update_option( 'avatar_rating', 'G' );

// 用户：作者、没有文章的用户。管理员（ID 1）由 Playground 创建。
$author = wp_insert_user(
	array(
		'user_login'   => 'zhangsan',
		'user_pass'    => 'x',
		'user_email'   => 'ZhangSan@Example.com ',
		'display_name' => '张三',
		'user_url'     => 'https://zhangsan.example.com',
		'role'         => 'author',
	)
);
$lurker = wp_insert_user(
	array(
		'user_login'   => 'lurker',
		'user_pass'    => 'x',
		'user_email'   => 'lurker@example.com',
		'display_name' => 'Lurker',
		'role'         => 'subscriber',
	)
);

// 分类与标签。
$cat_tech  = wp_insert_term( '技术', 'category', array( 'slug' => 'tech' ) )['term_id'];
$cat_life  = wp_insert_term( '生活', 'category' )['term_id'];
$cat_child = wp_insert_term( 'Go 语言', 'category', array( 'parent' => $cat_tech, 'slug' => 'golang' ) )['term_id'];
$cat_empty = wp_insert_term( '空分类', 'category', array( 'description' => '没有文章' ) )['term_id'];
$tags      = array();
foreach ( array( 'WordPress', '前端', 'Go', 'React', '随笔', '只用一次', 'A&B' ) as $name ) {
	$tags[ $name ] = wp_insert_term( $name, 'post_tag' )['term_id'];
}

$contents = array(
	"第一段，介绍一下“引号”和 \"英文引号\" -- 还有省略号...\n\n第二段\n单换行\n\n<!--more-->\n\n第三段",
	"<!-- wp:paragraph -->\n<p>区块编辑器写的段落。</p>\n<!-- /wp:paragraph -->\n\n<!-- wp:code -->\n<pre class=\"wp-block-code\"><code>func main() {\n\tfmt.Println(\"hi\")\n}</code></pre>\n<!-- /wp:code -->",
	"<pre class=\"brush: php\">&lt;?php echo 'hello'; ?&gt;</pre>\n\n普通段落 it's 1990's.",
	"[caption id=\"attachment_9\" align=\"alignnone\" width=\"300\"]<img class=\"size-medium wp-image-9\" src=\"https://best33.com/wp-content/uploads/a.jpg\" alt=\"\" width=\"300\" height=\"200\" /> 图片说明[/caption]\n\n文字 :) 结束",
	"<ul>\n<li>列表一</li>\n<li>列表二</li>\n</ul>\n<blockquote>引用一段话</blockquote>\n\n未注册的短代码 [foo bar=\"1\"] 保留",
	'Wordpress 拼写纠正，访问 http://example.com 看看。',
);

// 文章：日期递增，覆盖不同年份、月份；部分为其它状态。
$posts = array();
$specs = array(
	array( 'title' => 'Hello World', 'date' => '2019-03-01 10:00:00', 'cats' => array( $cat_tech ), 'tags' => array( 'WordPress', 'Go' ) ),
	array( 'title' => '“中文”标题 -- 测试', 'date' => '2019-12-31 23:30:00', 'cats' => array( $cat_life ), 'tags' => array( '随笔' ) ),
	array( 'title' => 'Go 与 React', 'date' => '2020-01-01 00:30:00', 'cats' => array( $cat_tech, $cat_child ), 'tags' => array( 'Go', 'React', '前端' ), 'author' => $author ),
	array( 'title' => '区块编辑器', 'date' => '2020-06-15 12:00:00', 'cats' => array( $cat_tech ), 'tags' => array( 'WordPress', '前端' ), 'author' => $author ),
	array( 'title' => '密码保护的文章', 'date' => '2020-07-01 08:00:00', 'cats' => array( $cat_life ), 'tags' => array( '随笔' ), 'password' => 'secret' ),
	array( 'title' => '草稿', 'date' => '2020-08-01 08:00:00', 'cats' => array( $cat_life ), 'tags' => array(), 'status' => 'draft' ),
	array( 'title' => '私密文章', 'date' => '2020-09-01 08:00:00', 'cats' => array( $cat_life ), 'tags' => array(), 'status' => 'private' ),
	array( 'title' => '置顶文章', 'date' => '2021-01-10 09:00:00', 'cats' => array( $cat_life ), 'tags' => array( '随笔', '只用一次' ), 'sticky' => true ),
	array( 'title' => 'A&B <em>强调</em>', 'date' => '2021-02-14 20:00:00', 'cats' => array( $cat_life ), 'tags' => array( 'A&B' ) ),
	array( 'title' => '搜索 关键词 测试', 'date' => '2021-05-05 05:05:05', 'cats' => array( $cat_tech ), 'tags' => array( 'Go' ) ),
	array( 'title' => '同一时间 A', 'date' => '2022-02-02 02:02:02', 'cats' => array( $cat_tech ), 'tags' => array() ),
	array( 'title' => '同一时间 B', 'date' => '2022-02-02 02:02:02', 'cats' => array( $cat_tech ), 'tags' => array() ),
	array( 'title' => 'Wordpress 小技巧', 'date' => '2023-11-11 11:11:11', 'cats' => array( $cat_tech ), 'tags' => array( 'WordPress' ), 'author' => $author ),
);
foreach ( $specs as $i => $spec ) {
	$id = wp_insert_post(
		array(
			'post_title'    => $spec['title'],
			'post_content'  => $contents[ $i % count( $contents ) ],
			'post_status'   => $spec['status'] ?? 'publish',
			'post_date'     => $spec['date'],
			'post_author'   => $spec['author'] ?? 1,
			'post_password' => $spec['password'] ?? '',
			'post_category' => $spec['cats'],
		)
	);
	wp_set_post_terms( $id, $spec['tags'], 'post_tag' );
	if ( ! empty( $spec['sticky'] ) ) {
		stick_post( $id );
	}
	$posts[ $spec['title'] ] = $id;
}
// Playground 自带的 Hello world! 文章与 Sample Page 删除，避免依赖其内容。
wp_delete_post( 1, true );
wp_delete_post( 2, true );
wp_delete_post( 3, true );

// 页面。
$page_about = wp_insert_post(
	array(
		'post_title'   => '关于',
		'post_name'    => 'about',
		'post_content' => "关于我\n\n欢迎来访。",
		'post_status'  => 'publish',
		'post_type'    => 'page',
		'post_date'    => '2019-01-01 00:00:00',
	)
);
$page_cn    = wp_insert_post(
	array(
		'post_title'   => '友情链接',
		'post_name'    => '友情链接',
		'post_content' => '链接们',
		'post_status'  => 'publish',
		'post_type'    => 'page',
		'post_date'    => '2019-01-02 00:00:00',
	)
);

// 修改时间固定下来，避免金标准随运行时间变化。
$wpdb->query( "UPDATE {$wpdb->posts} SET post_modified = '2024-01-01 12:00:00', post_modified_gmt = '2024-01-01 04:00:00'" );
wp_cache_flush();

// 评论。
$hello = $posts['Hello World'];
$c1    = wp_insert_comment(
	array(
		'comment_post_ID'      => $hello,
		'comment_author'       => '路人甲',
		'comment_author_email' => 'Passerby@Example.com',
		'comment_author_url'   => 'https://passerby.example.com',
		'comment_content'      => "第一条评论 http://example.com/a?b=1 :)\n\n第二段 \"引号\"",
		'comment_date'         => '2019-03-02 10:00:00',
		'comment_approved'     => 1,
	)
);
wp_insert_comment(
	array(
		'comment_post_ID'      => $hello,
		'comment_parent'       => $c1,
		'comment_author'       => '张三',
		'comment_author_email' => 'other@example.com',
		'user_id'              => $author,
		'comment_content'      => '<b>回复<i>未闭合',
		'comment_date'         => '2019-03-03 10:00:00',
		'comment_approved'     => 1,
	)
);
wp_insert_comment(
	array(
		'comment_post_ID'      => $hello,
		'comment_author'       => '无邮箱',
		'comment_author_email' => '',
		'comment_content'      => '没有邮箱的评论 www.example.org',
		'comment_date'         => '2019-03-03 10:00:00',
		'comment_approved'     => 1,
	)
);
wp_insert_comment(
	array(
		'comment_post_ID'  => $hello,
		'comment_author'   => '待审',
		'comment_content'  => '待审核评论',
		'comment_date'     => '2019-03-04 10:00:00',
		'comment_approved' => 0,
	)
);
wp_insert_comment(
	array(
		'comment_post_ID'  => $hello,
		'comment_author'   => 'spammer',
		'comment_content'  => 'spam',
		'comment_date'     => '2019-03-05 10:00:00',
		'comment_approved' => 'spam',
	)
);
wp_insert_comment(
	array(
		'comment_post_ID'  => $hello,
		'comment_author'   => 'Some Blog',
		'comment_content'  => 'pingback',
		'comment_type'     => 'pingback',
		'comment_date'     => '2019-03-06 10:00:00',
		'comment_approved' => 1,
	)
);
for ( $i = 1; $i <= 3; $i++ ) {
	wp_insert_comment(
		array(
			'comment_post_ID'      => $page_about,
			'comment_author'       => "访客 $i",
			'comment_author_email' => "guest$i@example.com",
			'comment_content'      => "页面评论 $i",
			'comment_date'         => "2019-02-0$i 10:00:00",
			'comment_approved'     => 1,
		)
	);
}
wp_insert_comment(
	array(
		'comment_post_ID'  => $posts['密码保护的文章'],
		'comment_author'   => '访客',
		'comment_content'  => '密码文章的评论',
		'comment_date'     => '2020-07-02 10:00:00',
		'comment_approved' => 1,
	)
);
wp_insert_comment(
	array(
		'comment_post_ID'  => $posts['草稿'],
		'comment_author'   => '访客',
		'comment_content'  => '草稿的评论',
		'comment_date'     => '2020-08-02 10:00:00',
		'comment_approved' => 1,
	)
);

// 以未登录访客身份发起请求。
wp_set_current_user( 0 );

$embed    = '_embed=author,wp:term';
$requests = array(
	'/',
	"/wp/v2/posts?$embed&page=1",
	"/wp/v2/posts?$embed&page=2",
	'/wp/v2/posts?page=99',
	'/wp/v2/posts?page=0',
	'/wp/v2/posts?per_page=101',
	'/wp/v2/posts?per_page=3&page=2',
	"/wp/v2/posts?$embed&categories=$cat_tech",
	"/wp/v2/posts?categories=$cat_child",
	"/wp/v2/posts?$embed&tags={$tags['Go']}",
	"/wp/v2/posts?tags={$tags['随笔']},{$tags['React']}",
	"/wp/v2/posts?author=$author",
	'/wp/v2/posts?search=' . rawurlencode( '关键词' ),
	'/wp/v2/posts?search=' . rawurlencode( '搜索 测试' ),
	'/wp/v2/posts?search=' . rawurlencode( '时间 -B' ),
	'/wp/v2/posts?search=' . rawurlencode( '"搜索 关键词"' ),
	'/wp/v2/posts?search=' . rawurlencode( '密码' ),
	'/wp/v2/posts?search=' . rawurlencode( 'the' ),
	'/wp/v2/posts?search=' . rawurlencode( '100%' ),
	'/wp/v2/posts?after=2019-12-31T16:00:00.000Z&before=2020-12-31T16:00:00.000Z',
	'/wp/v2/posts?after=2020-01-01T00:00:00&before=2021-01-01T00:00:00',
	'/wp/v2/posts?after=2022-02-02T02:02:02',
	'/wp/v2/posts?after=not-a-date',
	"/wp/v2/posts/{$posts['Hello World']}?$embed",
	"/wp/v2/posts/{$posts['区块编辑器']}?$embed",
	"/wp/v2/posts/{$posts['密码保护的文章']}?$embed",
	"/wp/v2/posts/{$posts['草稿']}",
	"/wp/v2/posts/{$posts['私密文章']}",
	"/wp/v2/posts/$page_about",
	'/wp/v2/posts/99999',
	"/wp/v2/pages?$embed&slug=about",
	"/wp/v2/pages?$embed&slug=" . rawurlencode( '友情链接' ),
	"/wp/v2/pages?$embed&slug=nope",
	'/wp/v2/categories',
	'/?_fields=name,description,url',
	'/wp/v2/tags?per_page=50&orderby=count&order=desc&_fields=id,name,slug,count',
	'/wp/v2/posts?per_page=2&_fields=id,title.rendered',
	'/wp/v2/categories?slug=tech',
	'/wp/v2/categories?slug=' . rawurlencode( '生活' ),
	'/wp/v2/categories?slug=TECH',
	'/wp/v2/tags?per_page=50&orderby=count&order=desc',
	'/wp/v2/tags?per_page=2&page=2',
	'/wp/v2/tags?slug=' . rawurlencode( '前端' ),
	'/wp/v2/tags?slug=a-b',
	'/wp/v2/users?slug=zhangsan',
	'/wp/v2/users?slug=lurker',
	'/wp/v2/users',
	"/wp/v2/comments?post=$hello&page=1&order=desc&per_page=100",
	"/wp/v2/comments?post=$hello&order=asc",
	"/wp/v2/comments?post=$page_about&per_page=2&page=2",
	"/wp/v2/comments?post=$page_about&per_page=2&page=5",
	"/wp/v2/comments?post={$posts['密码保护的文章']}",
	"/wp/v2/comments?post={$posts['草稿']}",
	'/wp/v2/comments',
);

/**
 * 恢复 the_content 默认的优先级 10 回调顺序。
 * do_blocks 会把 wpautop 重新挂到队尾，这一状态在真实环境中只在单个 HTTP 请求内有效。
 */
function i3w_reset_content_filters() {
	$callbacks = array( 'wptexturize', 'wpautop', 'shortcode_unautop', 'prepend_attachment', 'wp_replace_insecure_home_url' );
	foreach ( array_merge( $callbacks, array( '_restore_wpautop_hook' ) ) as $callback ) {
		$priority = has_filter( 'the_content', $callback );
		while ( false !== $priority ) {
			remove_filter( 'the_content', $callback, $priority );
			$priority = has_filter( 'the_content', $callback );
		}
	}
	foreach ( $callbacks as $callback ) {
		add_filter( 'the_content', $callback );
	}
}

/**
 * 去掉本服务不输出的 _links 与首页的 routes，减小金标准体积。
 */
function i3w_strip_links( $data ) {
	if ( ! is_array( $data ) ) {
		return $data;
	}
	unset( $data['_links'], $data['routes'] );
	return array_map( 'i3w_strip_links', $data );
}

// feed 先于 REST 渲染：主循环的图片计数、fetchpriority 标记等是进程级静态状态，真实环境中每个请求都是新进程。
/**
 * 以访客身份渲染 /feed/，返回 RSS2 输出。
 */
function i3w_render_feed( $options ) {
	global $wp_rewrite;
	foreach ( $options as $name => $value ) {
		update_option( $name, $value );
	}
	$wp_rewrite->init();
	i3w_reset_content_filters();
	// redirect_canonical() 会把 feed 地址规范到与固定链接结构一致的尾部斜杠形式。
	$_SERVER['REQUEST_URI'] = str_ends_with( $options['permalink_structure'], '/' ) ? '/feed/' : '/feed';
	wp( array( 'feed' => 'rss2' ) );
	ob_start();
	load_template( ABSPATH . WPINC . '/feed-rss2.php', false );
	return ob_get_clean();
}

$feed_variants = array(
	array(
		'permalink_structure'          => '/%post_id%.moe',
		'posts_per_rss'                => '20',
		'close_comments_for_old_posts' => '0',
	),
	array(
		'permalink_structure'          => '/%category%/%postname%/',
		'posts_per_rss'                => '3',
		'close_comments_for_old_posts' => '1',
	),
);
$feeds         = array();
foreach ( $feed_variants as $options ) {
	$feeds[] = array(
		'options' => $options,
		'body'    => i3w_render_feed( $options ),
	);
}
// 恢复默认选项，tables.json 中保存的是 REST 金标准对应的状态。
update_option( 'permalink_structure', '' );
update_option( 'posts_per_rss', '10' );
update_option( 'close_comments_for_old_posts', '0' );
$wp_rewrite->init();

$golden = array();
foreach ( $requests as $path ) {
	i3w_reset_content_filters();
	$url   = wp_parse_url( $path );
	$query = array();
	if ( isset( $url['query'] ) ) {
		parse_str( $url['query'], $query );
	}
	$request = new WP_REST_Request( 'GET', $url['path'] );
	$request->set_query_params( $query );
	$response = rest_do_request( $request );

	$embed_param = isset( $query['_embed'] ) ? rest_parse_embed_param( $query['_embed'] ) : false;
	$headers     = $response->get_headers();
	$golden[]    = array(
		'path'        => $path,
		'status'      => $response->get_status(),
		'total'       => $headers['X-WP-Total'] ?? null,
		'total_pages' => $headers['X-WP-TotalPages'] ?? null,
		'body'        => i3w_strip_links( rest_get_server()->response_to_data( $response, $embed_param ) ),
	);
}

$tables = array();
foreach ( array( 'options', 'users', 'posts', 'postmeta', 'terms', 'term_taxonomy', 'term_relationships', 'comments' ) as $table ) {
	$rows = $wpdb->get_results( "SELECT * FROM {$wpdb->$table}", ARRAY_A );
	if ( 'options' === $table ) {
		$rows = array_values(
			array_filter(
				$rows,
				function ( $row ) {
					return ! str_starts_with( $row['option_name'], '_' ) && strlen( $row['option_value'] ) < 2000;
				}
			)
		);
	}
	if ( 'users' === $table ) {
		foreach ( $rows as &$row ) {
			$row['user_pass'] = '';
		}
		unset( $row );
	}
	$tables[ $table ] = $rows;
}

$flags = JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES;
file_put_contents( '/work/tables.json', json_encode( $tables, $flags ) );
file_put_contents( '/work/golden.json', json_encode( $golden, $flags ) );
file_put_contents( '/work/feeds.json', json_encode( $feeds, $flags ) );
echo 'requests: ' . count( $golden ) . "\n";
