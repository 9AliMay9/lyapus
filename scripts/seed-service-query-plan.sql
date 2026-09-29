\set ON_ERROR_STOP on

BEGIN;

SET LOCAL statement_timeout = '120s';
SET LOCAL lock_timeout = '5s';
SET LOCAL search_path = public, pg_catalog;

DO $$
BEGIN
  IF current_database() <> 'lyapus_query_plan_test' THEN
    RAISE EXCEPTION 'refusing to seed a non-experiment database';
  END IF;
END
$$;

LOCK TABLE teams, services, environments
  IN SHARE ROW EXCLUSIVE MODE;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM teams)
    OR EXISTS (SELECT 1 FROM services)
    OR EXISTS (SELECT 1 FROM environments) THEN
    RAISE EXCEPTION 'experiment tables must be empty';
  END IF;
END
$$;

INSERT INTO teams (slug, name, created_at, updated_at)
SELECT
  'plan-team-' || n,
  'Plan Team ' || n,
  TIMESTAMPTZ '2026-01-01 00:00:00+00',
  TIMESTAMPTZ '2026-01-01 00:00:00+00'
FROM generate_series(1, 100) AS series(n)
ORDER BY n;

INSERT INTO services (
  team_id, slug, name, description, created_at, updated_at
)
SELECT
  t.id,
  'plan-service-' || n,
  'Plan Service ' || n,
  repeat('x', 100),
  TIMESTAMPTZ '2026-01-01 00:00:00+00'
    + (((n * 37) % 1000) / 10) * INTERVAL '1 second',
  TIMESTAMPTZ '2026-01-01 00:00:00+00'
    + (((n * 37) % 1000) / 10) * INTERVAL '1 second'
FROM teams AS t
CROSS JOIN generate_series(1, 1000) AS series(n)
ORDER BY n, t.id;

COMMIT;

ANALYZE public.teams;
ANALYZE public.services;

SELECT
  (SELECT count(*) FROM public.teams) AS team_count,
  (SELECT count(*) FROM public.services) AS service_count,
  (SELECT count(*) FROM public.environments) AS environment_count;
