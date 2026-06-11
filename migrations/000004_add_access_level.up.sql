CREATE TABLE IF NOT EXISTS schema_migrations (version bigint not null primary key, dirty boolean not null);

ALTER TABLE movies ADD COLUMN access_level INT DEFAULT 1;
ALTER TABLE series ADD COLUMN access_level INT DEFAULT 1;
ALTER TABLE live_tv_channels ADD COLUMN access_level INT DEFAULT 1;
ALTER TABLE sports_events ADD COLUMN access_level INT DEFAULT 1;
ALTER TABLE music_content ADD COLUMN access_level INT DEFAULT 1;
