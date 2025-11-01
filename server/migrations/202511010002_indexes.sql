-- +goose Up
-- Indexes and uniqueness

-- Unique category name within a project
ALTER TABLE category
  ADD CONSTRAINT category_project_name_unique UNIQUE (project_id, name);

-- Helpful indexes
CREATE INDEX IF NOT EXISTS category_project_id_idx ON category (project_id);
CREATE INDEX IF NOT EXISTS category_parent_category_id_idx ON category (parent_category_id);
CREATE INDEX IF NOT EXISTS time_entry_category_started_at_idx ON time_entry (category_id, started_at);

-- Optional: accelerate active timer lookups
CREATE INDEX IF NOT EXISTS time_entry_active_idx ON time_entry (category_id) WHERE stopped_at IS NULL;

-- +goose Down
-- Drop optional/regular indexes and constraint
DROP INDEX IF EXISTS time_entry_active_idx;
DROP INDEX IF EXISTS time_entry_category_started_at_idx;
DROP INDEX IF EXISTS category_parent_category_id_idx;
DROP INDEX IF EXISTS category_project_id_idx;
ALTER TABLE category DROP CONSTRAINT IF EXISTS category_project_name_unique;


