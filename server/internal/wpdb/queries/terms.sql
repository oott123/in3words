-- 排序方式由 sort 参数决定，取值为 name_asc、name_desc、count_asc、count_desc；
-- WP 的 get_terms 没有次级排序，这里补上 term_id 以保证结果稳定。

-- name: ListTerms :many
SELECT t.term_id, t.name, t.slug, tt.taxonomy, tt.description, tt.parent, tt.count
FROM `wp_terms` t
INNER JOIN `wp_term_taxonomy` tt ON tt.term_id = t.term_id
WHERE tt.taxonomy = sqlc.arg(taxonomy)
  AND (CAST(sqlc.arg(filter_slugs) AS SIGNED) = 0 OR t.slug IN (sqlc.slice(slugs)))
ORDER BY
  CASE WHEN CAST(sqlc.arg(sort) AS CHAR) = 'name_asc' THEN t.name END ASC,
  CASE WHEN CAST(sqlc.arg(sort) AS CHAR) = 'name_desc' THEN t.name END DESC,
  CASE WHEN CAST(sqlc.arg(sort) AS CHAR) = 'count_asc' THEN tt.count END ASC,
  CASE WHEN CAST(sqlc.arg(sort) AS CHAR) = 'count_desc' THEN tt.count END DESC,
  t.term_id ASC
LIMIT ? OFFSET ?;

-- name: CountTerms :one
SELECT COUNT(*)
FROM `wp_terms` t
INNER JOIN `wp_term_taxonomy` tt ON tt.term_id = t.term_id
WHERE tt.taxonomy = sqlc.arg(taxonomy)
  AND (CAST(sqlc.arg(filter_slugs) AS SIGNED) = 0 OR t.slug IN (sqlc.slice(slugs)));

-- 对应 wp_get_object_terms 的默认排序（按名称升序）。

-- name: ListObjectTerms :many
SELECT tr.object_id, t.term_id, t.name, t.slug, tt.taxonomy, tt.description, tt.parent, tt.count
FROM `wp_term_relationships` tr
INNER JOIN `wp_term_taxonomy` tt ON tt.term_taxonomy_id = tr.term_taxonomy_id
INNER JOIN `wp_terms` t ON t.term_id = tt.term_id
WHERE tr.object_id IN (sqlc.slice(object_ids))
  AND tt.taxonomy IN (sqlc.slice(taxonomies))
ORDER BY t.name ASC, t.term_id ASC;
