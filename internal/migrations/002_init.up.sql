CREATE TABLE IF NOT EXISTS users (
       id TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS segments (
       id TEXT PRIMARY KEY,
       description TEXT
);

CREATE TABLE IF NOT EXISTS users_segments (
       user_id TEXT,
       segment_id TEXT,
       CONSTRAINT user_id_fk FOREIGN KEY (user_id) REFERENCES users(id),
       CONSTRAINT segment_id_fk FOREIGN KEY (segment_id) REFERENCES segments(id),
       PRIMARY KEY (user_id, segment_id)
);