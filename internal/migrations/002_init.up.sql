CREATE TABLE IF NOT EXISTS users (
       id INT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS segments (
       id TEXT PRIMARY KEY,
       description TEXT
);

CREATE TABLE IF NOT EXISTS users_segments (
       user_id INT,
       segment_id TEXT,
       CONSTRAINT user_id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
       CONSTRAINT segment_id_fk FOREIGN KEY (segment_id) REFERENCES segments(id) ON DELETE CASCADE,
       PRIMARY KEY (user_id, segment_id)
);

CREATE INDEX IF NOT EXISTS users_segments_user_id_idx ON users_segments(user_id);
CREATE INDEX IF NOT EXISTS users_segments_segment_id_idx ON users_segments(segment_id);