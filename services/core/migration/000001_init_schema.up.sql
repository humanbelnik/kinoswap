
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
 CREATE EXTENSION IF NOT EXISTS "vector";

CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY, 
    _uuid UUID UNIQUE DEFAULT uuid_generate_v4(), 
    batch_size INT DEFAULT 16,
    preferences TEXT[] DEFAULT '{}',
    preferences_vec VECTOR(384)[], 
    participants INT DEFAULT 1, 
    status TEXT DEFAULT 'free',
    preference_vector VECTOR(384)
);

CREATE TABLE IF NOT EXISTS movies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT,
    year INT,
    rating NUMERIC,
    genres TEXT[] DEFAULT '{}', 
    overview TEXT,
    poster_link TEXT,
    movie_vector VECTOR(384)
);

CREATE TABLE IF NOT EXISTS results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    pass_count INT DEFAULT 0,
    UNIQUE(room_id, movie_id)
);


CREATE INDEX IF NOT EXISTS rooms_embedding_idx ON rooms 
USING ivfflat (preference_vector vector_cosine_ops) WITH (lists = 100);

CREATE INDEX IF NOT EXISTS movies_embedding_idx ON movies 
USING ivfflat (movie_vector vector_cosine_ops) WITH (lists = 100);