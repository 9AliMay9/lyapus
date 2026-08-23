-- name: CreateService :one
INSERT INTO services (team_id, slug, name, description)
VALUES ($1, $2, $3, $4)
RETURNING id, team_id, slug, name, description, created_at, updated_at;

-- name: CreateEnvironment :one
INSERT INTO environments (service_id, slug, name)
VALUES ($1, $2, $3)
RETURNING id, service_id, slug, name, created_at, updated_at;

-- name: GetServiceByID :one
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM services
WHERE id = $1;

-- name: ListEnvironmentsByServiceID :many
SELECT id, service_id, slug, name, created_at, updated_at
FROM environments
WHERE service_id = $1
ORDER BY created_at ASC, id ASC;

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
