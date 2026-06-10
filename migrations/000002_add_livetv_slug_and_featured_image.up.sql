ALTER TABLE featured_items ADD COLUMN image_url VARCHAR(512);

ALTER TABLE live_tv_channels ADD COLUMN slug VARCHAR(255);
UPDATE live_tv_channels SET slug = CONCAT('channel-', id) WHERE slug IS NULL;
ALTER TABLE live_tv_channels MODIFY COLUMN slug VARCHAR(255) NOT NULL, ADD UNIQUE (slug);
