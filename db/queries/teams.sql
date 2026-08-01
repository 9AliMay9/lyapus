-- name: CreateTeam :one
INSERT INTO teams (slug, name)
VALUES ($1, $2)
RETURNING id, slug, name, created_at, updated_at;

-- name: GetTeamByID :one
SELECT id, slug, name, created_at, updated_at
FROM teams
WHERE id = $1;
