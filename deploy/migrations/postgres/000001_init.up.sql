-- Utils

DO $$ BEGIN
    CREATE DOMAIN ksuid AS CHAR(27);
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE OR REPLACE FUNCTION ksuid() RETURNS TEXT AS $$
DECLARE
	v_time TIMESTAMP WITH TIME ZONE := NULL;
	v_seconds NUMERIC(50) := NULL;
	v_NUMERIC NUMERIC(50) := NULL;
	v_epoch NUMERIC(50) = 1400000000;
	v_base62 TEXT := '';
	v_alphabet CHAR ARRAY[62] := ARRAY[
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
		'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 
		'U', 'V', 'W', 'X', 'Y', 'Z', 
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 
		'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
		'u', 'v', 'w', 'x', 'y', 'z'];
	i INTEGER := 0;
BEGIN

	v_time := clock_timestamp();
	v_seconds := EXTRACT(EPOCH FROM v_time) - v_epoch;
	v_NUMERIC := v_seconds * pow(2::NUMERIC(50), 128) -- 32 bits for seconds and 128 bits for randomness
		+ ((random()::NUMERIC(70,20) * pow(2::NUMERIC(70,20), 48))::NUMERIC(50) * pow(2::NUMERIC(50), 80)::NUMERIC(50))
		+ ((random()::NUMERIC(70,20) * pow(2::NUMERIC(70,20), 40))::NUMERIC(50) * pow(2::NUMERIC(50), 40)::NUMERIC(50))
		+  (random()::NUMERIC(70,20) * pow(2::NUMERIC(70,20), 40))::NUMERIC(50);

	while v_NUMERIC <> 0 loop
		v_base62 := v_base62 || v_alphabet[mod(v_NUMERIC, 62) + 1];
		v_NUMERIC := div(v_NUMERIC, 62);
	end loop;
	v_base62 := reverse(v_base62);
	v_base62 := lpad(v_base62, 27, '0');

	return v_base62;
	
end $$ language plpgsql;

CREATE OR REPLACE FUNCTION escape_like_string(text) RETURNS TEXT LANGUAGE SQL IMMUTABLE AS $$ SELECT regexp_replace($1, '([\%_])', '\\\1', 'g'); $$;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Tables

CREATE TABLE IF NOT EXISTS music_groups (
    id ksuid DEFAULT ksuid() PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_music_groups_name ON music_groups (name);

CREATE TABLE IF NOT EXISTS songs (
    id ksuid DEFAULT ksuid() PRIMARY KEY,
    music_group_id ksuid NOT NULL REFERENCES music_groups(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    release_date DATE NOT NULL,
    link TEXT NOT NULL,
    UNIQUE (music_group_id, name)
);

CREATE INDEX IF NOT EXISTS idx_songs_name ON songs (name);
CREATE INDEX IF NOT EXISTS idx_songs_music_group_id ON songs (music_group_id);
CREATE INDEX IF NOT EXISTS idx_songs_release_date ON songs (release_date);
CREATE INDEX IF NOT EXISTS idx_songs_link ON songs (link);

CREATE TABLE IF NOT EXISTS song_couplets (
    song_id ksuid DEFAULT ksuid() NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    couplet_num INT NOT NULL,
    text TEXT NOT NULL,
    PRIMARY KEY (song_id, couplet_num)
);

CREATE INDEX IF NOT EXISTS IDX_SONG_COUPLETS_TEXT ON SONG_COUPLETS USING GIN (TEXT gin_trgm_ops);