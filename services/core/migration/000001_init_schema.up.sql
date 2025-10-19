
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_admin UUID NOT NULL,
    code TEXT UNIQUE,
    ready INT DEFAULT 0,
    status TEXT DEFAULT 'LOBBY'
);

CREATE TABLE IF NOT EXISTS participants (
    id UUID PRIMARY KEY NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE, 
    preference VECTOR(384)
);

CREATE TABLE IF NOT EXISTS reactions (
    id UUID PRIMARY KEY NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE, 
    movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    likes INT DEFAULT 0
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

