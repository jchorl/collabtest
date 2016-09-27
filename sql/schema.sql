CREATE TABLE projects (
  id SERIAL PRIMARY KEY,
  name VARCHAR(31) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE submissions (
  id SERIAL PRIMARY KEY,
  project_id INTEGER REFERENCES projects (id),
  stdout TEXT NOT NULL,
  stderr TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX submissions_project_id_idx ON submissions (project_id);
