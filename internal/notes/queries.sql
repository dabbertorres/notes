-- name: SaveNote :exec
INSERT INTO notes.notes (
  note_id,
  created_at,
  created_by,
  updated_at,
  updated_by,
  title,
  body
) VALUES (
  sqlc.arg(note_id),
  sqlc.arg(created_at),
  sqlc.arg(created_by),
  sqlc.arg(updated_at),
  sqlc.arg(updated_by),
  sqlc.arg(title),
  sqlc.arg(body)
) ON CONFLICT (note_id) DO UPDATE
  SET updated_at = excluded.updated_at,
      updated_by = excluded.updated_by,
      title      = excluded.title,
      body       = excluded.body
;

-- name: DeleteNote :execrows
DELETE FROM notes.notes
WHERE note_id = sqlc.arg(note_id)
;

-- name: GetNote :one
SELECT
  note_id,
  created_at,
  created_by,
  updated_at,
  updated_by,
  title,
  body
FROM notes.notes
WHERE note_id = sqlc.arg(note_id)
;

-- name: SetNoteTags :exec
MERGE INTO notes.note_tags
USING (SELECT $2::uuid AS set_tag_id) ON
  note_id = $1::uuid
WHEN MATCHED AND set_tag_id IS NULL THEN
  DELETE
WHEN MATCHED AND set_tag_id IS NOT NULL THEN
  UPDATE SET tag_id = set_tag_id
WHEN NOT MATCHED THEN
  INSERT (note_id, tag_id)
  VALUES ($1::uuid, set_tag_id)
;

-- name: SetNoteAccess :exec
MERGE INTO notes.user_note_access
USING (SELECT $1::uuid AS set_user_id,
              $2::notes.access_level AS set_access) ON
  note_id = $3::uuid
  AND user_id = set_user_id
WHEN MATCHED AND set_access IS NULL THEN
  DELETE
WHEN MATCHED AND set_access IS NOT NULL THEN
  UPDATE SET access = set_access
WHEN NOT MATCHED THEN
  INSERT (note_id, user_id, access)
  VALUES ($3::uuid, set_user_id, set_access)
;

-- name: GetNoteTags :many
SELECT
  tags.tag_id,
  name
FROM notes.note_tags
JOIN notes.tags ON
  note_tags.tag_id = tags.tag_id
JOIN notes.user_tag_access ON
  user_tag_access.tag_id = tags.tag_id
  AND user_id = sqlc.arg(user_id)
  -- NOTE: any access
WHERE note_id = sqlc.arg(note_id)
;

-- name: GetNoteAccess :many
SELECT
  user_note_access.user_id,
  access
FROM notes.user_note_access
JOIN notes.users ON
  user_note_access.user_id = users.user_id
WHERE note_id = sqlc.arg(note_id)
;

-- name: GetUserNoteAccess :one
SELECT
  access
FROM notes.user_note_access
WHERE note_id = sqlc.arg(note_id)
  AND user_id = sqlc.arg(user_id)
;

-- name: ListNotes :many
SELECT
  notes.note_id,
  title
FROM notes.notes
JOIN notes.user_note_access ON
  user_note_access.note_id = notes.note_id
  AND user_id = sqlc.arg(user_id)
  -- NOTE: any access
WHERE sqlc.narg(last_note_id)::uuid IS NULL OR notes.note_id > sqlc.narg(last_note_id)::uuid
ORDER BY notes.note_id ASC
LIMIT sqlc.arg(page_size)
;

-- name: SearchNotesWithText :many
SELECT
  notes.note_id,
  title,
  rank::float4,
  ts_headline(title || '\n' || body, query, 'StartSel=<<, StopSel=>>') AS match
FROM notes.notes
CROSS JOIN LATERAL websearch_to_tsquery(sqlc.arg(text_search)) AS query
CROSS JOIN LATERAL ts_rank_cd(search_index, query) AS rank
JOIN notes.user_note_access ON
  user_note_access.note_id = notes.note_id
  AND user_id = sqlc.arg(user_id)
  -- NOTE: any access
WHERE query @@ search_index
  AND sqlc.narg(last_rank)::float4 IS NULL OR rank::float4 < sqlc.narg(last_rank)::float4
ORDER BY rank::float4 DESC
LIMIT sqlc.arg(page_size)
;

-- name: SearchNotesWithTag :many
SELECT
  notes.note_id,
  title
FROM notes.notes
JOIN notes.user_note_access ON
  user_note_access.note_id = notes.note_id
  AND user_id = sqlc.arg(user_id)
  -- NOTE: any access
JOIN notes.note_tags ON
  note_tags.note_id = notes.note_id
WHERE note_tags.tag_id = sqlc.arg(tag_id)
  AND (sqlc.narg(last_note_id)::uuid IS NULL OR notes.note_id > sqlc.narg(last_note_id)::uuid)
ORDER BY notes.note_id ASC
LIMIT sqlc.arg(page_size)
;

-- name: SearchNotesWithTextAndTag :many
SELECT
  notes.note_id,
  title,
  rank::float4 AS rank,
  ts_headline(title || '\n' || body, query, 'StartSel=<<, StopSel=>>') AS match
FROM notes.notes
CROSS JOIN LATERAL websearch_to_tsquery(sqlc.arg(text_search)) AS query
CROSS JOIN LATERAL ts_rank_cd(search_index, query) AS rank
JOIN notes.user_note_access ON
  user_note_access.note_id = notes.note_id
  AND user_id = sqlc.arg(user_id)
  -- NOTE: any access
JOIN notes.note_tags ON
  note_tags.note_id = notes.note_id
WHERE tag_id = sqlc.arg(tag_id)
  AND query @@ search_index
  AND (sqlc.narg(last_rank)::float4 IS NULL OR rank::float4 < sqlc.narg(last_rank)::float4)
ORDER BY rank::float4 DESC
LIMIT sqlc.arg(page_size)
;
