
CREATE TABLE rooms (
    id TEXT PRIMARY KEY, 
    batch_size INT DEFAULT 16,
    preferences JSONB[] DEFAULT '{}',
    participants INT DEFAULT 1, 
    status TEXT DEFAULT 'free'
);
