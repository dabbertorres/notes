-- TODO: access control

-- name: SaveNote :one
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
RETURNING *
;

-- name: AddNoteTags :batchexec
INSERT INTO notes.note_tags (
  note_id, tag_id
) VALUES (
  sqlc.arg(note_id), sqlc.arg(tag_id)
) ON CONFLICT DO NOTHING;

-- name: DeleteNoteTags :batchexec
DELETE FROM notes.note_tags
WHERE 
  note_id = sqlc.arg(note_id)
  AND tag_id = sqlc.arg(tag_id)
;

-- name: AddNoteAccess :batchexec
INSERT INTO notes.user_note_access (
  note_id,
  user_id,
  access
) VALUES (
  sqlc.arg(note_id), sqlc.arg(user_id), sqlc.arg(access)
) ON CONFLICT DO UPDATE
  SET access = excluded.access
;

-- name: DeleteNoteAccess :batchexec
DELETE FROM notes.user_note_access
WHERE
  note_id = sqlc.arg(note_id)
  AND user_id = sqlc.arg(user_id)
;

-- name: DeleteNote :exec
DELETE FROM notes.notes 
WHERE 
  note_id = sqlc.arg(note_id)
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
WHERE
  note_id = sqlc.arg(note_id)
;

-- name: GetNoteTags :many
SELECT
  tags.tag_id,
  user_id,
  name
FROM notes.tags
JOIN notes.note_tags ON
  note_tags.tag_id = tags.tag_id
WHERE
  note_id = sqlc.arg(note_id)
;

-- name: GetNoteAccess :many
SELECT
  user_id,
  access
FROM notes.user_note_access
WHERE
  note_id = sqlc.arg(note_id)
;

-- name: SearchNotes :many
SELECT
  note_id,
  ts_rank_cd(search_index, sqlc.arg(search)::text)::float4 AS rank
  -- TODO: highlights: https://www.postgresql.org/docs/current/textsearch-controls.html#TEXTSEARCH-HEADLINE
FROM notes.notes
WHERE
  rank < sqlc.arg(rank)::float4
  AND search_index @@ to_tsquery(sqlc.arg(search))
ORDER BY
  rank DESC
;

-- name: ListTags :many
SELECT
  tag_id,
  ordered_id,
  name
FROM notes.tags
WHERE
  user_id = $1
  AND ordered_id >= $2
LIMIT $3
;
