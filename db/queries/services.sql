-- name: CreateService :one
INSERT INTO services (team_id, slug, name, description)
VALUES ($1, $2, $3, $4)
RETURNING id, team_id, slug, name, description, created_at, updated_at;

-- name: GetServiceByID :one
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM services
WHERE id = $1;

-- name: ListServicesFirstPage :many
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM services
ORDER BY created_at DESC, id DESC
LIMIT $1;

-- name: ListServicesFirstPageByTeamID :many
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM services
WHERE team_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: ListServicesAfterCursor :many
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM services
WHERE created_at < $1
  OR (created_at = $1 AND id < $2)
ORDER BY created_at DESC, id DESC
LIMIT $3;

-- name: ListServicesAfterCursorByTeamID :many
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM services
WHERE team_id = $1
  AND (
    created_at < $2
    OR (created_at = $2 AND id < $3)
  )
ORDER BY created_at DESC, id DESC
LIMIT $4;

-- name: UpdateService :one
UPDATE services
SET slug = COALESCE(sqlc.narg('slug')::text, slug),
    name = COALESCE(sqlc.narg('name')::text, name),
    description = CASE
      WHEN sqlc.arg('description_provided')::boolean
        THEN sqlc.narg('description')::text
      ELSE description
    END,
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, team_id, slug, name, description, created_at, updated_at;

-- name: DeleteService :one
DELETE FROM services
WHERE id = $1
RETURNING id;
