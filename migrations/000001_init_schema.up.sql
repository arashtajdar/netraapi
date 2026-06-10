CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    virtual_coins INT DEFAULT 500,
    user_level INT DEFAULT 1,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    avatar_url VARCHAR(512),
    settings JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email)
);

CREATE TABLE IF NOT EXISTS user_profiles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(512),
    is_kids_mode BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS watchlists (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    profile_id INT,
    name VARCHAR(255) NOT NULL DEFAULT 'Watch List',
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (profile_id) REFERENCES user_profiles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS movies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    release_date DATE,
    director VARCHAR(255),
    cast_members JSON,
    imdb_rating DECIMAL(3, 1),
    local_rating DECIMAL(3, 1),
    poster_url VARCHAR(512),
    backdrop_url VARCHAR(512),
    video_sources JSON,
    subtitles JSON,
    intro_start INT,
    intro_end INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS watchlist_items (
    watchlist_id INT NOT NULL,
    content_id INT NOT NULL,
    content_type ENUM('movie', 'series', 'music') NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (watchlist_id, content_id, content_type),
    FOREIGN KEY (watchlist_id) REFERENCES watchlists(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS series (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    director VARCHAR(255),
    cast_members JSON,
    rating DECIMAL(3, 1),
    poster_url VARCHAR(512),
    backdrop_url VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS seasons (
    id INT AUTO_INCREMENT PRIMARY KEY,
    series_id INT NOT NULL,
    season_number INT NOT NULL,
    title VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(series_id, season_number),
    FOREIGN KEY (series_id) REFERENCES series(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS episodes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    season_id INT NOT NULL,
    episode_number INT NOT NULL,
    title VARCHAR(255),
    description TEXT,
    video_sources JSON,
    subtitles JSON,
    intro_start INT,
    intro_end INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(season_id, episode_number),
    FOREIGN KEY (season_id) REFERENCES seasons(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS live_tv_channels (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    stream_url VARCHAR(512),
    logo_url VARCHAR(512),
    youtube_url VARCHAR(512),
    youtube_channel_url VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS epg (
    id INT AUTO_INCREMENT PRIMARY KEY,
    channel_id INT NOT NULL,
    program_title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES live_tv_channels(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sports_events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_live BOOLEAN DEFAULT FALSE,
    live_stream_url VARCHAR(512),
    video_sources JSON,
    start_time TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_watch_history (
    user_id INT,
    profile_id INT,
    movie_id INT,
    resume_position_seconds INT DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id, movie_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (profile_id) REFERENCES user_profiles(id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS watch_party_rooms (
    id INT AUTO_INCREMENT PRIMARY KEY,
    room_code VARCHAR(10) UNIQUE NOT NULL,
    host_id INT NOT NULL,
    current_movie_id INT,
    is_playing BOOLEAN DEFAULT FALSE,
    current_position_seconds INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_room_code (room_code),
    FOREIGN KEY (host_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (current_movie_id) REFERENCES movies(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS featured_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    content_type ENUM('movie', 'series', 'live_tv', 'sports', 'music') NOT NULL,
    content_id INT NOT NULL,
    custom_description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS app_settings (
    setting_key VARCHAR(255) UNIQUE NOT NULL,
    setting_value JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY(setting_key)
);

CREATE TABLE IF NOT EXISTS music_content (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    artist VARCHAR(255),
    video_sources JSON,
    audio_sources JSON,
    poster_url VARCHAR(512),
    backdrop_url VARCHAR(512),
    release_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Categories and Mappings
CREATE TABLE IF NOT EXISTS movie_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS movie_category_mapping (
    movie_id INT NOT NULL,
    category_id INT NOT NULL,
    PRIMARY KEY (movie_id, category_id),
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES movie_categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS series_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS series_category_mapping (
    series_id INT NOT NULL,
    category_id INT NOT NULL,
    PRIMARY KEY (series_id, category_id),
    FOREIGN KEY (series_id) REFERENCES series(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES series_categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS live_tv_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS live_tv_category_mapping (
    channel_id INT NOT NULL,
    category_id INT NOT NULL,
    PRIMARY KEY (channel_id, category_id),
    FOREIGN KEY (channel_id) REFERENCES live_tv_channels(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES live_tv_categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sports_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS sports_category_mapping (
    event_id INT NOT NULL,
    category_id INT NOT NULL,
    PRIMARY KEY (event_id, category_id),
    FOREIGN KEY (event_id) REFERENCES sports_events(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES sports_categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS music_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS music_category_mapping (
    music_id INT NOT NULL,
    category_id INT NOT NULL,
    PRIMARY KEY (music_id, category_id),
    FOREIGN KEY (music_id) REFERENCES music_content(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES music_categories(id) ON DELETE CASCADE
);
