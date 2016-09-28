CREATE TABLE projects (
  hash VARCHAR(8) PRIMARY KEY,
  name VARCHAR(31) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE submissions (
  id SERIAL PRIMARY KEY,
  project_hash VARCHAR(8) REFERENCES projects (hash),
  stdout TEXT NOT NULL,
  stderr TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX project_hash ON submissions (project_hash);
