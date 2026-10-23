-- 对应 wp_get_attachment_metadata()：附件文章必须存在，同一附件有多条时取最早的一条（get_post_meta 的 single 语义）。

-- name: ListAttachmentMetadata :many
SELECT pm.post_id, COALESCE(pm.meta_value, '') AS meta_value
FROM `wp_postmeta` pm
INNER JOIN `wp_posts` p ON p.ID = pm.post_id
WHERE pm.post_id IN (sqlc.slice(ids))
  AND pm.meta_key = '_wp_attachment_metadata'
ORDER BY pm.meta_id ASC;
