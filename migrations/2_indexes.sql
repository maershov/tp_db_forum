-- +migrate Up

-- fk user

-- fk forum
CREATE INDEX IF NOT EXISTS idx_forum__forum_user ON forum (forum_user);

-- fk post
CREATE INDEX IF NOT EXISTS idx_post__thread ON post (thread);
CREATE INDEX IF NOT EXISTS idx_post__post_author ON post (post_author);
CREATE INDEX IF NOT EXISTS idx_post__forum ON post (forum);

-- thread posts (parent_tree)
CREATE INDEX IF NOT EXISTS idx_post__thread_parent_path1_post_id ON post (thread, parent, path1, post_id);
CREATE INDEX IF NOT EXISTS idx_post__path1_path ON post (path1, path);

-- thread posts (order by path, tree)
CREATE INDEX idx_post__thread_path ON post (thread, path);

-- thread posts (flat, order by post_created, post_id)
CREATE INDEX idx_post__thread_post_created_post_id ON post (thread, post_created, post_id);

-- fk thread
CREATE INDEX IF NOT EXISTS idx_thread__thread_author ON thread (thread_author);
-- CREATE INDEX idx_thread__forum ON thread (forum);

-- get threads
CREATE INDEX IF NOT EXISTS idx_thread__forum_thread_created ON thread (forum, thread_created);

-- +migrate Down

DROP INDEX IF EXISTS idx_thread__forum_thread_created;
DROP INDEX IF EXISTS idx_thread__thread_author;
DROP INDEX IF EXISTS idx_post__forum;
DROP INDEX IF EXISTS idx_post__post_author;
DROP INDEX IF EXISTS idx_post__thread;
DROP INDEX IF EXISTS idx_forum__forum_user;
