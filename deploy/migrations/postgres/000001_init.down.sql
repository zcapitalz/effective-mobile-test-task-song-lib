DROP TABLE IF EXISTS song_couplets;
DROP TABLE IF EXISTS songs;
DROP TABLE IF EXISTS music_groups;

DROP EXTENSION IF EXISTS pg_trgm;
DROP FUNCTION IF EXISTS escape_like_string(text);
DROP FUNCTION IF EXISTS ksuid();
DROP DOMAIN IF EXISTS ksuid;