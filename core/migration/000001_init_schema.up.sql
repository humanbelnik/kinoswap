
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY, 
    batch_size INT DEFAULT 16,
    preferences TEXT[] DEFAULT '{}',
    participants INT DEFAULT 1, 
    status TEXT DEFAULT 'free'
);

CREATE TABLE IF NOT EXISTS movies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT,
    year INT,
    rating NUMERIC,
    genres TEXT[] DEFAULT '{}',
    overview TEXT,
    poster_link TEXT
);
