-- Seed Series
INSERT INTO series (id, title, description, director, cast_members, rating, poster_url, backdrop_url) 
VALUES (1, 'Sintel Journey', 'Sintel is an open-source CGI film about a girl searching for a baby dragon.', 'Colin Levy', '["Halina Reijn", "Thom Hoffman"]', 8.2, 'https://upload.wikimedia.org/wikipedia/commons/8/8c/Sintel_poster.jpg', 'https://durian.blender.org/wp-content/uploads/2010/10/sintel_and_scales.jpg')
ON DUPLICATE KEY UPDATE title=VALUES(title), description=VALUES(description), director=VALUES(director), rating=VALUES(rating), poster_url=VALUES(poster_url), backdrop_url=VALUES(backdrop_url);

-- Seed Seasons
INSERT INTO seasons (id, series_id, season_number, title, description)
VALUES 
(1, 1, 1, 'Season 1: Discovery', 'The journey begins as Sintel meets and trains the baby dragon.'),
(2, 1, 2, 'Season 2: Destiny', 'Sintel travels the world in search of her lost companion.')
ON DUPLICATE KEY UPDATE title=VALUES(title), description=VALUES(description);

-- Seed Episodes
INSERT INTO episodes (id, season_id, episode_number, title, description, video_sources, subtitles)
VALUES
(1, 1, 1, 'The Encounter', 'Sintel rescues a baby dragon and names him Scales.', '[{"url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/Sintel.mp4", "quality": "Original"}]', '{}'),
(2, 1, 2, 'Flight Training', 'Sintel teaches Scales how to fly and forage.', '[{"url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/TearsOfSteel.mp4", "quality": "Original"}]', '{}'),
(3, 2, 1, 'The Quest', 'Sintel sets off on a dangerous path across the dry lands.', '[{"url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4", "quality": "Original"}]', '{}'),
(4, 2, 2, 'Reunion', 'Sintel finally finds the dragon, but he has grown and forgotten her.', '[{"url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4", "quality": "Original"}]', '{}')
ON DUPLICATE KEY UPDATE title=VALUES(title), description=VALUES(description), video_sources=VALUES(video_sources);

-- Seed Settings
INSERT INTO app_settings (setting_key, setting_value) VALUES
('up_next_timer', '10'),
('frontend_menu', '[{"label":"Live TV", "path":"/live", "level_required": 1, "order": 1, "enabled": true}, {"label":"Movies", "path":"/movies", "level_required": 2, "order": 2, "enabled": true}, {"label":"Series", "path":"/series", "level_required": 2, "order": 3, "enabled": true}, {"label":"Music", "path":"/music", "level_required": 2, "order": 4, "enabled": true}, {"label":"Sports", "path":"/sports", "level_required": 2, "order": 5, "enabled": true}, {"label":"Profile", "path":"/profile", "level_required": 1, "order": 6, "enabled": true}]')
ON DUPLICATE KEY UPDATE setting_value=VALUES(setting_value);

-- Seed Categories
INSERT IGNORE INTO movie_categories (id, name, slug) VALUES 
(1, 'Action', 'action'), (2, 'Drama', 'drama'), (3, 'Comedy', 'comedy');

INSERT IGNORE INTO series_categories (id, name, slug) VALUES 
(1, 'Sci-Fi', 'sci-fi'), (2, 'Fantasy', 'fantasy'), (3, 'Documentary', 'documentary');

INSERT IGNORE INTO live_tv_categories (id, name, slug) VALUES 
(1, 'News', 'news'), (2, 'Entertainment', 'entertainment');

INSERT IGNORE INTO sports_categories (id, name, slug) VALUES 
(1, 'Football', 'football'), (2, 'Basketball', 'basketball');

INSERT IGNORE INTO music_categories (id, name, slug) VALUES 
(1, 'Pop', 'pop'), (2, 'Rock', 'rock'), (3, 'Concerts', 'concerts');

-- Seed Category Mapping (Sintel is a series, linking to Fantasy)
INSERT IGNORE INTO series_category_mapping (series_id, category_id) VALUES (1, 2);
