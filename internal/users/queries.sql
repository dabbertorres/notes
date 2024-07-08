-- name: SaveUser :exec
INSERT INTO notes.users (
  user_id,
  name,
  created_at,
  last_sign_in,
  active
) VALUES (
  sqlc.arg(user_id),
  sqlc.arg(name),
  sqlc.arg(created_at),
  sqlc.arg(last_sign_in),
  sqlc.arg(active)
) ON CONFLICT (user_id) DO UPDATE
  SET name         = excluded.name,
      last_sign_in = excluded.last_sign_in,
      active       = excluded.active
RETURNING *
;

-- name: DeleteUser :execrows
DELETE FROM notes.users
WHERE
  user_id = sqlc.arg(user_id)
;

-- name: GetUser :one
SELECT
  user_id,
  name,
  created_at,
  last_sign_in,
  active
FROM notes.users
WHERE
  user_id = sqlc.arg(note_id)
;
