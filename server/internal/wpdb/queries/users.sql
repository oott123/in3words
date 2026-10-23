-- has_published_posts 对应 WP_User_Query 的同名参数：未登录访客只能看到发表过文章的用户。

-- name: ListUsersByIDs :many
SELECT u.ID, u.user_login, u.user_nicename, u.user_email, u.user_url, u.display_name,
  EXISTS (
    SELECT 1
    FROM `wp_posts` p
    WHERE p.post_author = u.ID AND p.post_status = 'publish' AND p.post_type IN (sqlc.slice(post_types))
  ) AS has_published_posts
FROM `wp_users` u
WHERE u.ID IN (sqlc.slice(ids));

-- name: ListPublishedAuthors :many
SELECT u.ID, u.user_login, u.user_nicename, u.user_email, u.user_url, u.display_name
FROM `wp_users` u
WHERE (CAST(sqlc.arg(filter_slugs) AS SIGNED) = 0 OR u.user_nicename IN (sqlc.slice(slugs)))
  AND u.ID IN (
    SELECT p.post_author
    FROM `wp_posts` p
    WHERE p.post_status = 'publish' AND p.post_type IN (sqlc.slice(post_types))
  )
ORDER BY u.display_name ASC, u.ID ASC
LIMIT ? OFFSET ?;

-- name: CountPublishedAuthors :one
SELECT COUNT(*)
FROM `wp_users` u
WHERE (CAST(sqlc.arg(filter_slugs) AS SIGNED) = 0 OR u.user_nicename IN (sqlc.slice(slugs)))
  AND u.ID IN (
    SELECT p.post_author
    FROM `wp_posts` p
    WHERE p.post_status = 'publish' AND p.post_type IN (sqlc.slice(post_types))
  );
