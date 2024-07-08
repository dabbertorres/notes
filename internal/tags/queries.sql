-- name: SaveTag :exec
INSERT INTO notes.tags (
  tag_id,
  name
) VALUES (
  sqlc.arg(tag_id),
  sqlc.arg(name)
) ON CONFLICT (tag_id) DO UPDATE
  SET name = excluded.name
;

-- name: DeleteTag :execrows
DELETE FROM notes.tags
WHERE tag_id = sqlc.arg(tag_id)
;

-- name: GetTag :one
SELECT
  tag_id,
  name
FROM notes.tags
WHERE tag_id = sqlc.arg(tag_id)
;

-- name: ListTags :many
SELECT
  tags.tag_id,
  name,
  access
FROM notes.tags
JOIN notes.user_tag_access ON 
  tags.tag_id = user_tag_access.tag_id
  AND user_id = sqlc.arg(user_id) -- NOTE: any access
WHERE (sqlc.narg(last_tag_id)::uuid IS NULL OR tags.tag_id > sqlc.narg(last_tag_id)::uuid)
  AND LOWER(name) LIKE '%' || LOWER(sqlc.arg(search_string)) || '%'
ORDER BY tags.tag_id ASC
LIMIT sqlc.arg(page_size)
;

-- name: GetUserTagAccess :one
SELECT
  access
FROM notes.user_tag_access
WHERE tag_id = sqlc.arg(tag_id)
  AND user_id = sqlc.arg(user_id)
;

-- name: GetTagAccess :many
SELECT
  user_id,
  access
FROM notes.user_tag_access
WHERE tag_id = sqlc.arg(tag_id)
;

-- name: SetTagAccess :exec
MERGE INTO notes.user_tag_access
USING (SELECT $1::uuid AS set_user_id,
              $2::notes.access_level AS set_access) ON
  tag_id = $3::uuid
  AND user_id = set_user_id
WHEN MATCHED AND set_access IS NULL THEN
  DELETE
WHEN MATCHED AND set_access IS NOT NULL THEN
  UPDATE SET access = set_access
WHEN NOT MATCHED THEN
  INSERT (tag_id, user_id, access)
  VALUES ($3::uuid, set_user_id, set_access)
;
