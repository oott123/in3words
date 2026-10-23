-- 受密码保护的文章不进入 feed。

-- name: ListFeedPosts :many
SELECT *
FROM `wp_posts`
WHERE post_type = 'post'
  AND post_status = 'publish'
  AND post_password = ''
ORDER BY post_date DESC, ID DESC
LIMIT ?;

-- name: ListTaxonomyTerms :many
SELECT t.term_id, t.slug, tt.parent
FROM `wp_terms` t
INNER JOIN `wp_term_taxonomy` tt ON tt.term_id = t.term_id
WHERE tt.taxonomy = sqlc.arg(taxonomy);
