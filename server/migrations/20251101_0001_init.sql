-- +goose Up
-- Initial schema: project, category, time_entry

CREATE TABLE IF NOT EXISTS project (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  description text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS category (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL,
  parent_category_id uuid NULL,
  name text NOT NULL,
  description text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_category_project
    FOREIGN KEY (project_id)
    REFERENCES project (id)
    ON DELETE RESTRICT,
  CONSTRAINT fk_category_parent
    FOREIGN KEY (parent_category_id)
    REFERENCES category (id)
    ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS time_entry (
  id uuid PRIMARY KEY,
  category_id uuid NOT NULL,
  started_at timestamptz NOT NULL,
  stopped_at timestamptz NULL,
  duration_seconds integer NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_time_entry_category
    FOREIGN KEY (category_id)
    REFERENCES category (id)
    ON DELETE RESTRICT
);

-- +goose Down
DROP TABLE IF EXISTS time_entry;
DROP TABLE IF EXISTS category;
DROP TABLE IF EXISTS project;


