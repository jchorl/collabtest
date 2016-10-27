CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  github_id INTEGER,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE projects (
  hash VARCHAR(8) PRIMARY KEY,
  user_id INTEGER REFERENCES users (id),
  name VARCHAR(31) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE runs (
  id SERIAL PRIMARY KEY,
  project_hash VARCHAR(8) REFERENCES projects (hash),
  stdout TEXT NOT NULL,
  stderr TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX project_hash_idx ON runs (project_hash);
CREATE INDEX github_id_idx ON users (github_id);
