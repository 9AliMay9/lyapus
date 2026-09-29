\set ON_ERROR_STOP on

BEGIN;

SET LOCAL statement_timeout = '120s';
SET LOCAL lock_timeout = '5s';
SET LOCAL search_path = public, pg_catalog;

DO $$
BEGIN
    IF current_database() <> 'lyapus_query_plan_test' THEN
        RAISE EXCEPTION 'refusing to change a non-experiment database';
    END IF;
END
$$;

DROP INDEX public.services_team_id_created_at_id_idx;

CREATE INDEX services_team_id_id_idx
    ON public.services (team_id, id);

\echo 'Control index: first page'

EXPLAIN (ANALYZE, BUFFERS)
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM public.services
WHERE team_id = 1
ORDER BY created_at DESC, id DESC
LIMIT 21;

\echo 'Control index: deep page with original OR predicate'

EXPLAIN (ANALYZE, BUFFERS)
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM public.services
WHERE team_id = 1
  AND (
    created_at < TIMESTAMPTZ '2026-01-01 00:00:20+00'
    OR (
      created_at = TIMESTAMPTZ '2026-01-01 00:00:20+00'
      AND id < 35601
    )
  )
ORDER BY created_at DESC, id DESC
LIMIT 21;

ROLLBACK;
