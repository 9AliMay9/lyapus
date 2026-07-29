-- name: GetTeamByID :one
SELECT id, slug, name, created_at, updated_at
FROM teams
WHERE id = $1;
