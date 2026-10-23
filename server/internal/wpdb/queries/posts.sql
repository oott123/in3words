-- name: GetPost :one
SELECT *
FROM `wp_posts`
WHERE ID = sqlc.arg(id);

-- 文章列表的筛选条件与 WP_Query 保持一致；ListPosts 与 CountPosts 的 WHERE 必须同步修改。
-- 搜索词最多 9 个（WP_Query::parse_search 超过 9 个时整句作为一个词），
-- s1..s9 为包含词，x1..x9 为排除词（以 - 开头的词），值为已转义的 LIKE 模式。

-- name: ListPosts :many
SELECT *
FROM `wp_posts` p
WHERE p.post_type = sqlc.arg(post_type)
  AND p.post_status = 'publish'
  AND (CAST(sqlc.arg(filter_authors) AS SIGNED) = 0 OR p.post_author IN (sqlc.slice(authors)))
  AND (CAST(sqlc.arg(filter_slugs) AS SIGNED) = 0 OR p.post_name IN (sqlc.slice(slugs)))
  AND (CAST(sqlc.arg(filter_categories) AS SIGNED) = 0 OR EXISTS (
    SELECT 1
    FROM `wp_term_relationships` tr
    INNER JOIN `wp_term_taxonomy` tt ON tt.term_taxonomy_id = tr.term_taxonomy_id
    WHERE tr.object_id = p.ID AND tt.taxonomy = 'category' AND tt.term_id IN (sqlc.slice(categories))
  ))
  AND (CAST(sqlc.arg(filter_tags) AS SIGNED) = 0 OR EXISTS (
    SELECT 1
    FROM `wp_term_relationships` tr
    INNER JOIN `wp_term_taxonomy` tt ON tt.term_taxonomy_id = tr.term_taxonomy_id
    WHERE tr.object_id = p.ID AND tt.taxonomy = 'post_tag' AND tt.term_id IN (sqlc.slice(tags))
  ))
  AND (sqlc.narg(after) IS NULL OR p.post_date > sqlc.narg(after))
  AND (sqlc.narg(before) IS NULL OR p.post_date < sqlc.narg(before))
  AND (CAST(sqlc.arg(searching) AS SIGNED) = 0 OR p.post_password = '')
  AND (sqlc.narg(s1) IS NULL OR p.post_title LIKE sqlc.narg(s1) OR p.post_excerpt LIKE sqlc.narg(s1) OR p.post_content LIKE sqlc.narg(s1))
  AND (sqlc.narg(s2) IS NULL OR p.post_title LIKE sqlc.narg(s2) OR p.post_excerpt LIKE sqlc.narg(s2) OR p.post_content LIKE sqlc.narg(s2))
  AND (sqlc.narg(s3) IS NULL OR p.post_title LIKE sqlc.narg(s3) OR p.post_excerpt LIKE sqlc.narg(s3) OR p.post_content LIKE sqlc.narg(s3))
  AND (sqlc.narg(s4) IS NULL OR p.post_title LIKE sqlc.narg(s4) OR p.post_excerpt LIKE sqlc.narg(s4) OR p.post_content LIKE sqlc.narg(s4))
  AND (sqlc.narg(s5) IS NULL OR p.post_title LIKE sqlc.narg(s5) OR p.post_excerpt LIKE sqlc.narg(s5) OR p.post_content LIKE sqlc.narg(s5))
  AND (sqlc.narg(s6) IS NULL OR p.post_title LIKE sqlc.narg(s6) OR p.post_excerpt LIKE sqlc.narg(s6) OR p.post_content LIKE sqlc.narg(s6))
  AND (sqlc.narg(s7) IS NULL OR p.post_title LIKE sqlc.narg(s7) OR p.post_excerpt LIKE sqlc.narg(s7) OR p.post_content LIKE sqlc.narg(s7))
  AND (sqlc.narg(s8) IS NULL OR p.post_title LIKE sqlc.narg(s8) OR p.post_excerpt LIKE sqlc.narg(s8) OR p.post_content LIKE sqlc.narg(s8))
  AND (sqlc.narg(s9) IS NULL OR p.post_title LIKE sqlc.narg(s9) OR p.post_excerpt LIKE sqlc.narg(s9) OR p.post_content LIKE sqlc.narg(s9))
  AND (sqlc.narg(x1) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x1) AND p.post_excerpt NOT LIKE sqlc.narg(x1) AND p.post_content NOT LIKE sqlc.narg(x1)))
  AND (sqlc.narg(x2) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x2) AND p.post_excerpt NOT LIKE sqlc.narg(x2) AND p.post_content NOT LIKE sqlc.narg(x2)))
  AND (sqlc.narg(x3) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x3) AND p.post_excerpt NOT LIKE sqlc.narg(x3) AND p.post_content NOT LIKE sqlc.narg(x3)))
  AND (sqlc.narg(x4) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x4) AND p.post_excerpt NOT LIKE sqlc.narg(x4) AND p.post_content NOT LIKE sqlc.narg(x4)))
  AND (sqlc.narg(x5) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x5) AND p.post_excerpt NOT LIKE sqlc.narg(x5) AND p.post_content NOT LIKE sqlc.narg(x5)))
  AND (sqlc.narg(x6) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x6) AND p.post_excerpt NOT LIKE sqlc.narg(x6) AND p.post_content NOT LIKE sqlc.narg(x6)))
  AND (sqlc.narg(x7) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x7) AND p.post_excerpt NOT LIKE sqlc.narg(x7) AND p.post_content NOT LIKE sqlc.narg(x7)))
  AND (sqlc.narg(x8) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x8) AND p.post_excerpt NOT LIKE sqlc.narg(x8) AND p.post_content NOT LIKE sqlc.narg(x8)))
  AND (sqlc.narg(x9) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x9) AND p.post_excerpt NOT LIKE sqlc.narg(x9) AND p.post_content NOT LIKE sqlc.narg(x9)))
ORDER BY p.post_date DESC, p.ID DESC
LIMIT ? OFFSET ?;

-- name: CountPosts :one
SELECT COUNT(*)
FROM `wp_posts` p
WHERE p.post_type = sqlc.arg(post_type)
  AND p.post_status = 'publish'
  AND (CAST(sqlc.arg(filter_authors) AS SIGNED) = 0 OR p.post_author IN (sqlc.slice(authors)))
  AND (CAST(sqlc.arg(filter_slugs) AS SIGNED) = 0 OR p.post_name IN (sqlc.slice(slugs)))
  AND (CAST(sqlc.arg(filter_categories) AS SIGNED) = 0 OR EXISTS (
    SELECT 1
    FROM `wp_term_relationships` tr
    INNER JOIN `wp_term_taxonomy` tt ON tt.term_taxonomy_id = tr.term_taxonomy_id
    WHERE tr.object_id = p.ID AND tt.taxonomy = 'category' AND tt.term_id IN (sqlc.slice(categories))
  ))
  AND (CAST(sqlc.arg(filter_tags) AS SIGNED) = 0 OR EXISTS (
    SELECT 1
    FROM `wp_term_relationships` tr
    INNER JOIN `wp_term_taxonomy` tt ON tt.term_taxonomy_id = tr.term_taxonomy_id
    WHERE tr.object_id = p.ID AND tt.taxonomy = 'post_tag' AND tt.term_id IN (sqlc.slice(tags))
  ))
  AND (sqlc.narg(after) IS NULL OR p.post_date > sqlc.narg(after))
  AND (sqlc.narg(before) IS NULL OR p.post_date < sqlc.narg(before))
  AND (CAST(sqlc.arg(searching) AS SIGNED) = 0 OR p.post_password = '')
  AND (sqlc.narg(s1) IS NULL OR p.post_title LIKE sqlc.narg(s1) OR p.post_excerpt LIKE sqlc.narg(s1) OR p.post_content LIKE sqlc.narg(s1))
  AND (sqlc.narg(s2) IS NULL OR p.post_title LIKE sqlc.narg(s2) OR p.post_excerpt LIKE sqlc.narg(s2) OR p.post_content LIKE sqlc.narg(s2))
  AND (sqlc.narg(s3) IS NULL OR p.post_title LIKE sqlc.narg(s3) OR p.post_excerpt LIKE sqlc.narg(s3) OR p.post_content LIKE sqlc.narg(s3))
  AND (sqlc.narg(s4) IS NULL OR p.post_title LIKE sqlc.narg(s4) OR p.post_excerpt LIKE sqlc.narg(s4) OR p.post_content LIKE sqlc.narg(s4))
  AND (sqlc.narg(s5) IS NULL OR p.post_title LIKE sqlc.narg(s5) OR p.post_excerpt LIKE sqlc.narg(s5) OR p.post_content LIKE sqlc.narg(s5))
  AND (sqlc.narg(s6) IS NULL OR p.post_title LIKE sqlc.narg(s6) OR p.post_excerpt LIKE sqlc.narg(s6) OR p.post_content LIKE sqlc.narg(s6))
  AND (sqlc.narg(s7) IS NULL OR p.post_title LIKE sqlc.narg(s7) OR p.post_excerpt LIKE sqlc.narg(s7) OR p.post_content LIKE sqlc.narg(s7))
  AND (sqlc.narg(s8) IS NULL OR p.post_title LIKE sqlc.narg(s8) OR p.post_excerpt LIKE sqlc.narg(s8) OR p.post_content LIKE sqlc.narg(s8))
  AND (sqlc.narg(s9) IS NULL OR p.post_title LIKE sqlc.narg(s9) OR p.post_excerpt LIKE sqlc.narg(s9) OR p.post_content LIKE sqlc.narg(s9))
  AND (sqlc.narg(x1) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x1) AND p.post_excerpt NOT LIKE sqlc.narg(x1) AND p.post_content NOT LIKE sqlc.narg(x1)))
  AND (sqlc.narg(x2) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x2) AND p.post_excerpt NOT LIKE sqlc.narg(x2) AND p.post_content NOT LIKE sqlc.narg(x2)))
  AND (sqlc.narg(x3) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x3) AND p.post_excerpt NOT LIKE sqlc.narg(x3) AND p.post_content NOT LIKE sqlc.narg(x3)))
  AND (sqlc.narg(x4) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x4) AND p.post_excerpt NOT LIKE sqlc.narg(x4) AND p.post_content NOT LIKE sqlc.narg(x4)))
  AND (sqlc.narg(x5) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x5) AND p.post_excerpt NOT LIKE sqlc.narg(x5) AND p.post_content NOT LIKE sqlc.narg(x5)))
  AND (sqlc.narg(x6) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x6) AND p.post_excerpt NOT LIKE sqlc.narg(x6) AND p.post_content NOT LIKE sqlc.narg(x6)))
  AND (sqlc.narg(x7) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x7) AND p.post_excerpt NOT LIKE sqlc.narg(x7) AND p.post_content NOT LIKE sqlc.narg(x7)))
  AND (sqlc.narg(x8) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x8) AND p.post_excerpt NOT LIKE sqlc.narg(x8) AND p.post_content NOT LIKE sqlc.narg(x8)))
  AND (sqlc.narg(x9) IS NULL OR (p.post_title NOT LIKE sqlc.narg(x9) AND p.post_excerpt NOT LIKE sqlc.narg(x9) AND p.post_content NOT LIKE sqlc.narg(x9)));
