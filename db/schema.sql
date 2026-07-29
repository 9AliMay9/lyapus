CREATE TABLE teams (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug text NOT NULL,
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT teams_slug_key UNIQUE (slug),
  CONSTRAINT teams_slug_format_check
    CHECK (slug ~ '^[a-z][a-z0-9-]{0,62}$'),
  CONSTRAINT teams_name_check
    CHECK (name = btrim(name) AND char_length(name) BETWEEN 1 AND 100),
  CONSTRAINT teams_updated_at_check
    CHECK (updated_at >= created_at)
  );

  CREATE INDEX teams_created_at_id_idx
      ON teams (created_at DESC, id DESC);

  CREATE TABLE services (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_id bigint NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT services_team_id_slug_key UNIQUE (team_id, slug),
    CONSTRAINT services_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE RESTRICT,
    CONSTRAINT services_slug_format_check
        CHECK (slug ~ '^[a-z][a-z0-9-]{0,62}$'),
    CONSTRAINT services_name_check
        CHECK (name = btrim(name) AND char_length(name) BETWEEN 1 AND 100),
    CONSTRAINT services_description_length_check
        CHECK (description IS NULL OR char_length(description) <= 500),
    CONSTRAINT services_updated_at_check
        CHECK (updated_at >= created_at)
  );

  CREATE INDEX services_team_id_created_at_id_idx
      ON services (team_id, created_at DESC, id DESC);

  CREATE TABLE environments (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_id bigint NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT environments_service_id_slug_key UNIQUE (service_id, slug),
    CONSTRAINT environments_service_id_fkey
        FOREIGN KEY (service_id) REFERENCES services (id) ON DELETE RESTRICT,
    CONSTRAINT environments_slug_format_check
        CHECK (slug ~ '^[a-z][a-z0-9-]{0,62}$'),
    CONSTRAINT environments_name_check
        CHECK (name = btrim(name) AND char_length(name) BETWEEN 1 AND 100),
    CONSTRAINT environments_updated_at_check
        CHECK (updated_at >= created_at)
  );

  CREATE INDEX environments_service_id_created_at_id_idx
      ON environments (service_id, created_at DESC, id DESC);
