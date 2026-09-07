-- name: CreateEnvironment :one
INSERT INTO environments (service_id, slug, name)
VALUES ($1, $2, $3)
RETURNING id, service_id, slug, name, created_at, updated_at;

-- name: GetEnvironmentByID :one
SELECT id, service_id, slug, name, created_at, updated_at
FROM environments
WHERE id = $1;

-- name: ListEnvironmentsByServiceID :many
SELECT id, service_id, slug, name, created_at, updated_at
FROM environments
WHERE service_id = $1
ORDER BY created_at ASC, id ASC;

-- name: ListEnvironmentsFirstPage :many
SELECT id, service_id, slug, name, created_at, updated_at
FROM environments
ORDER BY created_at DESC, id DESC
LIMIT $1;

-- name: ListEnvironmentsFirstPageByServiceID :many
SELECT id, service_id, slug, name, created_at, updated_at
FROM environments
WHERE service_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: ListEnvironmentsAfterCursor :many
SELECT id, service_id, slug, name, created_at, updated_at
FROM environments
WHERE created_at < $1
  OR (created_at = $1 AND id < $2)
ORDER BY created_at DESC, id DESC
LIMIT $3;

-- name: ListEnvironmentsAfterCursorByServiceID :many
SELECT id, service_id, slug, name, created_at, updated_at
FROM environments
WHERE service_id = $1
  AND (
    created_at < $2
    OR (created_at = $2 AND id < $3)
  )
ORDER BY created_at DESC, id DESC
LIMIT $4;

-- name: UpdateEnvironment :one
UPDATE environments
SET slug = COALESCE(sqlc.narg('slug')::text, slug),
    name = COALESCE(sqlc.narg('name')::text, name),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, service_id, slug, name, created_at, updated_at;

-- name: DeleteEnvironment :one
DELETE FROM environments
WHERE id = $1
RETURNING id;
