-- Drop index "environments_service_id_id_idx" from table: "environments"
DROP INDEX "environments_service_id_id_idx";
-- Create index "environments_service_id_created_at_id_idx" to table: "environments"
CREATE INDEX "environments_service_id_created_at_id_idx" ON "environments" ("service_id", "created_at" DESC, "id" DESC);
-- Drop index "services_team_id_id_idx" from table: "services"
DROP INDEX "services_team_id_id_idx";
-- Create index "services_team_id_created_at_id_idx" to table: "services"
CREATE INDEX "services_team_id_created_at_id_idx" ON "services" ("team_id", "created_at" DESC, "id" DESC);
-- Create index "teams_created_at_id_idx" to table: "teams"
CREATE INDEX "teams_created_at_id_idx" ON "teams" ("created_at" DESC, "id" DESC);
