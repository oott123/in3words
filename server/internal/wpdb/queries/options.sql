-- name: ListOptions :many
SELECT option_name, option_value
FROM `wp_options`
WHERE option_name IN (sqlc.slice(names));
