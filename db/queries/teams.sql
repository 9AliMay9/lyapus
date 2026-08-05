-- name: CreateTeam :one
INSERT INTO teams (slug, name)
VALUES ($1, $2)
RETURNING id, slug, name, created_at, updated_at;

-- name: GetTeamByID :one
SELECT id, slug, name, created_at, updated_at
FROM teams
WHERE id = $1;

-- name: ListTeamsFirstPage :many
SELECT id, slug, name, created_at, updated_at
FROM teams
ORDER BY created_at DESC, id DESC
LIMIT $1;

-- name: ListTeamsAfterCursor :many
SELECT id, slug, name, created_at, updated_at
FROM teams
WHERE created_at < $1
  OR (created_at = $1 AND id < $2)
ORDER BY created_at DESC, id DESC
LIMIT $3;

-- name: UpdateTeam :one
UPDATE teams
SET slug = $2,
    name = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, slug, name, created_at, updated_at;

-- name: DeleteTeam :one
DELETE FROM teams
WHERE id = $1
RETURNING id;
