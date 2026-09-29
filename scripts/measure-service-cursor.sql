\set ON_ERROR_STOP on

BEGIN READ ONLY;

SET LOCAL statement_timeout = '120s';
SET LOCAL lock_timeout = '5s';

DO $$
BEGIN
  IF current_database() <> 'lyapus_query_plan_test' THEN
    RAISE EXCEPTION 'refusing to measure a non-experiment database';
  END IF;
END
$$;

\echo 'Deep page: original OR predicate'

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

\echo 'Deep page: row comparison'

EXPLAIN (ANALYZE, BUFFERS)
SELECT id, team_id, slug, name, description, created_at, updated_at
FROM public.services
WHERE team_id = 1
  AND (created_at, id) < (
      TIMESTAMPTZ '2026-01-01 00:00:20+00',
      35601::bigint
  )
ORDER BY created_at DESC, id DESC
LIMIT 21;

ROLLBACK;
