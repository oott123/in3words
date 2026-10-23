-- 访客只能读到已审核、类型为普通评论、且所属文章已公开发布并且没有密码的评论。
-- WP 是查询后再逐条过滤所属文章的，这里直接放进 WHERE，分页总数因此更准确。
-- ListComments 与 CountComments 的 WHERE 必须同步修改。

-- name: ListComments :many
SELECT c.comment_ID, c.comment_post_ID, c.comment_author, c.comment_author_email, c.comment_author_url,
  c.comment_date, c.comment_date_gmt, c.comment_content, c.comment_approved, c.comment_type,
  c.comment_parent, c.user_id, COALESCE(u.user_email, '') AS user_email
FROM `wp_comments` c
INNER JOIN `wp_posts` p ON p.ID = c.comment_post_ID
LEFT JOIN `wp_users` u ON u.ID = c.user_id
WHERE c.comment_approved = '1'
  AND c.comment_type IN ('', 'comment')
  AND p.post_status = 'publish'
  AND p.post_password = ''
  AND (CAST(sqlc.arg(filter_posts) AS SIGNED) = 0 OR c.comment_post_ID IN (sqlc.slice(posts)))
  AND (CAST(sqlc.arg(filter_parents) AS SIGNED) = 0 OR c.comment_parent IN (sqlc.slice(parents)))
ORDER BY
  CASE WHEN CAST(sqlc.arg(ascending) AS SIGNED) = 1 THEN c.comment_date_gmt END ASC,
  CASE WHEN CAST(sqlc.arg(ascending) AS SIGNED) = 1 THEN c.comment_ID END ASC,
  CASE WHEN CAST(sqlc.arg(ascending) AS SIGNED) = 0 THEN c.comment_date_gmt END DESC,
  CASE WHEN CAST(sqlc.arg(ascending) AS SIGNED) = 0 THEN c.comment_ID END DESC
LIMIT ? OFFSET ?;

-- name: CountComments :one
SELECT COUNT(*)
FROM `wp_comments` c
INNER JOIN `wp_posts` p ON p.ID = c.comment_post_ID
WHERE c.comment_approved = '1'
  AND c.comment_type IN ('', 'comment')
  AND p.post_status = 'publish'
  AND p.post_password = ''
  AND (CAST(sqlc.arg(filter_posts) AS SIGNED) = 0 OR c.comment_post_ID IN (sqlc.slice(posts)))
  AND (CAST(sqlc.arg(filter_parents) AS SIGNED) = 0 OR c.comment_parent IN (sqlc.slice(parents)));

-- name: ListPostsByIDs :many
SELECT ID, post_type, post_status, post_password
FROM `wp_posts`
WHERE ID IN (sqlc.slice(ids));
