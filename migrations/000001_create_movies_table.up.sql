CREATE TABLE IF NOT EXISTS movies (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    year integer NOT NULL,
    runtime integer NOT NULL,
    genres text[] NOT NULL,
    director varchar(255) NOT NULL,
    actors text[] NOT NULL,
    plot text NOT NULL,
    poster_url text NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'GMT+7'),
    version integer NOT NULL DEFAULT 1
);